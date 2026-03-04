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
  type KeyboardEvent,
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
  onResize?: () => void;
}

// 输入控制器：封装输入状态、快捷键和命令分词校验。
export function useInputEditorController({
  sessionId,
  envInfo,
  onSubmit,
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
      setValue(e.target.value);
    },
    [],
  );

  // 处理终端常用快捷键与提交行为。
  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLTextAreaElement>) => {
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
    [sessionId, value, navigateHistory, onSubmit],
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
    () => classifyCommandTokens(value, commandSet),
    [value, commandSet],
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
    handleContainerClick,
    handleOpenReviewPanel,
  };
}
