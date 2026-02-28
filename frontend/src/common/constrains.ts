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

import { ShellType } from "@wails/terminal";
import { AuthMethod, Environment } from "./enums/connection";

export enum ConnectionEnum {
  TERMINAL = "terminal",
  SSH = "ssh",
  MYSQL = "mysql",
  POSTGRESQL = "postgresql",
  MONGODB = "mongodb",
  REDIS = "redis",
}

export enum FileSystemType {
  FOLDER = "folder",
  UNKNOWN = "unknown",
}

// 数据库文件类型枚举
export enum DBFileType {
  TABLE_FOLDER = "table_folder",
  VIEW_FOLDER = "view_folder",
  QUERY_FOLDER = "query_folder",
  FUNCTION_FOLDER = "function_folder",
  DATABASE = "database",
  TABLE = "table",
}

export type PropertyType = ConnectionEnum | FileSystemType | DBFileType;

export enum TabType {
  TABLE = "table",
  TERMINAL = "terminal",
}

export function isDirType(type: PropertyType): boolean {
  switch (type) {
    case FileSystemType.UNKNOWN:
    case ConnectionEnum.SSH:
    case ConnectionEnum.TERMINAL:
      return false;
    case FileSystemType.FOLDER:
    case ConnectionEnum.MYSQL:
    case ConnectionEnum.POSTGRESQL:
    case ConnectionEnum.MONGODB:
    case ConnectionEnum.REDIS:
      return true;
    default:
      return false;
  }
}

// 判断是否是连接类型
export function isConnectionType(type: PropertyType): boolean {
  return Object.values(ConnectionEnum).includes(type as ConnectionEnum);
}

// 判断是否是数据库相关类型
export function isDBType(type: PropertyType): boolean {
  switch (type) {
    case ConnectionEnum.MYSQL:
    case ConnectionEnum.POSTGRESQL:
    case ConnectionEnum.MONGODB:
    case ConnectionEnum.REDIS:
    case DBFileType.DATABASE:
      return true;
    default:
      return false;
  }
}

export const AuthMethodOptions = [
  { label: "密码", value: AuthMethod.Password },
  { label: "每次询问", value: AuthMethod.InQuiry },
];

export const EnvironmentOptions = [
  { label: "无", value: Environment.None },
  { label: "开发", value: Environment.Development },
  { label: "测试", value: Environment.Testing },
  { label: "生产", value: Environment.Production },
];

export const ShellOptions = [
  { label: "auto", value: ShellType.ShellTypeAuto },
  { label: "cmd", value: ShellType.ShellTypeCmd },
  { label: "powershell", value: ShellType.ShellTypePowershell },
  { label: "pwsh", value: ShellType.ShellTypePwsh },
  { label: "bash", value: ShellType.ShellTypeBash },
  { label: "zsh", value: ShellType.ShellTypeZsh },
  { label: "sh", value: ShellType.ShellTypeSh },
];
