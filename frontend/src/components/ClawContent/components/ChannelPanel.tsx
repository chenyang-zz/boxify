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

import { FC, useCallback, useEffect, useMemo, useState } from "react";
import type { ManagedChannel as Channel } from "../types";
import { ChannelList } from "./ChannelList";
import { ChannelConfigPanel } from "./ChannelConfigPanel";
import { RadioIcon } from "lucide-react";
import { PanelHeader } from "./PanelHeader";
import { callWails } from "@/lib/utils";
import { ClawService } from "@wails/service";
import { channelDefinitions } from "../constants/channel-definitions";

type JsonObject = Record<string, unknown>;

/**
 * 将未知值转换为对象，避免访问配置字段时出现运行时错误。
 */
function toJsonObject(value: unknown): JsonObject {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as JsonObject;
  }
  return {};
}

/**
 * 判断配置值是否包含可视为“已配置”的内容。
 */
function hasConfiguredValue(value: unknown): boolean {
  if (value === null || value === undefined) {
    return false;
  }
  if (typeof value === "string") {
    return value.trim().length > 0;
  }
  if (typeof value === "number") {
    return true;
  }
  if (typeof value === "boolean") {
    return value;
  }
  if (Array.isArray(value)) {
    return value.some(hasConfiguredValue);
  }
  if (typeof value === "object") {
    return Object.values(value as JsonObject).some(hasConfiguredValue);
  }
  return false;
}

/**
 * 根据配置内容计算频道状态（已启用/已配置/未配置）。
 */
function resolveChannelStatus(config: JsonObject): Channel["status"] {
  if (config.enabled === true) {
    return "enabled";
  }
  const hasConfig = Object.entries(config).some(
    ([key, value]) => key !== "enabled" && hasConfiguredValue(value),
  );
  if (hasConfig) {
    return "configured";
  }
  return "unconfigured";
}

/**
 * 按主 ID + 别名顺序返回第一个存在的配置对象。
 */
function resolveChannelConfig(
  sourceMap: JsonObject,
  id: string,
  aliases: string[] = [],
): JsonObject {
  const candidates = [id, ...aliases];
  for (const candidate of candidates) {
    const config = toJsonObject(sourceMap[candidate]);
    if (Object.keys(config).length > 0) {
      return config;
    }
  }
  return {};
}

/**
 * 获取频道原始配置与命中的配置 ID（用于保存时复用原 key）。
 */
function resolveChannelConfigWithSource(
  sourceMap: JsonObject,
  id: string,
  aliases: string[] = [],
): { sourceId: string; config: JsonObject } {
  const candidates = [id, ...aliases];
  let best: { sourceId: string; config: JsonObject; score: number } | null =
    null;
  for (const candidate of candidates) {
    const config = toJsonObject(sourceMap[candidate]);
    if (Object.keys(config).length === 0) {
      continue;
    }
    // 优先级：enabled=true > 配置字段数量（排除 enabled）
    const nonEnabledCount = Object.keys(config).filter(
      (key) => key !== "enabled",
    ).length;
    const score = (config.enabled === true ? 1000 : 0) + nonEnabledCount;
    if (!best || score > best.score) {
      best = { sourceId: candidate, config, score };
    }
  }
  if (best) {
    return { sourceId: best.sourceId, config: best.config };
  }
  return { sourceId: id, config: {} };
}

/**
 * 聚合主 ID 与别名的配置，用于状态计算（enabled 取并集）。
 */
function mergeChannelConfigs(
  sourceMap: JsonObject,
  id: string,
  aliases: string[] = [],
): JsonObject {
  const merged: JsonObject = {};
  const candidates = [id, ...aliases];
  for (const candidate of candidates) {
    const config = toJsonObject(sourceMap[candidate]);
    Object.assign(merged, config);
  }
  const hasEnabled = candidates.some(
    (candidate) => toJsonObject(sourceMap[candidate]).enabled === true,
  );
  if (hasEnabled) {
    merged.enabled = true;
  }
  return merged;
}

/**
 * 将后端返回的频道 ID 归一化为前端定义 ID。
 */
function normalizeChannelId(rawId: string, fallbackId: string): string {
  const normalized = rawId.trim();
  if (!normalized) {
    return fallbackId;
  }
  const matched = channelDefinitions.find(
    (item) => item.id === normalized || item.aliases?.includes(normalized),
  );
  return matched?.id ?? normalized;
}

/**
 * 默认选中策略：优先已启用频道，其次飞书，最后首项。
 */
function resolveDefaultSelectedChannel(channels: Channel[]): string {
  const firstEnabled = channels.find((item) => item.status === "enabled");
  if (firstEnabled) {
    return firstEnabled.id;
  }
  const feishu = channels.find((item) => item.id === "feishu");
  if (feishu) {
    return feishu.id;
  }
  return channels[0]?.id ?? "";
}

/**
 * 频道配置面板组件
 * 左侧为频道列表，右侧为配置详情
 */
export const ChannelPanel: FC = () => {
  const [channels, setChannels] = useState<Channel[]>([]);
  const [selectedChannelId, setSelectedChannelId] = useState("");

  /**
   * 拉取 openclaw.json 的 channels/plugins 配置并映射为频道列表。
   */
  const refreshChannels = useCallback(async () => {
    const result = await callWails(ClawService.GetChannels);
    const channelConfigMap = toJsonObject(result.channels);
    const pluginConfigMap = toJsonObject(toJsonObject(result.plugins).entries);
    const nextChannels: Channel[] = channelDefinitions.map((item) => {
      const channelSource = resolveChannelConfigWithSource(
        channelConfigMap,
        item.id,
        item.aliases,
      );
      const channelMergedConfig = mergeChannelConfigs(
        channelConfigMap,
        item.id,
        item.aliases,
      );
      const pluginMergedConfig =
        item.type === "plugin"
          ? mergeChannelConfigs(pluginConfigMap, item.id, item.aliases)
          : {};
      const mergedConfig = {
        ...channelMergedConfig,
        ...pluginMergedConfig,
        enabled:
          channelMergedConfig.enabled === true ||
          pluginMergedConfig.enabled === true,
      };
      const feishuVariant =
        item.id === "feishu"
          ? toJsonObject(pluginConfigMap["feishu-openclaw-plugin"]).enabled ===
            true
            ? "official"
            : "clawteam"
          : undefined;
      return {
        id: item.id,
        name: item.name,
        managedBy: "由网关管理",
        description: item.description,
        type: item.type,
        icon: <RadioIcon />,
        config: channelSource.config,
        saveTargetId: channelSource.sourceId,
        feishuVariant,
        status: resolveChannelStatus(
          Object.keys(mergedConfig).length > 0
            ? mergedConfig
            : resolveChannelConfig(channelConfigMap, item.id, item.aliases),
        ),
      };
    });
    setChannels(nextChannels);
    setSelectedChannelId((prevRaw) => {
      const prev = normalizeChannelId(prevRaw, "");
      return prev && nextChannels.some((item) => item.id === prev)
        ? prev
        : resolveDefaultSelectedChannel(nextChannels);
    });
  }, []);

  useEffect(() => {
    void refreshChannels();
  }, [refreshChannels]);

  const selectedChannel = useMemo(
    () => channels.find((item) => item.id === selectedChannelId),
    [channels, selectedChannelId],
  );

  return (
    <div className="h-full w-full overflow-auto p-6">
      {/* 标题区域 */}
      <PanelHeader
        className="mb-6"
        title="频道管理"
        description="配置和管理所有消息频道"
      />
      <div className="flex gap-6">
        {/* 左侧频道列表 */}
        <ChannelList
          channels={channels}
          selectedChannelId={selectedChannelId}
          onChannelSelect={setSelectedChannelId}
        />

        {/* 右侧配置详情 */}
        <ChannelConfigPanel
          channel={selectedChannel}
          onSaved={refreshChannels}
        />
      </div>
    </div>
  );
};

export default ChannelPanel;
