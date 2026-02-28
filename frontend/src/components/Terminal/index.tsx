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

import { useMemo, useEffect, useRef } from "react";
import { TerminalCore } from "./TerminalCore";
import { terminalSessionManager } from "./lib/session-manager";
import { getPropertyItemByUUID } from "@/lib/property";

interface TerminalComponentProps {
  sessionId: string;
}

export default function Terminal({ sessionId }: TerminalComponentProps) {
  const isOpenedRef = useRef(false);

  const propertyItem = useMemo(
    () => getPropertyItemByUUID(sessionId),
    [sessionId],
  );

  // 清理：当组件卸载时销毁终端会话
  useEffect(() => {
    return () => {
      // 注意：这里不销毁会话，由 tabs.store.ts 中的 closeTab 处理
      // 这样可以避免在切换标签页时误销毁终端
    };
  }, [sessionId]);

  // 确保会话已创建
  useEffect(() => {
    if (!propertyItem?.terminalConfig) return;

    // 预先创建会话缓存
    terminalSessionManager.getOrCreate(sessionId);
    isOpenedRef.current = true;
  }, [sessionId, propertyItem]);

  if (!propertyItem?.terminalConfig) {
    return (
      <div className="terminal-error h-full w-full flex items-center justify-center text-muted-foreground">
        终端配置无效
      </div>
    );
  }

  return (
    <div className="terminal-wrapper h-full w-full" style={{ textAlign: "left" }}>
      <TerminalCore
        sessionId={sessionId}
        config={propertyItem.terminalConfig}
      />
    </div>
  );
}

// 导出子组件和 hooks
export { TerminalCore } from "./TerminalCore";
export { TerminalBlock } from "./components/TerminalBlock";
export { InputEditor } from "./components/InputEditor";
export { OutputRenderer } from "./components/OutputRenderer";
export { useTerminalTheme } from "./hooks/useTerminalTheme";
export {
  useTerminalStore,
  useSessionBlocks,
  useCurrentBlockId,
  useSessionTheme,
} from "./store/terminal.store";
export { terminalSessionManager } from "./lib/session-manager";
