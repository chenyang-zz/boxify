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

import { type FC, type RefObject } from "react";
import {
  Copy,
  CornerUpRight,
  Loader2,
  MessageCircle,
  RotateCcw,
  ThumbsDown,
  ThumbsUp,
} from "lucide-react";
import { Streamdown } from "streamdown";
import "streamdown/styles.css";
import { cjk } from "@streamdown/cjk";
import { code } from "@streamdown/code";
import { math } from "@streamdown/math";
import { mermaid } from "@streamdown/mermaid";

import type { ChatUIMessage } from "../domain/messages";

interface ChatMessageListProps {
  messages: ChatUIMessage[];
  loadingSession: boolean;
  selectedSessionId: string;
  selectedSessionPending: boolean;
  errorMessage: string;
  streamingAssistantId: string;
  messagesEndRef: RefObject<HTMLDivElement | null>;
}

const assistantActions = [
  { label: "复制", icon: Copy },
  { label: "赞", icon: ThumbsUp },
  { label: "踩", icon: ThumbsDown },
  { label: "重新生成", icon: RotateCcw },
  { label: "展开", icon: CornerUpRight },
];

/**
 * ChatMessageList 渲染聊天消息列表、空态和助手消息操作栏。
 */
export const ChatMessageList: FC<ChatMessageListProps> = ({
  messages,
  loadingSession,
  selectedSessionId,
  selectedSessionPending,
  errorMessage,
  streamingAssistantId,
  messagesEndRef,
}) => {
  return (
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
                  <div className="mt-2 text-xs text-destructive">发送失败</div>
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
  );
};

export default ChatMessageList;
