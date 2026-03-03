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
import { useSessionBlocks, useSessionTheme, useTerminalStore } from "./store/terminal.store";
import { TerminalBlock } from "./components/TerminalBlock";
import { InputEditor } from "./components/InputEditor";
import type { TerminalConfig } from "@/types/property";
import { TerminalEnvironmentInfo } from "@wails/types/models";

interface TerminalCoreProps {
  sessionId: string;
  config: TerminalConfig;
}

export function TerminalCore({ sessionId, config }: TerminalCoreProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const resizeRafRef = useRef<number | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [envInfo, setEnvInfo] = useState<TerminalEnvironmentInfo | undefined>(
    undefined,
  );

  const blocks = useSessionBlocks(sessionId);
  const theme = useSessionTheme();

  const addToHistory = useTerminalStore((state) => state.addToHistory);
  const createBlock = useTerminalStore((state) => state.createBlock);
  const terminalScrollStyle = useMemo(
    () => ({
      scrollbarWidth: "thin" as const,
      scrollbarColor: `${theme.brightBlack} transparent`,
    }),
    [theme.brightBlack],
  );

  // 使用 useCallback 缓存环境变化回调
  const handleEnvChange = useCallback((env: TerminalEnvironmentInfo) => {
    setEnvInfo(env);
  }, []);

  // 初始化后端会话
  useEffect(() => {
    let cancelled = false;

    const session = terminalSessionManager.getOrCreate(sessionId);

    // 设置环境变化回调
    terminalSessionManager.setEnvChangeCallback(sessionId, handleEnvChange);

    if (!session.isInitialized) {
      terminalSessionManager.initialize(sessionId, config).then((env) => {
        if (cancelled) return;
        setEnvInfo(env);
        setIsInitialized(true);
      });
    } else {
      setEnvInfo(session.environmentInfo);
      setIsInitialized(true);
    }

    return () => {
      cancelled = true;
      terminalSessionManager.setEnvChangeCallback(sessionId, undefined);
    };
  }, [sessionId, config, handleEnvChange]);

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
      const blockId = await terminalSessionManager.writeCommand(
        sessionId,
        trimmed,
      );

      // 使用后端返回的 blockId 创建 block
      if (blockId) {
        createBlock(sessionId, trimmed, blockId, {
          workPath: envInfo?.workPath,
          gitBranch:
            (envInfo as { gitInfo?: { branch?: string } } | undefined)?.gitInfo
              ?.branch || undefined,
        });
      }
      addToHistory(sessionId, trimmed);
      setTimeout(scrollToBottom, 50);
    },
    [sessionId, createBlock, addToHistory, scrollToBottom, envInfo],
  );

  // 监听容器大小变化，同步终端 rows/cols
  useEffect(() => {
    const updateSize = () => {
      if (containerRef.current) {
        const { clientWidth, clientHeight } = containerRef.current;
        const charWidth = 8;
        const charHeight = 16;
        const cols = Math.max(20, Math.floor(clientWidth / charWidth));
        const rows = Math.max(6, Math.floor(clientHeight / charHeight));
        terminalSessionManager.resize(sessionId, cols, rows);
      }
    };

    const scheduleResize = () => {
      if (resizeRafRef.current !== null) {
        cancelAnimationFrame(resizeRafRef.current);
      }
      resizeRafRef.current = requestAnimationFrame(() => {
        updateSize();
        resizeRafRef.current = null;
      });
    };

    scheduleResize();

    const observer = new ResizeObserver(() => {
      scheduleResize();
    });

    if (containerRef.current) {
      observer.observe(containerRef.current);
    }

    return () => {
      observer.disconnect();
      if (resizeRafRef.current !== null) {
        cancelAnimationFrame(resizeRafRef.current);
      }
    };
  }, [sessionId]);

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
          <TerminalBlock key={block.id} block={block} theme={theme} />
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
            theme={theme}
            onSubmit={handleCommandSubmit}
            onResize={scrollToBottom}
            envInfo={envInfo}
          />
        </div>
      )}
    </div>
  );
}
