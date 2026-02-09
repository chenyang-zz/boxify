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

import {
  ConnectionType,
  DBFileType,
  FileType,
  isDBConnection,
} from "@/common/constrains";
import { ConnectionConfig } from "@/types";
import { create } from "zustand";
import { persist } from "zustand/middleware";
import { v4 as uuid } from "uuid";
import { connection } from "@wails/models";
import { callWails } from "@/lib/utils";
import { DBGetDatabases } from "@wails/app/App";
import { StoreMethods } from "./common";

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
  triggerDirOpen: (uuid: string) => Promise<void>; // 打开/关闭文件夹
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
    if (item.isDir) {
      // 对数据库连接做特殊处理
      // 如果是数据库连接项，直接将其子项标记为未加载状态，不进行递归遍历
      if (isDBConnection(item.type)) {
        item.children = [];
        item.opened = false;
        item.loaded = false;
        continue;
      }

      if (item.opened && item.children && item.children.length > 0) {
        initTraverseTree(item.children);
      }
    }
  }
}

// 根据数据库查询结果创建PropertyItemType列表
function createPropertyItemListFromDBQueryResult(
  pLevel: number,
  pType: FileType,
  res: Record<string, any>[],
): PropertyItemType[] {
  const list = [] as PropertyItemType[];
  for (const row of res) {
    let item: PropertyItemType;
    switch (pType) {
      case ConnectionType.MYSQL:
        item = {
          uuid: uuid(),
          level: pLevel + 1,
          isDir: true,
          label: row["Database"],
          type: DBFileType.DATABASE,
          opened: false,
          loaded: false,
        };
        break;
      default:
        continue;
    }
    list.push(item);
  }
  return list;
}

// 根据数据库连接项加载其子项数据
export async function loadDBConnectionPropertyChildren(
  pLevel: number,
  type: FileType,
  config: ConnectionConfig,
) {
  let res: connection.QueryResult;
  try {
    switch (type) {
      case ConnectionType.MYSQL:
        res = await callWails(
          DBGetDatabases,
          connection.ConnectionConfig.createFrom(config),
        );
        break;
      default:
        throw new Error(`不支持 connection type: ${type}`);
    }

    return createPropertyItemListFromDBQueryResult(pLevel, type, res.data);
  } catch {}
}

export const usePropertyStore = create<PropertyState>()(
  persist(
    (set, get) => ({
      propertyList: [],
      setPropertyList: (list: PropertyItemType[]) => {
        set(() => ({ propertyList: list }));
      },
      triggerDirOpen: async (uuid: string) => {
        const item = FileTreeMap.get(uuid);
        if (!item) {
          console.warn("找不到该文件夹");
          return;
        }
        if (!item.isDir) {
          return;
        }
        const dir = item;

        // 如果加载过了，就直接切换打开状态
        dir.opened = !dir.opened;

        // 如果没有加载过，应该去后端请求获取子项数据，然后更新树数据
        if (!dir.loaded) {
          // 数据库连接
          if (isDBConnection(dir.type)) {
            const children = await loadDBConnectionPropertyChildren(
              dir.level,
              dir.type,
              dir.connectionConfig!,
            );
            dir.children = children;
          } else {
            // TODO: 其他连接
          }
          dir.loaded = true;
        }

        set(() => ({ propertyList: [...get().propertyList] }));
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

export const propertyStoreMethods = StoreMethods(usePropertyStore);
