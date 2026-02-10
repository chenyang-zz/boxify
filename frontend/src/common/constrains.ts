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

export enum ConnectionType {
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

export type FileType = ConnectionType | FileSystemType | DBFileType;

export function isDirType(type: FileType): boolean {
  switch (type) {
    case FileSystemType.UNKNOWN:
    case ConnectionType.SSH:
      return false;
    case FileSystemType.FOLDER:
    case ConnectionType.MYSQL:
    case ConnectionType.POSTGRESQL:
    case ConnectionType.MONGODB:
    case ConnectionType.REDIS:
      return true;
    default:
      return false;
  }
}

export function isDBConnection(type: FileType): boolean {
  switch (type) {
    case ConnectionType.MYSQL:
    case ConnectionType.POSTGRESQL:
    case ConnectionType.MONGODB:
    case ConnectionType.REDIS:
    case DBFileType.DATABASE:
      return true;
    default:
      return false;
  }
}
