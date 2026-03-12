import http from "node:http";

import { NATIVE_INBOX_PATH, NATIVE_STREAM_INBOX_PATH } from "./constants.js";
import { readJSONBody, waitForAbort } from "./protocol.js";

/**
 * NativeInboxGateway 管理每个账号对应的本地 HTTP inbox 生命周期。
 */
export class NativeInboxGateway {
  constructor({ resolveAccount, createRuntime, logger }) {
    this.resolveAccount = resolveAccount;
    this.createRuntime = createRuntime;
    this.logger = logger;
    this.activeServers = new Map();
  }

  /**
   * 启动指定账号的原生 inbox 服务。
   */
  async start(ctx) {
    const account = this.resolveAccount(ctx.cfg, ctx.accountId);
    if (!account.enabled) {
      ctx.log?.info?.(`[boxify] 账号 ${account.accountId} 已禁用，跳过启动`);
      return;
    }
    if (!account.configured) {
      throw new Error("channels.boxify.listenUrl 未配置");
    }

    await this.stopByAccount(account.accountId);

    const runtime = this.createRuntime(ctx.log);
    const listenURL = new URL(account.listenUrl);
    const inboxURL = `${account.listenUrl}${NATIVE_INBOX_PATH}`;
    const server = http.createServer(async (req, res) => {
      if (req.method !== "POST" || (req.url !== NATIVE_INBOX_PATH && req.url !== NATIVE_STREAM_INBOX_PATH)) {
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
        if (req.url === NATIVE_STREAM_INBOX_PATH) {
          res.writeHead(200, {
            "Content-Type": "text/event-stream",
            "Cache-Control": "no-cache, no-transform",
            Connection: "keep-alive",
          });
          await runtime.dispatchStream(account, payload, res);
          res.end();
          return;
        }

        const result = await runtime.dispatchBuffered(account, payload);
        res.setHeader("Content-Type", "application/json");
        res.end(JSON.stringify(result));
      } catch (error) {
        const message = String(error?.message || error || "unknown error");
        ctx.log?.error?.(`[boxify] inbox 处理失败: ${message}`);
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

    this.activeServers.set(account.accountId, server);
    ctx.setStatus?.({
      accountId: account.accountId,
      running: true,
      lastStartAt: Date.now(),
      lastError: null,
    });
    ctx.log?.info?.(`[boxify] native inbox listening on ${inboxURL}`);

    try {
      await waitForAbort(ctx.abortSignal);
    } finally {
      await this.stopByAccount(account.accountId);
      ctx.setStatus?.({
        accountId: account.accountId,
        running: false,
        lastStopAt: Date.now(),
      });
      ctx.log?.info?.(`[boxify] native inbox stopped for ${account.accountId}`);
    }
  }

  /**
   * 关闭指定账号的 HTTP inbox 服务。
   */
  async stop(ctx) {
    const accountId = this.resolveAccount(ctx.cfg, ctx.accountId).accountId;
    await this.stopByAccount(accountId);
    ctx.setStatus?.({
      accountId,
      running: false,
      lastStopAt: Date.now(),
    });
  }

  /**
   * 按账号 ID 关闭服务实例。
   */
  async stopByAccount(accountId) {
    const server = this.activeServers.get(accountId);
    if (!server) {
      return;
    }

    this.activeServers.delete(accountId);
    await new Promise((resolve) => {
      server.close(() => resolve());
    });
  }
}
