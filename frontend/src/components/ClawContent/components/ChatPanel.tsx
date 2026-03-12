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

import { FC } from "react";
import {
  Clock3,
  Loader2,
  MessageSquarePlus,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { PanelHeader } from "./PanelHeader";
import { ChatComposer } from "./ChatComposer";
import { ChatConversationList } from "./ChatConversationList";
import { ChatMessageList } from "./ChatMessageList";
import { getConversationTitle } from "../domain/chat-panel";
import { useChatPanelController } from "../hooks/use-chat-panel-controller";

/**
 * 聊天面板组件
 * 负责对接本地聊天会话、消息列表与发送能力。
 */
export const ChatPanel: FC = () => {
  const {
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
  } = useChatPanelController();

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
        <ChatConversationList
          conversations={conversations}
          selectedConversationId={selectedConversationId}
          isInitializing={isInitializing}
          onSelectConversation={handleSelectConversation}
        />

        <div className="flex min-h-0 flex-col rounded-2xl border bg-card/60">
          <div className="border-b px-4 py-3">
            <div className="text-sm font-semibold">消息</div>
            <div className="mt-1 text-xs text-muted-foreground">
              当前会在首个回复片段到达前显示骨架，占位后切换为流式文本。
            </div>
          </div>
          <ChatMessageList
            messages={renderedMessages}
            isLoadingMessages={isLoadingMessages}
            scrollToBottomToken={scrollToBottomToken}
          />

          <ChatComposer
            draft={draft}
            selectedConversation={selectedConversation}
            selectedConversationId={selectedConversationId}
            isSending={isSending}
            isCreatingConversation={isCreatingConversation}
            onDraftChange={setDraft}
            onComposerKeyDown={handleComposerKeyDown}
            onCreateConversation={handleCreateConversation}
            onSendMessage={handleSendMessage}
          />
        </div>
      </div>
    </div>
  );
};

export default ChatPanel;
