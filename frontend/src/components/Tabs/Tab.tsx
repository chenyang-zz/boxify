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
import { X, Pin, PinOff } from "lucide-react";
import { cn } from "@/lib/utils";
import { TabProps } from "./types";
import { Button } from "../ui/button";
import TabContextMenu from "./TabContextMenu";

const Tab: FC<TabProps> = ({
  tab,
  isActive,
  onClose,
  onPin,
  onUnpin,
  onSelect,
}) => {
  const handleClose = (e: React.MouseEvent) => {
    e.stopPropagation();
    onClose(tab.id);
  };

  const handlePin = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (tab.isPinned) {
      onUnpin(tab.id);
    } else {
      onPin(tab.id);
    }
  };

  return (
    <TabContextMenu tab={tab}>
      <div
        className={cn(
          "group flex items-center pl-3 pr-2 py-1.5 text-sm",
          "hover:bg-accent hover:text-accent-foreground",
          "transition-colors duration-150",
          "min-w-30 max-w-50 select-none",
          isActive && "bg-secondary border-b-2 border-b-primary",
        )}
        onClick={() => onSelect(tab.id)}
      >
        {/* 固定图标 */}
        {tab.isPinned && (
          <Pin className="size-3 text-muted-foreground shrink-0 mr-2" />
        )}

        {/* 标签标题 */}
        <span className="flex-1 truncate text-left mr-2">{tab.label}</span>

        {/* 固定按钮（悬停时显示，固定标签常显） */}
        <Button
          size="icon-sm"
          variant="ghost"
          className={cn(
            "hidden group-hover:block size-4 p-0 shrink-0",
            tab.isPinned && "opacity-100",
          )}
          onClick={handlePin}
        >
          {tab.isPinned ? (
            <PinOff className="size-3 " />
          ) : (
            <Pin className="size-3" />
          )}
        </Button>

        {/* 关闭按钮（悬停时显示，固定标签隐藏） */}
        {!tab.isPinned && (
          <Button
            size="icon-sm"
            variant="ghost"
            className="hidden group-hover:block size-4 p-0 shrink-0"
            onClick={handleClose}
          >
            <X className="size-3" />
          </Button>
        )}
      </div>
    </TabContextMenu>
  );
};

export default Tab;
