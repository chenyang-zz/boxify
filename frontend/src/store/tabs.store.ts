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
import { PropertyItemType } from "./property.store";
import { v4 as uuid } from "uuid";
import { TabType } from "@/common/constrains";
import { propertyTypeToTabType } from "@/lib/property";
import { StoreMethods } from "./common";
import { terminalManager } from "@/lib/terminal-manager";

// 标签页状态接口
export interface TabState {
  id: string; // 唯一 ID（使用 UUID）
  propertyUuid: string; // 关联的 PropertyItem UUID
  label: string; // 显示名称
  type: TabType; // 类型（table、view、query 等）
  isPinned: boolean; // 是否固定
}

interface TabsState {
  // 状态
  tabs: TabState[];
  activeTabId: string | null;

  // Actions
  openTab: (propertyItem: PropertyItemType) => string; // 返回 tabId
  closeTab: (tabId: string) => void;
  closeOtherTabs: (tabId: string) => void;
  closeAllTabs: () => void;
  closeTabsToRight: (tabId: string) => void;
  setActiveTab: (tabId: string) => void;
  pinTab: (tabId: string) => void;
  unpinTab: (tabId: string) => void;
  renameTab: (tabId: string, newLabel: string) => void;
  moveTab: (fromIndex: number, toIndex: number) => void;

  // 工具方法
  getTabByUUID: (uuid: string) => TabState | undefined;
  getActiveTab: () => TabState | undefined;
}

export const useTabsStore = create<TabsState>((set, get) => ({
  tabs: [],
  activeTabId: null,

  // 打开标签页（智能判断：存在则切换，不存在则新建）
  openTab: (propertyItem: PropertyItemType) => {
    const state = get();
    const existingTab = state.tabs.find(
      (t) => t.propertyUuid === propertyItem.uuid,
    );

    if (existingTab) {
      // 已存在，切换到该标签
      set({ activeTabId: existingTab.id });
      return existingTab.id;
    }

    // 不存在，新建标签
    const newTab: TabState = {
      id: `tab-${uuid()}`, // 使用 UUID 生成唯一 ID
      propertyUuid: propertyItem.uuid,
      label: propertyItem.label,
      type: propertyTypeToTabType(propertyItem.type),
      isPinned: false,
    };

    set((state) => ({
      tabs: [...state.tabs, newTab],
      activeTabId: newTab.id,
    }));

    return newTab.id;
  },

  // 关闭标签
  closeTab: (tabId: string) => {
    set((state) => {
      const closedTab = state.tabs.find((t) => t.id === tabId);
      const newTabs = state.tabs.filter((t) => t.id !== tabId);
      let newActiveTabId = state.activeTabId;

      // 如果关闭的是终端标签，销毁对应的终端实例
      if (closedTab?.type === TabType.TERMINAL) {
        terminalManager.destroy(closedTab.propertyUuid);
      }

      // 如果关闭的是当前激活标签，切换到最后一个标签
      if (state.activeTabId === tabId && newTabs.length > 0) {
        // 优先选择固定标签，否则选择最后一个
        const pinnedTabs = newTabs.filter((t) => t.isPinned);
        newActiveTabId =
          pinnedTabs.length > 0
            ? pinnedTabs[pinnedTabs.length - 1].id
            : newTabs[newTabs.length - 1].id;
      }

      return {
        tabs: newTabs,
        activeTabId: newTabs.length === 0 ? null : newActiveTabId,
      };
    });
  },

  // 关闭其他标签
  closeOtherTabs: (tabId: string) => {
    set((state) => {
      const closedTabs = state.tabs.filter((t) => t.id !== tabId && !t.isPinned);

      // 销毁被关闭的终端实例
      closedTabs.forEach((tab) => {
        if (tab.type === TabType.TERMINAL) {
          terminalManager.destroy(tab.propertyUuid);
        }
      });

      return {
        tabs: state.tabs.filter((t) => t.id === tabId || t.isPinned),
        activeTabId: tabId,
      };
    });
  },

  // 关闭所有标签
  closeAllTabs: () => {
    set((state) => {
      // 销毁所有终端实例
      state.tabs.forEach((tab) => {
        if (tab.type === TabType.TERMINAL) {
          terminalManager.destroy(tab.propertyUuid);
        }
      });

      return {
        tabs: [],
        activeTabId: null,
      };
    });
  },

  // 关闭右侧标签
  closeTabsToRight: (tabId: string) => {
    set((state) => {
      const index = state.tabs.findIndex((t) => t.id === tabId);
      if (index === -1) return state;

      const activatedIndex = state.tabs.findIndex(
        (t) => t.id === state.activeTabId,
      );

      const closedTabs = state.tabs
        .slice(index + 1)
        .filter((t) => !t.isPinned);

      // 销毁被关闭的终端实例
      closedTabs.forEach((tab) => {
        if (tab.type === TabType.TERMINAL) {
          terminalManager.destroy(tab.propertyUuid);
        }
      });

      return {
        tabs: [
          ...state.tabs.slice(0, index + 1),
          ...state.tabs.slice(index + 1).filter((t) => t.isPinned),
        ],
        activeTabId: activatedIndex > index ? tabId : state.activeTabId,
      };
    });
  },

  // 设置激活标签
  setActiveTab: (tabId: string) => {
    set({ activeTabId: tabId });
  },

  // 固定标签
  pinTab: (tabId: string) => {
    set((state) => ({
      tabs: state.tabs.map((t) =>
        t.id === tabId ? { ...t, isPinned: true } : t,
      ),
    }));
  },

  // 取消固定
  unpinTab: (tabId: string) => {
    set((state) => ({
      tabs: state.tabs.map((t) =>
        t.id === tabId ? { ...t, isPinned: false } : t,
      ),
    }));
  },

  // 重命名标签
  renameTab: (tabId: string, newLabel: string) => {
    set((state) => ({
      tabs: state.tabs.map((t) =>
        t.id === tabId ? { ...t, label: newLabel } : t,
      ),
    }));
  },

  // 移动标签（拖拽排序）
  moveTab: (fromIndex: number, toIndex: number) => {
    set((state) => {
      const newTabs = [...state.tabs];
      const [movedTab] = newTabs.splice(fromIndex, 1);
      newTabs.splice(toIndex, 0, movedTab);
      return { tabs: newTabs };
    });
  },

  // 根据 UUID 查找标签
  getTabByUUID: (uuid: string) => {
    return get().tabs.find((t) => t.propertyUuid === uuid);
  },

  // 获取当前激活标签
  getActiveTab: () => {
    const state = get();
    return state.tabs.find((t) => t.id === state.activeTabId);
  },
}));

export const tabStoreMethods = StoreMethods(useTabsStore);
