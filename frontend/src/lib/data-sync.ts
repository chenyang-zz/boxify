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

import { DataChannel } from "@/store/data-sync.store";
import { dataSyncStoreMethods } from "@/store/data-sync.store";

/**
 * 数据同步 API 封装
 */
export class DataSyncAPI {
  /**
   * 广播配置更新
   */
  static broadcastConfig(dataType: string, data: Record<string, unknown>) {
    dataSyncStoreMethods.broadcast(DataChannel.Config, dataType, data);
  }

  /**
   * 广播连接更新
   */
  static broadcastDatabase(dataType: string, data: Record<string, unknown>) {
    dataSyncStoreMethods.broadcast(DataChannel.Connection, dataType, data);
  }

  /**
   * 广播设置更新
   */
  static broadcastSettings(dataType: string, data: Record<string, unknown>) {
    dataSyncStoreMethods.broadcast(DataChannel.Settings, dataType, data);
  }

  /**
   * 发送到指定窗口
   */
  static sendToWindow(
    targetWindow: string,
    channel: DataChannel,
    dataType: string,
    data: Record<string, unknown>,
  ) {
    dataSyncStoreMethods.sendTo(targetWindow, channel, dataType, data);
  }
}

// 便捷函数
export const syncPropertyUpdate = (properties: unknown[]) => {
  DataSyncAPI.broadcastConfig("property:update", { properties });
};

export const syncDatabaseConfig = (config: unknown) => {
  DataSyncAPI.broadcastDatabase("connection:updated", { config });
};

export const syncSettingsUpdate = (settings: Record<string, unknown>) => {
  DataSyncAPI.broadcastSettings("settings:update", settings);
};
