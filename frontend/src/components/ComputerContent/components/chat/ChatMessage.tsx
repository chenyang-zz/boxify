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

import { cn } from "@/lib/utils";
import { UserMessage, UserMessageData } from "./UserMessage";
import { AIMessage, AIMessageData } from "./AIMessage";

/**
 * 消息类型
 */
export type ChatMessageKind = "user" | "assistant";

/**
 * 统一消息项
 */
export interface ChatMessageItem {
  /** 消息唯一标识 */
  id: string;
  /** 消息类型 */
  kind: ChatMessageKind;
  /** 消息数据 */
  data: UserMessageData | AIMessageData;
  /** 附件列表（可选，仅 user 类型） */
  attachments?: string[];
  /** 创建时间 */
  createdAt?: string;
}

export interface ChatMessageProps {
  className?: string;
  /** 消息项 */
  item: ChatMessageItem;
}

/**
 * 聊天消息统一分发器
 *
 * 根据 item.kind 自动渲染对应的消息组件：
 * - "user" → UserMessage（人类消息）
 * - "assistant" → AIMessage（AI 消息）
 */
export function ChatMessage({ className, item }: ChatMessageProps) {
  if (item.kind === "user") {
    return (
      <div className={cn("mt-4", className)}>
        <UserMessage
          data={item.data as UserMessageData}
          attachments={item.attachments}
          createdAt={item.createdAt}
        />
      </div>
    );
  }

  if (item.kind === "assistant") {
    return (
      <div className={cn("mt-4", className)}>
        <AIMessage
          data={item.data as AIMessageData}
          createdAt={item.createdAt}
        />
      </div>
    );
  }

  return null;
}

export default ChatMessage;
