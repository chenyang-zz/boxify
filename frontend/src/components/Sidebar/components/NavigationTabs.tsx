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

import { FC, useEffect } from "react";
import { BriefcaseBusiness, MessageCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { useActiveView, useSidebarStore } from "../store";
import type { SidebarView } from "../types";

interface TabConfig {
  view: SidebarView;
  icon: React.ReactNode;
  label: string;
}

const tabs: TabConfig[] = [
  {
    view: "chat",
    icon: <MessageCircle className="size-3.5 shrink-0" />,
    label: "Chat",
  },
  {
    view: "computer",
    icon: <BriefcaseBusiness className="size-3.5 shrink-0" />,
    label: "Work",
  },
];

/**
 * 导航标签组件
 * 渲染侧边栏顶部的轻量分段切换，当前只暴露 Chat 与 Work。
 */
export const NavigationTabs: FC = () => {
  const activeView = useActiveView();
  const setActiveView = useSidebarStore((state) => state.setActiveView);
  const hasVisibleActiveView = tabs.some((tab) => tab.view === activeView);

  useEffect(() => {
    if (!hasVisibleActiveView) {
      setActiveView("chat");
    }
  }, [hasVisibleActiveView, setActiveView]);

  return (
    <div className="mx-3 mt-3 mb-2 flex h-9 items-center gap-1 rounded-md border bg-muted/35 p-1">
      {tabs.map((tab) => {
        const isActive = hasVisibleActiveView && activeView === tab.view;
        return (
          <button
            key={tab.view}
            className={cn(
              "flex h-7 flex-1 items-center justify-center gap-1.5 rounded-md px-3 text-xs font-medium transition-colors",
              "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
              isActive
                ? "bg-card text-primary shadow-xs"
                : "text-muted-foreground hover:bg-card/60 hover:text-foreground",
            )}
            onClick={() => setActiveView(tab.view)}
            aria-label={tab.label}
          >
            {tab.icon}
            <span className="whitespace-nowrap">{tab.label}</span>
          </button>
        );
      })}
    </div>
  );
};

export default NavigationTabs;
