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

interface TerminalConfig {
  shell: ShellType;
  workpath: string;
  initialCommand: string;
}

interface PropertyItemType {
  uuid: string;
  level: number;
  loading?: boolean;
  isDir: boolean;
  label: string;
  type: PropertyType;
  loaded?: boolean; // 是否已加载

  authMethod?: AuthMethod; // 认证方式
  remark?: string; // 备注信息

  // dir 属性
  opened?: boolean;
  children?: PropertyItemType[];

  // connection 属性
  connectionConfig?: ConnectionConfig; // 连接配置，具体结构根据连接类型而定

  // terminal 属性
  terminalConfig?: TerminalConfig; // 终端配置，包含 shell 类型和可选的工作路径

  extra?: Record<string, any>; // 其他额外属性，根据需要添加
  parent?: PropertyItemType; // 可选的父级引用，方便向上访问
}
