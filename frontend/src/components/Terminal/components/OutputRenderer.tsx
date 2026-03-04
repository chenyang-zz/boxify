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

import { memo, useRef, useCallback } from "react";
import type { OutputLine } from "../types/block";

interface OutputRendererProps {
  output: OutputLine[];
}

export const OutputRenderer = memo(function OutputRenderer({
  output,
}: OutputRendererProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  // 渲染单行
  const renderLine = useCallback(
    (line: OutputLine, index: number) => {
      return (
        <div
          key={line.id || index}
          className="output-line whitespace-pre-wrap break-all py-0.5"
          style={{
            fontFamily:
              '"Sarasa Mono SC", "JetBrainsMono Nerd", "Fira Code", "Consolas", monospace',
            fontSize: 13,
            lineHeight: 1.4,
          }}
        >
          {line.formattedContent.map((char, charIndex) => {
            // 跳过控制字符（回车、退格等），只渲染可见字符和换行符
            if (char.char < " " && char.char !== "\n" && char.char !== "\t") {
              return null;
            }

            // 合并相邻的相同样式字符
            const style: React.CSSProperties = {
              color: char.style.fg || "#c9d1d9",
              backgroundColor: char.style.bg,
              fontWeight: char.style.bold ? "bold" : undefined,
              fontStyle: char.style.italic ? "italic" : undefined,
              textDecoration: char.style.underline
                ? "underline"
                : char.style.strikethrough
                  ? "line-through"
                  : undefined,
              opacity: char.style.dim ? 0.5 : char.style.hidden ? 0 : 1,
            };

            // 处理反色
            if (char.style.inverse) {
              const temp = style.color;
              style.color = style.backgroundColor || "#1e1e1e";
              style.backgroundColor = temp || "#c9d1d9";
            }

            return (
              <span key={charIndex} style={style}>
                {char.char}
              </span>
            );
          })}
        </div>
      );
    },
    [],
  );

  // 处理文本选择和复制
  const handleCopy = useCallback(async () => {
    const selection = window.getSelection();
    if (selection && selection.toString()) {
      // 浏览器默认复制行为
      return;
    }

    // 如果没有选择，复制所有输出
    const text = output.map((l) => l.content).join("\n");
    await navigator.clipboard.writeText(text);
  }, [output]);

  if (output.length === 0) return null;

  return (
    <div
      ref={containerRef}
      className="output-renderer py-1 select-text"
      onCopy={handleCopy}
    >
      {output.map((line, index) => renderLine(line, index))}
    </div>
  );
});
