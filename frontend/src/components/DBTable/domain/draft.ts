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

type FilterTokenType =
  | "identifier"
  | "string"
  | "number"
  | "operator"
  | "keyword"
  | "paren"
  | "comma";

interface FilterToken {
  type: FilterTokenType;
  value: string;
}

interface FilterColumnOperand {
  type: "column";
  column: string;
}

interface FilterFunctionOperand {
  type: "function";
  name: string;
  args: FilterOperandArg[];
}

interface FilterLiteralOperand {
  type: "literal";
  value: any;
}

type FilterOperand = FilterColumnOperand | FilterFunctionOperand;
type FilterOperandArg = FilterOperand | FilterLiteralOperand;

type FilterComparator =
  | "="
  | "!="
  | "<>"
  | ">"
  | ">="
  | "<"
  | "<="
  | "LIKE"
  | "ILIKE"
  | "NOT LIKE"
  | "NOT ILIKE"
  | "IS NULL"
  | "IS NOT NULL";

interface FilterComparisonNode {
  type: "comparison";
  leftOperand: FilterOperand;
  comparator: FilterComparator;
  value?: any;
}

interface FilterBinaryNode {
  type: "binary";
  operator: "AND" | "OR";
  left: FilterNode;
  right: FilterNode;
}

type FilterNode = FilterComparisonNode | FilterBinaryNode;

interface ParsedFilterExpression {
  valid: boolean;
  message?: string;
  ast?: FilterNode;
}

function tokenizeFilterExpression(expression: string): FilterToken[] {
  const tokens: FilterToken[] = [];
  let index = 0;

  while (index < expression.length) {
    const char = expression[index];

    if (/\s/.test(char)) {
      index += 1;
      continue;
    }

    if (char === "(" || char === ")") {
      tokens.push({ type: "paren", value: char });
      index += 1;
      continue;
    }

    if (char === ",") {
      tokens.push({ type: "comma", value: char });
      index += 1;
      continue;
    }

    if (char === "'" || char === "\"") {
      const quote = char;
      let cursor = index + 1;
      let text = "";
      while (cursor < expression.length) {
        const current = expression[cursor];
        if (current === quote) {
          break;
        }
        text += current;
        cursor += 1;
      }
      tokens.push({ type: "string", value: text });
      index = cursor < expression.length ? cursor + 1 : cursor;
      continue;
    }

    const threeCharOp = expression.slice(index, index + 3);
    const twoCharOp = expression.slice(index, index + 2);
    if (threeCharOp === "<=>") {
      tokens.push({ type: "operator", value: "=" });
      index += 3;
      continue;
    }
    if ([">=", "<=", "!=", "<>"].includes(twoCharOp)) {
      tokens.push({ type: "operator", value: twoCharOp });
      index += 2;
      continue;
    }
    if (["=", ">", "<"].includes(char)) {
      tokens.push({ type: "operator", value: char });
      index += 1;
      continue;
    }

    if (/[0-9]/.test(char)) {
      let cursor = index + 1;
      while (cursor < expression.length && /[0-9.]/.test(expression[cursor])) {
        cursor += 1;
      }
      tokens.push({ type: "number", value: expression.slice(index, cursor) });
      index = cursor;
      continue;
    }

    let cursor = index + 1;
    while (cursor < expression.length && /[A-Za-z0-9_]/.test(expression[cursor])) {
      cursor += 1;
    }
    const rawWord = expression.slice(index, cursor);
    const upper = rawWord.toUpperCase();
    if (["AND", "OR", "LIKE", "ILIKE", "NOT", "IS", "NULL", "TRUE", "FALSE"].includes(upper)) {
      tokens.push({ type: "keyword", value: upper });
    } else {
      tokens.push({ type: "identifier", value: rawWord });
    }
    index = cursor;
  }

  return tokens;
}

function parseFilterExpression(expression: string): ParsedFilterExpression {
  const tokens = tokenizeFilterExpression(expression);
  let cursor = 0;

  const peek = () => tokens[cursor];
  const consume = () => {
    const token = tokens[cursor];
    cursor += 1;
    return token;
  };

  const parseValue = (): any => {
    const token = peek();
    if (!token) {
      return undefined;
    }

    if (token.type === "string") {
      consume();
      return token.value;
    }
    if (token.type === "number") {
      consume();
      const parsed = Number(token.value);
      return Number.isNaN(parsed) ? token.value : parsed;
    }
    if (token.type === "keyword" && token.value === "NULL") {
      consume();
      return null;
    }
    if (token.type === "keyword" && token.value === "TRUE") {
      consume();
      return true;
    }
    if (token.type === "keyword" && token.value === "FALSE") {
      consume();
      return false;
    }
    if (token.type === "identifier" || token.type === "keyword") {
      consume();
      return token.value;
    }
    return undefined;
  };

  const parseOperandFromIdentifier = (identifier: string): FilterOperand | null => {
    const maybeOpen = peek();
    if (!(maybeOpen?.type === "paren" && maybeOpen.value === "(")) {
      return {
        type: "column",
        column: identifier,
      };
    }

    consume();
    const args: FilterOperandArg[] = [];
    while (true) {
      const tail = peek();
      if (!tail) {
        return null;
      }
      if (tail.type === "paren" && tail.value === ")") {
        consume();
        break;
      }

      const argToken = peek();
      if (!argToken) {
        return null;
      }

      if (argToken.type === "identifier") {
        consume();
        const operand = parseOperandFromIdentifier(argToken.value);
        if (!operand) {
          return null;
        }
        args.push(operand);
      } else {
        const literal = parseValue();
        if (literal === undefined) {
          return null;
        }
        args.push({ type: "literal", value: literal });
      }

      const separator = peek();
      if (separator?.type === "comma") {
        consume();
        continue;
      }
      if (separator?.type === "paren" && separator.value === ")") {
        consume();
        break;
      }
      return null;
    }

    return {
      type: "function",
      name: identifier.toUpperCase(),
      args,
    };
  };

  const parseLeftOperand = (): FilterOperand | null => {
    const token = consume();
    if (!token || token.type !== "identifier") {
      return null;
    }
    return parseOperandFromIdentifier(token.value);
  };

  const parseComparison = (): FilterNode | null => {
    const leftOperand = parseLeftOperand();
    if (!leftOperand) {
      return null;
    }

    const first = peek();
    if (!first) {
      return null;
    }

    if (first.type === "keyword" && first.value === "IS") {
      consume();
      const maybeNot = peek();
      if (maybeNot?.type === "keyword" && maybeNot.value === "NOT") {
        consume();
        const nil = peek();
        if (nil?.type === "keyword" && nil.value === "NULL") {
          consume();
          return {
            type: "comparison",
            leftOperand,
            comparator: "IS NOT NULL",
          };
        }
        return null;
      }
      const nil = peek();
      if (nil?.type === "keyword" && nil.value === "NULL") {
        consume();
        return {
          type: "comparison",
          leftOperand,
          comparator: "IS NULL",
        };
      }
      return null;
    }

    if (first.type === "keyword" && first.value === "NOT") {
      consume();
      const keyword = peek();
      if (!(keyword?.type === "keyword" && (keyword.value === "LIKE" || keyword.value === "ILIKE"))) {
        return null;
      }
      consume();
      const value = parseValue();
      if (value === undefined) {
        return null;
      }
      return {
        type: "comparison",
        leftOperand,
        comparator: keyword.value === "ILIKE" ? "NOT ILIKE" : "NOT LIKE",
        value,
      };
    }

    if (first.type === "keyword" && (first.value === "LIKE" || first.value === "ILIKE")) {
      consume();
      const value = parseValue();
      if (value === undefined) {
        return null;
      }
      return {
        type: "comparison",
        leftOperand,
        comparator: first.value as FilterComparator,
        value,
      };
    }

    if (first.type === "operator") {
      consume();
      const value = parseValue();
      if (value === undefined) {
        return null;
      }
      return {
        type: "comparison",
        leftOperand,
        comparator: first.value as FilterComparator,
        value,
      };
    }

    return null;
  };

  const parsePrimary = (): FilterNode | null => {
    const token = peek();
    if (!token) {
      return null;
    }
    if (token.type === "paren" && token.value === "(") {
      consume();
      const expr = parseOrExpression();
      const tail = peek();
      if (!expr || !(tail?.type === "paren" && tail.value === ")")) {
        return null;
      }
      consume();
      return expr;
    }
    return parseComparison();
  };

  const parseAndExpression = (): FilterNode | null => {
    let node = parsePrimary();
    if (!node) {
      return null;
    }

    while (peek()?.type === "keyword" && peek()?.value === "AND") {
      consume();
      const right = parsePrimary();
      if (!right) {
        return null;
      }
      node = {
        type: "binary",
        operator: "AND",
        left: node,
        right,
      };
    }

    return node;
  };

  const parseOrExpression = (): FilterNode | null => {
    let node = parseAndExpression();
    if (!node) {
      return null;
    }

    while (peek()?.type === "keyword" && peek()?.value === "OR") {
      consume();
      const right = parseAndExpression();
      if (!right) {
        return null;
      }
      node = {
        type: "binary",
        operator: "OR",
        left: node,
        right,
      };
    }

    return node;
  };

  if (tokens.length === 0) {
    return { valid: true };
  }

  const ast = parseOrExpression();
  if (!ast || cursor < tokens.length) {
    return {
      valid: false,
      message: "筛选语法无效，请检查字段名、操作符或括号",
    };
  }

  return {
    valid: true,
    ast,
  };
}

function toComparablePair(left: any, right: any): { left: any; right: any } {
  const numericLeft = Number(left);
  const numericRight = Number(right);
  if (!Number.isNaN(numericLeft) && !Number.isNaN(numericRight)) {
    return { left: numericLeft, right: numericRight };
  }
  return {
    left: asComparableValue(left).toLowerCase(),
    right: asComparableValue(right).toLowerCase(),
  };
}

function escapeRegExp(input: string): string {
  return input.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function evaluateOperand(
  operand: FilterOperandArg,
  rowValues: Record<string, any>,
): any {
  if (operand.type === "literal") {
    return operand.value;
  }
  if (operand.type === "column") {
    return rowValues[operand.column];
  }

  const args = operand.args.map((arg) => evaluateOperand(arg, rowValues));
  switch (operand.name) {
    case "LOWER":
      return asComparableValue(args[0]).toLowerCase();
    case "UPPER":
      return asComparableValue(args[0]).toUpperCase();
    case "TRIM":
    case "BTRIM":
      return asComparableValue(args[0]).trim();
    case "LENGTH":
      return asComparableValue(args[0]).length;
    case "ABS":
      return Math.abs(Number(args[0]));
    case "ROUND": {
      const value = Number(args[0]);
      const digits = args[1] === undefined ? 0 : Number(args[1]);
      if (Number.isNaN(value) || Number.isNaN(digits)) {
        return NaN;
      }
      const factor = 10 ** digits;
      return Math.round(value * factor) / factor;
    }
    case "CONCAT":
      return args.map((arg) => asComparableValue(arg)).join("");
    case "SUBSTRING": {
      const source = asComparableValue(args[0]);
      const start = Math.max(0, Number(args[1]) - 1);
      if (Number.isNaN(start)) {
        return "";
      }
      if (args[2] === undefined) {
        return source.slice(start);
      }
      const len = Number(args[2]);
      if (Number.isNaN(len)) {
        return "";
      }
      return source.slice(start, start + len);
    }
    case "COALESCE":
      return args.find((arg) => arg !== null && arg !== undefined && arg !== "");
    case "NOW":
      return new Date().toISOString();
    default:
      return undefined;
  }
}

function evaluateComparison(node: FilterComparisonNode, rowValues: Record<string, any>): boolean {
  const left = evaluateOperand(node.leftOperand, rowValues);
  const right = node.value;

  switch (node.comparator) {
    case "IS NULL":
      return left === null || left === undefined || left === "";
    case "IS NOT NULL":
      return !(left === null || left === undefined || left === "");
    case "LIKE":
    case "ILIKE":
    case "NOT LIKE": {
      const source = asComparableValue(left);
      const pattern = asComparableValue(right);
      const regex = new RegExp(
        `^${escapeRegExp(pattern).replace(/%/g, ".*").replace(/_/g, ".")}$`,
        node.comparator === "ILIKE" ? "i" : "",
      );
      const matched = regex.test(source);
      return node.comparator === "LIKE" || node.comparator === "ILIKE" ? matched : !matched;
    }
    case "NOT ILIKE":
      return !new RegExp(
        `^${escapeRegExp(asComparableValue(right)).replace(/%/g, ".*").replace(/_/g, ".")}$`,
        "i",
      ).test(asComparableValue(left));
    case "=":
      return asComparableValue(left) === asComparableValue(right);
    case "!=":
    case "<>":
      return asComparableValue(left) !== asComparableValue(right);
    case ">": {
      const pair = toComparablePair(left, right);
      return pair.left > pair.right;
    }
    case ">=": {
      const pair = toComparablePair(left, right);
      return pair.left >= pair.right;
    }
    case "<": {
      const pair = toComparablePair(left, right);
      return pair.left < pair.right;
    }
    case "<=": {
      const pair = toComparablePair(left, right);
      return pair.left <= pair.right;
    }
    default:
      return false;
  }
}

function evaluateFilterNode(node: FilterNode, rowValues: Record<string, any>): boolean {
  if (node.type === "comparison") {
    return evaluateComparison(node, rowValues);
  }
  if (node.operator === "AND") {
    return evaluateFilterNode(node.left, rowValues) && evaluateFilterNode(node.right, rowValues);
  }
  return evaluateFilterNode(node.left, rowValues) || evaluateFilterNode(node.right, rowValues);
}

function collectColumnsFromNode(node: FilterNode): string[] {
  if (node.type === "comparison") {
    const collectFromOperand = (operand: FilterOperandArg): string[] => {
      if (operand.type === "literal") {
        return [];
      }
      if (operand.type === "column") {
        return [operand.column];
      }
      return operand.args.flatMap(collectFromOperand);
    };
    return collectFromOperand(node.leftOperand);
  }
  return [...collectColumnsFromNode(node.left), ...collectColumnsFromNode(node.right)];
}

function validateFilterAstColumns(ast: FilterNode, columns: string[]): {
  valid: boolean;
  message?: string;
} {
  const allowed = new Set(columns.map((column) => column.toLowerCase()));
  const unknown = collectColumnsFromNode(ast).find(
    (column) => !allowed.has(column.toLowerCase()),
  );
  if (unknown) {
    return {
      valid: false,
      message: `筛选字段不存在: ${unknown}`,
    };
  }
  return { valid: true };
}

// 校验筛选表达式语法，供输入框提交前提示错误。
export function validateFilterExpression(
  expression: string,
  columns?: string[],
): {
  valid: boolean;
  message?: string;
} {
  const parsed = parseFilterExpression(expression);
  if (!parsed.valid || !parsed.ast) {
    return parsed;
  }
  if (!columns || columns.length === 0) {
    return parsed;
  }
  return validateFilterAstColumns(parsed.ast, columns);
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
  const keyword = filterKeyword.trim();
  const parsed = parseFilterExpression(keyword);
  const columnCheck =
    parsed.valid && parsed.ast ? validateFilterAstColumns(parsed.ast, columns) : { valid: false };
  const filtered = rows.filter((row) => {
    if (!keyword || !parsed.valid || !parsed.ast || !columnCheck.valid) {
      return true;
    }
    return evaluateFilterNode(parsed.ast, row.values);
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
