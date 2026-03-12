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

import { FC, useState } from "react";
import {
  Clock3,
  Loader2,
  MessageSquarePlus,
  PanelRightOpen,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
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
  const [isConversationDrawerOpen, setIsConversationDrawerOpen] =
    useState(false);
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

  /**
   * 切换会话后收起抽屉，避免内容区被持续遮挡。
   */
  async function handleSelectConversationFromDrawer(conversationId: string) {
    await handleSelectConversation(conversationId);
    setIsConversationDrawerOpen(false);
  }

  return (
    <div className="flex h-full w-full min-h-0 flex-col">
      {/* <PanelHeader
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
            <Sheet
              open={isConversationDrawerOpen}
              onOpenChange={setIsConversationDrawerOpen}
            >
              <SheetTrigger asChild>
                <Button
                  variant="outline"
                  size="sm"
                  className="h-10 px-4"
                  disabled={isInitializing}
                >
                  <PanelRightOpen className="size-4" />
                  会话列表
                </Button>
              </SheetTrigger>
              <SheetContent
                side="right"
                className="w-[min(92vw,28rem)] p-0 sm:max-w-md"
              >
                <SheetHeader>
                  <SheetTitle>会话列表</SheetTitle>
                  <SheetDescription>
                    当前只保存在 Boxify 进程内存中，可随时切换查看上下文。
                  </SheetDescription>
                </SheetHeader>
                <div className="min-h-0 flex-1 p-4">
                  <ChatConversationList
                    className="h-full rounded-xl"
                    conversations={conversations}
                    selectedConversationId={selectedConversationId}
                    isInitializing={isInitializing}
                    onSelectConversation={handleSelectConversationFromDrawer}
                  />
                </div>
              </SheetContent>
            </Sheet>
            <Separator orientation="vertical" className="mx-1 h-6" />
          </div>
        }
      /> */}

      <div className="flex min-h-0 flex-1 flex-col  relative">
        <ChatMessageList
          messages={renderedMessages}
          isLoadingMessages={isLoadingMessages}
          scrollToBottomToken={scrollToBottomToken}
        />

        <div className="pointer-events-none absolute inset-x-0 bottom-0 h-56 bg-linear-to-b from-background/0  via-background/95 to-background" />

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
  );
};

export default ChatPanel;
