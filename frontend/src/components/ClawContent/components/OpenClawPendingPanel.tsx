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
import { Loader2, Play, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

interface OpenClawPendingPanelProps {
  installed: boolean;
  configured: boolean;
  checking: boolean;
  starting: boolean;
  gatewayRunning: boolean;
  binaryPath: string;
  configPath: string;
  onRefresh: () => void;
  onStartGateway: () => void;
}

/**
 * OpenClaw 待安装/待初始化面板。
 */
export const OpenClawPendingPanel: FC<OpenClawPendingPanelProps> = ({
  installed,
  configured,
  checking,
  starting,
  gatewayRunning,
  binaryPath,
  configPath,
  onRefresh,
  onStartGateway,
}) => {
  const showGatewayStart = installed && configured && !gatewayRunning;
  const title = showGatewayStart
    ? "OpenClaw 网关未启动"
    : "OpenClaw 待安装或待初始化";
  const description = showGatewayStart
    ? "当前已检测到 OpenClaw 配置，但网关尚未启动，请先开启网管后再进入聊天。"
    : "当前未检测到可用 OpenClaw 配置，请先安装并执行初始化后再继续使用";
  const installStatus = !installed
    ? "未安装"
    : configured
      ? "已安装并已初始化"
      : "已安装（待初始化）";

  return (
    <div className="h-full w-full p-6">
      <div className="h-full rounded-lg">
        <div className="mx-auto flex h-full max-w-2xl flex-col items-center justify-center text-center">
          <h3 className="mt-5 text-lg font-semibold ">{title}</h3>
          <p className="mt-2 text-sm text-secondary-foreground">
            {description}
          </p>
          <div className="mt-4 w-full rounded-lg bg-card p-3 text-left text-xs text-card-foreground">
            <p>安装状态：{installStatus}</p>
            <p className="mt-1">网关状态：{gatewayRunning ? "运行中" : "未启动"}</p>
            {binaryPath && <p className="mt-1">可执行文件：{binaryPath}</p>}
            <p className="mt-1">配置文件：{configPath}</p>
            {!configured && (
              <p className="mt-2">
                推荐命令：
                <code>npm i -g openclaw@latest && openclaw init</code>
              </p>
            )}
          </div>
          <div className="mt-4 flex items-center gap-3">
            {showGatewayStart && (
              <Button
                onClick={onStartGateway}
                disabled={checking || starting}
              >
                {starting ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Play className="size-4" />
                )}
                开启网管
              </Button>
            )}
            <Button
              onClick={onRefresh}
              disabled={checking || starting}
              variant={showGatewayStart ? "outline" : "default"}
            >
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
    </div>
  );
};

export default OpenClawPendingPanel;
