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
  Braces,
  Clock3,
  Expand,
  RefreshCw,
  Sparkles,
  CornerDownLeft,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Separator } from "@/components/ui/separator";
import { PanelHeader } from "./PanelHeader";

interface HandlerSummary {
  name: string;
  count: number;
  description: string;
}

/**
 * 聊天面板演示数据
 * 用于展示首屏示例消息中的 Handler 统计表格
 */
const handlerSummaryList: HandlerSummary[] = [
  { name: "插件管理", count: 1, description: "插件生命周期管理" },
  { name: "进程管理", count: 1, description: "OpenClaw 启停控制" },
  { name: "会话管理", count: 1, description: "会话列表与详情" },
  { name: "技能管理", count: 1, description: "技能启用、Cron 任务" },
  { name: "软件环境", count: 1, description: "软件检测与一键安装" },
  { name: "系统状态", count: 1, description: "状态总览、系统环境" },
  { name: "系统操作", count: 1, description: "版本、备份、更新" },
  { name: "工作区", count: 1, description: "文件管理、自动清理" },
];

/**
 * 聊天消息内容区
 * 渲染示例回答卡片及结构化表格
 */
const ChatMessageContent: FC = () => {
  return (
    <div className="max-w-5xl rounded-xl border bg-card p-4 shadow-xs">
      <div className="overflow-x-auto rounded-md border">
        <table className="w-full border-collapse text-sm">
          <tbody>
            {handlerSummaryList.map((item) => (
              <tr key={item.name} className="not-last:border-b">
                <td className="w-28 border-r px-3 py-2 font-medium">
                  {item.name}
                </td>
                <td className="w-16 border-r px-3 py-2">{item.count}</td>
                <td className="px-3 py-2 text-muted-foreground">
                  {item.description}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="mt-3 text-sm leading-7 text-foreground/90">
        <p className="font-semibold">
          总计 19 个 Handler 文件，文档中包含了每个 Handler 的：
        </p>
        <ul className="mt-1 list-disc pl-5 text-muted-foreground">
          <li>API 路径和方法</li>
          <li>功能说明</li>
          <li>关键逻辑</li>
          <li>设计模式</li>
          <li>安全措施</li>
        </ul>
      </div>
      <div className="mt-4 text-xs text-muted-foreground">波波 21:01</div>
    </div>
  );
};

/**
 * 聊天面板组件
 * 根据设计稿渲染聊天页布局（标题、工具栏、消息区、输入区）
 */
export const ChatPanel: FC = () => {
  return (
    <div className="flex h-full w-full min-h-0 flex-col p-6">
      <PanelHeader
        className="mb-6"
        title="聊天"
        description="用于快速干预的直接网关聊天会话。"
        align="start"
        actions={
          <div className="flex flex-wrap items-center justify-end gap-2">
            <Button
              variant="outline"
              size="sm"
              className="h-10 min-w-40 justify-start text-sm text-muted-foreground"
            >
              Main Session
            </Button>
            <Button variant="outline" size="icon-sm" aria-label="刷新会话">
              <RefreshCw className="size-4" />
            </Button>
            <Separator orientation="vertical" className="mx-1 h-6" />
            <Button variant="outline" size="icon-sm" aria-label="切换代码模式">
              <Braces className="size-4" />
            </Button>
            <Button variant="outline" size="icon-sm" aria-label="切换全屏">
              <Expand className="size-4" />
            </Button>
            <Button
              variant="outline"
              size="icon-sm"
              className="border-destructive/80 text-destructive hover:bg-destructive/10 hover:text-destructive"
              aria-label="查看历史"
            >
              <Clock3 className="size-4" />
            </Button>
          </div>
        }
      />

      <div className="flex min-h-0 flex-1 flex-col p-2">
        <div className="flex-1 overflow-auto p-2">
          <div className="mb-4 flex items-start gap-3">
            <div className="mt-1 rounded-full bg-primary/10 p-1.5 text-primary">
              <Sparkles className="size-3.5" />
            </div>
            <ChatMessageContent />
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
        <Textarea
          placeholder="Message (↩ to send, Shift+↩ for line breaks, paste images)"
          className="min-h-16 resize-none bg-background"
        />
        <div className="flex items-center gap-3 sm:shrink-0">
          <Button variant="outline" className="h-10 px-5">
            New session
          </Button>
          <Button className="h-10 px-6">
            Send
            <CornerDownLeft className="size-3.5" />
          </Button>
        </div>
      </div>
    </div>
  );
};

export default ChatPanel;
