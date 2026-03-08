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
import { Bot, Search, Workflow, Globe } from "lucide-react";
import { FC, useMemo, useState } from "react";
import { ClawHubSkillCard, ClawHubSkillCardData } from "./ClawHubSkillCard";

/**
 * ClawHub 技能市场数据
 */
const clawHubSkills: ClawHubSkillCardData[] = [
  {
    id: "feishu-lark",
    name: "飞书 / Lark",
    version: "v1.1.0",
    category: "频道",
    description: "Feishu/Lark bot channel via WebSocket",
    descriptionZh: "飞书机器人通道插件，支持 WebSocket 连接",
    icon: Globe,
    docsUrl: "https://open.feishu.cn/",
    installed: false,
  },
  {
    id: "workflow-hub",
    name: "Workflow Hub",
    version: "v0.8.3",
    category: "自动化",
    description: "Orchestrate multi-step automation workflows",
    descriptionZh: "编排多步骤自动化流程，提高任务执行效率",
    icon: Workflow,
    docsUrl: "https://example.com/workflow-hub",
    installed: true,
  },
  {
    id: "bot-bridge",
    name: "Bot Bridge",
    version: "v2.0.1",
    category: "机器人",
    description: "Bridge and route messages across bot providers",
    descriptionZh: "跨机器人平台路由消息，统一管理多个适配器",
    icon: Bot,
    docsUrl: "https://example.com/bot-bridge",
    installed: false,
  },
];

/**
 * ClawHub 技能市场面板
 * 展示可从市场安装的技能列表
 */
export const ClawHubTab: FC = () => {
  const [searchValue, setSearchValue] = useState("");

  /**
   * 处理安装动作
   */
  const handleInstall = (id: string) => {
    console.log(`Installing skill: ${id}`);
    // TODO: 调用安装 API
  };

  /**
   * 根据搜索词过滤技能列表
   */
  const filteredSkills = useMemo(() => {
    const query = searchValue.trim().toLowerCase();
    if (!query) return clawHubSkills;

    return clawHubSkills.filter((skill) => {
      return (
        skill.name.toLowerCase().includes(query) ||
        skill.description.toLowerCase().includes(query) ||
        skill.descriptionZh.toLowerCase().includes(query) ||
        skill.category.toLowerCase().includes(query)
      );
    });
  }, [searchValue]);

  if (filteredSkills.length === 0) {
    return (
      <div className="flex flex-col gap-4">
        <div className="relative max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="搜索技能..."
            value={searchValue}
            onChange={(e) => setSearchValue(e.target.value)}
            className="pl-9"
          />
        </div>

        <div className="flex items-center justify-center py-12 text-muted-foreground text-sm">
          没有找到匹配的技能
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="relative w-full">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
        <Input
          placeholder="搜索ClawHub技能..."
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          className="pl-9"
        />
      </div>

      <div className="flex flex-wrap gap-4">
        {filteredSkills.map((skill) => (
          <ClawHubSkillCard
            key={skill.id}
            skill={skill}
            onInstall={handleInstall}
          />
        ))}
      </div>
    </div>
  );
};

export default ClawHubTab;
