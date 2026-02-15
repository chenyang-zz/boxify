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
import { Check, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { TAG_COLORS } from "@/constants/color-constants";

interface ColorTagSelectorProps {
  /** 当前选中的颜色值 */
  value: string;
  /** 颜色变化回调 */
  onChange: (color: string) => void;
  /** 自定义类名 */
  className?: string;
}

/**
 * 颜色标签选择器组件
 *
 * 提供一组预定义的颜色色块供用户选择，支持点击选择和键盘导航
 * 支持取消选择
 */
const ColorTagSelector: FC<ColorTagSelectorProps> = ({
  value,
  onChange,
  className,
}) => {
  return (
    <div className={cn("flex items-center gap-1.5 w-full", className)}>
      {/* 8种颜色按钮 */}
      <div className="flex gap-1.5 flex-1">
        {TAG_COLORS.map((color) => {
          const isSelected = value === color.value;

          return (
            <button
              key={color.value}
              type="button"
              className={cn(
                "relative size-4 rounded-full transition-all duration-200 shrink-0",
                "focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2",
                color.ringClass,
                color.bgClass,
                color.hoverClass,
                "hover:scale-110 hover:shadow-md",
              )}
              onClick={() => onChange(color.value)}
              aria-label={color.label}
              aria-pressed={isSelected}
              title={color.label}
            >
              {isSelected && (
                <Check className="size-3 text-white absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2" />
              )}
            </button>
          );
        })}
      </div>

      {/* 取消按钮 */}
      <button
        type="button"
        className={cn(
          "relative size-4 rounded-full transition-all duration-200 shrink-0",
          "focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2",
          "bg-muted hover:bg-muted-foreground/20",
          "hover:scale-110 hover:shadow-md",
          "flex items-center justify-center",
        )}
        onClick={() => onChange("")}
        aria-label="取消选择"
        title="取消选择"
      >
        <X className="size-3 text-muted-foreground" />
      </button>
    </div>
  );
};

export default ColorTagSelector;
