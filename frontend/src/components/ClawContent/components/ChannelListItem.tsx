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
import { cn } from "@/lib/utils";
import type { ManagedChannel } from "../types";

/**
 * 频道数据类型（兼容旧导出名）。
 */
export type Channel = ManagedChannel;

export interface ChannelListItemProps {
  channel: Channel;
  isSelected: boolean;
  onClick: () => void;
}

/**
 * 频道列表项组件
 */
export const ChannelListItem: FC<ChannelListItemProps> = ({
  channel,
  isSelected,
  onClick,
}) => {
  const statusClassName = cn(
    "size-2 rounded-full",
    channel.status === "enabled"
      ? "bg-emerald-500"
      : channel.status === "configured"
        ? "bg-amber-500"
        : "bg-muted-foreground/30",
  );

  return (
    <button
      onClick={onClick}
      className={cn(
        "group w-full flex items-center gap-3 px-2 py-2 rounded-lg border border-transparent transition-colors text-left",
        isSelected
          ? "bg-primary/20 text-foreground border-primary/40"
          : "text-muted-foreground hover:text-foreground hover:bg-accent/50",
      )}
    >
      <div
        className={cn(
          "flex items-center justify-center size-8 p-1.5 rounded-lg bg-secondary group-hover:bg-foreground text-foreground group-hover:text-background text-sm font-medium",
          isSelected && "bg-foreground text-background",
        )}
      >
        {channel.icon}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium truncate">{channel.name}</span>
          <div className={statusClassName} />
        </div>
        {channel.description && (
          <p className="text-[10px] text-muted-foreground/80 truncate">
            {channel.description}
          </p>
        )}
      </div>
    </button>
  );
};

export default ChannelListItem;
