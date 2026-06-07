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

import { FC } from "react";
import { ComputerSettings } from "./components/ComputerSettings";

/**
 * ComputerContent 核心渲染组件
 */
export const ComputerContentCore: FC = () => {
  return (
    <div className="h-full w-full flex flex-col">
      {/* 顶部工具栏 */}
      <div className="flex items-center justify-end p-2 border-b border-border">
        <ComputerSettings />
      </div>

      {/* 内容区域 — 测试入口 */}
      <div className="flex-1 flex flex-col items-center justify-center gap-4 p-4">
        <p className="text-sm text-muted-foreground">ComputerContent 内容区域</p>
        {/* TODO: 移除测试入口 */}
      </div>
    </div>
  );
};
