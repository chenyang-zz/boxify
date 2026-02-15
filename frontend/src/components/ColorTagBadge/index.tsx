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
import { TAG_COLORS } from "@/constants/color-constants";

interface ColorTagBadgeProps {
  /** 颜色值 */
  color: string;
  /** 自定义类名 */
  className?: string;
}

/**
 * 颜色标签徽章组件
 *
 * 显示选中颜色的预览徽章，使用对应颜色的浅色背景和深色文字
 */
const ColorTagBadge: FC<ColorTagBadgeProps> = ({ color, className }) => {
  const colorConfig = TAG_COLORS.find((c) => c.value === color);

  if (!colorConfig) {
    return null;
  }

  return (
    <span
      className={cn(
        "inline-flex items-center justify-center rounded-full px-2 py-0.5 text-xs font-medium",
        colorConfig.lightBg,
        colorConfig.textColor,
        className,
      )}
    >
      {colorConfig.label}
    </span>
  );
};

export default ColorTagBadge;
