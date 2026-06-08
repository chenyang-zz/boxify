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

import { Terminal } from "lucide-react";
import { ToolBadge } from "./ToolBadge";

export interface BashToolProps {
  /** Bash 命令文本 */
  label: string;
  /** 点击回调 */
  onClick?: () => void;
  className?: string;
}

/**
 * Bash 工具组件
 *
 * 展示 Bash 命令调用的工具徽章，使用 Terminal 图标。
 */
export function BashTool({ label, onClick, className }: BashToolProps) {
  return (
    <ToolBadge
      icon={Terminal}
      label={label}
      onClick={onClick}
      className={className}
    />
  );
}

export default BashTool;
