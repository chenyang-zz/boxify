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
import { FIleTreeList } from ".";
import { useFileTreeContext } from "./context";
import { cn } from "@/lib/utils";
import { ChevronRight, File } from "lucide-react";
import { ConnectionType, FileSystemType, FileType } from "@/common/constrains";
import { PropertyItemType } from "@/store/property.store";

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
        return <img className="size-5 mr-1 shrink-0" src="/icons/folder.svg" />;
      case ConnectionType.MYSQL:
        return <img className="size-5 mr-1 shrink-0" src="/icons/mysql.svg" />;
      case ConnectionType.REDIS:
        return <img className="size-5 mr-1 shrink-0" src="/icons/redis.svg" />;
      case ConnectionType.MONGODB:
        return (
          <img className="size-5 mr-1 shrink-0" src="/icons/mongodb.svg" />
        );
    }
  }
  switch (type) {
    case ConnectionType.SSH:
      return <img className="size-5 mr-1 shrink-0" src="/icons/terminal.svg" />;
    default:
      return <File className="size-5 mr-1 shrink-0" />;
  }
};

const FileTreeItem: FC<FileTreeItemProps> = ({ item }) => {
  const { selectedUUID, setSelectedUUID, triggerDirOpen } =
    useFileTreeContext();

  const handleClickItem = () => {
    setSelectedUUID(item.uuid);

    // TODO
  };

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
        onDoubleClick={() => triggerDirOpen(item.uuid)}
      >
        <ChevronRight
          className={cn("size-3 mr-0.5 shrink-0", !item.isDir && "opacity-0")}
          style={{
            transition: "transform 0.2s",
            transform: item.opened ? "rotate(90deg)" : "rotate(0deg)",
          }}
          onClick={() => triggerDirOpen(item.uuid)}
        />
        <FileIcon isDir={item.isDir} type={item.type} />
        <span className=" truncate">{item.label}</span>
      </div>
      {item.opened && <FIleTreeList data={item.children || []} />}
    </>
  );
};

export default FileTreeItem;
