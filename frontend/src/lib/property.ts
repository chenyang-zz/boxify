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
  isDBType,
} from "@/common/constrains";
import {
  FileTreeMap,
  PropertyItemType,
  propertyStoreMethods,
  usePropertyStore,
} from "@/store/property.store";
import { ConnectionConfig } from "@/types";
import { DBGetDatabases, DBGetTables } from "@wails/app/App";
import { callWails } from "./utils";
import { v4 as uuid } from "uuid";
import { connection } from "@wails/models";
import { getConnectionConfigByUUID } from "./connection";
import { use } from "react";
import { getDBTableColumnsByUUID } from "./dbTable";

// 根据UUID获取属性项的详细信息
export function getPropertyItemByUUID(uuid: string): PropertyItemType | null {
  return FileTreeMap.get(uuid) ?? null;
}

// 递归遍历文件树数据，将所有文件项存储到FileTreeMap中，方便后续通过UUID快速访问
export function initTraverseTree(
  data: PropertyItemType[],
  parent?: PropertyItemType,
) {
  for (const item of data) {
    item.loaded = false;
    FileTreeMap.set(item.uuid, item);

    // 设置父级引用，方便后续需要向上访问时使用
    if (parent) {
      item.parent = parent;
    }

    if (item.isDir) {
      // 对数据库连接做特殊处理
      // 如果是数据库连接项，直接将其子项标记为未加载状态，不进行递归遍历
      if (isDBType(item.type)) {
        item.children = [];
        item.opened = false;
        continue;
      }

      if (item.opened && item.children && item.children.length > 0) {
        item.loaded = true;
        initTraverseTree(item.children, item);
      }
    }
  }
}

// 根据数据库查询结果创建PropertyItemType列表
export function createPropertyItemListFromDBQueryResult(
  pItem: PropertyItemType,
  res: Record<string, any>[],
): PropertyItemType[] {
  const { level: pLevel, type: pType } = pItem;
  const list = [] as PropertyItemType[];
  for (const row of res) {
    let partialItem: Partial<PropertyItemType>;
    switch (pType) {
      case ConnectionType.MYSQL:
        partialItem = {
          isDir: true,
          label: row["Database"],
          type: DBFileType.DATABASE,
        };
        break;
      case DBFileType.DATABASE:
        partialItem = {
          isDir: true,
          label: row["Placeholder"],
          type: row["type"],
          children: row["children"] || [],
        };
        break;
      case DBFileType.TABLE_FOLDER:
        partialItem = {
          isDir: false,
          label: row["Table"],
          type: DBFileType.TABLE,
        };
        break;
      default:
        continue;
    }

    partialItem.uuid = uuid();
    partialItem.level = pLevel + 1;
    partialItem.opened = false;
    partialItem.loaded = false;

    const item = partialItem as PropertyItemType;
    item.parent = pItem; // 设置父级引用，方便后续向上访问

    // 如果查询结果中包含子项数据（如表列表），就直接设置到children属性中，并标记为已加载
    if (row["children"] && row["children"].length > 0) {
      if (!item.extra) item.extra = {};
      item.extra["count"] = row["children"].length; // 记录子项数量，方便前端展示
      item.children = createPropertyItemListFromDBQueryResult(
        item,
        row["children"],
      );
    }

    // 记录到全局Map中，方便后续通过UUID快速访问
    FileTreeMap.set(partialItem.uuid, item);
    list.push(item);
  }

  pItem.children = list; // 更新父级的children属性，确保树结构正确
  return list;
}

// 创建数据库属性项列表的占位符
export function createDatabaseQueryResult(): connection.QueryResult {
  return {
    success: true,
    data: [
      {
        Placeholder: "表",
        type: DBFileType.TABLE_FOLDER,
      },
      {
        Placeholder: "视图",
        type: DBFileType.VIEW_FOLDER,
      },
      {
        Placeholder: "查询",
        type: DBFileType.QUERY_FOLDER,
      },
      {
        Placeholder: "存储过程/函数",
        type: DBFileType.FUNCTION_FOLDER,
      },
    ],
    message: "占位符",
    fields: [],
  };
}

// 加载数据库下的子项：表、视图等
async function loadDBChildrenByDBName(
  dbName: string,
  type: FileType,
  config: ConnectionConfig,
): Promise<connection.QueryResult> {
  try {
    switch (type) {
      case DBFileType.TABLE_FOLDER:
        return await callWails(
          DBGetTables,
          connection.ConnectionConfig.createFrom(config),
          dbName,
        );

      default:
        throw new Error(`不支持的类型: ${type}`);
    }
  } catch (e) {
    throw e;
  }
}

// 根据数据库连接项加载其子项数据
export async function loadDBConnectionPropertyChildren(
  uuid: string,
): Promise<PropertyItemType[]> {
  const item = getPropertyItemByUUID(uuid);
  if (!item) {
    return [];
  }
  const { type, label } = item;

  const config = getConnectionConfigByUUID(item.uuid);

  let res: connection.QueryResult;
  try {
    if (!config) {
      throw new Error("缺少连接配置");
    }

    switch (type) {
      case ConnectionType.MYSQL:
        res = await callWails(
          DBGetDatabases,
          connection.ConnectionConfig.createFrom(config),
        );
        break;
      case DBFileType.DATABASE:
        res = createDatabaseQueryResult();
        // 需要立刻加载子项
        for (const row of res.data) {
          try {
            const rowChildren = await loadDBChildrenByDBName(
              label,
              row["type"],
              config,
            );
            row["children"] = rowChildren.data;
          } catch {
            continue;
          }
        }
        break;
      default:
        throw new Error(`不支持 connection type: ${type}`);
    }

    return createPropertyItemListFromDBQueryResult(item, res.data);
  } catch {
    return [];
  }
}

// 触发打开文件夹的操作
export async function triggerDirOpen(uuid: string) {
  const item = getPropertyItemByUUID(uuid);
  if (!item) {
    console.warn("找不到该文件夹");
    return;
  }

  // 如果加载过了，就直接切换打开状态
  item.opened = !item.opened;

  // 如果没有加载过，应该去后端请求获取子项数据，然后更新树数据
  if (!item.loaded) {
    // 数据库连接
    if (isDBType(item.type)) {
      await loadDBConnectionPropertyChildren(uuid);
    } else {
      // TODO: 其他连接
    }
    item.loaded = true;
  }

  propertyStoreMethods.setPropertyList([
    ...usePropertyStore.getState().propertyList,
  ]);
}

// 触发打开文件的操作
export async function triggerFileOpen(uuid: string) {
  const item = getPropertyItemByUUID(uuid);
  if (!item) {
    console.warn("找不到该文件");
    return;
  }

  switch (item.type) {
    case DBFileType.TABLE:
      getDBTableColumnsByUUID(item.uuid);
      break;
  }
}
