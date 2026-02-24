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

import { useCallback, useEffect, useRef } from "react";
import { Events } from "@wailsio/runtime";
import { MenuService, MenuDefinition, MenuClickEvent } from "@wails/service";
import { MenuConfig, MenuClickPayload, MenuItemDefinition } from "@/types/menu";
import { v4 as uuid } from "uuid";
import { currentWindowName } from "@/lib/utils";
import { MenuItemDefinition as GoMenuItemDefinition } from "@wails/service";

// ============================================================================
// 全局状态管理
// ============================================================================

/**
 * 全局回调注册表
 * 结构: menuId -> (action -> callback)
 */
const menuCallbackRegistry = new Map<
  string,
  Map<string, (payload: MenuClickPayload) => void>
>();

/**
 * 全局事件监听器解绑函数
 */
let globalListenerUnbind: (() => void) | null = null;

// ============================================================================
// 工具函数
// ============================================================================

/**
 * 生成唯一的菜单 ID
 */
const generateMenuId = (): string => {
  return `menu-${uuid()}`;
};

/**
 * 确保全局事件监听器已注册
 * 所有菜单共享一个监听器，提高效率
 */
const ensureGlobalListener = () => {
  if (globalListenerUnbind) {
    return; // 已经注册
  }

  globalListenerUnbind = Events.On(
    "menu:clicked",
    (event: { data: MenuClickEvent }) => {
      const { menuId, itemId, label, checked, contextData, itemData } =
        event.data;

      // 查找对应菜单的回调映射
      const callbackMap = menuCallbackRegistry.get(menuId);
      if (!callbackMap) {
        return; // 菜单不存在或已销毁
      }

      // 查找对应 itemId 的回调
      const callback = callbackMap.get(itemId);
      if (callback) {
        // 调用回调
        const payload: MenuClickPayload = {
          itemId,
          label,
          checked,
          contextData,
          itemData,
        };
        callback(payload);
      }
    },
  );
};

/**
 * 从 MenuItem 树中提取 onClick 回调，转换为 MenuItemDefinition
 * @param items - MenuItem 数组
 * @param callbackMap - 用于存储回调的 Map
 * @returns MenuItemDefinition 数组
 */
const extractCallbacks = (
  items: MenuItemDefinition[],
  callbackMap: Map<string, (payload: MenuClickPayload) => void>,
): MenuItemDefinition[] => {
  return items.map((item) => {
    const def: MenuItemDefinition = {
      id: "menu-item-" + uuid(), // 前缀避免冲突
      type: item.type,
      label: item.label || "",
      checked: item.checked ?? false,
      shortcut: item.shortcut || "",
      enabled: item.enabled ?? true,
      contextData: item.contextData ?? {},
      items: [], // 子菜单将在递归中处理
    };

    // 存储 onClick 回调
    if (item.onClick && def.id) {
      callbackMap.set(def.id, item.onClick);
    }

    // 递归处理子菜单
    if (item.items && item.items.length > 0) {
      def.items = extractCallbacks(item.items, callbackMap);
    }

    return def as MenuItemDefinition;
  });
};

// ============================================================================
// 类型定义
// ============================================================================

/**
 * 上下文菜单实例
 */
export interface ContextMenuInstance {
  /**
   * 打开菜单
   * @param pos - 位置信息（可选，默认使用上次右键位置或屏幕中心）
   */
  open: (pos?: { x: number; y: number }) => void;
  /**
   * 更新菜单
   * @param menuConfig - 新的菜单配置
   */
  update: (menuConfig: MenuConfig) => void;
}

// ============================================================================
// 新 API: useContextMenu with MenuConfig
// ============================================================================

/**
 * 使用上下文菜单 Hook（新 API）
 *
 * @param menuConfig - 菜单配置
 * @returns 上下文菜单实例
 *
 * @example
 * ```tsx
 * const MyComponent = () => {
 *   const menu = useContextMenu({
 *     items: [
 *       {
 *         type: "item",
 *         label: "刷新",
 *         action: "refresh",
 *         onClick: (payload) => {
 *           console.log("刷新点击", payload.action);
 *         }
 *       },
 *       {
 *         type: "checkbox",
 *         label: "显示隐藏",
 *         action: "toggle-hidden",
 *         checked: false,
 *         onClick: (payload) => {
 *           console.log("切换显示:", payload.checked);
 *         }
 *       },
 *     ],
 *     contextData: { source: "file-manager" }
 *   });
 *
 *   return <button onClick={() => menu.open()}>打开菜单</button>;
 * };
 * ```
 */
export function useContextMenu(menuConfig: MenuConfig): ContextMenuInstance {
  const menuIdRef = useRef<string>("");
  const callbackMapRef = useRef<
    Map<string, (payload: MenuClickPayload) => void>
  >(new Map());
  const lastPositionRef = useRef<{ x: number; y: number } | null>(null);

  // 组件挂载时创建菜单
  useEffect(() => {
    let isRegistered = true;
    // 生成唯一 menuId
    if (!menuIdRef.current) {
      menuIdRef.current = generateMenuId();
      isRegistered = false;
    }

    // 提取回调并转换为 MenuItemDefinition
    const callbackMap = new Map<string, (payload: MenuClickPayload) => void>();
    const items = extractCallbacks(menuConfig.items, callbackMap);
    callbackMapRef.current = callbackMap;

    // 注册到全局回调表
    menuCallbackRegistry.set(menuIdRef.current, callbackMap);

    // 确保全局监听器已注册
    ensureGlobalListener();

    if (!isRegistered) {
      // 构建后端 MenuDefinition
      const menuDefinition: MenuDefinition = {
        menuId: menuIdRef.current,
        label: `Auto Menu ${menuIdRef.current}`,
        window: menuConfig.window || currentWindowName(),
        items: items as GoMenuItemDefinition[],
        contextData: menuConfig.contextData ?? {},
      };

      // 调用后端创建菜单
      MenuService.CreateContextMenu(menuDefinition)
        .then((result) => {
          if (result?.success) {
            console.log("[useContextMenu] 菜单创建成功:", menuIdRef.current);
          } else {
            console.error("[useContextMenu] 菜单创建失败:", result?.message);
          }
        })
        .catch((error) => {
          console.error("[useContextMenu] 菜单创建异常:", error);
        });
    } else {
      // 已存在菜单，执行更新
      MenuService.UpdateMenu({
        menuId: menuIdRef.current,
        items: items as GoMenuItemDefinition[],
        contextData: menuConfig.contextData ?? {},
      })
        .then((result) => {
          if (result?.success) {
            console.log("[useContextMenu] 菜单更新成功:", menuIdRef.current);
          } else {
            console.error("[useContextMenu] 菜单更新失败:", result?.message);
          }
        })
        .catch((error) => {
          console.error("[useContextMenu] 菜单更新异常:", error);
        });
    }

    // 清理函数
    return () => {
      // 从全局回调表移除
      if (menuIdRef.current) {
        menuCallbackRegistry.delete(menuIdRef.current);
        console.log("[useContextMenu] 菜单回调已清理:", menuIdRef.current);
      }
    };
  }, [menuConfig]); // menuConfig 变化时重新创建

  useEffect(() => {
    return () => {
      // 组件卸载时清理全局监听器
      if (globalListenerUnbind) {
        globalListenerUnbind();
        globalListenerUnbind = null;
        console.log("[useContextMenu] 全局事件监听器已卸载");
      }

      // 注销后端菜单
      if (menuIdRef.current) {
        MenuService.UnregisterContextMenu(menuIdRef.current).catch((error) => {
          console.error("[useContextMenu] 菜单注销异常:", error);
        });
        menuIdRef.current = "";
      }
    };
  }, []);

  // 打开菜单
  const open = useCallback(
    (pos?: { x: number; y: number }) => {
      const targetX =
        pos?.x ?? lastPositionRef.current?.x ?? window.innerWidth / 2;
      const targetY =
        pos?.y ?? lastPositionRef.current?.y ?? window.innerHeight / 2;

      // 创建隐藏的触发元素
      const triggerElement = document.createElement("div");
      triggerElement.style.position = "absolute";
      triggerElement.style.left = "0";
      triggerElement.style.top = "0";
      triggerElement.style.width = "0";
      triggerElement.style.height = "0";
      triggerElement.dataset.contextMenu = menuIdRef.current;
      triggerElement.style.setProperty(
        "--custom-contextmenu",
        menuIdRef.current,
      );

      // 设置上下文数据
      const contextData = {
        ...(menuConfig.contextData || {}),
        _x: targetX,
        _y: targetY,
        _trigger: "manual",
      };
      triggerElement.dataset.contextMenuData = JSON.stringify(contextData);

      document.body.appendChild(triggerElement);

      // 创建并触发模拟的右键事件
      const mouseEvent = new MouseEvent("contextmenu", {
        bubbles: true,
        cancelable: true,
        clientX: targetX,
        clientY: targetY,
        button: 2,
        buttons: 2,
      });

      triggerElement.dispatchEvent(mouseEvent);

      // 记录位置
      lastPositionRef.current = { x: targetX, y: targetY };

      // 延迟清理触发元素
      setTimeout(() => {
        document.body.removeChild(triggerElement);
      }, 100);

      console.log("[useContextMenu] 手动触发菜单:", {
        menuId: menuIdRef.current,
        x: targetX,
        y: targetY,
      });
    },
    [menuConfig],
  );

  // 更新菜单
  const update = useCallback((newMenuConfig: MenuConfig) => {
    const menuId = menuIdRef.current;
    if (!menuId) {
      console.warn("[useContextMenu] 菜单未初始化，无法更新");
      return;
    }

    // 提取新回调
    const newCallbackMap = new Map<
      string,
      (payload: MenuClickPayload) => void
    >();
    const items = extractCallbacks(newMenuConfig.items, newCallbackMap);

    // 更新回调表
    menuCallbackRegistry.set(menuId, newCallbackMap);
    callbackMapRef.current = newCallbackMap;

    // 更新后端菜单
    MenuService.UpdateMenu({
      menuId,
      items: items as GoMenuItemDefinition[],
      contextData: newMenuConfig.contextData ?? {},
    })
      .then((result) => {
        if (result?.success) {
          console.log("[useContextMenu] 菜单更新成功:", menuId);
        } else {
          console.error("[useContextMenu] 菜单更新失败:", result?.message);
        }
      })
      .catch((error) => {
        console.error("[useContextMenu] 菜单更新异常:", error);
      });
  }, []);

  return { open, update };
}
