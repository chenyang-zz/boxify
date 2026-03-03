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

import { useState, useCallback, useMemo, useRef } from "react";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { FolderIcon, ChevronUpIcon, Loader2Icon, SearchIcon } from "lucide-react";
import { DirectoryInfo, ListDirectoryData } from "@wails/types/models";
import { FilesystemService } from "@wails/service";
import { Input } from "@/components/ui/input";

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
  const [searchKeyword, setSearchKeyword] = useState("");
  const menuOpenedAtRef = useRef(0);

  // 加载目录列表
  const loadDirectories = useCallback(async (path: string) => {
    setIsLoading(true);
    try {
      const result = await FilesystemService.ListDirectories(path);

      if (result?.success && result.data) {
        const data = result.data;
        const directories = data.directories.filter((d) => d !== null);

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
        menuOpenedAtRef.current = Date.now();
        setSearchKeyword("");
        // 避免打开时短暂显示上一次目录列表
        setDirectoryResult(null);
        setIsLoading(true);
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

  const handleDirectoryItemSelect = useCallback(
    (dir: DirectoryInfo, event: Event) => {
      // 防止菜单刚打开时由同一次点击/按键导致的误选
      if (Date.now() - menuOpenedAtRef.current < 150) {
        event.preventDefault();
        return;
      }

      handleDirectorySelect(dir);
    },
    [handleDirectorySelect],
  );

  const parentDirectory = useMemo(() => {
    if (!directoryResult?.parentPath) return null;
    if (directoryResult.parentPath === directoryResult.currentPath) return null;

    return {
      name: "..",
      path: directoryResult.parentPath,
      displayName: ".. (返回上一级)",
      isParent: true,
      children: undefined,
    } satisfies DirectoryInfo;
  }, [directoryResult]);

  const filteredDirectories = useMemo(() => {
    const dirs = directoryResult?.directories ?? [];
    const keyword = searchKeyword.trim().toLowerCase();
    if (!keyword) return dirs;

    return dirs.filter((dir) => {
      const name = dir?.displayName || dir?.name || "";
      return name.toLowerCase().includes(keyword);
    });
  }, [directoryResult?.directories, searchKeyword]);

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
        side="top"
        align="start"
        className="min-w-64 max-w-80 p-0 overflow-hidden"
      >
        {isLoading ? (
          <div className="flex items-center justify-center py-2">
            <Loader2Icon className="size-4 animate-spin" />
            <span className="ml-2 text-sm">加载中...</span>
          </div>
        ) : (
          <div className="w-full">
            <div className="sticky top-0 z-20 bg-popover border-b p-2 space-y-2">
              <div className="relative">
                <SearchIcon className="absolute left-2 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
                <Input
                  value={searchKeyword}
                  onChange={(e) => setSearchKeyword(e.target.value)}
                  placeholder="搜索当前目录内子文件夹"
                  className="h-8 pl-8 text-sm"
                  autoFocus
                />
              </div>

              {parentDirectory && (
                <DropdownMenuItem
                  key={parentDirectory.path}
                  onSelect={(event) =>
                    handleDirectoryItemSelect(parentDirectory, event)
                  }
                  className="cursor-pointer"
                >
                  <ChevronUpIcon className="size-4 text-muted-foreground" />
                  <span className="text-muted-foreground">
                    {parentDirectory.displayName}
                  </span>
                </DropdownMenuItem>
              )}
            </div>

            <div className="max-h-56 overflow-y-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden p-1">
              {filteredDirectories.length === 0 ? (
                <div className="px-2 py-1.5 text-sm text-muted-foreground">
                  {searchKeyword.trim() ? "未匹配到子目录" : "无子目录"}
                </div>
              ) : (
                filteredDirectories.map((dir) => (
                  <DropdownMenuItem
                    key={dir?.path}
                    onSelect={(event) => handleDirectoryItemSelect(dir!, event)}
                    className="cursor-pointer"
                  >
                    <FolderIcon className="size-4 text-cyan-200" />
                    <span>{dir?.displayName}</span>
                  </DropdownMenuItem>
                ))
              )}
            </div>
          </div>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
