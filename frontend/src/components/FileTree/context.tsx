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
  FileTreeMap,
  PropertyItemType,
  usePropertyStore,
} from "@/store/property.store";
import {
  createContext,
  FC,
  ReactNode,
  use,
  useContext,
  useEffect,
  useState,
} from "react";
import { DBGetDatabases } from "@wails/app/App";
import { DBFileType, isDBConnection } from "@/common/constrains";
import { connection } from "@wails/models";
import { v4 as uuid } from "uuid";
import { callWails } from "@/lib/utils";
import { useShallow } from "zustand/react/shallow";

export const fileTreeContext = {
  selectedUUID: "", // 当前选中的文件或文件夹的UUID
  setSelectedUUID: (uuid: string) => {}, // 更新选中的UUID
  triggerDirOpen: (uuid: string) => {}, // 打开/关闭文件夹
};

export type FileTreeContextType = typeof fileTreeContext;

const FileTreeContext = createContext<FileTreeContextType>(fileTreeContext);

// 根据数据库查询结果创建PropertyItemType列表
function createPropertyItemListFromDBQueryResult(
  pLevel: number,
  type: DBFileType,
  res: Record<string, any>[],
): PropertyItemType[] {
  const list = [] as PropertyItemType[];
  for (const row of res) {
    let item: PropertyItemType;
    switch (type) {
      case DBFileType.DATABASE:
        item = {
          uuid: uuid(),
          level: pLevel + 1,
          isDir: true,
          label: row["Database"],
          type,
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

const FileTreeProvider: FC<{
  children: ReactNode;
}> = ({ children }) => {
  const [selectedUUID, setSelectedUUID] = useState("");

  const propertyList = usePropertyStore((state) => state.propertyList);

  // 打开/关闭文件夹
  const triggerDirOpen = async (uuid: string) => {
    const item = FileTreeMap.get(uuid);
    if (!item) {
      throw new Error(`Directory with UUID ${uuid} not found`);
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
        try {
          const res = await callWails(
            DBGetDatabases,
            connection.ConnectionConfig.createFrom(dir.connectionConfig!),
          );
          const children = createPropertyItemListFromDBQueryResult(
            dir.level,
            DBFileType.DATABASE,
            res.data,
          );
          dir.children = children;
        } catch {}
      } else {
        // TODO: 其他连接
      }
      dir.loaded = true;
    }

    usePropertyStore.getState().setPropertyList([...propertyList]);
  };

  return (
    <FileTreeContext.Provider
      value={{
        selectedUUID,
        setSelectedUUID,
        triggerDirOpen,
      }}
    >
      {children}
    </FileTreeContext.Provider>
  );
};

export const useFileTreeContext = () => {
  const context = useContext(FileTreeContext);
  if (!context) {
    throw new Error(
      "useFileTreeContext must be used within a FileTreeProvider",
    );
  }
  return context;
};

export default FileTreeProvider;
