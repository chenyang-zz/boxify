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

import { cn, formatFileSize } from "@/lib/utils";
import { Eye, FileSearch, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";

/**
 * 附件文件数据
 */
export interface AttachmentFile {
  /** 文件唯一标识 */
  id?: string;
  /** 文件名 */
  filename: string;
  /** 扩展名 */
  extension?: string;
  /** 文件大小（字节） */
  size?: number;
  /** 预格式化的大小标签 */
  sizeLabel?: string;
}

export interface AttachmentsMessageProps {
  className?: string;
  /** 消息角色 */
  role: "user" | "assistant";
  /** 附件列表 */
  files: AttachmentFile[];
  /** 查看全部文件回调 */
  onViewAllFiles?: () => void;
  /** 文件点击回调 */
  onFileClick?: (file: AttachmentFile) => void;
}

const CARD_WIDTH = 280;
const CARD_HEIGHT = 72;

function FileCard({
  file,
  sizeLabel,
  role,
  onClick,
}: {
  file: AttachmentFile;
  index: number;
  sizeLabel: string;
  role: "user" | "assistant";
  onClick?: () => void;
}) {
  return (
    <div
      className={cn(
        "group flex shrink-0 items-center gap-3 rounded-lg border border-border bg-card p-3",
        role === "user" && "bg-card",
      )}
      style={{ width: CARD_WIDTH, height: CARD_HEIGHT }}
    >
      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
        <FileText className="size-4.5 shrink-0" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-semibold text-foreground">
          {file.filename}
        </p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          {file.extension} · {sizeLabel}
        </p>
      </div>

      {/* 预览按钮 */}
      <Button
        className=" hidden group-hover:flex"
        size="icon-sm"
        variant="ghost"
        onClick={onClick}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            onClick?.();
          }
        }}
      >
        <Eye className="size-4 shrink-0" />
      </Button>
    </div>
  );
}

/**
 * 附件消息组件
 *
 * 展示用户或 AI 发送的附件卡片列表，支持点击查看详情。
 */
export function AttachmentsMessage({
  className,
  role,
  files,
  onViewAllFiles,
  onFileClick,
}: AttachmentsMessageProps) {
  const sizeLabel = (f: AttachmentFile) =>
    f.sizeLabel ?? (typeof f.size === "number" ? formatFileSize(f.size) : "");

  if (role === "user") {
    return (
      <div
        className={cn(
          "flex flex-col flex-wrap items-end justify-end gap-2",
          className,
        )}
      >
        <div className="flex max-w-142 flex-wrap justify-end gap-2">
          {files.map((file, index) => (
            <FileCard
              key={file.id ? `${file.id}-${index}` : `file-${index}`}
              file={file}
              index={index}
              sizeLabel={sizeLabel(file)}
              role="user"
              onClick={() => onFileClick?.(file)}
            />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className={cn("flex flex-col justify-start ml-7", className)}>
      <div className="flex max-w-150 flex-wrap items-center gap-3">
        {files.map((file, index) => (
          <FileCard
            key={file.id ? `${file.id}-${index}` : `file-${index}`}
            file={file}
            index={index}
            sizeLabel={sizeLabel(file)}
            role="assistant"
            onClick={() => onFileClick?.(file)}
          />
        ))}
        {onViewAllFiles && (
          <Button
            variant="outline"
            size="sm"
            className="shrink-0 cursor-pointer gap-2 rounded-lg border-border bg-card px-3 py-2 text-muted-foreground hover:bg-accent/50 hover:text-foreground"
            style={{ width: CARD_WIDTH, height: CARD_HEIGHT }}
            onClick={onViewAllFiles}
          >
            <FileSearch className="size-4.5 shrink-0" />
            <span className="whitespace-nowrap text-sm">
              查看此任务中所有的文件
            </span>
          </Button>
        )}
      </div>
    </div>
  );
}

export default AttachmentsMessage;
