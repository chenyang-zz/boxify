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

import { DatabaseService } from "@wails/service";
import { callWails } from "./utils";
import {
  ColumnDefinition,
  ConnectionConfig,
  QueryResult,
} from "@wails/connection";
import { getPropertyItemByUUID } from "./property";
import { getConnectionConfigByUUID, getDatabaseNameByUUID } from "./connection";
import { DBFileType } from "@/common/constrains";
import { selectAllFrom } from "./sql";

// 根据表的 UUID 获取列信息
export async function getDBTableColumnsByUUID(
  uuid: string,
): Promise<ColumnDefinition[]> {
  const item = getPropertyItemByUUID(uuid);
  const config = getConnectionConfigByUUID(uuid);
  if (!item || !config) {
    return [];
  }

  if (item.type !== DBFileType.TABLE) {
    console.warn(`尝试获取非表类型的列信息，uuid: ${uuid}, type: ${item.type}`);
    return [];
  }

  const dbName = getDatabaseNameByUUID(uuid);
  if (!dbName) {
    console.warn(`无法获取表所属的数据库名称，uuid: ${uuid}`);
    return [];
  }

  try {
    const res = await callWails(
      DatabaseService.DBGetColumns,
      ConnectionConfig.createFrom(config),
      dbName,
      item.label,
    );

    return res.data as ColumnDefinition[];
  } catch {
    return [];
  }
}

// 根据表的 UUID 获取表数据
export async function getDBTableValuesByUUID(
  uuid: string,
): Promise<QueryResult | null> {
  const item = getPropertyItemByUUID(uuid);
  const config = getConnectionConfigByUUID(uuid);
  if (!item || !config) {
    return null;
  }

  if (item.type !== DBFileType.TABLE) {
    console.warn(`尝试获取非表类型的列信息，uuid: ${uuid}, type: ${item.type}`);
    return null;
  }

  const dbName = getDatabaseNameByUUID(uuid);
  if (!dbName) {
    console.warn(`无法获取表所属的数据库名称，uuid: ${uuid}`);
    return null;
  }

  const compile = selectAllFrom(item.label);

  try {
    const res = await callWails(
      DatabaseService.DBQuery,
      ConnectionConfig.createFrom(config),
      dbName,
      compile.sql,
      compile.parameters,
    );

    return res;
  } catch {
    return null;
  }
}
