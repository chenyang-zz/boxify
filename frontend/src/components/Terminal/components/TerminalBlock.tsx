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

import { memo, useCallback } from "react";
import { Copy, Loader2, CheckCircle, XCircle } from "lucide-react";
import type { TerminalBlock as TerminalBlockType } from "../types/block";
import { OutputRenderer } from "./OutputRenderer";
import { cn } from "@/lib/utils";

interface TerminalBlockProps {
  block: TerminalBlockType;
}

export const TerminalBlock = memo(function TerminalBlock({
  block,
}: TerminalBlockProps) {
  // 复制命令和输出
  const handleCopy = useCallback(async () => {
    const text = [
      `$ ${block.command}`,
      ...block.output.map((l) => l.content),
    ].join("\n");
    await navigator.clipboard.writeText(text);
  }, [block]);

  // 格式化执行时间
  const formatDuration = useCallback((ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`;
  }, []);

  // 状态图标
  const StatusIcon = {
    running: <Loader2 className="w-4 h-4 animate-spin text-blue-400" />,
    success: <CheckCircle className="w-4 h-4 text-green-400" />,
    error: <XCircle className="w-4 h-4 text-red-400" />,
    pending: (
      <div
        className="w-4 h-4 rounded-full border-2"
        style={{ borderColor: "#6e7681" }}
      />
    ),
  }[block.status];

  return (
    <div
      className={cn(
        "terminal-block border-t py-3 px-4 relative",
        block.status === "error" &&
          "bg-red-400/10 before:w-1 before:h-full before:bg-red-400 before:absolute before:left-0 before:top-0",
      )}
    >
      {/* Block 头部 */}
      <div className="block-header flex items-center gap-2">
        <div className=" select-none text-xs text-secondary-foreground flex gap-2">
          <span>base</span>
          {block.workPath && <span>{block.workPath}</span>}
          {block.gitBranch && <span>git:({block.gitBranch})</span>}
        </div>

        {/* 命令 */}
        <span className="command-text flex-1 font-mono text-sm truncate">
          <span className=" text-green-500 font-bold select-none">$ </span>
          {block.command || "(空命令)"}
        </span>

        {/* 执行时间 */}
        {block.endTime && (
          <span className="duration text-xs text-muted-foreground">
            {formatDuration(block.endTime - block.startTime)}
          </span>
        )}

        {/* 状态图标 */}
        <div className="status-indicator">{StatusIcon}</div>

        {/* 复制按钮 */}
        <button
          className="copy-btn p-1 rounded "
          onClick={(e) => {
            e.stopPropagation();
            handleCopy();
          }}
          title="复制"
        >
          <Copy className="w-4 h-4 text-muted-foreground" />
        </button>
      </div>

      {/* Block 输出 */}
      <div className="block-output ">
        <OutputRenderer output={block.output} />
      </div>
    </div>
  );
});
