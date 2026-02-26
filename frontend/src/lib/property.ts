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
  ConnectionEnum,
  DBFileType,
  FileSystemType,
  PropertyType,
  isConnectionType,
  isDBType,
  TabType,
} from "@/common/constrains";
import {
  FileTreeMap,
  PropertyItemType,
  propertyStoreMethods,
  usePropertyStore,
} from "@/store/property.store";
import { DatabaseService } from "@wails/service";
import { callWails } from "./utils";
import { v4 as uuid } from "uuid";
import { ConnectionConfig, QueryResult } from "@wails/connection";
import { getConnectionConfigByUUID } from "./connection";
import { tabStoreMethods } from "@/store/tabs.store";

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
      case ConnectionEnum.MYSQL:
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
export function createDatabaseQueryResult(): QueryResult {
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
  type: PropertyType,
  config: ConnectionConfig,
): Promise<QueryResult> {
  try {
    switch (type) {
      case DBFileType.TABLE_FOLDER:
        return await callWails(
          DatabaseService.DBGetTables,
          ConnectionConfig.createFrom(config),
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

  let res: QueryResult;
  try {
    if (!config) {
      throw new Error("缺少连接配置");
    }

    switch (type) {
      case ConnectionEnum.MYSQL:
        res = await callWails(
          DatabaseService.DBGetDatabases,
          ConnectionConfig.createFrom(config),
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
              ConnectionConfig.createFrom(config),
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
    let res: PropertyItemType[] = [];
    if (isDBType(item.type)) {
      res = await loadDBConnectionPropertyChildren(uuid);
    } else {
      // TODO: 其他连接
    }

    if (res.length > 0) {
      item.loaded = true;
    }
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
      tabStoreMethods.openTab(item);
      break;
  }
}

// 将属性类型转换为标签页类型
export function propertyTypeToTabType(type: PropertyType): TabType {
  switch (type) {
    case DBFileType.TABLE:
      return TabType.TABLE;
    case ConnectionEnum.TERMINAL:
      return TabType.TERMINAL;
    case ConnectionEnum.SSH:
      return TabType.TERMINAL;
    default:
      throw new Error(`不支持的属性类型: ${type}`);
  }
}

// 根据属性项UUID获取对应的标签页类型
export function getTabTypeFromProperty(uuid: string): TabType {
  const item = getPropertyItemByUUID(uuid);
  if (!item) {
    throw new Error("无法找到对应的属性项");
  }

  switch (item.type) {
    case DBFileType.TABLE:
      return TabType.TABLE;
    default:
      throw new Error(`不支持的属性类型: ${item.type}`);
  }
}

// 将新的属性项添加到指定父级下
export function addPropertyItemToParent(
  pUuid: string | null,
  item: PropertyItemType,
) {
  if (pUuid) {
    const parent = getPropertyItemByUUID(pUuid);
    if (!parent) {
      throw new Error("无法找到父级属性项");
    }

    if (!parent.children) {
      parent.children = [];
    }

    parent.children.push(item);
    item.parent = parent; // 设置父级引用，方便后续向上访问
  } else {
    // 如果没有指定父级UUID，说明是添加到根级
    const rootList = usePropertyStore.getState().propertyList;
    rootList.push(item);
  }

  FileTreeMap.set(item.uuid, item); // 同时记录到全局Map中

  propertyStoreMethods.setPropertyList([
    ...usePropertyStore.getState().propertyList,
  ]); // 触发状态更新，刷新UI
}

// 根据属性项UUID获取最近的文件夹
// 如果当前项是文件夹，则返回当前文件夹
// 如果当前项是文件，则返回其父级文件夹
// 如果没有父级文件夹（即已经是根级项），则返回null
export function getClosestFolder(uuid: string): PropertyItemType | null {
  let item = getPropertyItemByUUID(uuid);
  if (!item) {
    return null;
  }

  while (item && item.type !== FileSystemType.FOLDER) {
    item = item.parent ?? null;
  }

  return item;
}

// 根据UUID关闭连接项
export function closeConnectionByUUID(uuid: string) {
  const item = getPropertyItemByUUID(uuid);
  if (!item) {
    return;
  }

  if (!isConnectionType(item.type)) return;

  // 关闭连接项的操作，具体实现可以根据实际需求进行调整
  item.loaded = false;
  item.children = [];
  item.opened = false;

  propertyStoreMethods.setPropertyList([
    ...usePropertyStore.getState().propertyList,
  ]); // 触发状态更新，刷新UI
}

// 根据UUID删除连接项
export function deleteConnectionByUUID(uuid: string) {
  const item = getPropertyItemByUUID(uuid);
  if (!item) {
    return;
  }

  if (!isConnectionType(item.type)) return;

  const parent = item.parent;
  let rootList = usePropertyStore.getState().propertyList;

  if (parent) {
    // 从父级的children中移除该项
    parent.children = parent.children?.filter((child) => child.uuid !== uuid);
  } else {
    // 如果没有父级，说明是根级项，直接从根列表中移除
    rootList = rootList.filter((rootItem) => rootItem.uuid !== uuid);
  }

  propertyStoreMethods.setPropertyList([...rootList]); // 触发状态更新，刷新UI
}
