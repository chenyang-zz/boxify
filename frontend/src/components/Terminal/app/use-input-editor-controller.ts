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

import {
  useState,
  useRef,
  useCallback,
  useEffect,
  useMemo,
  type ChangeEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  type ClipboardEvent as ReactClipboardEvent,
} from "react";
import { terminalSessionManager } from "../lib/session-manager";
import { useTerminalStore } from "../store/terminal.store";
import { useEventStore } from "@/store/event.store";
import { EventType } from "@wails/events/models";
import { TerminalEnvironmentInfo } from "@wails/types/models";
import { classifyCommandTokens } from "../domain";

interface UseInputEditorControllerParams {
  sessionId: string;
  envInfo: TerminalEnvironmentInfo;
  onSubmit: (command: string) => void;
  inFullscreen: boolean;
  onResize?: () => void;
}

// 将 Ctrl+<key> 转换为控制字符（例如 Ctrl+C => \x03）。
function toControlChar(key: string): string | null {
  if (key.length !== 1) return null;
  const upper = key.toUpperCase();
  const code = upper.charCodeAt(0);
  if (code < 65 || code > 90) return null;
  return String.fromCharCode(code - 64);
}

// 将方向键等功能键映射为终端转义序列。
function mapSpecialKeyToSequence(key: string): string | null {
  switch (key) {
    case "Enter":
      return "\r";
    case "Backspace":
      return "\x7f";
    case "Tab":
      return "\t";
    case "Escape":
      return "\x1b";
    case "ArrowUp":
      return "\x1b[A";
    case "ArrowDown":
      return "\x1b[B";
    case "ArrowRight":
      return "\x1b[C";
    case "ArrowLeft":
      return "\x1b[D";
    case "Delete":
      return "\x1b[3~";
    case "Home":
      return "\x1b[H";
    case "End":
      return "\x1b[F";
    case "PageUp":
      return "\x1b[5~";
    case "PageDown":
      return "\x1b[6~";
    default:
      return null;
  }
}

// 全屏交互模式：把按键转换为 PTY 输入并直通后端。
function getInteractiveInput(
  e: Pick<KeyboardEvent, "key" | "ctrlKey" | "altKey" | "metaKey">,
): string | null {
  if (e.metaKey) return null;
  if (e.ctrlKey) {
    return toControlChar(e.key);
  }
  if (e.altKey && e.key.length === 1) {
    return `\x1b${e.key}`;
  }
  return mapSpecialKeyToSequence(e.key) ?? (e.key.length === 1 ? e.key : null);
}

// 输入控制器：封装输入状态、快捷键和命令分词校验。
export function useInputEditorController({
  sessionId,
  envInfo,
  onSubmit,
  inFullscreen,
  onResize,
}: UseInputEditorControllerParams) {
  const [value, setValue] = useState("");
  const inputRef = useRef<HTMLTextAreaElement>(null);

  const gitStatus = useEventStore(
    (state) => state.latestEvents[EventType.EventTypeGitStatusChanged],
  );

  const navigateHistory = useTerminalStore((state) => state.navigateHistory);
  const openReviewPanel = useTerminalStore((state) => state.openReviewPanel);

  // 根据内容自动调整 textarea 高度，保持输入区域自适应。
  const adjustInputHeight = useCallback(() => {
    const el = inputRef.current;
    if (!el) return;

    el.style.height = "0px";
    el.style.height = `${el.scrollHeight}px`;
    onResize?.();
  }, [onResize]);

  // 自动聚焦。
  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
      adjustInputHeight();
    }
  }, [adjustInputHeight]);

  useEffect(() => {
    adjustInputHeight();
  }, [value, adjustInputHeight]);

  // 聚焦输入框，供目录选择器等外部交互回调使用。
  const focusInput = useCallback(() => {
    inputRef.current?.focus();
  }, []);

  // 同步用户输入到本地状态。
  const handleChange = useCallback(
    (e: ChangeEvent<HTMLTextAreaElement>) => {
      if (inFullscreen) return;
      setValue(e.target.value);
    },
    [inFullscreen],
  );

  // 处理终端常用快捷键与提交行为。
  const handleKeyDown = useCallback(
    (e: ReactKeyboardEvent<HTMLTextAreaElement>) => {
      if (inFullscreen) {
        const data = getInteractiveInput({
          key: e.key,
          ctrlKey: e.ctrlKey,
          altKey: e.altKey,
          metaKey: e.metaKey,
        });
        if (data) {
          e.preventDefault();
          terminalSessionManager.write(sessionId, data);
        } else if (!e.metaKey) {
          e.preventDefault();
        }
        return;
      }

      switch (e.key) {
        case "Enter":
          if (!e.shiftKey) {
            e.preventDefault();
            onSubmit(value);
            setValue("");
          }
          break;

        case "ArrowUp":
          if (value === "" || e.ctrlKey) {
            e.preventDefault();
            const prevCmd = navigateHistory(sessionId, "up");
            if (prevCmd !== null) {
              setValue(prevCmd);
            }
          }
          break;

        case "ArrowDown":
          if (value === "" || e.ctrlKey) {
            e.preventDefault();
            const nextCmd = navigateHistory(sessionId, "down");
            setValue(nextCmd || "");
          }
          break;

        case "Tab":
          e.preventDefault();
          // TODO: 触发自动补全
          break;

        case "c":
          if (e.ctrlKey) {
            e.preventDefault();
            terminalSessionManager.write(sessionId, "\x03");
            setValue("");
          }
          break;

        case "l":
          if (e.ctrlKey) {
            e.preventDefault();
            terminalSessionManager.write(sessionId, "\x0c");
            setValue("");
          }
          break;

        case "a":
          if (e.ctrlKey) {
            e.preventDefault();
            const target = e.target as HTMLTextAreaElement;
            target.setSelectionRange(0, 0);
          }
          break;

        case "e":
          if (e.ctrlKey) {
            e.preventDefault();
            const target = e.target as HTMLTextAreaElement;
            const len = target.value.length;
            target.setSelectionRange(len, len);
          }
          break;

        case "u":
          if (e.ctrlKey) {
            e.preventDefault();
            const target = e.target as HTMLTextAreaElement;
            const pos = target.selectionStart;
            const newValue = value.slice(pos);
            setValue(newValue);
          }
          break;

        case "k":
          if (e.ctrlKey) {
            e.preventDefault();
            const target = e.target as HTMLTextAreaElement;
            const pos = target.selectionStart;
            const newValue = value.slice(0, pos);
            setValue(newValue);
          }
          break;
      }
    },
    [sessionId, value, navigateHistory, onSubmit, inFullscreen],
  );

  // 全屏交互模式下支持粘贴直通终端。
  const handlePaste = useCallback(
    (e: ReactClipboardEvent<HTMLTextAreaElement>) => {
      if (!inFullscreen) return;
      e.preventDefault();
      const text = e.clipboardData.getData("text");
      if (!text) return;
      terminalSessionManager.write(sessionId, text);
    },
    [inFullscreen, sessionId],
  );

  // 点击输入区域空白时聚焦 textarea。
  const handleContainerClick = useCallback(() => {
    inputRef.current?.focus();
  }, []);

  // 打开 Git Review 面板。
  const handleOpenReviewPanel = useCallback(() => {
    openReviewPanel(sessionId);
  }, [openReviewPanel, sessionId]);

  // 计算是否激活 Python 环境标签。
  const hasPythonEnv = useMemo(() => {
    return envInfo?.pythonEnv?.hasPython && envInfo?.pythonEnv?.envActive;
  }, [envInfo?.pythonEnv]);

  // 计算当前目录是否为 Git 仓库，用于展示 Git 面板入口。
  const isRepo = useMemo(() => {
    return !!(envInfo as { gitInfo?: { isRepo?: boolean } } | undefined)
      ?.gitInfo?.isRepo;
  }, [envInfo]);

  // 从 session 命令缓存构建命令集合，用于命令有效性校验。
  const commandSet = useMemo(() => {
    const cache = terminalSessionManager.getExecutableCommandCache(sessionId);
    const commandNames = cache?.commands
      ?.map((item) => item?.name?.trim())
      .filter((name): name is string => Boolean(name));
    const defaultCommands = cache?.defaultCommands
      ?.map((name) => name?.trim())
      .filter((name): name is string => Boolean(name));
    return new Set([...(commandNames ?? []), ...(defaultCommands ?? [])]);
  }, [sessionId, envInfo]);

  // 对当前输入进行分词与分类，驱动彩色渲染层。
  const highlightedTokens = useMemo(
    () => (inFullscreen ? [] : classifyCommandTokens(value, commandSet)),
    [value, commandSet, inFullscreen],
  );

  return {
    value,
    inputRef,
    gitStatus,
    hasPythonEnv,
    isRepo,
    highlightedTokens,
    focusInput,
    handleChange,
    handleKeyDown,
    handlePaste,
    handleContainerClick,
    handleOpenReviewPanel,
  };
}
