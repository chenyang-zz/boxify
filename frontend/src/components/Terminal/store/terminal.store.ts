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
import type { TerminalBlock, OutputLine, BlockStatus } from "../types/block";
import {
  appendBatchToBlockLastLine,
  createRunningBlock,
  finalizeBlock as finalizeBlockReducer,
  updateBlockStatus as updateBlockStatusReducer,
} from "../domain/block-reducer";
import {
  selectReviewPanelOpen,
  selectSessionBlocks,
} from "./selectors";

// 全局终端状态
interface TerminalState {
  // 所有会话的 blocks（按 sessionId 分组，使用普通对象）
  sessionBlocks: Record<string, TerminalBlock[]>;

  // 输入历史（按 sessionId）
  sessionHistory: Record<string, string[]>;
  historyIndexes: Record<string, number>;

  // 每个会话的代码审查面板开关
  reviewPanelOpenBySession: Record<string, boolean>;

  // === Block 操作 ===
  createBlock: (
    sessionId: string,
    command: string,
    blockId?: string,
    context?: { workPath?: string; gitBranch?: string },
  ) => string;
  appendOutputBatch: (
    sessionId: string,
    blockId: string,
    chunks: Array<{
      content: string;
      formattedContent: OutputLine["formattedContent"];
    }>,
  ) => void;
  finalizeBlock: (sessionId: string, blockId: string, exitCode: number) => void;
  updateBlockStatus: (
    sessionId: string,
    blockId: string,
    status: BlockStatus,
  ) => void;
  clearBlocks: (sessionId: string) => void;

  // === 历史操作 ===
  addToHistory: (sessionId: string, command: string) => void;
  navigateHistory: (
    sessionId: string,
    direction: "up" | "down",
  ) => string | null;
  resetHistoryIndex: (sessionId: string) => void;

  // === 审查面板 ===
  openReviewPanel: (sessionId: string) => void;
  closeReviewPanel: (sessionId: string) => void;

  // === 会话管理 ===
  clearSession: (sessionId: string) => void;
}

export const useTerminalStore = create<TerminalState>((set, get) => ({
  sessionBlocks: {},
  sessionHistory: {},
  historyIndexes: {},
  reviewPanelOpenBySession: {},

  createBlock: (
    sessionId: string,
    command: string,
    blockId?: string,
    context?: { workPath?: string; gitBranch?: string },
  ) => {
    const id = blockId || uuid();
    const newBlock = createRunningBlock(id, command, context, {
      now: Date.now,
    });

    set((state) => ({
      sessionBlocks: {
        ...state.sessionBlocks,
        [sessionId]: [...(state.sessionBlocks[sessionId] || []), newBlock],
      },
    }));

    return id;
  },

  appendOutputBatch: (sessionId, blockId, chunks) => {
    if (chunks.length === 0) return;

    set((state) => {
      const blocks = state.sessionBlocks[sessionId];
      if (!blocks) return state;
      const nextBlocks = appendBatchToBlockLastLine(blocks, blockId, chunks, {
        now: Date.now,
        createLineId: uuid,
      });
      if (!nextBlocks) return state;

      return {
        sessionBlocks: {
          ...state.sessionBlocks,
          [sessionId]: nextBlocks,
        },
      };
    });
  },

  finalizeBlock: (sessionId: string, blockId: string, exitCode: number) => {
    set((state) => {
      const blocks = state.sessionBlocks[sessionId];
      if (!blocks) return state;
      const nextBlocks = finalizeBlockReducer(blocks, blockId, exitCode, {
        now: Date.now,
      });
      if (!nextBlocks) return state;

      return {
        sessionBlocks: {
          ...state.sessionBlocks,
          [sessionId]: nextBlocks,
        },
      };
    });
  },

  updateBlockStatus: (
    sessionId: string,
    blockId: string,
    status: BlockStatus,
  ) => {
    set((state) => {
      const blocks = state.sessionBlocks[sessionId];
      if (!blocks) return state;
      const nextBlocks = updateBlockStatusReducer(blocks, blockId, status);
      if (!nextBlocks) return state;

      return {
        sessionBlocks: {
          ...state.sessionBlocks,
          [sessionId]: nextBlocks,
        },
      };
    });
  },

  clearBlocks: (sessionId: string) => {
    set((state) => ({
      sessionBlocks: {
        ...state.sessionBlocks,
        [sessionId]: [],
      },
    }));
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

  openReviewPanel: (sessionId: string) => {
    set((state) => ({
      reviewPanelOpenBySession: {
        ...state.reviewPanelOpenBySession,
        [sessionId]: true,
      },
    }));
  },

  closeReviewPanel: (sessionId: string) => {
    set((state) => ({
      reviewPanelOpenBySession: {
        ...state.reviewPanelOpenBySession,
        [sessionId]: false,
      },
    }));
  },

  clearSession: (sessionId: string) => {
    set((state) => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [sessionId]: _, ...restBlocks } = state.sessionBlocks;
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [sessionId]: ___, ...restHistory } = state.sessionHistory;
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [sessionId]: ____, ...restIndexes } = state.historyIndexes;
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [sessionId]: _____, ...restReviewOpen } = state.reviewPanelOpenBySession;

      return {
        sessionBlocks: restBlocks,
        sessionHistory: restHistory,
        historyIndexes: restIndexes,
        reviewPanelOpenBySession: restReviewOpen,
      };
    });
  },
}));

// 选择器 hooks
export function useSessionBlocks(sessionId: string): TerminalBlock[] {
  return useTerminalStore(selectSessionBlocks(sessionId));
}

export function useReviewPanelOpen(sessionId: string): boolean {
  return useTerminalStore(selectReviewPanelOpen(sessionId));
}
