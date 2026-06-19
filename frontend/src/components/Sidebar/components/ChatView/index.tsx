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
  type FC,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { Search, SquarePen } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { DeleteConfirmDialog } from "@/components/DeleteConfirmDialog";
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
import { useChatSessionStore } from "@/store/chat-session.store";
import {
  createChatSession,
  createSessionProject,
  deleteChatSession,
  deleteSessionProject,
  getSessionSidebar,
  handleSessionAuthError,
  updateChatSession,
} from "@/api/sessions";
import type {
  ListSessionItem,
  SessionSidebarResponse,
} from "@/types/api/session";
import {
  ActionRow,
  EmptyRow,
  InitialLoadingView,
  ProjectBlock,
  SectionTitle,
} from "./components";
import { SidebarItemContextMenu } from "./context-menu";
import { ChatMoveContext } from "./move-context";
import { NewProjectRow, SessionRow } from "./rows";
import type { DeleteTarget, PinnedItem } from "./types";
import {
  createOptimisticSession,
  moveSessionToProject,
  normalizeSidebar,
  projectContainsSession,
  removeOptimisticSession,
  replaceOptimisticSession,
  sessionTitle,
} from "./utils";

const emptySidebar: SessionSidebarResponse = {
  projects: [],
  standalone_conversations: [],
};

/**
 * ChatView 渲染 Chat tab 专属的分组式侧边栏。
 */
export const ChatView: FC = () => {
  const [sidebar, setSidebar] = useState<SessionSidebarResponse>(emptySidebar);
  const [expandedProjectIds, setExpandedProjectIds] = useState(
    () => new Set<string>(),
  );
  const [initialLoading, setInitialLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState("");
  const [projectDialogOpen, setProjectDialogOpen] = useState(false);
  const [projectName, setProjectName] = useState("");
  const [projectCreating, setProjectCreating] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null);
  const pendingSessionIdRef = useRef("");
  const selectedSessionId = useChatSessionStore(
    (state) => state.selectedSessionId,
  );
  const selectedSessionTitle = useChatSessionStore(
    (state) => state.selectedSessionTitle,
  );
  const setSelectedSession = useChatSessionStore(
    (state) => state.setSelectedSession,
  );
  const clearSelectedSessionId = useChatSessionStore(
    (state) => state.clearSelectedSessionId,
  );
  const sidebarRefreshVersion = useChatSessionStore(
    (state) => state.sidebarRefreshVersion,
  );

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

  useEffect(() => {
    if (sidebarRefreshVersion > 0) {
      void loadSidebar();
    }
  }, [loadSidebar, sidebarRefreshVersion]);

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
   * handleMoveSessionToProject 乐观移动会话并在失败时回滚。
   */
  const handleMoveSessionToProject = useCallback(
    (session: ListSessionItem, targetProjectId: string) => {
      const previousSidebar = sidebar;
      const previousExpandedProjectIds = new Set(expandedProjectIds);

      setSidebar((current) =>
        moveSessionToProject(current, session, targetProjectId),
      );
      setExpandedProjectIds((current) => {
        const next = new Set(current);
        next.add(targetProjectId);
        return next;
      });

      void (async () => {
        try {
          await updateChatSession(session.session_id, {
            project_id: targetProjectId,
          });
          await loadSidebar();
        } catch (error) {
          setSidebar(previousSidebar);
          setExpandedProjectIds(previousExpandedProjectIds);
          if (await handleSessionAuthError(error)) {
            return;
          }
          const message =
            error instanceof Error ? error.message : "移动会话失败";
          toast.error("移动会话失败", { description: message });
        }
      })();
    },
    [expandedProjectIds, loadSidebar, sidebar],
  );

  const moveContextValue = useMemo(
    () => ({
      projects,
      onMoveSessionToProject: handleMoveSessionToProject,
    }),
    [projects, handleMoveSessionToProject],
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
    const previousSelectedSessionTitle = selectedSessionTitle;
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
    setSelectedSession(tempId, sessionTitle(optimisticSession));

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
      setSelectedSession(created.session_id, sessionTitle(createdSession));
      await loadSidebar();
    } catch (error) {
      setSidebar((current) => removeOptimisticSession(current, tempId));
      if (previousSelectedSessionId) {
        setSelectedSession(
          previousSelectedSessionId,
          previousSelectedSessionTitle,
        );
      } else {
        clearSelectedSessionId();
      }
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
      await createSessionProject(name);
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
  const handleSelectSession = (session: ListSessionItem) => {
    setSelectedSession(session.session_id, sessionTitle(session));
  };

  /**
   * handleRequestDeleteSession 打开会话删除确认弹窗。
   */
  const handleRequestDeleteSession = (session: ListSessionItem) => {
    setDeleteTarget({ kind: "session", item: session });
  };

  /**
   * handleRequestDeleteProject 打开项目删除确认弹窗。
   */
  const handleRequestDeleteProject = (
    project: Extract<DeleteTarget, { kind: "project" }>["item"],
  ) => {
    setDeleteTarget({ kind: "project", item: project });
  };

  /**
   * handleConfirmDelete 确认后删除目标并同步侧栏。
   */
  const handleConfirmDelete = async () => {
    if (!deleteTarget) {
      return;
    }

    try {
      if (deleteTarget.kind === "session") {
        await deleteChatSession(deleteTarget.item.session_id);
        if (selectedSessionId === deleteTarget.item.session_id) {
          clearSelectedSessionId();
        }
      } else {
        await deleteSessionProject(deleteTarget.item.project_id);
        setExpandedProjectIds((current) => {
          const next = new Set(current);
          next.delete(deleteTarget.item.project_id);
          return next;
        });
        if (projectContainsSession(deleteTarget.item, selectedSessionId)) {
          clearSelectedSessionId();
        }
      }
      await loadSidebar();
      setDeleteTarget(null);
    } catch (error) {
      if (await handleSessionAuthError(error)) {
        throw error;
      }
      const message = error instanceof Error ? error.message : "删除失败";
      toast.error("删除失败", { description: message });
      throw error;
    }
  };

  const deleteDialogVariant =
    deleteTarget?.kind === "project" ? "project" : "chat";
  const deleteDialogTitle =
    deleteTarget?.kind === "project" ? "删除项目？" : "删除对话";
  const deleteDialogDescription =
    deleteTarget?.kind === "project"
      ? "这将永久删除所有项目文件和聊天。要保存聊天，请在删除前将其移至聊天列表或其他项目中。"
      : "此操作无法撤销。";
  const deleteDialogConfirmText =
    deleteTarget?.kind === "project" ? "删除项目" : "删除";

  return (
    <>
      <ChatMoveContext.Provider value={moveContextValue}>
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
                        <SidebarItemContextMenu
                          key={`pinned-session-${entry.item.session_id}`}
                          session={entry.item}
                          onDelete={() =>
                            handleRequestDeleteSession(entry.item)
                          }
                        >
                          <SessionRow
                            item={entry.item}
                            active={entry.item.session_id === selectedSessionId}
                            onClick={() =>
                              handleSelectSession(entry.item)
                            }
                          />
                        </SidebarItemContextMenu>
                      ) : (
                        <ProjectBlock
                          key={`pinned-project-${entry.item.project_id}`}
                          project={entry.item}
                          expanded={expandedProjectIds.has(
                            entry.item.project_id,
                          )}
                          selectedSessionId={selectedSessionId}
                          onToggle={() => toggleProject(entry.item.project_id)}
                          onSelectSession={handleSelectSession}
                          onRequestDeleteProject={handleRequestDeleteProject}
                          onRequestDeleteSession={handleRequestDeleteSession}
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
                        onRequestDeleteProject={handleRequestDeleteProject}
                        onRequestDeleteSession={handleRequestDeleteSession}
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
                      <SidebarItemContextMenu
                        key={item.session_id}
                        session={item}
                        onDelete={() => handleRequestDeleteSession(item)}
                      >
                        <SessionRow
                          item={item}
                          active={item.session_id === selectedSessionId}
                          onClick={() => handleSelectSession(item)}
                        />
                      </SidebarItemContextMenu>
                    ))
                  )}
                </div>
              </div>
            )}

            <div className="pointer-events-none absolute inset-x-0 top-0 h-6 bg-linear-to-b from-background to-transparent" />
            <div className="pointer-events-none absolute inset-x-0 bottom-0 h-8 bg-linear-to-t from-background to-transparent" />
          </div>
        </div>
      </ChatMoveContext.Provider>

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

      <DeleteConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open) {
            setDeleteTarget(null);
          }
        }}
        title={deleteDialogTitle}
        description={deleteDialogDescription}
        variant={deleteDialogVariant}
        confirmText={deleteDialogConfirmText}
        cancelText="取消"
        onConfirm={handleConfirmDelete}
      />
    </>
  );
};

export default ChatView;
