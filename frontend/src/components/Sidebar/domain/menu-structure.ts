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

import {
  MessageSquare,
  BarChart3,
  Link2,
  CircleDot,
  FileText,
  Activity,
  Star,
} from "lucide-react";
import type { MenuCategory, MenuItem } from "../types";

/**
 * 菜单分类数据
 */
export const menuCategories: MenuCategory[] = [
  {
    id: "chat",
    label: "聊天",
    items: [{ id: "chat", label: "聊天", icon: MessageSquare }],
  },
  {
    id: "control",
    label: "控制",
    items: [
      { id: "overview", label: "概览", icon: BarChart3 },
      { id: "channel", label: "频道", icon: Link2 },
      { id: "instance", label: "实例", icon: CircleDot },
      { id: "session", label: "会话", icon: FileText },
      { id: "usage", label: "使用情况", icon: Activity },
      { id: "scheduled", label: "定时任务", icon: Star },
    ],
  },
];

/**
 * 默认展开的分类 ID
 */
export const DEFAULT_EXPANDED_CATEGORIES = new Set(["chat", "control"]);

/**
 * 默认选中的菜单项 ID
 */
export const DEFAULT_SELECTED_ITEM = "overview";

/**
 * 根据 ID 查找菜单项
 */
export function findMenuItemById(itemId: string): MenuItem | undefined {
  for (const category of menuCategories) {
    const item = category.items.find((item) => item.id === itemId);
    if (item) return item;
  }
  return undefined;
}

/**
 * 根据 ID 查找菜单分类
 */
export function findMenuCategoryById(categoryId: string): MenuCategory | undefined {
  return menuCategories.find((category) => category.id === categoryId);
}

/**
 * 获取所有菜单项 ID
 */
export function getAllMenuItemIds(): string[] {
  return menuCategories.flatMap((category) => category.items.map((item) => item.id));
}
