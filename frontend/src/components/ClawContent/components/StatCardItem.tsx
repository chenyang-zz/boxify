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

import { FC } from "react";
import { cn } from "@/lib/utils";

export interface StatCardItemProps {
  label: string;
  value: string;
  icon: React.ComponentType<{ className?: string }>;
  iconBgColor: string;
  iconColor: string;
}

/**
 * 状态卡片组件
 * 用于展示单个统计数据项
 */
export const StatCardItem: FC<StatCardItemProps> = ({
  label,
  value,
  icon: Icon,
  iconBgColor,
  iconColor,
}) => (
  <div className="flex flex-1 items-center gap-3  p-4 bg-card rounded-xl min-w-50">
    <div
      className={cn(
        "flex items-center justify-center size-10 rounded-lg shrink-0",
        iconBgColor,
      )}
    >
      <Icon className={cn("size-5", iconColor)} />
    </div>
    <div className="flex flex-col  items-start gap-1">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="text-sm text-nowrap font-semibold">{value}</span>
    </div>
  </div>
);

export default StatCardItem;
