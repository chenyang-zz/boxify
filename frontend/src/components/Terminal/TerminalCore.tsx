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

import { useRef, useCallback } from "react";
import { terminalApplication, useTerminalController } from "./app";
import { useSessionBlocks } from "./store";
import { TerminalBlock } from "./components/TerminalBlock";
import { InputEditor } from "./components/InputEditor";
import type { TerminalConfig } from "@/types/property";

interface TerminalCoreProps {
  sessionId: string;
  config: TerminalConfig;
}

export function TerminalCore({ sessionId, config }: TerminalCoreProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const { containerRef, envInfo } = useTerminalController({ sessionId, config });

  const blocks = useSessionBlocks(sessionId);
  const terminalScrollStyle = {
    scrollbarWidth: "thin" as const,
    scrollbarColor: "#6e7681 transparent",
  };

  // 自动滚动到底部
  const scrollToBottom = useCallback(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, []);

  // 处理命令提交
  const handleCommandSubmit = useCallback(
    async (command: string) => {
      await terminalApplication.submitCommand(sessionId, command, {
        workPath: envInfo?.workPath,
        gitBranch:
          (envInfo as { gitInfo?: { branch?: string } } | undefined)?.gitInfo
            ?.branch || undefined,
      });
      setTimeout(scrollToBottom, 50);
    },
    [sessionId, scrollToBottom, envInfo],
  );

  return (
    <div
      ref={containerRef}
      className="terminal-core h-full w-full flex flex-col text-left bg-background"
    >
      {/* 输出区域 */}
      <div
        ref={scrollRef}
        className="output-area flex-1 overflow-auto"
        style={terminalScrollStyle}
      >
        {blocks.map((block) => (
          <TerminalBlock key={block.id} block={block} />
        ))}

        {/* 空状态提示 */}
        {blocks.length === 0 && (
          <div className="empty-state text-center py-8 opacity-50">
            <p>终端已就绪</p>
            <p className="text-sm mt-1">输入命令开始</p>
          </div>
        )}
      </div>

      {/* 输入区域 */}
      {envInfo && (
        <div className="input-area border-t">
          <InputEditor
            sessionId={sessionId}
            onSubmit={handleCommandSubmit}
            onResize={scrollToBottom}
            envInfo={envInfo}
          />
        </div>
      )}
    </div>
  );
}
