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

import { type FC } from "react";
import { MoreHorizontal } from "lucide-react";

import { Button } from "@/components/ui/button";

interface ChatHeaderProps {
  title: string;
}

/**
 * ChatHeader 渲染聊天区顶部标题和响应式覆盖层。
 */
export const ChatHeader: FC<ChatHeaderProps> = ({ title }) => {
  return (
    <header className="relative z-10 flex shrink-0 h-12  items-center px-5 bg-background text-foreground xl:absolute xl:inset-x-0 xl:top-0 xl:bg-background/0  ">
      <div className="flex min-w-0 items-center gap-1.5">
        <div className="truncate text-sm font-semibold leading-none">
          {title}
        </div>
        <Button
          type="button"
          variant="ghost"
          size="icon-xs"
          className="rounded-full text-muted-foreground hover:bg-muted hover:text-foreground"
          aria-label="更多聊天操作"
          title="更多聊天操作"
        >
          <MoreHorizontal />
        </Button>
      </div>
    </header>
  );
};

export default ChatHeader;
