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

import { FC, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Loader2, Sparkles } from "lucide-react";
import { useSelectedMenuItem } from "@/components/Sidebar/store";
import { cn } from "@/lib/utils";
import { OpenClawPendingPanel } from "./components/OpenClawPendingPanel";
import { useClawMenuContent, useOpenClawCheck } from "./hooks";

const OPENING_TRANSITION_MIN_MS = 1400;

interface OpenClawTransitionPanelProps {
  checking: boolean;
}

/**
 * 渲染 OpenClaw 检测与开启过程中的过渡态，避免面板切换闪烁。
 */
const OpenClawTransitionPanel: FC<OpenClawTransitionPanelProps> = ({
  checking,
}) => {
  const title = checking ? "正在检测 OpenClaw 环境" : "正在开启 OpenClaw";
  const description = checking
    ? "正在校验运行环境与网关状态，稍后会自动进入聊天界面。"
    : "网关已就绪，正在加载聊天工作区。";

  return (
    <div className="flex h-full w-full items-center justify-center overflow-hidden bg-background p-6">
      <div className="relative w-full max-w-xl overflow-hidden rounded-[28px] border border-border/70 bg-card/95 px-8 py-10 text-center shadow-[0_30px_80px_-40px_rgba(15,23,42,0.45)]">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(59,130,246,0.18),_transparent_55%),radial-gradient(circle_at_bottom_right,_rgba(16,185,129,0.14),_transparent_45%)]" />
        <div className="relative flex flex-col items-center">
          <div className="relative flex size-20 items-center justify-center rounded-full border border-primary/20 bg-primary/8">
            <div className="absolute inset-0 rounded-full border border-primary/25 animate-ping" />
            <Loader2 className="size-9 animate-spin text-primary" />
            <Sparkles className="absolute -right-1 -top-1 size-4 text-emerald-500" />
          </div>
          <h3 className="mt-6 text-xl font-semibold text-foreground">{title}</h3>
          <p className="mt-2 max-w-md text-sm leading-6 text-muted-foreground">
            {description}
          </p>
          <div className="mt-8 w-full max-w-sm space-y-3">
            <div className="h-2 overflow-hidden rounded-full bg-muted">
              <div className="h-full w-2/3 rounded-full bg-[linear-gradient(90deg,rgba(59,130,246,0.95),rgba(16,185,129,0.95))] animate-pulse" />
            </div>
            <div className="flex items-center justify-center gap-2 text-xs text-muted-foreground">
              <span
                className={cn(
                  "size-2 rounded-full bg-primary/40",
                  "animate-pulse",
                )}
              />
              <span>{checking ? "检测中" : "加载中"}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

/**
 * ClawContent 核心渲染组件
 * 根据选中的菜单项显示对应的面板
 */
export const ClawContentCore: FC = () => {
  const selectedMenuItem = useSelectedMenuItem();
  const menuContent = useClawMenuContent(selectedMenuItem);
  const {
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
  } = useOpenClawCheck();
  const [showOpeningTransition, setShowOpeningTransition] = useState(true);
  const transitionTimerRef = useRef<number | null>(null);

  const ready = useMemo(
    () => Boolean(openClawCheck?.configured && gatewayRunning),
    [gatewayRunning, openClawCheck?.configured],
  );
  const shouldShowPending = useMemo(() => {
    if (checking && !openClawCheck) {
      return false;
    }
    return !ready;
  }, [checking, openClawCheck, ready]);

  /**
   * 处理用户主动触发的 OpenClaw 状态刷新。
   */
  const handleRefreshOpenClawCheck = useCallback(() => {
    void refreshOpenClawCheck();
  }, [refreshOpenClawCheck]);

  /**
   * 处理用户主动启动 OpenClaw gateway。
   */
  const handleStartOpenClawGateway = useCallback(() => {
    void startOpenClawGateway();
  }, [startOpenClawGateway]);

  /**
   * 处理用户主动触发的 OpenClaw 自动安装。
   */
  const handleStartOpenClawSetup = useCallback(() => {
    void startOpenClawSetup();
  }, [startOpenClawSetup]);

  /**
   * 处理暂停安装任务。
   */
  const handlePauseOpenClawSetup = useCallback(() => {
    void pauseOpenClawSetup();
  }, [pauseOpenClawSetup]);

  /**
   * 处理恢复安装任务。
   */
  const handleResumeOpenClawSetup = useCallback(() => {
    void resumeOpenClawSetup();
  }, [resumeOpenClawSetup]);

  /**
   * 处理取消安装任务。
   */
  const handleCancelOpenClawSetup = useCallback(() => {
    void cancelOpenClawSetup();
  }, [cancelOpenClawSetup]);

  useEffect(() => {
    if (transitionTimerRef.current) {
      window.clearTimeout(transitionTimerRef.current);
      transitionTimerRef.current = null;
    }

    if (!ready) {
      setShowOpeningTransition(checking && !openClawCheck);
      return;
    }

    setShowOpeningTransition(true);
    transitionTimerRef.current = window.setTimeout(() => {
      setShowOpeningTransition(false);
      transitionTimerRef.current = null;
    }, OPENING_TRANSITION_MIN_MS);

    return () => {
      if (transitionTimerRef.current) {
        window.clearTimeout(transitionTimerRef.current);
        transitionTimerRef.current = null;
      }
    };
  }, [checking, openClawCheck, ready]);

  return (
    <div className="h-full w-full overflow-hidden bg-background">
      {shouldShowPending ? (
        <OpenClawPendingPanel
          installed={openClawCheck?.installed ?? false}
          configured={openClawCheck?.configured ?? false}
          nodeInstalled={openClawCheck?.nodeInstalled ?? false}
          nodeVersionSatisfied={openClawCheck?.nodeVersionSatisfied ?? false}
          npmInstalled={openClawCheck?.npmInstalled ?? false}
          checking={checking}
          starting={starting}
          installing={installing}
          gatewayRunning={gatewayRunning}
          binaryPath={openClawCheck?.binaryPath ?? ""}
          configPath={openClawCheck?.configPath ?? ""}
          nodePath={openClawCheck?.nodePath ?? ""}
          nodeVersion={openClawCheck?.nodeVersion ?? ""}
          npmPath={openClawCheck?.npmPath ?? ""}
          autoInstallSupported={openClawCheck?.autoInstallSupported ?? false}
          autoInstallHint={openClawCheck?.autoInstallHint ?? ""}
          installTask={installTask}
          onRefresh={handleRefreshOpenClawCheck}
          onStartGateway={handleStartOpenClawGateway}
          onStartAutoInstall={handleStartOpenClawSetup}
          onPauseAutoInstall={handlePauseOpenClawSetup}
          onResumeAutoInstall={handleResumeOpenClawSetup}
          onCancelAutoInstall={handleCancelOpenClawSetup}
        />
      ) : showOpeningTransition ? (
        <OpenClawTransitionPanel checking={checking} />
      ) : (
        menuContent
      )}
    </div>
  );
};

export default ClawContentCore;
