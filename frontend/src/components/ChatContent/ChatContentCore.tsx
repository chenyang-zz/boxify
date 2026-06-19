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

import { type FC, useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import {
  createChatSession,
  getChatSession,
  handleSessionAuthError,
  sendChatMessageStream,
} from "@/api/sessions";
import { useChatSessionStore } from "@/store/chat-session.store";

import { ChatComposer } from "./components/ChatComposer";
import { ChatHeader } from "./components/ChatHeader";
import { ChatMessageList } from "./components/ChatMessageList";
import {
  buildMessagesFromEvents,
  createLocalMessageId,
  formatEventTime,
  isOptimisticSessionId,
  type ChatUIMessage,
} from "./domain/messages";

/**
 * ChatContentCore 编排 Chat 主内容的数据加载、发送和布局。
 */
export const ChatContentCore: FC = () => {
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

  const selectedSessionPending = isOptimisticSessionId(selectedSessionId);
  const canSend =
    inputValue.trim().length > 0 && !sending && !selectedSessionPending;
  const headerTitle =
    selectedSessionTitle || (selectedSessionId ? "新对话" : "聊天功能");

  return (
    <section className="relative flex h-full min-w-0 flex-col bg-background text-foreground">
      <ChatHeader title={headerTitle} />
      <ChatMessageList
        messages={messages}
        loadingSession={loadingSession}
        selectedSessionId={selectedSessionId}
        selectedSessionPending={selectedSessionPending}
        errorMessage={errorMessage}
        streamingAssistantId={streamingAssistantId}
        messagesEndRef={messagesEndRef}
      />
      <ChatComposer
        inputValue={inputValue}
        sending={sending}
        canSend={canSend}
        onInputChange={setInputValue}
        onSend={() => void handleSend()}
      />
    </section>
  );
};

export default ChatContentCore;
