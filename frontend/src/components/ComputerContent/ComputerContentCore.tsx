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

import { FC, useState } from "react";
import { toast } from "sonner";
import { ComputerSettings } from "./components/ComputerSettings";
import { ChatInput, ChatInputFile } from "./components/ChatInput";

/**
 * ComputerContent 核心渲染组件
 */
export const ComputerContentCore: FC = () => {
  const [sending, setSending] = useState(false);

  const handleSend = async (message: string, files: ChatInputFile[]) => {
    setSending(true);
    try {
      // TODO: 接入真实发送逻辑
      await new Promise((resolve) => setTimeout(resolve, 800));
      toast.success("已发送", {
        description: `${message.slice(0, 20)}${message.length > 20 ? "…" : ""} ${files.length > 0 ? `（${files.length} 个附件）` : ""}`,
        style: { textAlign: "left" },
      });
    } finally {
      setSending(false);
    }
  };

  return (
    <div className="h-full w-full flex flex-col">
      {/* 顶部工具栏 */}
      <div className="flex items-center justify-end p-2 border-b border-border">
        <ComputerSettings />
      </div>

      {/* 内容区域 */}
      <div className="flex-1 flex flex-col min-h-0">
        {/* 消息区域占位 */}
        <div className="flex-1 overflow-y-auto p-4">
          <div className="mx-auto max-w-3xl h-full flex items-center justify-center">
            <p className="text-sm text-muted-foreground">消息列表区域</p>
          </div>
        </div>

        {/* 底部输入区 */}
        <div className="shrink-0 p-4 chat-column-inset">
          <ChatInput
            placeholder="分配一个任务或提问任何问题…"
            loading={sending}
            onSend={handleSend}
          />
        </div>
      </div>
    </div>
  );
};
