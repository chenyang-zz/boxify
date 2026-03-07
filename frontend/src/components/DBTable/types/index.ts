import type { ChangeSet, ColumnDefinition } from "@wails/connection";
import type { DBTableExportFormat } from "@/lib/db-table";

export type DBTableSortDirection = "none" | "asc" | "desc";

export interface DBTableSortState {
  column: string | null;
  direction: DBTableSortDirection;
}

export type DBTableRowMode = "existing" | "inserted";

export interface DBTableDraftRow {
  id: string;
  mode: DBTableRowMode;
  values: Record<string, any>;
  originalValues?: Record<string, any>;
  deleted?: boolean;
}

export interface DBTableRenderRow {
  row: DBTableDraftRow;
  index: number;
}

export interface DBTableActionState {
  inTransaction: boolean;
  dirty: boolean;
  pending: boolean;
  canUndo: boolean;
  canRedo: boolean;
  hasSelection: boolean;
  showFilterInput: boolean;
  filterKeyword: string;
  filterError: string | null;
  sortState: DBTableSortState;
}

export interface DBTableControllerResult {
  columns: string[];
  columnDefs: ColumnDefinition[];
  rows: DBTableRenderRow[];
  selectedRowIds: Set<string>;
  selectedColumn: string | null;
  actionState: DBTableActionState;
  load: () => Promise<void>;
  startTransaction: () => void;
  saveTransaction: () => Promise<void>;
  commitTransaction: () => Promise<void>;
  rollbackTransaction: () => void;
  addRow: () => void;
  deleteSelectedRows: () => void;
  undo: () => void;
  redo: () => void;
  endCellEditSession: () => void;
  toggleFilterInput: () => void;
  setFilterKeyword: (keyword: string) => void;
  applyFilter: (expression: string) => void;
  toggleSort: () => void;
  importData: () => Promise<void>;
  exportData: (format: DBTableExportFormat) => Promise<void>;
  setCellValue: (rowId: string, column: string, value: string) => void;
  toggleRowSelection: (rowId: string) => void;
  setSelectedColumn: (column: string | null) => void;
}

export interface DBTableChangeSummary {
  inserts: number;
  updates: number;
  deletes: number;
}

export interface DBTableBuiltChangeSet {
  changes: ChangeSet;
  summary: DBTableChangeSummary;
}
