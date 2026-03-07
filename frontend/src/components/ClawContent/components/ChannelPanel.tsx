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

import { FC, ReactNode, useState } from "react";
import { Channel } from "./ChannelListItem";
import { ChannelList } from "./ChannelList";
import { ChannelConfigPanel } from "./ChannelConfigPanel";
import { RadioIcon } from "lucide-react";

/**
 * 频道列表数据
 */
export const channels: Channel[] = [
  {
    id: "lark",
    name: "飞书",
    description: "飞书机器人 WebSocket (插件)",
    icon: <RadioIcon />,
    status: "configured",
  },

  {
    id: "whatsapp",
    name: "WhatsApp",
    icon: <RadioIcon />,
    status: "unconfigured",
  },
  {
    id: "telegram",
    name: "Telegram",
    icon: <RadioIcon />,
    status: "unconfigured",
  },
  {
    id: "discord",
    name: "Discord",
    icon: <RadioIcon />,
    status: "unconfigured",
  },
];

/**
 * 频道配置面板组件
 * 左侧为频道列表，右侧为配置详情
 */
export const ChannelPanel: FC = () => {
  const [selectedChannelId, setSelectedChannelId] = useState("lark");

  return (
    <div className="h-full w-full overflow-auto p-6">
      {/* 标题区域 */}
      <div className="flex items-start justify-between mb-6">
        <div className="flex flex-col gap-1 text-left">
          <h1 className="text-xl font-bold">通道管理</h1>
          <p className="text-sm text-muted-foreground">
            配置和管理所有消息通道 — 已启用 已配置未启用 未配置
          </p>
        </div>
      </div>
      <div className="flex gap-6">
        {/* 左侧频道列表 */}
        <ChannelList
          channels={channels}
          selectedChannelId={selectedChannelId}
          onChannelSelect={setSelectedChannelId}
        />

        {/* 右侧配置详情 */}
        <ChannelConfigPanel channelId={selectedChannelId} />
      </div>
    </div>
  );
};

export default ChannelPanel;
