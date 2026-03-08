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

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import {
  ArrowUpRight,
  Download,
  ExternalLinkIcon,
  LucideIcon,
} from "lucide-react";
import { FC } from "react";

/**
 * ClawHub 技能卡片数据
 */
export interface ClawHubSkillCardData {
  id: string;
  name: string;
  version: string;
  category: string;
  description: string;
  descriptionZh: string;
  icon: LucideIcon;
  docsUrl?: string;
  installed?: boolean;
}

interface ClawHubSkillCardProps {
  skill: ClawHubSkillCardData;
  onInstall: (id: string) => void;
}

/**
 * ClawHub 技能展示卡片
 */
export const ClawHubSkillCard: FC<ClawHubSkillCardProps> = ({
  skill,
  onInstall,
}) => {
  const Icon = skill.icon;

  /**
   * 打开技能文档链接
   */
  const handleOpenDocs = () => {
    if (!skill.docsUrl) return;
    window.open(skill.docsUrl, "_blank");
  };

  return (
    <section
      className={cn(
        "rounded-lg bg-card p-4",
        "flex flex-1 min-w-80  flex-col gap-4",
      )}
    >
      <div className="flex items-start gap-4 justify-between">
        <div className="flex items-center justify-center size-10 bg-primary/10 rounded-lg shrink-0">
          <Icon className="size-5 text-primary" />
        </div>

        <div className="flex min-w-0 flex-1 flex-col gap-1 items-start">
          <h3 className="truncate text-sm font-medium">{skill.name}</h3>
          <div className="flex flex-wrap items-center gap-2">
            <Badge className="text-[10px] px-1.5 py-0">{skill.version}</Badge>
            <Badge className="text-[10px] px-1.5 py-0 bg-primary/20 text-primary">
              {skill.category}
            </Badge>
          </div>
        </div>
      </div>

      <div className="space-y-1 text-xs text-muted-foreground text-left">
        <p>{skill.description}</p>
        <p>{skill.descriptionZh}</p>
      </div>

      <div className="mt-auto flex items-center gap-4">
        <Button variant="ghost" size="icon-sm" onClick={handleOpenDocs}>
          <ExternalLinkIcon className="size-5" />
        </Button>

        <Button
          className="flex-1 "
          onClick={() => onInstall(skill.id)}
          disabled={skill.installed}
        >
          <Download />
          {skill.installed ? "已安装" : "安装"}
        </Button>
      </div>
    </section>
  );
};

export default ClawHubSkillCard;
