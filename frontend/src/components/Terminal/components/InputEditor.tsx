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

import { useState, useRef, useCallback, useEffect, useMemo } from "react";
import { terminalSessionManager } from "../lib/session-manager";
import { useTerminalStore } from "../store/terminal.store";
import type { TerminalTheme } from "../types/theme";
import { Badge } from "@/components/ui/badge";
import { DiffIcon, FileIcon, GitBranchIcon, TerminalIcon } from "lucide-react";
import { TerminalEnvironmentInfo } from "@wails/types/models";
import { DirectorySelector } from "./DirectorySelector";
import { useEventStore } from "@/store/event.store";
import { EventType } from "@wails/events/models";

interface InputEditorProps {
  sessionId: string;
  theme: TerminalTheme;
  envInfo: TerminalEnvironmentInfo;
  onSubmit: (command: string) => void;
  onResize?: () => void;
}

export function InputEditor({
  sessionId,
  envInfo,
  onSubmit,
  onResize,
}: InputEditorProps) {
  const [value, setValue] = useState("");
  const inputRef = useRef<HTMLTextAreaElement>(null);

  const gitStatus = useEventStore(
    (state) => state.latestEvents[EventType.EventTypeGitStatusChanged],
  );

  // 获取 store 方法
  const navigateHistory = useTerminalStore((state) => state.navigateHistory);

  const adjustInputHeight = useCallback(() => {
    const el = inputRef.current;
    if (!el) return;

    el.style.height = "0px";
    el.style.height = `${el.scrollHeight}px`;
    onResize?.();
  }, [onResize]);

  // 自动聚焦
  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
      adjustInputHeight();
    }
  }, [adjustInputHeight]);

  useEffect(() => {
    adjustInputHeight();
  }, [value, adjustInputHeight]);

  // 聚焦输入框
  const focusInput = useCallback(() => {
    inputRef.current?.focus();
  }, []);

  // 处理输入变化
  const handleChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setValue(e.target.value);
  }, []);

  // 处理键盘事件
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      switch (e.key) {
        case "Enter":
          if (!e.shiftKey) {
            e.preventDefault();
            onSubmit(value);
            setValue("");
          }
          break;

        case "ArrowUp":
          // 历史导航
          if (value === "" || e.ctrlKey) {
            e.preventDefault();
            const prevCmd = navigateHistory(sessionId, "up");
            if (prevCmd !== null) {
              setValue(prevCmd);
            }
          }
          break;

        case "ArrowDown":
          // 历史导航
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
          // Ctrl+C 中断
          if (e.ctrlKey) {
            e.preventDefault();
            terminalSessionManager.write(sessionId, "\x03");
            setValue("");
          }
          break;

        case "l":
          // Ctrl+L 清屏
          if (e.ctrlKey) {
            e.preventDefault();
            terminalSessionManager.write(sessionId, "\x0c");
            setValue("");
          }
          break;

        case "a":
          // Ctrl+A 移动到开头
          if (e.ctrlKey) {
            e.preventDefault();
            const target = e.target as HTMLTextAreaElement;
            target.setSelectionRange(0, 0);
          }
          break;

        case "e":
          // Ctrl+E 移动到结尾
          if (e.ctrlKey) {
            e.preventDefault();
            const target = e.target as HTMLTextAreaElement;
            const len = target.value.length;
            target.setSelectionRange(len, len);
          }
          break;

        case "u":
          // Ctrl+U 删除到开头
          if (e.ctrlKey) {
            e.preventDefault();
            const target = e.target as HTMLTextAreaElement;
            const pos = target.selectionStart;
            const newValue = value.slice(pos);
            setValue(newValue);
          }
          break;

        case "k":
          // Ctrl+K 删除到结尾
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

  // 点击容器时聚焦输入框
  const handleContainerClick = useCallback(() => {
    inputRef.current?.focus();
  }, []);

  // 是否有 Python 环境
  const hasPythonEnv = useMemo(() => {
    return envInfo?.pythonEnv?.hasPython && envInfo?.pythonEnv?.envActive;
  }, [envInfo?.pythonEnv]);

  const isRepo = useMemo(() => {
    return !!(envInfo as { gitInfo?: { isRepo?: boolean } } | undefined)?.gitInfo
      ?.isRepo;
  }, [envInfo]);

  return (
    <div className="input-editor-wrapper flex flex-col items-start px-3 py-2 ">
      <div className="flex items-center gap-1.5 shrink-0">
        {hasPythonEnv && (
          <Badge variant="secondary" className="border text-yellow-200 ">
            <TerminalIcon />
            {envInfo?.pythonEnv?.envName}
          </Badge>
        )}
        <DirectorySelector
          workPath={envInfo?.workPath || "~"}
          onDirectorySelect={onSubmit}
          onFocus={focusInput}
        />
        {isRepo && gitStatus && (
          <Badge
            variant="secondary"
            className="border p-0 gap-0 flex items-center "
          >
            <Badge
              variant="ghost"
              className="text-green-200 hover:bg-accent cursor-pointer"
            >
              <GitBranchIcon className="text-xs" /> {gitStatus.data.status.head}
            </Badge>
            <span className="w-1 h-2 border-l" />
            <Badge variant="ghost" className="hover:bg-accent cursor-pointer">
              {gitStatus.data.status.files.length > 0 ? (
                <>
                  <FileIcon />
                  {gitStatus.data.status.files.length}
                  <span className="text-green-500 font-bold ml-1">
                    +{gitStatus.data.status.addedLines}
                  </span>
                  <span className="text-red-500 font-bold">
                    -{gitStatus.data.status.deletedLines}
                  </span>
                </>
              ) : (
                <>
                  <DiffIcon />0
                </>
              )}
            </Badge>
          </Badge>
        )}
      </div>
      <div
        className="flex items-start flex-1 pt-1 w-full"
        onClick={handleContainerClick}
      >
        <div className="input-field-wrapper relative flex-1">
          <textarea
            ref={inputRef}
            value={value}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            className="actual-input w-full block overflow-hidden bg-transparent outline-none resize-none text-sm leading-6 caret-primary"
            spellCheck={false}
            autoComplete="off"
            autoCorrect="off"
            autoCapitalize="off"
            rows={1}
          />
        </div>
      </div>
    </div>
  );
}
