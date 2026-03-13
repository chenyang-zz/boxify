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

import { FC, useCallback } from "react";
import { useSelectedMenuItem } from "@/components/Sidebar/store";
import { OpenClawPendingPanel } from "./components/OpenClawPendingPanel";
import { useClawMenuContent, useOpenClawCheck } from "./hooks";

/**
 * ClawContent 核心渲染组件
 * 根据选中的菜单项显示对应的面板
 */
export const ClawContentCore: FC = () => {
  const selectedMenuItem = useSelectedMenuItem();
  const menuContent = useClawMenuContent(selectedMenuItem);
  const {
    openClawCheck,
    checking,
    starting,
    gatewayRunning,
    refreshOpenClawCheck,
    startOpenClawGateway,
  } = useOpenClawCheck();

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

  return (
    <div className="h-full w-full overflow-hidden bg-background">
      {!openClawCheck?.configured || !gatewayRunning ? (
        <OpenClawPendingPanel
          installed={openClawCheck?.installed ?? false}
          configured={openClawCheck?.configured ?? false}
          checking={checking}
          starting={starting}
          gatewayRunning={gatewayRunning}
          binaryPath={openClawCheck?.binaryPath ?? ""}
          configPath={openClawCheck?.configPath ?? ""}
          onRefresh={handleRefreshOpenClawCheck}
          onStartGateway={handleStartOpenClawGateway}
        />
      ) : (
        menuContent
      )}
    </div>
  );
};

export default ClawContentCore;
