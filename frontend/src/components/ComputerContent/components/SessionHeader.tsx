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

import { useCallback, useState } from "react";
import { Download, FileText, PanelLeft } from "lucide-react";
import { cn, formatFileSize } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

/**
 * 会话文件描述
 */
export interface SessionFile {
  id: string;
  filename: string;
  size: number;
  extension?: string;
}

interface SessionHeaderProps {
  /** 任务/会话标题 */
  title?: string;
  /** 此任务下的文件列表 */
  files?: SessionFile[];
  /** 是否显示侧边栏切换按钮 */
  showSidebarToggle?: boolean;
  /** 点击侧边栏切换按钮的回调 */
  onToggleSidebar?: () => void;
  /** 文件列表弹窗打开状态变更 */
  onFileListOpenChange?: (open: boolean) => void;
  /** 弹窗打开时刷新文件列表 */
  onFetchFiles?: () => void;
  /** 点击文件的回调 */
  onFileClick?: (file: SessionFile) => void;
  /** 下载文件的回调 */
  onDownload?: (file: SessionFile) => void;
  /** 右侧额外内容 */
  rightSlot?: React.ReactNode;
}

/**
 * 会话标题栏组件
 *
 * 顶部粘性标题栏，包含：
 * - 侧边栏切换按钮（可选）
 * - 会话标题（截断显示）
 * - 文件列表弹窗（查看/下载此任务文件）
 */
export function SessionHeader({
  title = "",
  files,
  showSidebarToggle = false,
  onToggleSidebar,
  onFileListOpenChange,
  onFetchFiles,
  onFileClick,
  onDownload,
  rightSlot,
}: SessionHeaderProps) {
  const [internalOpen, setInternalOpen] = useState(false);
  const [downloadingId, setDownloadingId] = useState<string | null>(null);

  const isControlled = onFileListOpenChange !== undefined;
  const openState = isControlled ? internalOpen : internalOpen;
  const setOpenState = useCallback(
    (v: boolean) => {
      if (isControlled) {
        onFileListOpenChange?.(v);
      } else {
        setInternalOpen(v);
      }
      if (v && onFetchFiles) {
        onFetchFiles();
      }
    },
    [isControlled, onFileListOpenChange, onFetchFiles],
  );

  const fileList = Array.isArray(files) ? files : [];

  // 对相同文件名去重，保留最新的
  const uniqueFileList = fileList.reduce<SessionFile[]>((acc, file) => {
    const key = file.filename;
    const existingIndex = acc.findIndex((f) => f.filename === key);
    if (existingIndex >= 0) {
      acc[existingIndex] = file;
    } else {
      acc.push(file);
    }
    return acc;
  }, []);

  const handleDownload = useCallback(
    async (file: SessionFile, e: React.MouseEvent) => {
      e.stopPropagation();
      if (downloadingId) return;
      setDownloadingId(file.id);
      try {
        onDownload?.(file);
      } finally {
        setDownloadingId(null);
      }
    },
    [downloadingId, onDownload],
  );

  const handleFileClick = useCallback(
    (file: SessionFile) => {
      onFileClick?.(file);
      setOpenState(false);
    },
    [onFileClick, setOpenState],
  );

  return (
    <header className="flex items-center justify-between gap-2 sticky top-0 z-10 shrink-0  bg-background/95 backdrop-blur-sm px-4 py-3">
      <div className="text-foreground text-lg font-medium whitespace-nowrap text-ellipsis overflow-hidden flex-1 min-w-0 text-left">
        {title || "未命名任务"}
      </div>

      {/* 右侧：文件列表弹窗 */}
      <Dialog open={openState} onOpenChange={setOpenState}>
        <DialogTrigger asChild>
          <Button
            variant="ghost"
            size="icon-sm"
            className="cursor-pointer shrink-0 text-muted-foreground"
          >
            <FileText className="size-4" />
          </Button>
        </DialogTrigger>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>此任务中的所有文件</DialogTitle>
          </DialogHeader>
          <div className="max-h-125 overflow-y-auto flex flex-col gap-1 py-2">
            {uniqueFileList.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-4">
                暂无文件
              </p>
            ) : (
              uniqueFileList.map((file) => (
                <div
                  key={file.id}
                  className={cn(
                    "flex items-center gap-3 p-2 rounded-lg cursor-pointer transition-colors",
                    "hover:bg-muted",
                  )}
                  onClick={() => handleFileClick(file)}
                >
                  {/* 文件图标 */}
                  <div className="size-8 rounded-full bg-muted flex items-center justify-center shrink-0">
                    <FileText className="size-4 text-muted-foreground" />
                  </div>

                  {/* 文件信息 */}
                  <div className="flex-1 min-w-0 flex flex-col gap-0">
                    <p className="text-sm text-foreground truncate text-left">
                      {file.filename}
                    </p>
                    <p className="text-xs text-muted-foreground text-left">
                      {file.extension?.replace(/^\./, "") || "未知"} ·{" "}
                      {formatFileSize(file.size)}
                    </p>
                  </div>

                  {/* 下载按钮 */}
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    className="cursor-pointer shrink-0 text-muted-foreground hover:text-foreground"
                    onClick={(e) => handleDownload(file, e)}
                    disabled={downloadingId === file.id}
                    aria-label={`下载 ${file.filename}`}
                  >
                    <Download className="size-4" />
                  </Button>
                </div>
              ))
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* 自定义右侧插槽 */}
      {rightSlot}
    </header>
  );
}

export default SessionHeader;
