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
import { useTabsStore } from "@/store/tabs.store";
import TabContent from "./TabContent";
import { useShallow } from "zustand/react/shallow";

const Tabs: FC = () => {
  // 内容区只依赖标签数据和激活项，标题栏中的 TabBar 负责交互控制。
  const { tabs, activeTabId } = useTabsStore(
    useShallow((state) => ({
      tabs: state.tabs,
      activeTabId: state.activeTabId,
    })),
  );

  return (
    <div className="h-full w-full flex flex-col overflow-hidden ">
      {/* 标签内容区 */}
      <div className="flex-1 h-full overflow-hidden bg-background shadow-sm">
        <TabContent tabs={tabs} activeTabId={activeTabId} />
      </div>
    </div>
  );
};

export default Tabs;
