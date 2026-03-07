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
import { Button } from "@/components/ui/button";
import { BotIcon, FilesIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import { useActiveView, useSidebarStore } from "../store";
import type { SidebarView } from "../types";

/**
 * 导航标签组件
 * 用于在 files 和 control 视图之间切换
 */
export const NavigationTabs: FC = () => {
  const activeView = useActiveView();
  const setActiveView = useSidebarStore((state) => state.setActiveView);

  return (
    <div className="flex items-center px-2 py-1 gap-0.5">
      <NavigationTabButton
        view="files"
        activeView={activeView}
        onClick={() => setActiveView("files")}
        icon={<FilesIcon className="size-4" />}
        label="文件"
      />

      <NavigationTabButton
        view="control"
        activeView={activeView}
        onClick={() => setActiveView("control")}
        icon={<BotIcon className="size-4" />}
        label="控制"
      />
    </div>
  );
};

/**
 * 导航标签按钮
 */
interface NavigationTabButtonProps {
  view: SidebarView;
  activeView: SidebarView;
  onClick: () => void;
  icon: React.ReactNode;
  label: string;
}

const NavigationTabButton: FC<NavigationTabButtonProps> = ({
  view,
  activeView,
  onClick,
  icon,
}) => {
  return (
    <Button
      size="icon-xs"
      variant="ghost"
      className={cn("text-foreground", activeView === view && "bg-accent")}
      onClick={onClick}
      aria-label={view}
    >
      {icon}
    </Button>
  );
};

export default NavigationTabs;
