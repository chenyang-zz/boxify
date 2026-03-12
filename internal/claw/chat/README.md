# claw/chat

`internal/claw/chat` 负责把 Boxify 聊天面板里的会话与消息，投递到 OpenClaw 的 `boxify` 原生 channel，并把同步返回结果回写到 Boxify 本地会话。

## 包内职责

- `service.go`: 聊天主流程编排入口。
- `channel_client.go`: 本地 HTTP channel 客户端，负责调用原生 inbox。
- `store.go`: 会话与消息存储抽象，当前默认实现为内存版。
- `publisher.go`: 向前端广播聊天事件。
- `models.go`: 会话、消息、请求、响应事件等模型定义。

## 启动时如何装配

`ClawService.rebuildManagers()` 会创建 `chat.Service`：

1. 创建 `MemoryConversationStore` 作为会话存储。
2. 创建 `HTTPChannelClient`，基地址为 `http://127.0.0.1:<pluginPort>`。
3. 创建 `WailsEventPublisher`，用于给前端发事件。
4. 注入 `clawprocess.Manager`，便于发消息前尝试拉起 OpenClaw。

其中：

- `pluginPort` 默认是 `32124`，可通过 `BOXIFY_PLUGIN_INBOX_PORT` 覆盖。
- `chatToken` 优先读取 `BOXIFY_CHAT_SHARED_TOKEN`，否则从 `dataDir/boxify.json` 读取；若不存在则首次生成并写回。
- 原生 channel inbox 固定路径是 `/channels/boxify/inbox`。

## 一次发送流程

### 1. 创建 Boxify 会话

前端先调用 `CreateConversation`。

- `store.CreateConversation()` 生成 Boxify 会话 ID，例如 `conv_xxx`。
- 会话会记录目标 `agentId`，默认是 `main`。
- 此时还没有 OpenClaw 的 `sessionId`。

### 2. 发送用户消息

前端调用 `SendMessage(conversationID, text)` 后，`chat.Service` 会按下面顺序处理：

1. 校验 `conversationID` 和消息内容非空。
2. 从 `ConversationStore` 读取会话，拿到绑定的 `agentId`。
3. 如果注入了 `clawprocess.Manager`，先尝试 `manager.Start()`。
4. 生成本次发送的 `runId` 和用户消息 `msgId`。
5. 先把一条 `role=user`、`status=done` 的消息写入本地 store。
6. 通过 `ChannelClient.SendMessage()` 发起 HTTP 请求。

## HTTP 投递细节

`HTTPChannelClient` 会向下面的地址发起请求：

```text
POST http://127.0.0.1:<pluginPort>/channels/boxify/inbox
```

请求头：

- `Content-Type: application/json`
- `X-Boxify-Token: <sharedToken>`，仅当 token 非空时携带

请求体对应 `BridgeInboxRequest`：

```json
{
  "conversationId": "conv_xxx",
  "messageId": "msg_xxx",
  "agentId": "main",
  "text": "你好",
  "metadata": {
    "source": "boxify"
  }
}
```

这里的 `conversationId` 是 Boxify 自己的会话 ID，不是 OpenClaw 的 session key。

## 插件返回后如何处理

插件当前走同步 request/response 模式，返回结构是 `BridgeInboxResponse`：

- `ok`
- `conversationId`
- `sessionId`
- `text`
- `error`

`chat.Service.SendMessage()` 收到响应后会继续做三件事：

1. 如果响应里有 `sessionId`，调用 `store.UpdateOpenClawSessionID()`，把 Boxify 会话映射到 OpenClaw 会话。
2. 如果响应里有 `text`，追加一条 `role=assistant`、`status=done` 的本地消息。
3. 如果配置了 `EventPublisher`，广播一个 `assistant_done` 事件给前端。

这意味着当前实现下，前端看到的是：

- 先出现一条本地 `user` 消息
- 等待插件同步返回
- 再出现一条整段 `assistant` 消息

不是 token 级流式写回。

## 存储模型

### Conversation

`Conversation` 维护 Boxify 侧会话元信息：

- `ID`: Boxify 会话 ID
- `AgentID`: 目标 agent
- `OpenClawSessionID`: 插件返回后的 OpenClaw 会话 ID
- `CreatedAt` / `UpdatedAt`

### Message

`Message` 维护会话消息：

- `Role`: `user` / `assistant` / `system`
- `RunID`: 一次发送链路的执行 ID
- `Status`: 当前主要使用 `done`

### 当前存储实现

当前默认使用 `MemoryConversationStore`：

- 只保存在进程内存中
- 重启 Boxify 后会话与消息丢失
- 适合原型验证和当前本地联调

不过接口已经预留了流式能力：

- `UpdateAssistantDraft()`
- `FinalizeAssistantMessage()`

当前同步模式没有真正使用这两条接口，但后续如果改成分块回写，可以直接复用。

## 前端事件

当助手回复成功落库后，`WailsEventPublisher` 会通过 Wails 广播事件：

- 事件名：`EventTypeClawChatEvent`
- 负载为 `ChatEvent`

当前实际发送的事件类型是：

- `assistant_done`

虽然事件常量名还保留了 `Bridge` 字样，但它现在服务的是原生 channel 聊天链路。

## 失败路径

### 本地校验失败

以下情况会直接返回错误，不会发 HTTP：

- `conversationID` 为空
- `text` 为空
- 会话不存在

### OpenClaw 拉起失败

`manager.Start()` 失败时只会记 `Warn` 日志，不会中断后续 HTTP 请求。这样可以兼容：

- OpenClaw 已经在外部启动
- 本次启动探测失败但插件仍可达

### 插件请求失败

以下情况会返回错误给上层：

- HTTP 建连失败
- 插件返回非 2xx
- 响应 JSON 解析失败
- 插件返回 `ok=false`

此时：

- 本地 `user` 消息已经写入
- 不会追加 `assistant` 消息

## 当前实现边界

- 当前只有“同步整段返回”，没有真正的流式 token/chunk 更新。
- 当前 `ChannelClient` 是本地 HTTP 实现，没有做重试、退避和断线恢复。
- 当前默认存储是内存版，没有做持久化。
- 当前模型名里仍保留 `BridgeInboxRequest` / `BridgeInboxResponse` / `BridgeEvent`，这是历史命名遗留，语义上已经对应原生 channel inbox。

## 读代码建议

如果要继续扩展这个包，建议按下面顺序看：

1. `service.go`
2. `channel_client.go`
3. `models.go`
4. `store.go`
5. `publisher.go`

这样能先理解主流程，再看数据结构与扩展点。
