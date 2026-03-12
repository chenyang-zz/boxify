import type { FC } from "react";
import { Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Conversation as ChatConversation } from "../../../../bindings/github.com/chenyang-zz/boxify/internal/claw/chat/models";
import { formatTimeLabel, getConversationTitle } from "../domain/chat-panel";
import { DEFAULT_AGENT_ID } from "../types/chat-panel";

interface ChatConversationListProps {
  className?: string;
  conversations: ChatConversation[];
  selectedConversationId: string;
  isInitializing: boolean;
  onSelectConversation: (conversationId: string) => void | Promise<void>;
}

/**
 * 渲染聊天会话列表，负责展示当前可切换的本地会话。
 */
export const ChatConversationList: FC<ChatConversationListProps> = ({
  className,
  conversations,
  selectedConversationId,
  isInitializing,
  onSelectConversation,
}) => {
  return (
    <div className={cn("min-h-0 rounded-2xl border bg-card/70", className)}>
      <div className="border-b px-4 py-3">
        <div className="text-sm font-semibold">会话列表</div>
        <div className="mt-1 text-xs text-muted-foreground">
          当前只保存在 Boxify 进程内存中。
        </div>
      </div>
      <div className="h-full max-h-full overflow-auto p-2">
        {isInitializing ? (
          <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
            <Loader2 className="mr-2 size-4 animate-spin" />
            正在初始化聊天会话
          </div>
        ) : conversations.length === 0 ? (
          <div className="px-3 py-4 text-sm text-muted-foreground">
            暂无会话
          </div>
        ) : (
          conversations.map((conversation) => {
            const active = conversation.id === selectedConversationId;
            return (
              <button
                type="button"
                key={conversation.id}
                onClick={() => void onSelectConversation(conversation.id)}
                className={cn(
                  "w-full rounded-xl border px-3 py-3 text-left transition-colors",
                  active
                    ? "border-primary/40 bg-primary/8"
                    : "border-transparent hover:border-border hover:bg-muted/40",
                )}
              >
                <div className="flex items-center justify-between gap-3">
                  <div className="truncate text-sm font-medium">
                    {getConversationTitle(conversation)}
                  </div>
                  <div className="shrink-0 text-[11px] text-muted-foreground">
                    {formatTimeLabel(
                      conversation.updatedAt ?? conversation.createdAt,
                    )}
                  </div>
                </div>
                <div className="mt-1 text-xs text-muted-foreground">
                  Agent: {conversation.agentId || DEFAULT_AGENT_ID}
                </div>
                {conversation.openClawSessionId ? (
                  <div className="mt-1 truncate text-[11px] text-muted-foreground">
                    Session: {conversation.openClawSessionId}
                  </div>
                ) : null}
              </button>
            );
          })
        )}
      </div>
    </div>
  );
};
