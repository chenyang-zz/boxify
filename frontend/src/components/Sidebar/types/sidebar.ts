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

import type { FC } from "react";

/**
 * Sidebar 视图类型
 */
export type SidebarView = "files" | "control";

/**
 * 菜单项接口
 */
export interface MenuItem {
  id: string;
  label: string;
  icon: FC<{ className?: string }>;
}

/**
 * 菜单分类接口
 */
export interface MenuCategory {
  id: string;
  label: string;
  items: MenuItem[];
}

/**
 * Sidebar 状态接口
 */
export interface SidebarState {
  activeView: SidebarView;
  selectedMenuItem: string;
  expandedCategories: Set<string>;
}
