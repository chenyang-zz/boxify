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
import { Label } from "@/components/ui/label";

interface SettingFieldProps {
  /** 字段唯一标识，用于 label 的 htmlFor */
  id: string;
  /** 字段标题 */
  label: string;
  /** 字段描述 */
  description?: string;
  /** 右侧控件 */
  children: ReactNode;
}

/**
 * 设置表单字段 — label + description + control 的标准布局
 *
 * 左侧：标题和描述（自适应占据剩余空间）
 * 右侧：控件（固定宽度，不压缩）
 */
export const SettingField: FC<SettingFieldProps> = ({
  id,
  label,
  description,
  children,
}) => {
  return (
    <div className="flex items-start gap-6">
      <div className="flex-1 min-w-0">
        <Label htmlFor={id} className="text-sm font-medium text-foreground block text-left">
          {label}
        </Label>
        {description && (
          <p className="text-sm text-muted-foreground mt-1.5 leading-relaxed text-left">
            {description}
          </p>
        )}
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  );
};

export default SettingField;
