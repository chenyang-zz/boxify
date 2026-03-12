# OpenClaw 飞书插件调研整理

本文基于本机已安装的 OpenClaw `@openclaw/feishu` 插件源码整理，目标是为 `boxify-channel` 后续演进提供可继续深入的阅读入口。

## 结论速览

当前 `boxify-channel` 已经迁移为 OpenClaw 原生 channel 插件，但与官方飞书插件相比，能力深度仍有明显差距：

- `boxify-channel` 当前是原生 channel 的最小实现
  - 通过 `registerChannel`
  - 自己监听本地 HTTP inbox
  - 通过 `finalizeInboundContext + dispatchReplyWithBufferedBlockDispatcher` 接入 reply runtime
  - 目前仍是同步整段返回，不是细粒度流式 dispatcher 适配
- 飞书插件是原生 channel 插件
  - 通过 `registerChannel`
  - 直接接入 OpenClaw channel runtime
  - 自己负责账号管理、消息接入、鉴权、会话路由、回复分发、流式卡片、媒体和平台工具

相关引用：

- [plugins/experimental/boxify-channel/index.js](/Users/sheepzhao/WorkSpace/Boxify/plugins/experimental/boxify-channel/index.js)
- [plugins/experimental/boxify-channel/README.md](/Users/sheepzhao/WorkSpace/Boxify/plugins/experimental/boxify-channel/README.md)
- [extensions/feishu/index.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/index.ts)

## 1. 飞书插件的注册方式

飞书插件入口在 `extensions/feishu/index.ts`，关键逻辑是：

- `api.registerChannel({ plugin: feishuPlugin })`
- 同时注册了 doc/chat/wiki/drive/perm/bitable 等飞书平台工具

这说明飞书不是旁路服务，而是 OpenClaw 标准频道的一种实现。

相关引用：

- [extensions/feishu/index.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/index.ts)
- [plugin-sdk/types.d.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/dist/plugin-sdk/plugins/types.d.ts)

补充：

- `package.json` 中的 `openclaw.channel` 也声明了 `id/label/docsPath/aliases`
- `openclaw.plugin.json` 中声明了 `channels: ["feishu"]`

相关引用：

- [extensions/feishu/package.json](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/package.json)
- [extensions/feishu/openclaw.plugin.json](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/openclaw.plugin.json)

## 2. 飞书频道声明了哪些能力

飞书频道主体定义在 `extensions/feishu/src/channel.ts`。

可以直接看到这不是一个最小实现，而是完整接入了 OpenClaw channel 能力：

- `capabilities`
  - `threads: true`
  - `media: true`
  - `reactions: true`
  - `edit: true`
  - `reply: true`
- `pairing`
  - 未授权私聊用户可以走 pairing
- `groups`
  - 群工具策略和群权限控制
- `directory`
  - 查询用户和群
- `messaging`
  - 统一目标 ID 规范化
- `outbound`
  - 使用飞书专属发送适配器
- `onboarding`
  - 安装配置向导

相关引用：

- [extensions/feishu/src/channel.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/channel.ts)

## 3. 账号与配置解析方式

飞书账号解析在 `extensions/feishu/src/accounts.ts`。

核心做法：

- 允许顶层 `channels.feishu` 作为默认配置
- 允许 `channels.feishu.accounts.<id>` 做账号级覆盖
- 解析时会把顶层配置和账号配置 merge
- 最后得到 `ResolvedFeishuAccount`
  - `accountId`
  - `enabled`
  - `configured`
  - `appId`
  - `appSecret`
  - `verificationToken`
  - `encryptKey`
  - `domain`
  - `config`

这意味着飞书插件本身就支持多账号，并且配置继承是内建能力，不是外围面板自己拼出来的。

相关引用：

- [extensions/feishu/src/accounts.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/accounts.ts)
- [extensions/feishu/src/types.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/types.ts)
- [docs/channels/feishu.md](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/docs/channels/feishu.md)

额外确认到的配置项：

- `connectionMode`: `websocket` 或 `webhook`
- `replyInThread`
- `streaming`
- `typingIndicator`
- `resolveSenderNames`
- `dynamicAgentCreation`
- `groupSessionScope`
- `topicSessionMode`

相关引用：

- [extensions/feishu/src/channel.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/channel.ts)

## 4. 入站消息链路

飞书入站入口主要分三层：

### 4.1 monitor 启动账号监听

`monitorFeishuProvider` 会：

- 找到所有已启用账号
- 逐个做启动前探测
- 为每个账号启动独立监听

相关引用：

- [extensions/feishu/src/monitor.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/monitor.ts)

### 4.2 transport 层决定用 websocket 还是 webhook

`monitor.transport.ts` 中：

- `monitorWebSocket`
  - 创建飞书 WS client
  - 通过长连接接收事件
- `monitorWebhook`
  - 本地起 HTTP server
  - 安装 body limit、限流、基础请求校验
  - 交给飞书 SDK 的 webhook adapter 处理

这说明飞书插件的接入是实时事件驱动，而不是简单的同步 inbox 请求模型。

相关引用：

- [extensions/feishu/src/monitor.transport.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/monitor.transport.ts)
- [extensions/feishu/src/client.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/client.ts)

### 4.3 bot 层把平台事件转成 agent 输入

`handleFeishuMessage` 是真正的核心入口。它会：

- 对消息做内存 + 持久化双重去重
- 解析消息文本、mention、reply、thread、media
- 检查群策略和私聊策略
- 按规则构造 OpenClaw 入站上下文
- 调用 channel runtime 分发给目标 agent session

相关引用：

- [extensions/feishu/src/bot.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/bot.ts)

## 5. 飞书消息进入 OpenClaw 后如何做会话路由

这是飞书插件最值得参考的一层。

### 5.1 不是简单按 chatId 建会话

`resolveFeishuGroupSession` 会根据配置选择不同粒度的群会话范围：

- `group`
- `group_sender`
- `group_topic`
- `group_topic_sender`

因此同一个群可以：

- 全群共享一个 session
- 每个发言者独立 session
- 每个话题独立 session
- 每个话题下每个发言者独立 session

相关引用：

- [extensions/feishu/src/bot.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/bot.ts)

### 5.2 群聊有 mention gate 和 allowlist gate

在群聊里，插件会判断：

- 群是否启用
- `groupPolicy`
- `groupAllowFrom`
- 每群/全局 sender allowlist
- 是否要求 `mention bot`

如果未 mention 且配置要求 mention，会把消息记入 pending history，但不会触发 agent。

相关引用：

- [extensions/feishu/src/policy.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/policy.ts)
- [extensions/feishu/src/bot.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/bot.ts)

### 5.3 私聊有 pairing / allowlist 机制

私聊策略基于 `dmPolicy`：

- `open`
- `pairing`
- `allowlist`

未授权时如果是 `pairing`，插件会自动发送配对提示消息，而不是直接进入 agent。

相关引用：

- [extensions/feishu/src/channel.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/channel.ts)
- [extensions/feishu/src/bot.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/bot.ts)

### 5.4 最终通过 channel runtime 分发

`handleFeishuMessage` 最终会调用：

- `core.channel.routing.resolveAgentRoute(...)`
- `core.channel.reply.dispatchReplyFromConfig(...)`

这说明飞书插件直接接在 OpenClaw 运行时的 routing/reply 抽象上。

相关引用：

- [extensions/feishu/src/bot.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/bot.ts)

## 6. 出站回复是怎么实现的

### 6.1 普通出站适配器

`extensions/feishu/src/outbound.ts` 提供了飞书出站适配器：

- 文本走 `sendMessageFeishu`
- markdown/表格/代码块可自动切换卡片
- 支持媒体发送
- 支持本地图片路径自动识别上传

相关引用：

- [extensions/feishu/src/outbound.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/outbound.ts)
- [extensions/feishu/src/send.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/send.ts)

### 6.2 reply dispatcher 是关键

飞书真正和 agent 输出做对接的是 `createFeishuReplyDispatcher`。

它负责：

- 接住 OpenClaw 的 reply payload
- 决定文本还是卡片
- 决定是否启用 streaming card
- 决定是否在 thread 里回复
- 处理 typing indicator
- 处理媒体输出

相关引用：

- [extensions/feishu/src/reply-dispatcher.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/reply-dispatcher.ts)

### 6.3 流式不是“多条消息”，而是“更新同一张卡片”

`reply-dispatcher.ts` 中有 `FeishuStreamingSession`，配合 partial/final 回调逐步更新卡片内容。

这和当前 `boxify-channel` 的“同步整段返回”是两个层级：

- `boxify-channel` 当前只消费最终文本结果
- 飞书插件已经接入更完整的 reply dispatcher，可消费 agent 的中间输出

相关引用：

- [extensions/feishu/src/reply-dispatcher.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/reply-dispatcher.ts)

### 6.4 typing indicator 通过 reaction 实现

飞书插件不是模拟“正在输入”文本，而是给原消息加 `Typing` reaction，并在结束时移除。

相关引用：

- [extensions/feishu/src/typing.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/typing.ts)

## 7. 卡片交互如何回流到 agent

`handleFeishuCardAction` 会把卡片按钮点击转成一个 synthetic message event，然后再次走 `handleFeishuMessage`。

这意味着：

- 卡片动作不会走单独的旁路逻辑
- 它仍然复用同一套鉴权、会话、reply 分发链路

相关引用：

- [extensions/feishu/src/card-action.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/card-action.ts)
- [extensions/feishu/src/bot.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/bot.ts)

## 8. Dynamic agent creation 是怎么做的

飞书插件支持在私聊第一次接触某个用户时，动态创建一个 agent：

- agent id 形如 `feishu-<senderOpenId>`
- 自动创建 workspace 和 agentDir
- 自动写入 `agents.list`
- 自动写入 `bindings`
- 然后重新按新的 binding 路由消息

所以它不是“把用户映射到 session”，而是可以进一步“把用户映射到独立 agent”。

相关引用：

- [extensions/feishu/src/dynamic-agent.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/dynamic-agent.ts)
- [extensions/feishu/src/bot.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/bot.ts)

## 9. 对 `boxify-channel` 的直接启发

`boxify-channel` 已经完成了“从桥接式插件转到原生 channel 插件”这一步。

当前更值得参考飞书插件的地方，不再是“是否改成 `registerChannel`”，而是后续怎么继续补齐 native channel 能力：

- 把 Boxify 会话进一步抽象成更稳定的 OpenClaw `peer/session` 模型
- 从“同步整段返回”升级到更完整的 reply dispatcher 输出消费
- 根据需要补 `streaming / threading / directory / actions` 等 channel 能力
- 继续减少 Boxify 私有协议，让 Boxify 更像一个标准 channel 宿主 UI

当前 `boxify-channel` 的实现特征：

- 起本地 HTTP inbox
- 收到消息后构造标准 inbound context
- 通过 runtime dispatcher 获取最终回复
- 以同步 JSON 响应返回 `text/sessionId`

相关引用：

- [plugins/experimental/boxify-channel/index.js](/Users/sheepzhao/WorkSpace/Boxify/plugins/experimental/boxify-channel/index.js)
- [plugins/experimental/boxify-channel/README.md](/Users/sheepzhao/WorkSpace/Boxify/plugins/experimental/boxify-channel/README.md)

## 10. 建议的继续阅读顺序

如果后面要继续理解并迁移设计，建议按下面顺序读源码：

1. 先读频道声明和配置结构
   - [extensions/feishu/src/channel.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/channel.ts)
   - [extensions/feishu/src/accounts.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/accounts.ts)
2. 再读入站链路
   - [extensions/feishu/src/monitor.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/monitor.ts)
   - [extensions/feishu/src/monitor.transport.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/monitor.transport.ts)
   - [extensions/feishu/src/bot.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/bot.ts)
3. 最后读出站与流式回复
   - [extensions/feishu/src/outbound.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/outbound.ts)
   - [extensions/feishu/src/send.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/send.ts)
   - [extensions/feishu/src/reply-dispatcher.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/reply-dispatcher.ts)

## 11. 当前可确认的判断边界

目前能从本机源码直接确认的事实：

- 飞书插件是 `registerChannel` 原生实现
- 飞书插件没有走同步 inbox + 最终文本返回这种最小实现
- 飞书插件已经接入 OpenClaw 的 routing / reply / typing / streaming 体系
- 飞书插件的流式回复依赖 reply dispatcher，而不是单独开一条桥接协议

目前还没继续往下展开的点：

- OpenClaw channel runtime 内部 `dispatchReplyFromConfig` 的具体调用栈
- `FeishuStreamingSession` 的卡片增量更新细节
- 其他 channel 插件与飞书插件的抽象共性边界

这些适合作为下一轮继续读的目标。

## 12. OpenClaw channel runtime 的关键抽象

继续往下追后，可以确认飞书插件依赖的是 OpenClaw runtime 暴露给 channel 的几组核心能力：

- `routing.resolveAgentRoute`
- `reply.finalizeInboundContext`
- `reply.dispatchReplyFromConfig`
- `reply.withReplyDispatcher`
- `reply.createReplyDispatcherWithTyping`

这组能力在插件 runtime 类型里是正式导出的，不是飞书插件私有调用。

相关引用：

- [plugin-sdk runtime types-channel.d.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/dist/plugin-sdk/plugins/runtime/types-channel.d.ts)
- [extensions/feishu/src/bot.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/bot.ts)
- [extensions/feishu/src/reply-dispatcher.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/reply-dispatcher.ts)

### 12.1 `finalizeInboundContext` 做什么

`finalizeInboundContext` 的职责不是路由，而是把 channel 构造出来的入站上下文做最后规范化。

它会处理：

- `Body / RawBody / CommandBody / Transcript` 的换行和系统标签清洗
- `BodyForAgent / BodyForCommands` 的兜底填充
- `ConversationLabel` 的自动生成
- `ChatType` 的标准化
- `CommandAuthorized` 的布尔归一
- 多媒体字段和 `MediaTypes` 的补全

也就是说，飞书插件在 `bot.ts` 里先尽量完整地组装上下文，然后再交给 runtime 做统一收口。

相关引用：

- [plugin-sdk/inbound-context-BFKvjYKo.js](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/dist/plugin-sdk/inbound-context-BFKvjYKo.js)

### 12.2 `resolveAgentRoute` 怎么决定消息去哪

`resolveAgentRoute` 会把下面这些输入一起考虑：

- `channel`
- `accountId`
- `peer`
- `parentPeer`
- `guildId`
- `teamId`
- `memberRoleIds`
- `cfg.bindings`

它不是只看一个 peer key，而是分层匹配：

- `binding.peer`
- `binding.peer.parent`
- `binding.guild+roles`
- `binding.guild`
- `binding.team`
- `binding.account`
- `binding.channel`
- 最后回退 `default`

匹配成功后，runtime 会生成：

- `agentId`
- `sessionKey`
- `mainSessionKey`
- `matchedBy`

这解释了为什么飞书插件只需要产出正确的 `peer / parentPeer / accountId`，后面的 session key 不必自己手工拼整套逻辑。

相关引用：

- [plugin-sdk/resolve-route-OwCmBiZ2.js](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/dist/plugin-sdk/resolve-route-OwCmBiZ2.js)
- [extensions/feishu/src/bot.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/bot.ts)

## 13. `dispatchReplyFromConfig` 的真实职责

这部分已经可以从 runtime bundle 里直接确认。

### 13.1 它不是“发消息 API”，而是 reply orchestration

`dispatchReplyFromConfig` 在 runtime 中承担的是调度角色。它会：

- 检查 inbound dedupe
- 记录 session/message diagnostics
- 触发 `message_received` hook
- 根据策略判断是否允许发送
- 调用 `getReplyFromConfig(...)` 真正取 agent 回复
- 接住中间的 `onToolResult / onBlockReply`
- 把最终回复通过 dispatcher 发给具体 channel

也就是说：

- “生成回复”与“投递回复”在 runtime 里是分开的
- channel 插件只要提供 dispatcher，就能消费同一套 agent 输出

相关引用：

- [plugin-sdk/reply-DFFRlayb.js](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/dist/plugin-sdk/reply-DFFRlayb.js)
- [plugin-sdk/auto-reply/reply/dispatch-from-config.d.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/dist/plugin-sdk/auto-reply/reply/dispatch-from-config.d.ts)

### 13.2 dispatcher 如何接住不同阶段的输出

runtime 会把不同类型回复拆成三种：

- `tool`
- `block`
- `final`

对应 dispatcher 接口：

- `sendToolResult`
- `sendBlockReply`
- `sendFinalReply`

飞书插件的 `createFeishuReplyDispatcher` 正是围绕这三种输出做适配。

相关引用：

- [plugin-sdk/auto-reply/reply/reply-dispatcher.d.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/dist/plugin-sdk/auto-reply/reply/reply-dispatcher.d.ts)
- [extensions/feishu/src/reply-dispatcher.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/reply-dispatcher.ts)

### 13.3 `withReplyDispatcher` 解决的是“何时算发送完成”

`withReplyDispatcher` 的逻辑非常直接：

- 执行 `run()`
- 无论成功失败都调用 `dispatcher.markComplete()`
- 等待 `dispatcher.waitForIdle()`
- 最后再跑 `onSettled`

这意味着 channel 插件不需要自己管理复杂的“还有没有 pending reply”，runtime 已经有统一收尾。

相关引用：

- [plugin-sdk/auto-reply/dispatch.d.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/dist/plugin-sdk/auto-reply/dispatch.d.ts)
- [plugin-sdk/reply-DFFRlayb.js](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/dist/plugin-sdk/reply-DFFRlayb.js)

## 14. 飞书流式卡片的内部细节

继续展开后，可以确认飞书所谓的 streaming 不是多条消息刷屏，而是基于 Card Kit 的“同卡更新”。

### 14.1 流式卡片对象怎么启动

`FeishuStreamingSession.start(...)` 会：

1. 先拿 tenant access token
2. 调 `cardkit/v1/cards` 创建一个 `streaming_mode: true` 的卡片实体
3. 再把这张卡片作为 `interactive` message 发到飞书会话里
4. 保存 `cardId / messageId / sequence / currentText`

这说明飞书流式回复本质上是：

- 先创建卡片资源
- 再创建承载该卡片的消息

相关引用：

- [extensions/feishu/src/streaming-card.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/streaming-card.ts)

### 14.2 增量更新如何避免乱序和刷频

`FeishuStreamingSession` 内部做了几层控制：

- `mergeStreamingText`
  - 尝试把 cumulative partial 和 fragmented partial 合并
- `queue`
  - 所有更新串行排队，避免并发 PUT
- `sequence`
  - 每次更新卡片元素时自增，保证服务端顺序
- `updateThrottleMs = 100`
  - 最多约 10 次/秒，避免过频更新
- `pendingText`
  - 节流期间先记住新文本，下一次再合并

所以它不是“partial 来一段就立即写一次”，而是带顺序控制和节流。

相关引用：

- [extensions/feishu/src/streaming-card.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/streaming-card.ts)

### 14.3 关闭流式时会切回普通卡片

`close(finalText)` 会：

- 先把 pending 内容和 final 内容合并
- 如果最终文本和当前卡片文本不同，补最后一次更新
- 再 PATCH 卡片 settings，把 `streaming_mode` 改成 `false`
- 同时写入 `summary`

所以飞书流式卡片的结束不是“再发一条 final 消息”，而是“收敛并封口当前卡片”。

相关引用：

- [extensions/feishu/src/streaming-card.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/streaming-card.ts)

### 14.4 reply dispatcher 如何把 partial 接进卡片流

飞书的 `createFeishuReplyDispatcher` 做了这一层连接：

- 当 `streamingEnabled` 且 renderMode 允许时，`onReplyStart` 会启动 streaming session
- `replyOptions.onPartialReply` 会把 partial 文本喂给 `queueStreamingUpdate`
- `deliver(..., info.kind === "block")` 时也会兼容某些 runtime 只给 block、不走 partial 回调的情况
- `info.kind === "final"` 时会关闭 streaming session

这说明飞书插件对 runtime 的 partial/block/final 三种输出都做了兼容，而不是只依赖一种回调。

相关引用：

- [extensions/feishu/src/reply-dispatcher.ts](/Users/sheepzhao/.local/share/fnm/node-versions/v22.21.1/installation/lib/node_modules/openclaw/extensions/feishu/src/reply-dispatcher.ts)

## 15. 对 Boxify 下一步设计更明确的启发

到这一层已经可以更具体地判断：

`boxify-channel` 现在已经是原生 channel 模式，所以下一步更实际的判断是：

- 如果继续维持当前最小 native channel 方案
  - 实现简单
  - Boxify 只需要同步请求/响应
  - 但仍然拿不到完整的 tool/block/final/partial 原生输出体验
- 如果继续向飞书插件靠拢
  - 需要把 Boxify 会话进一步抽象成更稳定的 `peer/session`
  - 需要给 Boxify 实现更完整的 reply dispatcher 适配
  - 需要决定哪些 channel 能力值得补到 Boxify：例如 `streaming / threading / actions`
  - 这样才能更完整地接住 OpenClaw runtime 的统一回复链

换句话说，飞书插件最值得学习的不是飞书 API 细节，而是它如何把“平台消息收发”接到 OpenClaw channel runtime 上。
