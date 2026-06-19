import { handleApiAuthError, requestWithAuth } from "@/api/client";

export { handleApiAuthError };

export interface AuthMeUser {
  id: string;
  username: string;
  email?: string | null;
  avatar_url?: string | null;
  oauth_provider?: string | null;
  is_active: boolean;
  is_admin: boolean;
  created_at: string;
  updated_at: string;
}

/**
 * getCurrentUserProfile 从远端认证服务读取当前登录用户详情。
 */
export async function getCurrentUserProfile(): Promise<AuthMeUser> {
  return requestWithAuth<AuthMeUser>("/api/auth/me");
}
