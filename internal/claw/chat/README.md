# claw/chat

`internal/claw/chat` 负责承接 Boxify 内部聊天能力的后端链路：创建会话、保存消息、将用户输入投递到 OpenClaw 的 `boxify` 原生 channel inbox、消费插件返回的 SSE 流式事件，并把结果同步到本地状态与前端事件总线。

更高层的双端通信背景可参考：[docs/boxify-channel-architecture.md](/Users/sheepzhao/WorkSpace/Boxify/docs/boxify-channel-architecture.md)。

## 目标与边界

这个包只处理 Boxify 聊天链路中的“后端协调层”职责，不负责：

- 前端视图渲染与交互状态管理
- OpenClaw 插件进程的完整生命周期管理
- 持久化数据库实现
- SSE 重连、重试、断点续传等高级传输能力

这个包明确负责：

- 创建和查询 Boxify 侧聊天会话
- 保存会话消息
- 构造发送请求上下文
- 调用原生 channel inbox 的 HTTP / SSE 接口
- 消费 `start / delta / done / error` 流事件
- 将流式回复整理为本地 assistant 草稿/完成消息
- 向前端广播统一格式的聊天事件

## 文件结构

- `models.go`
  定义会话、消息、请求、响应、流事件、前端事件等核心模型。
- `store.go`
  定义 `ConversationStore` 抽象，并提供当前默认实现 `MemoryConversationStore`。
- `request_factory.go`
  负责把一次发送动作整理成 `SendMessageEnvelope`。
- `channel_client.go`
  定义 `ChannelClient` 抽象，并提供基于本地 HTTP 的 `HTTPChannelClient`。
- `stream_handler.go`
  负责把插件 SSE 事件翻译为本地状态变化和前端事件。
- `publisher.go`
  定义 `EventPublisher` 抽象，并提供 Wails 事件发布实现。
- `channel_coordinator.go`
  包的主入口，负责编排一次完整的发送流程。

## 对外角色关系

### 1. `ChannelCoordinator`

`ChannelCoordinator` 是包内的总协调器，对上提供简单的会话和消息接口：

- `CreateConversation(agentID string)`
- `ListConversations()`
- `ListMessages(conversationID string)`
- `SendMessage(ctx, conversationID, text)`

它不直接解析 SSE，也不直接拼接流式 assistant 文本，而是依赖以下协作者：

- `ConversationStore`
- `ChannelClient`
- `EventPublisher`
- `StreamEventHandler`
- `clawprocess.Manager`

### 2. `ConversationStore`

`ConversationStore` 抽象本地会话与消息存储，当前接口如下：

```go
type ConversationStore interface {
	CreateConversation(agentID string) (*Conversation, error)
	ListConversations() ([]Conversation, error)
	GetConversation(id string) (*Conversation, error)
	UpdateOpenClawSessionID(conversationID, sessionID string) error
	AppendMessage(msg Message) error
	ListMessages(conversationID string) ([]Message, error)
	UpdateAssistantDraft(conversationID, runID, chunk string) error
	FinalizeAssistantMessage(conversationID, runID string) error
}
```

当前默认实现为 `MemoryConversationStore`，只保存在进程内存里，应用重启后数据会丢失。

### 3. `ChannelClient`

`ChannelClient` 抽象 Boxify 到 OpenClaw channel inbox 的传输能力：

```go
type ChannelClient interface {
	SendMessage(ctx context.Context, req ChannelInboxRequest) (*ChannelInboxResponse, error)
	SendMessageStream(ctx context.Context, req ChannelInboxRequest, onEvent func(ChatStreamEvent) error) (*ChannelInboxResponse, error)
}
```

当前主要使用 `SendMessageStream`，由 `HTTPChannelClient` 通过本地 HTTP + SSE 实现。

### 4. `StreamEventHandler`

`StreamEventHandler` 负责处理插件返回的流式事件：

- 更新 OpenClaw `sessionId` 映射
- 将 `delta` 合并到 assistant 草稿
- 将 `done` 收敛为完成态消息
- 将统一格式事件广播给前端

### 5. `EventPublisher`

`EventPublisher` 抽象“如何把聊天事件发给前端”：

```go
type EventPublisher interface {
	PublishConversationEvent(conversationID string, event ChatReplyEvent)
}
```

当前实现 `WailsEventPublisher` 会通过 Wails 广播 `claw:chat-event`。

## 运行时装配

`ClawService.rebuildManagers()` 会创建本包所需依赖，并最终装配出一个 `ChannelCoordinator`：

```go
s.chatCoordinator = clawchat.NewChannelCoordinator(
	clawchat.NewMemoryConversationStore(),
	clawchat.NewHTTPChannelClient(fmt.Sprintf("http://127.0.0.1:%d", s.pluginPort), s.chatToken),
	clawchat.NewWailsEventPublisher(s.App(), s.Logger()),
	s.manager,
	s.Logger(),
)
```

当前默认组合为：

1. `MemoryConversationStore`
2. `HTTPChannelClient`
3. `WailsEventPublisher`
4. `clawprocess.Manager`
5. `StreamEventHandler` 由 `NewChannelCoordinator()` 内部创建

## 运行参数与配置来源

### 插件地址

`HTTPChannelClient` 默认基地址为：

```text
http://127.0.0.1:32124
```

在 `ClawService` 中可通过环境变量覆盖：

- `BOXIFY_PLUGIN_INBOX_PORT`

最终访问的路径常量定义在 `channel_client.go`：

- `BoxifyInboxPath = /channels/boxify/inbox`
- `BoxifyInboxStreamPath = /channels/boxify/inbox/stream`

### 共享令牌

聊天通道的共享令牌优先级如下：

1. 环境变量 `BOXIFY_CHAT_SHARED_TOKEN`
2. `dataDir/boxify.json` 中保存的本地 token
3. 若不存在则首次自动生成，并回写 `boxify.json`

首次自动生成后，`ClawService` 还会尝试将该 token 同步写入 `openclaw.json` 的 `channels.boxify.sharedToken`。

### OpenClaw 进程拉起

如果 `ChannelCoordinator` 注入了 `clawprocess.Manager`，发送消息前会调用：

```go
manager.Start()
```

这里的策略是“尽量拉起，但不阻断发送”：

- 启动成功：继续请求插件
- 启动失败：只记录 `Warn` 日志，仍然继续请求 HTTP inbox

这样可以兼容 OpenClaw 已经由外部启动的场景。

## 核心数据模型

### `Conversation`

`Conversation` 表示 Boxify 侧维护的聊天会话：

```go
type Conversation struct {
	ID                string
	Title             string
	AgentID           string
	OpenClawSessionID string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
```

字段说明：

- `ID`
  Boxify 本地会话 ID，当前格式类似 `conv_<unix_nano>`。
- `Title`
  当前内存实现固定初始化为 `新会话`。
- `AgentID`
  当前会话绑定的目标 agent；空值时自动回退为 `main`。
- `OpenClawSessionID`
  与 OpenClaw 对应的会话 ID 映射，初始可能为空。
- `CreatedAt` / `UpdatedAt`
  会话创建时间与最近更新时间。

### `Message`

`Message` 表示会话中的单条消息：

```go
type Message struct {
	ID             string
	ConversationID string
	RunID          string
	Role           string
	Content        string
	Status         string
	CreatedAt      time.Time
}
```

字段说明：

- `ID`
  Boxify 本地消息 ID，当前格式类似 `msg_<unix_nano>`。
- `ConversationID`
  所属会话 ID。
- `RunID`
  一次用户发送动作对应的执行 ID，用户消息与本轮 assistant 回复共用同一个 `runId`。
- `Role`
  当前约定值：`user`、`assistant`、`system`。
- `Content`
  消息文本内容。
- `Status`
  当前实际使用值：`streaming`、`done`，错误回复可能以事件表达而非消息状态表达。
- `CreatedAt`
  消息创建时间。

### `ChannelInboxRequest`

发往插件 inbox 的请求结构：

```go
type ChannelInboxRequest struct {
	ConversationID string
	MessageID      string
	RunID          string
	AgentID        string
	Text           string
	Metadata       map[string]interface{}
}
```

当前由 `BuildSendMessageEnvelope()` 统一构造，`Metadata` 默认包含：

```json
{
  "source": "boxify"
}
```

### `ChannelInboxResponse`

插件同步返回的聚合结果：

```go
type ChannelInboxResponse struct {
	OK             bool
	ConversationID string
	SessionID      string
	Text           string
	Error          string
}
```

对于流式请求来说，这个结构是客户端在消费完整个 SSE 流后整理出来的结果汇总，不一定对应插件一次性返回的 JSON。

### `ChatStreamEvent`

插件 SSE 事件结构：

```go
type ChatStreamEvent struct {
	EventType      ChatEventType
	ConversationID string
	SessionID      string
	RunID          string
	Text           string
	Error          string
	Payload        map[string]interface{}
}
```

当前支持的流式事件类型：

- `start`
- `delta`
- `done`
- `error`

### `ChatEvent`

发送给前端的统一事件结构：

```go
type ChatEvent struct {
	ConversationID string
	SessionID      string
	RunID          string
	EventType      ChatEventType
	Payload        map[string]interface{}
	Timestamp      int64
}
```

这是前端实际监听到的 Wails 事件负载。

## 会话与消息的本地存储语义

### 创建会话

`MemoryConversationStore.CreateConversation()` 的行为：

- 生成 `conv_<unix_nano>` 作为会话 ID
- 空 `agentID` 自动回退到 `main`
- 标题初始化为 `新会话`
- 记录 `CreatedAt` 和 `UpdatedAt`

### 列表排序

`ListConversations()` 会按 `UpdatedAt` 倒序返回，最近活跃会话排在最前。

### 用户消息写入

每次调用 `SendMessage()` 时，会先创建一条用户消息并写入 store：

- `Role = user`
- `Status = done`
- `RunID = 当前发送 runId`

也就是说，即便后续请求插件失败，用户消息通常已经落库。

### assistant 草稿更新

`UpdateAssistantDraft()` 的行为：

1. 倒序查找当前会话中最近一条：
   - `Role == assistant`
   - `RunID == 当前 runId`
   - `Status == streaming`
2. 如果找到，则把新 `chunk` 追加到原有 `Content`
3. 如果找不到，则新建一条 assistant 消息：
   - `Role = assistant`
   - `Status = streaming`
   - `Content = 当前 chunk`

这保证了一次回复流会在本地只形成一条持续增长的 assistant 草稿消息。

### assistant 完成收敛

`FinalizeAssistantMessage()` 会倒序找到当前 `runId` 对应的 assistant 消息，并将其状态改为：

```text
done
```

如果没有找到对应 assistant 消息，当前实现直接返回 `nil`，不会报错。

## 一次完整发送流程

下面是当前实现中的真实时序。

### 第 1 步：前端创建会话

前端通过 `ClawService.CreateChatConversation(agentID)` 创建会话，最终调用：

```go
chatCoordinator.CreateConversation(agentID)
```

此时只会创建 Boxify 本地会话：

- 有 `conversationId`
- 有 `agentId`
- 没有 `OpenClawSessionID`

### 第 2 步：前端发送消息

前端通过 `ClawService.SendChatMessage(conversationID, text)` 发送消息，最终调用：

```go
chatCoordinator.SendMessage(ctx, conversationID, text)
```

`SendMessage()` 的顺序如下：

1. 清理输入参数首尾空白
2. 校验 `conversationID` 非空
3. 校验 `text` 非空
4. 从 store 读取会话
5. 尝试通过 `manager.Start()` 拉起 OpenClaw
6. 调用 `BuildSendMessageEnvelope()` 生成：
   - `runId`
   - 用户消息
   - `ChannelInboxRequest`
7. 先把用户消息写入 store
8. 若未配置 `ChannelClient`，仅本地入库后直接返回 `runId`
9. 若已配置 `ChannelClient`，调用 `SendMessageStream()`
10. 将每个 SSE 事件交给 `StreamEventHandler.Handle()`
11. 请求结束后，若聚合结果中带 `sessionId`，再更新一次会话映射
12. 返回本次发送的 `runId`

### 第 3 步：构造发送载荷

`BuildSendMessageEnvelope()` 会生成两个共享同一标识上下文的对象：

- 一条本地用户消息
- 一份投递给插件的 inbox 请求

生成规则：

- `runId = run_<unix_nano>`
- `messageId = msg_<unix_nano>`

对应关系如下：

| 对象 | 关键字段 |
| --- | --- |
| 用户消息 | `Role=user`、`Status=done`、`RunID=runId` |
| Inbox 请求 | `MessageID=messageId`、`RunID=runId`、`Text=text` |

### 第 4 步：发起 HTTP / SSE 请求

`HTTPChannelClient.SendMessageStream()` 会向插件发起：

```text
POST /channels/boxify/inbox/stream
```

完整地址形如：

```text
POST http://127.0.0.1:<pluginPort>/channels/boxify/inbox/stream
```

请求头：

- `Content-Type: application/json`
- `Accept: text/event-stream`
- `X-Boxify-Token: <sharedToken>`，仅在 token 非空时携带

请求体示例：

```json
{
  "conversationId": "conv_1741760000000000000",
  "messageId": "msg_1741760000000000100",
  "runId": "run_1741760000000000000",
  "agentId": "main",
  "text": "你好",
  "metadata": {
    "source": "boxify"
  }
}
```

注意：

- 这里的 `conversationId` 是 Boxify 本地会话 ID
- 它不是 OpenClaw 的 `sessionId`
- OpenClaw 的真实会话映射由后续返回的 `sessionId` 建立

## SSE 解析与事件处理

### SSE 读取规则

`HTTPChannelClient.SendMessageStream()` 使用 `bufio.Reader` 按行读取响应，并按标准 SSE 语义处理：

- `event:` 行记录事件名
- `data:` 行累积事件体
- 空行触发一次事件 flush

flush 时会：

1. 将多行 `data:` 用 `\n` 拼接
2. 反序列化为 `ChatStreamEvent`
3. 若 JSON 内未写 `eventType`，则回退使用 `event:` 行的值
4. 更新聚合结果 `ChannelInboxResponse`
5. 调用上层传入的 `onEvent`

### 聚合结果规则

在流式读取过程中，`HTTPChannelClient` 会维护一个聚合结果：

- `ConversationID`
  记录事件中最新的 `conversationId`
- `SessionID`
  记录事件中最新的 `sessionId`
- `Text`
  在收到 `done` 事件时记录最终文本
- `OK`
  初始为 `true`，收到 `error` 事件后改为 `false`
- `Error`
  记录错误文本

流结束后：

- 若 `OK == true`，返回聚合结果
- 若 `OK == false`，返回聚合结果和错误

### `StreamEventHandler` 的处理规则

`StreamEventHandler.Handle()` 先做两件通用工作：

1. 计算本次事件使用的 `runId`
   - 优先使用事件内的 `event.RunID`
   - 若为空，则回退到 `fallbackRunID`
2. 若事件里带了 `sessionId`，尝试更新本地会话的 `OpenClawSessionID`

之后按 `eventType` 分支：

#### `delta`

- 空文本直接忽略
- 调用 `store.UpdateAssistantDraft(conversationID, runID, event.Text)`
- 广播前端事件 `assistant_delta`

事件负载格式：

```json
{
  "text": "<增量文本>"
}
```

#### `done`

- 调用 `store.FinalizeAssistantMessage(conversationID, runID)`
- 广播前端事件 `assistant_done`

事件负载格式：

```json
{
  "text": "<最终文本>"
}
```

#### `error`

- 不会把 assistant 消息状态改成 `error`
- 直接广播前端事件 `assistant_error`

事件负载格式：

```json
{
  "error": "<错误信息>"
}
```

#### `start`

当前实现不会落库，也不会向前端转发 `start`。

## 前端事件发布

`WailsEventPublisher.PublishConversationEvent()` 会发出 Wails 事件：

```text
claw:chat-event
```

事件类型常量来源于：

- `internal/events/types.go`
- `EventTypeClawChatEvent = "claw:chat-event"`

实际广播的载荷为 `ChatEvent`，字段包括：

- `conversationId`
- `sessionId`
- `runId`
- `eventType`
- `payload`
- `timestamp`

当前前端主要需要处理的事件类型：

- `assistant_delta`
- `assistant_done`
- `assistant_error`

## 错误与失败语义

### 输入校验失败

`ChannelCoordinator.SendMessage()` 在以下情况会直接返回错误，不会请求插件：

- `conversationID` 为空
- `text` 为空
- 会话不存在

对应错误示例：

- `会话 ID 不能为空`
- `消息内容不能为空`
- `会话不存在`

### 本地入库后未配置 client

如果 `ChannelCoordinator.client == nil`：

- 用户消息仍然会先写入 store
- 不会请求插件
- 会记录 `Warn` 日志
- 仍然返回 `runId`

这适合本地开发阶段只验证消息入库行为。

### HTTP 请求失败

以下情况会返回错误：

- 请求构造失败
- 连接插件失败
- HTTP 状态码非 2xx
- 解析同步 JSON 响应失败
- 解析 SSE 事件失败
- `onEvent` 回调返回错误

需要注意：

- 用户消息通常已经入库
- assistant 消息可能完全没有创建
- 也可能停留在 `streaming` 状态，取决于失败发生时机

### 插件拒绝请求

`channel_client.go` 定义了：

```go
var ErrChannelRequestRejected = errors.New("插件端拒绝请求")
```

下列情况会返回该错误或其包装错误：

- 插件返回非 2xx
- 同步响应 `OK == false`
- SSE 流中收到 `error` 事件

如果插件附带了具体错误文本，最终错误会包装该文本。

### OpenClaw 会话映射更新失败

无论在流式处理中，还是在请求结束后的聚合结果处理中，只要 `sessionId` 存在，代码都会尝试调用：

```go
store.UpdateOpenClawSessionID(conversationID, sessionID)
```

如果更新失败：

- 只记录 `Warn`
- 不中断主流程

这意味着“会话映射更新失败”被视为可降级问题，而不是致命错误。

## 当前实现中的几个重要细节

### 1. `runId` 是一轮消息交互的主关联键

同一轮发送中的：

- 本地用户消息
- assistant 流式草稿
- assistant 完成消息
- 前端流事件

都通过 `runId` 进行关联。

### 2. `sessionId` 不一定在一开始就有

会话刚创建时只有 Boxify 本地 `conversationId`。

只有插件在流式或最终响应中返回 `sessionId` 后，Boxify 才会建立与 OpenClaw 的会话映射。

### 3. `assistant_error` 当前主要是事件语义，不是消息状态语义

当前 store 中的 `Message.Status` 并不会在错误时自动改成 `error`。也就是说：

- 错误态主要通过前端事件感知
- 本地 assistant 消息可能仍然是已有草稿，或没有落库

如果后续要支持聊天历史中的明确失败态，需要扩展 store 接口和消息模型。

### 4. `start` 事件当前被消费但不对外透出

客户端能解析 `start`，但 `StreamEventHandler` 不会落库，也不会向前端广播。若未来需要展示“已建立流”或“模型正在响应”，可以从这里扩展。

### 5. 当前没有持久化恢复能力

默认 `MemoryConversationStore` 重启即丢失，因此：

- 聊天记录不会跨进程保存
- `OpenClawSessionID` 映射也不会跨进程保留

## 扩展建议

如果后续要演进这个包，通常会落在下面几个方向：

- 持久化 store
  用 SQLite 或本地文件替换 `MemoryConversationStore`。
- 错误态消息模型
  为 assistant 回复增加明确的 `error` 状态和错误信息字段。
- `start` 事件透传
  支持前端更细粒度的流状态显示。
- 重试与超时策略
  为 `HTTPChannelClient` 增加更稳健的网络错误处理。
- 幂等与重复保护
  防止重复提交同一轮消息导致本地消息重复。
- 会话标题生成
  当前标题固定为 `新会话`，未来可在首轮消息后自动摘要生成标题。

## 推荐阅读顺序

建议按下面顺序读代码，最容易建立整体认知：

1. [channel_coordinator.go](/Users/sheepzhao/WorkSpace/Boxify/internal/claw/chat/channel_coordinator.go)
2. [request_factory.go](/Users/sheepzhao/WorkSpace/Boxify/internal/claw/chat/request_factory.go)
3. [channel_client.go](/Users/sheepzhao/WorkSpace/Boxify/internal/claw/chat/channel_client.go)
4. [stream_handler.go](/Users/sheepzhao/WorkSpace/Boxify/internal/claw/chat/stream_handler.go)
5. [store.go](/Users/sheepzhao/WorkSpace/Boxify/internal/claw/chat/store.go)
6. [publisher.go](/Users/sheepzhao/WorkSpace/Boxify/internal/claw/chat/publisher.go)
7. [models.go](/Users/sheepzhao/WorkSpace/Boxify/internal/claw/chat/models.go)

## 一句话总结

`internal/claw/chat` 本质上是 Boxify 聊天能力的后端编排层：它把“本地会话/消息状态”、“OpenClaw channel HTTP/SSE 通信”、“前端实时事件广播”三件事连接在一起，并通过 `ChannelCoordinator` 对外提供统一入口。
