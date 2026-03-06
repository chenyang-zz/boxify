import { useCallback, useMemo, useState } from "react";
import { toast } from "sonner";
import type { ColumnDefinition } from "@wails/connection";
import {
  applyDBTableChangesByUUID,
  exportDBTableByUUID,
  getDBTableColumnsByUUID,
  getDBTableValuesByUUID,
  importDBTableByUUID,
  type DBTableExportFormat,
} from "@/lib/db-table";
import {
  buildChangeSet,
  cloneDraftRows,
  createDraftRows,
  createInsertedRow,
  hasPendingChanges,
  toRenderRows,
  toggleRowsDeleted,
  updateCellValue,
} from "../domain/draft";
import type {
  DBTableControllerResult,
  DBTableDraftRow,
  DBTableSortDirection,
  DBTableSortState,
} from "../types";

interface UseDBTableControllerOptions {
  sessionId: string;
}

// DBTable 控制器：负责数据加载、编辑事务与头部动作编排。
export function useDBTableController({
  sessionId,
}: UseDBTableControllerOptions): DBTableControllerResult {
  const [columns, setColumns] = useState<string[]>([]);
  const [columnDefs, setColumnDefs] = useState<ColumnDefinition[]>([]);
  const [rows, setRows] = useState<DBTableDraftRow[]>([]);
  const [pending, setPending] = useState(false);
  const [inTransaction, setInTransaction] = useState(false);
  const [selectedRowIds, setSelectedRowIds] = useState<Set<string>>(new Set());
  const [selectedColumn, setSelectedColumn] = useState<string | null>(null);
  const [showFilterInput, setShowFilterInput] = useState(false);
  const [filterKeyword, setFilterKeyword] = useState("");
  const [sortState, setSortState] = useState<DBTableSortState>({
    column: null,
    direction: "none",
  });
  const [insertCounter, setInsertCounter] = useState(0);
  const [undoStack, setUndoStack] = useState<DBTableDraftRow[][]>([]);
  const [redoStack, setRedoStack] = useState<DBTableDraftRow[][]>([]);

  const primaryColumns = useMemo(
    () => columnDefs.filter((col) => col.key === "PRI").map((col) => col.name),
    [columnDefs],
  );

  const dirty = useMemo(
    () => hasPendingChanges(rows, columns, primaryColumns),
    [rows, columns, primaryColumns],
  );

  // 加载列定义和数据，并重置编辑状态。
  const load = useCallback(async () => {
    if (!sessionId) {
      return;
    }

    setPending(true);
    try {
      const [columnRes, valueRes] = await Promise.all([
        getDBTableColumnsByUUID(sessionId),
        getDBTableValuesByUUID(sessionId),
      ]);

      setColumnDefs(columnRes);
      setColumns(valueRes?.fields ?? []);
      setRows(createDraftRows((valueRes?.data as Record<string, any>[]) ?? []));
      setSelectedRowIds(new Set());
      setSelectedColumn(null);
      setInTransaction(false);
      setUndoStack([]);
      setRedoStack([]);
      setInsertCounter(0);
    } finally {
      setPending(false);
    }
  }, [sessionId]);

  const pushHistory = useCallback((nextRows: DBTableDraftRow[]) => {
    setUndoStack((prev) => [...prev, cloneDraftRows(rows)]);
    setRedoStack([]);
    setRows(nextRows);
  }, [rows]);

  // 切换编辑事务状态；关闭时会丢弃未保存草稿。
  const toggleTransaction = useCallback(() => {
    if (!inTransaction) {
      setInTransaction(true);
      return;
    }

    if (dirty) {
      setRows((prev) =>
        prev
          .filter((row) => row.mode === "existing")
          .map((row) => ({
            ...row,
            values: { ...(row.originalValues ?? row.values) },
            deleted: false,
          })),
      );
      setUndoStack([]);
      setRedoStack([]);
      toast.info("已放弃未保存更改");
    }

    setInTransaction(false);
  }, [dirty, inTransaction]);

  // 事务内新增一行空数据。
  const addRow = useCallback(() => {
    if (!inTransaction) {
      toast.warning("请先开始事务");
      return;
    }

    const nextRow = createInsertedRow(columns, insertCounter + 1);
    setInsertCounter((prev) => prev + 1);
    pushHistory([...rows, nextRow]);
    setSelectedRowIds(new Set([nextRow.id]));
  }, [columns, inTransaction, insertCounter, pushHistory, rows]);

  // 标记删除或恢复选中行。
  const deleteSelectedRows = useCallback(() => {
    if (!inTransaction) {
      toast.warning("请先开始事务");
      return;
    }

    if (selectedRowIds.size === 0) {
      toast.warning("请先选择要删除的行");
      return;
    }

    const nextRows = toggleRowsDeleted(rows, selectedRowIds);
    pushHistory(nextRows);
    setSelectedRowIds(new Set());
  }, [inTransaction, pushHistory, rows, selectedRowIds]);

  // 保存草稿变更并刷新数据。
  const save = useCallback(async () => {
    if (!inTransaction) {
      toast.warning("请先开始事务");
      return;
    }

    const built = buildChangeSet(rows, columns, primaryColumns);
    const total = built.summary.inserts + built.summary.updates + built.summary.deletes;
    if (total === 0) {
      toast.info("没有可保存的变更");
      return;
    }

    setPending(true);
    try {
      await applyDBTableChangesByUUID(sessionId, built.changes);
      toast.success(
        `保存成功：新增 ${built.summary.inserts}，更新 ${built.summary.updates}，删除 ${built.summary.deletes}`,
      );
      await load();
    } finally {
      setPending(false);
    }
  }, [columns, inTransaction, load, primaryColumns, rows, sessionId]);

  // 撤销最近一次草稿编辑。
  const undo = useCallback(() => {
    if (!inTransaction) {
      toast.warning("请先开始事务");
      return;
    }
    if (undoStack.length === 0) {
      return;
    }

    const previous = undoStack[undoStack.length - 1];
    setUndoStack((prev) => prev.slice(0, -1));
    setRedoStack((prev) => [...prev, cloneDraftRows(rows)]);
    setRows(cloneDraftRows(previous));
  }, [inTransaction, rows, undoStack]);

  // 重做最近一次被撤销的草稿编辑。
  const redo = useCallback(() => {
    if (!inTransaction) {
      toast.warning("请先开始事务");
      return;
    }
    if (redoStack.length === 0) {
      return;
    }

    const next = redoStack[redoStack.length - 1];
    setRedoStack((prev) => prev.slice(0, -1));
    setUndoStack((prev) => [...prev, cloneDraftRows(rows)]);
    setRows(cloneDraftRows(next));
  }, [inTransaction, redoStack, rows]);

  // 切换筛选输入框显示状态。
  const toggleFilterInput = useCallback(() => {
    setShowFilterInput((prev) => !prev);
    if (showFilterInput) {
      setFilterKeyword("");
    }
  }, [showFilterInput]);

  // 按当前选中列切换排序方向。
  const toggleSort = useCallback(() => {
    const targetColumn = selectedColumn ?? columns[0] ?? null;
    if (!targetColumn) {
      return;
    }

    let nextDirection: DBTableSortDirection = "asc";
    if (sortState.column === targetColumn) {
      if (sortState.direction === "asc") {
        nextDirection = "desc";
      } else if (sortState.direction === "desc") {
        nextDirection = "none";
      }
    }

    setSortState({
      column: nextDirection === "none" ? null : targetColumn,
      direction: nextDirection,
    });
  }, [columns, selectedColumn, sortState.column, sortState.direction]);

  // 更新指定单元格值。
  const setCellValue = useCallback(
    (rowId: string, column: string, value: string) => {
      if (!inTransaction) {
        return;
      }

      pushHistory(updateCellValue(rows, rowId, column, value));
    },
    [inTransaction, pushHistory, rows],
  );

  // 切换行选中状态。
  const toggleRowSelection = useCallback((rowId: string) => {
    setSelectedRowIds((prev) => {
      const next = new Set(prev);
      if (next.has(rowId)) {
        next.delete(rowId);
      } else {
        next.add(rowId);
      }
      return next;
    });
  }, []);

  // 导入数据文件并刷新。
  const importData = useCallback(async () => {
    setPending(true);
    try {
      await importDBTableByUUID(sessionId);
      await load();
      toast.success("导入完成");
    } finally {
      setPending(false);
    }
  }, [load, sessionId]);

  // 导出为指定格式。
  const exportData = useCallback(async (format: DBTableExportFormat) => {
    setPending(true);
    try {
      await exportDBTableByUUID(sessionId, format);
      toast.success(`导出成功 (${format.toUpperCase()})`);
    } finally {
      setPending(false);
    }
  }, [sessionId]);

  const renderRows = useMemo(
    () => toRenderRows(rows, columns, filterKeyword, sortState),
    [columns, filterKeyword, rows, sortState],
  );

  return {
    columns,
    columnDefs,
    rows: renderRows,
    selectedRowIds,
    selectedColumn,
    actionState: {
      inTransaction,
      dirty,
      pending,
      canUndo: undoStack.length > 0,
      canRedo: redoStack.length > 0,
      hasSelection: selectedRowIds.size > 0,
      showFilterInput,
      filterKeyword,
      sortState,
    },
    load,
    toggleTransaction,
    addRow,
    deleteSelectedRows,
    save,
    undo,
    redo,
    toggleFilterInput,
    setFilterKeyword,
    toggleSort,
    importData,
    exportData,
    setCellValue,
    toggleRowSelection,
    setSelectedColumn,
  };
}
