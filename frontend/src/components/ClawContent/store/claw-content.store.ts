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
import type { SystemOverview } from "../types";
import { mockSystemOverview } from "../domain";

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
      // TODO: 调用后端 API 获取数据
      // const data = await fetchOverview();
      // set({ overview: data });
    } finally {
      set({ isLoading: false });
    }
  },
}));
