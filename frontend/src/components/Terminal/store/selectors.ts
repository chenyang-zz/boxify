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

import type { TerminalBlock } from "../types/block";

interface TerminalSelectorState {
  sessionBlocks: Record<string, TerminalBlock[]>;
  reviewPanelOpenBySession: Record<string, boolean>;
}

const EMPTY_BLOCKS: TerminalBlock[] = [];

// 读取会话 blocks，缺省返回稳定空数组以减少渲染抖动。
export function selectSessionBlocks(sessionId: string) {
  return (state: TerminalSelectorState): TerminalBlock[] =>
    state.sessionBlocks[sessionId] ?? EMPTY_BLOCKS;
}

// 读取会话级审查面板开关。
export function selectReviewPanelOpen(sessionId: string) {
  return (state: TerminalSelectorState): boolean =>
    state.reviewPanelOpenBySession[sessionId] ?? false;
}
