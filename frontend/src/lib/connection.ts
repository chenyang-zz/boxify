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

import { DBFileType, isConnectionType } from "@/common/constrains";
import {
  PropertyItemType,
  propertyStoreMethods,
  usePropertyStore,
} from "@/store/property.store";
import { getPropertyItemByUUID } from "./property";

// 根据UUID获取所属的数据库连接项的名称
export function getConnectionConfigByUUID(uuid: string) {
  const item = getPropertyItemByUUID(uuid);
  if (!item) {
    return null;
  }

  let node = item as PropertyItemType | undefined;

  while (node) {
    if (isConnectionType(node.type)) {
      return node.connectionConfig ?? null;
    } else {
      node = node.parent;
    }
  }
  return null;
}

// 根据UUID获取所属的数据库的名称
export function getDatabaseNameByUUID(uuid: string) {
  const item = getPropertyItemByUUID(uuid);
  if (!item) {
    return null;
  }

  let node = item as PropertyItemType | undefined;

  while (node) {
    if (node.type === DBFileType.DATABASE) {
      return node.label;
    } else {
      node = node.parent;
    }
  }

  return null;
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
