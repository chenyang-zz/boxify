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

import { create } from "zustand";
import { callWails } from "@/lib/utils";
import { ClawService } from "@wails/service";
import type { SystemOverview } from "../types";
import { mockSystemOverview } from "../domain";
import type { ClawOverviewData } from "@wails/types/models";

/**
 * ClawContent Store 状态接口
 */
interface ClawContentStore {
  // 状态
  overview: SystemOverview;
  isLoading: boolean;

  // Actions
  setOverview: (overview: SystemOverview) => void;
  refreshOverview: () => Promise<void>;
}

/**
 * 将后端概览数据转换为前端展示结构。
 */
function mapOverviewDataToState(data: ClawOverviewData): SystemOverview {
  return {
    systemStatus:
      data.systemStatus === "error" || data.systemStatus === "warning"
        ? data.systemStatus
        : "normal",
    activeChannels: data.activeChannels ?? 0,
    aiModel: data.aiModel || "-",
    uptime: data.uptime || "-",
    memoryUsage: data.memoryUsage || "-",
    todayMessages: data.todayMessages ?? 0,
    channels: (data.channels ?? []).map((channel) => ({
      id: channel.id || "",
      name: channel.name || channel.id || "-",
      type: channel.type === "plugin" ? "plugin" : "built-in",
      status: channel.status === "disabled" ? "disabled" : "enabled",
      managedBy: channel.managedBy || "由网关管理",
    })),
  };
}

/**
 * ClawContent 全局状态 Store
 */
export const useClawContentStore = create<ClawContentStore>((set) => ({
  overview: mockSystemOverview,
  isLoading: false,

  setOverview: (overview: SystemOverview) => {
    set({ overview });
  },

  refreshOverview: async () => {
    set({ isLoading: true });
    try {
      const result = await callWails(ClawService.GetOverview);
      if (!result?.data) {
        return;
      }
      set({ overview: mapOverviewDataToState(result.data) });
    } finally {
      set({ isLoading: false });
    }
  },
}));
