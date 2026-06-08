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
import { Wrench, Loader2, Trash } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { SettingList } from "./SettingList";

export interface MCPServerItem {
  server_name: string;
  transport: string;
  enabled: boolean;
  tools: string[];
}

interface MCPSettingProps {
  servers?: MCPServerItem[];
  loading?: boolean;
  onToggleEnabled?: (serverName: string, enabled: boolean) => void;
  onDelete?: (serverName: string) => void;
  onAdd?: (config: string) => Promise<boolean>;
}

const MCP_CONFIG_PLACEHOLDER = `{
  "mcpServers": {
    "example": {
      "command": "uvx",
      "args": [
        "example-mcp-server"
      ],
      "env": {
        "ACCESS_KEY": "YOUR_ACCESS_KEY"
      }
    }
  }
}`;

/** 单条 MCP 服务器卡片 */
const MCPServerCard: FC<{
  server: MCPServerItem;
  onToggleEnabled?: (name: string, enabled: boolean) => void;
  onDelete?: (name: string) => void;
}> = ({ server, onToggleEnabled, onDelete }) => (
  <div className="bg-muted rounded-xl p-4 flex flex-col gap-2">
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-2">
        <span className="text-sm font-medium text-foreground">
          {server.server_name}
        </span>
        <Badge variant="secondary">{server.transport}</Badge>
        {!server.enabled && <Badge>禁用</Badge>}
      </div>
      <div className="flex items-center gap-2">
        <Button
          type="button"
          variant="ghost"
          size="icon-xs"
          className="cursor-pointer text-muted-foreground hover:text-foreground"
          onClick={() => onDelete?.(server.server_name)}
        >
          <Trash className="size-4" />
        </Button>
        <Switch
          checked={server.enabled}
          onCheckedChange={(checked) =>
            onToggleEnabled?.(server.server_name, checked)
          }
        />
      </div>
    </div>
    {server.tools.length > 0 && (
      <div className="flex flex-wrap items-center gap-x-2 gap-y-1 pt-1">
        <Wrench className="size-3 text-muted-foreground" />
        {server.tools.map((tool) => (
          <Badge key={tool} variant="outline" className="text-xs">
            {tool}
          </Badge>
        ))}
      </div>
    )}
  </div>
);

/** MCP 添加服务器弹窗 */
const MCPAddDialog: FC<{
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdd: (config: string) => Promise<void>;
}> = ({ open, onOpenChange, onAdd }) => {
  const [config, setConfig] = useState("");
  const [adding, setAdding] = useState(false);

  const handleSubmit = async () => {
    if (!config.trim()) return;
    setAdding(true);
    try {
      await onAdd(config.trim());
      setConfig("");
    } finally {
      setAdding(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>
        <Button type="button" size="xs" className="cursor-pointer">
          添加服务器
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader className="text-left!">
          <DialogTitle>添加新的 MCP 服务器</DialogTitle>
          <DialogDescription>
            使用标准的 JSON MCP 配置来创建新服务器。
            请将配置粘贴到下方，然后点击"添加"。
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div>
            <label
              htmlFor="mcp_config"
              className="text-sm font-medium text-foreground block mb-1.5 text-left"
            >
              MCP 配置
            </label>
            <Textarea
              id="mcp_config"
              placeholder={MCP_CONFIG_PLACEHOLDER}
              value={config}
              onChange={(e) => setConfig(e.target.value)}
              className="min-h-50 font-mono text-xs"
              disabled={adding}
            />
          </div>
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button
              variant="outline"
              className="cursor-pointer"
              disabled={adding}
            >
              取消
            </Button>
          </DialogClose>
          <Button
            className="cursor-pointer"
            onClick={handleSubmit}
            disabled={adding}
          >
            {adding && <Loader2 className="size-4 animate-spin" />}
            添加
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export const MCPSetting: FC<MCPSettingProps> = ({
  servers = [],
  loading = false,
  onToggleEnabled,
  onDelete,
  onAdd,
}) => {
  const [addDialogOpen, setAddDialogOpen] = useState(false);

  const handleAdd = async (config: string) => {
    if (!onAdd) {
      setAddDialogOpen(false);
      return;
    }
    const success = await onAdd(config);
    if (success) {
      setAddDialogOpen(false);
    }
  };

  return (
    <div className="w-full">
      {/* 主标题 */}
      <div className="flex items-start justify-between mb-1">
        <h2 className="text-xl font-semibold text-foreground text-left">
          MCP 服务器
        </h2>
        <MCPAddDialog
          open={addDialogOpen}
          onOpenChange={setAddDialogOpen}
          onAdd={handleAdd}
        />
      </div>
      <p className="text-sm text-muted-foreground mb-6 text-left">
        MCP 协议通过集成外部工具来增强应用能力，如私有搜索、网页浏览等。
      </p>

      {/* 列表（含加载态和空态） */}
      <SettingList
        loading={loading}
        empty={servers.length === 0}
        emptyText="暂无 MCP 服务器"
        emptyDescription="请点击上方按钮添加"
      >
        {servers.map((server) => (
          <MCPServerCard
            key={server.server_name}
            server={server}
            onToggleEnabled={onToggleEnabled}
            onDelete={onDelete}
          />
        ))}
      </SettingList>
    </div>
  );
};

export default MCPSetting;
