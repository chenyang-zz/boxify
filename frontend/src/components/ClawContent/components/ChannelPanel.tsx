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

import { FC, useState } from "react";
import { Channel } from "./ChannelListItem";
import { ChannelList } from "./ChannelList";
import { ChannelConfigPanel } from "./ChannelConfigPanel";
import { RadioIcon } from "lucide-react";
import { PanelHeader } from "./PanelHeader";

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
      <PanelHeader
        className="mb-6"
        title="频道管理"
        description="配置和管理所有消息频道 — 已启用 已配置未启用 未配置"
      />
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
