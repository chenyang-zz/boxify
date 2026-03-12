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

/**
 * 频道定义：用于渲染频道管理列表。
 */
export interface ChannelDefinition {
  id: string;
  name: string;
  description: string;
  type: "built-in" | "plugin";
  aliases?: string[];
}

/**
 * 参考 ClawPanel 维护的频道定义。
 */
export const channelDefinitions: ChannelDefinition[] = [
  {
    id: "whatsapp",
    name: "WhatsApp",
    description: "Baileys QR 扫码配对",
    type: "built-in",
  },
  {
    id: "telegram",
    name: "Telegram",
    description: "Bot API via grammY",
    type: "built-in",
  },
  {
    id: "discord",
    name: "Discord",
    description: "Discord Bot API + Gateway",
    type: "built-in",
  },
  {
    id: "irc",
    name: "IRC",
    description: "经典 IRC 服务器",
    type: "built-in",
  },
  {
    id: "slack",
    name: "Slack",
    description: "Bolt SDK 工作区应用",
    type: "built-in",
  },
  {
    id: "signal",
    name: "Signal",
    description: "signal-cli REST API",
    type: "built-in",
  },
  {
    id: "googlechat",
    name: "Google Chat",
    description: "Google Chat API Webhook",
    type: "built-in",
  },
  {
    id: "bluebubbles",
    name: "BlueBubbles",
    description: "macOS iMessage",
    type: "built-in",
  },
  {
    id: "webchat",
    name: "WebChat",
    description: "Gateway WebChat UI",
    type: "built-in",
  },
  {
    id: "feishu",
    name: "飞书 / Lark",
    description: "飞书机器人 WebSocket (插件)",
    type: "plugin",
    aliases: ["lark", "feishu-openclaw-plugin"],
  },
  {
    id: "boxify",
    name: "Boxify",
    description: "本地原生 channel inbox (插件)",
    type: "plugin",
  },
  {
    id: "qqbot",
    name: "QQ 官方机器人",
    description: "QQ 开放平台官方 Bot API (插件)",
    type: "plugin",
  },
  {
    id: "dingtalk",
    name: "钉钉",
    description: "钉钉机器人 (插件)",
    type: "plugin",
  },
  {
    id: "wecom",
    name: "企业微信",
    description: "企业微信应用消息 (插件)",
    type: "plugin",
  },
  {
    id: "msteams",
    name: "Microsoft Teams",
    description: "Bot Framework (插件)",
    type: "plugin",
  },
  {
    id: "mattermost",
    name: "Mattermost",
    description: "Bot API + WebSocket (插件)",
    type: "plugin",
  },
  {
    id: "line",
    name: "LINE",
    description: "LINE Messaging API (插件)",
    type: "plugin",
  },
  {
    id: "matrix",
    name: "Matrix",
    description: "Matrix 协议 (插件)",
    type: "plugin",
  },
  {
    id: "twitch",
    name: "Twitch",
    description: "Twitch Chat via IRC (插件)",
    type: "plugin",
  },
];
