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

import { useEffect, useRef } from "react";
import {
  useDataSyncStore,
  DataChannel,
  dataSyncStoreMethods,
} from "@/store/data-sync.store";
import { DataSyncEvent } from "@wails/service";
import { parseDefaultValueExpression } from "node_modules/kysely/dist/esm/parser/default-value-parser";

/**
 * 数据同步 Hook
 *
 * @param channel 监听的数据频道
 * @param callback 收到消息时的回调
 */
export const useDataSync = (
  channel: DataChannel,
  callback: (event: DataSyncEvent) => void,
) => {
  const { currentWindow } = useDataSyncStore();
  const callbackRef = useRef(callback);

  // 更新回调引用
  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  // 注册监听器
  useEffect(() => {
    const listener = (event: DataSyncEvent) => {
      // 忽略自己发送的消息
      if (event.source === currentWindow) return;
      callbackRef.current(event);
    };

    dataSyncStoreMethods.addListener(channel, listener);

    return () => {
      dataSyncStoreMethods.removeListener(channel);
    };
  }, [channel, currentWindow]);
};

/**
 * 数据发送 Hook
 *
 * @returns 广播和定向发送方法
 */
export const useDataSyncSend = () => {
  const { broadcast, sendTo, currentWindow } = useDataSyncStore();

  return {
    broadcast: (
      channel: DataChannel,
      dataType: string,
      data: Record<string, unknown>,
    ) => {
      broadcast(channel, dataType, data);
    },

    sendTo: (
      targetWindow: string,
      channel: DataChannel,
      dataType: string,
      data: Record<string, unknown>,
    ) => {
      sendTo(targetWindow, channel, dataType, data);
    },

    currentWindow,
  };
};
