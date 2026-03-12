# Boxify Channel

`boxify-channel` 是一个 OpenClaw 原生 `channel` 插件，用于把 Boxify 作为本地聊天入口接入 OpenClaw。

## 设计定位

当前版本已经不是早期的桥接插件，而是标准 native channel：

- 使用 `registerChannel`
- channel id 为 `boxify`
- 由 `gateway.startAccount` 启动本地 HTTP inbox
- 入站消息通过 OpenClaw 标准 reply runtime 分发
- 不再调用 `openclaw agent --json`
- 不再回推旧版 `assistant_chunk/run_finished` 事件

插件内部当前采用同步 request/response 协议：

1. Boxify 向插件发送 `POST /channels/boxify/inbox`
2. 插件内部调用 `finalizeInboundContext + dispatchReplyWithBufferedBlockDispatcher`
3. 插件同步返回最终文本与 `sessionId`

## 请求与响应

监听基地址由 `channels.boxify.listenUrl` 决定，实际 inbox 路径固定为 `/channels/boxify/inbox`，默认基地址为 `http://127.0.0.1:32124`。

入站请求示例：

```json
{
  "conversationId": "conv_123",
  "messageId": "msg_456",
  "agentId": "main",
  "text": "你好",
  "metadata": {
    "source": "boxify"
  }
}
```

成功响应示例：

```json
{
  "ok": true,
  "conversationId": "conv_123",
  "sessionId": "agent:main:boxify-conv_123",
  "text": "你好，我是 OpenClaw。"
}
```

失败响应示例：

```json
{
  "ok": false,
  "conversationId": "conv_123",
  "sessionId": "agent:main:boxify-conv_123",
  "error": "错误信息"
}
```

## 安装与配置

安装：

```bash
openclaw plugins install /Users/sheepzhao/WorkSpace/Boxify/plugins/experimental/boxify-channel
```

`openclaw.plugin.json` 已声明 `channels: ["boxify"]`，`package.json` 已声明 `openclaw.channel` 元数据。

最小配置示例：

```json
{
  "channels": {
    "boxify": {
      "enabled": true,
      "listenUrl": "http://127.0.0.1:32124",
      "sharedToken": "your-token",
      "defaultAgent": "main"
    }
  }
}
```

字段说明：

- `enabled`: 是否启用该账号
- `listenUrl`: 插件监听基础地址，插件内部会拼接 `/channels/boxify/inbox`
- `sharedToken`: Boxify 请求插件时携带的共享令牌
- `defaultAgent`: 未显式传入 `agentId` 时使用的默认 agent

## 多账号

插件支持 `channels.boxify.accounts.<id>` 形式的账号覆盖配置。未指定账号时使用默认账号。

## 调研文档

原始调研记录见 [research-feishu-plugin.md](/Users/sheepzhao/WorkSpace/Boxify/plugins/experimental/boxify-channel/research-feishu-plugin.md)。

## 测试文档

手工联调步骤见 [TESTING.md](/Users/sheepzhao/WorkSpace/Boxify/plugins/experimental/boxify-channel/TESTING.md)。

## 后续增强

1. 改成 `registerHttpRoute`，避免插件自己额外监听端口
2. 接入 embedded runner / subscribe，拿到更细粒度的原生流式块
3. 补稳定的 `conversationId -> sessionId` 持久化映射
