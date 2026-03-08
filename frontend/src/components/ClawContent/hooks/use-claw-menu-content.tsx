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

import { useMemo, type ReactNode } from "react";
import { clawMenuPanelRenderers, isClawMenuItemId } from "../constants";

/**
 * 根据菜单项解析 Claw 主内容区域。
 */
export function useClawMenuContent(selectedMenuItem: string): ReactNode {
  return useMemo(() => {
    if (!isClawMenuItemId(selectedMenuItem)) {
      return (
        <div className="h-full flex items-center justify-center text-muted-foreground">
          <p>请从左侧选择一个菜单项</p>
        </div>
      );
    }

    return clawMenuPanelRenderers[selectedMenuItem]();
  }, [selectedMenuItem]);
}
