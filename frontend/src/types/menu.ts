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

import { MenuItemDefinition as _MenuItemDefinition } from "@wails/service";

export interface MenuItemDefinition extends Omit<
  NullToOptional<_MenuItemDefinition>,
  "items" | "id" | "contextData"
> {
  id?: string;
  onClick?: (payload: MenuClickPayload) => void | Promise<void>;
  items?: MenuItemDefinition[];
  contextData?: Record<string, any>;
}

/**
 * 菜单更新请求
 */
export interface MenuUpdateRequest {
  /** 菜单 ID */
  menuId: string;
  /** 新的菜单项列表 */
  items: MenuItemDefinition[];
  /** 新的上下文数据 */
  contextData?: Record<string, any>;
}

/**
 * 菜单项更新
 */
export interface MenuItemUpdate {
  /** 菜单项 ID */
  itemId: string;
  /** 新标签 */
  label?: string;
  /** 启用状态 */
  enabled?: boolean;
  /** 选中状态 */
  checked?: boolean;
  /** 新的上下文数据 */
  contextData?: Record<string, any>;
}

/**
 * 菜单信息
 */
export interface MenuInfo {
  /** 菜单 ID */
  menuId: string;
  /** 菜单标签 */
  label: string;
  /** 关联窗口 */
  window: string;
  /** 创建时间戳 */
  createdAt: number;
}

/**
 * 菜单类型
 */
export enum MenuType {
  /** 普通菜单项 */
  Item = "item",
  /** 复选框 */
  Checkbox = "checkbox",
  /** 单选项 */
  Radio = "radio",
  /** 分隔符 */
  Separator = "separator",
  /** 子菜单 */
  Submenu = "submenu",
}

/**
 * 预定义菜单模板
 */
export enum MenuTemplate {
  /** 树项菜单 */
  TreeItem = "tree-item",
  /** 表格菜单 */
  Table = "table",
  /** 标签页菜单 */
  Tab = "tab",
  /** 文件树菜单 */
  FileTree = "file-tree",
}

/**
 * 菜单点击载荷（简化版）
 * 只包含前端常用的字段
 */
export interface MenuClickPayload {
  /** 菜单项 ID */
  itemId: string;
  /** 菜单项标签 */
  label: string;
  /** 当前选中状态 */
  checked: boolean;
  /** 菜单级别的上下文数据 */
  contextData?: Record<string, any>;
  /** 菜单项级别的上下文数据 */
  itemData?: Record<string, any>;
}

/**
 * 菜单配置（新 API）
 * 不包含 menuId 和 label，由 hook 自动生成
 */
export interface MenuConfig {
  /** 菜单项列表 */
  items: MenuItemDefinition[];
  /** 全局上下文数据 */
  contextData?: Record<string, any>;
  /** 关联窗口（可选，默认当前窗口） */
  window?: string;
}
