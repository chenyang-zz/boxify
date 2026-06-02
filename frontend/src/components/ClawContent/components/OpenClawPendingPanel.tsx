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

import { FC, ReactNode, useEffect, useState } from "react";
import {
  Download,
  Loader2,
  Pause,
  Play,
  Rocket,
  RefreshCw,
  Square,
  Wrench,
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

interface SetupStep {
  index: number;
  title: string;
  description: string;
  status: "done" | "active" | "pending";
  progress?: number;
  detail: string;
}

interface StepBadgeProps {
  index: number;
  title: string;
  done: boolean;
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
 * 渲染 OpenClaw 环境搭建步骤卡片。
 */
function SetupStepCard({
  index,
  title,
  description,
  status,
  progress,
  detail,
}: SetupStep) {
  const statusText =
    status === "done" ? "已完成" : status === "active" ? "进行中" : "待处理";

  return (
    <div
      className={cn(
        "rounded-xl border p-4 text-left transition-colors",
        status === "done"
          ? "border-emerald-500/40 bg-emerald-500/5"
          : status === "active"
            ? "border-primary/40 bg-primary/5"
            : "border-border/70 bg-background/70",
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-3">
          <div
            className={cn(
              "flex size-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold",
              status === "done"
                ? "bg-emerald-500 text-white"
                : status === "active"
                  ? "bg-primary text-primary-foreground"
                  : "bg-muted text-muted-foreground",
            )}
          >
            {index}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h4 className="text-sm font-semibold text-foreground">{title}</h4>
              <span
                className={cn(
                  "rounded-full px-2 py-0.5 text-[11px]",
                  status === "done"
                    ? "bg-emerald-500/10 text-emerald-600"
                    : status === "active"
                      ? "bg-primary/10 text-primary"
                      : "bg-muted text-muted-foreground",
                )}
              >
                {statusText}
              </span>
            </div>
            <p className="mt-1 text-xs leading-5 text-muted-foreground">
              {description}
            </p>
          </div>
        </div>
        {typeof progress === "number" && (
          <span className="shrink-0 text-xs tabular-nums text-muted-foreground">
            {progress}%
          </span>
        )}
      </div>
      {typeof progress === "number" && (
        <div className="mt-3 h-2 overflow-hidden rounded-full bg-muted">
          <div
            className={cn(
              "h-full rounded-full transition-[width] duration-300",
              status === "done"
                ? "bg-emerald-500"
                : status === "active"
                  ? "bg-primary"
                  : "bg-muted-foreground/40",
            )}
            style={{ width: `${Math.max(0, Math.min(100, progress))}%` }}
          />
        </div>
      )}
      <p className="mt-3 text-xs leading-5 text-foreground/80">{detail}</p>
    </div>
  );
}

/**
 * 渲染顶部步骤导航，仅用于表达串行进度。
 */
function StepBadge({ index, title, done }: StepBadgeProps) {
  return (
    <div
      className={cn(
        "flex items-center gap-2 rounded-full border px-3 py-1.5 text-xs",
        done
          ? "border-emerald-500/40 bg-emerald-500/5 text-emerald-700"
          : "border-border/70 bg-background/70 text-muted-foreground",
      )}
    >
      <span
        className={cn(
          "flex size-5 items-center justify-center rounded-full text-[11px] font-semibold",
          done ? "bg-emerald-500 text-white" : "bg-muted text-muted-foreground",
        )}
      >
        {index}
      </span>
      <span>{title}</span>
    </div>
  );
}

/**
 * 渲染当前步骤专属的环境信息卡片。
 */
function StepInfoCard({ children }: { children: ReactNode }) {
  return (
    <div className="mt-4 w-full rounded-lg bg-card p-3 text-left text-xs leading-5 text-card-foreground">
      {children}
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
  const [pluginStepConfirmed, setPluginStepConfirmed] = useState(false);
  const showGatewayStart = installed && configured && !gatewayRunning;
  const nodeStatus = !nodeInstalled
    ? "未安装"
    : nodeVersionSatisfied
      ? "已安装（符合要求）"
      : "已安装（版本不符合要求）";
  const showNodeVersionWarning = nodeInstalled && !nodeVersionSatisfied;
  const title = "OpenClaw 环境准备";
  const description =
    "按照下面的步骤完成环境准备。当前先支持 Node 与 OpenClaw 自动安装，插件安装步骤先预留位置。";
  const installStatus = !installed
    ? "未安装"
    : configured
      ? "已安装并已初始化"
      : "已安装（待初始化）";
  const taskLog = installTask?.log?.[installTask.log.length - 1] ?? "";
  const taskPaused = Boolean(installTask?.paused);
  const taskControllable =
    installTask?.status === "running" || installTask?.status === "paused";
  const nodeReady = nodeInstalled && nodeVersionSatisfied && npmInstalled;
  const openClawReady = installed && configured;
  const pluginReady = pluginStepConfirmed;
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
  const currentTaskStage = installTask?.stage ?? "";

  useEffect(() => {
    if (!openClawReady) {
      setPluginStepConfirmed(false);
    }
  }, [openClawReady]);

  const currentStep = !nodeReady
    ? 1
    : !openClawReady
      ? 2
      : !pluginReady
        ? 3
        : !gatewayRunning
          ? 4
          : 4;
  const stepOneStatus: SetupStep["status"] = nodeReady
    ? "done"
    : currentStep === 1
      ? "active"
      : "pending";
  const stepTwoStatus: SetupStep["status"] = openClawReady
    ? "done"
    : currentStep === 2
      ? "active"
      : "pending";
  const stepThreeStatus: SetupStep["status"] = pluginReady
    ? "done"
    : currentStep === 3
      ? "active"
      : "pending";
  const stepFourStatus: SetupStep["status"] = gatewayRunning
    ? "done"
    : currentStep === 4
      ? "active"
      : "pending";
  const steps: SetupStep[] = [
    {
      index: 1,
      title: "安装 Node.js",
      description: "确保 Node.js 与 npm 可用，并满足 OpenClaw 最低版本要求。",
      status: stepOneStatus,
      progress:
        stepOneStatus === "done"
          ? 100
          : taskControllable || installing
            ? nodeStage.progress
            : 0,
      detail: nodeReady
        ? `当前状态：${nodeStatus}${nodeVersion ? `，版本 ${nodeVersion}` : ""}`
        : nodeStage.message || "将自动安装或切换到符合要求的 Node.js 版本",
    },
    {
      index: 2,
      title: "安装 OpenClaw 并初始化",
      description: "安装 openclaw 可执行文件，并生成可用的初始化配置。",
      status: stepTwoStatus,
      progress:
        stepTwoStatus === "done"
          ? 100
          : taskControllable || installing
            ? openClawStage.progress
            : 0,
      detail: openClawReady
        ? `当前状态：${installStatus}`
        : openClawStage.message || "将自动安装 OpenClaw 并执行初始化",
    },
    {
      index: 3,
      title: "安装必要插件",
      description: "预留步骤，后续补充默认插件安装与校验流程。",
      status: stepThreeStatus,
      detail: pluginReady
        ? "已确认跳过插件安装预留步骤。"
        : "当前暂未启用，确认后进入下一步。",
    },
    {
      index: 4,
      title: "开启网管",
      description: "启动 OpenClaw gateway，准备进入聊天工作区。",
      status: stepFourStatus,
      detail: gatewayRunning
        ? "网关已启动，可以进入聊天。"
        : showGatewayStart
          ? "基础环境已完成，下一步直接启动网管。"
          : "需先完成前两步后才能开启网管。",
    },
  ];
  const currentStepData = steps[currentStep - 1];
  const renderStepInfo = () => {
    if (currentStep === 1) {
      return (
        <StepInfoCard>
          <p>Node 环境：{nodeStatus}</p>
          {nodePath && <p className="mt-1">Node 路径：{nodePath}</p>}
          {nodeVersion && <p className="mt-1">Node 版本：{nodeVersion}</p>}
          <p className="mt-1">npm 环境：{npmInstalled ? "已安装" : "未安装"}</p>
          {npmPath && <p className="mt-1">npm 路径：{npmPath}</p>}
          <p className="mt-1">Node 要求：&gt;= 22.16.0，推荐 24.x</p>
          {showNodeVersionWarning && (
            <p className="mt-2 text-amber-600">
              当前 Node 版本不满足 OpenClaw 要求，将自动升级或切换到合规版本。
            </p>
          )}
          {!autoInstallSupported && autoInstallHint && (
            <p className="mt-2 text-amber-600">{autoInstallHint}</p>
          )}
          {(taskControllable || installing || nodeStage.progress > 0) && (
            <div className="mt-3">
              <InstallProgressBar {...nodeStage} />
            </div>
          )}
        </StepInfoCard>
      );
    }

    if (currentStep === 2) {
      return (
        <StepInfoCard>
          <p>安装状态：{installStatus}</p>
          {binaryPath && <p className="mt-1">可执行文件：{binaryPath}</p>}
          <p className="mt-1">配置文件：{configPath}</p>
          {!configured && (
            <p className="mt-2">
              推荐命令：
              <code>npm i -g openclaw@latest && openclaw init</code>
            </p>
          )}
          {(taskControllable || installing || openClawStage.progress > 0) && (
            <div className="mt-3">
              <InstallProgressBar {...openClawStage} />
            </div>
          )}
        </StepInfoCard>
      );
    }

    if (currentStep === 3) {
      return (
        <StepInfoCard>
          <p>插件步骤：预留中</p>
          <p className="mt-1">
            当前暂未接入默认插件安装逻辑，这一步仅作为后续扩展的占位。
          </p>
          <p className="mt-1">确认后会进入第 4 步开启网管。</p>
        </StepInfoCard>
      );
    }

    return (
      <StepInfoCard>
        <p>网关状态：{gatewayRunning ? "运行中" : "未启动"}</p>
        <p className="mt-1">
          {gatewayRunning
            ? "OpenClaw gateway 已准备就绪。"
            : "当前基础环境已就绪，可以直接开启网管。"}
        </p>
        <p className="mt-1">配置文件：{configPath}</p>
      </StepInfoCard>
    );
  };

  return (
    <div className="h-full w-full p-6">
      <div className="h-full rounded-lg">
        <div className="mx-auto flex h-full max-w-3xl flex-col items-center justify-center text-center">
          <h3 className="mt-5 text-lg font-semibold ">{title}</h3>
          <p className="mt-2 text-sm text-secondary-foreground">
            {description}
          </p>
          <div className="mt-4 flex w-full flex-wrap justify-center gap-2">
            {steps.map((step) => (
              <StepBadge
                key={step.index}
                index={step.index}
                title={step.title}
                done={step.status === "done"}
              />
            ))}
          </div>
          <div className="mt-4 w-full">
            <SetupStepCard {...currentStepData} />
          </div>
          {renderStepInfo()}
          {taskLog && (
            <p className="mt-2 w-full break-all rounded-lg bg-muted/40 px-3 py-2 text-left text-xs text-muted-foreground">
              当前任务：{taskLog}
            </p>
          )}
          <div className="mt-4 flex flex-wrap items-center gap-3">
            {currentStep <= 2 && !openClawReady && (
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
                {!nodeReady ? "执行第 1 步和第 2 步" : "执行第 2 步"}
              </Button>
            )}
            {currentStep === 3 && (
              <Button
                onClick={() => setPluginStepConfirmed(true)}
                disabled={checking || starting || installing}
              >
                <Wrench className="size-4" />
                确认并进入第 4 步
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
            {currentStep === 4 && showGatewayStart && (
              <Button
                onClick={onStartGateway}
                disabled={checking || starting || installing}
              >
                {starting ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Rocket className="size-4" />
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
