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

/**
 * Chat 消息组件统一导出
 *
 * 每种消息类型独立成组件，便于后续扩展：
 * - UserMessage: 人类消息（右对齐气泡）
 * - AIMessage: AI 消息（左对齐，带头像）
 * - AttachmentsMessage: 附件消息（用户/AI 通用）
 * - ChatMessage: 统一分发器，根据 kind 渲染对应组件
 */

export {
  UserMessage,
  type UserMessageProps,
  type UserMessageData,
} from "./UserMessage";

export {
  AIMessage,
  type AIMessageProps,
  type AIMessageData,
} from "./AIMessage";

export {
  AttachmentsMessage,
  type AttachmentsMessageProps,
  type AttachmentFile,
} from "./AttachmentsMessage";

export {
  ChatMessage,
  type ChatMessageProps,
  type ChatMessageItem,
  type ChatMessageKind,
} from "./ChatMessage";

export {
  StepBlock,
  type StepBlockProps,
  type StepBlockData,
  type StepData,
  type ToolEvent,
  type StepStatus,
} from "./StepBlock";
