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
import { FC, ReactNode } from "react";

export interface CardItemProps {
  label: string;
  badge?: React.ReactNode;
  icon?: React.ReactNode;
  description?: ReactNode;
  action?: React.ReactNode;
}

/**
 * 通用卡片组件
 */
export const CardItem: FC<CardItemProps> = ({
  label,
  icon,
  badge,
  description,
  action,
}) => {
  return (
    <div
      className={cn(
        "flex items-center flex-1 gap-4 p-4 bg-card rounded-lg  min-w-80 overflow-hidden",
      )}
    >
      {icon}
      <div className="flex flex-col justify-between flex-1  overflow-hidden">
        <div className="flex items-center gap-2 justify-between">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium">{label}</span>
            {badge}
          </div>

          {action && <div className="shrink-0">{action}</div>}
        </div>
        {description && (
          <div className="text-xs leading-5 text-muted-foreground text-left line-clamp-2">
            {description}
          </div>
        )}
      </div>
    </div>
  );
};

export default CardItem;
