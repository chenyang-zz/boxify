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

import { Input } from "@/components/ui/input";
import { Search } from "lucide-react";
import { FC, ReactNode, useMemo, useState } from "react";
import SkillListItem, { SkillListItemProps } from "./SkillListItem";

interface PluginTabProps {
  plugins: Omit<SkillListItemProps, "onToggle" | "onSettingsClick">[];
  loading?: boolean;
  onToggle: (id: string, enabled: boolean) => void | Promise<void>;
}

/**
 * 插件面板
 * 展示已安装插件列表，提供启用/禁用入口
 */
export const PluginTab: FC<PluginTabProps> = ({
  plugins,
  loading = false,
  onToggle,
}) => {
  const [searchValue, setSearchValue] = useState("");

  /** 预留设置入口，后续补充配置弹窗。 */
  const handleSettingsClick = (id: string) => {
    console.log(`Plugin ${id} settings clicked`);
  };

  /** 根据搜索词过滤插件。 */
  const filteredPlugins = useMemo(() => {
    const query = searchValue.trim().toLowerCase();
    if (!query) return plugins;

    return plugins.filter(
      (plugin) =>
        plugin.name.toLowerCase().includes(query) ||
        plugin.description.toLowerCase().includes(query),
    );
  }, [plugins, searchValue]);

  let content: ReactNode;
  if (loading) {
    content = (
      <div className="flex items-center justify-center py-12 text-muted-foreground text-sm">
        正在加载插件...
      </div>
    );
  } else if (filteredPlugins.length === 0) {
    content = (
      <div className="flex items-center justify-center py-12 text-muted-foreground text-sm">
        {searchValue ? "没有找到匹配的插件" : "暂无插件"}
      </div>
    );
  } else {
    content = (
      <div className="flex flex-col gap-2">
        {filteredPlugins.map((plugin) => (
          <SkillListItem
            key={plugin.id}
            {...plugin}
            onToggle={onToggle}
            onSettingsClick={handleSettingsClick}
          />
        ))}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="relative w-full">
        <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="搜索插件..."
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          className="pl-9"
        />
      </div>

      {content}
    </div>
  );
};

export default PluginTab;
