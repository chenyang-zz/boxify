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

import { FC, useState, useCallback } from "react";
import {
  Plus,
  CircuitBoard,
  Loader2,
  MoreHorizontal,
  Trash,
  Monitor,
} from "lucide-react";
import { cn, formatRelativeDate } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { Kbd } from "@/components/ui/kbd";

/**
 * 会话数据接口
 */
interface Session {
  session_id: string;
  title: string;
  latest_message: string;
  latest_message_at: string;
  status: "running" | "waiting" | "completed";
}

/**
 * 模拟会话数据
 */
const mockSessions: Session[] = [
  {
    session_id: "1",
    title: "代码审查任务",
    latest_message: "正在分析代码结构...",
    latest_message_at: "2026-06-07T10:30:00Z",
    status: "running",
  },
  {
    session_id: "2",
    title: "文档生成",
    latest_message: "已生成 API 文档",
    latest_message_at: "2026-06-07T09:15:00Z",
    status: "completed",
  },
  {
    session_id: "3",
    title: "数据迁移",
    latest_message: "等待用户确认",
    latest_message_at: "2026-06-06T18:20:00Z",
    status: "waiting",
  },
  {
    session_id: "4",
    title: "测试用例编写",
    latest_message: "已完成 80%",
    latest_message_at: "2026-06-06T14:00:00Z",
    status: "running",
  },
];

/**
 * 单个会话列表项
 */
interface SessionItemProps {
  session: Session;
  isActive: boolean;
  onClick: (sessionId: string) => void;
  onDelete: (session: Session) => void;
}

const SessionItem: FC<SessionItemProps> = ({
  session,
  isActive,
  onClick,
  onDelete,
}) => {
  const handleClick = useCallback(() => {
    onClick(session.session_id);
  }, [onClick, session.session_id]);

  const handleDelete = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      onDelete(session);
    },
    [onDelete, session],
  );

  const description = session.latest_message || "暂无消息";
  const dateLabel = formatRelativeDate(session.latest_message_at);
  const isRunning =
    session.status === "running" || session.status === "waiting";

  return (
    <div
      className={cn(
        "group/item flex items-center gap-2 p-2 rounded-md cursor-pointer transition-colors",
        isActive ? "bg-card shadow-sm" : "hover:bg-accent/50",
      )}
      onClick={handleClick}
    >
      {/* 左侧图标 */}
      <div className="flex shrink-0 items-center justify-center size-8 rounded-sm bg-muted">
        {isRunning ? (
          <Loader2 className="size-4 animate-spin text-primary" />
        ) : (
          <CircuitBoard className="size-4 text-muted-foreground" />
        )}
      </div>

      {/* 中间内容 */}
      <div className="flex-1 flex flex-col gap-0 min-w-0">
        <p className="text-sm font-medium truncate text-left">
          {session.title || "新任务"}
        </p>
        <p className="text-xs text-muted-foreground truncate text-left">
          {description}
        </p>
      </div>

      {/* 右侧操作区 */}
      <div className="flex flex-col items-end gap-0 self-start pt-0.5">
        <span className="text-xs text-muted-foreground whitespace-nowrap">
          {dateLabel}
        </span>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              size="icon-xs"
              variant="ghost"
              className="cursor-pointer opacity-0 group-hover/item:opacity-100 transition-opacity"
              onClick={(e) => e.stopPropagation()}
            >
              <MoreHorizontal className="size-3" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="center" side="bottom">
            <DropdownMenuItem
              variant="destructive"
              className="cursor-pointer"
              onClick={handleDelete}
            >
              <Trash className="size-4" />
              删除
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
};

/**
 * Computer 视图组件
 * 参照 left-panel 样式实现，包含新建任务按钮和会话列表
 */
export const ComputerView: FC = () => {
  const [activeSessionId, setActiveSessionId] = useState<string>("1");

  const handleSessionClick = useCallback((sessionId: string) => {
    setActiveSessionId(sessionId);
  }, []);

  const handleDeleteSession = useCallback((session: Session) => {
    // eslint-disable-next-line no-console
    console.log("删除会话:", session.title);
  }, []);

  const handleNewTask = useCallback(() => {
    // eslint-disable-next-line no-console
    console.log("新建任务");
  }, []);

  return (
    <div className="h-full w-full flex flex-col">
      {/* 内容区域 */}
      <div className="flex-1 flex flex-col gap-2 p-2 overflow-auto">
        {/* 新建任务按钮 */}
        <Button
          variant="outline"
          className="w-full cursor-pointer justify-center"
          onClick={handleNewTask}
        >
          <span className="flex items-center gap-2">
            <Plus className="size-4" />
            新建任务
          </span>
          <span className="flex items-center gap-0.5 ml-2">
            <Kbd>⌘</Kbd>
            <Kbd>K</Kbd>
          </span>
        </Button>

        {/* 会话列表 */}
        <div className="flex flex-col gap-1">
          {mockSessions.map((session) => (
            <SessionItem
              key={session.session_id}
              session={session}
              isActive={session.session_id === activeSessionId}
              onClick={handleSessionClick}
              onDelete={handleDeleteSession}
            />
          ))}
        </div>
      </div>

      {/* 底部状态栏 */}
      <div className="flex items-center justify-center gap-1.5 py-2 border-t border-border">
        <Monitor className="size-3.5 text-muted-foreground" />
        <span className="text-xs text-muted-foreground">Computer 已连接</span>
      </div>
    </div>
  );
};

export default ComputerView;
