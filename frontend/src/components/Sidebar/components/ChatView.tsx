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
  type ElementType,
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  Archive,
  ChevronRight,
  Folder,
  FolderOpen,
  MoreHorizontal,
  Pin,
  Search,
  SquarePen,
} from "lucide-react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import {
  createChatSession,
  getSessionSidebar,
  handleSessionAuthError,
} from "@/api/sessions";
import type {
  ListSessionItem,
  SessionSidebarResponse,
  SidebarProjectItem,
} from "@/types/api/session";

type PinnedItem =
  | { kind: "session"; item: ListSessionItem }
  | { kind: "project"; item: SidebarProjectItem };

/**
 * sessionTitle 返回侧边栏会话标题。
 */
function sessionTitle(session: ListSessionItem): string {
  return (
    session.title?.trim() ||
    session.latest_message?.trim() ||
    "新对话"
  );
}

/**
 * formatSessionTime 将会话时间格式化为紧凑相对时间。
 */
function formatSessionTime(value?: string | null): string {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  const time = date.getTime();
  if (Number.isNaN(time)) {
    return "";
  }

  const diffMs = Math.max(0, Date.now() - time);
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);
  if (diffMins < 1) return "刚刚";
  if (diffMins < 60) return `${diffMins}分`;
  if (diffHours < 24) return `${diffHours}时`;
  if (diffDays < 7) return `${diffDays}天`;
  if (diffDays < 30) return `${Math.floor(diffDays / 7)}周`;
  return `${Math.floor(diffDays / 30)}个月`;
}

/**
 * ActionRow 渲染 Chat 侧栏顶部的快捷入口。
 */
const ActionRow: FC<{
  icon: ElementType;
  label: string;
  disabled?: boolean;
  onClick?: () => void;
}> = ({ icon: Icon, label, disabled = false, onClick }) => {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className="flex h-8 w-full items-center gap-2 rounded-md px-2 text-left text-sm text-foreground transition-colors hover:bg-accent/60 disabled:cursor-not-allowed disabled:opacity-60"
    >
      <Icon className="size-4 shrink-0 text-muted-foreground" />
      <span className="truncate">{label}</span>
    </button>
  );
};

/**
 * SectionTitle 渲染置顶、项目、对话等分组标题。
 */
const SectionTitle: FC<{ children: ReactNode }> = ({ children }) => {
  return (
    <div className="px-2 pt-4 pb-1 text-xs font-medium text-muted-foreground">
      {children}
    </div>
  );
};

/**
 * EmptyRow 渲染分组空态。
 */
const EmptyRow: FC<{ children: ReactNode }> = ({ children }) => {
  return (
    <div className="flex h-8 items-center px-2 text-sm text-muted-foreground">
      <span className="truncate">{children}</span>
    </div>
  );
};

/**
 * ProjectRow 渲染项目条目。
 */
const ProjectRow: FC<{
  item: SidebarProjectItem;
  expanded: boolean;
  onClick: () => void;
}> = ({ item, expanded, onClick }) => {
  const FolderIcon = expanded ? FolderOpen : Folder;

  return (
    <button
      type="button"
      onClick={onClick}
      aria-expanded={expanded}
      className="group flex h-8 w-full min-w-0 items-center gap-2 rounded-md px-2 text-left text-sm text-foreground transition-colors hover:bg-accent/60"
    >
      <FolderIcon className="size-4 shrink-0 text-muted-foreground" />
      <div className="flex min-w-0 flex-1 items-center gap-2">
        <span className="min-w-0 truncate">{item.name}</span>
        <ChevronRight
          className={cn(
            "size-3.5 shrink-0 text-muted-foreground opacity-0 transition duration-150 ease-out group-hover:opacity-100",
            expanded && "rotate-90 opacity-100",
          )}
        />
      </div>

      <span className="hidden shrink-0 items-center gap-2 text-muted-foreground group-hover:flex">
        <MoreHorizontal className="size-3.5" />
        <SquarePen className="size-3.5" />
      </span>
    </button>
  );
};

/**
 * SessionRow 渲染单行会话记录。
 */
const SessionRow: FC<{
  item: ListSessionItem;
  active?: boolean;
  inset?: boolean;
  onClick?: () => void;
}> = ({ item, active = false, inset = false, onClick }) => {
  const time = formatSessionTime(item.latest_message_at);

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "group flex h-8 w-full min-w-0 items-center gap-2 rounded-md text-left text-sm transition-colors hover:bg-accent/60",
        inset ? "pl-8 pr-2" : "px-2",
        active ? "bg-muted text-foreground" : "text-foreground",
      )}
    >
      <span className="min-w-0 flex-1 truncate">{sessionTitle(item)}</span>
      {time ? (
        <span className="shrink-0 text-xs text-muted-foreground group-hover:hidden">
          {time}
        </span>
      ) : null}
      <span className="hidden shrink-0 items-center gap-2 text-muted-foreground group-hover:flex">
        <Pin className="size-3.5" />
        <Archive className="size-3.5" />
      </span>
    </button>
  );
};

/**
 * ProjectBlock 渲染项目和展开后的会话。
 */
const ProjectBlock: FC<{
  project: SidebarProjectItem;
  expanded: boolean;
  selectedSessionId: string;
  onToggle: () => void;
  onSelectSession: (sessionId: string) => void;
}> = ({ project, expanded, selectedSessionId, onToggle, onSelectSession }) => {
  return (
    <div className="flex flex-col gap-0.5">
      <ProjectRow item={project} expanded={expanded} onClick={onToggle} />
      {expanded
        ? (project.sessions ?? []).map((session) => (
            <SessionRow
              key={session.session_id}
              item={session}
              inset
              active={session.session_id === selectedSessionId}
              onClick={() => onSelectSession(session.session_id)}
            />
          ))
        : null}
    </div>
  );
};

/**
 * ChatView 渲染 Chat tab 专属的分组式侧边栏。
 */
export const ChatView: FC = () => {
  const [sidebar, setSidebar] = useState<SessionSidebarResponse | null>(null);
  const [expandedProjectIds, setExpandedProjectIds] = useState(
    () => new Set<string>(),
  );
  const [selectedSessionId, setSelectedSessionId] = useState("");
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  /**
   * loadSidebar 加载远端侧边栏结构。
   */
  const loadSidebar = useCallback(async () => {
    setLoading(true);
    setErrorMessage("");
    try {
      const data = await getSessionSidebar();
      setSidebar({
        projects: data.projects ?? [],
        standalone_conversations: data.standalone_conversations ?? [],
      });
    } catch (error) {
      if (await handleSessionAuthError(error)) {
        return;
      }
      const message =
        error instanceof Error ? error.message : "加载会话侧边栏失败";
      setErrorMessage(message);
      toast.error("加载会话失败", { description: message });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadSidebar();
  }, [loadSidebar]);

  const projects = sidebar?.projects ?? [];
  const standaloneConversations = sidebar?.standalone_conversations ?? [];

  const pinnedItems = useMemo<PinnedItem[]>(() => {
    const pinnedStandaloneSessions = standaloneConversations
      .filter((session) => session.is_pinned)
      .map((item): PinnedItem => ({ kind: "session", item }));
    const pinnedProjectSessions = projects
      .flatMap((project) => project.sessions ?? [])
      .filter((session) => session.is_pinned)
      .map((item): PinnedItem => ({ kind: "session", item }));
    const pinnedProjects = projects
      .filter((project) => project.is_pinned)
      .map((item): PinnedItem => ({ kind: "project", item }));

    return [
      ...pinnedStandaloneSessions,
      ...pinnedProjectSessions,
      ...pinnedProjects,
    ];
  }, [projects, standaloneConversations]);

  const regularProjects = useMemo(
    () => projects.filter((project) => !project.is_pinned),
    [projects],
  );

  const regularConversations = useMemo(
    () => standaloneConversations.filter((session) => !session.is_pinned),
    [standaloneConversations],
  );

  /**
   * toggleProject 切换项目的本地展开状态。
   */
  const toggleProject = (projectId: string) => {
    setExpandedProjectIds((current) => {
      const next = new Set(current);
      if (next.has(projectId)) {
        next.delete(projectId);
      } else {
        next.add(projectId);
      }
      return next;
    });
  };

  /**
   * handleCreateSession 创建新 Chat 会话并刷新侧栏。
   */
  const handleCreateSession = async () => {
    if (creating) {
      return;
    }
    setCreating(true);
    try {
      const created = await createChatSession(null);
      setSelectedSessionId(created.session_id);
      await loadSidebar();
    } catch (error) {
      if (await handleSessionAuthError(error)) {
        return;
      }
      const message = error instanceof Error ? error.message : "创建会话失败";
      toast.error("创建会话失败", { description: message });
    } finally {
      setCreating(false);
    }
  };

  /**
   * handleSelectSession 记录当前选中的侧边栏会话。
   */
  const handleSelectSession = (sessionId: string) => {
    setSelectedSessionId(sessionId);
  };

  return (
    <div className="flex h-full w-full min-w-0 flex-col overflow-hidden bg-sidebar">
      <div className="shrink-0 px-2 pt-2 pb-1">
        <div className="flex flex-col gap-0.5">
          <ActionRow
            icon={SquarePen}
            label={creating ? "创建中..." : "新对话"}
            disabled={creating}
            onClick={() => void handleCreateSession()}
          />
          <ActionRow icon={Search} label="搜索" />
        </div>
      </div>

      <div className="relative min-h-0 flex-1">
        <div className="h-full overflow-auto px-2 pb-2">
          <SectionTitle>置顶</SectionTitle>
          <div className="flex flex-col gap-0.5">
            {loading ? (
              <EmptyRow>加载中...</EmptyRow>
            ) : pinnedItems.length === 0 ? (
              <EmptyRow>暂无置顶</EmptyRow>
            ) : (
              pinnedItems.map((entry) =>
                entry.kind === "session" ? (
                  <SessionRow
                    key={`pinned-session-${entry.item.session_id}`}
                    item={entry.item}
                    active={entry.item.session_id === selectedSessionId}
                    onClick={() => handleSelectSession(entry.item.session_id)}
                  />
                ) : (
                  <ProjectBlock
                    key={`pinned-project-${entry.item.project_id}`}
                    project={entry.item}
                    expanded={expandedProjectIds.has(entry.item.project_id)}
                    selectedSessionId={selectedSessionId}
                    onToggle={() => toggleProject(entry.item.project_id)}
                    onSelectSession={handleSelectSession}
                  />
                ),
              )
            )}
          </div>

          <SectionTitle>项目</SectionTitle>
          <div className="flex flex-col gap-0.5">
            {loading ? (
              <EmptyRow>加载中...</EmptyRow>
            ) : regularProjects.length === 0 ? (
              <EmptyRow>暂无项目</EmptyRow>
            ) : (
              regularProjects.map((project) => (
                <ProjectBlock
                  key={project.project_id}
                  project={project}
                  expanded={expandedProjectIds.has(project.project_id)}
                  selectedSessionId={selectedSessionId}
                  onToggle={() => toggleProject(project.project_id)}
                  onSelectSession={handleSelectSession}
                />
              ))
            )}
          </div>

          <SectionTitle>对话</SectionTitle>
          <div className="flex flex-col gap-0.5">
            {loading ? (
              <EmptyRow>加载中...</EmptyRow>
            ) : errorMessage ? (
              <EmptyRow>{errorMessage}</EmptyRow>
            ) : regularConversations.length === 0 ? (
              <EmptyRow>暂无对话</EmptyRow>
            ) : (
              regularConversations.map((item) => (
                <SessionRow
                  key={item.session_id}
                  item={item}
                  active={item.session_id === selectedSessionId}
                  onClick={() => handleSelectSession(item.session_id)}
                />
              ))
            )}
          </div>
        </div>

        <div className="pointer-events-none absolute inset-x-0 top-0 h-6 bg-gradient-to-b from-sidebar to-transparent" />
        <div className="pointer-events-none absolute inset-x-0 bottom-0 h-8 bg-gradient-to-t from-sidebar to-transparent" />
      </div>
    </div>
  );
};

export default ChatView;
