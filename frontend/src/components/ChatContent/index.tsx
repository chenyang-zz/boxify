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

import { FC } from "react";
import { MessageCircle, Paperclip, SendHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";

/**
 * ChatContent 渲染 Chat tab 的独立主内容区域。
 */
const ChatContent: FC = () => {
  return (
    <section className="flex h-full min-w-0 flex-col bg-background">
      <header className="flex h-10 shrink-0 items-center border-b px-4">
        <div className="flex min-w-0 items-center gap-2">
          <div className="flex size-6 items-center justify-center rounded-md bg-primary/10 text-primary">
            <MessageCircle className="size-3.5" />
          </div>
          <div className="truncate text-sm font-semibold">Chat</div>
        </div>
      </header>

      <div className="flex flex-1 items-center justify-center overflow-hidden px-6">
        <div className="flex w-full max-w-2xl flex-col items-center gap-3 text-center">
          <div className="flex size-12 items-center justify-center rounded-full bg-muted text-primary">
            <MessageCircle className="size-6" />
          </div>
          <div className="text-lg font-semibold">新对话</div>
          <p className="max-w-md text-sm leading-6 text-muted-foreground">
            暂无消息记录
          </p>
        </div>
      </div>

      <div className="shrink-0 border-t bg-card/80 p-4">
        <div className="mx-auto flex max-w-3xl flex-col gap-2 rounded-md border bg-background p-3 shadow-xs">
          <Textarea
            aria-label="输入消息"
            placeholder="输入消息..."
            className="min-h-20 resize-none border-0 bg-transparent px-0 shadow-none focus-visible:ring-0"
          />
          <div className="flex items-center justify-between">
            <Button variant="ghost" size="icon-sm" aria-label="添加附件">
              <Paperclip />
            </Button>
            <Button size="sm">
              <SendHorizontal />
              发送
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
};

export default ChatContent;
