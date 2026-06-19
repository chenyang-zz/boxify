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

import type { FC } from "react";
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuGroup,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuSub,
  ContextMenuSubContent,
  ContextMenuSubTrigger,
  ContextMenuTrigger,
} from "@/components/ui/context-menu";
import { cn } from "@/lib/utils";
import { useChatMoveContext } from "./move-context";
import type { SidebarItemContextMenuProps } from "./types";

/**
 * SidebarItemContextMenu 渲染侧边栏条目的右键菜单。
 */
export const SidebarItemContextMenu: FC<SidebarItemContextMenuProps> = ({
  children,
  onDelete,
  type = "session",
  session,
}) => {
  const moveContext = useChatMoveContext();
  const projects = moveContext?.projects ?? [];

  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>{children}</ContextMenuTrigger>
      <ContextMenuContent className="min-w-32">
        {type === "session" && session && (
          <>
            <ContextMenuSub>
              <ContextMenuSubTrigger>添加到项目</ContextMenuSubTrigger>
              <ContextMenuSubContent>
                <ContextMenuGroup>
                  {projects.length === 0 ? (
                    <ContextMenuItem disabled>暂无项目</ContextMenuItem>
                  ) : (
                    projects.map((project) => {
                      const disabled =
                        project.project_id === session.project_id;

                      return (
                        <ContextMenuItem
                          key={project.project_id}
                          disabled={disabled}
                          onSelect={() => {
                            if (!disabled) {
                              moveContext?.onMoveSessionToProject(
                                session,
                                project.project_id,
                              );
                            }
                          }}
                        >
                          <span
                            className={cn(
                              "max-w-48 truncate",
                              disabled && "text-muted-foreground",
                            )}
                          >
                            {project.name}
                          </span>
                        </ContextMenuItem>
                      );
                    })
                  )}
                </ContextMenuGroup>
              </ContextMenuSubContent>
            </ContextMenuSub>
            <ContextMenuSeparator />
          </>
        )}

        <ContextMenuGroup>
          <ContextMenuItem>置顶</ContextMenuItem>
          <ContextMenuItem>归档</ContextMenuItem>
          <ContextMenuItem>重命名</ContextMenuItem>
          <ContextMenuItem onSelect={onDelete}>删除</ContextMenuItem>
        </ContextMenuGroup>
      </ContextMenuContent>
    </ContextMenu>
  );
};
