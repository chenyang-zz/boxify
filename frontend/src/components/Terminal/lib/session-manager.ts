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

import { Events } from "@wailsio/runtime";
import { TerminalService } from "@wails/service";
import { TerminalConfig as GoTerminalConfig } from "@wails/terminal";
import type { TerminalConfig } from "@/types/property";
import { AnsiParser } from "./ansi-parser";
import { useTerminalStore } from "../store/terminal.store";
import { defaultTheme } from "../types/theme";
import { callWails } from "@/lib/utils";
import { TerminalEnvironmentInfo } from "@wails/types/models";

interface CachedSession {
  isInitialized: boolean;
  parser: AnsiParser;
  unbindCallbacks: (() => void)[];
  environmentInfo?: TerminalEnvironmentInfo;
  onEnvChange?: (env: TerminalEnvironmentInfo) => void;
}

class TerminalSessionManager {
  private sessions = new Map<string, CachedSession>();
  private defaultRows = 24;
  private defaultCols = 80;

  getOrCreate(sessionId: string): CachedSession {
    let session = this.sessions.get(sessionId);

    if (!session) {
      session = {
        isInitialized: false,
        parser: new AnsiParser(defaultTheme),
        unbindCallbacks: [],
      };
      this.sessions.set(sessionId, session);
      this.setupEventListeners(sessionId);
    }

    return session;
  }

  // 设置环境信息变化回调
  setEnvChangeCallback(
    sessionId: string,
    callback: (env: TerminalEnvironmentInfo) => void,
  ): void {
    const session = this.sessions.get(sessionId);
    if (session) {
      session.onEnvChange = callback;
    }
  }

  private setupEventListeners(sessionId: string): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    const unbindOutput = Events.On(
      "terminal:output",
      (event: {
        data: { sessionId: string; blockId?: string; data: string };
      }) => {
        if (event.data.sessionId === sessionId) {
          this.handleOutput(sessionId, event.data.data, event.data.blockId);
        }
      },
    );

    const unbindError = Events.On(
      "terminal:error",
      (event: { data: { sessionId: string; message: string } }) => {
        if (event.data.sessionId === sessionId) {
          this.handleError(sessionId, event.data.message);
        }
      },
    );

    // 命令结束事件
    const unbindCommandEnd = Events.On(
      "terminal:command_end",
      (event: {
        data: { sessionId: string; blockId: string; exitCode: number };
      }) => {
        if (event.data.sessionId === sessionId) {
          this.handleCommandEnd(
            sessionId,
            event.data.blockId,
            event.data.exitCode,
          );
        }
      },
    );

    // 工作路径更新事件
    const unbindPwdUpdate = Events.On(
      "terminal:pwd_update",
      (event: { data: { sessionId: string; pwd: string } }) => {
        if (event.data.sessionId === sessionId) {
          this.handlePwdUpdate(sessionId, event.data.pwd);
        }
      },
    );

    // Git 状态更新事件
    const unbindGitUpdate = Events.On(
      "terminal:git_update",
      (event: {
        data: {
          sessionId: string;
          git: {
            isRepo: boolean;
            branch?: string;
            modifiedFiles: number;
            addedLines: number;
            deletedLines: number;
          };
        };
      }) => {
        if (event.data.sessionId === sessionId) {
          this.handleGitUpdate(sessionId, event.data.git);
        }
      },
    );

    session.unbindCallbacks = [
      unbindOutput,
      unbindError,
      unbindCommandEnd,
      unbindPwdUpdate,
      unbindGitUpdate,
    ];
  }

  private handleOutput(
    sessionId: string,
    encodedData: string,
    blockId?: string,
  ): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    try {
      // 使用 TextDecoder 正确处理 UTF-8 编码
      const binaryString = atob(encodedData);
      const bytes = Uint8Array.from(binaryString, (c) => c.charCodeAt(0));
      const decoded = new TextDecoder("utf-8").decode(bytes);

      const formattedContent = session.parser.parse(decoded);

      const store = useTerminalStore.getState();

      // 优先使用事件中的 blockId，否则回退到 store 中的当前 blockId
      const targetBlockId = blockId ?? store.currentBlockIds[sessionId] ?? null;

      // 只在有 block 时追加输出
      if (targetBlockId) {
        store.appendToLastLine(
          sessionId,
          targetBlockId,
          decoded,
          formattedContent,
        );
      }
    } catch (e) {
      console.error("处理终端输出失败:", e);
    }
  }

  private handleError(sessionId: string, message: string): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    const store = useTerminalStore.getState();
    const currentBlockId = store.currentBlockIds[sessionId];

    if (currentBlockId) {
      const formattedContent = session.parser.parse(
        `\x1b[31m错误: ${message}\x1b[0m`,
      );
      store.appendToLastLine(
        sessionId,
        currentBlockId,
        `\n错误: ${message}`,
        formattedContent,
      );
      store.updateBlockStatus(sessionId, currentBlockId, "error");
      store.finalizeBlock(sessionId, currentBlockId, 1);
    }
  }

  private handleCommandEnd(
    sessionId: string,
    blockId: string,
    exitCode: number,
  ): void {
    const store = useTerminalStore.getState();

    // 更新 block 状态
    if (exitCode === 0) {
      store.updateBlockStatus(sessionId, blockId, "success");
    } else {
      store.updateBlockStatus(sessionId, blockId, "error");
    }

    // 完成 block
    store.finalizeBlock(sessionId, blockId, exitCode);
  }

  // 通用的环境信息更新方法
  private updateEnvironmentInfo(
    sessionId: string,
    updater: (env: TerminalEnvironmentInfo) => TerminalEnvironmentInfo,
  ): void {
    const session = this.sessions.get(sessionId);
    if (!session || !session.environmentInfo) return;

    session.environmentInfo = updater(session.environmentInfo);

    if (session.onEnvChange) {
      session.onEnvChange(session.environmentInfo);
    }
  }

  private handlePwdUpdate(sessionId: string, pwd: string): void {
    const session = this.sessions.get(sessionId);
    if (!session || !session.environmentInfo) return;

    // 检查路径是否真的变化了
    if (session.environmentInfo.workPath === pwd) return;

    this.updateEnvironmentInfo(sessionId, (env) => ({
      ...env,
      workPath: pwd,
    }));

    // 通知后端更新工作路径（用于 Git 监听器）
    TerminalService.UpdateWorkPath(sessionId, pwd).catch((err) => {
      console.error("更新工作路径失败:", err);
    });
  }

  private handleGitUpdate(
    sessionId: string,
    git: {
      isRepo: boolean;
      branch?: string;
      modifiedFiles: number;
      addedLines: number;
      deletedLines: number;
    },
  ): void {
    this.updateEnvironmentInfo(sessionId, (env) => ({
      ...env,
      gitInfo: {
        isRepo: git.isRepo,
        branch: git.branch ?? "",
        modifiedFiles: git.modifiedFiles,
        addedLines: git.addedLines,
        deletedLines: git.deletedLines,
      },
    }));
  }

  async initialize(
    sessionId: string,
    terminalConfig: TerminalConfig,
  ): Promise<TerminalEnvironmentInfo | undefined> {
    const session = this.sessions.get(sessionId);
    if (!session || session.isInitialized) return;

    try {
      const res = await callWails(
        TerminalService.Create,
        GoTerminalConfig.createFrom({
          id: sessionId,
          shell: terminalConfig.shell,
          workPath: terminalConfig.workpath,
          initialCommand: terminalConfig.initialCommand,
          rows: this.defaultRows,
          cols: this.defaultCols,
        }),
      );

      session.isInitialized = true;
      session.environmentInfo = res.data ?? undefined;
      console.log("终端会话创建成功:", sessionId);
      return session.environmentInfo;
    } catch (err) {
      console.error("创建终端失败:", err);
    }
  }

  async write(sessionId: string, data: string): Promise<void> {
    const session = this.sessions.get(sessionId);
    if (!session || !session.isInitialized) return;

    const encoded = btoa(data);
    await TerminalService.Write(sessionId, encoded);
  }

  /**
   * 写入命令并返回 block ID
   * 用于追踪命令输出，实现 block 关联
   */
  async writeCommand(sessionId: string, command: string): Promise<string> {
    const session = this.sessions.get(sessionId);
    if (!session || !session.isInitialized) return "";

    try {
      const blockId = await TerminalService.WriteCommand(sessionId, command);
      return blockId || "";
    } catch (err) {
      console.error("写入命令失败:", err);
      return "";
    }
  }

  async resize(sessionId: string, cols: number, rows: number): Promise<void> {
    const session = this.sessions.get(sessionId);
    if (!session || !session.isInitialized) return;

    await TerminalService.Resize(sessionId, rows, cols);
  }

  destroy(sessionId: string): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    session.unbindCallbacks.forEach((unbind) => unbind());

    TerminalService.Close(sessionId).catch((err) => {
      console.error("关闭终端失败:", err);
    });

    useTerminalStore.getState().clearSession(sessionId);

    this.sessions.delete(sessionId);
  }

  has(sessionId: string): boolean {
    return this.sessions.has(sessionId);
  }

  isInitialized(sessionId: string): boolean {
    const session = this.sessions.get(sessionId);
    return session?.isInitialized ?? false;
  }
}

export const terminalSessionManager = new TerminalSessionManager();
