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

import { FC } from "react";
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from "../ui/context-menu";
import { Pin, PinOff, X, XCircle, RotateCcw, Edit2 } from "lucide-react";
import { TabContextMenuProps } from "./types";
import { tabStoreMethods } from "@/store/tabs.store";

const TabContextMenu: FC<TabContextMenuProps> = ({ tab, children }) => {
  const { id } = tab;

  // 关闭当前标签的逻辑
  const onClose = () => {
    tabStoreMethods.closeTab(id);
  };

  // 关闭其他标签的逻辑
  const onCloseOthers = () => {
    tabStoreMethods.closeOtherTabs(id);
  };

  // 关闭所有标签的逻辑
  const onCloseAll = () => {
    tabStoreMethods.closeAllTabs();
  };

  const onCloseToRight = () => {
    tabStoreMethods.closeTabsToRight(id);
  };

  // 固定标签的逻辑
  const onPin = () => {
    tabStoreMethods.pinTab(id);
  };

  // 取消固定标签的逻辑
  const onUnpin = () => {
    tabStoreMethods.unpinTab(id);
  };

  // 重命名标签的逻辑
  const onRename = () => {
    // TODO: 实现重命名功能
  };
  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>{children}</ContextMenuTrigger>
      <ContextMenuContent>
        <ContextMenuItem onClick={onClose}>
          <X className="size-4" />
          关闭标签
        </ContextMenuItem>
        <ContextMenuItem onClick={onCloseOthers}>
          <XCircle className="size-4 " />
          关闭其他标签
        </ContextMenuItem>
        <ContextMenuItem onClick={onCloseToRight}>
          <XCircle className="size-4 " />
          关闭右侧标签
        </ContextMenuItem>
        <ContextMenuSeparator />
        <ContextMenuItem onClick={onCloseAll}>
          <RotateCcw className="size-4 " />
          关闭所有标签
        </ContextMenuItem>
        <ContextMenuSeparator />
        {tab.isPinned ? (
          <ContextMenuItem onClick={onUnpin}>
            <PinOff className="size-4 " />
            取消固定
          </ContextMenuItem>
        ) : (
          <ContextMenuItem onClick={onPin}>
            <Pin className="size-4 " />
            固定标签
          </ContextMenuItem>
        )}
        <ContextMenuItem onClick={onRename}>
          <Edit2 className="size-4 " />
          重命名
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
};

export default TabContextMenu;
