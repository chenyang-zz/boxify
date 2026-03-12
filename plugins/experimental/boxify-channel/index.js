import http from "node:http";
import { randomUUID } from "node:crypto";
import {
  DEFAULT_ACCOUNT_ID,
  collectStatusIssuesFromLastError,
  createDefaultChannelRuntimeState,
  emptyPluginConfigSchema,
} from "openclaw/plugin-sdk";

let pluginRuntime = null;

const activeServers = new Map();
const NATIVE_INBOX_PATH = "/channels/boxify/inbox";
const DEFAULT_LISTEN_URL = "http://127.0.0.1:32124";

/**
 * 返回原生 channel 配置，并兼容单账号/多账号结构。
 */
function resolveChannelConfig(cfg, accountId) {
  const root = cfg?.channels?.boxify && typeof cfg.channels.boxify === "object" ? cfg.channels.boxify : {};
  const accounts = root.accounts && typeof root.accounts === "object" ? root.accounts : {};
  const resolvedAccountId = normalizeAccountId(accountId);
  const accountOverrides =
    resolvedAccountId !== DEFAULT_ACCOUNT_ID &&
    accounts[resolvedAccountId] &&
    typeof accounts[resolvedAccountId] === "object"
      ? accounts[resolvedAccountId]
      : {};
  const merged = { ...root, ...accountOverrides };
  const rawListenURL = stringOrDefault(merged.listenUrl, DEFAULT_LISTEN_URL);

  return {
    accountId: resolvedAccountId,
    name: stringOrDefault(merged.name, resolvedAccountId === DEFAULT_ACCOUNT_ID ? "Boxify" : resolvedAccountId),
    enabled: merged.enabled !== false,
    listenUrl: normalizeListenURL(rawListenURL),
    sharedToken: stringOrDefault(merged.sharedToken, ""),
    defaultAgent: stringOrDefault(merged.defaultAgent, "main"),
    configured: Boolean(rawListenURL.trim()),
  };
}

/**
 * 规整账号 ID，保证默认账号语义稳定。
 */
function normalizeAccountId(accountId) {
  const trimmed = String(accountId || "").trim();
  return trimmed || DEFAULT_ACCOUNT_ID;
}

/**
 * 读取字符串配置并提供默认值。
 */
function stringOrDefault(value, fallback) {
  const trimmed = String(value ?? "").trim();
  return trimmed || fallback;
}

/**
 * 规整监听地址，仅保留基础地址。
 */
function normalizeListenURL(listenURL) {
  const trimmed = String(listenURL || "").trim();
  if (!trimmed) {
    return DEFAULT_LISTEN_URL;
  }
  return trimmed.replace(/\/$/, "");
}

/**
 * 将会话 ID 规整为安全格式，保持与旧桥接版兼容。
 */
function toSessionId(conversationId) {
  const safe = String(conversationId || "")
    .trim()
    .replace(/[^a-zA-Z0-9._-]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return safe ? `boxify-${safe}` : `boxify-${randomUUID()}`;
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
function readJSONBody(req) {
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
function waitForAbort(signal) {
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
 * 关闭账号对应的 HTTP server。
 */
async function closeActiveServer(accountId) {
  const server = activeServers.get(accountId);
  if (!server) {
    return;
  }

  activeServers.delete(accountId);
  await new Promise((resolve) => {
    server.close(() => resolve());
  });
}

/**
 * 处理一条 Boxify 入站消息，并交给原生 channel dispatcher。
 */
async function handleInboxRequest(account, incoming, logger) {
  const runtime = pluginRuntime;
  const finalizeInboundContext = runtime?.channel?.reply?.finalizeInboundContext;
  const dispatchReply = runtime?.channel?.reply?.dispatchReplyWithBufferedBlockDispatcher;
  if (typeof finalizeInboundContext !== "function" || typeof dispatchReply !== "function") {
    throw new Error("OpenClaw reply runtime 不可用");
  }

  const conversationId = String(incoming?.conversationId || "").trim();
  const messageId = String(incoming?.messageId || "").trim();
  const text = String(incoming?.text || "").trim();
  const agentId = String(incoming?.agentId || account.defaultAgent || "").trim();
  const senderId = String(incoming?.senderId || incoming?.metadata?.senderId || conversationId).trim();
  const chatType = incoming?.chatType === "group" ? "group" : "direct";
  const sessionId = toSessionKey(agentId, conversationId);
  const senderName = String(incoming?.metadata?.senderName || "Boxify User").trim();
  const conversationLabel = String(incoming?.metadata?.conversationLabel || conversationId).trim();
  const replyChunks = [];

  if (!conversationId || !text) {
    throw new Error("conversationId 和 text 不能为空");
  }

  try {
    const cfg = await runtime.config.loadConfig();
    const msgCtx = finalizeInboundContext({
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
      ConversationLabel: conversationLabel || conversationId,
      MessageSid: messageId || undefined,
      Timestamp: Date.now(),
      CommandAuthorized: true,
      metadata: incoming?.metadata,
    });

    await dispatchReply({
      ctx: msgCtx,
      cfg,
      dispatcherOptions: {
        deliver: async (payload) => {
          const chunk = String(payload?.text ?? payload?.body ?? "").trim();
          if (chunk) {
            replyChunks.push(chunk);
          }
        },
        onReplyStart: () => {
          logger?.info?.(`[boxify-channel] 原生回复开始，conversation=${conversationId}`);
        },
      },
    });

    return {
      ok: true,
      conversationId,
      sessionId,
      text: replyChunks.join("\n\n"),
    };
  } catch (error) {
    const message = String(error?.message || error || "OpenClaw 执行失败");
    logger?.error?.(`[boxify-channel] 原生 channel 处理失败: ${message}`);
    return {
      ok: false,
      conversationId,
      sessionId,
      error: message,
      text: "",
    };
  }
}

/**
 * 启动 native channel 的本地 inbox。
 */
async function startInboxServer(ctx) {
  const account = resolveChannelConfig(ctx.cfg, ctx.accountId);
  if (!account.enabled) {
    ctx.log?.info?.(`[boxify-channel] 账号 ${account.accountId} 已禁用，跳过启动`);
    return;
  }
  if (!account.configured) {
    throw new Error("channels.boxify.listenUrl 未配置");
  }

  await closeActiveServer(account.accountId);

  const listenURL = new URL(account.listenUrl);
  const inboxURL = `${account.listenUrl}${NATIVE_INBOX_PATH}`;
  const server = http.createServer(async (req, res) => {
    if (req.method !== "POST" || req.url !== NATIVE_INBOX_PATH) {
      res.statusCode = 404;
      res.end("not found");
      return;
    }

    if (account.sharedToken && req.headers["x-boxify-token"] !== account.sharedToken) {
      res.statusCode = 401;
      res.end("invalid token");
      return;
    }

    try {
      const payload = await readJSONBody(req);
      const result = await handleInboxRequest(account, payload, ctx.log);
      res.setHeader("Content-Type", "application/json");
      res.end(JSON.stringify(result));
    } catch (error) {
      const message = String(error?.message || error || "unknown error");
      ctx.log?.error?.(`[boxify-channel] inbox 处理失败: ${message}`);
      res.statusCode = 500;
      res.setHeader("Content-Type", "application/json");
      res.end(JSON.stringify({ ok: false, error: message }));
    }
  });

  await new Promise((resolve, reject) => {
    server.once("error", reject);
    server.listen(Number(listenURL.port || 80), listenURL.hostname, () => {
      server.off("error", reject);
      resolve();
    });
  });

  activeServers.set(account.accountId, server);
  ctx.setStatus?.({
    accountId: account.accountId,
    running: true,
    lastStartAt: Date.now(),
    lastError: null,
  });
  ctx.log?.info?.(`[boxify-channel] native inbox listening on ${inboxURL}`);

  try {
    await waitForAbort(ctx.abortSignal);
  } finally {
    await closeActiveServer(account.accountId);
    ctx.setStatus?.({
      accountId: account.accountId,
      running: false,
      lastStopAt: Date.now(),
    });
    ctx.log?.info?.(`[boxify-channel] native inbox stopped for ${account.accountId}`);
  }
}

const boxifyNativeChannel = {
  id: "boxify",
  meta: {
    id: "boxify",
    label: "Boxify",
    selectionLabel: "Boxify",
    docsPath: "/channels/boxify",
    docsLabel: "boxify",
    blurb: "Boxify native local channel for OpenClaw.",
    quickstartAllowFrom: false,
  },
  capabilities: {
    chatTypes: ["direct"],
    reactions: false,
    threads: false,
    media: false,
    nativeCommands: false,
    blockStreaming: true,
  },
  reload: { configPrefixes: ["channels.boxify"] },
  configSchema: {
    schema: {
      type: "object",
      additionalProperties: false,
      properties: {
        enabled: { type: "boolean", title: "启用 Boxify 通道", default: true },
        name: { type: "string", title: "账号名称", default: "Boxify" },
        listenUrl: {
          type: "string",
          title: "插件监听地址",
          default: DEFAULT_LISTEN_URL,
        },
        sharedToken: { type: "string", title: "共享令牌" },
        defaultAgent: { type: "string", title: "默认 Agent", default: "main" },
        accounts: {
          type: "object",
          title: "多账号配置",
          additionalProperties: {
            type: "object",
            additionalProperties: false,
            properties: {
              enabled: { type: "boolean" },
              name: { type: "string" },
              listenUrl: { type: "string" },
              sharedToken: { type: "string" },
              defaultAgent: { type: "string" },
            },
          },
        },
      },
    },
  },
  config: {
    listAccountIds: (cfg) => {
      const root = cfg?.channels?.boxify && typeof cfg.channels.boxify === "object" ? cfg.channels.boxify : {};
      const ids = Object.keys(root.accounts && typeof root.accounts === "object" ? root.accounts : {});
      return ids.length > 0 ? ids : [DEFAULT_ACCOUNT_ID];
    },
    resolveAccount: (cfg, accountId) => resolveChannelConfig(cfg, accountId),
    defaultAccountId: () => DEFAULT_ACCOUNT_ID,
    isConfigured: (account) => Boolean(account?.configured),
    isEnabled: (account) => account?.enabled !== false,
    describeAccount: (account) => ({
      accountId: account.accountId,
      name: account.name,
      enabled: account.enabled,
      configured: account.configured,
      listenUrl: account.listenUrl,
    }),
  },
  messaging: {
    normalizeTarget: (target) => {
      const trimmed = String(target || "").trim();
      return trimmed || undefined;
    },
    targetResolver: {
      looksLikeId: (input) => Boolean(String(input || "").trim()),
      hint: "<conversationId>",
    },
  },
  directory: {
    self: async () => null,
    listPeers: async () => [],
    listGroups: async () => [],
  },
  status: {
    defaultRuntime: createDefaultChannelRuntimeState(DEFAULT_ACCOUNT_ID),
    collectStatusIssues: (accounts) => collectStatusIssuesFromLastError("boxify", accounts),
    buildChannelSummary: ({ snapshot }) => ({
      configured: snapshot.configured ?? false,
      running: snapshot.running ?? false,
      lastStartAt: snapshot.lastStartAt ?? null,
      lastStopAt: snapshot.lastStopAt ?? null,
      lastError: snapshot.lastError ?? null,
    }),
    buildAccountSnapshot: ({ account, runtime }) => ({
      accountId: account.accountId,
      name: account.name,
      enabled: account.enabled,
      configured: account.configured,
      listenUrl: account.listenUrl,
      running: runtime?.running ?? false,
      lastStartAt: runtime?.lastStartAt ?? null,
      lastStopAt: runtime?.lastStopAt ?? null,
      lastError: runtime?.lastError ?? null,
    }),
  },
  gateway: {
    startAccount: async (ctx) => {
      await startInboxServer(ctx);
    },
    stopAccount: async (ctx) => {
      await closeActiveServer(normalizeAccountId(ctx.accountId));
      ctx.setStatus?.({
        accountId: normalizeAccountId(ctx.accountId),
        running: false,
        lastStopAt: Date.now(),
      });
    },
  },
};

const plugin = {
  id: "boxify-channel",
  name: "Boxify Channel",
  description: "将 Boxify 作为 OpenClaw 原生本地聊天通道接入",
  configSchema: emptyPluginConfigSchema(),
  register(api) {
    pluginRuntime = api.runtime;
    api.registerChannel({ plugin: boxifyNativeChannel });
  },
};

export default plugin;
