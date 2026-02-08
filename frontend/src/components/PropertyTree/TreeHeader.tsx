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

import { CopyMinus, CopyPlus } from "lucide-react";
import { FC, ReactNode, useState } from "react";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";

interface TreeHeaderAction {
  icon: ReactNode;
  label: string;
  onClick: () => void;
}

const TreeHeader: FC = () => {
  const [isCollapsed, setIsCollapsed] = useState(true);

  const actions: TreeHeaderAction[] = [
    {
      icon: isCollapsed ? (
        <CopyMinus className="size-4" />
      ) : (
        <CopyPlus className="size-4" />
      ),
      label: isCollapsed ? "点击全部折叠" : "点击全部展开",
      onClick: () => {
        setIsCollapsed((prev) => !prev);
      },
    },
  ];

  return (
    <nav className="flex items-center p-2 justify-between text-foreground">
      <span className="text-sm font-bold">资产列表</span>
      <div className="flex gap-2">
        {actions.map((action, index) => (
          <Tooltip key={index}>
            <TooltipTrigger asChild>
              <button
                className=" cursor-pointer"
                onClick={action.onClick}
                title={action.label}
              >
                {action.icon}
              </button>
            </TooltipTrigger>
            <TooltipContent side="bottom">
              <p>{action.label}</p>
            </TooltipContent>
          </Tooltip>
        ))}
      </div>
    </nav>
  );
};

export default TreeHeader;
