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

import type { TerminalBlock, BlockStatus } from "./block";
import type { TerminalTheme } from "./theme";

export interface TerminalSize {
  cols: number;
  rows: number;
  width: number;
  height: number;
}

export interface TerminalSession {
  id: string;
  config: TerminalSessionConfig;
  blocks: TerminalBlock[];
  currentBlockId: string | null;
  inputHistory: string[];
  historyIndex: number;
  size: TerminalSize;
}

export interface TerminalSessionConfig {
  shell: string;
  workPath: string;
  initialCommand?: string;
}

export interface TerminalState {
  sessions: Map<string, TerminalSession>;
  currentTheme: TerminalTheme;
}

// AI 建议类型（预留）
export interface CommandSuggestion {
  id: string;
  command: string;
  description?: string;
  source: "history" | "ai" | "snippet";
}

// 终端事件
export interface TerminalOutputEvent {
  sessionId: string;
  data: string;
}

export interface TerminalErrorEvent {
  sessionId: string;
  message: string;
}

export interface TerminalResizeEvent {
  sessionId: string;
  cols: number;
  rows: number;
}

// 输入编辑器状态
export interface InputEditorState {
  value: string;
  cursorPosition: number;
  selection: {
    start: number;
    end: number;
  } | null;
  isMultiline: boolean;
}
