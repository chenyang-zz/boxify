import { buildInboundContext, computeDeltaText, writeSSE } from "./protocol.js";

/**
 * BoxifyChannelRuntime 封装 Boxify inbox 到 OpenClaw reply runtime 的翻译逻辑。
 */
export class BoxifyChannelRuntime {
  constructor(runtime, logger) {
    this.runtime = runtime;
    this.logger = logger;
  }

  /**
   * 按同步模式处理一次 inbox 请求。
   */
  async dispatchBuffered(account, incoming) {
    const dispatchReply = this.runtime?.channel?.reply?.dispatchReplyWithBufferedBlockDispatcher;
    if (typeof dispatchReply !== "function") {
      throw new Error("OpenClaw buffered reply runtime 不可用");
    }

    const { cfgPromise, msgCtx, conversationId, sessionId } = buildInboundContext(this.runtime, account, incoming);
    const replyChunks = [];

    try {
      await dispatchReply({
        ctx: msgCtx,
        cfg: await cfgPromise,
        dispatcherOptions: {
          deliver: async (payload) => {
            const chunk = String(payload?.text ?? payload?.body ?? "").trim();
            if (chunk) {
              replyChunks.push(chunk);
            }
          },
          onReplyStart: () => {
            this.logger?.info?.(`[boxify] 原生回复开始，conversation=${conversationId}`);
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
      this.logger?.error?.(`[boxify] 原生 channel 处理失败: ${message}`);
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
   * 按流式模式处理一次 inbox 请求。
   */
  async dispatchStream(account, incoming, res) {
    const createReplyDispatcherWithTyping = this.runtime?.channel?.reply?.createReplyDispatcherWithTyping;
    const dispatchReplyFromConfig = this.runtime?.channel?.reply?.dispatchReplyFromConfig;
    const withReplyDispatcher = this.runtime?.channel?.reply?.withReplyDispatcher;
    if (
      typeof createReplyDispatcherWithTyping !== "function" ||
      typeof dispatchReplyFromConfig !== "function" ||
      typeof withReplyDispatcher !== "function"
    ) {
      throw new Error("OpenClaw streaming reply runtime 不可用");
    }

    const { cfgPromise, msgCtx, conversationId, sessionId, runId } = buildInboundContext(
      this.runtime,
      account,
      incoming,
    );
    let finalText = "";
    let streamedText = "";
    let lastPartial = "";
    const emitDelta = (text) => {
      const normalized = String(text ?? "");
      if (!normalized) {
        return;
      }
      const { delta, merged } = computeDeltaText(streamedText, normalized);
      streamedText = merged;
      if (!delta) {
        return;
      }
      writeSSE(res, "delta", {
        eventType: "delta",
        conversationId,
        sessionId,
        runId,
        text: delta,
      });
    };

    const { dispatcher, replyOptions, markDispatchIdle } = createReplyDispatcherWithTyping({
      deliver: async (payload, info) => {
        const text = String(payload?.text ?? payload?.body ?? "");
        if (!text.trim()) {
          return;
        }
        if (info?.kind === "final") {
          finalText = text;
          return;
        }
        emitDelta(text);
      },
      onReplyStart: () => {
        this.logger?.info?.(`[boxify] 原生流式回复开始，conversation=${conversationId}`);
        writeSSE(res, "start", {
          eventType: "start",
          conversationId,
          sessionId,
          runId,
        });
      },
      onError: (error, info) => {
        this.logger?.error?.(`[boxify] ${info?.kind || "reply"} 流式回复失败: ${String(error)}`);
      },
    });

    try {
      await withReplyDispatcher({
        dispatcher,
        run: async () =>
          dispatchReplyFromConfig({
            ctx: msgCtx,
            cfg: await cfgPromise,
            dispatcher,
            replyOptions: {
              ...replyOptions,
              onPartialReply: async (payload) => {
                const text = String(payload?.text ?? payload?.body ?? "");
                if (!text.trim() || text === lastPartial) {
                  return;
                }
                lastPartial = text;
                emitDelta(text);
              },
            },
          }),
        onSettled: () => {
          markDispatchIdle();
        },
      });

      emitDelta(finalText);

      writeSSE(res, "done", {
        eventType: "done",
        conversationId,
        sessionId,
        runId,
        text: streamedText || finalText,
      });
    } catch (error) {
      const message = String(error?.message || error || "OpenClaw 执行失败");
      this.logger?.error?.(`[boxify] 原生 channel 流式处理失败: ${message}`);
      writeSSE(res, "error", {
        eventType: "error",
        conversationId,
        sessionId,
        runId,
        error: message,
      });
    }
  }
}
