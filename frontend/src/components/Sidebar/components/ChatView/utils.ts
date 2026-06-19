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

import type {
  ListSessionItem,
  SessionSidebarResponse,
  SidebarProjectItem,
} from "@/types/api/session";

/**
 * normalizeSidebar 将远端或本地数据规整为稳定数组结构。
 */
export function normalizeSidebar(
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
export function sessionTitle(session: ListSessionItem): string {
  return session.title?.trim() || session.latest_message?.trim() || "新对话";
}

/**
 * formatSessionTime 将会话时间格式化为紧凑相对时间。
 */
export function formatSessionTime(value?: string | null): string {
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
export function createOptimisticSession(sessionId: string): ListSessionItem {
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
export function replaceOptimisticSession(
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
export function removeOptimisticSession(
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
 * moveSessionToProject 乐观地把会话移动到目标项目顶部。
 */
export function moveSessionToProject(
  sidebar: SessionSidebarResponse | null,
  session: ListSessionItem,
  targetProjectId: string,
): Required<SessionSidebarResponse> {
  const current = normalizeSidebar(sidebar);
  let movedSession: ListSessionItem = session;

  const standaloneConversations = current.standalone_conversations.filter(
    (item) => {
      if (item.session_id === session.session_id) {
        movedSession = item;
        return false;
      }
      return true;
    },
  );

  const projects = current.projects.map((project) => {
    const sessions = project.sessions ?? [];
    return {
      ...project,
      sessions: sessions.filter((item) => {
        if (item.session_id === session.session_id) {
          movedSession = item;
          return false;
        }
        return true;
      }),
    };
  });

  return {
    projects: projects.map((project) =>
      project.project_id === targetProjectId
        ? {
            ...project,
            sessions: [
              { ...movedSession, project_id: targetProjectId },
              ...(project.sessions ?? []),
            ],
          }
        : project,
    ),
    standalone_conversations: standaloneConversations,
  };
}

/**
 * updateSessionPinned 在侧栏结构中更新指定会话置顶状态。
 */
export function updateSessionPinned(
  sidebar: SessionSidebarResponse | null,
  sessionId: string,
  isPinned: boolean,
): Required<SessionSidebarResponse> {
  const current = normalizeSidebar(sidebar);
  return {
    projects: current.projects.map((project) => ({
      ...project,
      sessions: (project.sessions ?? []).map((session) =>
        session.session_id === sessionId
          ? { ...session, is_pinned: isPinned }
          : session,
      ),
    })),
    standalone_conversations: current.standalone_conversations.map((session) =>
      session.session_id === sessionId
        ? { ...session, is_pinned: isPinned }
        : session,
    ),
  };
}

/**
 * updateProjectPinned 在侧栏结构中更新指定项目置顶状态。
 */
export function updateProjectPinned(
  sidebar: SessionSidebarResponse | null,
  projectId: string,
  isPinned: boolean,
): Required<SessionSidebarResponse> {
  const current = normalizeSidebar(sidebar);
  return {
    ...current,
    projects: current.projects.map((project) =>
      project.project_id === projectId
        ? { ...project, is_pinned: isPinned }
        : project,
    ),
  };
}

/**
 * projectContainsSession 判断项目是否包含指定会话。
 */
export function projectContainsSession(
  project: SidebarProjectItem,
  sessionId: string,
): boolean {
  return (project.sessions ?? []).some(
    (session) => session.session_id === sessionId,
  );
}
