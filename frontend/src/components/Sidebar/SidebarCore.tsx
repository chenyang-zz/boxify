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
import { useActiveView } from "./store";
import { NavigationTabs } from "./components/NavigationTabs";
import { FileTreeView } from "./components/FileTreeView";
import { ClawMenu } from "./components/ClawMenu";

/**
 * Sidebar 核心渲染组件
 * 组合所有子组件的渲染逻辑
 */
export const SidebarCore: FC = () => {
  const activeView = useActiveView();

  return (
    <>
      <NavigationTabs />
      {activeView === "files" && <FileTreeView />}
      {activeView === "control" && <ClawMenu />}
    </>
  );
};

export default SidebarCore;
