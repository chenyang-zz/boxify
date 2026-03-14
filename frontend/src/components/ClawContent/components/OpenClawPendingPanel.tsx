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
import {
  Download,
  Loader2,
  Pause,
  Play,
  RefreshCw,
  Square,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { Task } from "@wails/claw/taskman/models";

interface OpenClawPendingPanelProps {
  installed: boolean;
  configured: boolean;
  nodeInstalled: boolean;
  nodeVersionSatisfied: boolean;
  npmInstalled: boolean;
  checking: boolean;
  starting: boolean;
  installing: boolean;
  gatewayRunning: boolean;
  binaryPath: string;
  configPath: string;
  nodePath: string;
  nodeVersion: string;
  npmPath: string;
  autoInstallSupported: boolean;
  autoInstallHint: string;
  installTask: Task | null;
  onRefresh: () => void;
  onStartGateway: () => void;
  onStartAutoInstall: () => void;
  onPauseAutoInstall: () => void;
  onResumeAutoInstall: () => void;
  onCancelAutoInstall: () => void;
}

interface InstallStageProgress {
  label: string;
  progress: number;
  message: string;
  active: boolean;
  completed: boolean;
}

/**
 * 将任务阶段状态映射为前端可渲染的进度数据。
 */
function buildStageProgress(
  label: string,
  progress: number | undefined,
  message: string | undefined,
  active: boolean,
): InstallStageProgress {
  const normalizedProgress = Math.max(0, Math.min(100, progress ?? 0));
  return {
    label,
    progress: normalizedProgress,
    message: message?.trim() || "等待开始",
    active,
    completed: normalizedProgress >= 100,
  };
}

/**
 * 渲染单个安装阶段的进度条。
 */
function InstallProgressBar({
  label,
  progress,
  message,
  active,
  completed,
}: InstallStageProgress) {
  return (
    <div className="rounded-lg border border-border/70 bg-background/70 p-3">
      <div className="flex items-center justify-between gap-3 text-xs">
        <span className="font-medium text-foreground">{label}</span>
        <span
          className={cn(
            "tabular-nums",
            completed
              ? "text-emerald-600"
              : active
                ? "text-foreground"
                : "text-muted-foreground",
          )}
        >
          {progress}%
        </span>
      </div>
      <div className="mt-2 h-2 overflow-hidden rounded-full bg-muted">
        <div
          className={cn(
            "h-full rounded-full transition-[width] duration-300",
            completed
              ? "bg-emerald-500"
              : active
                ? "bg-primary"
                : "bg-muted-foreground/40",
          )}
          style={{ width: `${progress}%` }}
        />
      </div>
      <p className="mt-2 text-xs text-muted-foreground">{message}</p>
    </div>
  );
}

/**
 * OpenClaw 待安装/待初始化面板。
 */
export const OpenClawPendingPanel: FC<OpenClawPendingPanelProps> = ({
  installed,
  configured,
  nodeInstalled,
  nodeVersionSatisfied,
  npmInstalled,
  checking,
  starting,
  installing,
  gatewayRunning,
  binaryPath,
  configPath,
  nodePath,
  nodeVersion,
  npmPath,
  autoInstallSupported,
  autoInstallHint,
  installTask,
  onRefresh,
  onStartGateway,
  onStartAutoInstall,
  onPauseAutoInstall,
  onResumeAutoInstall,
  onCancelAutoInstall,
}) => {
  const showGatewayStart = installed && configured && !gatewayRunning;
  const showAutoInstall = !showGatewayStart;
  const nodeStatus = !nodeInstalled
    ? "未安装"
    : nodeVersionSatisfied
      ? "已安装（符合要求）"
      : "已安装（版本不符合要求）";
  const showNodeVersionWarning = nodeInstalled && !nodeVersionSatisfied;
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
  const taskLog = installTask?.log?.[installTask.log.length - 1] ?? "";
  const taskPaused = Boolean(installTask?.paused);
  const taskControllable =
    installTask?.status === "running" || installTask?.status === "paused";
  const nodeStage = buildStageProgress(
    "Node.js 安装",
    installTask?.nodeProgress,
    installTask?.nodeMessage,
    installTask?.stage === "node" || (installTask?.nodeProgress ?? 0) > 0,
  );
  const openClawStage = buildStageProgress(
    "OpenClaw 安装",
    installTask?.openClawProgress,
    installTask?.openClawMessage,
    installTask?.stage === "openclaw" ||
      installTask?.stage === "done" ||
      (installTask?.openClawProgress ?? 0) > 0,
  );

  return (
    <div className="h-full w-full p-6">
      <div className="h-full rounded-lg">
        <div className="mx-auto flex h-full max-w-2xl flex-col items-center justify-center text-center">
          <h3 className="mt-5 text-lg font-semibold ">{title}</h3>
          <p className="mt-2 text-sm text-secondary-foreground">
            {description}
          </p>
          <div className="mt-4 w-full rounded-lg bg-card p-3 text-left text-xs leading-5 text-card-foreground">
            <p>Node 环境：{nodeStatus}</p>
            {nodePath && <p className="mt-1">Node 路径：{nodePath}</p>}
            {nodeVersion && <p className="mt-1">Node 版本：{nodeVersion}</p>}
            <p className="mt-1">
              npm 环境：{npmInstalled ? "已安装" : "未安装"}
            </p>
            {npmPath && <p className="mt-1">npm 路径：{npmPath}</p>}
            <p className="mt-1">Node 要求：&gt;= 22.16.0，推荐 24.x</p>
            {showNodeVersionWarning && (
              <p className="mt-2 text-amber-600">
                当前 Node 版本不满足 OpenClaw 要求，将自动升级或切换到合规版本。
              </p>
            )}
            <p>安装状态：{installStatus}</p>
            <p className="mt-1">
              网关状态：{gatewayRunning ? "运行中" : "未启动"}
            </p>
            {binaryPath && <p className="mt-1">可执行文件：{binaryPath}</p>}
            <p className="mt-1">配置文件：{configPath}</p>
            {!configured && (
              <p className="mt-2">
                推荐命令：
                <code>
                  node -v && npm i -g openclaw@latest && openclaw init
                </code>
              </p>
            )}
            {installTask && (
              <div className="mt-3 space-y-3">
                <div className="flex items-center justify-between gap-3 text-xs">
                  <span className="font-medium text-foreground">
                    自动安装总进度
                  </span>
                  <span className="tabular-nums text-muted-foreground">
                    {installTask.progress ?? 0}%
                  </span>
                </div>
                <div className="h-2 overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full rounded-full bg-primary transition-[width] duration-300"
                    style={{
                      width: `${Math.max(0, Math.min(100, installTask.progress ?? 0))}%`,
                    }}
                  />
                </div>
                <div className="grid gap-3 md:grid-cols-2">
                  <InstallProgressBar {...nodeStage} />
                  <InstallProgressBar {...openClawStage} />
                </div>
              </div>
            )}
            {taskLog && (
              <p className="mt-2 break-all text-muted-foreground">
                当前任务：{taskLog}
              </p>
            )}
            {!autoInstallSupported && autoInstallHint && (
              <p className="mt-2 text-amber-600">{autoInstallHint}</p>
            )}
          </div>
          <div className="mt-4 flex items-center gap-3">
            {showAutoInstall && (
              <Button
                onClick={onStartAutoInstall}
                disabled={
                  checking ||
                  starting ||
                  installing ||
                  taskControllable ||
                  !autoInstallSupported
                }
              >
                {installing ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Download className="size-4" />
                )}
                自动安装/修复 OpenClaw 环境
              </Button>
            )}
            {taskControllable && !taskPaused && (
              <Button
                onClick={onPauseAutoInstall}
                disabled={checking || starting}
                variant="outline"
              >
                <Pause className="size-4" />
                暂停下载
              </Button>
            )}
            {taskControllable && taskPaused && (
              <Button
                onClick={onResumeAutoInstall}
                disabled={checking || starting}
                variant="outline"
              >
                <Play className="size-4" />
                继续下载
              </Button>
            )}
            {taskControllable && (
              <Button
                onClick={onCancelAutoInstall}
                disabled={checking || starting}
                variant="outline"
              >
                <Square className="size-4" />
                取消下载
              </Button>
            )}
            {showGatewayStart && (
              <Button
                onClick={onStartGateway}
                disabled={checking || starting || installing}
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
              disabled={checking || starting || installing}
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
