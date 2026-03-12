import {
  DEFAULT_ACCOUNT_ID,
  collectStatusIssuesFromLastError,
  createDefaultChannelRuntimeState,
} from "openclaw/plugin-sdk";

import { BoxifyChannelRuntime } from "./runtime.js";
import { DEFAULT_LISTEN_URL } from "./constants.js";
import { listAccountIds, normalizeAccountId, resolveChannelConfig } from "./config.js";
import { NativeInboxGateway } from "./gateway.js";

/**
 * 创建 Boxify native channel 定义。
 */
export function createBoxifyNativeChannel(runtimeRef) {
  const gateway = new NativeInboxGateway({
    resolveAccount: resolveChannelConfig,
    createRuntime: (logger) => new BoxifyChannelRuntime(runtimeRef.current, logger),
  });

  return {
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
      listAccountIds,
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
        await gateway.start(ctx);
      },
      stopAccount: async (ctx) => {
        await gateway.stop({
          ...ctx,
          accountId: normalizeAccountId(ctx.accountId),
        });
      },
    },
  };
}
