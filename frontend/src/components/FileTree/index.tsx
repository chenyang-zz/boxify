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
import FileTreeItem from "./FileTreeItem";
import { usePropertyStore } from "@/store/property.store";
import { useDataSync } from "@/hooks/useDataSync";
import { DataChannel } from "@/store/data-sync.store";
import {
  addPropertyItemToParent,
  closeConnectionByUUID,
  getClosestFolder,
  getPropertyItemByUUID,
} from "@/lib/property";
import { ConnectionEnum } from "@/common/constrains";
import { ConnectionConfig, ConnectionType } from "@wails/connection";
import { v4 } from "uuid";
import { PropertyItemType } from "@/types/property";

interface FileTreeProps {
  data: PropertyItemType[];
}

export const FIleTreeList: FC<FileTreeProps> = ({ data }) => {
  return (
    <div className="text-foreground text-sm flex-1 overflow-auto">
      {data.map((item, index) => (
        <FileTreeItem key={index} item={item} />
      ))}
    </div>
  );
};

const FileTree: FC = () => {
  // const folder = getClosestFolder()
  const selectedUUID = usePropertyStore((state) => state.selectedUUID);

  useDataSync(DataChannel.Connection, (event) => {
    if (event.dataType === "connection:save") {
      console.log("[主窗口] 收到连接保存", event.data);
      let pUuid: string | null = null;
      let folder: PropertyItemType | null = null;
      if (selectedUUID) {
        folder = getClosestFolder(selectedUUID);
        pUuid = folder?.uuid ?? null;
      }

      let newPropertyItem: PropertyItemType | undefined;
      switch (event.data.type) {
        case ConnectionEnum.MYSQL:
          newPropertyItem = {
            uuid: v4(),
            level: folder ? folder.level + 1 : 1,
            isDir: true,
            label: event.data.name,
            type: ConnectionEnum.MYSQL,
            remark: event.data.remark,
            authMethod: event.data.authMethod,
            connectionConfig: ConnectionConfig.createFrom({
              type: ConnectionType.ConnectionTypeMySQL,
              host: event.data.host,
              port: event.data.port,
              user: event.data.user,
              password: event.data.password,
              useSSH: false,
            }),
          };
          break;
        case ConnectionEnum.TERMINAL:
          newPropertyItem = {
            uuid: v4(),
            level: folder ? folder.level + 1 : 1,
            isDir: false,
            label: event.data.name,
            type: ConnectionEnum.TERMINAL,
            remark: event.data.remark ?? "",
            terminalConfig: {
              shell: event.data.shell,
              workpath: event.data.workpath ?? "",
              initialCommand: event.data.initialCommand ?? "",
            },
          };
          break;
      }

      if (newPropertyItem) {
        addPropertyItemToParent(pUuid, newPropertyItem);
      }
    } else if (event.dataType === "connection:update") {
      console.log("[主窗口] 收到连接更新", event.data);
      const uuid = event.data.uuid;
      const item = getPropertyItemByUUID(uuid);
      if (!item) return;

      switch (event.data.type) {
        case ConnectionEnum.MYSQL:
          item.label = event.data.name;
          item.remark = event.data.remark;
          item.authMethod = event.data.authMethod;
          item.connectionConfig = ConnectionConfig.createFrom({
            type: ConnectionType.ConnectionTypeMySQL,
            host: event.data.host,
            port: event.data.port,
            user: event.data.user,
            password: event.data.password,
            useSSH: false,
          });
          break;
        case ConnectionEnum.TERMINAL:
          item.label = event.data.name;
          item.remark = event.data.remark ?? "";
          if (item.terminalConfig) {
            item.terminalConfig.shell = event.data.shell;
            item.terminalConfig.workpath = event.data.workpath ?? "";
          } else {
            item.terminalConfig = {
              shell: event.data.shell,
              workpath: event.data.workpath ?? "",
              initialCommand: event.data.initialCommand ?? "",
            };
          }
          break;
      }

      // 需要重新加载
      closeConnectionByUUID(uuid);
    }
  });
  const propertyList = usePropertyStore((state) => state.propertyList);

  return <FIleTreeList data={propertyList}></FIleTreeList>;
};

export default FileTree;
