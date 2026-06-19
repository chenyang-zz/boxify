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
