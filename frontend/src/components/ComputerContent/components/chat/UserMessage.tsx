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

import { cn, formatRelativeDate } from "@/lib/utils";

/**
 * 用户消息数据
 */
export interface UserMessageData {
  /** 消息内容 */
  message: string;
}

export interface UserMessageProps {
  className?: string;
  /** 消息数据 */
  data: UserMessageData;
  /** 附件列表（可选） */
  attachments?: string[];
  /** 创建时间 */
  createdAt?: string;
}

/**
 * 人类消息组件
 *
 * 右对齐气泡，白色背景带边框。
 */
export function UserMessage({
  className,
  data,
  attachments,
  createdAt,
}: UserMessageProps) {
  const timeLabel = createdAt ? formatRelativeDate(createdAt) : undefined;

  return (
    <div
      className={cn("group flex w-auto flex-col items-end gap-1", className)}
    >
      <div className="relative flex max-w-[90%] flex-col items-end gap-2">
        {/* 时间标签 — hover 时显示 */}
        {timeLabel && (
          <span
            className={cn(
              "shrink-0 text-xs text-muted-foreground text-right transition-opacity duration-150",
              "opacity-0 group-hover:opacity-100",
            )}
          >
            {timeLabel}
          </span>
        )}

        {/* 消息气泡 */}
        <div className="text-foreground flex items-center rounded-lg border bg-card p-3 text-sm shadow-sm">
          {data.message}
        </div>

        {/* 附件预览 */}
        {attachments && attachments.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {attachments.map((file, idx) => (
              <div
                key={idx}
                className="flex items-center gap-1.5 rounded-md border bg-muted px-2 py-1 text-xs text-muted-foreground"
              >
                <span className="truncate max-w-30">{file}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default UserMessage;
