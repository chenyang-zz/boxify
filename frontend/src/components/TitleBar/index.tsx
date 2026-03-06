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
import { System, Window } from "@wailsio/runtime";
import { useShallow } from "zustand/react/shallow";
import { tabStoreMethods, useTabsStore } from "@/store/tabs.store";
import TabBar from "../Tabs/TabBar";
import { Button } from "../ui/button";
import { LayoutGrid, PanelLeftIcon } from "lucide-react";
import { appStoreMethods, useAppStore } from "@/store/app.store";
import { cn } from "@/lib/utils";

const macControlButtonClass =
  "h-3 w-3 rounded-full transition-opacity hover:opacity-85";

const TitleBar: FC = () => {
  // 仅在 macOS 渲染窗口控制按钮，保持各平台原生习惯一致。
  const isMac = System.IsMac();
  const isOpen = useAppStore(useShallow((state) => state.isPropertyOpen));

  // 标题栏托管标签状态，保证 TabBar 放在窗口顶部时仍可完整操作标签。
  const { tabs, activeTabId } = useTabsStore(
    useShallow((state) => ({
      tabs: state.tabs,
      activeTabId: state.activeTabId,
    })),
  );

  // 关闭当前窗口。
  const handleWindowClose = () => {
    Window.Close().catch((err) => {
      console.error("关闭窗口失败:", err);
    });
  };

  // 最小化当前窗口。
  const handleWindowMinimise = () => {
    Window.Minimise().catch((err) => {
      console.error("最小化窗口失败:", err);
    });
  };

  // 切换窗口最大化状态。
  const handleWindowToggleMaximise = () => {
    Window.ToggleMaximise().catch((err) => {
      console.error("切换窗口最大化失败:", err);
    });
  };

  const handleTabSelect = (tabId: string) => {
    tabStoreMethods.setActiveTab(tabId);
  };

  return (
    <header
      className="w-full shrink-0 cursor-default pl-4 bg-card border-b flex items-center"
      style={{ "--wails-draggable": "drag" } as React.CSSProperties}
    >
      {isMac && (
        <div
          className="flex items-center gap-2 mr-3"
          style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
        >
          <button
            type="button"
            aria-label="关闭窗口"
            title="关闭"
            className={`${macControlButtonClass} bg-[#FF5F57]`}
            onClick={handleWindowClose}
          />
          <button
            type="button"
            aria-label="最小化窗口"
            title="最小化"
            className={`${macControlButtonClass} bg-[#FEBC2E]`}
            onClick={handleWindowMinimise}
          />
          <button
            type="button"
            aria-label="缩放窗口"
            title="缩放"
            className={`${macControlButtonClass} bg-[#28C840]`}
            onClick={handleWindowToggleMaximise}
          />
        </div>
      )}
      <Button
        size="icon-xs"
        variant="ghost"
        className={cn("text-foreground mr-1", isOpen && "bg-accent")}
        onClick={() => {
          appStoreMethods.setIsPropertyOpen(!isOpen);
        }}
      >
        <PanelLeftIcon className="size-4" />
      </Button>
      <Button size="icon-xs" variant="ghost" className="text-foreground mr-1">
        <LayoutGrid className="size-4" />
      </Button>
      <div className="ml-2 flex-1 overflow-hidden">
        <TabBar
          tabs={tabs}
          activeTabId={activeTabId}
          onTabSelect={handleTabSelect}
        />
      </div>
    </header>
  );
};

export default TitleBar;
