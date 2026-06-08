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

import { FC, useRef, useState } from "react";
import { toast } from "sonner";
import { ComputerSettings } from "./components/ComputerSettings";
import { ChatInput, ChatInputFile, ChatInputRef } from "./components/ChatInput";
import { SessionHeader } from "./components/SessionHeader";

const SUGGESTIONS = [
  "与最高的建筑相比，埃菲尔铁塔有多高？",
  "GitHub上最热门的存储库有哪些？",
  "如何看待中国的外卖大战？",
  "超加工食品与健康有关吗？超加工食品的历史怎样？",
];

/**
 * ComputerContent 核心渲染组件
 */
export const ComputerContentCore: FC = () => {
  const [sending, setSending] = useState(false);
  const [hasMessages, setHasMessages] = useState(false);
  const chatInputRef = useRef<ChatInputRef>(null);

  const handleSend = async (message: string, files: ChatInputFile[]) => {
    setSending(true);
    try {
      // TODO: 接入真实发送逻辑
      await new Promise((resolve) => setTimeout(resolve, 800));
      toast.success("已发送", {
        description: `${message.slice(0, 20)}${message.length > 20 ? "…" : ""} ${files.length > 0 ? `（${files.length} 个附件）` : ""}`,
        style: { textAlign: "left" },
      });
      setHasMessages(true);
    } finally {
      setSending(false);
    }
  };

  const handleSuggestionClick = (text: string) => {
    chatInputRef.current?.setInputText(text);
  };

  return (
    <div className="h-full w-full flex flex-col">
      {hasMessages ? (
        /* 有消息时：会话标题 + 消息列表 + 底部输入框 */
        <>
          <SessionHeader title="新任务" rightSlot={<ComputerSettings />} />
          <div className="flex-1 overflow-y-auto p-4">
            <div className="mx-auto max-w-3xl">
              <p className="text-sm text-muted-foreground">消息列表区域</p>
            </div>
          </div>
          <div className="shrink-0 p-4 chat-column-inset">
            <ChatInput
              ref={chatInputRef}
              placeholder="分配一个任务或提问任何问题…"
              loading={sending}
              onSend={handleSend}
            />
          </div>
        </>
      ) : (
        /* 无消息时：问候语 + 输入框 + 推荐任务居中 */
        <div className="flex-1 flex flex-col items-center justify-center p-4 overflow-y-auto">
          <div className="mx-auto max-w-3xl w-full flex flex-col items-center gap-8">
            {/* 问候语 */}
            <div className="text-center w-full">
              <h1 className="text-3xl font-bold text-foreground">
                您想要做什么？
              </h1>
            </div>

            {/* 输入框 */}
            <div className="w-full">
              <ChatInput
                ref={chatInputRef}
                placeholder="给MooCManus一个任务"
                loading={sending}
                onSend={handleSend}
              />
            </div>

            {/* 推荐任务 */}
            <div className="w-full flex flex-wrap gap-3">
              {SUGGESTIONS.map((suggestion) => (
                <button
                  key={suggestion}
                  type="button"
                  onClick={() => handleSuggestionClick(suggestion)}
                  className="rounded-full border border-border bg-card px-4 py-2 text-sm text-foreground transition hover:bg-accent hover:text-accent-foreground"
                >
                  {suggestion}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
