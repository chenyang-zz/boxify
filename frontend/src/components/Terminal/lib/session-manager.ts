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
import { GitService } from "@wails/service";
import {
  ShellType,
  TerminalConfig as GoTerminalConfig,
} from "@wails/terminal";
import type { TerminalConfig } from "@/types/property";
import { AnsiParser } from "./ansi-parser";
import { useTerminalStore } from "../store/terminal.store";
import { defaultTheme } from "../types/theme";
import { callWails } from "@/lib/utils";
import {
  TerminalEnvironmentInfo,
  TerminalListExecutableCommandsData,
} from "@wails/types/models";
import { useEventStore } from "@/store/event.store";
import { EventType } from "@wails/events/models";

interface SessionGitInfo {
  isRepo: boolean;
  branch?: string;
  modifiedFiles: number;
  addedLines: number;
  deletedLines: number;
}

type SessionEnvironmentInfo = TerminalEnvironmentInfo & {
  gitInfo?: SessionGitInfo;
};

interface CachedSession {
  isInitialized: boolean;
  parser: AnsiParser;
  unbindCallbacks: (() => void)[];
  environmentInfo?: SessionEnvironmentInfo;
  executableCommands?: TerminalListExecutableCommandsData;
  onEnvChange?: (env: SessionEnvironmentInfo) => void;
  currentGitRepoKey?: string;
}

class TerminalSessionManager {
  private sessions = new Map<string, CachedSession>();
  private outputQueues = new Map<
    string,
    Array<{ content: string; formattedContent: ReturnType<AnsiParser["parse"]> }>
  >();
  private flushFrameId: number | null = null;
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
    callback?: (env: SessionEnvironmentInfo) => void,
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
          console.log("[terminal:error]:", event.data);
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
          console.log("[terminal:command_end]:", event.data);
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
          console.log("[terminal:pwd_update]:", event.data);
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
          console.log("[terminal:git_update]:", event.data);
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

      const store = useTerminalStore.getState();

      // 优先使用事件中的 blockId，否则回退到 store 中的当前 blockId
      const targetBlockId = blockId ?? store.currentBlockIds[sessionId] ?? null;

      // 只在有 block 时追加输出
      if (targetBlockId) {
        const formattedContent = session.parser.parse(decoded);
        this.enqueueOutput(sessionId, targetBlockId, decoded, formattedContent);
      }
    } catch (e) {
      console.error("处理终端输出失败:", e);
    }
  }

  private enqueueOutput(
    sessionId: string,
    blockId: string,
    content: string,
    formattedContent: ReturnType<AnsiParser["parse"]>,
  ): void {
    const queueKey = `${sessionId}:${blockId}`;
    const queue = this.outputQueues.get(queueKey) ?? [];
    queue.push({ content, formattedContent });
    this.outputQueues.set(queueKey, queue);

    if (this.flushFrameId !== null) return;

    this.flushFrameId = requestAnimationFrame(() => {
      this.flushOutputQueue();
    });
  }

  private flushOutputQueue(): void {
    this.flushFrameId = null;
    if (this.outputQueues.size === 0) return;

    const store = useTerminalStore.getState();

    this.outputQueues.forEach((chunks, key) => {
      if (chunks.length === 0) return;
      const separator = key.indexOf(":");
      if (separator === -1) return;
      const sessionId = key.slice(0, separator);
      const blockId = key.slice(separator + 1);
      store.appendOutputBatch(sessionId, blockId, chunks);
    });

    this.outputQueues.clear();
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
      store.appendOutputBatch(
        sessionId,
        currentBlockId,
        [{ content: `\n错误: ${message}`, formattedContent }],
      );
      store.finalizeBlock(sessionId, currentBlockId, 1);
    }
  }

  private handleCommandEnd(
    sessionId: string,
    blockId: string,
    exitCode: number,
  ): void {
    const store = useTerminalStore.getState();
    store.finalizeBlock(sessionId, blockId, exitCode);
  }

  // 通用的环境信息更新方法
  private updateEnvironmentInfo(
    sessionId: string,
    updater: (env: SessionEnvironmentInfo) => SessionEnvironmentInfo,
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

    this.syncGitWatchByWorkPath(sessionId, pwd);
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

  private clearGitStatusEvent() {
    useEventStore
      .getState()
      .clearEvent(EventType.EventTypeGitStatusChanged);
  }

  private async syncExecutableCommands(
    sessionId: string,
    shellType: ShellType,
  ): Promise<void> {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    try {
      const res = await callWails(TerminalService.ListExecutableCommands, shellType);
      const data = res.data;
      if (!data) return;
      session.executableCommands = data;
    } catch (err) {
      console.error("获取可执行命令失败:", err);
    }
  }

  private async syncGitWatchByWorkPath(
    sessionId: string,
    workPath: string,
  ): Promise<void> {
    const session = this.sessions.get(sessionId);
    if (!session || !workPath) return;

    try {
      // repoKey 传空，让后端按仓库根目录归一化 key，避免子目录重复注册。
      const registerRes = await GitService.RegisterRepo("", workPath);
      if (!registerRes?.success || !registerRes.data?.repoKey) {
        throw new Error(registerRes?.message || "仓库注册失败");
      }

      const nextRepoKey = registerRes.data.repoKey;

      // 切换仓库时先停止旧监听。
      if (
        session.currentGitRepoKey &&
        session.currentGitRepoKey !== nextRepoKey
      ) {
        await GitService.StopRepoWatch(session.currentGitRepoKey);
      }

      session.currentGitRepoKey = nextRepoKey;

      // 激活当前仓库并仅保留它的监听。
      await GitService.SetActiveRepo(nextRepoKey, true, true);

      // 立即拉取一次状态并写入事件缓存，确保前端展示及时更新。
      const initialRes = await GitService.GetInitialStatusEvent(nextRepoKey);
      if (initialRes?.success && initialRes.data?.event) {
        useEventStore
          .getState()
          .setEvent(EventType.EventTypeGitStatusChanged, initialRes.data.event);
      }

      this.updateEnvironmentInfo(sessionId, (env) => ({
        ...env,
        gitInfo: {
          isRepo: true,
          branch: env.gitInfo?.branch ?? "",
          modifiedFiles: env.gitInfo?.modifiedFiles ?? 0,
          addedLines: env.gitInfo?.addedLines ?? 0,
          deletedLines: env.gitInfo?.deletedLines ?? 0,
        },
      }));
    } catch {
      // 当前目录不在 git 仓库时，关闭旧监听并清理前端 git 展示。
      if (session.currentGitRepoKey) {
        await GitService.StopRepoWatch(session.currentGitRepoKey).catch(
          () => undefined,
        );
      }
      session.currentGitRepoKey = undefined;
      this.clearGitStatusEvent();

      this.updateEnvironmentInfo(sessionId, (env) => ({
        ...env,
        gitInfo: {
          isRepo: false,
          branch: "",
          modifiedFiles: 0,
          addedLines: 0,
          deletedLines: 0,
        },
      }));
    }
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
      session.environmentInfo = (res.data as SessionEnvironmentInfo) ?? undefined;
      void this.syncExecutableCommands(sessionId, terminalConfig.shell);
      if (terminalConfig.workpath) {
        void this.syncGitWatchByWorkPath(sessionId, terminalConfig.workpath);
      }
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

    if (session.currentGitRepoKey) {
      GitService.StopRepoWatch(session.currentGitRepoKey).catch(() => undefined);
    }

    TerminalService.Close(sessionId).catch((err) => {
      console.error("关闭终端失败:", err);
    });

    useTerminalStore.getState().clearSession(sessionId);

    for (const key of this.outputQueues.keys()) {
      if (key.startsWith(`${sessionId}:`)) {
        this.outputQueues.delete(key);
      }
    }
    if (this.flushFrameId !== null && this.outputQueues.size === 0) {
      cancelAnimationFrame(this.flushFrameId);
      this.flushFrameId = null;
    }

    this.sessions.delete(sessionId);
  }

  has(sessionId: string): boolean {
    return this.sessions.has(sessionId);
  }

  isInitialized(sessionId: string): boolean {
    const session = this.sessions.get(sessionId);
    return session?.isInitialized ?? false;
  }

  getExecutableCommandCache(
    sessionId: string,
  ): TerminalListExecutableCommandsData | undefined {
    return this.sessions.get(sessionId)?.executableCommands;
  }
}

export const terminalSessionManager = new TerminalSessionManager();
