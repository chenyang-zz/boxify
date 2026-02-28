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

export type BlockStatus = "running" | "success" | "error" | "pending";

export interface TextStyle {
  fg?: string;
  bg?: string;
  bold?: boolean;
  italic?: boolean;
  underline?: boolean;
  dim?: boolean;
  blink?: boolean;
  inverse?: boolean;
  hidden?: boolean;
  strikethrough?: boolean;
}

export interface FormattedChar {
  char: string;
  style: TextStyle;
}

export interface OutputLine {
  id: string;
  content: string;
  formattedContent: FormattedChar[];
  timestamp: number;
}

export interface TerminalBlock {
  id: string;
  command: string;
  output: OutputLine[];
  status: BlockStatus;
  startTime: number;
  endTime?: number;
  exitCode?: number;
  isCollapsed: boolean;
}

// 用于创建新 block 的参数
export interface CreateBlockParams {
  command: string;
}

// 用于更新 block 输出的参数
export interface UpdateBlockOutputParams {
  blockId: string;
  content: string;
  append?: boolean;
}
