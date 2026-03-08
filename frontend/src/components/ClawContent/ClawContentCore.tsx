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
import { useSelectedMenuItem } from "@/components/Sidebar/store";
import { OverviewPanel } from "./components/OverviewPanel";
import { ChannelPanel } from "./components/ChannelPanel";
import { SkillPanel } from "./components/SkillPanel";

/**
 * ClawContent 核心渲染组件
 * 根据选中的菜单项显示对应的面板
 */
export const ClawContentCore: FC = () => {
  const selectedMenuItem = useSelectedMenuItem();

  const renderContent = () => {
    switch (selectedMenuItem) {
      case "overview":
        return <OverviewPanel />;
      case "channel":
        return <ChannelPanel />;
      case "skill":
        return <SkillPanel />;
      case "instance":
        return <div className="p-6 text-muted-foreground">实例面板（开发中）</div>;
      case "session":
        return <div className="p-6 text-muted-foreground">会话面板（开发中）</div>;
      case "usage":
        return <div className="p-6 text-muted-foreground">使用情况面板（开发中）</div>;
      case "scheduled":
        return <div className="p-6 text-muted-foreground">定时任务面板（开发中）</div>;
      case "chat":
        return <div className="p-6 text-muted-foreground">聊天面板（开发中）</div>;
      default:
        return (
          <div className="h-full flex items-center justify-center text-muted-foreground">
            <p>请从左侧选择一个菜单项</p>
          </div>
        );
    }
  };

  return <div className="h-full w-full overflow-hidden bg-background">{renderContent()}</div>;
};

export default ClawContentCore;
