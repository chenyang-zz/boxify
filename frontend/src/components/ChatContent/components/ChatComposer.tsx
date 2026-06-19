// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { type FC, type KeyboardEvent } from "react";
import {
  ArrowUp,
  ChevronDown,
  Loader2,
  Mic,
  Plus,
  Settings,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";

interface ChatComposerProps {
  inputValue: string;
  sending: boolean;
  canSend: boolean;
  onInputChange: (value: string) => void;
  onSend: () => void;
}

/**
 * ChatComposer 渲染底部输入框和工具栏。
 */
export const ChatComposer: FC<ChatComposerProps> = ({
  inputValue,
  sending,
  canSend,
  onInputChange,
  onSend,
}) => {
  /**
   * handleComposerKeyDown 支持 Enter 发送、Shift+Enter 换行。
   */
  const handleComposerKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      onSend();
    }
  };

  return (
    <div className="shrink-0 px-3 pb-3 sm:px-6">
      <div className="shadow-composer mx-auto flex w-full max-w-3xl flex-col rounded-[20px] border border-border/70 bg-card px-3 pb-2 pt-4 text-card-foreground">
        <Textarea
          aria-label="输入消息"
          placeholder="有问题，尽管问"
          value={inputValue}
          onChange={(event) => onInputChange(event.target.value)}
          onKeyDown={handleComposerKeyDown}
          className="min-h-12 resize-none border-0 bg-transparent px-0 py-0 text-sm text-foreground shadow-none placeholder:text-muted-foreground/80 focus-visible:ring-0"
        />
        <div className="flex flex-wrap items-center justify-between gap-2 gap-y-1 pt-1">
          <div className="flex w-full min-w-0 items-center gap-1 sm:w-auto">
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="rounded-full text-muted-foreground hover:bg-muted hover:text-foreground"
              aria-label="添加内容"
            >
              <Plus />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="min-w-0 rounded-full px-2 text-muted-foreground hover:bg-muted hover:text-foreground"
              aria-label="选择自定义配置"
            >
              <Settings />
              <span className="truncate">自定义</span>
              <ChevronDown />
            </Button>
          </div>

          <div className="flex w-full min-w-0 flex-wrap items-center justify-between gap-1 sm:w-auto sm:justify-end">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="rounded-full px-2 text-muted-foreground hover:bg-muted hover:text-foreground"
              aria-label="选择模型强度"
            >
              <span>5.5 高</span>
              <ChevronDown />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="hidden rounded-full text-muted-foreground hover:bg-muted hover:text-foreground sm:inline-flex"
              aria-label="语音输入"
            >
              <Mic />
            </Button>
            <Button
              type="button"
              size="icon-sm"
              disabled={!canSend}
              onClick={onSend}
              className="rounded-full bg-primary text-primary-foreground hover:bg-primary/90"
              aria-label="发送消息"
            >
              {sending ? <Loader2 className="animate-spin" /> : <ArrowUp />}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ChatComposer;
