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

import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

export interface ToolBadgeProps {
  /** 图标组件 */
  icon: LucideIcon;
  /** 徽章文字 */
  label: string;
  /** 点击回调 */
  onClick?: () => void;
  className?: string;
}

/**
 * 工具徽章组件
 *
 * 用于展示单个工具调用的缩略标签，带图标和名称。
 */
export function ToolBadge({
  icon: Icon,
  label,
  onClick,
  className,
}: ToolBadgeProps) {
  return (
    <div
      role={onClick ? "button" : undefined}
      tabIndex={onClick ? 0 : undefined}
      onClick={onClick}
      onKeyDown={
        onClick
          ? (e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                onClick();
              }
            }
          : undefined
      }
      className={cn(
        "inline-flex max-w-full min-w-0 items-center gap-1.5 rounded-lg border border-border bg-muted px-2.5 py-1 text-sm text-foreground transition-colors",
        onClick && "cursor-pointer hover:bg-accent/50",
        className,
      )}
    >
      <span className="flex shrink-0 items-center justify-center text-muted-foreground">
        <Icon className="size-4 shrink-0" />
      </span>
      <span className="max-w-30 truncate">{label}</span>
    </div>
  );
}

export default ToolBadge;
