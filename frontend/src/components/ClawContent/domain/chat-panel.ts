import type { Conversation as ChatConversation } from "../../../../bindings/github.com/chenyang-zz/boxify/internal/claw/chat/models";

/**
 * 提取会话标题，避免空标题直接出现在会话列表中。
 */
export function getConversationTitle(conversation: ChatConversation): string {
  const title = conversation.title?.trim();
  if (title) {
    return title;
  }
  return conversation.agentId?.trim()
    ? `${conversation.agentId} 会话`
    : "未命名会话";
}

/**
 * 将时间字符串格式化为简短的本地时间。
 */
export function formatTimeLabel(value?: string): string {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return new Intl.DateTimeFormat("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    month: "2-digit",
    day: "2-digit",
  }).format(date);
}
