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
import {
  Archive,
  ChevronRight,
  Folder,
  MoreHorizontal,
  Pencil,
  Pin,
  Search,
  SquarePen,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface FolderItem {
  id: string;
  title: string;
}

interface ConversationItem {
  id: string;
  title: string;
  time: string;
}

interface ProjectItem extends FolderItem {
  conversations: ConversationItem[];
}

const pinnedItems: FolderItem[] = [
  { id: "boxify-api", title: "boxify/api" },
  { id: "boxify", title: "Boxify" },
];

const dtInsightConversations: ConversationItem[] = [
  { id: "handoff-doc", title: "编写离职交接文档", time: "" },
  {
    id: "staged-summary",
    title: "看看我的 staged，帮我生成80...",
    time: "3 周",
  },
  { id: "tag-sync", title: "跨项目标签同步发布", time: "3 周" },
  { id: "pdf-export", title: "pdf导出问题修复", time: "1 个月" },
  {
    id: "staged-generate",
    title: "根据我staged的内容，生成...",
    time: "1 个月",
  },
  { id: "show-more", title: "展开显示", time: "" },
];

const projectItems: ProjectItem[] = [
  {
    id: "claude-code-src",
    title: "claude_code_src",
    conversations: [
      { id: "claude-migrate", title: "claude-code迁移", time: "2 周" },
      { id: "tui-output", title: "修复 TUI 流式输出问题", time: "1 个月" },
      {
        id: "migrate-go",
        title: "migrate claude code to golang",
        time: "1 个月",
      },
    ],
  },
  {
    id: "flowmind",
    title: "flowmind",
    conversations: [{ id: "flowmind-empty", title: "暂无对话", time: "" }],
  },
  {
    id: "feat-19604",
    title: "feat_19604",
    conversations: [
      {
        id: "ai-chat-index",
        title: "Add index selector to AI chat",
        time: "1 个月",
      },
      {
        id: "custom-tsx-sync",
        title: "同步更新 CustomTsx 数据",
        time: "1 个月",
      },
    ],
  },
  {
    id: "feat-19175",
    title: "feat_19175",
    conversations: [
      { id: "ai-flow", title: "分析 AI 交互数据流", time: "1 个月" },
    ],
  },
  {
    id: "dt-insight-studio",
    title: "dt-insight-studio",
    conversations: dtInsightConversations,
  },
];

const activeProjectId = "dt-insight-studio";
const initialExpandedProjectIds = [activeProjectId];

const conversations: ConversationItem[] = [
  { id: "pino-refresh", title: "修复 Node/pino 索引刷新失败", time: "3 天" },
  { id: "pet-shatch", title: "$hatch-pet 创建一个宠物基于我...", time: "4 天" },
  {
    id: "clash-connect",
    title: "为什么我使用clash时，只有连接...",
    time: "4 天",
  },
  {
    id: "install-pets",
    title: "Install this pet: npx codex-pets ...",
    time: "4 天",
  },
  { id: "rar", title: "转换压缩包为RAR", time: "1 周" },
  { id: "resume", title: "更新简历", time: "1 周" },
  {
    id: "brew-clean",
    title: "帮我清除一下使用“brew install b...",
    time: "2 周",
  },
  { id: "zsh-fish", title: "将当前zsh配置迁移到fish中}}", time: "1 个月" },
];

/**
 * ActionRow 渲染 Chat 侧栏顶部的快捷入口。
 */
const ActionRow: FC<{
  icon: React.ElementType;
  label: string;
}> = ({ icon: Icon, label }) => {
  return (
    <button
      type="button"
      className="flex h-8 w-full items-center gap-2 rounded-md px-2 text-left text-sm text-foreground transition-colors hover:bg-accent/60"
    >
      <Icon className="size-4 shrink-0 text-muted-foreground" />
      <span className="truncate">{label}</span>
    </button>
  );
};

/**
 * SectionTitle 渲染置顶、项目、对话等分组标题。
 */
const SectionTitle: FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <div className="px-2 pt-4 pb-1 text-xs font-medium text-muted-foreground">
      {children}
    </div>
  );
};

/**
 * FolderRow 渲染文件夹类条目。
 */
const FolderRow: FC<{
  item: FolderItem;
  active?: boolean;
  expanded?: boolean;
  onClick?: () => void;
}> = ({ item, active = false, expanded = false, onClick }) => {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-expanded={onClick ? expanded : undefined}
      className={cn(
        "group flex h-8 w-full min-w-0 items-center gap-2 rounded-md px-2 text-left text-sm transition-colors",
        active
          ? "bg-muted text-foreground"
          : "text-foreground hover:bg-accent/60",
      )}
    >
      <Folder className="size-4 shrink-0 text-muted-foreground" />
      <div className="flex flex-1 items-center gap-2">
        <span className="min-w-0 truncate">{item.title}</span>
        <ChevronRight
          className={cn(
            "size-3.5 shrink-0 text-muted-foreground opacity-0  transition duration-150 ease-out group-hover:opacity-100",
            expanded && "rotate-90",
          )}
        />
      </div>

      <span className="hidden shrink-0 items-center gap-2 text-muted-foreground group-hover:flex">
        <MoreHorizontal className="size-3.5" />
        <SquarePen className="size-3.5" />
      </span>
    </button>
  );
};

/**
 * ConversationRow 渲染单行会话记录。
 */
const ConversationRow: FC<{
  item: ConversationItem;
  inset?: boolean;
  muted?: boolean;
}> = ({ item, inset = false, muted = false }) => {
  return (
    <button
      type="button"
      className={cn(
        "group flex h-8 w-full min-w-0 items-center gap-2 rounded-md text-left text-sm transition-colors hover:bg-accent/60",
        inset ? "pl-8 pr-2" : "px-2",
        muted ? "text-muted-foreground" : "text-foreground",
      )}
    >
      <span className="min-w-0 flex-1 truncate">{item.title}</span>
      {item.time ? (
        <span className="shrink-0 text-xs text-muted-foreground group-hover:hidden">
          {item.time}
        </span>
      ) : null}
      {!muted ? (
        <span className="hidden shrink-0 items-center gap-2 text-muted-foreground group-hover:flex">
          <Pin className="size-3.5" />
          <Archive className="size-3.5" />
        </span>
      ) : null}
    </button>
  );
};

/**
 * ChatView 渲染 Chat tab 专属的分组式侧边栏。
 */
export const ChatView: FC = () => {
  const [expandedProjectIds, setExpandedProjectIds] = useState(
    () => new Set(initialExpandedProjectIds),
  );

  /**
   * toggleProject 切换项目的本地展开状态。
   */
  const toggleProject = (projectId: string) => {
    setExpandedProjectIds((current) => {
      const next = new Set(current);
      if (next.has(projectId)) {
        next.delete(projectId);
      } else {
        next.add(projectId);
      }
      return next;
    });
  };

  return (
    <div className="flex h-full w-full min-w-0 flex-col overflow-hidden bg-sidebar">
      <div className="shrink-0 px-2 pt-2 pb-1">
        <div className="flex flex-col gap-0.5">
          <ActionRow icon={SquarePen} label="新对话" />
          <ActionRow icon={Search} label="搜索" />
        </div>
      </div>

      <div className="relative min-h-0 flex-1">
        <div className="h-full overflow-auto px-2 pb-2">
          <SectionTitle>置顶</SectionTitle>
          <div className="flex flex-col gap-0.5">
            {pinnedItems.map((item) => (
              <FolderRow key={item.id} item={item} />
            ))}
          </div>

          <SectionTitle>项目</SectionTitle>
          <div className="flex flex-col gap-0.5">
            {projectItems.map((item) => {
              const expanded = expandedProjectIds.has(item.id);
              return (
                <div key={item.id} className="flex flex-col gap-0.5">
                  <FolderRow
                    item={item}
                    active={item.id === activeProjectId}
                    expanded={expanded}
                    onClick={() => toggleProject(item.id)}
                  />
                  {expanded
                    ? item.conversations.map((conversation) => (
                        <ConversationRow
                          key={conversation.id}
                          item={conversation}
                          inset
                          muted={
                            conversation.id.endsWith("-empty") ||
                            conversation.id === "show-more"
                          }
                        />
                      ))
                    : null}
                </div>
              );
            })}
          </div>

          <SectionTitle>对话</SectionTitle>
          <div className="flex flex-col gap-0.5">
            {conversations.map((item) => (
              <ConversationRow key={item.id} item={item} />
            ))}
          </div>
        </div>

        <div className="pointer-events-none absolute inset-x-0 top-0 h-6 bg-gradient-to-b from-sidebar to-transparent" />
        <div className="pointer-events-none absolute inset-x-0 bottom-0 h-8 bg-gradient-to-t from-sidebar to-transparent" />
      </div>
    </div>
  );
};

export default ChatView;
