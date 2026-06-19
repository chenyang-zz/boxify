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
  useRef,
  useState,
} from "react";
import {
  Archive,
  ChevronRight,
  Folder,
  FolderPlus,
  FolderOpen,
  MoreHorizontal,
  Pin,
  Search,
  SquarePen,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { cn } from "@/lib/utils";
import {
  createChatSession,
  createSessionProject,
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

const emptySidebar: SessionSidebarResponse = {
  projects: [],
  standalone_conversations: [],
};

/**
 * normalizeSidebar 将远端或本地数据规整为稳定数组结构。
 */
function normalizeSidebar(
  data?: SessionSidebarResponse | null,
): Required<SessionSidebarResponse> {
  return {
    projects: data?.projects ?? [],
    standalone_conversations: data?.standalone_conversations ?? [],
  };
}

/**
 * sessionTitle 返回侧边栏会话标题。
 */
function sessionTitle(session: ListSessionItem): string {
  return session.title?.trim() || session.latest_message?.trim() || "新对话";
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
 * createOptimisticSession 生成本地临时会话。
 */
function createOptimisticSession(sessionId: string): ListSessionItem {
  return {
    session_id: sessionId,
    title: "新对话",
    latest_message: "",
    latest_message_at: new Date().toISOString(),
    status: "pending",
    unread_message_count: 0,
    type: "chat",
    project_id: null,
    is_pinned: false,
  };
}

/**
 * replaceOptimisticSession 用真实会话 id 替换临时会话。
 */
function replaceOptimisticSession(
  sidebar: SessionSidebarResponse | null,
  tempId: string,
  session: ListSessionItem,
): Required<SessionSidebarResponse> {
  const current = normalizeSidebar(sidebar);
  return {
    ...current,
    standalone_conversations: current.standalone_conversations.map((item) =>
      item.session_id === tempId ? { ...item, ...session } : item,
    ),
  };
}

/**
 * removeOptimisticSession 移除创建失败的临时会话。
 */
function removeOptimisticSession(
  sidebar: SessionSidebarResponse | null,
  tempId: string,
): Required<SessionSidebarResponse> {
  const current = normalizeSidebar(sidebar);
  return {
    ...current,
    standalone_conversations: current.standalone_conversations.filter(
      (item) => item.session_id !== tempId,
    ),
  };
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
 * InitialLoadingView 渲染 Chat 分组区首次加载状态。
 */
const InitialLoadingView: FC = () => {
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
 * NewProjectRow 渲染项目区的新建项目入口。
 */
const NewProjectRow: FC<{ onClick: () => void }> = ({ onClick }) => {
  return (
    <button
      type="button"
      onClick={onClick}
      className="group flex h-8 w-full min-w-0 items-center gap-2 rounded-md px-2 text-left text-sm font-medium text-foreground transition-colors hover:bg-accent/60"
    >
      <FolderPlus className="size-4 shrink-0 text-muted-foreground" />
      <span className="min-w-0 flex-1 truncate">新建项目</span>
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
  const [sidebar, setSidebar] = useState<SessionSidebarResponse>(emptySidebar);
  const [expandedProjectIds, setExpandedProjectIds] = useState(
    () => new Set<string>(),
  );
  const [selectedSessionId, setSelectedSessionId] = useState("");
  const [initialLoading, setInitialLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState("");
  const [projectDialogOpen, setProjectDialogOpen] = useState(false);
  const [projectName, setProjectName] = useState("");
  const [projectCreating, setProjectCreating] = useState(false);
  const pendingSessionIdRef = useRef("");

  /**
   * loadSidebar 静默同步远端侧边栏结构。
   */
  const loadSidebar = useCallback(async (isInitial = false) => {
    setErrorMessage("");
    try {
      const data = await getSessionSidebar();
      setSidebar(normalizeSidebar(data));
    } catch (error) {
      if (await handleSessionAuthError(error)) {
        return;
      }
      const message =
        error instanceof Error ? error.message : "加载会话侧边栏失败";
      setErrorMessage(message);
      toast.error("加载会话失败", { description: message });
    } finally {
      if (isInitial) {
        setInitialLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    void loadSidebar(true);
  }, [loadSidebar]);

  const projects = sidebar.projects ?? [];
  const standaloneConversations = sidebar.standalone_conversations ?? [];

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
   * handleCreateSession 乐观创建新 Chat 会话并刷新侧栏。
   */
  const handleCreateSession = async () => {
    if (pendingSessionIdRef.current) {
      return;
    }

    const previousSelectedSessionId = selectedSessionId;
    const tempId = `optimistic-session-${Date.now()}`;
    const optimisticSession = createOptimisticSession(tempId);
    pendingSessionIdRef.current = tempId;
    setSidebar((current) => {
      const normalized = normalizeSidebar(current);
      return {
        ...normalized,
        standalone_conversations: [
          optimisticSession,
          ...normalized.standalone_conversations,
        ],
      };
    });
    setSelectedSessionId(tempId);

    try {
      const created = await createChatSession(null);
      const createdSession: ListSessionItem = {
        ...optimisticSession,
        session_id: created.session_id,
        type: created.type,
        project_id: created.project_id ?? null,
        is_pinned: created.is_pinned ?? false,
      };
      setSidebar((current) =>
        replaceOptimisticSession(current, tempId, createdSession),
      );
      setSelectedSessionId(created.session_id);
      await loadSidebar();
    } catch (error) {
      setSidebar((current) => removeOptimisticSession(current, tempId));
      setSelectedSessionId(previousSelectedSessionId);
      if (await handleSessionAuthError(error)) {
        return;
      }
      const message = error instanceof Error ? error.message : "创建会话失败";
      toast.error("创建会话失败", { description: message });
    } finally {
      pendingSessionIdRef.current = "";
    }
  };

  /**
   * handleCreateProject 请求成功后再把新项目同步到侧栏。
   */
  const handleCreateProject = async () => {
    const name = projectName.trim();
    if (!name || projectCreating) {
      return;
    }

    setProjectCreating(true);

    try {
      const created = await createSessionProject(name);
      await loadSidebar();
      setProjectDialogOpen(false);
      setProjectName("");
    } catch (error) {
      if (await handleSessionAuthError(error)) {
        return;
      }
      const message = error instanceof Error ? error.message : "创建项目失败";
      toast.error("创建项目失败", { description: message });
    } finally {
      setProjectCreating(false);
    }
  };

  /**
   * handleSelectSession 记录当前选中的侧边栏会话。
   */
  const handleSelectSession = (sessionId: string) => {
    setSelectedSessionId(sessionId);
  };

  return (
    <>
      <div className="flex h-full w-full min-w-0 flex-col overflow-hidden bg-sidebar">
        <div className="shrink-0 px-2 pt-2 pb-1">
          <div className="flex flex-col gap-0.5">
            <ActionRow
              icon={SquarePen}
              label="新对话"
              onClick={() => void handleCreateSession()}
            />
            <ActionRow icon={Search} label="搜索" />
          </div>
        </div>

        <div className="relative min-h-0 flex-1">
          {initialLoading ? (
            <InitialLoadingView />
          ) : (
            <div className="h-full overflow-auto px-2 pb-2">
              <SectionTitle>置顶</SectionTitle>
              <div className="flex flex-col gap-0.5">
                {pinnedItems.length === 0 ? (
                  <EmptyRow>暂无置顶</EmptyRow>
                ) : (
                  pinnedItems.map((entry) =>
                    entry.kind === "session" ? (
                      <SessionRow
                        key={`pinned-session-${entry.item.session_id}`}
                        item={entry.item}
                        active={entry.item.session_id === selectedSessionId}
                        onClick={() =>
                          handleSelectSession(entry.item.session_id)
                        }
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
                <NewProjectRow onClick={() => setProjectDialogOpen(true)} />
                {regularProjects.length === 0 ? (
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
                {errorMessage ? (
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
          )}

          <div className="pointer-events-none absolute inset-x-0 top-0 h-6 bg-gradient-to-b from-sidebar to-transparent" />
          <div className="pointer-events-none absolute inset-x-0 bottom-0 h-8 bg-gradient-to-t from-sidebar to-transparent" />
        </div>
      </div>

      <Dialog
        open={projectDialogOpen}
        onOpenChange={(open) => {
          if (!projectCreating) {
            setProjectDialogOpen(open);
          }
        }}
      >
        <DialogContent className="sm:max-w-sm">
          <form
            className="flex flex-col gap-4"
            onSubmit={(event) => {
              event.preventDefault();
              void handleCreateProject();
            }}
          >
            <DialogHeader>
              <DialogTitle>新建项目</DialogTitle>
              <DialogDescription>
                创建后会出现在 Chat 侧边栏项目分组中。
              </DialogDescription>
            </DialogHeader>

            <Input
              autoFocus
              disabled={projectCreating}
              value={projectName}
              onChange={(event) => setProjectName(event.target.value)}
              placeholder="项目名称"
            />

            <DialogFooter>
              <DialogClose asChild>
                <Button
                  type="button"
                  variant="outline"
                  disabled={projectCreating}
                >
                  取消
                </Button>
              </DialogClose>
              <Button
                type="submit"
                disabled={!projectName.trim() || projectCreating}
              >
                {projectCreating ? <Spinner data-icon="inline-start" /> : null}
                创建项目
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
};

export default ChatView;
