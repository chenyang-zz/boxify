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

import { Badge } from "@/components/ui/badge";
import { DiffIcon, FileIcon, GitBranchIcon, TerminalIcon } from "lucide-react";
import { TerminalEnvironmentInfo } from "@wails/types/models";
import { DirectorySelector } from "./DirectorySelector";
import { useInputEditorController } from "../app";
import { commandTokenClassName } from "../domain";

interface InputEditorProps {
  sessionId: string;
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
  const {
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
  } = useInputEditorController({
    sessionId,
    envInfo,
    onSubmit,
    onResize,
  });

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
                className={commandTokenClassName(token)}
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
