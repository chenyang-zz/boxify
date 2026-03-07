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

import { Link2, Bot, Clock, MemoryStick, MessageSquare } from "lucide-react";
import type { StatCard, ChannelCard, SystemOverview } from "../types";

/**
 * 状态卡片配置数据
 */
export const statCardsConfig: Omit<StatCard, "value">[] = [
  {
    id: "activeChannels",
    label: "活跃频道",
    icon: Link2,
    iconBgColor: "bg-emerald-500/20",
    iconColor: "text-emerald-500",
  },
  {
    id: "aiModel",
    label: "AI 模型",
    icon: Bot,
    iconBgColor: "bg-violet-500/20",
    iconColor: "text-violet-500",
  },
  {
    id: "uptime",
    label: "运行时间",
    icon: Clock,
    iconBgColor: "bg-blue-500/20",
    iconColor: "text-blue-500",
  },
  {
    id: "memoryUsage",
    label: "内存占用",
    icon: MemoryStick,
    iconBgColor: "bg-cyan-500/20",
    iconColor: "text-cyan-500",
  },
  {
    id: "todayMessages",
    label: "今日消息",
    icon: MessageSquare,
    iconBgColor: "bg-orange-500/20",
    iconColor: "text-orange-500",
  },
];

/**
 * 模拟的通道数据
 */
export const mockChannels: ChannelCard[] = [
  {
    id: "lark",
    name: "飞书 / Lark",
    type: "built-in",
    status: "enabled",
    managedBy: "由网关管理",
  },
  {
    id: "acpx",
    name: "acpx",
    type: "plugin",
    status: "enabled",
    managedBy: "由网关管理",
  },
];

/**
 * 模拟的系统概览数据
 */
export const mockSystemOverview: SystemOverview = {
  systemStatus: "normal",
  activeChannels: 2,
  aiModel: "zai/glm-5",
  uptime: "1 时",
  memoryUsage: "23 MB",
  todayMessages: 1,
  channels: mockChannels,
};

/**
 * 获取状态卡片的值
 */
export function getStatCardValue(
  cardId: string,
  overview: SystemOverview,
): string {
  switch (cardId) {
    case "activeChannels":
      return `${overview.activeChannels} 个`;
    case "aiModel":
      return overview.aiModel;
    case "uptime":
      return overview.uptime;
    case "memoryUsage":
      return overview.memoryUsage;
    case "todayMessages":
      return `${overview.todayMessages} 条`;
    default:
      return "-";
  }
}

/**
 * 获取通道类型标签文本
 */
export function getChannelTypeLabel(type: ChannelCard["type"]): string {
  return type === "built-in" ? "内置通道" : "插件通道";
}

/**
 * 获取状态标签配置
 */
export function getStatusBadgeConfig(status: ChannelCard["status"]) {
  switch (status) {
    case "enabled":
      return {
        text: "已启用",
        className: "bg-blue-500/10 text-blue-500",
      };
    case "disabled":
      return {
        text: "已禁用",
        className: "bg-gray-500/10 text-gray-500",
      };
  }
}
