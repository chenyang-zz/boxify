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

import { Events } from "@wailsio/runtime";
import { useEffect } from "react";

export interface WindowInfo {
  id: number;
  name: string;
  title: string;
}

export const useWindowListener = () => {
  useEffect(() => {
    // 监听窗口打开事件（用于动态更新）
    const unbindOpened = Events.On(
      "window:opened",
      (event: { data: Record<string, unknown> }) => {
        console.log("🪟 窗口打开事件:", event.data);
      },
    );

    const unbindClosed = Events.On(
      "window:closed",
      (event: { data: Record<string, unknown> }) => {
        console.log("🪟 窗口关闭事件:", event.data);
      },
    );

    return () => {
      unbindOpened();
      unbindClosed();
    };
  }, []);
};
