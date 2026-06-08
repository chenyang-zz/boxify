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

import { FC, ReactNode } from "react";
import { Loader2 } from "lucide-react";

interface SettingListProps {
  /** 是否正在加载 */
  loading?: boolean;
  /** 数据是否为空 */
  empty?: boolean;
  /** 空态主标题 */
  emptyText?: string;
  /** 空态副描述（可选） */
  emptyDescription?: string;
  /** 列表项渲染 */
  children: ReactNode;
}

/**
 * 设置列表容器 — 统一处理加载态和空态
 */
export const SettingList: FC<SettingListProps> = ({
  loading = false,
  empty = false,
  emptyText = "暂无数据",
  emptyDescription,
  children,
}) => {
  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (empty) {
    return (
      <div className="bg-card rounded-xl p-8 text-center border border-border">
        <p className="text-sm font-medium text-foreground">{emptyText}</p>
        {emptyDescription && (
          <p className="text-sm text-muted-foreground mt-1">{emptyDescription}</p>
        )}
      </div>
    );
  }

  return <div className="flex flex-col gap-3">{children}</div>;
};

export default SettingList;
