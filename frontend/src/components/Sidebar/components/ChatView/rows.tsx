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

import { forwardRef, type ComponentPropsWithoutRef, type FC } from "react";
import {
  Archive,
  ChevronRight,
  Folder,
  FolderPlus,
  FolderOpen,
  MoreHorizontal,
  Pin,
  SquarePen,
} from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import type {
  ListSessionItem,
  SidebarProjectItem,
} from "@/types/api/session";
import { formatSessionTime, sessionTitle } from "./utils";

/**
 * ProjectRow 渲染项目条目。
 */
export const ProjectRow = forwardRef<
  HTMLButtonElement,
  {
    item: SidebarProjectItem;
    expanded: boolean;
    onClick: () => void;
  } & ComponentPropsWithoutRef<"button">
>(({ item, expanded, onClick, className, ...props }, ref) => {
  const FolderIcon = expanded ? FolderOpen : Folder;

  return (
    <button
      ref={ref}
      type="button"
      onClick={onClick}
      aria-expanded={expanded}
      className={cn(
        "group flex h-8 w-full min-w-0 items-center gap-2 rounded-md px-2 text-left text-sm transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
        className,
      )}
      {...props}
    >
      <FolderIcon className="size-4 shrink-0" />
      <div className="flex min-w-0 flex-1 items-center gap-2">
        <span className="min-w-0 truncate">{item.name}</span>
        <ChevronRight
          className={cn(
            "size-3.5 shrink-0 opacity-0 transition duration-150 ease-out group-hover:opacity-100",
            expanded && "rotate-90 opacity-100",
          )}
        />
      </div>

      <span className="hidden shrink-0 items-center gap-2 group-hover:flex">
        <MoreHorizontal className="size-3.5" />
        <SquarePen className="size-3.5" />
      </span>
    </button>
  );
});
ProjectRow.displayName = "ProjectRow";

/**
 * NewProjectRow 渲染项目区的新建项目入口。
 */
export const NewProjectRow: FC<{ onClick: () => void }> = ({ onClick }) => {
  return (
    <button
      type="button"
      onClick={onClick}
      className="group flex h-8 w-full min-w-0 items-center gap-2 rounded-md px-2 text-left text-sm font-medium transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
    >
      <FolderPlus className="size-4 shrink-0" />
      <span className="min-w-0 flex-1 truncate">新建项目</span>
    </button>
  );
};

/**
 * SessionRow 渲染单行会话记录。
 */
export const SessionRow = forwardRef<
  HTMLButtonElement,
  {
    item: ListSessionItem;
    active?: boolean;
    inset?: boolean;
    onClick?: () => void;
  } & ComponentPropsWithoutRef<"button">
>(
  (
    { item, active = false, inset = false, onClick, className, ...props },
    ref,
  ) => {
    const time = formatSessionTime(item.latest_message_at);

    return (
      <button
        ref={ref}
        type="button"
        onClick={onClick}
        className={cn(
          "group flex h-8 w-full min-w-0 items-center gap-2 rounded-md text-left text-sm transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
          inset ? "pl-8 pr-2" : "px-2",
          active ? "bg-accent text-accent-foreground" : "text-foreground",
          className,
        )}
        {...props}
      >
        <span className="min-w-0 flex-1 truncate">{sessionTitle(item)}</span>
        {time ? (
          <span className="shrink-0 text-xs text-muted-foreground group-hover:hidden">
            {time}
          </span>
        ) : null}
        <span className="hidden shrink-0 items-center gap-2 text-muted-foreground group-hover:flex">
          <Tooltip>
            <TooltipTrigger asChild>
              <Pin className="size-3.5 hover:text-sidebar-accent-foreground" />
            </TooltipTrigger>
            <TooltipContent>
              <p>置顶对话</p>
            </TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <Archive className="size-3.5 hover:text-sidebar-accent-foreground" />
            </TooltipTrigger>
            <TooltipContent>
              <p>归档对话</p>
            </TooltipContent>
          </Tooltip>
        </span>
      </button>
    );
  },
);
SessionRow.displayName = "SessionRow";
