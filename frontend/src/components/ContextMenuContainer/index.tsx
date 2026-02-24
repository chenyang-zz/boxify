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

import { FC, ReactNode } from "react";

import { MenuConfig } from "@/types/menu";
import { useContextMenu } from "@/hooks/use-context-menu";

/**
 * 菜单容器组件的 Props
 */
export interface ContextMenuContainerProps {
  /** 菜单定义 */
  menu: MenuConfig;
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
  const { open } = useContextMenu(menu);
  return (
    <div
      className={className}
      style={style}
      onContextMenu={(e) => {
        e.preventDefault();
        e.stopPropagation();

        open({
          x: e.clientX,
          y: e.clientY,
        });
      }}
    >
      {children}
    </div>
  );
};

export default ContextMenuContainer;
