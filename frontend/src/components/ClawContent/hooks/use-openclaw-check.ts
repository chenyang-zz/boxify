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

import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { callWails } from "@/lib/utils";
import { ClawService } from "@wails/service";
import type { ClawOpenClawCheckResult } from "@wails/types/models";
import type { Task } from "@wails/claw/taskman/models";

/**
 * OpenClaw 检测状态。
 */
interface OpenClawCheckState {
  openClawCheck: ClawOpenClawCheckResult | null;
  installTask: Task | null;
  checking: boolean;
  starting: boolean;
  installing: boolean;
  gatewayRunning: boolean;
  refreshOpenClawCheck: () => Promise<void>;
  startOpenClawGateway: () => Promise<void>;
  startOpenClawSetup: () => Promise<void>;
  pauseOpenClawSetup: () => Promise<void>;
  resumeOpenClawSetup: () => Promise<void>;
  cancelOpenClawSetup: () => Promise<void>;
}

/**
 * 管理 OpenClaw 安装与配置检测生命周期。
 */
export function useOpenClawCheck(): OpenClawCheckState {
  const [openClawCheck, setOpenClawCheck] =
    useState<ClawOpenClawCheckResult | null>(null);
  const [installTask, setInstallTask] = useState<Task | null>(null);
  const [checking, setChecking] = useState(true);
  const [starting, setStarting] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [gatewayRunning, setGatewayRunning] = useState(false);
  const completedTaskRef = useRef<string>("");

  /**
   * 拉取 OpenClaw 安装与配置状态。
   */
  const refreshOpenClawCheck = useCallback(async () => {
    setChecking(true);
    try {
      const [checkResult, statusResult] = await Promise.all([
        callWails(ClawService.CheckOpenClaw),
        callWails(ClawService.GetStatus),
      ]);
      setOpenClawCheck(checkResult ?? null);
      setGatewayRunning(Boolean(statusResult?.data?.running));
    } finally {
      setChecking(false);
    }
  }, []);

  /**
   * 显式启动 OpenClaw gateway，并在成功后刷新状态。
   */
  const startOpenClawGateway = useCallback(async () => {
    setStarting(true);
    try {
      await callWails(ClawService.StartProcess);
      await refreshOpenClawCheck();
    } finally {
      setStarting(false);
    }
  }, [refreshOpenClawCheck]);

  /**
   * 启动 OpenClaw 自动安装任务，并记录任务句柄供后续轮询。
   */
  const startOpenClawSetup = useCallback(async () => {
    setInstalling(true);
    completedTaskRef.current = "";
    try {
      const result = await callWails(ClawService.StartOpenClawSetup);
      setInstallTask(result?.task ?? null);
      toast.success("已开始自动安装 OpenClaw", {
        description:
          "将按当前系统环境自动补齐 Node.js（要求 >=22.16，优先 24）、OpenClaw 与初始化配置",
        style: { textAlign: "left" },
      });
    } catch (error) {
      setInstalling(false);
      throw error;
    }
  }, []);

  /**
   * 暂停当前 OpenClaw 自动安装任务。
   */
  const pauseOpenClawSetup = useCallback(async () => {
    if (!installTask?.id) {
      return;
    }
    await callWails(ClawService.PauseTask, installTask.id);
    const detail = await callWails(ClawService.GetTaskDetail, installTask.id);
    setInstallTask(detail?.task ?? null);
  }, [installTask?.id]);

  /**
   * 恢复当前 OpenClaw 自动安装任务。
   */
  const resumeOpenClawSetup = useCallback(async () => {
    if (!installTask?.id) {
      return;
    }
    await callWails(ClawService.ResumeTask, installTask.id);
    const detail = await callWails(ClawService.GetTaskDetail, installTask.id);
    setInstallTask(detail?.task ?? null);
  }, [installTask?.id]);

  /**
   * 取消当前 OpenClaw 自动安装任务。
   */
  const cancelOpenClawSetup = useCallback(async () => {
    if (!installTask?.id) {
      return;
    }
    await callWails(ClawService.CancelTask, installTask.id);
    const detail = await callWails(ClawService.GetTaskDetail, installTask.id);
    setInstallTask(detail?.task ?? null);
    setInstalling(false);
    await refreshOpenClawCheck();
  }, [installTask?.id, refreshOpenClawCheck]);

  useEffect(() => {
    if (!installTask?.id) {
      return;
    }
    if (
      installTask.status !== "pending" &&
      installTask.status !== "running" &&
      installTask.status !== "paused"
    ) {
      setInstalling(false);
      return;
    }

    const timer = window.setInterval(async () => {
      const detail = await ClawService.GetTaskDetail(installTask.id);
      if (!detail?.success || !detail.task) {
        return;
      }

      setInstallTask(detail.task);
      if (
        detail.task.status === "pending" ||
        detail.task.status === "running" ||
        detail.task.status === "paused" ||
        completedTaskRef.current === detail.task.id
      ) {
        return;
      }

      completedTaskRef.current = detail.task.id;
      setInstalling(false);

      if (detail.task.status === "success") {
        toast.success("OpenClaw 安装完成", {
          description: "环境检测已刷新，可以继续进入聊天面板",
          style: { textAlign: "left" },
        });
        await refreshOpenClawCheck();
        return;
      }

      toast.error("OpenClaw 自动安装失败", {
        description:
          detail.task.error ||
          detail.task.log[detail.task.log.length - 1] ||
          "请根据任务日志排查安装环境",
        style: { textAlign: "left" },
      });
      await refreshOpenClawCheck();
    }, 2000);

    return () => {
      window.clearInterval(timer);
    };
  }, [installTask, refreshOpenClawCheck]);

  useEffect(() => {
    void refreshOpenClawCheck();
  }, [refreshOpenClawCheck]);

  return {
    openClawCheck,
    installTask,
    checking,
    starting,
    installing,
    gatewayRunning,
    refreshOpenClawCheck,
    startOpenClawGateway,
    startOpenClawSetup,
    pauseOpenClawSetup,
    resumeOpenClawSetup,
    cancelOpenClawSetup,
  };
}
