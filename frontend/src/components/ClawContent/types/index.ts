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
 * 通道卡片数据类型
 */
export interface ChannelCard {
  id: string;
  name: string;
  type: "built-in" | "plugin";
  status: "enabled" | "disabled";
  managedBy: string;
  icon?: React.ComponentType<{ className?: string }>;
}

/**
 * 系统概览数据类型
 */
export interface SystemOverview {
  systemStatus: "normal" | "warning" | "error";
  activeChannels: number;
  aiModel: string;
  uptime: string;
  memoryUsage: string;
  todayMessages: number;
  channels: ChannelCard[];
}
