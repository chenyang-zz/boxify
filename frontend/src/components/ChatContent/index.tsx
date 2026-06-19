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
import {
  ArrowUp,
  ChevronDown,
  MessageCircle,
  Mic,
  Plus,
  Settings,
} from "lucide-react";
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

      <div className="shrink-0 px-3 pb-3 pt-2 sm:px-6">
        <div className="mx-auto flex w-full max-w-3xl flex-col rounded-[20px] border border-border/60 bg-card px-3 pb-2 pt-4 shadow-lg shadow-black/10">
          <Textarea
            aria-label="输入消息"
            placeholder="要求后续变更"
            className="min-h-12 resize-none border-0 bg-transparent px-0 py-0 text-sm shadow-none placeholder:text-muted-foreground/80 focus-visible:ring-0"
          />
          <div className="flex flex-wrap items-center justify-between gap-2 gap-y-1 pt-1">
            <div className="flex w-full min-w-0 items-center gap-1 sm:w-auto">
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="rounded-full text-muted-foreground"
                aria-label="添加内容"
              >
                <Plus />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="min-w-0 rounded-full px-2 text-muted-foreground"
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
                className="rounded-full px-2 text-muted-foreground"
                aria-label="选择模型强度"
              >
                <span>5.5 高</span>
                <ChevronDown />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="hidden rounded-full text-muted-foreground sm:inline-flex"
                aria-label="语音输入"
              >
                <Mic />
              </Button>
              <Button
                type="button"
                size="icon-sm"
                className="rounded-full bg-muted text-muted-foreground hover:bg-muted/90"
                aria-label="发送消息"
              >
                <ArrowUp />
              </Button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default ChatContent;
