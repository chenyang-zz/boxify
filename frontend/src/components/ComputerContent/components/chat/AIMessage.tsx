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
import { Bot } from "lucide-react";

/**
 * AI 消息数据
 */
export interface AIMessageData {
  /** 消息内容 */
  message: string;
}

export interface AIMessageProps {
  className?: string;
  /** 消息数据 */
  data: AIMessageData;
  /** 创建时间 */
  createdAt?: string;
}

/**
 * AI 消息组件
 *
 * 左对齐，带 AI 头像和名称，内容使用卡片样式展示。
 * 支持 markdown 内容渲染（后续扩展）。
 */
export function AIMessage({ className, data, createdAt }: AIMessageProps) {
  const timeLabel = createdAt ? formatRelativeDate(createdAt) : undefined;

  return (
    <div className={cn("group flex w-full flex-col gap-2", className)}>
      {/* 头部：AI 标识 + hover 时间 */}
      <div className="flex items-center justify-between relative">
        <div className="flex items-center gap-1.5 text-foreground">
          <div className="flex size-6 items-center justify-center rounded-md bg-primary/10">
            <Bot className="size-4 text-primary" />
          </div>
          <span className="text-xs font-medium">Boxify</span>
        </div>
        {/* 时间标签 — hover 时显示 */}
        {timeLabel && (
          <span
            className={cn(
              " absolute right-0  text-xs text-muted-foreground transition-opacity duration-150",
              " opacity-0 group-hover:opacity-100",
            )}
          >
            {timeLabel}
          </span>
        )}
      </div>

      {/* 消息内容 */}
      <div className="max-w-none pl-7 text-foreground">
        <div className="text-sm leading-relaxed whitespace-pre-wrap">
          {data.message}
        </div>
      </div>
    </div>
  );
}

export default AIMessage;
