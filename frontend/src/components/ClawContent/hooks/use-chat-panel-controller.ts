import {
  KeyboardEvent,
  startTransition,
  useEffect,
  useEffectEvent,
  useMemo,
  useState,
} from "react";
import { Events } from "@wailsio/runtime";
import { callWails } from "@/lib/utils";
import { ClawService } from "@wails/service";
import { EventType } from "@wails/events/models";
import type {
  ChatEvent,
  Conversation as ChatConversation,
  Message as ChatMessage,
} from "../../../../bindings/github.com/chenyang-zz/boxify/internal/claw/chat/models";
import { ChatEventType } from "../../../../bindings/github.com/chenyang-zz/boxify/internal/claw/chat/models";
import {
  DEFAULT_AGENT_ID,
  PENDING_ASSISTANT_ID,
  type PendingAssistantDraft,
  type RenderedChatMessage,
} from "../types/chat-panel";

interface ChatPanelController {
  conversations: ChatConversation[];
  selectedConversation: ChatConversation | null;
  selectedConversationId: string;
  renderedMessages: RenderedChatMessage[];
  draft: string;
  isInitializing: boolean;
  isRefreshing: boolean;
  isCreatingConversation: boolean;
  isSending: boolean;
  isLoadingMessages: boolean;
  scrollToBottomToken: number;
  setDraft: (value: string) => void;
  handleCreateConversation: () => Promise<void>;
  handleRefresh: () => Promise<void>;
  handleSelectConversation: (conversationId: string) => Promise<void>;
  handleSendMessage: () => Promise<void>;
  handleComposerKeyDown: (
    event: KeyboardEvent<HTMLTextAreaElement>,
  ) => Promise<void>;
}

/**
 * 管理聊天面板的会话、消息与流式事件订阅。
 */
export function useChatPanelController(): ChatPanelController {
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
  const [scrollToBottomToken, setScrollToBottomToken] = useState(0);

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
  const renderedMessages = useMemo<RenderedChatMessage[]>(() => {
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
      const rightTs = new Date(
        right.updatedAt ?? right.createdAt ?? 0,
      ).getTime();
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
      const conversationId = result.data?.id ?? "";
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
          runId: "",
          content: "",
          status: "loading",
        });
        setScrollToBottomToken((current) => current + 1);
      });
      const runId = await callWails(
        ClawService.SendChatMessage,
        selectedConversationId,
        content,
      );
      startTransition(() => {
        setPendingAssistantDraft((current) =>
          current && current.conversationId === selectedConversationId
            ? {
                ...current,
                runId: typeof runId === "string" ? runId : "",
              }
            : current,
        );
      });
      setDraft("");
    } catch (error) {
      setPendingAssistantDraft({
        conversationId: selectedConversationId,
        runId: "",
        content: "发送失败，请检查 boxify 配置或 OpenClaw 运行状态。",
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
   * 收到聊天流式事件后，增量更新当前会话的草稿或最终消息。
   */
  const handleChatEvent = useEffectEvent(async (payload: ChatEvent) => {
    if (
      !payload.conversationId ||
      payload.conversationId !== selectedConversationId
    ) {
      return;
    }
    if (payload.eventType === ChatEventType.ChatEventTypeAssistantDelta) {
      const text = String(payload.payload?.text ?? "");
      setPendingAssistantDraft((current) => ({
        conversationId: payload.conversationId,
        runId: payload.runId ?? current?.runId ?? "",
        content:
          current?.conversationId === payload.conversationId
            ? (current?.content ?? "") + text
            : text,
        status: "streaming",
      }));
      return;
    }
    if (payload.eventType === ChatEventType.ChatEventTypeAssistantDone) {
      const finalContent = String(payload.payload?.text ?? "");
      const finalizedAt = new Date().toISOString();
      setMessages((current) => {
        const nextContent =
          finalContent || pendingAssistantDraft?.content?.trim() || "";
        if (!nextContent) {
          return current;
        }
        return [
          ...current,
          {
            id: `assistant-${payload.runId || Date.now()}`,
            conversationId: payload.conversationId,
            runId: payload.runId ?? "",
            role: "assistant",
            content: nextContent,
            status: "done",
            createdAt: finalizedAt,
          },
        ];
      });
      setConversations((current) => {
        const nextItems = current.map((conversation) =>
          conversation.id === payload.conversationId
            ? {
                ...conversation,
                openClawSessionId:
                  payload.sessionId || conversation.openClawSessionId,
                updatedAt: finalizedAt,
              }
            : conversation,
        );
        return nextItems.sort((left, right) => {
          const leftTs = new Date(left.updatedAt ?? left.createdAt ?? 0).getTime();
          const rightTs = new Date(
            right.updatedAt ?? right.createdAt ?? 0,
          ).getTime();
          return rightTs - leftTs;
        });
      });
      setPendingAssistantDraft(null);
      return;
    }
    if (payload.eventType === ChatEventType.ChatEventTypeAssistantError) {
      setPendingAssistantDraft((current) => ({
        conversationId: payload.conversationId,
        runId: payload.runId ?? current?.runId ?? "",
        content:
          String(payload.payload?.error ?? "") ||
          "发送失败，请检查 boxify 配置或 OpenClaw 运行状态。",
        status: "error",
      }));
    }
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

  return {
    conversations,
    selectedConversation,
    selectedConversationId,
    renderedMessages,
    draft,
    isInitializing,
    isRefreshing,
    isCreatingConversation,
    isSending,
    isLoadingMessages,
    scrollToBottomToken,
    setDraft,
    handleCreateConversation,
    handleRefresh,
    handleSelectConversation,
    handleSendMessage,
    handleComposerKeyDown,
  };
}
