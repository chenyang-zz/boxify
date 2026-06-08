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
import { PlanPanel, PlanStep } from "./components/PlanPanel";
import { ChatMessage, ChatMessageItem } from "./components/chat";

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
  const [hasMessages, setHasMessages] = useState(true);
  const chatInputRef = useRef<ChatInputRef>(null);

  // Demo 消息数据 — TODO: 接入真实消息数据
  const messages: ChatMessageItem[] = [
    {
      id: "msg-1",
      kind: "user",
      data: { message: "帮我分析一下当前项目的代码结构" },
      createdAt: "2026-06-08T10:00:00Z",
    },
    {
      id: "msg-attachments",
      kind: "attachments",
      role: "user",
      files: [
        {
          id: "att-1",
          filename: "project-overview.pdf",
          extension: "PDF",
          size: 245760,
        },
        {
          id: "att-2",
          filename: "screenshot-design.png",
          extension: "PNG",
          size: 1048576,
          sizeLabel: "1.0 MB",
        },
        {
          id: "att-3",
          filename: "screenshot-design.png",
          extension: "PNG",
          size: 1048576,
          sizeLabel: "1.0 MB",
        },
        {
          id: "att-4",
          filename: "screenshot-design.png",
          extension: "PNG",
          size: 1048576,
          sizeLabel: "1.0 MB",
        },
      ],
      createdAt: "2026-06-08T10:00:08Z",
    },
    {
      id: "msg-2",
      kind: "assistant",
      data: {
        message:
          "好的，我来为您分析项目代码结构。\n\n首先，让我查看项目根目录的文件列表...\n\n项目采用了以下架构：\n1. 前端使用 React + TypeScript + Tailwind CSS\n2. 后端使用 Go + Wails 框架\n3. 组件库使用 shadcn/ui\n\n主要目录结构：\n- frontend/src/components/  UI 组件\n- frontend/src/lib/         工具函数\n- backend/                  Go 后端代码",
      },
      createdAt: "2026-06-08T10:00:05Z",
    },
    {
      id: "step-1",
      kind: "step",
      step: {
        data: {
          id: "1",
          description: "分析任务需求并制定执行计划",
          status: "completed",
        },
        tools: [
          { name: "read_directory", startTime: "2026-06-08T10:00:01Z" },
          {
            name: "bash",
            args: { command: "ls -la" },
            startTime: "2026-06-08T10:00:01Z",
          },
          { name: "search_files", startTime: "2026-06-08T10:00:02Z" },
        ],
      },
    },
    {
      id: "step-2",
      kind: "step",
      step: {
        data: {
          id: "2",
          description: "搜索相关资料和参考文档",
          status: "completed",
        },
        tools: [
          { name: "web_search", startTime: "2026-06-08T10:00:03Z" },
          { name: "fetch_page", startTime: "2026-06-08T10:00:04Z" },
        ],
      },
    },
    {
      id: "step-3",
      kind: "step",
      step: {
        data: {
          id: "3",
          description: "编写核心代码实现",
          status: "in_progress",
        },
        tools: [{ name: "write_file", startTime: "2026-06-08T10:00:05Z" }],
      },
    },

    {
      id: "msg-attachments-ai",
      kind: "attachments",
      role: "assistant",
      files: [
        {
          id: "att-3",
          filename: "analysis-report.md",
          extension: "MD",
          size: 5120,
        },
        {
          id: "att-2",
          filename: "screenshot-design.png",
          extension: "PNG",
          size: 1048576,
          sizeLabel: "1.0 MB",
        },
        {
          id: "att-3",
          filename: "screenshot-design.png",
          extension: "PNG",
          size: 1048576,
          sizeLabel: "1.0 MB",
        },
      ],
      createdAt: "2026-06-08T10:00:12Z",
    },
  ];

  // Demo plan steps — TODO: 接入真实计划数据
  // Demo session files — TODO: 接入真实会话文件
  const sessionFiles = [
    {
      id: "sf-1",
      filename: "project-overview.pdf",
      size: 245760,
      extension: "pdf",
    },
    {
      id: "sf-2",
      filename: "screenshot-design.png",
      size: 1048576,
      extension: "png",
    },
    {
      id: "sf-3",
      filename: "analysis-report.md",
      size: 5120,
      extension: "md",
    },
    {
      id: "sf-4",
      filename: "todo-list.txt",
      size: 1024,
      extension: "txt",
    },
    {
      id: "sf-5",
      filename: "architecture.drawio",
      size: 3145728,
      extension: "drawio",
    },
    {
      id: "sf-6",
      filename: "todo-list1.txt",
      size: 1024,
      extension: "txt",
    },
    {
      id: "sf-7",
      filename: "architecture1.drawio",
      size: 3145728,
      extension: "drawio",
    },
  ];

  // Demo plan steps — TODO: 接入真实计划数据
  const planSteps: PlanStep[] = [
    { id: "1", description: "分析任务需求并制定执行计划", status: "completed" },
    { id: "2", description: "搜索相关资料和参考文档", status: "completed" },
    { id: "3", description: "编写核心代码实现", status: "in_progress" },
    { id: "4", description: "测试并验证结果", status: "pending" },
  ];

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
          <SessionHeader
            title="新任务"
            files={sessionFiles}
            rightSlot={<ComputerSettings />}
          />
          {/* 消息列表 */}
          <div className="flex-1 overflow-y-auto p-4">
            <div className="mx-auto max-w-3xl">
              {messages.map((item) => (
                <ChatMessage key={item.id} item={item} />
              ))}
            </div>
          </div>
          {/* 计划面板 — 固定在输入框上方 */}
          <div className="shrink-0 px-4 pb-2 chat-column-inset">
            <PlanPanel steps={planSteps} />
          </div>
          <div className="shrink-0 p-4 pt-2 chat-column-inset">
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
