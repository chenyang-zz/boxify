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
  FC,
  forwardRef,
  ReactNode,
  useCallback,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from "react";
import {
  ArrowUp,
  FileText,
  Loader2,
  Paperclip,
  Pause,
  X,
} from "lucide-react";
import { cn, formatFileSize } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

/**
 * 附件文件描述
 *
 * 当前仅用于 UI 交互展示，不直接依赖后端 API。
 */
export interface ChatInputFile {
  /** 本地唯一标识 */
  id: string;
  /** 文件名 */
  name: string;
  /** 文件大小（字节） */
  size: number;
  /** MIME 类型 */
  type: string;
  /** 原始 File 对象（可选，便于调用方二次上传） */
  raw?: File;
}

export interface ChatInputRef {
  /** 设置输入框文本并聚焦 */
  setInputText: (text: string) => void;
  /** 获取当前输入值 */
  getInputValue: () => string;
  /** 获取当前已选文件 */
  getFiles: () => ChatInputFile[];
  /** 清空输入框与文件 */
  clear: () => void;
}

interface ChatInputProps {
  className?: string;
  placeholder?: string;
  disabled?: boolean;
  /** 发送消息时的加载状态 */
  loading?: boolean;
  /** 任务是否正在运行中，显示暂停按钮 */
  isRunning?: boolean;
  /** 发送回调，成功后会自动清空输入 */
  onSend?: (message: string, files: ChatInputFile[]) => void | Promise<void>;
  /** 暂停回调 */
  onStop?: () => void;
  /** 输入值变化回调 */
  onInputChange?: (value: string) => void;
  /** 文件被选择后的回调（UI 已将其加入列表） */
  onAttachFiles?: (files: ChatInputFile[]) => void;
}

const genFileId = () =>
  `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 9)}`;

const fileFromRaw = (raw: File): ChatInputFile => ({
  id: genFileId(),
  name: raw.name,
  size: raw.size,
  type: raw.type,
  raw,
});

/** 工具栏图标按钮 — 无背景 hover，避免 ghost variant 优先级问题 */
const ToolbarIcon: FC<{
  children: React.ReactNode;
  title?: string;
  onClick?: () => void;
  disabled?: boolean;
  className?: string;
}> = ({ children, title, onClick, disabled, className }) => (
  <button
    type="button"
    title={title}
    onClick={onClick}
    disabled={disabled}
    className={cn(
      "inline-flex size-8 items-center justify-center rounded-full text-muted-foreground transition",
      "hover:bg-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-40",
      className,
    )}
  >
    {children}
  </button>
);

export const ChatInput = forwardRef<ChatInputRef, ChatInputProps>(
  (
    {
      className,
      placeholder = "分配一个任务或提问任何问题…",
      disabled = false,
      loading = false,
      isRunning = false,
      onSend,
      onStop,
      onInputChange,
      onAttachFiles,
    },
    ref,
  ) => {
    const [inputValue, setInputValue] = useState("");
    const [files, setFiles] = useState<ChatInputFile[]>([]);
    const [sending, setSending] = useState(false);

    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const fileInputRef = useRef<HTMLInputElement>(null);

    const isBusy = disabled || loading || sending;

    const resetInput = useCallback(() => {
      setInputValue("");
      setFiles([]);
      onInputChange?.("");
      if (textareaRef.current) {
        textareaRef.current.style.height = "auto";
      }
    }, [onInputChange]);

    const adjustHeight = useCallback(() => {
      const el = textareaRef.current;
      if (!el) return;
      el.style.height = "auto";
      const nextHeight = Math.min(Math.max(el.scrollHeight, 48), 240);
      el.style.height = `${nextHeight}px`;
    }, []);

    useEffect(() => {
      adjustHeight();
    }, [inputValue, adjustHeight]);

    useImperativeHandle(
      ref,
      () => ({
        setInputText: (text) => {
          setInputValue(text);
          onInputChange?.(text);
          textareaRef.current?.focus();
        },
        getInputValue: () => inputValue,
        getFiles: () => files,
        clear: resetInput,
      }),
      [files, inputValue, onInputChange, resetInput],
    );

    const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      const value = e.target.value;
      setInputValue(value);
      onInputChange?.(value);
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === "Enter" && (e.ctrlKey || e.metaKey) && !isBusy) {
        e.preventDefault();
        handleSend();
      }
    };

    const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
      const selected = e.target.files;
      if (!selected || selected.length === 0) return;

      const added = Array.from(selected).map(fileFromRaw);
      setFiles((prev) => [...prev, ...added]);
      onAttachFiles?.(added);

      if (fileInputRef.current) fileInputRef.current.value = "";
    };

    const handleRemoveFile = (id: string) => {
      setFiles((prev) => prev.filter((f) => f.id !== id));
    };

    const handleSend = async () => {
      const trimmed = inputValue.trim();
      if (!trimmed) {
        textareaRef.current?.focus();
        return;
      }
      if (!onSend || isBusy) return;

      setSending(true);
      try {
        await onSend(trimmed, files);
        resetInput();
      } catch {
        // 调用方处理业务错误；组件内仅保证 sending 状态正确重置
      } finally {
        setSending(false);
      }
    };

    const canSend = !isBusy && inputValue.trim().length > 0;

    return (
      <div
        className={cn(
          "relative flex w-full flex-col overflow-hidden rounded-[20px] border border-input bg-card shadow-sm transition",
          className,
        )}
      >
        {/* 附件列表 */}
        {files.length > 0 && (
          <div className="w-full px-4 pt-4">
            <div className="flex w-full gap-2 overflow-x-auto scrollbar-hide">
              {files.map((file) => (
                <Badge
                  key={file.id}
                  variant="secondary"
                  className="inline-flex shrink-0 items-center gap-1.5 rounded-full border border-border px-2.5 py-1 text-xs font-normal text-secondary-foreground"
                >
                  <FileText className="size-3.5 text-muted-foreground" />
                  <span className="max-w-[160px] truncate">{file.name}</span>
                  <span className="text-muted-foreground">
                    · {formatFileSize(file.size)}
                  </span>
                  <button
                    type="button"
                    onClick={() => handleRemoveFile(file.id)}
                    disabled={isBusy}
                    className="ml-0.5 inline-flex items-center justify-center rounded-full p-0.5 text-muted-foreground transition hover:bg-accent hover:text-accent-foreground disabled:opacity-50"
                  >
                    <X className="size-3" />
                  </button>
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* 输入区 */}
        <div className="w-full px-4 pt-4">
          <textarea
            ref={textareaRef}
            rows={1}
            value={inputValue}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={isBusy}
            className={cn(
              "min-h-[44px] w-full resize-none bg-transparent pb-2.5 text-[14px] text-foreground outline-none placeholder:text-muted-foreground",
              "scrollbar-hide disabled:cursor-not-allowed disabled:opacity-60",
            )}
          />
        </div>

        {/* 工具栏 */}
        <div className="flex items-center justify-between px-2 pb-2 pt-0.5">
          <div className="flex items-center">
            <input
              ref={fileInputRef}
              type="file"
              multiple
              className="sr-only"
              onChange={handleFileSelect}
              disabled={isBusy}
            />
            <ToolbarIcon
              title="添加附件"
              onClick={() => fileInputRef.current?.click()}
              disabled={isBusy}
            >
              <Paperclip className="size-[18px]" />
            </ToolbarIcon>
          </div>

          <div className="flex items-center gap-2 pr-1">
            {isRunning ? (
              <Button
                type="button"
                variant="outline"
                size="icon-sm"
                className="rounded-lg text-foreground"
                onClick={onStop}
                disabled={!onStop || isBusy}
              >
                <Pause className="size-4 fill-current" />
              </Button>
            ) : (
              <Button
                type="button"
                size="icon-sm"
                className={cn(
                  "rounded-lg transition",
                  canSend
                    ? "bg-primary text-primary-foreground hover:bg-primary/90"
                    : "bg-muted text-muted-foreground",
                )}
                onClick={handleSend}
                disabled={!canSend}
              >
                {sending || loading ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <ArrowUp className="size-4" />
                )}
              </Button>
            )}
          </div>
        </div>
      </div>
    );
  },
);

ChatInput.displayName = "ChatInput";

export default ChatInput;
