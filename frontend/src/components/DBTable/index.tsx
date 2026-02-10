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

import { FC, useLayoutEffect, useRef, useState } from "react";
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/table";
import { useDBTable } from "../../hooks/useDBTable";
import { usePropertyStore } from "@/store/property.store";
import { cn } from "@/lib/utils";
import { useResizeObserver } from "@/hooks/use-resize-observer";

const DBTable: FC = () => {
  const selectedUUID = usePropertyStore((state) => state.selectedUUID);
  const { columns, values } = useDBTable(selectedUUID);
  const headerRefs = useRef<Array<HTMLTableCellElement | null>>([]);
  const [stickyLefts, setStickyLefts] = useState<number[]>([0, 0, 0]);
  const { ref: containerRef, size } = useResizeObserver<HTMLDivElement>();

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
  }, [columns, size.width]);

  const stickyStyle = (colIndex: number): React.CSSProperties | undefined => {
    if (colIndex > 2) {
      return undefined;
    }
    return {
      left: stickyLefts[colIndex] ?? 0,
      boxShadow: "1px 0 0 var(--border)",
    };
  };

  const stickyHeadClass = (colIndex: number) =>
    colIndex <= 2 ? "sticky z-20 bg-muted" : "";

  const stickyCellClass = (colIndex: number) =>
    colIndex <= 2 ? "sticky z-10 bg-card" : "";

  return (
    <div ref={containerRef} className="bg-card h-full rounded-lg flex flex-col">
      <div className="h-full flex flex-col">
        <header className="shrink-0">toubu</header>
        <main className="flex-1 flex text-xs outline outline-background">
          <aside className="shrink-0 flex flex-col h-full outline outline-background">
            {new Array(values.length + 1).fill(0).map((_, index) => (
              <span
                key={index}
                className="h-8 flex px-2 justify-center items-center "
              >
                {index === 0 ? "" : index}
              </span>
            ))}
          </aside>
          <section className="flex-1">
            <Table className="w-full ">
              <TableHeader className="bg-muted">
                <TableRow>
                  <TableHead
                    ref={setHeaderRef(0)}
                    className={cn(
                      "py-0 h-8 outline outline-border border-y",
                      stickyHeadClass(0),
                    )}
                    style={stickyStyle(0)}
                  />
                  {columns.map((col, index) => {
                    const colIndex = index + 1;
                    const isSticky = colIndex <= 2;
                    return (
                      <TableHead
                        key={col}
                        ref={isSticky ? setHeaderRef(colIndex) : undefined}
                        className={cn(
                          "px-4 py-0 h-8 text-center truncate border",
                          stickyHeadClass(colIndex),
                        )}
                        style={stickyStyle(colIndex)}
                      >
                        {col}
                      </TableHead>
                    );
                  })}
                </TableRow>
              </TableHeader>
              <TableBody className="shadow">
                {values.map((row, rowIndex) => (
                  <TableRow key={rowIndex} className="h-8">
                    <TableCell
                      className={cn(
                        "px-3 py-0 outline outline-border border-y",
                        stickyCellClass(0),
                      )}
                      style={stickyStyle(0)}
                    >
                      {rowIndex + 1}
                    </TableCell>
                    {columns.map((col, index) => {
                      const colIndex = index + 1;
                      return (
                        <TableCell
                          key={col}
                          className={cn(
                            "px-4 py-0 max-w-150 truncate border text-left",
                            stickyCellClass(colIndex),
                          )}
                          style={stickyStyle(colIndex)}
                        >
                          {row[col]}
                        </TableCell>
                      );
                    })}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </section>
        </main>
        <footer className="shrink-0">dibu</footer>
      </div>
    </div>
  );
};

export default DBTable;
