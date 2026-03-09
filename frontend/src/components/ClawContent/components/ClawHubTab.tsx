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
import { ClawHubSkillCard, ClawHubSkillCardData } from "./ClawHubSkillCard";

interface ClawHubTabProps {
  skills: ClawHubSkillCardData[];
  loading?: boolean;
  onInstall: (id: string) => void | Promise<void>;
}

/**
 * ClawHub 技能市场面板
 * 展示可从仓库安装的插件列表
 */
export const ClawHubTab: FC<ClawHubTabProps> = ({
  skills,
  loading = false,
  onInstall,
}) => {
  const [searchValue, setSearchValue] = useState("");

  /** 根据搜索词过滤技能列表。 */
  const filteredSkills = useMemo(() => {
    const query = searchValue.trim().toLowerCase();
    if (!query) return skills;

    return skills.filter((skill) => {
      return (
        skill.name.toLowerCase().includes(query) ||
        skill.description.toLowerCase().includes(query) ||
        skill.descriptionZh.toLowerCase().includes(query) ||
        skill.category.toLowerCase().includes(query)
      );
    });
  }, [searchValue, skills]);

  let content: ReactNode;
  if (loading) {
    content = (
      <div className="flex items-center justify-center py-12 text-muted-foreground text-sm">
        正在加载 ClawHub 列表...
      </div>
    );
  } else if (filteredSkills.length === 0) {
    content = (
      <div className="flex items-center justify-center py-12 text-muted-foreground text-sm">
        {searchValue ? "没有找到匹配的技能" : "暂无可安装插件"}
      </div>
    );
  } else {
    content = (
      <div className="flex flex-wrap gap-4">
        {filteredSkills.map((skill) => (
          <ClawHubSkillCard key={skill.id} skill={skill} onInstall={onInstall} />
        ))}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="relative w-full">
        <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="搜索 ClawHub 技能..."
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          className="pl-9"
        />
      </div>

      {content}
    </div>
  );
};

export default ClawHubTab;
