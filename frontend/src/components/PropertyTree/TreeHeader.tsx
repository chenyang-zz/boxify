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

import { CopyMinus, CopyPlus, PlusIcon } from "lucide-react";
import { DOMAttributes, FC, ReactNode, useMemo, useState } from "react";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import { useContextMenu } from "@/hooks/use-context-menu";
import { MenuConfig } from "@/types/menu";
import { MenuItemType } from "@wails/service";
import { useOpenWindowWithData } from "@/hooks/useOpenWindowWithData";
import { ConnectionEditInitialData } from "@/types/initial-data";
import { ConnectionEnum } from "@/common/constrains";

interface TreeHeaderAction {
  icon: ReactNode;
  label: string;
  onClick: DOMAttributes<HTMLElement>["onClick"];
}

const TreeHeader: FC = () => {
  const [isCollapsed, setIsCollapsed] = useState(true);

  const { openWindowWithData } = useOpenWindowWithData();

  const addPropertyMenu: MenuConfig = useMemo(
    () => ({
      items: [
        {
          label: "目录",
          type: MenuItemType.MenuItemTypeItem,
          onClick: async (payload) => {
            console.log(payload);
          },
        },
        {
          label: "本地终端",
          type: MenuItemType.MenuItemTypeItem,
          onClick: () => {
            // 打开连接弹窗
            openWindowWithData("connection-edit", {
              type: ConnectionEnum.TERMINAL,
              title: "终端配置编辑",
            } as ConnectionEditInitialData);
          },
        },
        {
          label: "Docker",
          type: MenuItemType.MenuItemTypeItem,
          onClick: async (payload) => {
            console.log(payload);
          },
        },
        {
          label: "远程连接",
          type: MenuItemType.MenuItemTypeSubmenu,
          items: [
            {
              label: "SSH",
              type: MenuItemType.MenuItemTypeItem,
              onClick: () => {
                // 打开连接弹窗
                openWindowWithData("connection-edit", {
                  type: ConnectionEnum.SSH,
                  title: "SSH配置编辑",
                } as ConnectionEditInitialData);
              },
            },
            {
              label: "SSH隧道",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
            {
              label: "RDP",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
            {
              label: "Telnet",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
            {
              label: "串口",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
          ],
        },
        {
          label: "数据库",
          type: MenuItemType.MenuItemTypeSubmenu,
          items: [
            {
              label: "Redis",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
            {
              label: "MySql",
              type: MenuItemType.MenuItemTypeItem,
              onClick: () => {
                // 打开连接弹窗
                openWindowWithData("connection-edit", {
                  type: ConnectionEnum.MYSQL,
                  title: "MySQL配置编辑",
                } as ConnectionEditInitialData);
              },
            },
            {
              label: "MariaDB",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
            {
              label: "PostgreSQL",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
            {
              label: "SqlServer",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
            {
              label: "ClickHouse",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
            {
              label: "SQLite",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
            {
              label: "Oracle",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
            {
              label: "达梦",
              type: MenuItemType.MenuItemTypeItem,
              onClick: async (payload) => {
                console.log(payload);
              },
            },
          ],
        },
      ],
    }),
    [openWindowWithData],
  );

  const addMenu = useContextMenu(addPropertyMenu);

  const actions: TreeHeaderAction[] = [
    {
      icon: isCollapsed ? (
        <CopyMinus className="size-4" />
      ) : (
        <CopyPlus className="size-4" />
      ),
      label: isCollapsed ? "点击全部折叠" : "点击全部展开",
      onClick: (e) => {
        setIsCollapsed((prev) => !prev);
      },
    },
    {
      icon: <PlusIcon className="size-4" />,
      label: "创建资产",
      onClick: (e) => {
        addMenu.open({
          x: e.clientX,
          y: e.clientY,
        });
      },
    },
  ];

  return (
    <nav className="flex items-center p-2 justify-between text-foreground shrink-0">
      <span className="text-sm font-bold">资产列表</span>
      <div className="flex gap-2" onClick={(e) => {}}>
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
