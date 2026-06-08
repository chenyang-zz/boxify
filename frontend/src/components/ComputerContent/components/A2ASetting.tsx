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
import { LayoutList, Loader2, Trash } from "lucide-react";
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
import { Input } from "@/components/ui/input";
import { SettingList } from "./SettingList";

export interface A2AServerItem {
  id: string;
  name: string;
  description?: string;
  enabled: boolean;
  input_modes?: string[];
  output_modes?: string[];
  streaming?: boolean;
  push_notifications?: boolean;
}

interface A2ASettingProps {
  servers?: A2AServerItem[];
  loading?: boolean;
  onToggleEnabled?: (id: string, enabled: boolean) => void;
  onDelete?: (id: string) => void;
  onAdd?: (baseUrl: string) => Promise<boolean>;
}

/** A2A 能力标签映射 */
const A2A_CAPABILITY_LABELS: Record<string, string> = {
  streaming: "流式输出",
  push_notifications: "推送通知",
};

/** 单条 A2A 服务器卡片 */
const A2AServerCard: FC<{
  server: A2AServerItem;
  onToggleEnabled?: (id: string, enabled: boolean) => void;
  onDelete?: (id: string) => void;
}> = ({ server, onToggleEnabled, onDelete }) => (
  <div className="bg-muted rounded-xl p-4 flex flex-col gap-2">
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-2">
        <span className="text-sm font-medium text-foreground">
          {server.name}
        </span>
        {!server.enabled && <Badge>禁用</Badge>}
      </div>
      <div className="flex items-center gap-2">
        <Button
          type="button"
          variant="ghost"
          size="icon-xs"
          className="cursor-pointer text-muted-foreground hover:text-foreground"
          onClick={() => onDelete?.(server.id)}
        >
          <Trash className="size-4" />
        </Button>
        <Switch
          checked={server.enabled}
          onCheckedChange={(checked) => onToggleEnabled?.(server.id, checked)}
        />
      </div>
    </div>
    {server.description && (
      <p className="text-sm text-muted-foreground">{server.description}</p>
    )}
    <div className="flex flex-wrap items-center gap-x-2 gap-y-1 pt-1">
      <LayoutList className="size-3 text-muted-foreground" />
      {server.input_modes?.map((mode) => (
        <Badge key={`in-${mode}`} variant="secondary">
          输入: {mode}
        </Badge>
      ))}
      {server.output_modes?.map((mode) => (
        <Badge key={`out-${mode}`} variant="secondary">
          输出: {mode}
        </Badge>
      ))}
      <Badge variant={server.streaming ? "secondary" : "outline"}>
        流式输出: {server.streaming ? "开启" : "关闭"}
      </Badge>
      <Badge variant={server.push_notifications ? "secondary" : "outline"}>
        推送通知: {server.push_notifications ? "开启" : "关闭"}
      </Badge>
    </div>
  </div>
);

/** A2A 添加服务器弹窗 */
const A2AAddDialog: FC<{
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdd: (url: string) => Promise<void>;
}> = ({ open, onOpenChange, onAdd }) => {
  const [url, setUrl] = useState("");
  const [adding, setAdding] = useState(false);

  const handleSubmit = async () => {
    if (!url.trim()) return;
    setAdding(true);
    try {
      await onAdd(url.trim());
      setUrl("");
    } finally {
      setAdding(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>
        <Button type="button" size="xs" className="cursor-pointer">
          添加远程 Agent
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader className="!text-left">
          <DialogTitle>添加远程 Agent</DialogTitle>
          <DialogDescription>
            使用标准的 A2A 协议来连接远程 Agent。
            <br />
            请将配置粘贴到下方，然后点击"添加"。
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div>
            <label
              htmlFor="a2a_base_url"
              className="text-sm font-medium text-foreground block mb-1.5 text-left"
            >
              Agent 地址
            </label>
            <Input
              id="a2a_base_url"
              type="url"
              placeholder="Example: https://example.com/weather-agent"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
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

export const A2ASetting: FC<A2ASettingProps> = ({
  servers = [],
  loading = false,
  onToggleEnabled,
  onDelete,
  onAdd,
}) => {
  const [addDialogOpen, setAddDialogOpen] = useState(false);

  const handleAdd = async (url: string) => {
    if (!onAdd) {
      setAddDialogOpen(false);
      return;
    }
    const success = await onAdd(url);
    if (success) {
      setAddDialogOpen(false);
    }
  };

  return (
    <div className="w-full">
      {/* 主标题 */}
      <div className="flex items-start justify-between mb-1">
        <h2 className="text-xl font-semibold text-foreground text-left">
          A2A Agent
        </h2>
        <A2AAddDialog
          open={addDialogOpen}
          onOpenChange={setAddDialogOpen}
          onAdd={handleAdd}
        />
      </div>
      <p className="text-sm text-muted-foreground mb-6 text-left">
        A2A 协议让你可以连接外部 Agent 来扩展能力，如天气查询、数据分析等。
      </p>

      {/* 列表（含加载态和空态） */}
      <SettingList
        loading={loading}
        empty={servers.length === 0}
        emptyText="暂无 A2A Agent"
        emptyDescription="请点击上方按钮添加"
      >
        {servers.map((server) => (
          <A2AServerCard
            key={server.id}
            server={server}
            onToggleEnabled={onToggleEnabled}
            onDelete={onDelete}
          />
        ))}
      </SettingList>
    </div>
  );
};

export default A2ASetting;
