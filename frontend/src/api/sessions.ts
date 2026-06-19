import { handleApiAuthError, requestWithAuth } from "@/api/client";
import type {
  CreateSessionProjectRequest,
  CreateSessionResponse,
  SessionProjectResponse,
  SessionSidebarResponse,
} from "@/types/api/session";

export { handleApiAuthError as handleSessionAuthError };

/**
 * getSessionSidebar 获取 Chat 侧边栏项目和会话结构。
 */
export async function getSessionSidebar(): Promise<SessionSidebarResponse> {
  return requestWithAuth<SessionSidebarResponse>("/api/sessions/sidebar");
}

/**
 * createChatSession 创建新的 Chat 会话。
 */
export async function createChatSession(
  projectId: string | null = null,
): Promise<CreateSessionResponse> {
  return requestWithAuth<CreateSessionResponse>("/api/sessions", {
    method: "POST",
    body: JSON.stringify({
      type: "chat",
      project_id: projectId,
      is_pinned: false,
    }),
  });
}

/**
 * createSessionProject 创建新的 Chat 项目。
 */
export async function createSessionProject(
  name: string,
): Promise<SessionProjectResponse> {
  const body: CreateSessionProjectRequest = {
    name,
    sort_order: 0,
    is_pinned: false,
  };

  return requestWithAuth<SessionProjectResponse>("/api/sessions/projects", {
    method: "POST",
    body: JSON.stringify(body),
  });
}
