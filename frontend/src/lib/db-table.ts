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

import { DBFileType } from "@/common/constrains";
import { DatabaseService } from "@wails/service";
import {
  ChangeSet,
  ColumnDefinition,
  ConnectionConfig,
  QueryResult,
} from "@wails/connection";
import { getPropertyItemByUUID } from "./property";
import { getConnectionConfigByUUID, getDatabaseNameByUUID } from "./connection";
import { callWails, callWailsWithOptions } from "./utils";

export type DBTableExportFormat = "csv" | "json" | "md";

interface DBTableContext {
  config: ConnectionConfig;
  dbName: string;
  tableName: string;
}

// 根据数据库类型包装标识符，避免表名包含特殊字符导致查询失败。
function quoteIdentifierByType(
  dbType: string | undefined,
  identifier: string,
): string {
  if (dbType === "postgresql") {
    return `"${identifier.replace(/"/g, `""`)}"`;
  }
  return `\`${identifier.replace(/`/g, "``")}\``;
}

// 生成表数据查询 SQL，支持可选筛选表达式。
function buildTableQuerySQL(
  dbType: string | undefined,
  tableName: string,
  filterExpression = "",
  offset = 0,
  limit = 500,
): string {
  const quotedTable = quoteIdentifierByType(dbType, tableName);
  const whereClause = filterExpression.trim()
    ? ` WHERE ${filterExpression.trim()}`
    : "";

  if (dbType === "postgresql") {
    return `SELECT * FROM ${quotedTable}${whereClause} LIMIT ${limit} OFFSET ${offset}`;
  }
  return `SELECT * FROM ${quotedTable}${whereClause} LIMIT ${offset},${limit}`;
}

// 解析并校验表操作所需上下文。
function resolveDBTableContext(uuid: string): DBTableContext | null {
  const item = getPropertyItemByUUID(uuid);
  const config = getConnectionConfigByUUID(uuid);
  if (!item || !config) {
    return null;
  }

  if (item.type !== DBFileType.TABLE) {
    console.warn(`尝试操作非表类型节点，uuid: ${uuid}, type: ${item.type}`);
    return null;
  }

  const dbName = getDatabaseNameByUUID(uuid);
  if (!dbName) {
    console.warn(`无法获取表所属数据库名称，uuid: ${uuid}`);
    return null;
  }

  return {
    config: ConnectionConfig.createFrom(config),
    dbName,
    tableName: item.label,
  };
}

// 根据表的 UUID 获取列信息。
export async function getDBTableColumnsByUUID(
  uuid: string,
): Promise<ColumnDefinition[]> {
  const ctx = resolveDBTableContext(uuid);
  if (!ctx) {
    return [];
  }

  try {
    const res = await callWailsWithOptions(
      DatabaseService.DBGetColumns,
      [ctx.config, ctx.dbName, ctx.tableName],
      {
        timeoutMs: 30000,
        timeoutMessage: "加载字段超时，请检查连接状态后重试",
      },
    );

    return res.data as ColumnDefinition[];
  } catch {
    return [];
  }
}

// 根据表的 UUID 获取表数据。
export async function getDBTableValuesByUUID(
  uuid: string,
  filterExpression = "",
): Promise<QueryResult | null> {
  const ctx = resolveDBTableContext(uuid);
  if (!ctx) {
    return null;
  }

  const querySQL = buildTableQuerySQL(
    ctx.config.type,
    ctx.tableName,
    filterExpression,
  );

  try {
    return await callWailsWithOptions(
      DatabaseService.DBQuery,
      [ctx.config, ctx.dbName, querySQL, []],
      {
        timeoutMs: 60000,
        timeoutMessage: "加载表数据超时，请缩小数据范围或调整连接超时",
      },
    );
  } catch {
    return null;
  }
}

// 将前端编辑变更集应用到目标表。
export async function applyDBTableChangesByUUID(
  uuid: string,
  changes: ChangeSet,
): Promise<QueryResult | null> {
  const ctx = resolveDBTableContext(uuid);
  if (!ctx) {
    return null;
  }

  return callWailsWithOptions(
    DatabaseService.ApplyChanges,
    [ctx.config, ctx.dbName, ctx.tableName, ChangeSet.createFrom(changes)],
    {
      timeoutMs: 30000,
      timeoutMessage: "保存超时，请检查锁等待或连接状态后重试",
    },
  );
}

// 导入数据文件并写入目标表。
export async function importDBTableByUUID(uuid: string): Promise<QueryResult | null> {
  const ctx = resolveDBTableContext(uuid);
  if (!ctx) {
    return null;
  }

  try {
    return await callWails(
      DatabaseService.ImportData,
      ctx.config,
      ctx.dbName,
      ctx.tableName,
    );
  } catch {
    return null;
  }
}

// 按指定格式导出目标表。
export async function exportDBTableByUUID(
  uuid: string,
  format: DBTableExportFormat,
): Promise<QueryResult | null> {
  const ctx = resolveDBTableContext(uuid);
  if (!ctx) {
    return null;
  }

  try {
    return await callWails(
      DatabaseService.ExportTable,
      ctx.config,
      ctx.dbName,
      ctx.tableName,
      format,
    );
  } catch {
    return null;
  }
}
