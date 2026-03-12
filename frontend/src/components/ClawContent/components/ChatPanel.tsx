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
  FC,
  KeyboardEvent,
  startTransition,
  useEffect,
  useEffectEvent,
  useMemo,
  useState,
} from "react";
import {
  Bot,
  Clock3,
  CornerDownLeft,
  Loader2,
  MessageSquarePlus,
  RefreshCw,
  User,
} from "lucide-react";
import { Events } from "@wailsio/runtime";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Separator } from "@/components/ui/separator";
import { callWails, cn } from "@/lib/utils";
import { ClawService } from "@wails/service";
import { EventType } from "@wails/events/models";
import type {
  ChatEvent,
  Conversation as ChatConversation,
  Message as ChatMessage,
} from "../../../../bindings/github.com/chenyang-zz/boxify/internal/claw/chat/models";
import { PanelHeader } from "./PanelHeader";

const DEFAULT_AGENT_ID = "main";
const PENDING_ASSISTANT_ID = "pending-assistant";

interface PendingAssistantDraft {
  conversationId: string;
  content: string;
  status: string;
}

/**
 * 提取会话标题，避免空标题直接出现在会话列表中。
 */
function getConversationTitle(conversation: ChatConversation): string {
  const title = conversation.title?.trim();
  if (title) {
    return title;
  }
  return conversation.agentId?.trim()
    ? `${conversation.agentId} 会话`
    : "未命名会话";
}

/**
 * 将时间字符串格式化为简短的本地时间。
 */
function formatTimeLabel(value?: string): string {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return new Intl.DateTimeFormat("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    month: "2-digit",
    day: "2-digit",
  }).format(date);
}

/**
 * 聊天面板组件
 * 负责对接本地聊天会话、消息列表与发送能力。
 */
export const ChatPanel: FC = () => {
  const [conversations, setConversations] = useState<ChatConversation[]>([]);
  const [selectedConversationId, setSelectedConversationId] = useState("");
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [pendingAssistantDraft, setPendingAssistantDraft] =
    useState<PendingAssistantDraft | null>(null);
  const [isInitializing, setIsInitializing] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isCreatingConversation, setIsCreatingConversation] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [isLoadingMessages, setIsLoadingMessages] = useState(false);

  const selectedConversation = useMemo(
    () =>
      conversations.find(
        (conversation) => conversation.id === selectedConversationId,
      ) ?? null,
    [conversations, selectedConversationId],
  );

  /**
   * 合并正式消息与发送中的助手草稿，减少同步返回前的空白等待。
   */
  const renderedMessages = useMemo(() => {
    if (
      !pendingAssistantDraft ||
      pendingAssistantDraft.conversationId !== selectedConversationId
    ) {
      return messages;
    }
    return [
      ...messages,
      {
        id: PENDING_ASSISTANT_ID,
        conversationId: pendingAssistantDraft.conversationId,
        runId: "",
        role: "assistant",
        content: pendingAssistantDraft.content,
        status: pendingAssistantDraft.status,
        createdAt: undefined,
      },
    ];
  }, [messages, pendingAssistantDraft, selectedConversationId]);

  /**
   * 拉取聊天会话列表，并在必要时补建默认会话。
   */
  async function loadConversations(preferredConversationId?: string) {
    const result = await callWails(ClawService.ListChatConversations);
    const nextItems = [...(result.items ?? [])].sort((left, right) => {
      const leftTs = new Date(left.updatedAt ?? left.createdAt ?? 0).getTime();
      const rightTs = new Date(right.updatedAt ?? right.createdAt ?? 0).getTime();
      return rightTs - leftTs;
    });

    if (nextItems.length === 0) {
      const created = await callWails(
        ClawService.CreateChatConversation,
        DEFAULT_AGENT_ID,
      );
      const conversation = created.data;
      const seededItems = conversation ? [conversation] : [];
      startTransition(() => {
        setConversations(seededItems);
        setSelectedConversationId(conversation?.id ?? "");
      });
      return conversation?.id ?? "";
    }

    const fallbackId = preferredConversationId?.trim();
    const nextSelectedId =
      (fallbackId &&
      nextItems.some((conversation) => conversation.id === fallbackId)
        ? fallbackId
        : selectedConversationId &&
            nextItems.some(
              (conversation) => conversation.id === selectedConversationId,
            )
          ? selectedConversationId
          : nextItems[0]?.id) ?? "";

    startTransition(() => {
      setConversations(nextItems);
      setSelectedConversationId(nextSelectedId);
    });
    return nextSelectedId;
  }

  /**
   * 拉取指定会话的全部消息。
   */
  async function loadMessages(conversationId: string) {
    const resolvedConversationId = conversationId.trim();
    if (!resolvedConversationId) {
      setMessages([]);
      return;
    }

    setIsLoadingMessages(true);
    try {
      const result = await callWails(
        ClawService.GetChatMessages,
        resolvedConversationId,
      );
      const nextItems = [...(result.items ?? [])].sort((left, right) => {
        const leftTs = new Date(left.createdAt ?? 0).getTime();
        const rightTs = new Date(right.createdAt ?? 0).getTime();
        return leftTs - rightTs;
      });
      startTransition(() => {
        setMessages(nextItems);
      });
    } finally {
      setIsLoadingMessages(false);
    }
  }

  /**
   * 初始化聊天状态，保证面板首次打开即可进入可用会话。
   */
  async function initializeChatPanel() {
    setIsInitializing(true);
    try {
      const conversationId = await loadConversations();
      if (conversationId) {
        await loadMessages(conversationId);
      } else {
        setMessages([]);
      }
    } finally {
      setIsInitializing(false);
    }
  }

  /**
   * 创建新的默认聊天会话，并切换到该会话。
   */
  async function handleCreateConversation() {
    setIsCreatingConversation(true);
    try {
      const result = await callWails(
        ClawService.CreateChatConversation,
        DEFAULT_AGENT_ID,
      );
      const conversation = result.data;
      const conversationId = conversation?.id ?? "";
      await loadConversations(conversationId);
      await loadMessages(conversationId);
      setDraft("");
    } finally {
      setIsCreatingConversation(false);
    }
  }

  /**
   * 刷新会话与消息列表，保留当前选中的会话。
   */
  async function handleRefresh() {
    setIsRefreshing(true);
    try {
      const conversationId = await loadConversations(selectedConversationId);
      await loadMessages(conversationId);
    } finally {
      setIsRefreshing(false);
    }
  }

  /**
   * 切换当前会话并加载对应消息。
   */
  async function handleSelectConversation(conversationId: string) {
    if (!conversationId || conversationId === selectedConversationId) {
      return;
    }
    setSelectedConversationId(conversationId);
    await loadMessages(conversationId);
  }

  /**
   * 发送当前输入框中的消息，并在完成后刷新会话与消息。
   */
  async function handleSendMessage() {
    const content = draft.trim();
    if (!selectedConversationId || !content || isSending) {
      return;
    }

    setIsSending(true);
    const optimisticUserMessage: ChatMessage = {
      id: `optimistic-user-${Date.now()}`,
      conversationId: selectedConversationId,
      runId: "",
      role: "user",
      content,
      status: "done",
      createdAt: new Date().toISOString(),
    };

    try {
      startTransition(() => {
        setMessages((current) => [...current, optimisticUserMessage]);
        setPendingAssistantDraft({
          conversationId: selectedConversationId,
          content: "正在等待 OpenClaw 返回完整回复...",
          status: "loading",
        });
      });
      await callWails(
        ClawService.SendChatMessage,
        selectedConversationId,
        content,
      );
      setDraft("");
      await loadConversations(selectedConversationId);
      await loadMessages(selectedConversationId);
      setPendingAssistantDraft(null);
    } catch (error) {
      setPendingAssistantDraft({
        conversationId: selectedConversationId,
        content: "发送失败，请检查 boxify-channel 配置或 OpenClaw 运行状态。",
        status: "error",
      });
      throw error;
    } finally {
      setIsSending(false);
    }
  }

  /**
   * 处理输入框快捷键，支持 Enter 发送和 Shift+Enter 换行。
   */
  async function handleComposerKeyDown(
    event: KeyboardEvent<HTMLTextAreaElement>,
  ) {
    if (event.key !== "Enter" || event.shiftKey) {
      return;
    }
    event.preventDefault();
    await handleSendMessage();
  }

  /**
   * 收到当前会话的助手完成事件后，刷新消息列表。
   */
  const handleChatEvent = useEffectEvent(async (payload: ChatEvent) => {
    if (
      !payload.conversationId ||
      payload.conversationId !== selectedConversationId ||
      payload.eventType !== "assistant_done"
    ) {
      return;
    }
    setPendingAssistantDraft(null);
    await loadMessages(payload.conversationId);
  });

  useEffect(() => {
    void initializeChatPanel();
  }, []);

  useEffect(() => {
    const unbind = Events.On(
      EventType.EventTypeClawChatEvent,
      (event: { data: ChatEvent }) => {
        void handleChatEvent(event.data);
      },
    );
    return () => {
      unbind();
    };
  }, [handleChatEvent]);

  return (
    <div className="flex h-full w-full min-h-0 flex-col p-6">
      <PanelHeader
        className="mb-6"
        title="聊天"
        description="通过 Boxify 本地会话直接投递到 OpenClaw boxify channel。"
        align="start"
        actions={
          <div className="flex flex-wrap items-center justify-end gap-2">
            <Button
              variant="outline"
              size="sm"
              className="h-10 min-w-48 justify-start text-sm"
              disabled
            >
              {selectedConversation
                ? getConversationTitle(selectedConversation)
                : "未选择会话"}
            </Button>
            <Button
              variant="outline"
              size="icon-sm"
              aria-label="刷新会话"
              onClick={() => void handleRefresh()}
              disabled={isInitializing || isRefreshing || isSending}
            >
              {isRefreshing ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <RefreshCw className="size-4" />
              )}
            </Button>
            <Separator orientation="vertical" className="mx-1 h-6" />
            <Button
              variant="outline"
              size="sm"
              className="h-10 px-4"
              onClick={() => void handleCreateConversation()}
              disabled={isCreatingConversation || isSending}
            >
              {isCreatingConversation ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <MessageSquarePlus className="size-4" />
              )}
              新建会话
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-10 px-4 text-muted-foreground"
              disabled
            >
              <Clock3 className="size-4" />
              {conversations.length} 个会话
            </Button>
          </div>
        }
      />

      <div className="grid min-h-0 flex-1 gap-4 lg:grid-cols-[260px_minmax(0,1fr)]">
        <div className="min-h-0 rounded-2xl border bg-card/70">
          <div className="border-b px-4 py-3">
            <div className="text-sm font-semibold">会话列表</div>
            <div className="mt-1 text-xs text-muted-foreground">
              当前只保存在 Boxify 进程内存中。
            </div>
          </div>
          <div className="h-full max-h-full overflow-auto p-2">
            {isInitializing ? (
              <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
                <Loader2 className="mr-2 size-4 animate-spin" />
                正在初始化聊天会话
              </div>
            ) : conversations.length === 0 ? (
              <div className="px-3 py-4 text-sm text-muted-foreground">
                暂无会话
              </div>
            ) : (
              conversations.map((conversation) => {
                const active = conversation.id === selectedConversationId;
                return (
                  <button
                    type="button"
                    key={conversation.id}
                    onClick={() => void handleSelectConversation(conversation.id)}
                    className={cn(
                      "w-full rounded-xl border px-3 py-3 text-left transition-colors",
                      active
                        ? "border-primary/40 bg-primary/8"
                        : "border-transparent hover:border-border hover:bg-muted/40",
                    )}
                  >
                    <div className="flex items-center justify-between gap-3">
                      <div className="truncate text-sm font-medium">
                        {getConversationTitle(conversation)}
                      </div>
                      <div className="shrink-0 text-[11px] text-muted-foreground">
                        {formatTimeLabel(
                          conversation.updatedAt ?? conversation.createdAt,
                        )}
                      </div>
                    </div>
                    <div className="mt-1 text-xs text-muted-foreground">
                      Agent: {conversation.agentId || DEFAULT_AGENT_ID}
                    </div>
                    {conversation.openClawSessionId ? (
                      <div className="mt-1 truncate text-[11px] text-muted-foreground">
                        Session: {conversation.openClawSessionId}
                      </div>
                    ) : null}
                  </button>
                );
              })
            )}
          </div>
        </div>

        <div className="flex min-h-0 flex-col rounded-2xl border bg-card/60">
          <div className="border-b px-4 py-3">
            <div className="text-sm font-semibold">消息</div>
            <div className="mt-1 text-xs text-muted-foreground">
              当前使用同步 request/response，发送后会等待完整回复落库。
            </div>
          </div>

          <div className="min-h-0 flex-1 overflow-auto px-4 py-4">
            {isLoadingMessages ? (
              <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
                <Loader2 className="mr-2 size-4 animate-spin" />
                正在加载消息
              </div>
            ) : renderedMessages.length === 0 ? (
              <div className="flex h-full min-h-48 items-center justify-center rounded-xl border border-dashed text-sm text-muted-foreground">
                发送第一条消息开始当前会话
              </div>
            ) : (
              <div className="space-y-4">
                {renderedMessages.map((message) => {
                  const isUser = message.role === "user";
                  const isPendingAssistantMessage =
                    message.id === PENDING_ASSISTANT_ID;
                  return (
                    <div
                      key={message.id}
                      className={cn(
                        "flex gap-3",
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
                          isUser
                            ? "border-primary/20 bg-primary/8"
                            : "bg-background",
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
                        <div className="whitespace-pre-wrap break-words text-sm leading-7 text-foreground/92">
                          {isPendingAssistantMessage &&
                          message.status === "loading" ? (
                            <span className="inline-flex items-center gap-2 text-muted-foreground">
                              <Loader2 className="size-4 animate-spin" />
                              {message.content}
                            </span>
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

          <div className="border-t p-4">
            <div className="flex flex-col gap-3">
              <Textarea
                value={draft}
                onChange={(event) => setDraft(event.target.value)}
                onKeyDown={(event) => void handleComposerKeyDown(event)}
                placeholder="输入消息，按 Enter 发送，Shift+Enter 换行"
                className="min-h-24 resize-none bg-background"
                disabled={!selectedConversationId || isSending}
              />
              <div className="flex items-center justify-between gap-3">
                <div className="text-xs text-muted-foreground">
                  {selectedConversation ? (
                    <>
                      当前 Agent:{" "}
                      {selectedConversation.agentId || DEFAULT_AGENT_ID}
                    </>
                  ) : (
                    "请先创建会话"
                  )}
                </div>
                <div className="flex items-center gap-3">
                  <Button
                    variant="outline"
                    className="h-10 px-5"
                    onClick={() => void handleCreateConversation()}
                    disabled={isCreatingConversation || isSending}
                  >
                    新建会话
                  </Button>
                  <Button
                    className="h-10 px-6"
                    onClick={() => void handleSendMessage()}
                    disabled={!selectedConversationId || !draft.trim() || isSending}
                  >
                    {isSending ? (
                      <Loader2 className="size-4 animate-spin" />
                    ) : (
                      <CornerDownLeft className="size-3.5" />
                    )}
                    发送
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ChatPanel;
