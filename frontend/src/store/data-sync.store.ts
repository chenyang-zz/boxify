// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with License.
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
import { persist } from "zustand/middleware";
import { Events } from "@wailsio/runtime";
import { StoreMethods } from "./common";
import { DataSyncEvent } from "@wails/service";

// 数据频道
export const DataChannel = {
  Config: "config",
  Connection: "connection",
  Settings: "settings",
  Custom: "custom",
} as const;

export type DataChannel = (typeof DataChannel)[keyof typeof DataChannel];

// 监听器清理函数
type ListenerCleanup = {
  unbind: () => void;
};

interface DataSyncState {
  currentWindow: string;
  setCurrentWindow: (name: string) => void;

  // 监听器管理
  listeners: Map<string, ListenerCleanup>;
  addListener: (
    channel: DataChannel,
    listener: (event: DataSyncEvent) => void,
  ) => void;
  removeListener: (channel: DataChannel) => void;

  // 发送数据
  broadcast: (
    channel: DataChannel,
    dataType: string,
    data: Record<string, unknown>,
  ) => void;
  sendTo: (
    targetWindow: string,
    channel: DataChannel,
    dataType: string,
    data: Record<string, unknown>,
  ) => void;
}

export const useDataSyncStore = create<DataSyncState>((set, get) => ({
  currentWindow: "",
  setCurrentWindow: (name: string) => set({ currentWindow: name }),

  listeners: new Map(),

  addListener: (channel, listener) => {
    const state = get();
    const key = `channel:${channel}`;

    // 避免重复监听
    if (state.listeners.has(key)) {
      console.warn(`[DataSync] 通道 ${channel} 已有监听器`);
      return;
    }

    // 监听广播事件
    const unbindBroadcast = Events.On(
      "data-sync:broadcast",
      (event: { data: DataSyncEvent }) => {
        if (event.data.channel === channel) {
          listener(event.data);
        }
      },
    );

    // 监听定向事件
    const unbindTargeted = Events.On(
      "data-sync:targeted",
      (event: { data: DataSyncEvent }) => {
        if (
          event.data.target === state.currentWindow &&
          event.data.channel === channel
        ) {
          listener(event.data);
        }
      },
    );

    // 保存监听器（包含清理函数）
    state.listeners.set(key, {
      unbind: () => {
        unbindBroadcast();
        unbindTargeted();
      },
    });

    console.log(`[DataSync] 已监听频道: ${channel}`);
  },

  removeListener: (channel) => {
    const state = get();
    const key = `channel:${channel}`;
    const listener = state.listeners.get(key);

    if (listener) {
      listener.unbind();
      state.listeners.delete(key);
      console.log(`[DataSync] 已移除频道监听: ${channel}`);
    }
  },

  broadcast: (channel, dataType, data) => {
    const state = get();
    const event: DataSyncEvent = {
      source: state.currentWindow,
      target: "",
      channel,
      dataType,
      data,
      timestamp: Date.now(),
      id: generateID(),
    };

    Events.Emit("data-sync:broadcast", event);
    console.log(`[DataSync] 广播消息`, { channel, dataType, data });
  },

  sendTo: (targetWindow, channel, dataType, data) => {
    const state = get();
    const event: DataSyncEvent = {
      source: state.currentWindow,
      target: targetWindow,
      channel,
      dataType,
      data,
      timestamp: Date.now(),
      id: generateID(),
    };

    Events.Emit("data-sync:targeted", event);
    console.log(`[DataSync] 发送定向消息`, {
      targetWindow,
      channel,
      dataType,
      data,
    });
  },
}));

// 辅助函数：生成唯一ID
function generateID(): string {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

// 导出非 React 环境调用的方法
export const dataSyncStoreMethods = StoreMethods(useDataSyncStore);
