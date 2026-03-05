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
import {
  useInteractiveMode,
  useSessionBlocks,
  useSelectedBlockId,
  useTerminalStore,
} from "./store";
import { TerminalBlock } from "./components/TerminalBlock";
import { InputEditor } from "./components/InputEditor";
import { FullscreenTerminal } from "./components/FullscreenTerminal";
import type { TerminalConfig } from "@/types/property";

interface TerminalCoreProps {
  sessionId: string;
  config: TerminalConfig;
}

export function TerminalCore({ sessionId, config }: TerminalCoreProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const inInteractive = useInteractiveMode(sessionId);
  const { containerRef, envInfo } = useTerminalController({
    sessionId,
    config,
    autoResize: !inInteractive,
  });

  const blocks = useSessionBlocks(sessionId);
  const selectedBlockId = useSelectedBlockId(sessionId);
  const setSelectedBlock = useTerminalStore((state) => state.setSelectedBlock);
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

  if (inInteractive) {
    return (
      <div
        ref={containerRef}
        className="terminal-core h-full w-full text-left bg-background"
      >
        <FullscreenTerminal sessionId={sessionId} />
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className="terminal-core h-full w-full flex flex-col text-left bg-background"
    >
      {/* 输出区域 */}
      <div
        ref={scrollRef}
        className="output-area flex-1 overflow-y-auto overflow-x-hidden min-w-0"
        style={terminalScrollStyle}
        onClick={() => setSelectedBlock(sessionId, undefined)}
      >
        <div className="min-h-full flex flex-col justify-end min-w-0">
          {blocks.map((block) => (
            <TerminalBlock
              key={block.id}
              block={block}
              isActive={selectedBlockId === block.id}
              onSelect={() => setSelectedBlock(sessionId, block.id)}
            />
          ))}
        </div>
      </div>

      {/* 输入区域 */}
      {envInfo && (
        <div className="input-area border-t">
          <InputEditor
            sessionId={sessionId}
            onSubmit={handleCommandSubmit}
            onResize={scrollToBottom}
            envInfo={envInfo}
            inInteractive={inInteractive}
          />
        </div>
      )}
    </div>
  );
}
