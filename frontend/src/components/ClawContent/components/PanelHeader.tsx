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

import { FC, ReactNode } from "react";
import { cn } from "@/lib/utils";

export interface PanelHeaderProps {
  title: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  className?: string;
  titleClassName?: string;
  descriptionClassName?: string;
  contentClassName?: string;
  align?: "center" | "start";
}

/**
 * 面板标题组件
 * 统一渲染 panel 的标题、副标题与右侧操作区
 */
export const PanelHeader: FC<PanelHeaderProps> = ({
  title,
  description,
  actions,
  className,
  titleClassName,
  descriptionClassName,
  contentClassName,
  align = "center",
}) => {
  return (
    <div
      className={cn(
        "flex justify-between gap-4",
        align === "start" ? "items-start" : "items-center",
        className,
      )}
    >
      <div className={cn("flex flex-col gap-1 text-left", contentClassName)}>
        <h1 className={cn("text-xl font-bold", titleClassName)}>{title}</h1>
        {description && (
          <p className={cn("text-sm text-muted-foreground", descriptionClassName)}>
            {description}
          </p>
        )}
      </div>
      {actions}
    </div>
  );
};

export default PanelHeader;
