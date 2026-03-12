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
import { Check, Eye, EyeOff, Loader2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { callWails, cn } from "@/lib/utils";
import CardItem from "./CardItem";
import { PanelHeader } from "./PanelHeader";
import type { ManagedChannel as Channel } from "../types";
import { ClawService } from "@wails/service";
import { toast } from "sonner";
import { Checkbox } from "@/components/ui/checkbox";

type JsonObject = Record<string, unknown>;

/**
 * 将未知值转换为对象，避免访问配置字段时报错。
 */
function toJsonObject(value: unknown): JsonObject {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as JsonObject;
  }
  return {};
}

/**
 * 从配置中读取字符串字段，支持多个别名 key。
 */
function readString(config: JsonObject, keys: string[]): string {
  for (const key of keys) {
    const value = config[key];
    if (typeof value === "string" && value.trim().length > 0) {
      return value;
    }
  }
  return "";
}

/**
 * 从配置中读取布尔字段，支持多个别名 key。
 */
function readBoolean(
  config: JsonObject,
  keys: string[],
  fallback: boolean,
): boolean {
  for (const key of keys) {
    const value = config[key];
    if (typeof value === "boolean") {
      return value;
    }
  }
  return fallback;
}

export interface ChannelConfigPanelProps {
  channel?: Channel;
  onSaved?: () => Promise<void> | void;
}

interface ChatChannelInfoState {
  channelInboxURL: string;
  sharedToken: string;
}

/**
 * 频道配置详情组件
 */
export const ChannelConfigPanel: FC<ChannelConfigPanelProps> = ({
  channel,
  onSaved,
}) => {
  const feishuOfficialId = "feishu-openclaw-plugin";
  const feishuClawteamId = "feishu";
  const boxifyPluginId = "boxify-channel";
  const [isSaving, setIsSaving] = useState(false);
  const [isTogglingEnabled, setIsTogglingEnabled] = useState(false);
  const isConfigured =
    channel?.status === "configured" || channel?.status === "enabled";
  const isFeishuChannel = channel?.id === "feishu";
  const isBoxifyChannel = channel?.id === "boxify";
  const channelConfig = useMemo(
    () => toJsonObject(channel?.config),
    [channel?.config],
  );
  const [chatChannelInfo, setChatChannelInfo] =
    useState<ChatChannelInfoState | null>(null);
  const [isLoadingChannelInfo, setIsLoadingChannelInfo] = useState(false);
  const [showBoxifyToken, setShowBoxifyToken] = useState(false);

  // 飞书版本选择
  const [feishuVariant, setFeishuVariant] = useState<"official" | "clawteam">(
    "official",
  );

  // 启用状态
  const [isEnabled, setIsEnabled] = useState(isConfigured);

  // 飞书功能开关
  const [streaming, setStreaming] = useState(false);
  const [threadSession, setThreadSession] = useState(false);
  const [replyInThread, setReplyInThread] = useState(false);
  const [typingIndicator, setTypingIndicator] = useState(false);
  const [resolveSenderNames, setResolveSenderNames] = useState(false);
  const [dynamicAgentCreation, setDynamicAgentCreation] = useState(false);

  // 输入框状态
  const [appId, setAppId] = useState("");
  const [appSecret, setAppSecret] = useState("");
  const [boxifyDefaultAgent, setBoxifyDefaultAgent] = useState("main");
  const [showSecret, setShowSecret] = useState(false);

  // 验证状态
  const isAppIdValid = appId.length > 0;
  const isAppSecretValid = appSecret.length > 0;

  /**
   * 切换频道时重置默认表单状态，避免不同频道间相互污染。
   */
  useEffect(() => {
    const nextVariant =
      channel?.feishuVariant === "official" ? "official" : "clawteam";
    const nextEnabled = readBoolean(channelConfig, ["enabled"], isConfigured);

    setFeishuVariant(nextVariant);
    setIsEnabled(nextEnabled);
    setStreaming(readBoolean(channelConfig, ["streaming"], false));
    setThreadSession(readBoolean(channelConfig, ["threadSession"], false));
    setReplyInThread(readBoolean(channelConfig, ["replyInThread"], false));
    setTypingIndicator(readBoolean(channelConfig, ["typingIndicator"], false));
    setResolveSenderNames(
      readBoolean(channelConfig, ["resolveSenderNames"], false),
    );
    setDynamicAgentCreation(
      readBoolean(channelConfig, ["dynamicAgentCreation"], false),
    );
    setAppId(readString(channelConfig, ["appId", "app_id", "appID"]));
    setAppSecret(
      readString(channelConfig, ["appSecret", "app_secret", "appKey"]),
    );
    setBoxifyDefaultAgent(readString(channelConfig, ["defaultAgent"]) || "main");
    setShowSecret(false);
    setShowBoxifyToken(false);
  }, [
    channel?.feishuVariant,
    channel?.id,
    channel?.saveTargetId,
    channelConfig,
    isConfigured,
  ]);

  /**
   * 拉取 Boxify 原生 channel inbox 的联调信息。
   */
  const loadChatChannelInfo = useCallback(async () => {
    if (!isBoxifyChannel) {
      setChatChannelInfo(null);
      return;
    }

    setIsLoadingChannelInfo(true);
    try {
      const result = await callWails(ClawService.GetChatChannelInfo);
      const data = result.data;
      setChatChannelInfo(
        data
          ? {
              channelInboxURL: data.channelInboxURL ?? "",
              sharedToken: data.sharedToken ?? "",
            }
          : null,
      );
    } finally {
      setIsLoadingChannelInfo(false);
    }
  }, [isBoxifyChannel]);

  /**
   * 切到 Boxify 频道时自动刷新一次联调信息。
   */
  useEffect(() => {
    void loadChatChannelInfo();
  }, [loadChatChannelInfo]);

  /**
   * 保存当前频道配置到 openclaw.json。
   */
  const handleSave = useCallback(async () => {
    if (!channel) {
      return;
    }

    if (isFeishuChannel && (!appId.trim() || !appSecret.trim())) {
      toast.error("保存失败", {
        description: "请先填写 App ID 和 App Secret",
        style: { textAlign: "left" },
      });
      return;
    }

    const payload: JsonObject = { enabled: isEnabled };

    if (isFeishuChannel) {
      payload.appId = appId.trim();
      payload.appSecret = appSecret.trim();
      payload.streaming = streaming;
      payload.threadSession = threadSession;
      payload.replyInThread = replyInThread;
      payload.typingIndicator = typingIndicator;
      payload.resolveSenderNames = resolveSenderNames;
      payload.dynamicAgentCreation = dynamicAgentCreation;
    } else if (isBoxifyChannel) {
      payload.listenUrl =
        chatChannelInfo?.channelInboxURL ||
        readString(channelConfig, ["listenUrl"]) ||
        "";
      payload.sharedToken =
        chatChannelInfo?.sharedToken ||
        readString(channelConfig, ["sharedToken"]) ||
        "";
      payload.defaultAgent = boxifyDefaultAgent.trim() || "main";
    }

    const saveId = isFeishuChannel
      ? feishuVariant === "official"
        ? feishuOfficialId
        : feishuClawteamId
      : channel.saveTargetId?.trim() || channel.id;
    const inactiveFeishuId =
      feishuVariant === "official" ? feishuClawteamId : feishuOfficialId;
    setIsSaving(true);
    try {
      if (isFeishuChannel) {
        await callWails(ClawService.SaveChannel, "feishu", payload);
        await callWails(ClawService.SavePlugin, saveId, {
          enabled: isEnabled,
        });
        await callWails(ClawService.SavePlugin, inactiveFeishuId, {
          enabled: false,
        });
      } else if (channel.type === "plugin") {
        await callWails(ClawService.SaveChannel, saveId, payload);
        await callWails(
          ClawService.SavePlugin,
          isBoxifyChannel ? boxifyPluginId : saveId,
          {
          enabled: isEnabled,
          },
        );
      } else {
        await callWails(ClawService.SaveChannel, saveId, payload);
      }
      toast.success("保存成功", {
        description: `${channel.name} 配置已保存`,
        style: { textAlign: "left" },
      });
      await onSaved?.();
    } finally {
      setIsSaving(false);
    }
  }, [
    appId,
    appSecret,
    channel,
    dynamicAgentCreation,
    feishuVariant,
    boxifyDefaultAgent,
    channelConfig,
    chatChannelInfo?.channelInboxURL,
    chatChannelInfo?.sharedToken,
    isEnabled,
    isBoxifyChannel,
    isFeishuChannel,
    onSaved,
    replyInThread,
    resolveSenderNames,
    streaming,
    threadSession,
    typingIndicator,
  ]);

  /**
   * 启用状态切换后立即持久化，其他字段仍通过“保存配置”手动提交。
   */
  const handleEnabledChange = useCallback(
    async (nextEnabled: boolean) => {
      if (!channel) {
        return;
      }
      setIsEnabled(nextEnabled);
      setIsTogglingEnabled(true);
      try {
        if (isFeishuChannel) {
          const activeFeishuId =
            feishuVariant === "official" ? feishuOfficialId : feishuClawteamId;
          const inactiveFeishuId =
            feishuVariant === "official" ? feishuClawteamId : feishuOfficialId;
          await callWails(ClawService.ToggleChannel, "feishu", nextEnabled);
          await callWails(ClawService.SavePlugin, activeFeishuId, {
            enabled: nextEnabled,
          });
          await callWails(ClawService.SavePlugin, inactiveFeishuId, {
            enabled: false,
          });
        } else if (channel.type === "plugin") {
          const targetId = channel.saveTargetId?.trim() || channel.id;
          await callWails(ClawService.ToggleChannel, targetId, nextEnabled);
          await callWails(
            ClawService.SavePlugin,
            isBoxifyChannel ? boxifyPluginId : targetId,
            {
            enabled: nextEnabled,
            },
          );
        } else {
          const targetId = channel.saveTargetId?.trim() || channel.id;
          await callWails(ClawService.ToggleChannel, targetId, nextEnabled);
        }
        await onSaved?.();
      } catch {
        setIsEnabled((prev) => !prev);
      } finally {
        setIsTogglingEnabled(false);
      }
    },
    [channel, feishuVariant, isBoxifyChannel, isFeishuChannel, onSaved],
  );

  if (!channel) {
    return (
      <div className="h-full overflow-auto flex-1">
        <div className="text-sm text-muted-foreground mt-4">暂无可配置频道</div>
      </div>
    );
  }

  return (
    <div className="h-full overflow-auto flex-1">
      <div className="space-y-6">
        {/* 标题区域 */}
        <PanelHeader
          align="start"
          className="mt-4"
          title={
            <div className="flex items-center gap-2">
              <span>{channel?.name} 配置</span>
              {isConfigured && (
                <Badge
                  variant="outline"
                  className="bg-emerald-500/10 text-emerald-500 border-emerald-500/20"
                >
                  已配置
                </Badge>
              )}
            </div>
          }
          titleClassName="text-base font-semibold"
          description={`配置 ${channel?.name} 机器人以接收和发送消息`}
          actions={
            <div className="flex items-center gap-2">
              <span
                className={cn(
                  "text-sm",
                  isEnabled ? "text-emerald-500" : "text-muted-foreground",
                )}
              >
                {isEnabled ? "启用中" : "已禁用"}
              </span>
              <Switch
                checked={isEnabled}
                disabled={isSaving || isTogglingEnabled}
                onCheckedChange={(checked) => void handleEnabledChange(checked)}
              />
            </div>
          }
        />

        {isFeishuChannel ? (
          <>
            {/* 飞书版本选择 */}
            <div className="text-left">
              <label className="text-sm font-medium">当前飞书实现</label>
              <div className="flex flex-col gap-3 mt-2">
                <label
                  className={cn(
                    "flex items-start gap-3 p-3 rounded-lg border-2 cursor-pointer transition-all",
                    feishuVariant === "official"
                      ? "border-primary bg-primary/5"
                      : "border-border hover:border-primary/50",
                  )}
                >
                  <Checkbox
                    id="terms-checkbox"
                    checked={feishuVariant === "official"}
                    onCheckedChange={() => setFeishuVariant("official")}
                    name="terms-checkbox"
                  />
                  <div>
                    <div className="text-sm font-medium">飞书官方版</div>
                    <div className="text-[11px] text-muted-foreground mt-0.5">
                      支持流式卡片、话题独立上下文
                    </div>
                  </div>
                </label>
                <label
                  className={cn(
                    "flex items-start gap-3 p-3 rounded-lg border-2 cursor-pointer transition-all",
                    feishuVariant === "clawteam"
                      ? "border-primary bg-primary/5"
                      : "border-border hover:border-primary/50",
                  )}
                >
                  <Checkbox
                    id="clawteam-variant"
                    checked={feishuVariant === "clawteam"}
                    onCheckedChange={() => setFeishuVariant("clawteam")}
                    name="feishu-variant"
                    value="clawteam"
                  />
                  <div>
                    <div className="text-sm font-medium">ClawTeam 社区版</div>
                    <div className="text-[11px] text-muted-foreground mt-0.5">
                      支持话题内回复、输入中提示等
                    </div>
                  </div>
                </label>
              </div>
            </div>

            {/* 配置输入区域 */}
            <div className="space-y-4 text-left">
              <div>
                <label className="text-sm font-medium">App ID</label>
                <div className="relative mt-2">
                  <Input
                    value={appId}
                    onChange={(e) => setAppId(e.target.value)}
                    placeholder="请输入 App ID"
                    className="pr-10"
                  />
                  {isAppIdValid && (
                    <Check className="absolute right-3 top-1/2 -translate-y-1/2 size-4 text-emerald-500" />
                  )}
                </div>
              </div>

              <div>
                <label className="text-sm font-medium">App Secret</label>
                <div className="relative mt-2">
                  <Input
                    type={showSecret ? "text" : "password"}
                    value={appSecret}
                    onChange={(e) => setAppSecret(e.target.value)}
                    placeholder="请输入 App Secret"
                    className="pr-16"
                  />
                  <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
                    {isAppSecretValid && (
                      <Check className="size-4 text-emerald-500" />
                    )}
                    <button
                      type="button"
                      onClick={() => setShowSecret(!showSecret)}
                      className="p-1 hover:bg-accent rounded transition-colors"
                    >
                      {showSecret ? (
                        <EyeOff className="size-4 text-muted-foreground" />
                      ) : (
                        <Eye className="size-4 text-muted-foreground" />
                      )}
                    </button>
                  </div>
                </div>
              </div>
            </div>

            {/* 功能开关区域 */}
            <div className="text-left">
              <h2 className="text-sm font-semibold">功能设置</h2>
              <div className="space-y-4 mt-2">
                {/* 图片卡片输出 */}
                <CardItem
                  label="流式卡片输出"
                  description="仅飞书官方版支持，开启后回复以流式卡片形式呈现"
                  action={
                    <Switch
                      checked={streaming}
                      onCheckedChange={setStreaming}
                    />
                  }
                />

                <CardItem
                  label="话题独立上下文"
                  description="仅飞书官方版支持，每个话题拥有独立会话并可并行"
                  action={
                    <Switch
                      checked={threadSession}
                      onCheckedChange={setThreadSession}
                    />
                  }
                />

                <CardItem
                  label="话题内回复"
                  description="仅 ClawTeam 版支持，优先在话题内回复"
                  action={
                    <Switch
                      checked={replyInThread}
                      onCheckedChange={setReplyInThread}
                    />
                  }
                />

                <CardItem
                  label="输入中提示"
                  description="仅 ClawTeam 版支持"
                  action={
                    <Switch
                      checked={typingIndicator}
                      onCheckedChange={setTypingIndicator}
                    />
                  }
                />

                <CardItem
                  label="解析发送者名称"
                  description="仅 ClawTeam 版支持，自动解析飞书用户显示名"
                  action={
                    <Switch
                      checked={resolveSenderNames}
                      onCheckedChange={setResolveSenderNames}
                    />
                  }
                />

                <CardItem
                  label="动态创建 Agent"
                  description="仅 ClawTeam 版支持，按场景动态创建 Agent"
                  action={
                    <Switch
                      checked={dynamicAgentCreation}
                      onCheckedChange={setDynamicAgentCreation}
                    />
                  }
                />
              </div>
            </div>
          </>
        ) : isBoxifyChannel ? (
          <div className="space-y-4 text-left">
            <div>
              <label className="text-sm font-medium">Channel URL</label>
              <div className="relative mt-2">
                <Input
                  readOnly
                  value={chatChannelInfo?.channelInboxURL ?? ""}
                  placeholder={
                    isLoadingChannelInfo ? "正在读取 Channel URL" : "暂无 Channel URL"
                  }
                  className="pr-10"
                />
                {chatChannelInfo?.channelInboxURL ? (
                  <Check className="absolute right-3 top-1/2 -translate-y-1/2 size-4 text-emerald-500" />
                ) : null}
              </div>
            </div>

            <div>
              <label className="text-sm font-medium">Secret Token</label>
              <div className="relative mt-2">
                <Input
                  readOnly
                  type={showBoxifyToken ? "text" : "password"}
                  value={chatChannelInfo?.sharedToken ?? ""}
                  placeholder={
                    isLoadingChannelInfo ? "正在读取 Secret Token" : "暂无 Secret Token"
                  }
                  className="pr-16"
                />
                <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
                  {chatChannelInfo?.sharedToken ? (
                    <Check className="size-4 text-emerald-500" />
                  ) : null}
                  <button
                    type="button"
                    onClick={() => setShowBoxifyToken(!showBoxifyToken)}
                    className="p-1 hover:bg-accent rounded transition-colors"
                  >
                    {showBoxifyToken ? (
                      <EyeOff className="size-4 text-muted-foreground" />
                    ) : (
                      <Eye className="size-4 text-muted-foreground" />
                    )}
                  </button>
                </div>
              </div>
            </div>
          </div>
        ) : (
          <div className="text-sm text-muted-foreground text-left">
            当前频道的详细配置项将逐步接入，现阶段可先保存启用状态。
          </div>
        )}

        <div className="pt-2">
          <Button
            className="w-full"
            onClick={() => void handleSave()}
            disabled={isSaving}
          >
            {isSaving ? <Loader2 className="size-4 animate-spin" /> : null}
            {isSaving ? "保存中..." : "保存配置"}
          </Button>
        </div>
      </div>
    </div>
  );
};

export default ChannelConfigPanel;
