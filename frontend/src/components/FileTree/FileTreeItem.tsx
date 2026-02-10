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
import { ChevronRight, File } from "lucide-react";
import {
  ConnectionType,
  DBFileType,
  FileSystemType,
  FileType,
} from "@/common/constrains";
import {
  getPropertyItemByUUID,
  PropertyItemType,
  propertyStoreMethods,
  usePropertyStore,
} from "@/store/property.store";
import { Spinner } from "../ui/spinner";

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
      case ConnectionType.MYSQL:
        return (
          <img className="size-4.5 mr-1 shrink-0" src="/icons/mysql.svg" />
        );
      case ConnectionType.REDIS:
        return (
          <img className="size-4.5 mr-1 shrink-0" src="/icons/redis.svg" />
        );
      case ConnectionType.MONGODB:
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
    case ConnectionType.SSH:
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

  const handleClickItem = () => {
    propertyStoreMethods.setSelectedUUID(item.uuid);

    // TODO
  };

  const handleToggleDir = async () => {
    const propertyItem = getPropertyItemByUUID(item.uuid);
    if (!propertyItem) return;
    if (!propertyItem.loaded) {
      setLoading(true);
    }

    try {
      await propertyStoreMethods.triggerDirOpen(item.uuid);
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
    <>
      <div
        className={cn(
          "px-2 py-0.5 flex items-center hover:bg-accent hover:text-accent-foreground cursor-default select-none",
          selectedUUID === item.uuid &&
            "bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground",
        )}
        style={{
          paddingLeft: item.level * 8, // 根据层级增加左侧缩进
        }}
        onClick={handleClickItem}
        onDoubleClick={handleToggleDir}
      >
        <ChevronRight
          className={cn("size-3 mr-0.5 shrink-0", !item.isDir && "opacity-0")}
          style={{
            transition: "transform 0.2s",
            transform: item.opened ? "rotate(90deg)" : "rotate(0deg)",
          }}
          onClick={handleToggleDir}
        />
        {fileIcon}
        <span className=" truncate">{item.label}</span>
      </div>
      {item.opened && <FIleTreeList data={item.children || []} />}
    </>
  );
};

export default FileTreeItem;
