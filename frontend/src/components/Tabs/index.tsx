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
import TabBar from "./TabBar";
import TabContent from "./TabContent";
import { useShallow } from "zustand/react/shallow";

const Tabs: FC = () => {
  const { tabs, activeTabId } = useTabsStore(
    useShallow((state) => ({
      tabs: state.tabs,
      activeTabId: state.activeTabId,
    })),
  );

  const handleTabClose = (tabId: string) => {
    tabStoreMethods.closeTab(tabId);
  };

  const handleTabPin = (tabId: string) => {
    tabStoreMethods.pinTab(tabId);
  };

  const handleTabUnpin = (tabId: string) => {
    tabStoreMethods.unpinTab(tabId);
  };

  const handleTabSelect = (tabId: string) => {
    tabStoreMethods.setActiveTab(tabId);
  };

  return (
    <div className="h-full w-full flex flex-col rounded-lg overflow-hidden">
      {/* 标签栏 */}
      <TabBar
        tabs={tabs}
        activeTabId={activeTabId}
        onTabClose={handleTabClose}
        onTabPin={handleTabPin}
        onTabUnpin={handleTabUnpin}
        onTabSelect={handleTabSelect}
      />

      {/* 标签内容区 */}
      <TabContent tabs={tabs} activeTabId={activeTabId} />
    </div>
  );
};

export default Tabs;
