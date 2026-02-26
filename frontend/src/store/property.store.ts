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

import { PropertyType } from "@/common/constrains";
import { ConnectionConfig } from "@/types";
import { create } from "zustand";
import { persist } from "zustand/middleware";

import { safeStorage, StoreMethods } from "./common";
import { initTraverseTree } from "@/lib/property";
import { AuthMethod } from "@/common/enums/connection";

export const FileTreeMap = new Map<string, PropertyItemType>();

export interface PropertyItemType {
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

  extra?: Record<string, any>; // 其他额外属性，根据需要添加
  parent?: PropertyItemType; // 可选的父级引用，方便向上访问
}

interface PropertyState {
  propertyList: PropertyItemType[];
  setPropertyList: (list: PropertyItemType[]) => void;
  selectedUUID: string; // 当前选中的文件或文件夹的UUID
  setSelectedUUID: (uuid: string) => void; // 更新选中的UUID
}

export const usePropertyStore = create<PropertyState>()(
  persist(
    (set, get) => ({
      selectedUUID: "",
      setSelectedUUID: (uuid: string) => {
        set(() => ({ selectedUUID: uuid }));
      },
      propertyList: [],
      setPropertyList: (list: PropertyItemType[]) => {
        set(() => ({ propertyList: list }));
      },
    }),
    {
      name: "boxify-property-store",
      version: 1,
      onRehydrateStorage: () => (state) => {
        if (state?.propertyList) {
          initTraverseTree(state.propertyList);
        }
      },
      migrate: (persistedState, version) => {
        console.log(persistedState);
        return persistedState as PropertyState;
      },
      storage: safeStorage,
      partialize: (state) => ({
        propertyList: state.propertyList,
      }),
    },
  ),
);

export const propertyStoreMethods = StoreMethods(usePropertyStore);
