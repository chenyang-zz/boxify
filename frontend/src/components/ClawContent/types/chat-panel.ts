import type { Message as ChatMessage } from "../../../../bindings/github.com/chenyang-zz/boxify/internal/claw/chat/models";

export const DEFAULT_AGENT_ID = "main";
export const PENDING_ASSISTANT_ID = "pending-assistant";

export interface PendingAssistantDraft {
  conversationId: string;
  runId: string;
  content: string;
  status: string;
}

export interface RenderedChatMessage extends ChatMessage {
  id: string;
}
