import { handleApiAuthError, requestWithAuth } from "@/api/client";
import type {
  CreateSessionResponse,
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
