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

import { FC, useRef, useEffect, ReactNode } from "react";

import { MenuService, MenuDefinition } from "@wails/service";

/**
 * 菜单容器组件的 Props
 */
export interface ContextMenuContainerProps {
  /** 菜单定义 */
  menu: MenuDefinition;
  /** 子元素 */
  children: ReactNode;
  /** 额外的 className */
  className?: string;
  /** 额外的样式 */
  style?: React.CSSProperties;
}

/**
 * 菜单容器组件
 *
 * 专门用于支持右键触发的菜单容器
 *
 * @example
 * ```tsx
 * <ContextMenuContainer menu={menuDefinition} className="p-4 border">
 *   <div>在这个区域内右键即可触发菜单</div>
 * </ContextMenuContainer>
 * ```
 */
const ContextMenuContainer: FC<ContextMenuContainerProps> = ({
  menu,
  children,
  className = "",
  style,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // 调用后端创建菜单
    MenuService.CreateContextMenu(menu)
      .then((result) => {
        if (result?.success) {
          console.log("[ContextMenuContainer] 菜单创建成功:", menu.menuId);
        } else {
          console.error(
            "[ContextMenuContainer] 菜单创建失败:",
            result?.message,
          );
        }
      })
      .catch((error) => {
        console.error("[ContextMenuContainer] 菜单创建异常:", error);
      });

    return () => {
      // 清理菜单
      MenuService.UnregisterContextMenu(menu.menuId)
        .then((result) => {
          if (result?.success) {
            console.log("[ContextMenuContainer] 菜单已注销:", menu.menuId);
          }
        })
        .catch((error) => {
          console.error("[ContextMenuContainer] 菜单注销异常:", error);
        });
    };
  }, [menu]);

  // 处理右键菜单
  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();

    const element = containerRef.current;
    if (!element) return;

    // 设置上下文数据
    const contextData = {
      ...menu.contextData,
      _x: e.clientX,
      _y: e.clientY,
      _trigger: "contextmenu",
    };

    // 更新 data-context-menu-data 属性
    element.dataset.contextMenuData = JSON.stringify(contextData);

    console.log("[ContextMenuContainer] 右键触发菜单:", {
      menuId: menu.menuId,
      x: e.clientX,
      y: e.clientY,
    });
  };

  return (
    <div
      ref={containerRef}
      className={className}
      style={style}
      data-context-menu={menu.menuId}
      onContextMenu={handleContextMenu}
    >
      {children}
    </div>
  );
};

export default ContextMenuContainer;
