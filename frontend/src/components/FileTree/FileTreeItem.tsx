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
import { FIleTreeList } from ".";
import { cn } from "@/lib/utils";
import { ChevronRight, File, Pencil } from "lucide-react";
import {
  ConnectionEnum,
  DBFileType,
  FileSystemType,
  FileType,
  isConnectionType,
} from "@/common/constrains";
import {
  PropertyItemType,
  propertyStoreMethods,
  usePropertyStore,
} from "@/store/property.store";
import { tabStoreMethods } from "@/store/tabs.store";
import { Spinner } from "../ui/spinner";
import { Badge } from "../ui/badge";
import {
  getPropertyItemByUUID,
  triggerDirOpen,
  triggerFileOpen,
} from "@/lib/property";
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from "@/components/ui/context-menu";
import { X, Edit, Trash2, RefreshCw } from "lucide-react";
import { closeConnectionByUUID } from "@/lib/connection";

interface FileTreeItemProps {
  item: PropertyItemType;
}

interface FileIconProps {
  isDir: boolean;
  type?: FileType;
}
const FileIcon: FC<FileIconProps> = ({ isDir, type }) => {
  if (isDir) {
    switch (type) {
      case FileSystemType.FOLDER:
        return (
          <img className="size-4.5 mr-1 shrink-0" src="/icons/folder.svg" />
        );
      case ConnectionEnum.MYSQL:
        return (
          <img className="size-4.5 mr-1 shrink-0" src="/icons/mysql.svg" />
        );
      case ConnectionEnum.REDIS:
        return (
          <img className="size-4.5 mr-1 shrink-0" src="/icons/redis.svg" />
        );
      case ConnectionEnum.MONGODB:
        return (
          <img className="size-4.5 mr-1 shrink-0" src="/icons/mongodb.svg" />
        );
      case DBFileType.DATABASE:
        return (
          <img className="size-4.5 mr-1 shrink-0" src="/icons/database.svg" />
        );
      case DBFileType.TABLE_FOLDER:
        return (
          <img className="size-4.5 mr-1 shrink-0" src="/icons/table.svg" />
        );
      case DBFileType.VIEW_FOLDER:
        return <img className="size-4.5 mr-1 shrink-0" src="/icons/view.svg" />;
      case DBFileType.QUERY_FOLDER:
        return (
          <img className="size-4.5 mr-1 shrink-0" src="/icons/query.svg" />
        );
      case DBFileType.FUNCTION_FOLDER:
        return (
          <img className="size-4.5 mr-1 shrink-0" src="/icons/function.svg" />
        );
    }
  }
  switch (type) {
    case ConnectionEnum.SSH:
      return (
        <img className="size-4.5 mr-1 shrink-0" src="/icons/terminal.svg" />
      );
    case DBFileType.TABLE:
      return <img className="size-4.5 mr-1 shrink-0" src="/icons/table.svg" />;
    default:
      return <File className="size-4.5 mr-1 shrink-0" />;
  }
};

const FileTreeItem: FC<FileTreeItemProps> = ({ item }) => {
  const [loading, setLoading] = useState(false);
  const selectedUUID = usePropertyStore((state) => state.selectedUUID);

  const isConnection = isConnectionType(item.type);

  // 菜单项处理函数
  const handleClose = async () => {
    closeConnectionByUUID(item.uuid);
    console.log("关闭连接:", item.uuid);
  };

  const handleEdit = async () => {
    // TODO: 实现编辑连接逻辑
    console.log("编辑连接:", item.uuid);
  };

  const handleDelete = async () => {
    // TODO: 实现删除连接逻辑
    console.log("删除连接:", item.uuid);
  };

  const handleRename = async () => {
    // TODO: 实现重命名连接逻辑
    console.log("重命名连接:", item.uuid);
  };

  const handleRefresh = async () => {
    // 刷新节点
    await triggerDirOpen(item.uuid);
  };

  const handleClickItem = () => {
    propertyStoreMethods.setSelectedUUID(item.uuid);
  };

  // 打开/关闭 列表项
  const handleTogglePropertyItem = async () => {
    try {
      if (!item.loaded) {
        setLoading(true);
      }
      if (!item.isDir) {
        await triggerFileOpen(item.uuid);
      } else {
        await triggerDirOpen(item.uuid);
      }
    } finally {
      setLoading(false);
    }
  };

  // 如果正在加载，显示加载图标；否则根据文件类型显示对应的图标
  const fileIcon = loading ? (
    <Spinner className="size-3 mr-1" />
  ) : (
    <FileIcon isDir={item.isDir} type={item.type} />
  );

  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>
        <div
          className={cn(
            "px-2 py-0.5 flex items-center justify-between overflow-hidden hover:bg-accent hover:text-accent-foreground cursor-default select-none",
            selectedUUID === item.uuid &&
              "bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground",
          )}
          style={{
            paddingLeft: item.level * 8, // 根据层级增加左侧缩进
          }}
          onClick={handleClickItem}
          onDoubleClick={handleTogglePropertyItem}
        >
          <div className="flex items-center">
            <ChevronRight
              className={cn(
                "size-3 mr-0.5 shrink-0",
                !item.isDir && "opacity-0",
              )}
              style={{
                transition: "transform 0.2s",
                transform: item.opened ? "rotate(90deg)" : "rotate(0deg)",
              }}
              onClick={handleTogglePropertyItem}
            />
            {fileIcon}
            <span className=" truncate">{item.label}</span>
          </div>

          {typeof item.extra?.["count"] === "number" && (
            <Badge className="py-0" variant="secondary">
              {item.extra["count"]}
            </Badge>
          )}
        </div>
      </ContextMenuTrigger>
      <ContextMenuContent>
        {isConnection && (
          <ContextMenuItem onClick={handleClose} disabled={!item?.loaded}>
            <X className="size-4" />
            关闭
          </ContextMenuItem>
        )}
        {isConnection && <ContextMenuSeparator />}
        <ContextMenuItem onClick={handleEdit}>
          <Edit className="size-4" />
          编辑
        </ContextMenuItem>
        <ContextMenuItem onClick={handleDelete}>
          <Trash2 className="size-4" />
          删除
        </ContextMenuItem>
        <ContextMenuItem onClick={handleRename}>
          <Pencil className="size-4" />
          重命名
        </ContextMenuItem>
        <ContextMenuSeparator />
        <ContextMenuItem onClick={handleRefresh}>
          <RefreshCw className="size-4" />
          刷新
        </ContextMenuItem>
      </ContextMenuContent>
      {item.opened && <FIleTreeList data={item.children || []} />}
    </ContextMenu>
  );
};

export default FileTreeItem;
