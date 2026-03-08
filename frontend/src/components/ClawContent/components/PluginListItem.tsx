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
import { LucideIcon } from "lucide-react";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { CardItem } from "./CardItem";

export interface PluginListItemProps {
  /** 技能 ID */
  id: string;
  /** 技能名称 */
  name: string;
  /** 技能描述 */
  description: string;
  /** 技能图标 */
  icon: LucideIcon;
  /** 是否启用 */
  enabled: boolean;
  /** 开关状态变更回调 */
  onToggle: (id: string, enabled: boolean) => void;
  /** 设置按钮点击回调 */
  onSettingsClick?: (id: string) => void;
}

/**
 * 插件列表项组件
 * 用于展示单个插件的配置信息（列表行布局）
 */
export const PluginListItem: FC<PluginListItemProps> = ({
  id,
  name,
  description,
  icon: Icon,
  enabled,
  onToggle,
}) => {
  const [isEnabled, setIsEnabled] = useState(enabled);

  const handleToggle = (checked: boolean) => {
    setIsEnabled(checked);
    onToggle(id, checked);
  };

  return (
    <CardItem
      label={name}
      icon={
        <div className="flex items-center justify-center size-10 bg-primary/10 rounded-lg shrink-0">
          <Icon className="size-5 text-primary" />
        </div>
      }
      badge={
        <Badge
          variant={isEnabled ? "default" : "secondary"}
          className="text-[10px] px-1.5 py-0"
        >
          {isEnabled ? "已启用" : "未启用"}
        </Badge>
      }
      description={description}
      action={<Switch checked={isEnabled} onCheckedChange={handleToggle} />}
    />
  );
};

export default PluginListItem;
