import { type FC, useLayoutEffect, useRef, useState } from "react";
import { Bot, Loader2, User } from "lucide-react";
import { cn } from "@/lib/utils";
import { formatTimeLabel } from "../domain/chat-panel";
import {
  PENDING_ASSISTANT_ID,
  type RenderedChatMessage,
} from "../types/chat-panel";

interface ChatMessageListProps {
  messages: RenderedChatMessage[];
  isLoadingMessages: boolean;
  scrollToBottomToken: number;
}

/**
 * 渲染消息区，统一处理用户消息、助手消息与流式占位态。
 */
export const ChatMessageList: FC<ChatMessageListProps> = ({
  messages,
  isLoadingMessages,
  scrollToBottomToken,
}) => {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [shouldFollowBottom, setShouldFollowBottom] = useState(true);
  const hasPendingAssistantMessage = messages.some(
    (message) =>
      message.id === PENDING_ASSISTANT_ID &&
      (message.status === "loading" || message.status === "streaming"),
  );

  /**
   * 判断用户是否仍停留在底部附近，只有这种情况下才继续跟随流式输出。
   */
  function handleScroll() {
    const container = containerRef.current;
    if (!container) {
      return;
    }
    const distanceToBottom =
      container.scrollHeight - container.clientHeight - container.scrollTop;
    setShouldFollowBottom(distanceToBottom <= 96);
  }

  /**
   * 仅在用户主动发送消息后滚到底部，避免完成态刷新干扰阅读位置。
   */
  useLayoutEffect(() => {
    if (scrollToBottomToken <= 0) {
      return;
    }
    const container = containerRef.current;
    if (!container) {
      return;
    }
    container.scrollTop = container.scrollHeight;
    setShouldFollowBottom(true);
  }, [scrollToBottomToken]);

  /**
   * 在助手流式生成期间，如果用户仍靠近底部，则随着文本增长继续贴底。
   */
  useLayoutEffect(() => {
    if (!hasPendingAssistantMessage || !shouldFollowBottom) {
      return;
    }
    const container = containerRef.current;
    if (!container) {
      return;
    }
    container.scrollTop = container.scrollHeight;
  }, [messages, hasPendingAssistantMessage, shouldFollowBottom]);

  return (
    <div
      ref={containerRef}
      onScroll={handleScroll}
      className="min-h-0 flex-1 overflow-auto p-6 pb-52"
    >
      {isLoadingMessages ? (
        <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
          <Loader2 className="mr-2 size-4 animate-spin" />
          正在加载消息
        </div>
      ) : messages.length === 0 ? (
        <div className="flex h-full min-h-48 items-center justify-center rounded-xl border border-dashed text-sm text-muted-foreground">
          发送第一条消息开始当前会话
        </div>
      ) : (
        <div className="space-y-4">
          {messages.map((message) => {
            const isUser = message.role === "user";
            const isPendingAssistantMessage =
              message.id === PENDING_ASSISTANT_ID;
            return (
              <div
                key={message.id}
                className={cn(
                  "flex gap-3 items-start",
                  isUser ? "justify-end" : "justify-start",
                )}
              >
                {!isUser ? (
                  <div className="mt-1 rounded-full bg-primary/10 p-2 text-primary">
                    <Bot className="size-4" />
                  </div>
                ) : null}
                <div
                  className={cn(
                    "max-w-[82%] rounded-2xl border px-4 py-3 shadow-xs",
                    isUser ? "border-primary/20 bg-primary/8" : "bg-background",
                  )}
                >
                  <div className="mb-2 flex items-center gap-2 text-xs text-muted-foreground">
                    {isUser ? (
                      <User className="size-3.5" />
                    ) : (
                      <Bot className="size-3.5" />
                    )}
                    <span>{isUser ? "你" : "OpenClaw"}</span>
                    {message.status ? (
                      <span>
                        {message.status === "loading"
                          ? "生成中"
                          : message.status === "error"
                            ? "失败"
                            : message.status}
                      </span>
                    ) : null}
                    <span>{formatTimeLabel(message.createdAt)}</span>
                  </div>
                  <div className="whitespace-pre-wrap wrap-break-word text-sm leading-7 text-foreground/92 text-justify">
                    {isPendingAssistantMessage &&
                    message.status === "loading" ? (
                      <div className="space-y-2 py-1">
                        <div className="h-4 w-40 animate-pulse rounded-full bg-muted" />
                        <div className="h-4 w-56 animate-pulse rounded-full bg-muted/90" />
                        <div className="h-4 w-28 animate-pulse rounded-full bg-muted/80" />
                      </div>
                    ) : (
                      message.content
                    )}
                  </div>
                </div>
                {isUser ? (
                  <div className="mt-1 rounded-full bg-emerald-500/10 p-2 text-emerald-600">
                    <User className="size-4" />
                  </div>
                ) : null}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
