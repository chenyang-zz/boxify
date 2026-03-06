import { ChangeSet } from "@wails/connection";
import type {
  DBTableBuiltChangeSet,
  DBTableDraftRow,
  DBTableRenderRow,
  DBTableSortState,
} from "../types";

// 将后端返回数据初始化为可编辑草稿行。
export function createDraftRows(data: Record<string, any>[]): DBTableDraftRow[] {
  return data.map((row, index) => ({
    id: `existing-${index}`,
    mode: "existing",
    values: { ...row },
    originalValues: { ...row },
    deleted: false,
  }));
}

// 复制草稿行，避免历史记录共享引用。
export function cloneDraftRows(rows: DBTableDraftRow[]): DBTableDraftRow[] {
  return rows.map((row) => ({
    ...row,
    values: { ...row.values },
    originalValues: row.originalValues ? { ...row.originalValues } : undefined,
  }));
}

// 创建一个空插入行。
export function createInsertedRow(
  columns: string[],
  nextId: number,
): DBTableDraftRow {
  const values = columns.reduce<Record<string, any>>((acc, column) => {
    acc[column] = "";
    return acc;
  }, {});

  return {
    id: `inserted-${nextId}`,
    mode: "inserted",
    values,
    deleted: false,
  };
}

// 更新单元格值。
export function updateCellValue(
  rows: DBTableDraftRow[],
  rowId: string,
  column: string,
  value: string,
): DBTableDraftRow[] {
  return rows.map((row) => {
    if (row.id !== rowId) {
      return row;
    }

    return {
      ...row,
      values: {
        ...row.values,
        [column]: value,
      },
    };
  });
}

// 标记或取消删除指定行。
export function toggleRowsDeleted(
  rows: DBTableDraftRow[],
  selectedRowIds: Set<string>,
): DBTableDraftRow[] {
  return rows
    .map((row) => {
      if (!selectedRowIds.has(row.id)) {
        return row;
      }

      if (row.mode === "inserted") {
        return null;
      }

      return {
        ...row,
        deleted: !row.deleted,
      };
    })
    .filter((row): row is DBTableDraftRow => row !== null);
}

function asComparableValue(value: any): string {
  if (value === null || value === undefined) {
    return "";
  }
  if (typeof value === "object") {
    return JSON.stringify(value);
  }
  return String(value);
}

function isValueChanged(before: any, after: any): boolean {
  return asComparableValue(before) !== asComparableValue(after);
}

function getPrimaryKeys(
  row: DBTableDraftRow,
  columns: string[],
  primaryColumns: string[],
): Record<string, any> {
  const keyColumns = primaryColumns.length > 0 ? primaryColumns : columns;
  const source = row.originalValues ?? row.values;

  return keyColumns.reduce<Record<string, any>>((acc, key) => {
    acc[key] = source[key];
    return acc;
  }, {});
}

function getKeyColumns(columns: string[], primaryColumns: string[]): string[] {
  return primaryColumns.length > 0 ? primaryColumns : columns;
}

function buildPrimaryKeysFromValues(
  values: Record<string, any>,
  columns: string[],
  primaryColumns: string[],
): Record<string, any> {
  return getKeyColumns(columns, primaryColumns).reduce<Record<string, any>>((acc, key) => {
    acc[key] = values[key];
    return acc;
  }, {});
}

function buildRowSignature(
  values: Record<string, any>,
  columns: string[],
  primaryColumns: string[],
): string {
  const keyColumns = getKeyColumns(columns, primaryColumns);
  return keyColumns
    .map((key) => `${key}:${asComparableValue(values[key])}`)
    .join("\u001f");
}

// 将草稿行转换为后端可消费的 ChangeSet。
export function buildChangeSet(
  rows: DBTableDraftRow[],
  columns: string[],
  primaryColumns: string[],
): DBTableBuiltChangeSet {
  const changes = ChangeSet.createFrom({
    inserts: [],
    updates: [],
    deletes: [],
  });

  for (const row of rows) {
    if (row.mode === "inserted") {
      if (!row.deleted) {
        changes.inserts.push({ ...row.values });
      }
      continue;
    }

    if (row.deleted) {
      changes.deletes.push(getPrimaryKeys(row, columns, primaryColumns));
      continue;
    }

    const originalValues = row.originalValues ?? {};
    const changedValues: Record<string, any> = {};
    for (const column of columns) {
      if (isValueChanged(originalValues[column], row.values[column])) {
        changedValues[column] = row.values[column];
      }
    }

    if (Object.keys(changedValues).length === 0) {
      continue;
    }

    changes.updates.push({
      keys: getPrimaryKeys(row, columns, primaryColumns),
      values: changedValues,
    });
  }

  return {
    changes,
    summary: {
      inserts: changes.inserts.length,
      updates: changes.updates.length,
      deletes: changes.deletes.length,
    },
  };
}

// 比较两组行并生成从 fromRows 同步到 toRows 的变更集。
export function buildChangeSetBetween(
  fromRows: DBTableDraftRow[],
  toRows: DBTableDraftRow[],
  columns: string[],
  primaryColumns: string[],
): DBTableBuiltChangeSet {
  const changes = ChangeSet.createFrom({
    inserts: [],
    updates: [],
    deletes: [],
  });

  const fromMap = new Map<string, Record<string, any>>();
  for (const row of fromRows) {
    fromMap.set(buildRowSignature(row.values, columns, primaryColumns), { ...row.values });
  }

  const toMap = new Map<string, Record<string, any>>();
  for (const row of toRows) {
    toMap.set(buildRowSignature(row.values, columns, primaryColumns), { ...row.values });
  }

  for (const [signature, fromValues] of fromMap.entries()) {
    const toValues = toMap.get(signature);
    if (!toValues) {
      changes.deletes.push(buildPrimaryKeysFromValues(fromValues, columns, primaryColumns));
      continue;
    }

    const changedValues: Record<string, any> = {};
    for (const column of columns) {
      if (isValueChanged(fromValues[column], toValues[column])) {
        changedValues[column] = toValues[column];
      }
    }
    if (Object.keys(changedValues).length > 0) {
      changes.updates.push({
        keys: buildPrimaryKeysFromValues(fromValues, columns, primaryColumns),
        values: changedValues,
      });
    }
  }

  for (const [signature, toValues] of toMap.entries()) {
    if (!fromMap.has(signature)) {
      changes.inserts.push({ ...toValues });
    }
  }

  return {
    changes,
    summary: {
      inserts: changes.inserts.length,
      updates: changes.updates.length,
      deletes: changes.deletes.length,
    },
  };
}

// 计算当前是否存在待提交变更。
export function hasPendingChanges(
  rows: DBTableDraftRow[],
  columns: string[],
  primaryColumns: string[],
): boolean {
  const built = buildChangeSet(rows, columns, primaryColumns);
  return (
    built.summary.inserts > 0 || built.summary.updates > 0 || built.summary.deletes > 0
  );
}

// 按关键字与排序状态得到渲染行。
export function toRenderRows(
  rows: DBTableDraftRow[],
  columns: string[],
  filterKeyword: string,
  sortState: DBTableSortState,
): DBTableRenderRow[] {
  const keyword = filterKeyword.trim().toLowerCase();
  const filtered = rows.filter((row) => {
    if (!keyword) {
      return true;
    }

    return columns.some((column) =>
      asComparableValue(row.values[column]).toLowerCase().includes(keyword),
    );
  });

  const sorted = [...filtered];
  if (sortState.column && sortState.direction !== "none") {
    sorted.sort((a, b) => {
      const va = asComparableValue(a.values[sortState.column!]);
      const vb = asComparableValue(b.values[sortState.column!]);
      const result = va.localeCompare(vb, "zh-CN", { numeric: true });
      return sortState.direction === "asc" ? result : -result;
    });
  }

  return sorted.map((row, index) => ({ row, index }));
}
