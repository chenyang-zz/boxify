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
import { Brain, Loader2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

interface OpenClawPendingPanelProps {
  installed: boolean;
  checking: boolean;
  binaryPath: string;
  configPath: string;
  onRefresh: () => void;
}

/**
 * OpenClaw 待安装/待初始化面板。
 */
export const OpenClawPendingPanel: FC<OpenClawPendingPanelProps> = ({
  installed,
  checking,
  binaryPath,
  configPath,
  onRefresh,
}) => {
  return (
    <div className="h-full w-full p-6">
      <div className="h-full rounded-lg">
        <div className="mx-auto flex h-full max-w-2xl flex-col items-center justify-center text-center">
          <h3 className="mt-5 text-lg font-semibold ">
            OpenClaw 待安装或待初始化
          </h3>
          <p className="mt-2 text-sm text-secondary-foreground">
            当前未检测到可用 OpenClaw 配置，请先安装并执行初始化后再继续使用
          </p>
          <div className="mt-4 w-full rounded-lg bg-card p-3 text-left text-xs text-card-foreground">
            <p>安装状态：{installed ? "已安装（待初始化）" : "未安装"}</p>
            {binaryPath && <p className="mt-1">可执行文件：{binaryPath}</p>}
            <p className="mt-1">配置文件：{configPath}</p>
            <p className="mt-2">
              推荐命令：<code>npm i -g openclaw@latest && openclaw init</code>
            </p>
          </div>
          <Button onClick={onRefresh} disabled={checking} className="mt-4 ">
            {checking ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <RefreshCw className="size-4" />
            )}
            重新检查
          </Button>
        </div>
      </div>
    </div>
  );
};

export default OpenClawPendingPanel;
