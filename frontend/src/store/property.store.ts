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

import { ConnectionType, FileType, isDBConnection } from "@/common/constrains";
import { ConnectionConfig } from "@/types";
import { create } from "zustand";
import { persist } from "zustand/middleware";

export const FileTreeMap = new Map<string, PropertyItemType>();

export interface PropertyItemType {
  uuid: string;
  level: number;
  loading?: boolean;
  isDir: boolean;
  label: string;
  type: FileType;
  opened?: boolean;
  children?: PropertyItemType[];

  // dir 属性
  loaded?: boolean; // 目录是否已加载过子项
  isConnection?: boolean; // 是否是连接项

  // connection 属性
  connectionConfig?: ConnectionConfig; // 连接配置，具体结构根据连接类型而定
}

interface PropertyState {
  propertyList: PropertyItemType[];
  setPropertyList: (list: PropertyItemType[]) => void;
}

// 根据UUID获取属性项的详细信息
export function getPropertyItemByUUID(
  uuid: string,
): PropertyItemType | undefined {
  return FileTreeMap.get(uuid);
}

// 递归遍历文件树数据，将所有文件项存储到FileTreeMap中，方便后续通过UUID快速访问
function initTraverseTree(data: PropertyItemType[]) {
  for (const item of data) {
    FileTreeMap.set(item.uuid, item);
    if (
      item.isDir &&
      item.opened &&
      item.children &&
      item.children.length > 0
    ) {
      // 对数据库连接做特殊处理
      // 如果是数据库连接项，直接将其子项标记为未加载状态，不进行递归遍历
      if (isDBConnection(item.type)) {
        item.children = [];
        item.opened = false;
        item.loaded = false;
        continue;
      }
      initTraverseTree(item.children);
    }
  }
}

export const usePropertyStore = create<PropertyState>()(
  persist(
    (set) => ({
      propertyList: [],
      setPropertyList: (list: PropertyItemType[]) => {
        set(() => ({ propertyList: list }));
      },
    }),
    {
      name: "boxify-property-store",
      version: 1,
      onRehydrateStorage: () => (state) => {
        if (state?.propertyList) {
          initTraverseTree(state.propertyList);
          state.setPropertyList(state.propertyList);
        }
      },
      migrate: (persistedState, version) => {
        return persistedState as PropertyState;
      },
      partialize: (state) => ({
        propertyList: state.propertyList,
      }),
    },
  ),
);
