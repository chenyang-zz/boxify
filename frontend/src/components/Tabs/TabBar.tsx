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

import { FC, useEffect, useState } from "react";
import { List, Pin } from "lucide-react";
import { TabBarProps } from "./types";
import Tab from "./Tab";
import { useHorizontalScroll } from "@/hooks/useHorizontalScroll";
import { cn } from "@/lib/utils";
import { Button } from "../ui/button";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from "../ui/dropdown-menu";

const TabBar: FC<TabBarProps> = ({
  tabs,
  activeTabId,
  onTabSelect,
  onTabClose,
  onTabPin,
  onTabUnpin,
}) => {
  const pinnedTabs = tabs.filter((t) => t.isPinned);
  const unpinnedTabs = tabs.filter((t) => !t.isPinned);
  const hasDivider = pinnedTabs.length > 0 && unpinnedTabs.length > 0;

  const scrollRef = useHorizontalScroll({ hideScrollbar: true });
  const [hasOverflow, setHasOverflow] = useState(false);

  // 检测标签栏是否溢出
  useEffect(() => {
    const container = scrollRef.current;
    if (!container) return;

    const checkOverflow = () => {
      setHasOverflow(container.scrollWidth > container.clientWidth);
    };

    const resizeObserver = new ResizeObserver(checkOverflow);
    resizeObserver.observe(container);

    checkOverflow();

    return () => resizeObserver.disconnect();
  }, [tabs.length]);

  return (
    <div className="relative flex w-full items-center bg-card border-b-2 border-border">
      {/* 标签滚动区域 */}
      <div ref={scrollRef} className="flex items-center overflow-auto flex-1">
        {/* 固定标签区域 */}
        {pinnedTabs.map((tab) => (
          <Tab
            key={tab.id}
            tab={tab}
            isActive={activeTabId === tab.id}
            onClose={onTabClose}
            onPin={onTabPin}
            onUnpin={onTabUnpin}
            onSelect={onTabSelect}
          />
        ))}

        {/* 分隔线 */}
        {hasDivider && (
          <div className="w-px shrink-0 h-6 bg-border mx-1 shadow shadow-black" />
        )}

        {/* 普通标签区域 */}
        {unpinnedTabs.map((tab) => (
          <Tab
            key={tab.id}
            tab={tab}
            isActive={activeTabId === tab.id}
            onClose={onTabClose}
            onPin={onTabPin}
            onUnpin={onTabUnpin}
            onSelect={onTabSelect}
          />
        ))}
      </div>

      {/* 标签列表下拉按钮（仅在溢出时显示） */}
      {hasOverflow && (
        <div className="shrink-0 border-l border-border">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon-sm" className="h-8 px-2">
                <List className="size-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="min-w-50 max-h-100">
              {tabs.map((tab) => (
                <DropdownMenuItem
                  key={tab.id}
                  onClick={() => onTabSelect(tab.id)}
                  className={cn(
                    "flex items-center gap-2",
                    activeTabId === tab.id && "bg-accent",
                  )}
                >
                  {tab.isPinned && <Pin className="size-3 shrink-0" />}
                  <span className="truncate">{tab.label}</span>
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )}
    </div>
  );
};

export default TabBar;
