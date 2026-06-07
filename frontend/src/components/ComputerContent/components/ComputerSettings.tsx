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

import { FC, useState } from "react";
import { Settings, Languages, Wrench, LayoutGrid } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Separator } from "@/components/ui/separator";

/**
 * 设置选项卡类型
 */
type SettingTab = "common-setting" | "llm-setting" | "a2a-setting" | "mcp-setting";

/**
 * 设置菜单配置
 */
const SETTING_MENUS: Array<{
  key: SettingTab;
  icon: React.ComponentType<{ className?: string }>;
  title: string;
}> = [
  { key: "common-setting", icon: Settings, title: "通用配置" },
  { key: "llm-setting", icon: Languages, title: "模型提供商" },
  { key: "a2a-setting", icon: LayoutGrid, title: "A2A Agent 配置" },
  { key: "mcp-setting", icon: Wrench, title: "MCP 服务器" },
];

/**
 * Computer 设置弹窗组件
 */
export const ComputerSettings: FC = () => {
  const [open, setOpen] = useState(false);
  const [activeSetting, setActiveSetting] = useState<SettingTab>("common-setting");

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {/* 触发按钮 */}
      <DialogTrigger asChild>
        <Button variant="outline" size="icon-sm" className="cursor-pointer">
          <Settings className="size-4" />
        </Button>
      </DialogTrigger>

      {/* 弹窗内容 */}
      <DialogContent className="!max-w-[850px] p-0 gap-0 overflow-hidden">
        {/* 头部 */}
        <DialogHeader className="px-6 py-4 border-b border-border">
          <DialogTitle>设置</DialogTitle>
          <DialogDescription>管理您的 Computer 配置。</DialogDescription>
        </DialogHeader>

        {/* 中间主体 */}
        <div className="flex flex-row min-h-[400px]">
          {/* 左侧导航菜单 */}
          <div className="w-[180px] p-3 shrink-0">
            <div className="flex flex-col gap-1">
              {SETTING_MENUS.map((menu) => (
                <Button
                  key={menu.key}
                  variant={activeSetting === menu.key ? "default" : "ghost"}
                  className="w-full cursor-pointer justify-start gap-2"
                  onClick={() => setActiveSetting(menu.key)}
                >
                  <menu.icon className="size-4" />
                  {menu.title}
                </Button>
              ))}
            </div>
          </div>

          {/* 分隔符 */}
          <Separator orientation="vertical" />

          {/* 右侧内容区域 */}
          <div className="flex-1 p-6 overflow-y-auto">
            {/* TODO: 根据 activeSetting 渲染不同内容 */}
          </div>
        </div>

        {/* 底部 */}
        <DialogFooter className="px-6 py-4 border-t border-border">
          <Button variant="outline" className="cursor-pointer" onClick={() => setOpen(false)}>
            取消
          </Button>
          <Button className="cursor-pointer">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default ComputerSettings;
