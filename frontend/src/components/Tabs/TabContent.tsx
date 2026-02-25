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

import { Activity, FC } from "react";
import { TabContentProps } from "./types";
import { TabType } from "@/common/constrains";
import DBTable from "../DBTable";
import Terminal from "../Terminal";

const TabContent: FC<TabContentProps> = ({ tabs, activeTabId }) => {
  const activeTab = tabs.find((t) => t.id === activeTabId);

  return (
    <div className="flex-1 overflow-hidden">
      {/* React 19 Activity: 所有标签都渲染，使用 hidden 隐藏非活动标签
          这样每个标签的组件状态（滚动、分页等）都会被保留 */}
      {tabs.map((tab) => (
        <Activity
          key={tab.propertyUuid}
          mode={tab.id === activeTabId ? "visible" : "hidden"}
        >
          {tab.type === TabType.TABLE ? (
            <DBTable key={tab.propertyUuid} sessionId={tab.propertyUuid} />
          ) : tab.type === TabType.TERMINAL ? (
            <Terminal key={tab.propertyUuid} sessionId={tab.propertyUuid} />
          ) : null}
        </Activity>
      ))}

      {/* 空状态 */}
      {!activeTab && (
        <div className="h-full flex items-center justify-center text-muted-foreground">
          <p>请从左侧选择一个表或视图</p>
        </div>
      )}
    </div>
  );
};

export default TabContent;
