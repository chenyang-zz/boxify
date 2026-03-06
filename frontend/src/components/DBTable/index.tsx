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
import { useDBTableController } from "./app/use-db-table-controller";

interface DBTableProps {
  sessionId: string;
}

// DBTable 组件：负责表格渲染与交互绑定。
const DBTable: FC<DBTableProps> = ({ sessionId: uuid }) => {
  const headerRefs = useRef<Array<HTMLTableCellElement | null>>([]);
  const [stickyLefts, setStickyLefts] = useState<number[]>([0, 0, 0]);
  const { ref: containerRef, size } = useResizeObserver<HTMLDivElement>();
  const tableScrollRef = useHorizontalScroll({ hideScrollbar: false });

  const propertyItem = useMemo(() => getPropertyItemByUUID(uuid), [uuid]);
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
    colIndex <= 2 ? "sticky z-20 bg-muted" : "";

  const stickyCellClass = (colIndex: number) =>
    colIndex <= 2 ? "sticky z-10 bg-card" : "";

  const sql = `SELECT * FROM ${propertyItem.label} LIMIT 0,500`;

  return (
    <div ref={containerRef} className="bg-card h-full flex flex-col">
      <div className="h-full flex flex-col text-xs">
        <HeaderAction
          state={controller.actionState}
          onToggleTransaction={controller.toggleTransaction}
          onRefresh={controller.load}
          onAddRow={controller.addRow}
          onDeleteRows={controller.deleteSelectedRows}
          onSave={controller.save}
          onUndo={controller.undo}
          onRedo={controller.redo}
          onToggleFilter={controller.toggleFilterInput}
          onSort={controller.toggleSort}
          onImport={controller.importData}
          onExport={controller.exportData}
        />
        {controller.actionState.showFilterInput && (
          <div className="px-2 py-1 border-b border-border">
            <Input
              value={controller.actionState.filterKeyword}
              onChange={(event) =>
                controller.setFilterKeyword(event.target.value)
              }
              placeholder="输入关键字筛选当前结果"
              className="h-7 text-xs"
            />
          </div>
        )}
        <main className="flex-1 flex outline outline-background min-h-0">
          <aside className="shrink-0 flex flex-col h-full outline outline-background bg-muted/20">
            {new Array(controller.rows.length + 1).fill(0).map((_, index) => (
              <span
                key={index}
                className="h-8 flex px-2 justify-center items-center outline outline-background first:bg-muted"
              >
                {index === 0 ? "" : index}
              </span>
            ))}
          </aside>
          <section ref={tableScrollRef} className="flex-1 overflow-auto">
            <Table className="w-full">
              <TableHeader className="bg-muted">
                <TableRow className="border-0">
                  {controller.columns.map((col, index) => {
                    const colIndex = index;
                    const isSticky = colIndex <= 2;
                    return (
                      <TableHead
                        key={col}
                        ref={isSticky ? setHeaderRef(colIndex) : undefined}
                        className={cn(
                          "px-4 py-0 h-8 text-center truncate outline outline-background cursor-pointer",
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
              <TableBody className="shadow">
                {controller.rows.map(({ row }) => {
                  const selected = controller.selectedRowIds.has(row.id);
                  return (
                    <TableRow
                      key={row.id}
                      className={cn(
                        "h-8 border-0",
                        selected && "bg-accent/40",
                        row.deleted && "opacity-55",
                      )}
                      onClick={() => controller.toggleRowSelection(row.id)}
                    >
                      {controller.columns.map((col, index) => {
                        const colIndex = index;
                        const value = row.values[col];
                        const displayValue =
                          value === null || value === undefined
                            ? ""
                            : String(value);

                        return (
                          <TableCell
                            key={`${row.id}-${col}`}
                            className={cn(
                              "px-2 py-0 max-w-150 truncate text-left outline outline-background",
                              stickyCellClass(colIndex),
                              row.deleted && "line-through",
                            )}
                            style={stickyStyle(colIndex)}
                            onClick={(event) => {
                              event.stopPropagation();
                              controller.setSelectedColumn(col);
                            }}
                          >
                            {controller.actionState.inTransaction &&
                            !row.deleted ? (
                              <Input
                                value={displayValue}
                                className="h-6 border-0 bg-transparent px-2 text-xs"
                                onChange={(event) =>
                                  controller.setCellValue(
                                    row.id,
                                    col,
                                    event.target.value,
                                  )
                                }
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
        <footer className="shrink-0 text-left px-0.5 flex items-center text-[10px]">
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
