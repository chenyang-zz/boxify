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
import {
  BarChart3,
  BrainCircuit,
  Code2,
  FileCode2,
  Globe,
  Palette,
  Search,
} from "lucide-react";
import { FC, useState } from "react";
import { SkillListItemProps } from "./SkillListItem";
import SkillListItem from "./SkillListItem";

/**
 * 已安装技能配置数据
 */
const installedSkills: Omit<
  SkillListItemProps,
  "onToggle" | "onSettingsClick"
>[] = [
  {
    id: "ai-drawing",
    name: "AI绘画",
    description: "DALL-E/SD/Midjourney等AI绘画工具的调用适配",
    icon: Palette,
    enabled: true,
  },
  {
    id: "web-search",
    name: "联网搜索",
    description: "联网搜索功能，支持搜索引擎选择",
    icon: Globe,
    enabled: true,
  },
  {
    id: "code-interpreter",
    name: "代码解释器",
    description: "Python代码执行环境，支持文件操作",
    icon: Code2,
    enabled: true,
  },
  {
    id: "web-parser",
    name: "网页解析",
    description: "网页内容解析，支持多种格式输出",
    icon: FileCode2,
    enabled: false,
  },
  {
    id: "data-analysis",
    name: "数据分析",
    description: "数据分析和可视化功能",
    icon: BarChart3,
    enabled: false,
  },
  {
    id: "chain-of-thought",
    name: "思维链",
    description: "思维链推理功能，用于复杂问题求解",
    icon: BrainCircuit,
    enabled: false,
  },
  {
    id: "code-interpreter",
    name: "代码解释器",
    description: "Python代码执行环境，支持文件操作",
    icon: Code2,
    enabled: true,
  },
  {
    id: "web-parser",
    name: "网页解析",
    description: "网页内容解析，支持多种格式输出",
    icon: FileCode2,
    enabled: false,
  },
  {
    id: "data-analysis",
    name: "数据分析",
    description: "数据分析和可视化功能",
    icon: BarChart3,
    enabled: false,
  },
  {
    id: "chain-of-thought",
    name: "思维链",
    description: "思维链推理功能，用于复杂问题求解",
    icon: BrainCircuit,
    enabled: false,
  },
  {
    id: "code-interpreter",
    name: "代码解释器",
    description: "Python代码执行环境，支持文件操作",
    icon: Code2,
    enabled: true,
  },
  {
    id: "web-parser",
    name: "网页解析",
    description: "网页内容解析，支持多种格式输出",
    icon: FileCode2,
    enabled: false,
  },
  {
    id: "data-analysis",
    name: "数据分析",
    description: "数据分析和可视化功能",
    icon: BarChart3,
    enabled: false,
  },
  {
    id: "chain-of-thought",
    name: "思维链",
    description: "思维链推理功能，用于复杂问题求解",
    icon: BrainCircuit,
    enabled: false,
  },
];

/**
 * 可用技能配置数据（市场中未安装的技能）
 */
const availableSkills: Omit<
  SkillListItemProps,
  "onToggle" | "onSettingsClick"
>[] = [
  {
    id: "weather",
    name: "天气查询",
    description: "实时天气查询和天气预报功能",
    icon: Globe,
    enabled: false,
  },
  {
    id: "translation",
    name: "多语言翻译",
    description: "支持多种语言的实时翻译功能",
    icon: Code2,
    enabled: false,
  },
];

/**
 * 已安装技能面板
 * 展示已安装技能的列表，提供启用/禁用和设置入口
 */
export const InstalledSkillTab: FC = () => {
  const [searchValue, setSearchValue] = useState("");

  const handleToggle = (id: string, enabled: boolean) => {
    console.log(`Skill ${id} toggled: ${enabled}`);
    // TODO: 更新后端状态
  };

  const handleSettingsClick = (id: string) => {
    console.log(`Skill ${id} settings clicked`);
    // TODO: 打开设置弹窗
  };

  const filterSkills = (skills: typeof installedSkills) => {
    if (!searchValue.trim()) return skills;
    const query = searchValue.toLowerCase();
    return skills.filter(
      (skill) =>
        skill.name.toLowerCase().includes(query) ||
        skill.description.toLowerCase().includes(query),
    );
  };

  const renderSkillList = (skills: typeof installedSkills) => {
    const filteredSkills = filterSkills(skills);

    if (filteredSkills.length === 0) {
      return (
        <div className="flex items-center justify-center py-12 text-muted-foreground text-sm">
          {searchValue ? "没有找到匹配的技能" : "暂无技能"}
        </div>
      );
    }

    return (
      <div className="flex flex-col gap-2">
        {filteredSkills.map((skill) => (
          <SkillListItem
            key={skill.id}
            {...skill}
            onToggle={handleToggle}
            onSettingsClick={handleSettingsClick}
          />
        ))}
      </div>
    );
  };

  return (
    <Tabs defaultValue="all" className="w-full">
      <div className="flex justify-between gap-6 mb-3 ">
        {/* 搜索框 */}
        <div className="relative max-w-md flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
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

      {/* 全部 Tab */}
      <TabsContent value="all">
        {renderSkillList([...installedSkills, ...availableSkills])}
      </TabsContent>

      {/* 已安装 Tab */}
      <TabsContent value="enabled">
        {renderSkillList(installedSkills)}
      </TabsContent>

      {/* 可用 Tab */}
      <TabsContent value="disabled">
        {renderSkillList(availableSkills)}
      </TabsContent>
    </Tabs>
  );
};

export default InstalledSkillTab;
