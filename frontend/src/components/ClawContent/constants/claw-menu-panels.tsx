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

import type { ReactNode } from "react";
import { ChannelPanel } from "../components/ChannelPanel";
import { ChatPanel } from "../components/ChatPanel";
import { ComingSoonPanel } from "../components/ComingSoonPanel";
import { OverviewPanel } from "../components/OverviewPanel";
import { SkillPanel } from "../components/SkillPanel";

/**
 * Claw 菜单项 ID。
 */
export type ClawMenuItemId =
  | "overview"
  | "channel"
  | "skill"
  | "instance"
  | "session"
  | "usage"
  | "scheduled"
  | "chat";

/**
 * 菜单内容渲染器表。
 */
export const clawMenuPanelRenderers: Record<ClawMenuItemId, () => ReactNode> = {
  overview: () => <OverviewPanel />,
  channel: () => <ChannelPanel />,
  skill: () => <SkillPanel />,
  instance: () => <ComingSoonPanel text="实例面板（开发中）" />,
  session: () => <ComingSoonPanel text="会话面板（开发中）" />,
  usage: () => <ComingSoonPanel text="使用情况面板（开发中）" />,
  scheduled: () => <ComingSoonPanel text="定时任务面板（开发中）" />,
  chat: () => <ChatPanel />,
};

/**
 * 判断给定字符串是否为 Claw 菜单项。
 */
export function isClawMenuItemId(value: string): value is ClawMenuItemId {
  return value in clawMenuPanelRenderers;
}
