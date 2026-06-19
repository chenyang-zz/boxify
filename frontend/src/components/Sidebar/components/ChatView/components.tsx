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
import { Spinner } from "@/components/ui/spinner";
import { SidebarItemContextMenu } from "./context-menu";
import { ProjectRow, SessionRow } from "./rows";
import type { ActionRowProps, ChildrenProps, ProjectBlockProps } from "./types";

/**
 * ActionRow 渲染 Chat 侧栏顶部的快捷入口。
 */
export const ActionRow: FC<ActionRowProps> = ({
  icon: Icon,
  label,
  disabled = false,
  onClick,
}) => {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className="flex h-8 w-full items-center gap-2 rounded-md px-2 text-left text-sm transition-colors disabled:cursor-not-allowed disabled:opacity-60 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
    >
      <Icon className="size-4 shrink-0" />
      <span className="truncate">{label}</span>
    </button>
  );
};

/**
 * InitialLoadingView 渲染 Chat 分组区首次加载状态。
 */
export const InitialLoadingView: FC = () => {
  return (
    <div className="flex h-full min-h-40 flex-col items-center justify-center gap-3 px-4 text-sm text-muted-foreground">
      <Spinner className="size-5" />
      <span>正在加载会话...</span>
    </div>
  );
};

/**
 * SectionTitle 渲染置顶、项目、对话等分组标题。
 */
export const SectionTitle: FC<ChildrenProps> = ({ children }) => {
  return (
    <div className="px-2 pt-4 pb-1 text-xs font-medium text-muted-foreground">
      {children}
    </div>
  );
};

/**
 * EmptyRow 渲染分组空态。
 */
export const EmptyRow: FC<ChildrenProps> = ({ children }) => {
  return (
    <div className="flex h-8 items-center px-2 text-sm text-muted-foreground">
      <span className="truncate">{children}</span>
    </div>
  );
};

/**
 * ProjectBlock 渲染项目和展开后的会话。
 */
export const ProjectBlock: FC<ProjectBlockProps> = ({
  project,
  expanded,
  selectedSessionId,
  onToggle,
  onSelectSession,
  onRequestDeleteProject,
  onRequestDeleteSession,
}) => {
  return (
    <div className="flex flex-col gap-0.5">
      <SidebarItemContextMenu
        type="project"
        onDelete={() => onRequestDeleteProject(project)}
      >
        <ProjectRow item={project} expanded={expanded} onClick={onToggle} />
      </SidebarItemContextMenu>
      {expanded
        ? (project.sessions ?? []).map((session) => (
            <SidebarItemContextMenu
              key={session.session_id}
              session={session}
              onDelete={() => onRequestDeleteSession(session)}
            >
              <SessionRow
                item={session}
                inset
                active={session.session_id === selectedSessionId}
                onClick={() => onSelectSession(session)}
              />
            </SidebarItemContextMenu>
          ))
        : null}
    </div>
  );
};
