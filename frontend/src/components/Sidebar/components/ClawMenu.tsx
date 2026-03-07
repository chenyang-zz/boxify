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
import { ChevronDown, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  useSelectedMenuItem,
  useExpandedCategories,
  useSidebarStore,
} from "../store";
import { menuCategories } from "../domain";

/**
 * Claw 菜单组件
 * 用于显示聊天和控制菜单
 */
export const ClawMenu: FC = () => {
  const selectedMenuItem = useSelectedMenuItem();
  const expandedCategories = useExpandedCategories();
  const { selectMenuItem, toggleCategory } = useSidebarStore();

  return (
    <div className="h-full w-full p-1 flex flex-col overflow-auto">
      {menuCategories.map((category) => {
        const isExpanded = expandedCategories.has(category.id);
        return (
          <div key={category.id} className="mb-1">
            {/* Category Header */}
            <button
              onClick={() => toggleCategory(category.id)}
              className="w-full flex items-center gap-1 px-2 py-1.5 text-xs font-semibold text-muted-foreground hover:text-foreground hover:bg-accent/50 rounded transition-colors"
            >
              {isExpanded ? (
                <ChevronDown className="size-3" />
              ) : (
                <ChevronRight className="size-3" />
              )}
              <span>{category.label}</span>
            </button>

            {/* Category Items */}
            {isExpanded && (
              <div className="ml-2">
                {category.items.map((item) => {
                  const isSelected = selectedMenuItem === item.id;
                  const Icon = item.icon;
                  return (
                    <button
                      key={item.id}
                      onClick={() => selectMenuItem(item.id)}
                      className={cn(
                        "w-full flex items-center gap-2 px-2 py-1.5 text-sm rounded transition-colors",
                        isSelected
                          ? "bg-primary/20  text-foreground"
                          : "text-muted-foreground hover:text-foreground hover:bg-accent/50",
                      )}
                    >
                      <Icon
                        className={cn("size-4", isSelected && "text-primary")}
                      />
                      <span>{item.label}</span>
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
};

export default ClawMenu;
