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

import type { ClawOverviewChannel, ClawOverviewData } from "@wails/types/models";

/**
 * 状态卡片数据类型
 */
export interface StatCard {
  id: string;
  label: string;
  value: string;
  icon: React.ComponentType<{ className?: string }>;
  iconBgColor: string;
  iconColor: string;
}

/**
 * 概览通道卡片类型（复用后端类型定义）。
 */
export type ChannelCard = ClawOverviewChannel;

/**
 * 系统概览数据类型（复用后端类型定义）。
 */
export type SystemOverview = ClawOverviewData;

/**
 * 频道管理面板的通道状态。
 */
export type ManagedChannelStatus = "enabled" | "configured" | "unconfigured";

/**
 * 频道管理面板通道类型：复用后端基础字段并扩展前端视图字段。
 */
export interface ManagedChannel extends Omit<ClawOverviewChannel, "status"> {
  description?: string;
  icon: React.ReactNode;
  status: ManagedChannelStatus;
  config?: Record<string, unknown>;
  saveTargetId?: string;
  feishuVariant?: "official" | "clawteam";
}

export {
  DEFAULT_AGENT_ID,
  PENDING_ASSISTANT_ID,
  type PendingAssistantDraft,
  type RenderedChatMessage,
} from "./chat-panel";
