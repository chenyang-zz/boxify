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

import { useState } from "react";
import { Check, ChevronDown, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { BashTool } from "../tool-use/BashTool";

/**
 * 工具调用项
 */
export interface ToolEvent {
  /** 工具名称 */
  name: string;
  /** 工具参数 */
  args?: Record<string, unknown>;
  /** 执行结果 */
  result?: unknown;
  /** 开始时间 */
  startTime?: string;
  /** 结束时间 */
  endTime?: string;
}

/**
 * 步骤状态
 */
export type StepStatus = "completed" | "in_progress" | "pending";

/**
 * 步骤数据
 */
export interface StepData {
  /** 步骤 ID */
  id: string;
  /** 步骤描述 */
  description: string;
  /** 步骤状态 */
  status: StepStatus;
}

/**
 * 步骤块数据项
 */
export interface StepBlockData {
  /** 步骤数据 */
  data: StepData;
  /** 工具调用列表 */
  tools?: ToolEvent[];
}

export interface StepBlockProps {
  className?: string;
  /** 步骤项数据 */
  stepItem: StepBlockData;
  /** 工具点击回调 */
  onToolClick?: (tool: ToolEvent) => void;
}

/**
 * 步骤块组件
 *
 * 展示单条计划步骤，支持折叠/展开查看下属工具调用。
 * - 标题行：状态图标 + 步骤描述 + 展开/收起箭头
 * - 展开后：左侧虚线时间线 + 工具列表
 */
export function StepBlock({
  className,
  stepItem,
  onToolClick,
}: StepBlockProps) {
  const [expanded, setExpanded] = useState(false);
  const { data, tools = [] } = stepItem;
  const isCompleted = data.status === "completed";
  const isInProgress = data.status === "in_progress";

  const handleToggle = () => setExpanded((prev) => !prev);

  return (
    <div className={cn("flex flex-col pl-7", className)}>
      {/* 标题行 */}
      <button
        type="button"
        onClick={handleToggle}
        className="group/header flex cursor-pointer items-center  gap-2 rounded-md py-1 text-sm text-foreground text-left"
      >
        <div className="flex min-w-0 flex-row items-center justify-start gap-2 truncate">
          {/* 状态图标 */}
          <StepStatusIcon status={data.status} />

          {/* 步骤描述 */}
          <div className="min-w-0 flex-1 truncate font-medium">
            {data.description}
          </div>

          {/* 展开/收起箭头 */}
          <ChevronDown
            className={cn(
              "size-4 shrink-0 text-muted-foreground transition-transform duration-150",
              expanded && "rotate-180",
            )}
          />
        </div>
      </button>

      {/* 工具列表 */}
      {expanded && tools.length > 0 && (
        <div className="flex">
          {/* 左侧虚线时间线 */}
          <div className="relative w-6 shrink-0">
            <div className="absolute bottom-0 left-[7px] top-2 w-px border-l border-dashed border-border" />
          </div>

          {/* 工具项 */}
          <div className="flex min-w-0 flex-1 flex-col gap-3 overflow-hidden pt-2 transition-[max-height,opacity] duration-150 ease-in-out">
            {tools.map((tool, idx) => (
              <ToolRow
                key={`${data.id}-tool-${idx}`}
                tool={tool}
                onClick={onToolClick ? () => onToolClick(tool) : undefined}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

/** 步骤状态图标 */
function StepStatusIcon({ status }: { status: StepStatus }) {
  if (status === "completed") {
    return (
      <div className="flex size-4 shrink-0 items-center justify-center rounded-full border bg-primary">
        <Check className="size-3 text-primary-foreground" />
      </div>
    );
  }

  if (status === "in_progress") {
    return (
      <div className="flex size-4 shrink-0 items-center justify-center rounded-full border border-primary bg-primary/10">
        <Loader2 className="size-3 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="flex size-4 shrink-0 items-center justify-center rounded-full border bg-muted" />
  );
}

/** 单条工具行 */
function ToolRow({ tool, onClick }: { tool: ToolEvent; onClick?: () => void }) {
  if (tool.name === "bash") {
    const command =
      typeof tool.args?.command === "string" ? tool.args.command : "bash";
    return (
      <div className="flex items-center justify-between gap-2">
        <BashTool label={command} onClick={onClick} />
      </div>
    );
  }

  return (
    <div
      className={cn(
        "flex items-center justify-between gap-2",
        onClick && "cursor-pointer",
      )}
      onClick={onClick}
    >
      <div className="flex min-w-0 items-center gap-2">
        <span className="truncate text-sm text-muted-foreground">
          {tool.name}
        </span>
      </div>
    </div>
  );
}

export default StepBlock;
