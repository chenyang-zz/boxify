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

import { useState, useCallback } from "react";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { FolderIcon, ChevronUpIcon, Loader2Icon } from "lucide-react";
import { DirectoryInfo, ListDirectoryData } from "@wails/types/models";
import { FilesystemService } from "@wails/service";

interface DirectorySelectorProps {
  workPath: string;
  onDirectorySelect: (command: string) => void;
  onFocus?: () => void;
}

export function DirectorySelector({
  workPath,
  onDirectorySelect,
  onFocus,
}: DirectorySelectorProps) {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [directoryResult, setDirectoryResult] =
    useState<ListDirectoryData | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // 加载目录列表
  const loadDirectories = useCallback(async (path: string) => {
    setIsLoading(true);
    try {
      const result = await FilesystemService.ListDirectories(path);

      if (result?.success && result.data) {
        const data = result.data;
        // 添加返回上一级选项
        const directories = data.directories.filter((d) => d !== null);

        // 如果有父目录，添加返回上一级选项
        if (data.parentPath && data.parentPath !== data.currentPath) {
          directories.unshift({
            name: "..",
            path: data.parentPath,
            displayName: ".. (返回上一级)",
            isParent: true,
            children: undefined,
          });
        }

        setDirectoryResult({
          currentPath: data.currentPath,
          parentPath: data.parentPath || "",
          directories,
        });
      }
    } catch (error) {
      console.error("加载目录列表失败:", error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // 处理菜单打开
  const handleMenuOpenChange = useCallback(
    (open: boolean) => {
      setIsMenuOpen(open);
      if (open && workPath) {
        // 展开路径（将 ~ 替换为用户目录）
        FilesystemService.ExpandPath(workPath).then((expandedPath) => {
          loadDirectories(expandedPath);
        });
      }
    },
    [workPath, loadDirectories],
  );

  // 处理目录选择
  const handleDirectorySelect = useCallback(
    (dir: DirectoryInfo) => {
      if (dir.isParent) {
        // 返回上一级
        onDirectorySelect(`cd ..`);
      } else {
        // 进入子目录（使用相对路径）
        onDirectorySelect(`cd ./${dir.name}`);
      }
      setIsMenuOpen(false);
      onFocus?.();
    },
    [onDirectorySelect, onFocus],
  );

  return (
    <DropdownMenu open={isMenuOpen} onOpenChange={handleMenuOpenChange}>
      <DropdownMenuTrigger asChild>
        <Badge
          variant="secondary"
          className="border text-cyan-200 hover:bg-accent cursor-pointer"
        >
          <FolderIcon /> {workPath}
        </Badge>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="start"
        className="max-h-64 overflow-y-auto min-w-48"
      >
        {isLoading ? (
          <div className="flex items-center justify-center py-2">
            <Loader2Icon className="size-4 animate-spin" />
            <span className="ml-2 text-sm">加载中...</span>
          </div>
        ) : directoryResult?.directories.length === 0 ? (
          <div className="px-2 py-1.5 text-sm text-muted-foreground">
            无子目录
          </div>
        ) : (
          directoryResult?.directories.map((dir) => (
            <DropdownMenuItem
              key={dir?.path}
              onClick={() => handleDirectorySelect(dir!)}
              className="cursor-pointer"
            >
              {dir?.isParent ? (
                <ChevronUpIcon className="size-4 text-muted-foreground" />
              ) : (
                <FolderIcon className="size-4 text-cyan-200" />
              )}
              <span className={dir?.isParent ? "text-muted-foreground" : ""}>
                {dir?.displayName}
              </span>
            </DropdownMenuItem>
          ))
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
