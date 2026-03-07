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

import { SidebarCore } from "./SidebarCore";

/**
 * Sidebar 组件主入口
 */
const Sidebar = () => {
  return <SidebarCore />;
};

export default Sidebar;

// 导出子组件和 hooks
export { SidebarCore } from "./SidebarCore";
export { NavigationTabs } from "./components/NavigationTabs";
export { FileTreeView } from "./components/FileTreeView";
export { ClawMenu } from "./components/ClawMenu";
export * from "./store";
export * from "./domain";
export * from "./types";
