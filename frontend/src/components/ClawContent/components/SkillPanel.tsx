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
import { SparklesIcon, PlugIcon, StoreIcon } from "lucide-react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import InstalledSkillTab from "./InstalledSkillTab";
import PluginTab from "./PluginTab";
import ClawHubTab from "./ClawHubTab";
import { PanelHeader } from "./PanelHeader";

/**
 * 技能管理面板组件
 * 显示所有可用技能的配置和管理
 */
export const SkillPanel: FC = () => {
  const handleRefresh = () => {
    console.log("Refresh skill list");
    // TODO: 刷新技能列表
  };

  return (
    <div className="h-full w-full overflow-auto p-6">
      {/* 标题区域 */}
      <PanelHeader
        className="mb-6"
        title="技能中心"
        description="配置和管理 OpenClaw 的技能模块"
        actions={
          <Button variant="secondary" size="sm" onClick={handleRefresh}>
            刷新列表
          </Button>
        }
      />

      <Tabs defaultValue="installed">
        <TabsList variant="line" className="mb-3">
          <TabsTrigger value="installed">
            <SparklesIcon />
            已安装技能
            <Badge variant="secondary">11</Badge>
          </TabsTrigger>
          <TabsTrigger value="plugins">
            <PlugIcon />
            插件
            <Badge variant="secondary">11</Badge>
          </TabsTrigger>
          <TabsTrigger value="clawhub">
            <StoreIcon />
            ClawHub
            <Badge variant="secondary">11</Badge>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="installed">
          <InstalledSkillTab />
        </TabsContent>
        <TabsContent value="plugins">
          <PluginTab />
        </TabsContent>
        <TabsContent value="clawhub">
          <ClawHubTab />
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default SkillPanel;
