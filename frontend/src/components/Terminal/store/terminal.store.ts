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

import { create } from "zustand";
import { v4 as uuid } from "uuid";
import type {
  TerminalBlock,
  OutputLine,
  BlockStatus,
} from "../types/block";
import type { TerminalTheme } from "../types/theme";
import { defaultTheme } from "../types/theme";

// 全局终端状态
interface TerminalState {
  // 所有会话的 blocks（按 sessionId 分组，使用普通对象）
  sessionBlocks: Record<string, TerminalBlock[]>;

  // 当前活动的 block（按 sessionId）
  currentBlockIds: Record<string, string | null>;

  // 输入历史（按 sessionId）
  sessionHistory: Record<string, string[]>;
  historyIndexes: Record<string, number>;

  // 主题
  currentTheme: TerminalTheme;

  // === Block 操作 ===
  createBlock: (sessionId: string, command: string, blockId?: string) => string;
  updateBlockOutput: (
    sessionId: string,
    blockId: string,
    content: string,
    formattedContent: OutputLine["formattedContent"]
  ) => void;
  appendToLastLine: (
    sessionId: string,
    blockId: string,
    content: string,
    formattedContent: OutputLine["formattedContent"]
  ) => void;
  finalizeBlock: (sessionId: string, blockId: string, exitCode: number) => void;
  updateBlockStatus: (sessionId: string, blockId: string, status: BlockStatus) => void;
  toggleBlockCollapse: (sessionId: string, blockId: string) => void;

  // === 历史操作 ===
  addToHistory: (sessionId: string, command: string) => void;
  navigateHistory: (sessionId: string, direction: "up" | "down") => string | null;
  resetHistoryIndex: (sessionId: string) => void;

  // === 会话管理 ===
  clearSession: (sessionId: string) => void;

  // === 主题 ===
  setTheme: (theme: TerminalTheme) => void;
}

export const useTerminalStore = create<TerminalState>((set, get) => ({
  sessionBlocks: {},
  currentBlockIds: {},
  sessionHistory: {},
  historyIndexes: {},
  currentTheme: defaultTheme,

  createBlock: (sessionId: string, command: string, blockId?: string) => {
    const id = blockId || uuid();
    const newBlock: TerminalBlock = {
      id,
      command,
      output: [],
      status: "running",
      startTime: Date.now(),
      isCollapsed: false,
    };

    set((state) => ({
      sessionBlocks: {
        ...state.sessionBlocks,
        [sessionId]: [...(state.sessionBlocks[sessionId] || []), newBlock],
      },
      currentBlockIds: {
        ...state.currentBlockIds,
        [sessionId]: id,
      },
    }));

    return id;
  },

  updateBlockOutput: (
    sessionId: string,
    blockId: string,
    content: string,
    formattedContent: OutputLine["formattedContent"]
  ) => {
    const newLine: OutputLine = {
      id: uuid(),
      content,
      formattedContent,
      timestamp: Date.now(),
    };

    set((state) => {
      const blocks = state.sessionBlocks[sessionId];
      if (!blocks) return state;

      return {
        sessionBlocks: {
          ...state.sessionBlocks,
          [sessionId]: blocks.map((block) =>
            block.id === blockId
              ? { ...block, output: [...block.output, newLine] }
              : block
          ),
        },
      };
    });
  },

  appendToLastLine: (
    sessionId: string,
    blockId: string,
    content: string,
    formattedContent: OutputLine["formattedContent"]
  ) => {
    set((state) => {
      const blocks = state.sessionBlocks[sessionId];
      if (!blocks) return state;

      const blockIndex = blocks.findIndex((b) => b.id === blockId);
      if (blockIndex === -1) return state;

      const block = blocks[blockIndex];
      const lastLine = block.output[block.output.length - 1];

      let newOutput: OutputLine[];

      if (lastLine) {
        newOutput = [
          ...block.output.slice(0, -1),
          {
            ...lastLine,
            content: lastLine.content + content,
            formattedContent: [...lastLine.formattedContent, ...formattedContent],
          },
        ];
      } else {
        const newLine: OutputLine = {
          id: uuid(),
          content,
          formattedContent,
          timestamp: Date.now(),
        };
        newOutput = [newLine];
      }

      return {
        sessionBlocks: {
          ...state.sessionBlocks,
          [sessionId]: blocks.map((b, i) =>
            i === blockIndex ? { ...b, output: newOutput } : b
          ),
        },
      };
    });
  },

  finalizeBlock: (sessionId: string, blockId: string, exitCode: number) => {
    set((state) => {
      const blocks = state.sessionBlocks[sessionId];
      if (!blocks) return state;

      return {
        sessionBlocks: {
          ...state.sessionBlocks,
          [sessionId]: blocks.map((block) =>
            block.id === blockId
              ? {
                  ...block,
                  status: exitCode === 0 ? "success" : "error",
                  endTime: Date.now(),
                  exitCode,
                }
              : block
          ),
        },
        currentBlockIds: {
          ...state.currentBlockIds,
          [sessionId]: null,
        },
      };
    });
  },

  updateBlockStatus: (sessionId: string, blockId: string, status: BlockStatus) => {
    set((state) => {
      const blocks = state.sessionBlocks[sessionId];
      if (!blocks) return state;

      return {
        sessionBlocks: {
          ...state.sessionBlocks,
          [sessionId]: blocks.map((block) =>
            block.id === blockId ? { ...block, status } : block
          ),
        },
      };
    });
  },

  toggleBlockCollapse: (sessionId: string, blockId: string) => {
    set((state) => {
      const blocks = state.sessionBlocks[sessionId];
      if (!blocks) return state;

      return {
        sessionBlocks: {
          ...state.sessionBlocks,
          [sessionId]: blocks.map((block) =>
            block.id === blockId
              ? { ...block, isCollapsed: !block.isCollapsed }
              : block
          ),
        },
      };
    });
  },

  addToHistory: (sessionId: string, command: string) => {
    if (!command.trim()) return;

    set((state) => ({
      sessionHistory: {
        ...state.sessionHistory,
        [sessionId]: [...(state.sessionHistory[sessionId] || []), command],
      },
      historyIndexes: {
        ...state.historyIndexes,
        [sessionId]: -1,
      },
    }));
  },

  navigateHistory: (sessionId: string, direction: "up" | "down") => {
    const { sessionHistory, historyIndexes } = get();
    const history = sessionHistory[sessionId] || [];
    const currentIndex = historyIndexes[sessionId] ?? -1;

    if (history.length === 0) return null;

    let newIndex: number;
    if (direction === "up") {
      newIndex = Math.min(currentIndex + 1, history.length - 1);
    } else {
      newIndex = Math.max(currentIndex - 1, -1);
    }

    set((state) => ({
      historyIndexes: {
        ...state.historyIndexes,
        [sessionId]: newIndex,
      },
    }));

    if (newIndex === -1) return "";
    return history[history.length - 1 - newIndex];
  },

  resetHistoryIndex: (sessionId: string) => {
    set((state) => ({
      historyIndexes: {
        ...state.historyIndexes,
        [sessionId]: -1,
      },
    }));
  },

  clearSession: (sessionId: string) => {
    set((state) => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [sessionId]: _, ...restBlocks } = state.sessionBlocks;
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [sessionId]: __, ...restBlockIds } = state.currentBlockIds;
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [sessionId]: ___, ...restHistory } = state.sessionHistory;
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [sessionId]: ____, ...restIndexes } = state.historyIndexes;

      return {
        sessionBlocks: restBlocks,
        currentBlockIds: restBlockIds,
        sessionHistory: restHistory,
        historyIndexes: restIndexes,
      };
    });
  },

  setTheme: (theme: TerminalTheme) => {
    set({ currentTheme: theme });
  },
}));

// 选择器 hooks
export function useSessionBlocks(sessionId: string): TerminalBlock[] {
  return useTerminalStore((state) => state.sessionBlocks[sessionId] || []);
}

export function useCurrentBlockId(sessionId: string): string | null {
  return useTerminalStore((state) => state.currentBlockIds[sessionId] ?? null);
}

export function useSessionTheme(): TerminalTheme {
  return useTerminalStore((state) => state.currentTheme);
}
