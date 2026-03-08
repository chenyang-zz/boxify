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
import { ChannelListItem } from "./ChannelListItem";
import type { ManagedChannel as Channel } from "../types";

/**
 * 频道列表组件属性
 */
export interface ChannelListProps {
  /** 频道数据列表 */
  channels: Channel[];
  /** 当前选中的频道 ID */
  selectedChannelId: string;
  /** 频道选中回调 */
  onChannelSelect: (channelId: string) => void;
}

/**
 * 频道列表组件
 * 展示所有可配置的通道列表
 */
export const ChannelList: FC<ChannelListProps> = ({
  channels,
  selectedChannelId,
  onChannelSelect,
}) => {
  const statusOrder: Record<Channel["status"], number> = {
    enabled: 0,
    configured: 1,
    unconfigured: 2,
  };
  const sortByStatus = (a: Channel, b: Channel) =>
    statusOrder[a.status] - statusOrder[b.status];
  const builtInChannels = channels
    .filter((channel) => channel.type === "built-in")
    .sort(sortByStatus);
  const pluginChannels = channels
    .filter((channel) => channel.type === "plugin")
    .sort(sortByStatus);

  return (
    <div className="w-62.5 max-h-[70vh]  overflow-auto bg-card rounded-lg flex flex-col">
      <h2 className="text-base font-semibold text-left p-4 pb-0 ">频道列表</h2>
      <div className="flex-1 overflow-auto p-2 space-y-1 text-left">
        <h3 className="text-xs px-2 text-muted-foreground py-1">内置频道</h3>
        {builtInChannels.map((channel) => (
          <ChannelListItem
            key={channel.id}
            channel={channel}
            isSelected={selectedChannelId === channel.id}
            onClick={() => onChannelSelect(channel.id)}
          />
        ))}
        <h3 className="text-xs px-2 text-muted-foreground py-1">插件频道</h3>
        {pluginChannels.map((channel) => (
          <ChannelListItem
            key={channel.id}
            channel={channel}
            isSelected={selectedChannelId === channel.id}
            onClick={() => onChannelSelect(channel.id)}
          />
        ))}
      </div>
    </div>
  );
};

export default ChannelList;
