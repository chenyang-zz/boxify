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

import type { SidebarView } from "../types";

/**
 * Sidebar Store 选择器状态接口
 */
interface SidebarSelectorState {
  activeView: SidebarView;
  selectedMenuItem: string;
  expandedCategories: Set<string>;
}

/**
 * 选择当前视图
 */
export function selectActiveView(state: SidebarSelectorState): SidebarView {
  return state.activeView;
}

/**
 * 选择当前选中的菜单项
 */
export function selectSelectedMenuItem(state: SidebarSelectorState): string {
  return state.selectedMenuItem;
}

/**
 * 选择展开的分类集合
 */
export function selectExpandedCategories(
  state: SidebarSelectorState,
): Set<string> {
  return state.expandedCategories;
}

/**
 * 判断分类是否展开
 */
export function selectIsCategoryExpanded(categoryId: string) {
  return (state: SidebarSelectorState): boolean =>
    state.expandedCategories.has(categoryId);
}

/**
 * 判断菜单项是否选中
 */
export function selectIsMenuItemSelected(itemId: string) {
  return (state: SidebarSelectorState): boolean =>
    state.selectedMenuItem === itemId;
}
