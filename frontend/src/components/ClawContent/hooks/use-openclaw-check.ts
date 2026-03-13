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

import { useCallback, useEffect, useState } from "react";
import { callWails } from "@/lib/utils";
import { ClawService } from "@wails/service";
import type { ClawOpenClawCheckResult } from "@wails/types/models";

/**
 * OpenClaw 检测状态。
 */
interface OpenClawCheckState {
  openClawCheck: ClawOpenClawCheckResult | null;
  checking: boolean;
  starting: boolean;
  gatewayRunning: boolean;
  refreshOpenClawCheck: () => Promise<void>;
  startOpenClawGateway: () => Promise<void>;
}

/**
 * 管理 OpenClaw 安装与配置检测生命周期。
 */
export function useOpenClawCheck(): OpenClawCheckState {
  const [openClawCheck, setOpenClawCheck] =
    useState<ClawOpenClawCheckResult | null>(null);
  const [checking, setChecking] = useState(true);
  const [starting, setStarting] = useState(false);
  const [gatewayRunning, setGatewayRunning] = useState(false);

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

  useEffect(() => {
    void refreshOpenClawCheck();
  }, [refreshOpenClawCheck]);

  return {
    openClawCheck,
    checking,
    starting,
    gatewayRunning,
    refreshOpenClawCheck,
    startOpenClawGateway,
  };
}
