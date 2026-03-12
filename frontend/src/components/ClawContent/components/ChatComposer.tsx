import type { FC, KeyboardEvent } from "react";
import { CornerDownLeft, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import type { Conversation as ChatConversation } from "../../../../bindings/github.com/chenyang-zz/boxify/internal/claw/chat/models";
import { DEFAULT_AGENT_ID } from "../types/chat-panel";

interface ChatComposerProps {
  draft: string;
  selectedConversation: ChatConversation | null;
  selectedConversationId: string;
  isSending: boolean;
  isCreatingConversation: boolean;
  onDraftChange: (value: string) => void;
  onComposerKeyDown: (
    event: KeyboardEvent<HTMLTextAreaElement>,
  ) => void | Promise<void>;
  onCreateConversation: () => void | Promise<void>;
  onSendMessage: () => void | Promise<void>;
}

/**
 * 渲染消息输入区，负责发送入口和当前会话上下文提示。
 */
export const ChatComposer: FC<ChatComposerProps> = ({
  draft,
  selectedConversation,
  selectedConversationId,
  isSending,
  isCreatingConversation,
  onDraftChange,
  onComposerKeyDown,
  onCreateConversation,
  onSendMessage,
}) => {
  return (
    <div className="absolute bottom-0 left-0 right-0 p-6 ">
      <div className="flex flex-col gap-3">
        <div className="transition-all duration-300 ease-out motion-safe:will-change-transform focus-within:-translate-y-2 focus-within:scale-[1.01]">
          <Textarea
            value={draft}
            onChange={(event) => onDraftChange(event.target.value)}
            onKeyDown={(event) => void onComposerKeyDown(event)}
            placeholder="输入消息，按 Enter 发送，Shift+Enter 换行"
            className="min-h-24 resize-none rounded-lg border-border/70 bg-background/90 shadow-lg shadow-black/5 transition-[box-shadow,border-color,background-color] duration-300 ease-out focus-visible:border-primary/40 focus-visible:bg-background focus-visible:shadow-2xl focus-visible:shadow-primary/10"
            disabled={!selectedConversationId || isSending}
          />
        </div>
        <div className="flex items-center justify-end gap-3">
          <div className="flex items-center gap-3">
            <Button
              variant="outline"
              className="h-10 px-5"
              onClick={() => void onCreateConversation()}
              disabled={isCreatingConversation || isSending}
            >
              新建会话
            </Button>
            <Button
              className="h-10 px-6"
              onClick={() => void onSendMessage()}
              disabled={!selectedConversationId || !draft.trim() || isSending}
            >
              {isSending ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <CornerDownLeft className="size-3.5" />
              )}
              发送
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};
