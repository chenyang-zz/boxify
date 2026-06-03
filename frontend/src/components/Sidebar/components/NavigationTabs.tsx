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
import { Wrench, Boxes, Monitor } from "lucide-react";
import { cn } from "@/lib/utils";
import { useActiveView, useSidebarStore } from "../store";
import type { SidebarView } from "../types";

interface TabConfig {
  view: SidebarView;
  icon: React.ReactNode;
  label: string;
}

const tabs: TabConfig[] = [
  { view: "tools", icon: <Wrench className="size-4 shrink-0" />, label: "工具" },
  { view: "boxclaw", icon: <Boxes className="size-4 shrink-0" />, label: "BoxClaw" },
  { view: "computer", icon: <Monitor className="size-4 shrink-0" />, label: "Computer" },
];

/**
 * 导航标签组件
 * 选中的 tab 显示完整文字和白色背景，未选中的只显示图标
 */
export const NavigationTabs: FC = () => {
  const activeView = useActiveView();
  const setActiveView = useSidebarStore((state) => state.setActiveView);

  return (
    <div className="flex mx-2 mt-2 mb-1.5 rounded-xl bg-muted p-1 gap-1">
      {tabs.map((tab) => {
        const isActive = activeView === tab.view;
        return (
          <button
            key={tab.view}
            className={cn(
              "flex-1 flex items-center justify-center rounded-lg py-2 text-sm font-medium",
              "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
              isActive
                ? "bg-card text-foreground shadow-sm px-3"
                : "text-muted-foreground hover:text-foreground px-2"
            )}
            onClick={() => setActiveView(tab.view)}
            aria-label={tab.label}
          >
            {tab.icon}
            {isActive && <span className="ml-1.5 whitespace-nowrap">{tab.label}</span>}
          </button>
        );
      })}
    </div>
  );
};

export default NavigationTabs;
