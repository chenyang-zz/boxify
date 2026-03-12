# Boxify Channel Testing

本文档描述 `boxify-channel` 的本地联调方式。

## 前置条件

- 已安装 OpenClaw
- 已安装 Boxify
- `boxify-channel` 已安装到 OpenClaw
- OpenClaw 配置中存在可用的 `main` agent，或你计划测试的其他 agent

安装插件：

```bash
openclaw plugins install /Users/sheepzhao/WorkSpace/Boxify/plugins/experimental/boxify-channel
```

## 1. 获取 Boxify 提供的连接信息

在 Boxify 中调用 `GetChatChannelInfo`，记录以下字段：

- `channelInboxURL`
- `sharedToken`

当前原生版只需要这两个字段。

## 2. 配置 OpenClaw

将 OpenClaw 配置中的 `channels.boxify` 设置为：

```json
{
  "channels": {
    "boxify": {
      "enabled": true,
      "listenUrl": "http://127.0.0.1:32124",
      "sharedToken": "替换为 Boxify 返回的 sharedToken",
      "defaultAgent": "main"
    }
  }
}
```

要求：

- `listenUrl` 与 Boxify 的 `channelInboxURL` 基础地址一致
- `sharedToken` 与 Boxify 返回值一致
- `defaultAgent` 必须是 OpenClaw 中真实存在的 agent

## 3. 启动顺序

建议顺序：

1. 启动 Boxify
2. 启动 OpenClaw
3. 确认 OpenClaw 已加载 `boxify` channel

## 4. 基础收发验证

在 Boxify 聊天面板创建一个会话并发送一条消息，例如：

```text
你好，请用一句话介绍自己
```

期望结果：

- Boxify 本地先写入一条 `user` 消息
- 插件收到 `POST /channels/boxify/inbox`
- OpenClaw 通过 `boxify` native channel 处理该消息
- Boxify 收到同步响应并写入一条 `assistant` 消息

当前版本不是事件流式写回，而是同步整段返回。

## 5. HTTP 手工验证

如果需要绕过 Boxify UI，可以直接请求插件 inbox：

```bash
curl -X POST http://127.0.0.1:32124/channels/boxify/inbox \
  -H 'Content-Type: application/json' \
  -H 'X-Boxify-Token: your-token' \
  -d '{
    "conversationId": "manual-test-001",
    "messageId": "msg-001",
    "agentId": "main",
    "text": "你好，请回复测试成功",
    "metadata": {
      "source": "manual-test"
    }
  }'
```

期望返回：

```json
{
  "ok": true,
  "conversationId": "manual-test-001",
  "sessionId": "agent:main:boxify-manual-test-001",
  "text": "..."
}
```

## 6. 失败场景验证

可以重点验证以下场景：

- `X-Boxify-Token` 错误
  期望：返回 `401`
- `agentId` 不存在
  期望：返回 `ok=false` 或 HTTP 500，响应体包含错误信息
- OpenClaw 未启动
  期望：Boxify 发送失败，界面不追加 assistant 消息
- `listenUrl` 配置错误
  期望：Boxify 请求插件失败

## 7. 观察点

建议关注：

- Boxify 日志中是否出现“请求插件 inbox 失败”
- OpenClaw 日志中是否出现 `[boxify-channel] native inbox listening`
- 返回的 `sessionId` 是否稳定
- 同一 `conversationId` 的多轮消息是否复用同一 session key

## 8. 已知限制

- 当前返回的是整段文本，不是 token 级流式输出
- 当前 inbox 由插件自己监听本地端口，还没有切到 `registerHttpRoute`
- 当前 session key 为 `agent:<agentId>:boxify-<conversationId>`，还没有做额外持久化映射
