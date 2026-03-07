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
import { Badge } from "@/components/ui/badge";
import { getChannelTypeLabel, getStatusBadgeConfig } from "../domain";

export interface ChannelCardItemProps {
  name: string;
  type: "built-in" | "plugin";
  status: "enabled" | "disabled";
  managedBy: string;
}

/**
 * 通道卡片组件
 * 用于展示单个通道的连接状态信息
 */
export const ChannelCardItem: FC<ChannelCardItemProps> = ({
  name,
  type,
  status,
  managedBy,
}) => {
  const statusConfig = getStatusBadgeConfig(status);
  const typeLabel = getChannelTypeLabel(type);

  return (
    <div className="flex items-center flex-1 gap-4 p-4 bg-card rounded-xl min-w-80">
      <div className="flex items-center justify-center size-12 bg-background rounded-lg border">
        <span className="text-sm font-medium">
          {name.charAt(0).toUpperCase()}
        </span>
      </div>
      <div className="flex flex-col gap-1 flex-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{name}</span>
          <Badge className={statusConfig.className}>{statusConfig.text}</Badge>
        </div>
        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          <span>类型：{typeLabel}</span>
          <span>状态：{managedBy}</span>
        </div>
      </div>
    </div>
  );
};

export default ChannelCardItem;
