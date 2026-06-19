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

import type { MessageSSEEvent, SessionEvent } from "@/types/api/session";

export interface ChatUIMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  createdAtLabel?: string;
  status?: "streaming" | "error";
}

/**
 * isOptimisticSessionId 判断侧边栏乐观会话临时 id。
 */
export function isOptimisticSessionId(sessionId: string) {
  return sessionId.startsWith("optimistic-session-");
}

/**
 * createLocalMessageId 生成前端消息临时 id。
 */
export function createLocalMessageId(prefix: string) {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return `${prefix}-${crypto.randomUUID()}`;
  }
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

/**
 * formatEventTime 将接口时间格式化为消息时间标签。
 */
export function formatEventTime(value?: string | number | null) {
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
export function buildMessagesFromEvents(events: SessionEvent[]) {
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
