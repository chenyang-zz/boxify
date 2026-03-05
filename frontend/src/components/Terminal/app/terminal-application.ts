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

import { terminalSessionManager } from "../lib/session-manager";
import type { TerminalSessionEvent } from "../lib/session-manager";
import { useTerminalStore } from "../store/terminal.store";
import { v4 as uuid } from "uuid";

interface CommandExecutionContext {
  workPath?: string;
  gitBranch?: string;
}

class TerminalApplication {
  private boundSessions = new Set<string>();

  // 绑定会话事件到 store，集中处理终端业务状态更新。
  bindSession(sessionId: string): void {
    if (this.boundSessions.has(sessionId)) return;

    terminalSessionManager.setEventCallback(sessionId, (event) => {
      this.handleSessionEvent(event);
    });
    this.boundSessions.add(sessionId);
  }

  // 解除会话绑定，通常由会话销毁流程触发。
  unbindSession(sessionId: string): void {
    if (!this.boundSessions.has(sessionId)) return;
    terminalSessionManager.setEventCallback(sessionId, undefined);
    this.boundSessions.delete(sessionId);
  }

  // 提交命令：统一处理空命令、block 创建与历史记录。
  async submitCommand(
    sessionId: string,
    command: string,
    context?: CommandExecutionContext,
  ): Promise<void> {
    const trimmed = command.trim();
    const store = useTerminalStore.getState();

    // 空命令只发送回车。
    if (!trimmed) {
      await terminalSessionManager.write(sessionId, "\r");
      return;
    }

    // clear 直接清空当前会话 block 列表，不创建新的命令 block。
    if (trimmed === "clear") {
      store.clearBlocks(sessionId);
      return;
    }

    const blockId = uuid();
    store.createBlock(sessionId, trimmed, blockId, context);
    store.addToHistory(sessionId, trimmed);

    const resolvedBlockId = await terminalSessionManager.writeCommand(
      sessionId,
      trimmed,
      blockId,
    );
    if (!resolvedBlockId) {
      store.finalizeBlock(sessionId, blockId, 1);
    }
  }

  private handleSessionEvent(event: TerminalSessionEvent): void {
    const store = useTerminalStore.getState();

    switch (event.type) {
      case "output_batch":
        store.appendOutputBatch(event.sessionId, event.blockId, event.chunks);
        break;
      case "error":
        if (event.blockId) {
          store.appendOutputBatch(event.sessionId, event.blockId, [
            {
              content: event.content,
              formattedContent: event.formattedContent,
            },
          ]);
          store.finalizeBlock(event.sessionId, event.blockId, 1);
        }
        break;
      case "command_end":
        store.finalizeBlock(event.sessionId, event.blockId, event.exitCode);
        break;
      case "session_destroyed":
        store.clearSession(event.sessionId);
        this.unbindSession(event.sessionId);
        break;
    }
  }
}

export const terminalApplication = new TerminalApplication();
