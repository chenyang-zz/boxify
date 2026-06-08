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
import { Settings } from "lucide-react";
import { cn } from "@/lib/utils";
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
import { CommonSetting } from "./CommonSetting";
import { LLMSetting } from "./LLMSetting";
import { A2ASetting } from "./A2ASetting";
import { MCPSetting } from "./MCPSetting";

/**
 * 设置选项卡类型
 */
type SettingTab =
  | "common-setting"
  | "llm-setting"
  | "a2a-setting"
  | "mcp-setting";

/**
 * 设置菜单配置
 */
const SETTING_MENUS: Array<{
  key: SettingTab;
  title: string;
}> = [
  { key: "common-setting", title: "通用配置" },
  { key: "llm-setting", title: "模型提供商" },
  { key: "a2a-setting", title: "A2A Agent" },
  { key: "mcp-setting", title: "MCP 服务器" },
];

/**
 * Computer 设置弹窗组件
 */
export const ComputerSettings: FC = () => {
  const [open, setOpen] = useState(false);
  const [activeSetting, setActiveSetting] =
    useState<SettingTab>("common-setting");

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {/* 触发按钮 */}
      <DialogTrigger asChild>
        <Button variant="outline" size="icon-sm" className="cursor-pointer">
          <Settings className="size-4" />
        </Button>
      </DialogTrigger>

      {/* 弹窗内容 */}
      <DialogContent className="max-w-212.5! p-0 gap-0 overflow-hidden rounded-2xl border">
        {/* 头部 */}
        <DialogHeader className="px-6 py-5 text-left! border-b">
          <DialogTitle className="text-xl">设置</DialogTitle>
          <DialogDescription>管理您的 Computer 配置。</DialogDescription>
        </DialogHeader>

        {/* 中间主体 — 固定高度，右侧内部滚动 */}
        <div className="flex flex-row h-125">
          {/* 左侧导航菜单 */}
          <div className="w-50 p-3 shrink-0 overflow-y-auto border-r bg-sidebar">
            <div className="flex flex-col gap-1">
              {SETTING_MENUS.map((menu) => (
                <Button
                  key={menu.key}
                  variant={activeSetting === menu.key ? "secondary" : "ghost"}
                  size="sm"
                  className={cn(
                    "w-full justify-start cursor-pointer text-sm font-medium transition-colors",
                    activeSetting === menu.key
                      ? "bg-primary text-primary-foreground hover:bg-primary/90"
                      : "text-muted-foreground hover:text-foreground hover:bg-accent hover:text-accent-foreground",
                  )}
                  onClick={() => setActiveSetting(menu.key)}
                >
                  {menu.title}
                </Button>
              ))}
            </div>
          </div>

          {/* 右侧内容区域 */}
          <div className="flex-1 p-6 overflow-y-auto">
            {activeSetting === "common-setting" && <CommonSetting />}
            {activeSetting === "llm-setting" && <LLMSetting />}
            {activeSetting === "a2a-setting" && <A2ASetting />}
            {activeSetting === "mcp-setting" && <MCPSetting />}
          </div>
        </div>

        {/* 底部 */}
        <DialogFooter className="px-6 py-4 border-t">
          <Button
            variant="outline"
            className="cursor-pointer text-foreground"
            onClick={() => setOpen(false)}
          >
            取消
          </Button>
          <Button className="cursor-pointer px-6">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default ComputerSettings;
