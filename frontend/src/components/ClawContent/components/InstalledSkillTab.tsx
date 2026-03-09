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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Search } from "lucide-react";
import { FC, useMemo, useState } from "react";
import SkillListItem, { SkillListItemProps } from "./SkillListItem";

interface InstalledSkillTabProps {
  skills: Omit<SkillListItemProps, "onToggle" | "onSettingsClick">[];
  loading?: boolean;
  onToggle: (id: string, enabled: boolean) => void | Promise<void>;
}

/**
 * 已安装技能面板
 * 展示已安装技能的列表，提供启用/禁用入口
 */
export const InstalledSkillTab: FC<InstalledSkillTabProps> = ({
  skills,
  loading = false,
  onToggle,
}) => {
  const [searchValue, setSearchValue] = useState("");

  /** 预留设置入口，后续补充配置弹窗。 */
  const handleSettingsClick = (id: string) => {
    console.log(`Skill ${id} settings clicked`);
  };

  /** 根据搜索词过滤技能。 */
  const filteredSkills = useMemo(() => {
    const query = searchValue.trim().toLowerCase();
    if (!query) return skills;

    return skills.filter(
      (skill) =>
        skill.name.toLowerCase().includes(query) ||
        skill.description.toLowerCase().includes(query),
    );
  }, [searchValue, skills]);

  const enabledSkills = useMemo(
    () => filteredSkills.filter((skill) => skill.enabled),
    [filteredSkills],
  );
  const disabledSkills = useMemo(
    () => filteredSkills.filter((skill) => !skill.enabled),
    [filteredSkills],
  );

  /** 渲染当前筛选结果列表。 */
  const renderSkillList = (items: typeof filteredSkills) => {
    if (loading) {
      return (
        <div className="flex items-center justify-center py-12 text-muted-foreground text-sm">
          正在加载技能...
        </div>
      );
    }

    if (items.length === 0) {
      return (
        <div className="flex items-center justify-center py-12 text-muted-foreground text-sm">
          {searchValue ? "没有找到匹配的技能" : "暂无技能"}
        </div>
      );
    }

    return (
      <div className="flex flex-col gap-2">
        {items.map((skill) => (
          <SkillListItem
            key={skill.id}
            {...skill}
            onToggle={onToggle}
            onSettingsClick={handleSettingsClick}
          />
        ))}
      </div>
    );
  };

  return (
    <Tabs defaultValue="all" className="w-full">
      <div className="mb-3 flex justify-between gap-6">
        <div className="relative max-w-md flex-1">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="搜索技能..."
            value={searchValue}
            onChange={(e) => setSearchValue(e.target.value)}
            className="pl-9"
          />
        </div>

        <TabsList>
          <TabsTrigger value="all">全部</TabsTrigger>
          <TabsTrigger value="enabled">已启用</TabsTrigger>
          <TabsTrigger value="disabled">已禁用</TabsTrigger>
        </TabsList>
      </div>

      <TabsContent value="all">{renderSkillList(filteredSkills)}</TabsContent>
      <TabsContent value="enabled">
        {renderSkillList(enabledSkills)}
      </TabsContent>
      <TabsContent value="disabled">
        {renderSkillList(disabledSkills)}
      </TabsContent>
    </Tabs>
  );
};

export default InstalledSkillTab;
