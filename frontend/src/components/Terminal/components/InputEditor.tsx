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

type TokenType =
  | "whitespace"
  | "command"
  | "option"
  | "operator"
  | "variable"
  | "string"
  | "path"
  | "argument";

interface CommandToken {
  value: string;
  type: TokenType;
  valid?: boolean;
}

const OPERATOR_SET = new Set([
  "|",
  "||",
  "&&",
  ";",
  ">",
  ">>",
  "<",
  "<<",
  "2>",
  "2>>",
  "&",
]);

const ASSIGNMENT_RE = /^[A-Za-z_][A-Za-z0-9_]*=.*/;

// 将输入拆分为保留空白和操作符的 token 列表，供后续高亮使用。
function splitCommandTokens(input: string): string[] {
  const tokens: string[] = [];
  let buffer = "";
  let quote: "'" | '"' | null = null;

  for (let i = 0; i < input.length; i += 1) {
    const ch = input[i];

    if (quote) {
      buffer += ch;
      if (ch === quote) {
        quote = null;
      }
      continue;
    }

    if (ch === "'" || ch === '"') {
      buffer += ch;
      quote = ch;
      continue;
    }

    if (/\s/.test(ch)) {
      if (buffer) {
        tokens.push(buffer);
        buffer = "";
      }
      let spaceRun = ch;
      while (i + 1 < input.length && /\s/.test(input[i + 1])) {
        i += 1;
        spaceRun += input[i];
      }
      tokens.push(spaceRun);
      continue;
    }

    if (ch === "&" || ch === "|" || ch === ";" || ch === "<" || ch === ">") {
      if (buffer) {
        tokens.push(buffer);
        buffer = "";
      }
      let operator = ch;
      const next = input[i + 1];
      if (
        (ch === "&" || ch === "|" || ch === ">" || ch === "<") &&
        next === ch
      ) {
        i += 1;
        operator += next;
      }
      tokens.push(operator);
      continue;
    }

    if (ch === "2" && (input[i + 1] === ">" || input[i + 1] === "<")) {
      if (buffer) {
        tokens.push(buffer);
        buffer = "";
      }
      const operator = `${ch}${input[i + 1]}`;
      i += 1;
      if (input[i + 1] === ">") {
        i += 1;
        tokens.push(`${operator}>`);
      } else {
        tokens.push(operator);
      }
      continue;
    }

    buffer += ch;
  }

  if (buffer) {
    tokens.push(buffer);
  }

  return tokens;
}

// 规范化命令 token（去掉首尾引号），用于命令有效性匹配。
function normalizeCommandToken(token: string): string {
  return token.replace(/^["']|["']$/g, "");
}

// 对 token 进行语义分类，并对主命令做有效性标记。
function classifyTokens(
  input: string,
  commandSet: Set<string>,
): CommandToken[] {
  const rawTokens = splitCommandTokens(input);
  const commandIndex = rawTokens.findIndex((token) => {
    if (!token || /^\s+$/.test(token)) return false;
    if (ASSIGNMENT_RE.test(token)) return false;
    return !OPERATOR_SET.has(token);
  });

  return rawTokens.map((token, index) => {
    if (!token) return { value: token, type: "argument" };
    if (/^\s+$/.test(token)) return { value: token, type: "whitespace" };
    if (OPERATOR_SET.has(token)) return { value: token, type: "operator" };

    if (index === commandIndex) {
      const normalized = normalizeCommandToken(token);
      const valid = normalized.length > 0 && commandSet.has(normalized);
      return { value: token, type: "command", valid };
    }

    if (token.startsWith("-")) return { value: token, type: "option" };
    if (token.startsWith("$")) return { value: token, type: "variable" };
    if (
      (token.startsWith('"') && token.endsWith('"')) ||
      (token.startsWith("'") && token.endsWith("'"))
    ) {
      return { value: token, type: "string" };
    }
    if (
      token.includes("/") ||
      token.startsWith("./") ||
      token.startsWith("../") ||
      token.startsWith("~/")
    ) {
      return { value: token, type: "path" };
    }

    return { value: token, type: "argument" };
  });
}

// 返回不同 token 对应的展示样式类名。
function tokenClassName(token: CommandToken): string {
  switch (token.type) {
    case "command":
      if (token.valid) return "text-emerald-400";
      return "text-foreground underline decoration-red-400 decoration-dashed underline-offset-4";
    case "option":
      return "text-amber-300";
    case "operator":
      return "text-pink-300";
    case "variable":
      return "text-sky-300";
    case "string":
      return "text-teal-300";
    case "path":
      return "text-violet-300";
    default:
      return "text-foreground";
  }
}

interface InputEditorProps {
  sessionId: string;
  theme: TerminalTheme;
  envInfo: TerminalEnvironmentInfo;
  onSubmit: (command: string) => void;
  onResize?: () => void;
}

// 终端输入组件：负责输入、快捷键、分词高亮与命令有效性提示。
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
  const openReviewPanel = useTerminalStore((state) => state.openReviewPanel);

  // 根据内容自动调整 textarea 高度，保持输入区域自适应。
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
  // 供外部目录选择器回调使用，确保点选后仍能回到输入框。
  const focusInput = useCallback(() => {
    inputRef.current?.focus();
  }, []);

  // 同步用户输入到本地状态。
  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      setValue(e.target.value);
    },
    [],
  );

  // 处理终端常用快捷键与提交行为。
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

  // 点击输入区域空白时聚焦 textarea。
  const handleContainerClick = useCallback(() => {
    inputRef.current?.focus();
  }, []);

  // 计算是否激活 Python 环境标签。
  const hasPythonEnv = useMemo(() => {
    return envInfo?.pythonEnv?.hasPython && envInfo?.pythonEnv?.envActive;
  }, [envInfo?.pythonEnv]);

  // 计算当前目录是否为 Git 仓库，用于展示 Git 面板入口。
  const isRepo = useMemo(() => {
    return !!(envInfo as { gitInfo?: { isRepo?: boolean } } | undefined)
      ?.gitInfo?.isRepo;
  }, [envInfo]);

  // 打开 Git Review 面板。
  const handleOpenReviewPanel = useCallback(() => {
    openReviewPanel(sessionId);
  }, [openReviewPanel, sessionId]);

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
    () => classifyTokens(value, commandSet),
    [value, commandSet],
  );

  return (
    <div className="input-editor-wrapper flex flex-col items-start px-3 py-2 ">
      <div className="flex items-center gap-1.5 shrink-0 flex-wrap">
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
            className="border p-0 gap-0 flex items-center cursor-pointer select-none"
          >
            <Badge
              variant="ghost"
              className="text-green-200 hover:bg-accent cursor-pointer select-none"
            >
              <GitBranchIcon className="text-xs" /> {gitStatus.data.status.head}
            </Badge>
            <span className="w-1 h-2 border-l" />
            <Badge
              variant="ghost"
              className="hover:bg-accent cursor-pointer select-none"
              onClick={handleOpenReviewPanel}
            >
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
          <pre
            className="pointer-events-none absolute inset-0 m-0 whitespace-pre-wrap wrap-break-word text-sm leading-6 font-sans"
            aria-hidden="true"
          >
            {highlightedTokens.map((token, index) => (
              <span
                key={`${index}-${token.value}`}
                className={tokenClassName(token)}
              >
                {token.value}
              </span>
            ))}
          </pre>
          <textarea
            ref={inputRef}
            value={value}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            className="actual-input relative w-full block overflow-hidden bg-transparent outline-none resize-none text-sm leading-6 text-transparent caret-primary"
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
