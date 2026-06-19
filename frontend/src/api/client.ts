import { AuthService, InitialDataService, WindowService } from "@wails/service";
import { callWails, currentPageId } from "@/lib/utils";
import type { ApiResponse } from "@/types/api/common";

interface AuthAccessTokenResponse {
  accessToken?: string;
  tokenType?: string;
}

interface RequestWithAuthOptions extends RequestInit {
  allowEmptyData?: boolean;
}

const DEFAULT_API_BASE_URL = "http://localhost:8000";
const LOGIN_EXPIRED_REASON = "登录已过期，请重新登录";

/**
 * ApiAuthExpiredError 表示远端 HTTP API 需要重新登录。
 */
export class ApiAuthExpiredError extends Error {
  constructor(message = LOGIN_EXPIRED_REASON) {
    super(message);
    this.name = "ApiAuthExpiredError";
  }
}

/**
 * getApiBaseUrl 返回远端 HTTP API 基础地址。
 */
function getApiBaseUrl(): string {
  const value = import.meta.env.VITE_BOXIFY_API_BASE_URL;
  return typeof value === "string" && value.trim()
    ? value.trim().replace(/\/$/, "")
    : DEFAULT_API_BASE_URL;
}

/**
 * normalizeBearerScheme 规范化 token type 为 Authorization scheme。
 */
function normalizeBearerScheme(tokenType?: string): string {
  const normalized = tokenType?.trim();
  if (!normalized || normalized.toLowerCase() === "bearer") {
    return "Bearer";
  }
  return normalized;
}

/**
 * redirectToLoginWithReason 打开登录窗口并传递过期提示。
 */
async function redirectToLoginWithReason(reason: string) {
  try {
    const windowNameResult = await callWails(
      WindowService.GetWindowNameByPageID,
      "login",
    );
    const targetWindow =
      typeof windowNameResult.data === "string" ? windowNameResult.data : "login";

    await InitialDataService.SaveInitialData(
      currentPageId(),
      targetWindow,
      { reason },
      5,
    );
    await callWails(WindowService.OpenPage, "login");
    await callWails(WindowService.ClosePage, currentPageId());
  } catch (error) {
    console.error("[API] 跳转登录窗口失败:", error);
  }
}

/**
 * getAccessToken 从 Wails 鉴权服务读取当前有效 token。
 */
async function getAccessToken(): Promise<AuthAccessTokenResponse> {
  try {
    const result = await AuthService.GetAccessToken();
    if (!result?.success) {
      throw new ApiAuthExpiredError(result?.message || LOGIN_EXPIRED_REASON);
    }
    const data = result.data as AuthAccessTokenResponse | undefined;
    if (!data?.accessToken) {
      throw new ApiAuthExpiredError();
    }
    return data;
  } catch (error) {
    if (error instanceof ApiAuthExpiredError) {
      throw error;
    }
    throw new ApiAuthExpiredError();
  }
}

/**
 * handleApiAuthError 统一处理登录过期错误。
 */
export async function handleApiAuthError(error: unknown): Promise<boolean> {
  if (error instanceof ApiAuthExpiredError) {
    await redirectToLoginWithReason(error.message || LOGIN_EXPIRED_REASON);
    return true;
  }
  return false;
}

/**
 * requestWithAuth 使用当前登录 token 调用远端 HTTP API。
 */
export async function requestWithAuth<T>(
  path: string,
  init: RequestWithAuthOptions = {},
): Promise<T> {
  const { allowEmptyData = false, ...requestInit } = init;
  const token = await getAccessToken();
  const headers = new Headers(requestInit.headers);
  headers.set(
    "Authorization",
    `${normalizeBearerScheme(token.tokenType)} ${token.accessToken}`,
  );
  if (requestInit.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${getApiBaseUrl()}${path}`, {
    ...requestInit,
    headers,
  });
  if (response.status === 401 || response.status === 403) {
    throw new ApiAuthExpiredError();
  }
  if (!response.ok) {
    throw new Error(`请求失败: HTTP ${response.status}`);
  }

  const payload = (await response.json()) as ApiResponse<T>;
  if (payload.code !== undefined && payload.code !== 0 && payload.code !== 200) {
    throw new Error(payload.msg || "接口返回失败");
  }
  if (payload.data === undefined || payload.data === null) {
    if (allowEmptyData) {
      return undefined as T;
    }
    throw new Error("接口响应缺少 data");
  }
  return payload.data;
}
