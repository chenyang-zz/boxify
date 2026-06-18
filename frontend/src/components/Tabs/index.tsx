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
import { tabStoreMethods, useTabsStore } from "@/store/tabs.store";
import TabContent from "./TabContent";
import { useShallow } from "zustand/react/shallow";
import TabBar from "./TabBar";

const Tabs: FC = () => {
  // 内容区托管标签数据，保证顶栏独立后仍可完整操作标签。
  const { tabs, activeTabId } = useTabsStore(
    useShallow((state) => ({
      tabs: state.tabs,
      activeTabId: state.activeTabId,
    })),
  );

  // 切换当前激活标签。
  const handleTabSelect = (tabId: string) => {
    tabStoreMethods.setActiveTab(tabId);
  };

  return (
    <div className="flex h-full w-full flex-col overflow-hidden">
      {tabs.length > 0 && (
        <div className="shrink-0 border-b bg-card px-2">
          <TabBar
            tabs={tabs}
            activeTabId={activeTabId}
            onTabSelect={handleTabSelect}
          />
        </div>
      )}
      {/* 标签内容区 */}
      <div className="h-full flex-1 overflow-hidden bg-background shadow-sm">
        <TabContent tabs={tabs} activeTabId={activeTabId} />
      </div>
    </div>
  );
};

export default Tabs;
