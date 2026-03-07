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

import { create } from "zustand";
import type { SidebarView } from "../types";
import {
  DEFAULT_EXPANDED_CATEGORIES,
  DEFAULT_SELECTED_ITEM,
} from "../domain";
import {
  selectActiveView,
  selectSelectedMenuItem,
  selectExpandedCategories,
} from "./selectors";

/**
 * Sidebar Store 状态接口
 */
interface SidebarStore {
  // 状态
  activeView: SidebarView;
  selectedMenuItem: string;
  expandedCategories: Set<string>;

  // Actions
  setActiveView: (view: SidebarView) => void;
  selectMenuItem: (itemId: string) => void;
  toggleCategory: (categoryId: string) => void;
  expandCategory: (categoryId: string) => void;
  collapseCategory: (categoryId: string) => void;
  reset: () => void;
}

/**
 * 初始状态
 */
const initialState = {
  activeView: "files" as SidebarView,
  selectedMenuItem: DEFAULT_SELECTED_ITEM,
  expandedCategories: DEFAULT_EXPANDED_CATEGORIES,
};

/**
 * Sidebar 全局状态 Store
 */
export const useSidebarStore = create<SidebarStore>((set) => ({
  ...initialState,

  setActiveView: (view: SidebarView) => {
    set({ activeView: view });
  },

  selectMenuItem: (itemId: string) => {
    set({ selectedMenuItem: itemId });
  },

  toggleCategory: (categoryId: string) => {
    set((state) => {
      const next = new Set(state.expandedCategories);
      if (next.has(categoryId)) {
        next.delete(categoryId);
      } else {
        next.add(categoryId);
      }
      return { expandedCategories: next };
    });
  },

  expandCategory: (categoryId: string) => {
    set((state) => {
      const next = new Set(state.expandedCategories);
      next.add(categoryId);
      return { expandedCategories: next };
    });
  },

  collapseCategory: (categoryId: string) => {
    set((state) => {
      const next = new Set(state.expandedCategories);
      next.delete(categoryId);
      return { expandedCategories: next };
    });
  },

  reset: () => {
    set(initialState);
  },
}));

// 选择器 hooks
export function useActiveView(): SidebarView {
  return useSidebarStore(selectActiveView);
}

export function useSelectedMenuItem(): string {
  return useSidebarStore(selectSelectedMenuItem);
}

export function useExpandedCategories(): Set<string> {
  return useSidebarStore(selectExpandedCategories);
}
