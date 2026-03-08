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

import { FC, useEffect, useState } from "react";
import { useSelectedMenuItem } from "@/components/Sidebar/store";
import { callWails } from "@/lib/utils";
import { ClawService } from "@wails/service";
import type { ClawOpenClawCheckResult } from "@wails/types/models";
import { OverviewPanel } from "./components/OverviewPanel";
import { ChannelPanel } from "./components/ChannelPanel";
import { SkillPanel } from "./components/SkillPanel";
import { ChatPanel } from "./components/ChatPanel";
import { OpenClawPendingPanel } from "./components/OpenClawPendingPanel";

/**
 * ClawContent 核心渲染组件
 * 根据选中的菜单项显示对应的面板
 */
export const ClawContentCore: FC = () => {
  const selectedMenuItem = useSelectedMenuItem();
  const [openClawCheck, setOpenClawCheck] =
    useState<ClawOpenClawCheckResult | null>(null);
  const [checking, setChecking] = useState(true);

  /**
   * 拉取 OpenClaw 安装与配置状态。
   */
  const fetchOpenClawCheck = async () => {
    setChecking(true);
    try {
      const result = await callWails(ClawService.CheckOpenClaw);
      if (!result) {
        return;
      }
      setOpenClawCheck(result);
    } finally {
      setChecking(false);
    }
  };

  useEffect(() => {
    // fetchOpenClawCheck();
  }, []);

  const renderContent = () => {
    switch (selectedMenuItem) {
      case "overview":
        return <OverviewPanel />;
      case "channel":
        return <ChannelPanel />;
      case "skill":
        return <SkillPanel />;
      case "instance":
        return (
          <div className="p-6 text-muted-foreground">实例面板（开发中）</div>
        );
      case "session":
        return (
          <div className="p-6 text-muted-foreground">会话面板（开发中）</div>
        );
      case "usage":
        return (
          <div className="p-6 text-muted-foreground">
            使用情况面板（开发中）
          </div>
        );
      case "scheduled":
        return (
          <div className="p-6 text-muted-foreground">
            定时任务面板（开发中）
          </div>
        );
      case "chat":
        return <ChatPanel />;
      default:
        return (
          <div className="h-full flex items-center justify-center text-muted-foreground">
            <p>请从左侧选择一个菜单项</p>
          </div>
        );
    }
  };

  return (
    <div className="h-full w-full overflow-hidden bg-background">
      {!openClawCheck?.configured ? (
        <OpenClawPendingPanel
          installed={openClawCheck?.installed ?? false}
          checking={checking}
          binaryPath={openClawCheck?.binaryPath ?? ""}
          configPath={openClawCheck?.configPath ?? ""}
          onRefresh={fetchOpenClawCheck}
        />
      ) : (
        renderContent()
      )}
    </div>
  );
};

export default ClawContentCore;
