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
import { ShellType, TerminalConfig as GoTerminalConfig } from "@wails/terminal";
import type { TerminalConfig } from "@/types/property";
import { AnsiParser } from "./ansi-parser";
import { callWails } from "@/lib/utils";
import {
  TerminalEnvironmentInfo,
  TerminalInteractionModeChangedEvent,
  TerminalListExecutableCommandsData,
} from "@wails/types/models";
import { useEventStore } from "@/store/event.store";
import { EventType } from "@wails/events/models";
import type { OutputLine } from "../types/block";

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
  inInteractive?: boolean;
  executableCommands?: TerminalListExecutableCommandsData;
  onEnvChange?: (env: SessionEnvironmentInfo) => void;
  onEvent?: (event: TerminalSessionEvent) => void;
  currentGitRepoKey?: string;
  currentBlockId?: string;
  interactiveModeBlockId?: string;
  // 记录“本次命令曾进入交互态”的 block，命令结束后强制收敛为 "%"。
  interactiveBlockIds: Set<string>;
}

interface OutputChunk {
  content: string;
  formattedContent: OutputLine["formattedContent"];
}

export type TerminalSessionEvent =
  | {
      type: "output_batch";
      sessionId: string;
      blockId: string;
      chunks: OutputChunk[];
    }
  | {
      type: "error";
      sessionId: string;
      blockId?: string;
      message: string;
      content: string;
      formattedContent: OutputLine["formattedContent"];
    }
  | {
      type: "command_end";
      sessionId: string;
      blockId: string;
      exitCode: number;
    }
  | {
      type: "interactive_placeholder";
      sessionId: string;
      blockId: string;
      chunk: OutputChunk;
    }
  | {
      type: "session_destroyed";
      sessionId: string;
    }
  | {
      type: "interaction_mode_changed";
      sessionId: string;
      inInteractive: boolean;
      changedAtUnix: number;
    };

class TerminalSessionManager {
  private sessions = new Map<string, CachedSession>();
  private outputQueues = new Map<
    string,
    Array<{
      content: string;
      formattedContent: ReturnType<AnsiParser["parse"]>;
    }>
  >();
  private flushTimerId: number | null = null;
  private defaultRows = 24;
  private defaultCols = 80;

  getOrCreate(sessionId: string): CachedSession {
    let session = this.sessions.get(sessionId);

    if (!session) {
      session = {
        isInitialized: false,
        parser: new AnsiParser(),
        unbindCallbacks: [],
        interactiveBlockIds: new Set<string>(),
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

  // 设置终端事件回调，由应用层决定如何落地到状态管理。
  setEventCallback(
    sessionId: string,
    callback?: (event: TerminalSessionEvent) => void,
  ): void {
    const session = this.sessions.get(sessionId);
    if (session) {
      session.onEvent = callback;
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

    // 交互模式切换事件
    const unbindInteractionModeChange = Events.On(
      "terminal:interaction_mode_change",
      (event: { data: TerminalInteractionModeChangedEvent }) => {
        if (event.data.sessionId === sessionId) {
          console.log("[terminal:interaction_mode_change]:", event.data);
          this.handleInteractionModeChange(sessionId, event.data);
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
      unbindInteractionModeChange,
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

      // 仅消费带 blockId 的输出，避免初始化阶段输出误挂到用户命令 block。
      const targetBlockId = blockId ?? null;

      // 交互模式期间产生的 block 输出在结束后只保留一个 "%" 占位，避免回灌大段交互内容。
      if (targetBlockId && session.inInteractive) {
        session.interactiveBlockIds.add(targetBlockId);
        session.interactiveModeBlockId = targetBlockId;
        return;
      }

      // 交互模式结束后，到命令结束前仍可能有尾部输出，继续丢弃该 block 的原始内容。
      if (
        targetBlockId &&
        (session.interactiveModeBlockId === targetBlockId ||
          session.interactiveBlockIds.has(targetBlockId))
      ) {
        return;
      }

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

    if (this.flushTimerId !== null) return;

    // 使用定时器而非 requestAnimationFrame，避免 WebView 节流导致输出延迟到命令结束。
    this.flushTimerId = window.setTimeout(() => {
      this.flushOutputQueue();
    }, 16);
  }

  private emitEvent(sessionId: string, event: TerminalSessionEvent): void {
    const session = this.sessions.get(sessionId);
    if (!session?.onEvent) return;
    session.onEvent(event);
  }

  private flushOutputQueue(): void {
    this.flushTimerId = null;
    if (this.outputQueues.size === 0) return;

    this.outputQueues.forEach((chunks, key) => {
      if (chunks.length === 0) return;
      const separator = key.indexOf(":");
      if (separator === -1) return;
      const sessionId = key.slice(0, separator);
      const blockId = key.slice(separator + 1);
      this.emitEvent(sessionId, {
        type: "output_batch",
        sessionId,
        blockId,
        chunks,
      });
    });

    this.outputQueues.clear();
  }

  private handleError(sessionId: string, message: string): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    const currentBlockId = session.currentBlockId;

    if (currentBlockId) {
      const formattedContent = session.parser.parse(
        `\x1b[31m错误: ${message}\x1b[0m`,
      );
      this.emitEvent(sessionId, {
        type: "error",
        sessionId,
        blockId: currentBlockId,
        message,
        content: `\n错误: ${message}`,
        formattedContent,
      });
    }
  }

  private handleCommandEnd(
    sessionId: string,
    blockId: string,
    exitCode: number,
  ): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    // 交互命令结束后统一覆盖输出，保证 block 只展示一个 "%".
    if (session.interactiveBlockIds.has(blockId)) {
      this.emitEvent(sessionId, {
        type: "interactive_placeholder",
        sessionId,
        blockId,
        chunk: {
          content: "%",
          formattedContent: session.parser.parse("%"),
        },
      });
      session.interactiveBlockIds.delete(blockId);
      if (session.interactiveModeBlockId === blockId) {
        session.interactiveModeBlockId = undefined;
      }
    }

    if (session?.currentBlockId === blockId) {
      session.currentBlockId = undefined;
    }
    this.emitEvent(sessionId, {
      type: "command_end",
      sessionId,
      blockId,
      exitCode,
    });
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

  // 处理交互模式切换，通知应用层切换输入策略。
  private handleInteractionModeChange(
    sessionId: string,
    event: TerminalInteractionModeChangedEvent,
  ): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;
    session.inInteractive = event.inInteractive;
    if (event.inInteractive && session.currentBlockId) {
      session.interactiveModeBlockId = session.currentBlockId;
      session.interactiveBlockIds.add(session.currentBlockId);
    }
    this.emitEvent(sessionId, {
      type: "interaction_mode_changed",
      sessionId,
      inInteractive: event.inInteractive,
      changedAtUnix: event.changedAtUnix,
    });
  }

  private clearGitStatusEvent() {
    useEventStore.getState().clearEvent(EventType.EventTypeGitStatusChanged);
  }

  private async syncExecutableCommands(
    sessionId: string,
    shellType: ShellType,
  ): Promise<void> {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    try {
      const res = await callWails(
        TerminalService.ListExecutableCommands,
        shellType,
      );
      const data = res.data;
      if (!data) return;
      session.executableCommands = data;
      // 命令缓存更新后触发一次回调，确保输入区校验状态及时刷新
      if (session.environmentInfo && session.onEnvChange) {
        session.onEnvChange({ ...session.environmentInfo });
      }
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
      session.environmentInfo =
        (res.data as SessionEnvironmentInfo) ?? undefined;
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
  async writeCommand(
    sessionId: string,
    command: string,
    blockId?: string,
  ): Promise<string> {
    const session = this.sessions.get(sessionId);
    if (!session || !session.isInitialized) return "";

    try {
      let resolvedBlockId = "";

      if (
        blockId &&
        typeof TerminalService.WriteCommandWithBlock === "function"
      ) {
        resolvedBlockId = await TerminalService.WriteCommandWithBlock(
          sessionId,
          blockId,
          command,
        );
      } else {
        resolvedBlockId = await TerminalService.WriteCommand(
          sessionId,
          command,
        );
      }

      session.currentBlockId = resolvedBlockId || undefined;
      if (resolvedBlockId) {
        session.interactiveModeBlockId = undefined;
        session.interactiveBlockIds.delete(resolvedBlockId);
      }
      return resolvedBlockId || "";
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
      GitService.StopRepoWatch(session.currentGitRepoKey).catch(
        () => undefined,
      );
    }

    TerminalService.Close(sessionId).catch((err) => {
      console.error("关闭终端失败:", err);
    });

    this.emitEvent(sessionId, {
      type: "session_destroyed",
      sessionId,
    });

    for (const key of this.outputQueues.keys()) {
      if (key.startsWith(`${sessionId}:`)) {
        this.outputQueues.delete(key);
      }
    }
    if (this.flushTimerId !== null && this.outputQueues.size === 0) {
      clearTimeout(this.flushTimerId);
      this.flushTimerId = null;
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

  // 查询当前会话是否处于交互模式。
  isInInteractive(sessionId: string): boolean {
    return this.sessions.get(sessionId)?.inInteractive ?? false;
  }

  getExecutableCommandCache(
    sessionId: string,
  ): TerminalListExecutableCommandsData | undefined {
    return this.sessions.get(sessionId)?.executableCommands;
  }
}

export const terminalSessionManager = new TerminalSessionManager();
