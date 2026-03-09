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

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { callWails } from "@/lib/utils";
import { FC, useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import {
  Bot,
  Boxes,
  BrainCircuit,
  Code2,
  Cpu,
  Globe,
  Package,
  PlugIcon,
  SparklesIcon,
  StoreIcon,
  Workflow,
} from "lucide-react";
import {
  GetSkillPlugins,
  GetSkills,
  GetPluginList,
  InstallPlugin,
  ToggleSkill,
  TogglePlugin,
} from "../../../../bindings/github.com/chenyang-zz/boxify/internal/service/clawservice";
import ClawHubTab from "./ClawHubTab";
import type { ClawHubSkillCardData } from "./ClawHubSkillCard";
import InstalledSkillTab from "./InstalledSkillTab";
import { PanelHeader } from "./PanelHeader";
import PluginTab from "./PluginTab";
import type { SkillListItemProps } from "./SkillListItem";

type SkillListData = Omit<SkillListItemProps, "onToggle" | "onSettingsClick">;

interface InstalledSkillData {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  source?: string;
  path?: string;
  version?: string;
}

interface InstalledPluginData {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  author?: string;
  source?: string;
  path?: string;
  dir?: string;
  installedAt?: string;
  version?: string;
  category?: string;
  tags?: string[];
}

interface RegistryPluginData {
  id: string;
  name: string;
  version: string;
  description: string;
  category?: string;
  homepage?: string;
  repository?: string;
  readme?: string;
}

/** 根据插件分类推断默认图标。 */
function resolvePluginIcon(category?: string, tags?: string[]) {
  const hint = `${category ?? ""} ${(tags ?? []).join(" ")}`.toLowerCase();
  if (hint.includes("workflow") || hint.includes("自动化")) return Workflow;
  if (hint.includes("bot") || hint.includes("机器人")) return Bot;
  if (hint.includes("ai")) return BrainCircuit;
  if (hint.includes("code") || hint.includes("开发")) return Code2;
  if (hint.includes("web") || hint.includes("search") || hint.includes("联网")) return Globe;
  if (hint.includes("tool") || hint.includes("工具")) return Cpu;
  if (hint.includes("message") || hint.includes("channel") || hint.includes("频道")) return PlugIcon;
  if (hint.includes("plugin") || hint.includes("插件")) return Package;
  return Boxes;
}

/** 将已安装技能转换为列表项。 */
function toSkillListData(plugin: InstalledSkillData): SkillListData {
  return {
    id: plugin.id,
    name: plugin.name || plugin.id,
    description: plugin.description || "暂无描述",
    enabled: plugin.enabled,
    icon: resolvePluginIcon(undefined, [plugin.source ?? "", plugin.path ?? ""]),
  };
}

/** 将插件列表转换为列表项。 */
function toPluginListData(plugin: InstalledPluginData): SkillListData {
  return {
    id: plugin.id,
    name: plugin.name || plugin.id,
    description: plugin.description || "暂无描述",
    enabled: plugin.enabled,
    icon: resolvePluginIcon(plugin.category, plugin.tags),
  };
}

/** 将仓库插件转换为市场卡片。 */
function toClawHubCard(
  plugin: RegistryPluginData,
  installedIDs: Set<string>,
): ClawHubSkillCardData {
  return {
    id: plugin.id,
    name: plugin.name || plugin.id,
    version: plugin.version || "-",
    category: plugin.category || "插件",
    description: plugin.description || "暂无描述",
    descriptionZh: plugin.description || "暂无描述",
    icon: resolvePluginIcon(plugin.category),
    docsUrl: plugin.homepage || plugin.repository || plugin.readme,
    installed: installedIDs.has(plugin.id),
  };
}

/**
 * 技能管理面板组件
 * 显示所有可用技能的配置和管理
 */
export const SkillPanel: FC = () => {
  const [installedSkills, setInstalledSkills] = useState<InstalledSkillData[]>([]);
  const [installedPlugins, setInstalledPlugins] = useState<InstalledPluginData[]>([]);
  const [registryPlugins, setRegistryPlugins] = useState<RegistryPluginData[]>([]);
  const [activeTab, setActiveTab] = useState("installed");
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  /** 拉取已安装技能列表。 */
  const loadInstalledSkills = useCallback(async () => {
    const skillsResult = await callWails(GetSkills);
    setInstalledSkills(
      ((skillsResult.skills ?? []).filter(Boolean) as InstalledSkillData[]).slice(),
    );
  }, []);

  /** 拉取已安装插件列表。 */
  const loadInstalledPlugins = useCallback(async () => {
    const skillPluginsResult = await callWails(GetSkillPlugins);
    setInstalledPlugins(
      ((skillPluginsResult.plugins ?? []).filter(Boolean) as InstalledPluginData[]).slice(),
    );
  }, []);

  /** 拉取 ClawHub 仓库插件列表。 */
  const loadRegistryPlugins = useCallback(async () => {
    const pluginListResult = await callWails(GetPluginList);
    setRegistryPlugins(
      ((pluginListResult.registry ?? []).filter(Boolean) as RegistryPluginData[]).slice(),
    );
  }, []);

  /** 首次进入时全量拉取技能中心数据。 */
  const loadAllSkillData = useCallback(async () => {
    setLoading(true);
    try {
      await Promise.all([
        loadInstalledSkills(),
        loadInstalledPlugins(),
        loadRegistryPlugins(),
      ]);
    } finally {
      setLoading(false);
    }
  }, [loadInstalledPlugins, loadInstalledSkills, loadRegistryPlugins]);

  /** 按当前页签刷新对应列表。 */
  const refreshCurrentTab = useCallback(async () => {
    setRefreshing(true);
    try {
      if (activeTab === "installed") {
        await loadInstalledSkills();
        toast.success("技能列表已刷新", {
          style: {
            textAlign: "left",
          },
        });
        return;
      }

      if (activeTab === "plugins") {
        await loadInstalledPlugins();
        toast.success("插件列表已刷新", {
          style: {
            textAlign: "left",
          },
        });
        return;
      }

      await loadRegistryPlugins();
      toast.success("ClawHub 列表已刷新", {
        style: {
          textAlign: "left",
        },
      });
    } finally {
      setRefreshing(false);
    }
  }, [activeTab, loadInstalledPlugins, loadInstalledSkills, loadRegistryPlugins]);

  useEffect(() => {
    void loadAllSkillData();
  }, [loadAllSkillData]);

  /** 切换技能启用状态后，同步本地展示。 */
  const handleToggleSkill = useCallback(async (id: string, enabled: boolean) => {
    await callWails(ToggleSkill, id, enabled);
    setInstalledSkills((current) =>
      current.map((skill) =>
        skill.id === id ? { ...skill, enabled } : skill,
      ),
    );
    toast.success(enabled ? "技能已启用" : "技能已禁用", {
      style: {
        textAlign: "left",
      },
    });
  }, []);

  /** 切换插件启用状态后，同步本地展示。 */
  const handleTogglePlugin = useCallback(async (id: string, enabled: boolean) => {
    await callWails(TogglePlugin, id, enabled);
    setInstalledPlugins((current) =>
      current.map((plugin) =>
        plugin.id === id ? { ...plugin, enabled } : plugin,
      ),
    );
    toast.success(enabled ? "插件已启用" : "插件已禁用", {
      style: {
        textAlign: "left",
      },
    });
  }, []);

  /** 安装插件后刷新已安装与仓库视图。 */
  const handleInstallPlugin = useCallback(
    async (id: string) => {
      await callWails(InstallPlugin, id, "");
      toast.success("插件安装成功", {
        style: {
          textAlign: "left",
        },
      });
      await Promise.all([loadInstalledPlugins(), loadRegistryPlugins()]);
    },
    [loadInstalledPlugins, loadRegistryPlugins],
  );

  const scannedSkillItems = useMemo(
    () => installedSkills.map(toSkillListData),
    [installedSkills],
  );
  const installedIDs = useMemo(
    () => new Set(installedPlugins.map((plugin) => plugin.id)),
    [installedPlugins],
  );
  const installedPluginItems = useMemo(
    () => installedPlugins.map(toPluginListData),
    [installedPlugins],
  );
  const clawHubItems = useMemo(
    () => registryPlugins.map((plugin) => toClawHubCard(plugin, installedIDs)),
    [installedIDs, registryPlugins],
  );

  return (
    <div className="h-full w-full overflow-auto p-6">
      <PanelHeader
        className="mb-6"
        title="技能中心"
        description="配置和管理 OpenClaw 的技能模块"
        actions={
          <Button
            variant="secondary"
            size="sm"
            onClick={() => void refreshCurrentTab()}
            disabled={loading || refreshing}
          >
            {refreshing ? "刷新中..." : "刷新列表"}
          </Button>
        }
      />

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList variant="line" className="mb-3">
          <TabsTrigger value="installed">
            <SparklesIcon />
            已安装技能
            <Badge variant="secondary">{scannedSkillItems.length}</Badge>
          </TabsTrigger>
          <TabsTrigger value="plugins">
            <PlugIcon />
            插件
            <Badge variant="secondary">{installedPluginItems.length}</Badge>
          </TabsTrigger>
          <TabsTrigger value="clawhub">
            <StoreIcon />
            ClawHub
            <Badge variant="secondary">{clawHubItems.length}</Badge>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="installed">
          <InstalledSkillTab
            skills={scannedSkillItems}
            loading={loading}
            onToggle={handleToggleSkill}
          />
        </TabsContent>
        <TabsContent value="plugins">
          <PluginTab
            plugins={installedPluginItems}
            loading={loading}
            onToggle={handleTogglePlugin}
          />
        </TabsContent>
        <TabsContent value="clawhub">
          <ClawHubTab
            skills={clawHubItems}
            loading={loading}
            onInstall={handleInstallPlugin}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default SkillPanel;
