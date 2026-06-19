// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import {
  type FC,
  type KeyboardEvent,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import {
  ArrowUp,
  ChevronDown,
  Copy,
  CornerUpRight,
  Loader2,
  MessageCircle,
  Mic,
  MoreHorizontal,
  Plus,
  RotateCcw,
  Settings,
  ThumbsDown,
  ThumbsUp,
} from "lucide-react";
import { Streamdown } from "streamdown";
import "streamdown/styles.css";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  createChatSession,
  getChatSession,
  handleSessionAuthError,
  sendChatMessageStream,
} from "@/api/sessions";
import { useChatSessionStore } from "@/store/chat-session.store";
import type { MessageSSEEvent, SessionEvent } from "@/types/api/session";
import { cjk } from "@streamdown/cjk";
import { code } from "@streamdown/code";
import { math } from "@streamdown/math";
import { mermaid } from "@streamdown/mermaid";

interface ChatUIMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  createdAtLabel?: string;
  status?: "streaming" | "error";
}

const assistantActions = [
  { label: "复制", icon: Copy },
  { label: "赞", icon: ThumbsUp },
  { label: "踩", icon: ThumbsDown },
  { label: "重新生成", icon: RotateCcw },
  { label: "展开", icon: CornerUpRight },
];

/**
 * isOptimisticSessionId 判断侧边栏乐观会话临时 id。
 */
function isOptimisticSessionId(sessionId: string) {
  return sessionId.startsWith("optimistic-session-");
}

/**
 * createLocalMessageId 生成前端消息临时 id。
 */
function createLocalMessageId(prefix: string) {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return `${prefix}-${crypto.randomUUID()}`;
  }
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

/**
 * formatEventTime 将接口时间格式化为消息时间标签。
 */
function formatEventTime(value?: string | number | null) {
  if (value === undefined || value === null) {
    return "";
  }

  const timestamp =
    typeof value === "number" && value < 1_000_000_000_000
      ? value * 1000
      : value;
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) {
    return "";
  }

  return new Intl.DateTimeFormat("zh-CN", {
    weekday: "short",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(date);
}

/**
 * isMessageEvent 判断会话事件是否为可渲染消息。
 */
function isMessageEvent(event: SessionEvent): event is MessageSSEEvent {
  return event.event === "message" && typeof event.data === "object";
}

/**
 * buildMessagesFromEvents 将后端事件流合并为 Chat UI 消息列表。
 */
function buildMessagesFromEvents(events: SessionEvent[]) {
  const messages: ChatUIMessage[] = [];

  for (const event of events) {
    if (!isMessageEvent(event)) {
      continue;
    }

    const content = event.data.message ?? "";
    if (!content) {
      continue;
    }

    const previous = messages.at(-1);
    if (previous?.role === event.data.role) {
      previous.content += content;
      previous.createdAtLabel =
        formatEventTime(event.data.created_at) || previous.createdAtLabel;
      continue;
    }

    messages.push({
      id: event.data.event_id || createLocalMessageId(event.data.role),
      role: event.data.role,
      content,
      createdAtLabel: formatEventTime(event.data.created_at),
    });
  }

  return messages;
}

/**
 * ChatContent 渲染 Chat tab 的独立主内容区域。
 */
const ChatContent: FC = () => {
  const selectedSessionId = useChatSessionStore(
    (state) => state.selectedSessionId,
  );
  const selectedSessionTitle = useChatSessionStore(
    (state) => state.selectedSessionTitle,
  );
  const setSelectedSession = useChatSessionStore(
    (state) => state.setSelectedSession,
  );
  const requestSidebarRefresh = useChatSessionStore(
    (state) => state.requestSidebarRefresh,
  );
  const [messages, setMessages] = useState<ChatUIMessage[]>([]);
  const [inputValue, setInputValue] = useState("");
  const [loadingSession, setLoadingSession] = useState(false);
  const [sending, setSending] = useState(false);
  const [streamingAssistantId, setStreamingAssistantId] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const loadRequestIdRef = useRef(0);
  const skipNextLoadSessionIdRef = useRef("");
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ block: "end" });
  }, [messages, loadingSession]);

  useEffect(() => {
    if (!selectedSessionId) {
      setMessages([]);
      setErrorMessage("");
      setLoadingSession(false);
      return;
    }

    if (isOptimisticSessionId(selectedSessionId)) {
      setMessages([]);
      setErrorMessage("");
      setLoadingSession(true);
      return;
    }

    if (skipNextLoadSessionIdRef.current === selectedSessionId) {
      skipNextLoadSessionIdRef.current = "";
      return;
    }

    const requestId = loadRequestIdRef.current + 1;
    loadRequestIdRef.current = requestId;
    setMessages([]);
    setLoadingSession(true);
    setErrorMessage("");

    void (async () => {
      try {
        const session = await getChatSession(selectedSessionId);
        if (loadRequestIdRef.current !== requestId) {
          return;
        }
        setSelectedSession(selectedSessionId, session.title || "新对话");
        setMessages(buildMessagesFromEvents(session.events ?? []));
      } catch (error) {
        if (loadRequestIdRef.current !== requestId) {
          return;
        }
        if (await handleSessionAuthError(error)) {
          return;
        }
        const message =
          error instanceof Error ? error.message : "加载会话消息失败";
        setErrorMessage(message);
        toast.error("加载会话消息失败", { description: message });
      } finally {
        if (loadRequestIdRef.current === requestId) {
          setLoadingSession(false);
        }
      }
    })();
  }, [selectedSessionId, setSelectedSession]);

  /**
   * appendAssistantDelta 将流式 assistant 片段追加到当前占位消息。
   */
  const appendAssistantDelta = useCallback(
    (
      assistantId: string,
      delta: string,
      createdAt?: string | number | null,
    ) => {
      if (!delta) {
        return;
      }
      setMessages((current) => {
        let found = false;
        const next = current.map((message) => {
          if (message.id !== assistantId) {
            return message;
          }
          found = true;
          return {
            ...message,
            content: `${message.content}${delta}`,
            createdAtLabel:
              formatEventTime(createdAt) || message.createdAtLabel,
            status: "streaming" as const,
          };
        });

        return found
          ? next
          : [
              ...next,
              {
                id: assistantId,
                role: "assistant",
                content: delta,
                createdAtLabel: formatEventTime(createdAt),
                status: "streaming",
              },
            ];
      });
    },
    [],
  );

  /**
   * markAssistantDone 结束当前 assistant 流式消息状态。
   */
  const markAssistantDone = useCallback((assistantId: string) => {
    setMessages((current) =>
      current.flatMap((message) => {
        if (message.id !== assistantId) {
          return [message];
        }
        return message.content.trim()
          ? [{ ...message, status: undefined }]
          : [];
      }),
    );
    setStreamingAssistantId("");
  }, []);

  /**
   * markAssistantError 展示发送失败状态并保留用户消息。
   */
  const markAssistantError = useCallback(
    (assistantId: string, detail: string) => {
      setMessages((current) =>
        current.map((message) =>
          message.id === assistantId
            ? {
                ...message,
                content: message.content || detail,
                status: "error",
              }
            : message,
        ),
      );
      setStreamingAssistantId("");
    },
    [],
  );

  /**
   * handleSend 发送当前输入并消费后端流式响应。
   */
  const handleSend = useCallback(async () => {
    const content = inputValue.trim();
    if (!content || sending || isOptimisticSessionId(selectedSessionId)) {
      return;
    }

    let sessionId = selectedSessionId;
    let assistantMessageId = "";
    setSending(true);
    setErrorMessage("");

    try {
      if (!sessionId) {
        const created = await createChatSession(null);
        sessionId = created.session_id;
        skipNextLoadSessionIdRef.current = sessionId;
        setSelectedSession(sessionId, "新对话");
        requestSidebarRefresh();
      }

      const userMessageId = createLocalMessageId("user");
      assistantMessageId = createLocalMessageId("assistant");
      setInputValue("");
      setStreamingAssistantId(assistantMessageId);
      setMessages((current) => [
        ...current,
        {
          id: userMessageId,
          role: "user",
          content,
          createdAtLabel: formatEventTime(Date.now()),
        },
        {
          id: assistantMessageId,
          role: "assistant",
          content: "",
          status: "streaming",
        },
      ]);

      await sendChatMessageStream(
        sessionId,
        {
          message: content,
          attachments: [],
        },
        {
          onMessage: (event) => {
            if (event.data.role === "assistant") {
              appendAssistantDelta(
                assistantMessageId,
                event.data.message ?? "",
                event.data.created_at,
              );
            }
          },
          onDone: () => {
            markAssistantDone(assistantMessageId);
          },
        },
      );

      markAssistantDone(assistantMessageId);
      requestSidebarRefresh();
    } catch (error) {
      if (await handleSessionAuthError(error)) {
        return;
      }
      const message = error instanceof Error ? error.message : "发送消息失败";
      setErrorMessage(message);
      toast.error("发送消息失败", { description: message });
      if (assistantMessageId) {
        markAssistantError(assistantMessageId, "请求失败，请稍后重试。");
      }
    } finally {
      setSending(false);
    }
  }, [
    appendAssistantDelta,
    inputValue,
    markAssistantDone,
    markAssistantError,
    requestSidebarRefresh,
    selectedSessionId,
    sending,
    setSelectedSession,
  ]);

  /**
   * handleComposerKeyDown 支持 Enter 发送、Shift+Enter 换行。
   */
  const handleComposerKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      void handleSend();
    }
  };

  const selectedSessionPending = isOptimisticSessionId(selectedSessionId);
  const canSend =
    inputValue.trim().length > 0 && !sending && !selectedSessionPending;
  const headerTitle =
    selectedSessionTitle || (selectedSessionId ? "新对话" : "新对话");

  return (
    <section className="relative flex h-full min-w-0 flex-col bg-background text-foreground">
      <header className="relative z-10 flex h-11 shrink-0 items-center bg-background px-3 text-foreground xl:absolute xl:inset-x-0 xl:top-0 xl:h-12 xl:bg-background/0 xl:px-5">
        <div className="flex min-w-0 items-center gap-1.5">
          <div className="truncate text-sm font-semibold leading-none">
            {headerTitle}
          </div>
          <Button
            type="button"
            variant="ghost"
            size="icon-xs"
            className="rounded-full text-muted-foreground hover:bg-muted hover:text-foreground"
            aria-label="更多聊天操作"
            title="更多聊天操作"
          >
            <MoreHorizontal />
          </Button>
        </div>
      </header>

      <div className="scrollbar-hide min-h-0 flex-1 overflow-auto px-3 pb-8 pt-4 sm:px-6 sm:pt-8 lg:pt-14">
        <div className="mx-auto flex w-full max-w-3xl flex-col gap-10">
          {loadingSession ? (
            <div className="flex min-h-48 items-center justify-center text-sm text-muted-foreground">
              <Loader2 className="mr-2 size-4 animate-spin" />
              正在加载消息
            </div>
          ) : messages.length === 0 ? (
            <div className="flex min-h-48 flex-col items-center justify-center gap-3 text-center text-muted-foreground">
              <div className="flex size-12 items-center justify-center rounded-full bg-muted text-primary">
                <MessageCircle className="size-6" />
              </div>
              <div className="text-sm">
                {selectedSessionPending
                  ? "正在创建新对话"
                  : selectedSessionId
                    ? "暂无消息，发送第一条消息开始对话"
                    : "输入消息后会自动创建新对话"}
              </div>
              {errorMessage ? (
                <div className="max-w-md text-xs text-destructive">
                  {errorMessage}
                </div>
              ) : null}
            </div>
          ) : (
            messages.map((message) =>
              message.role === "user" ? (
                <div key={message.id} className="flex justify-end">
                  <div className="max-w-[82%] rounded-2xl bg-muted px-4 py-3 text-sm leading-6 text-foreground shadow-sm">
                    <span className="whitespace-pre-wrap">
                      {message.content}
                    </span>
                  </div>
                </div>
              ) : (
                <article
                  key={message.id}
                  className="w-full text-sm leading-7 text-foreground"
                >
                  {message.content ? (
                    <Streamdown
                      plugins={{ code, math, mermaid, cjk }}
                      mode="static"
                      className="chat-streamdown"
                    >
                      {message.content}
                    </Streamdown>
                  ) : (
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <Loader2 className="size-4 animate-spin" />
                      正在思考
                    </div>
                  )}
                  {message.status === "error" ? (
                    <div className="mt-2 text-xs text-destructive">
                      发送失败
                    </div>
                  ) : null}
                  <div className="mt-3 flex flex-wrap items-center gap-2 text-muted-foreground">
                    {assistantActions.map(({ label, icon: Icon }) => (
                      <button
                        key={label}
                        type="button"
                        className="inline-flex size-5 items-center justify-center rounded-md transition hover:bg-muted hover:text-foreground"
                        aria-label={label}
                        title={label}
                      >
                        <Icon className="size-3.5" />
                      </button>
                    ))}
                    {message.createdAtLabel ? (
                      <span className="ml-1 text-xs">
                        {message.createdAtLabel}
                      </span>
                    ) : null}
                    {message.id === streamingAssistantId ? (
                      <span className="ml-1 text-xs">生成中</span>
                    ) : null}
                  </div>
                </article>
              ),
            )
          )}
          <div ref={messagesEndRef} />
        </div>
      </div>

      <div className="shrink-0 px-3 pb-3 sm:px-6">
        <div className="shadow-composer mx-auto flex w-full max-w-3xl flex-col rounded-[20px] border border-border/70 bg-card px-3 pb-2 pt-4 text-card-foreground">
          <Textarea
            aria-label="输入消息"
            placeholder="要求后续变更"
            value={inputValue}
            onChange={(event) => setInputValue(event.target.value)}
            onKeyDown={handleComposerKeyDown}
            className="min-h-12 resize-none border-0 bg-transparent px-0 py-0 text-sm text-foreground shadow-none placeholder:text-muted-foreground/80 focus-visible:ring-0"
          />
          <div className="flex flex-wrap items-center justify-between gap-2 gap-y-1 pt-1">
            <div className="flex w-full min-w-0 items-center gap-1 sm:w-auto">
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="rounded-full text-muted-foreground hover:bg-muted hover:text-foreground"
                aria-label="添加内容"
              >
                <Plus />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="min-w-0 rounded-full px-2 text-muted-foreground hover:bg-muted hover:text-foreground"
                aria-label="选择自定义配置"
              >
                <Settings />
                <span className="truncate">自定义</span>
                <ChevronDown />
              </Button>
            </div>

            <div className="flex w-full min-w-0 flex-wrap items-center justify-between gap-1 sm:w-auto sm:justify-end">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="rounded-full px-2 text-muted-foreground hover:bg-muted hover:text-foreground"
                aria-label="选择模型强度"
              >
                <span>5.5 高</span>
                <ChevronDown />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="hidden rounded-full text-muted-foreground hover:bg-muted hover:text-foreground sm:inline-flex"
                aria-label="语音输入"
              >
                <Mic />
              </Button>
              <Button
                type="button"
                size="icon-sm"
                disabled={!canSend}
                onClick={() => void handleSend()}
                className="rounded-full bg-primary text-primary-foreground hover:bg-primary/90"
                aria-label="发送消息"
              >
                {sending ? <Loader2 className="animate-spin" /> : <ArrowUp />}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default ChatContent;
