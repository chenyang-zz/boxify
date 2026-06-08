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
import { Check, ChevronDown, ChevronUp, Clock, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";

/**
 * 计划步骤状态
 */
export type PlanStepStatus = "completed" | "in_progress" | "pending";

/**
 * 计划步骤
 */
export interface PlanStep {
  /** 步骤唯一标识 */
  id: string;
  /** 步骤描述 */
  description: string;
  /** 步骤状态 */
  status: PlanStepStatus;
}

export interface PlanPanelProps {
  className?: string;
  /** 计划步骤列表 */
  steps?: PlanStep[];
}

/**
 * 计划面板组件
 *
 * 展示任务执行计划的步骤列表，支持折叠/展开。
 * - 折叠态：显示最新步骤 + 完成进度
 * - 展开态：显示完整步骤列表
 */
export function PlanPanel({
  className,
  steps: stepsProp = [],
}: PlanPanelProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const steps = stepsProp;

  // 无步骤时不渲染
  if (steps.length === 0) return null;

  const completedCount = steps.filter((s) => s.status === "completed").length;
  const totalCount = steps.length;

  // 折叠态展示正在执行的步骤；没有则展示最后一个完成的步骤
  const activeStep =
    steps.find((s) => s.status === "in_progress") ??
    steps.findLast((s) => s.status === "completed");

  return (
    <Card
      className={cn(
        "overflow-hidden transition-shadow py-3 gap-0 ",
        isExpanded && "shadow-sm",
        className,
      )}
    >
      {/* 折叠状态 */}
      {!isExpanded && (
        <button
          type="button"
          aria-expanded="false"
          onClick={() => setIsExpanded(true)}
          className="flex w-full items-center justify-between gap-2 px-4 text-left transition-colors"
        >
          {/* 左侧：最新步骤 */}
          <div className="flex min-w-0 items-center gap-2.5">
            <Loader2 className="size-4 shrink-0 text-primary animate-spin" />
            <span className="truncate text-sm text-muted-foreground">
              {activeStep?.description ?? "暂无步骤"}
            </span>
          </div>

          {/* 右侧：进度 + 展开图标 */}
          <div className="flex shrink-0 items-center gap-2">
            <Badge variant="secondary" className="text-xs font-normal">
              {completedCount} / {totalCount}
            </Badge>
            <ChevronUp className="size-4 text-muted-foreground" />
          </div>
        </button>
      )}

      {/* 展开状态 */}
      {isExpanded && (
        <>
          <CardHeader className="flex-row items-start justify-end px-3 ">
            <Button
              type="button"
              aria-expanded="true"
              aria-label="收起计划面板"
              onClick={() => setIsExpanded(false)}
              variant="ghost"
              size="icon-xs"
            >
              <ChevronDown className="size-4" />
            </Button>
          </CardHeader>

          <CardContent className="px-4 pb-1">
            <div className="rounded-lg bg-muted/30 px-4 py-4 pb-3">
              {/* 标题栏 */}
              <div className="flex items-center justify-between pl-2 pb-2">
                <span className="text-sm font-semibold text-foreground">
                  任务进度
                </span>
                <Badge variant="secondary" className="text-xs font-normal">
                  {completedCount} / {totalCount}
                </Badge>
              </div>

              {/* 步骤列表 */}
              <div>
                {steps.map((step, index) => (
                  <div key={step.id}>
                    <StepItem step={step} />
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </>
      )}
    </Card>
  );
}

/** 单条步骤项 */
function StepItem({ step }: { step: PlanStep }) {
  return (
    <div className="flex items-center gap-2.5 px-2 py-1 ">
      <StepStatusIcon status={step.status} />
      <span
        className={cn(
          "truncate text-xs",
          step.status === "completed" && "text-muted-foreground line-through",
          step.status === "in_progress" && "font-medium text-foreground",
          step.status === "pending" && "text-muted-foreground",
        )}
      >
        {step.description}
      </span>
    </div>
  );
}

/** 步骤状态图标 */
function StepStatusIcon({ status }: { status: PlanStepStatus }) {
  const iconClass = "size-4 shrink-0";

  if (status === "completed") {
    return <Check className={cn(iconClass, "text-primary")} />;
  }
  if (status === "in_progress") {
    return <Loader2 className={cn(iconClass, "animate-spin text-primary")} />;
  }
  return <Clock className={cn(iconClass, "text-muted-foreground")} />;
}

export default PlanPanel;
