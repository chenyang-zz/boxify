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

import { useEffect, useRef, useCallback, useState, useMemo } from "react";
import { terminalSessionManager } from "./lib/session-manager";
import { useTerminalStore } from "./store/terminal.store";
import { TerminalBlock } from "./components/TerminalBlock";
import { InputEditor } from "./components/InputEditor";
import type { TerminalConfig } from "@/types/property";

interface TerminalCoreProps {
  sessionId: string;
  config: TerminalConfig;
}

export function TerminalCore({ sessionId, config }: TerminalCoreProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const [isInitialized, setIsInitialized] = useState(false);

  // 直接从 store 获取数据
  const sessionBlocks = useTerminalStore((state) => state.sessionBlocks);
  const theme = useTerminalStore((state) => state.currentTheme);
  const toggleBlockCollapse = useTerminalStore((state) => state.toggleBlockCollapse);
  const addToHistory = useTerminalStore((state) => state.addToHistory);
  const createBlock = useTerminalStore((state) => state.createBlock);

  // 使用 useMemo 稳定 blocks 引用
  const blocks = useMemo(() => {
    return sessionBlocks[sessionId] || [];
  }, [sessionBlocks, sessionId]);

  // 初始化后端会话
  useEffect(() => {
    if (isInitialized) return;

    const session = terminalSessionManager.getOrCreate(sessionId);

    if (!session.isInitialized) {
      terminalSessionManager.initialize(sessionId, config).then(() => {
        setIsInitialized(true);
      });
    } else {
      setIsInitialized(true);
    }
  }, [sessionId, config, isInitialized]);

  // 自动滚动到底部
  const scrollToBottom = useCallback(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, []);

  // 处理命令提交
  const handleCommandSubmit = useCallback(
    async (command: string) => {
      const trimmed = command.trim();

      // 空命令只发送回车
      if (!trimmed) {
        await terminalSessionManager.write(sessionId, "\r");
        return;
      }

      // 后端生成并返回 blockId
      const blockId = await terminalSessionManager.writeCommand(sessionId, trimmed);

      // 使用后端返回的 blockId 创建 block
      if (blockId) {
        createBlock(sessionId, trimmed, blockId);
      }
      addToHistory(sessionId, trimmed);
      setTimeout(scrollToBottom, 50);
    },
    [sessionId, createBlock, addToHistory, scrollToBottom],
  );

  // 处理终端大小变化
  useEffect(() => {
    const handleResize = () => {
      if (containerRef.current) {
        const { clientWidth, clientHeight } = containerRef.current;
        const charWidth = 8;
        const charHeight = 16;
        const cols = Math.floor(clientWidth / charWidth);
        const rows = Math.floor(clientHeight / charHeight);
        terminalSessionManager.resize(sessionId, cols, rows);
      }
    };

    handleResize();
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, [sessionId]);

  // 处理折叠
  const handleToggleCollapse = useCallback(
    (blockId: string) => {
      toggleBlockCollapse(sessionId, blockId);
    },
    [sessionId, toggleBlockCollapse],
  );

  return (
    <div
      ref={containerRef}
      className="terminal-core h-full w-full flex flex-col"
      style={{
        backgroundColor: theme.background,
        color: theme.foreground,
        fontFamily: theme.fontFamily,
        fontSize: theme.fontSize,
        lineHeight: theme.lineHeight,
        textAlign: "left",
      }}
    >
      {/* 输出区域 */}
      <div
        ref={scrollRef}
        className="output-area flex-1 overflow-auto p-2"
        style={{
          scrollbarWidth: "thin",
          scrollbarColor: `${theme.brightBlack} transparent`,
        }}
      >
        {blocks.map((block) => (
          <TerminalBlock
            key={block.id}
            block={block}
            theme={theme}
            onToggleCollapse={() => handleToggleCollapse(block.id)}
          />
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
      <div
        className="input-area border-t"
        style={{ borderColor: theme.blockStyle.borderColor }}
      >
        <InputEditor
          sessionId={sessionId}
          theme={theme}
          onSubmit={handleCommandSubmit}
          onResize={scrollToBottom}
        />
      </div>
    </div>
  );
}
