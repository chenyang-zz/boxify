import { randomUUID } from "node:crypto";

/**
 * 将 Boxify 会话 ID 转成 OpenClaw 侧更稳定的展示/路由别名。
 * 例如 conv_123 -> boxify_session_123。
 */
function toBoxifySessionAlias(conversationId) {
  const safe = String(conversationId || "")
    .trim()
    .replace(/^conv[-_]+/i, "")
    .replace(/[^a-zA-Z0-9._-]+/g, "_")
    .replace(/^_+|_+$/g, "");
  return safe ? `boxify_session_${safe}` : `boxify_session_${randomUUID()}`;
}

/**
 * 将会话别名规整为安全的 session id。
 */
function toSessionId(conversationId) {
  return toBoxifySessionAlias(conversationId);
}

/**
 * 生成符合 OpenClaw native channel 习惯的 session key。
 */
function toSessionKey(agentId, conversationId) {
  const resolvedAgentId = String(agentId || "main").trim() || "main";
  return `agent:${resolvedAgentId}:${toSessionId(conversationId)}`;
}

/**
 * 读取 HTTP 请求体。
 */
export function readJSONBody(req) {
  return new Promise((resolve, reject) => {
    let raw = "";
    req.on("data", (chunk) => {
      raw += String(chunk);
    });
    req.on("end", () => {
      try {
        resolve(raw ? JSON.parse(raw) : {});
      } catch (error) {
        reject(error);
      }
    });
    req.on("error", reject);
  });
}

/**
 * 等待 abort，供 gateway 生命周期阻塞使用。
 */
export function waitForAbort(signal) {
  if (!signal) {
    return new Promise(() => {});
  }
  if (signal.aborted) {
    return Promise.resolve();
  }
  return new Promise((resolve) => {
    signal.addEventListener("abort", resolve, { once: true });
  });
}

/**
 * 为 SSE 连接写入一个事件。
 */
export function writeSSE(res, event, data) {
  res.write(`event: ${event}\n`);
  res.write(`data: ${JSON.stringify(data)}\n\n`);
}

/**
 * 将 cumulative/fragmented 文本规整为真正的增量片段。
 */
export function computeDeltaText(currentText, nextText) {
  if (!nextText) {
    return { delta: "", merged: currentText };
  }
  if (!currentText) {
    return { delta: nextText, merged: nextText };
  }
  if (nextText.startsWith(currentText)) {
    return {
      delta: nextText.slice(currentText.length),
      merged: nextText,
    };
  }
  if (currentText.endsWith(nextText)) {
    return { delta: "", merged: currentText };
  }
  return {
    delta: nextText,
    merged: `${currentText}${nextText}`,
  };
}

/**
 * 构造 Boxify 入站消息上下文。
 */
export function buildInboundContext(runtime, account, incoming) {
  const finalizeInboundContext = runtime?.channel?.reply?.finalizeInboundContext;
  if (typeof finalizeInboundContext !== "function") {
    throw new Error("OpenClaw reply runtime 不可用");
  }

  const conversationId = String(incoming?.conversationId || "").trim();
  const messageId = String(incoming?.messageId || "").trim();
  const runId = String(incoming?.runId || "").trim();
  const text = String(incoming?.text || "").trim();
  const agentId = String(incoming?.agentId || account.defaultAgent || "").trim();
  const senderId = String(incoming?.senderId || incoming?.metadata?.senderId || conversationId).trim();
  const chatType = incoming?.chatType === "group" ? "group" : "direct";
  const sessionId = toSessionKey(agentId, conversationId);
  const senderName = String(incoming?.metadata?.senderName || "Boxify User").trim();
  const conversationLabel = String(incoming?.metadata?.conversationLabel || toBoxifySessionAlias(conversationId)).trim();
  if (!conversationId || !text) {
    throw new Error("conversationId 和 text 不能为空");
  }

  return {
    cfgPromise: runtime.config.loadConfig(),
    conversationId,
    runId,
    sessionId,
    msgCtx: finalizeInboundContext({
      Body: text,
      RawBody: text,
      CommandBody: text,
      From: `boxify:${senderId || conversationId}`,
      To: `boxify:${conversationId}`,
      SessionKey: sessionId,
      AccountId: account.accountId,
      OriginatingChannel: "boxify",
      OriginatingTo: `boxify:${conversationId}`,
      ChatType: chatType,
      SenderName: senderName || undefined,
      SenderId: senderId || conversationId,
      Provider: "boxify",
      Surface: "boxify",
      ConversationLabel: conversationLabel || toBoxifySessionAlias(conversationId),
      MessageSid: messageId || undefined,
      Timestamp: Date.now(),
      CommandAuthorized: true,
      metadata: incoming?.metadata,
    }),
  };
}
