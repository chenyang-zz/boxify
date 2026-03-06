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

import { TabBarProps } from "./types";
import Tab from "./Tab";
import { useHorizontalScroll } from "@/hooks/useHorizontalScroll";

import { tabStoreMethods } from "@/store/tabs.store";

const TabBar: FC<TabBarProps> = ({ tabs, activeTabId, onTabSelect }) => {
  const scrollRef = useHorizontalScroll({ hideScrollbar: true });

  // 关闭指定标签，避免在 Tab 子组件内耦合 store 访问。
  const handleTabClose = (tabId: string) => {
    tabStoreMethods.closeTab(tabId);
  };

  return (
    <div className="relative flex w-full items-center">
      {/* 标签滚动区域 */}
      <div
        ref={scrollRef}
        className="flex h-9 flex-1 items-center overflow-hidden rounded-none border-l border-r border-border/30"
      >
        {tabs.map((tab) => (
          <Tab
            key={tab.id}
            tab={tab}
            isActive={activeTabId === tab.id}
            onSelect={onTabSelect}
            onClose={handleTabClose}
          />
        ))}
      </div>
    </div>
  );
};

export default TabBar;
