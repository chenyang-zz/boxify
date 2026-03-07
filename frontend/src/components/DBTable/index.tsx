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
  FC,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/table";
import { cn, copyText } from "@/lib/utils";
import { useResizeObserver } from "@/hooks/use-resize-observer";
import { CopyIcon } from "lucide-react";
import { Button } from "../ui/button";
import { getPropertyItemByUUID } from "@/lib/property";
import { useHorizontalScroll } from "@/hooks/useHorizontalScroll";
import { Input } from "../ui/input";
import HeaderAction from "./components/HeaderAction";
import FilterExpressionInput from "./components/FilterExpressionInput";
import { useDBTableController } from "./app/use-db-table-controller";
import { ConnectionEnum } from "@/common/constrains";
import type { PropertyItemType } from "@/types/property";

interface DBTableProps {
  sessionId: string;
}

// 从当前属性节点向上追溯连接节点，解析数据库类型。
function resolveDatabaseTypeFromProperty(
  item: PropertyItemType | null,
): ConnectionEnum.MYSQL | ConnectionEnum.POSTGRESQL | null {
  let cursor = item;
  while (cursor) {
    if (
      cursor.type === ConnectionEnum.MYSQL ||
      cursor.type === ConnectionEnum.POSTGRESQL
    ) {
      return cursor.type;
    }
    cursor = cursor.parent ?? null;
  }
  return null;
}

// DBTable 组件：负责表格渲染与交互绑定。
const DBTable: FC<DBTableProps> = ({ sessionId: uuid }) => {
  type HighlightMode = "row" | "cell";

  const headerRefs = useRef<Array<HTMLTableCellElement | null>>([]);
  const [stickyLefts, setStickyLefts] = useState<number[]>([0, 0, 0]);
  const [editingCell, setEditingCell] = useState<{
    rowId: string;
    column: string;
  } | null>(null);
  const [highlightMode, setHighlightMode] = useState<HighlightMode>("row");
  const [selectedCell, setSelectedCell] = useState<{
    rowId: string;
    column: string;
  } | null>(null);
  const { ref: containerRef, size } = useResizeObserver<HTMLDivElement>();
  const tableScrollRef = useHorizontalScroll({ hideScrollbar: false });

  const propertyItem = useMemo(() => getPropertyItemByUUID(uuid), [uuid]);
  const databaseType = useMemo(
    () => resolveDatabaseTypeFromProperty(propertyItem),
    [propertyItem],
  );
  const controller = useDBTableController({ sessionId: uuid });

  useEffect(() => {
    controller.load();
  }, [uuid]);

  if (!propertyItem) {
    return null;
  }

  const setHeaderRef = (index: number) => (el: HTMLTableCellElement | null) => {
    headerRefs.current[index] = el;
  };

  useLayoutEffect(() => {
    const widths = headerRefs.current
      .slice(0, 3)
      .map((el) => el?.offsetWidth ?? 0);
    if (widths.length < 3) {
      return;
    }
    const lefts = [1, widths[0] + 1, widths[0] + 1 + widths[1] + 1];
    setStickyLefts(lefts);
  }, [controller.columns, size.width]);

  const stickyStyle = (colIndex: number): React.CSSProperties | undefined => {
    if (colIndex > 2) {
      return undefined;
    }
    return {
      left: stickyLefts[colIndex] ?? 0,
      boxShadow: "1px 0 4px var(--border)",
    };
  };

  const stickyHeadClass = (colIndex: number) =>
    colIndex <= 2 ? "sticky z-20" : "";

  const stickyCellClass = (colIndex: number) =>
    colIndex <= 2 ? "sticky z-10" : "";

  const sql = `SELECT * FROM ${propertyItem.label} LIMIT 0,500`;

  // 处理左侧行标点击，切换对应数据行选中状态。
  const handleSidebarRowClick = (rowId: string) => {
    setHighlightMode("row");
    setSelectedCell(null);
    controller.setSelectedColumn(null);
    controller.toggleRowSelection(rowId);
  };

  return (
    <div ref={containerRef} className="h-full bg-card flex flex-col">
      <div className="h-full flex flex-col text-xs">
        <HeaderAction
          state={controller.actionState}
          onStartTransaction={controller.startTransaction}
          onSaveTransaction={controller.saveTransaction}
          onCommitTransaction={controller.commitTransaction}
          onRollbackTransaction={controller.rollbackTransaction}
          onRefresh={controller.load}
          onAddRow={controller.addRow}
          onDeleteRows={controller.deleteSelectedRows}
          onUndo={controller.undo}
          onToggleFilter={controller.toggleFilterInput}
          onSort={controller.toggleSort}
          onImport={controller.importData}
          onExport={controller.exportData}
        />
        {controller.actionState.showFilterInput && (
          <FilterExpressionInput
            value={controller.actionState.filterKeyword}
            columns={controller.columns}
            databaseType={databaseType}
            pending={controller.actionState.pending}
            error={controller.actionState.filterError ?? undefined}
            onChange={controller.setFilterKeyword}
            onApply={controller.applyFilter}
          />
        )}
        <main className="flex-1 flex outline outline-background min-h-0">
          <aside className="shrink-0 flex flex-col h-full outline outline-background bg-background">
            <div className="h-8 w-14 outline outline-background bg-card" />
            {controller.rows.map(({ row }, index) => {
              const selected = controller.selectedRowIds.has(row.id);
              return (
                <button
                  key={row.id}
                  type="button"
                  className={cn(
                    "h-8 w-14 flex items-center justify-center outline outline-background cursor-pointer",
                    selected && "bg-accent",
                  )}
                  onClick={() => handleSidebarRowClick(row.id)}
                >
                  {selected ? "●" : index + 1}
                </button>
              );
            })}
          </aside>
          <section
            ref={tableScrollRef}
            className="flex-1 overflow-auto bg-background"
          >
            <Table className="w-full">
              <TableHeader>
                <TableRow className="border-0">
                  {controller.columns.map((col, index) => {
                    const colIndex = index;
                    const isSticky = colIndex <= 2;
                    return (
                      <TableHead
                        key={col}
                        ref={isSticky ? setHeaderRef(colIndex) : undefined}
                        className={cn(
                          "px-4 py-0 h-8 text-center truncate outline outline-background cursor-pointer bg-card",
                          stickyHeadClass(colIndex),
                          controller.selectedColumn === col && "bg-accent",
                        )}
                        style={stickyStyle(colIndex)}
                        onClick={() => controller.setSelectedColumn(col)}
                      >
                        {col}
                      </TableHead>
                    );
                  })}
                </TableRow>
              </TableHeader>
              <TableBody className="shadow bg-background">
                {controller.rows.map(({ row }) => {
                  const selected = controller.selectedRowIds.has(row.id);
                  const rowHighlighted = selected && highlightMode === "row";
                  return (
                    <TableRow
                      key={row.id}
                      className={cn(
                        "h-8 border-0",
                        rowHighlighted && "bg-accent/40",
                        row.deleted && "opacity-55",
                      )}
                      onClick={() => handleSidebarRowClick(row.id)}
                    >
                      {controller.columns.map((col, index) => {
                        const colIndex = index;
                        const value = row.values[col];
                        const cellHighlighted =
                          selected &&
                          highlightMode === "cell" &&
                          selectedCell?.rowId === row.id &&
                          selectedCell?.column === col;
                        const isEditing =
                          editingCell?.rowId === row.id &&
                          editingCell.column === col;
                        const displayValue =
                          value === null || value === undefined
                            ? ""
                            : String(value);

                        return (
                          <TableCell
                            key={`${row.id}-${col}`}
                            className={cn(
                              "p-0 h-8 max-w-150 truncate text-left outline outline-background",
                              stickyCellClass(colIndex),
                              (rowHighlighted || cellHighlighted) &&
                                "bg-primary/30 ",
                              row.deleted && "line-through",
                            )}
                            style={stickyStyle(colIndex)}
                            onClick={(event) => {
                              event.stopPropagation();
                              setHighlightMode("cell");
                              setSelectedCell({ rowId: row.id, column: col });
                              controller.toggleRowSelection(row.id);
                              controller.setSelectedColumn(col);
                            }}
                            onDoubleClick={(event) => {
                              event.stopPropagation();
                              if (row.deleted) {
                                return;
                              }
                              setHighlightMode("cell");
                              setSelectedCell({ rowId: row.id, column: col });
                              controller.toggleRowSelection(row.id);
                              controller.setSelectedColumn(col);
                              setEditingCell({ rowId: row.id, column: col });
                            }}
                          >
                            {!row.deleted && isEditing ? (
                              <Input
                                value={displayValue}
                                className="h-full w-full rounded-none border-0 bg-transparent px-0 py-0 text-xs shadow-none focus-visible:border-0 focus-visible:ring-0"
                                autoFocus
                                onChange={(event) =>
                                  controller.setCellValue(
                                    row.id,
                                    col,
                                    event.target.value,
                                  )
                                }
                                onBlur={() => {
                                  controller.endCellEditSession();
                                  setEditingCell(null);
                                }}
                                onClick={(event) => event.stopPropagation()}
                                onKeyDown={(event) => {
                                  if (
                                    event.key === "Enter" ||
                                    event.key === "Escape"
                                  ) {
                                    controller.endCellEditSession();
                                    setEditingCell(null);
                                  }
                                }}
                              />
                            ) : (
                              displayValue
                            )}
                          </TableCell>
                        );
                      })}
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </section>
        </main>
        <footer className="shrink-0 text-left px-0.5 bg-card flex items-center text-[10px]">
          <Button
            size="icon"
            variant="ghost"
            className="rounded-full size-7"
            onClick={() => copyText(sql)}
          >
            <CopyIcon className="size-3" />
          </Button>
          {sql}
        </footer>
      </div>
    </div>
  );
};

export default DBTable;
