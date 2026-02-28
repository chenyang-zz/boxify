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
import { Copy, ChevronRight, Loader2, CheckCircle, XCircle } from "lucide-react";
import type { TerminalBlock as TerminalBlockType } from "../types/block";
import type { TerminalTheme } from "../types/theme";
import { OutputRenderer } from "./OutputRenderer";

interface TerminalBlockProps {
  block: TerminalBlockType;
  theme: TerminalTheme;
  onToggleCollapse: () => void;
}

export const TerminalBlock = memo(function TerminalBlock({
  block,
  theme,
  onToggleCollapse,
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
    running: <Loader2 className="w-4 h-4 animate-spin" style={{ color: theme.blue }} />,
    success: <CheckCircle className="w-4 h-4" style={{ color: theme.green }} />,
    error: <XCircle className="w-4 h-4" style={{ color: theme.red }} />,
    pending: <div className="w-4 h-4 rounded-full border-2" style={{ borderColor: theme.brightBlack }} />,
  }[block.status];

  return (
    <div
      className="terminal-block group mb-2"
      style={{
        background: theme.blockStyle.background,
        border: `1px solid ${theme.blockStyle.borderColor}`,
        borderRadius: theme.blockStyle.borderRadius,
      }}
    >
      {/* Block 头部 */}
      <div
        className="block-header flex items-center gap-2 px-3 py-2 cursor-pointer select-none"
        onClick={onToggleCollapse}
      >
        {/* 折叠按钮 */}
        <button
          className="collapse-btn p-0.5 rounded hover:bg-white/10 transition-colors"
          onClick={(e) => {
            e.stopPropagation();
            onToggleCollapse();
          }}
        >
          <ChevronRight
            className="w-4 h-4 transition-transform duration-200"
            style={{
              transform: block.isCollapsed ? "rotate(0deg)" : "rotate(90deg)",
              color: theme.brightBlack,
            }}
          />
        </button>

        {/* 命令 */}
        <span
          className="command-text flex-1 font-mono text-sm truncate"
          style={{ color: theme.foreground }}
        >
          <span style={{ color: theme.green, fontWeight: "bold" }}>$ </span>
          {block.command || "(空命令)"}
        </span>

        {/* 执行时间 */}
        {block.endTime && (
          <span
            className="duration text-xs"
            style={{ color: theme.brightBlack }}
          >
            {formatDuration(block.endTime - block.startTime)}
          </span>
        )}

        {/* 状态图标 */}
        <div className="status-indicator">{StatusIcon}</div>

        {/* 复制按钮 */}
        <button
          className="copy-btn p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-white/10 transition-all"
          onClick={(e) => {
            e.stopPropagation();
            handleCopy();
          }}
          title="复制"
        >
          <Copy className="w-4 h-4" style={{ color: theme.brightBlack }} />
        </button>
      </div>

      {/* Block 输出 */}
      {!block.isCollapsed && block.output.length > 0 && (
        <div
          className="block-output border-t"
          style={{ borderColor: theme.blockStyle.borderColor }}
        >
          <OutputRenderer
            output={block.output}
            theme={theme}
            blockId={block.id}
          />
        </div>
      )}
    </div>
  );
});
