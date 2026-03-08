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
import { CircleCheck } from "lucide-react";
import { useClawContentStore } from "../store";
import { statCardsConfig, getStatCardValue } from "../domain";
import { StatCardItem } from "./StatCardItem";
import { ChannelCardItem } from "./ChannelCardItem";
import { PanelHeader } from "./PanelHeader";

/**
 * 概览面板组件
 * 显示系统运行状态总览
 */
export const OverviewPanel: FC = () => {
  const overview = useClawContentStore((state) => state.overview);

  return (
    <div className="h-full w-full overflow-auto p-6">
      {/* 标题区域 */}
      <PanelHeader
        className="mb-6"
        title="仪表盘"
        description="OpenClaw 运行状态总览"
        actions={
          <div className="flex items-center gap-2">
            <CircleCheck className="size-3.5 text-emerald-500" />
            <span className="text-sm text-emerald-500">系统运行正常</span>
          </div>
        }
      />

      {/* 状态卡片组 */}
      <div className="flex flex-wrap gap-4 mb-6">
        {statCardsConfig.map((card) => (
          <StatCardItem
            key={card.id}
            label={card.label}
            value={getStatCardValue(card.id, overview)}
            icon={card.icon}
            iconBgColor={card.iconBgColor}
            iconColor={card.iconColor}
          />
        ))}
      </div>

      {/* 已连接通道区域 */}
      <div className="flex flex-col gap-3">
        <h2 className="text-base font-semibold text-left">已连接通道</h2>
        <div className="flex flex-wrap gap-4 ">
          {overview.channels.map((channel) => (
            <ChannelCardItem
              key={channel.id}
              name={channel.name}
              type={channel.type}
              status={channel.status}
              managedBy={channel.managedBy}
            />
          ))}
        </div>
      </div>
    </div>
  );
};

export default OverviewPanel;
