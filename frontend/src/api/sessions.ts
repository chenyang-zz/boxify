import {
  handleApiAuthError,
  requestStreamWithAuth,
  requestWithAuth,
} from "@/api/client";
import type {
  ChatRequest,
  CreateSessionProjectRequest,
  CreateSessionResponse,
  DoneSSEEvent,
  GetSessionResponse,
  ListSessionItem,
  MessageSSEEvent,
  SessionProjectResponse,
  SessionSidebarResponse,
  UnknownSessionEvent,
  UpdateSessionRequest,
} from "@/types/api/session";

export { handleApiAuthError as handleSessionAuthError };

/**
 * getSessionSidebar 获取 Chat 侧边栏项目和会话结构。
 */
export async function getSessionSidebar(): Promise<SessionSidebarResponse> {
  return requestWithAuth<SessionSidebarResponse>("/api/sessions/sidebar");
}

/**
 * getChatSession 获取指定 Chat 会话详情和历史事件。
 */
export async function getChatSession(
  sessionId: string,
): Promise<GetSessionResponse> {
  return requestWithAuth<GetSessionResponse>(
    `/api/sessions/${encodeURIComponent(sessionId)}`,
  );
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

export interface ChatStreamHandlers {
  onMessage?: (event: MessageSSEEvent) => void;
  onDone?: (event: DoneSSEEvent) => void;
  onEvent?: (event: MessageSSEEvent | DoneSSEEvent | UnknownSessionEvent) => void;
}

/**
 * isRecord 判断未知值是否为普通对象。
 */
function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

/**
 * normalizeStreamEvent 规范化 SSE/JSON line 为会话事件结构。
 */
function normalizeStreamEvent(
  eventName: string | null,
  payload: unknown,
): MessageSSEEvent | DoneSSEEvent | UnknownSessionEvent | null {
  if (!isRecord(payload)) {
    return null;
  }

  const nestedEvent =
    typeof payload.event === "string" ? payload.event : eventName;
  const data = isRecord(payload.data) ? payload.data : payload;
  const event =
    nestedEvent ??
    (typeof data.role === "string" || typeof data.message === "string"
      ? "message"
      : "created_at" in data || "event_id" in data
        ? "done"
        : "");

  if (event === "message") {
    return {
      event: "message",
      data: {
        event_id:
          typeof data.event_id === "string" || data.event_id === null
            ? data.event_id
            : undefined,
        created_at:
          typeof data.created_at === "string" ||
          typeof data.created_at === "number" ||
          data.created_at === null
            ? data.created_at
            : undefined,
        role: data.role === "user" ? "user" : "assistant",
        message: typeof data.message === "string" ? data.message : "",
        attachments: Array.isArray(data.attachments)
          ? data.attachments.filter(isRecord)
          : [],
      },
    };
  }

  if (event === "done") {
    return {
      event: "done",
      data: {
        event_id:
          typeof data.event_id === "string" || data.event_id === null
            ? data.event_id
            : undefined,
        created_at:
          typeof data.created_at === "string" ||
          typeof data.created_at === "number" ||
          data.created_at === null
            ? data.created_at
            : undefined,
      },
    };
  }

  return event ? { event, data } : null;
}

/**
 * parseSSEFrame 解析一个完整 SSE frame。
 */
function parseSSEFrame(
  frame: string,
): MessageSSEEvent | DoneSSEEvent | UnknownSessionEvent | null {
  let eventName: string | null = null;
  const dataLines: string[] = [];

  for (const line of frame.split("\n")) {
    if (line.startsWith("event:")) {
      eventName = line.slice("event:".length).trim();
    } else if (line.startsWith("data:")) {
      dataLines.push(line.slice("data:".length).trimStart());
    }
  }

  const dataText = dataLines.join("\n").trim();
  if (!dataText && eventName === "done") {
    return normalizeStreamEvent("done", {});
  }
  if (!dataText || dataText === "[DONE]") {
    return dataText === "[DONE]" ? normalizeStreamEvent("done", {}) : null;
  }

  try {
    return normalizeStreamEvent(eventName, JSON.parse(dataText));
  } catch {
    return null;
  }
}

/**
 * parseJSONLine 兼容直接换行输出的 JSON 事件。
 */
function parseJSONLine(
  line: string,
): MessageSSEEvent | DoneSSEEvent | UnknownSessionEvent | null {
  const trimmed = line.trim();
  if (!trimmed || trimmed.startsWith("event:") || trimmed.startsWith("data:")) {
    return null;
  }

  try {
    return normalizeStreamEvent(null, JSON.parse(trimmed));
  } catch {
    return null;
  }
}

/**
 * dispatchChatStreamEvent 分发已解析的会话流事件。
 */
function dispatchChatStreamEvent(
  event: MessageSSEEvent | DoneSSEEvent | UnknownSessionEvent | null,
  handlers: ChatStreamHandlers,
) {
  if (!event) {
    return;
  }
  handlers.onEvent?.(event);
  if (event.event === "message") {
    handlers.onMessage?.(event as MessageSSEEvent);
  } else if (event.event === "done") {
    handlers.onDone?.(event as DoneSSEEvent);
  }
}

/**
 * drainChatStreamBuffer 解析缓冲区中的完整 SSE frame 或 JSON line。
 */
function drainChatStreamBuffer(
  buffer: string,
  handlers: ChatStreamHandlers,
  flush = false,
): string {
  let nextBuffer = buffer.replace(/\r\n/g, "\n");

  while (nextBuffer.includes("\n\n")) {
    const boundaryIndex = nextBuffer.indexOf("\n\n");
    const frame = nextBuffer.slice(0, boundaryIndex);
    nextBuffer = nextBuffer.slice(boundaryIndex + 2);
    dispatchChatStreamEvent(parseSSEFrame(frame), handlers);
  }

  if (!nextBuffer.includes("event:") && !nextBuffer.includes("data:")) {
    const lines = nextBuffer.split("\n");
    const completeLines = flush ? lines : lines.slice(0, -1);
    nextBuffer = flush ? "" : (lines.at(-1) ?? "");
    for (const line of completeLines) {
      dispatchChatStreamEvent(parseJSONLine(line), handlers);
    }
  } else if (flush) {
    dispatchChatStreamEvent(parseSSEFrame(nextBuffer), handlers);
    nextBuffer = "";
  }

  return nextBuffer;
}

/**
 * sendChatMessageStream 向指定会话发送消息并处理流式响应。
 */
export async function sendChatMessageStream(
  sessionId: string,
  body: ChatRequest,
  handlers: ChatStreamHandlers = {},
): Promise<void> {
  const response = await requestStreamWithAuth(
    `/api/sessions/${encodeURIComponent(sessionId)}/chat`,
    {
      method: "POST",
      body: JSON.stringify(body),
    },
  );

  if (!response.body) {
    throw new Error("接口未返回流式响应");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }
    buffer += decoder.decode(value, { stream: true });
    buffer = drainChatStreamBuffer(buffer, handlers);
  }

  buffer += decoder.decode();
  drainChatStreamBuffer(buffer, handlers, true);
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

/**
 * updateChatSession 更新指定 Chat 会话的基础信息。
 */
export async function updateChatSession(
  sessionId: string,
  body: UpdateSessionRequest,
): Promise<ListSessionItem> {
  return requestWithAuth<ListSessionItem>(
    `/api/sessions/${encodeURIComponent(sessionId)}/update`,
    {
      method: "POST",
      body: JSON.stringify(body),
    },
  );
}

/**
 * deleteChatSession 删除指定 Chat 会话。
 */
export async function deleteChatSession(sessionId: string): Promise<void> {
  await requestWithAuth<void>(
    `/api/sessions/${encodeURIComponent(sessionId)}/delete`,
    {
      method: "POST",
      allowEmptyData: true,
    },
  );
}

/**
 * deleteSessionProject 删除指定 Chat 项目及其会话。
 */
export async function deleteSessionProject(projectId: string): Promise<void> {
  await requestWithAuth<void>(
    `/api/sessions/projects/${encodeURIComponent(projectId)}/delete`,
    {
      method: "POST",
      allowEmptyData: true,
    },
  );
}
