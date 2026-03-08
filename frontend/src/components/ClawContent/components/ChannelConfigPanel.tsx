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

import { FC, useState } from "react";
import { Check, Circle, Eye, EyeOff } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { channels } from "./ChannelPanel";
import CardItem from "./CardItem";
import { PanelHeader } from "./PanelHeader";

export interface ChannelConfigPanelProps {
  channelId: string;
}

/**
 * 频道配置详情组件
 */
export const ChannelConfigPanel: FC<ChannelConfigPanelProps> = ({
  channelId,
}) => {
  const channel = channels.find((c) => c.id === channelId);
  const isConfigured = channel?.status === "configured";

  // 配置版本选择
  const [configVersion, setConfigVersion] = useState<"official" | "community">(
    "official",
  );

  // 启用状态
  const [isEnabled, setIsEnabled] = useState(isConfigured);

  // 功能开关
  const [imageCardOutput, setImageCardOutput] = useState(false);
  const [independentContext, setIndependentContext] = useState(false);

  // 输入框状态
  const [appId, setAppId] = useState(isConfigured ? "cli_9f280bbc3bbddcb" : "");
  const [appSecret, setAppSecret] = useState(
    isConfigured ? "••••••••••••••••" : "",
  );
  const [showSecret, setShowSecret] = useState(false);

  // 验证状态
  const isAppIdValid = appId.length > 0;
  const isAppSecretValid = appSecret.length > 0;

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
              <Switch checked={isEnabled} onCheckedChange={setIsEnabled} />
            </div>
          }
        />

        {/* 配置版本选择 */}
        <div className="text-left">
          <label className="text-sm font-medium">配置版本</label>
          <div className="flex gap-4 mt-2">
            <button
              onClick={() => setConfigVersion("official")}
              className={cn(
                "flex items-center gap-2 px-4 py-3 rounded-lg transition-colors",
                configVersion === "official"
                  ? "bg-primary/10  text-primary"
                  : "bg-card hover:bg-accent border-border",
              )}
            >
              <Circle
                className={cn(
                  "size-4",
                  configVersion === "official"
                    ? "fill-primary text-primary"
                    : "fill-transparent",
                )}
              />
              <span className="text-sm">{channel?.name}官方版</span>
            </button>
            <button
              onClick={() => setConfigVersion("community")}
              className={cn(
                "flex items-center gap-2 px-4 py-3 rounded-lg transition-colors",
                configVersion === "community"
                  ? "bg-primary/10  text-primary"
                  : "bg-card hover:bg-accent border-border",
              )}
            >
              <Circle
                className={cn(
                  "size-4",
                  configVersion === "community"
                    ? "fill-primary text-primary"
                    : "fill-transparent",
                )}
              />
              <span className="text-sm">ClawTeam 社区版</span>
            </button>
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
              label="图片卡片输出"
              description="将图片消息以卡片形式展示，提升阅读体验"
              action={
                <Switch
                  checked={imageCardOutput}
                  onCheckedChange={setImageCardOutput}
                />
              }
            />

            {/* 逐题独立上下文 */}
            <CardItem
              label="逐题独立上下文"
              description="每个问题使用独立的对话上下文，避免历史消息干扰"
              action={
                <Switch
                  checked={independentContext}
                  onCheckedChange={setIndependentContext}
                />
              }
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default ChannelConfigPanel;
