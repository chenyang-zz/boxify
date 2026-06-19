export type SessionStatus = "pending" | "running" | "waiting" | "completed";
export type SessionType = "task" | "chat";

export interface ListSessionItem {
  session_id: string;
  title?: string;
  latest_message?: string;
  latest_message_at?: string | null;
  status?: SessionStatus;
  unread_message_count?: number | null;
  type?: SessionType;
  project_id?: string | null;
  is_pinned?: boolean;
}

export interface SidebarProjectItem {
  project_id: string;
  name: string;
  sort_order: number;
  is_pinned?: boolean;
  sessions?: ListSessionItem[];
}

export interface SessionSidebarResponse {
  projects?: SidebarProjectItem[];
  standalone_conversations?: ListSessionItem[];
}

export interface CreateSessionResponse {
  session_id: string;
  type: SessionType;
  project_id?: string | null;
  is_pinned?: boolean;
}

export interface CreateSessionProjectRequest {
  name: string;
  sort_order?: number;
  is_pinned?: boolean;
}

export interface SessionProjectResponse {
  project_id: string;
  name: string;
  sort_order: number;
  is_pinned?: boolean;
}

export interface UpdateSessionRequest {
  title?: string | null;
  project_id?: string | null;
  is_pinned?: boolean | null;
}

export interface ChatRequest {
  message?: string | null;
  attachments?: string[] | null;
  event_id?: string | null;
  timestamp?: number | null;
}

export interface SessionAttachment {
  file_id?: string;
  filename?: string;
  name?: string;
  url?: string;
  [key: string]: unknown;
}

export interface MessageEventData {
  event_id?: string | null;
  created_at?: string | number | null;
  role: "user" | "assistant";
  message?: string;
  attachments?: SessionAttachment[];
}

export interface DoneEventData {
  event_id?: string | null;
  created_at?: string | number | null;
}

export interface MessageSSEEvent {
  event: "message";
  data: MessageEventData;
}

export interface DoneSSEEvent {
  event: "done";
  data: DoneEventData;
}

export interface UnknownSessionEvent {
  event: string;
  data?: unknown;
}

export type SessionEvent =
  | MessageSSEEvent
  | DoneSSEEvent
  | UnknownSessionEvent;

export interface GetSessionResponse {
  session_id: string;
  title?: string | null;
  status: SessionStatus;
  events: SessionEvent[];
}
