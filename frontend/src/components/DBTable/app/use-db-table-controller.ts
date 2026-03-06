import { useCallback, useMemo, useRef, useState } from "react";
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
  buildChangeSetBetween,
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
  const [transactionSnapshot, setTransactionSnapshot] = useState<DBTableDraftRow[] | null>(null);
  const editingCellKeyRef = useRef<string | null>(null);

  const primaryColumns = useMemo(
    () => columnDefs.filter((col) => col.key === "PRI").map((col) => col.name),
    [columnDefs],
  );

  const dirty = useMemo(
    () => hasPendingChanges(rows, columns, primaryColumns),
    [rows, columns, primaryColumns],
  );

  // 刷新表格数据并按需决定是否保持在事务中。
  const reloadWithTransactionState = useCallback(
    async (keepTransaction: boolean) => {
      if (!sessionId) {
        return;
      }

      const [columnRes, valueRes] = await Promise.all([
        getDBTableColumnsByUUID(sessionId),
        getDBTableValuesByUUID(sessionId),
      ]);
      const nextRows = createDraftRows((valueRes?.data as Record<string, any>[]) ?? []);

      setColumnDefs(columnRes);
      setColumns(valueRes?.fields ?? []);
      setRows(nextRows);
      setSelectedRowIds(new Set());
      setSelectedColumn(null);
      setInTransaction(keepTransaction);
      setTransactionSnapshot(keepTransaction ? transactionSnapshot ?? cloneDraftRows(nextRows) : null);
      setUndoStack([]);
      setRedoStack([]);
      setInsertCounter(0);
    },
    [sessionId, transactionSnapshot],
  );

  // 加载列定义和数据，并保持当前事务开关状态。
  const load = useCallback(async () => {
    if (!sessionId) {
      return;
    }

    setPending(true);
    try {
      await reloadWithTransactionState(inTransaction);
    } finally {
      setPending(false);
    }
  }, [inTransaction, reloadWithTransactionState, sessionId]);

  const pushHistory = useCallback((nextRows: DBTableDraftRow[]) => {
    setUndoStack((prev) => [...prev, cloneDraftRows(rows)]);
    setRedoStack([]);
    setRows(nextRows);
  }, [rows]);

  // 开启编辑事务并记录起始快照，用于后续回退。
  const startTransaction = useCallback(() => {
    if (inTransaction) {
      return;
    }

    setTransactionSnapshot(cloneDraftRows(rows));
    setUndoStack([]);
    setRedoStack([]);
    setInTransaction(true);
  }, [inTransaction, rows]);

  // 执行保存：未开启事务时普通保存；开启事务时可选择是否结束事务。
  const applyTransactionChanges = useCallback(async (closeTransaction: boolean) => {
    const built = buildChangeSet(rows, columns, primaryColumns);
    const total = built.summary.inserts + built.summary.updates + built.summary.deletes;
    if (total === 0) {
      if (inTransaction && closeTransaction) {
        setInTransaction(false);
        setTransactionSnapshot(null);
        setUndoStack([]);
        setRedoStack([]);
      }
      return;
    }

    setPending(true);
    try {
      const applyResult = await applyDBTableChangesByUUID(sessionId, built.changes);
      if (!applyResult?.success) {
        return;
      }
      toast.success(
        `保存成功：新增 ${built.summary.inserts}，更新 ${built.summary.updates}，删除 ${built.summary.deletes}`,
      );
      await reloadWithTransactionState(inTransaction && !closeTransaction);
    } catch {
      // 调用层已统一弹出错误提示；这里中断后续成功提示和刷新。
    } finally {
      setPending(false);
    }
  }, [columns, inTransaction, primaryColumns, reloadWithTransactionState, rows, sessionId]);

  // 保存事务内草稿变更，但保持事务开启。
  const saveTransaction = useCallback(async () => {
    await applyTransactionChanges(false);
  }, [applyTransactionChanges]);

  // 提交事务内草稿变更到数据库并结束事务。
  const commitTransaction = useCallback(async () => {
    await applyTransactionChanges(true);
  }, [applyTransactionChanges]);

  // 回退事务：将数据库和界面恢复到开始事务前的快照状态。
  const rollbackTransaction = useCallback(async () => {
    const snapshot = transactionSnapshot ? cloneDraftRows(transactionSnapshot) : [];

    setPending(true);
    try {
      const valueRes = await getDBTableValuesByUUID(sessionId);
      const dbRows = createDraftRows((valueRes?.data as Record<string, any>[]) ?? []);
      const built = buildChangeSetBetween(dbRows, snapshot, columns, primaryColumns);
      const total = built.summary.inserts + built.summary.updates + built.summary.deletes;

      if (total > 0) {
        await applyDBTableChangesByUUID(sessionId, built.changes);
      }

      setRows(snapshot);
      setSelectedRowIds(new Set());
      setUndoStack([]);
      setRedoStack([]);
      setTransactionSnapshot(null);
      setInTransaction(false);
      toast.info("已回退事务更改");
    } finally {
      setPending(false);
    }
  }, [columns, primaryColumns, sessionId, transactionSnapshot]);

  // 新增一行空数据；不自动切换事务状态。
  const addRow = useCallback(() => {
    const nextRow = createInsertedRow(columns, insertCounter + 1);
    setInsertCounter((prev) => prev + 1);
    pushHistory([...rows, nextRow]);
    setSelectedRowIds(new Set([nextRow.id]));
  }, [columns, insertCounter, pushHistory, rows]);

  // 标记删除或恢复选中行。
  const deleteSelectedRows = useCallback(() => {
    if (selectedRowIds.size === 0) {
      toast.warning("请先选择要删除的行");
      return;
    }

    const nextRows = toggleRowsDeleted(rows, selectedRowIds);
    pushHistory(nextRows);
    setSelectedRowIds(new Set());
  }, [pushHistory, rows, selectedRowIds]);

  // 撤销最近一次草稿编辑。
  const undo = useCallback(() => {
    if (undoStack.length === 0) {
      return;
    }

    const previous = undoStack[undoStack.length - 1];
    setUndoStack((prev) => prev.slice(0, -1));
    setRedoStack((prev) => [...prev, cloneDraftRows(rows)]);
    setRows(cloneDraftRows(previous));
  }, [rows, undoStack]);

  // 重做最近一次被撤销的草稿编辑。
  const redo = useCallback(() => {
    if (redoStack.length === 0) {
      return;
    }

    const next = redoStack[redoStack.length - 1];
    setRedoStack((prev) => prev.slice(0, -1));
    setUndoStack((prev) => [...prev, cloneDraftRows(rows)]);
    setRows(cloneDraftRows(next));
  }, [redoStack, rows]);

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
      const cellKey = `${rowId}::${column}`;
      if (editingCellKeyRef.current === cellKey) {
        setRedoStack([]);
        setRows((prev) => updateCellValue(prev, rowId, column, value));
        return;
      }

      editingCellKeyRef.current = cellKey;
      pushHistory(updateCellValue(rows, rowId, column, value));
    },
    [pushHistory, rows],
  );

  // 结束当前单元格编辑会话，使后续编辑重新计入一条撤销记录。
  const endCellEditSession = useCallback(() => {
    editingCellKeyRef.current = null;
  }, []);

  // 切换行选中状态。
  const toggleRowSelection = useCallback((rowId: string) => {
    setSelectedRowIds(new Set([rowId]));
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
    startTransaction,
    saveTransaction,
    commitTransaction,
    rollbackTransaction,
    addRow,
    deleteSelectedRows,
    undo,
    redo,
    endCellEditSession,
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
