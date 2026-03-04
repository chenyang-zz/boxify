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

import { useCallback, useEffect, useRef, useState, type RefObject } from "react";
import type { TerminalConfig } from "@/types/property";
import { TerminalEnvironmentInfo } from "@wails/types/models";
import { terminalSessionManager } from "../lib/session-manager";

interface UseTerminalControllerParams {
  sessionId: string;
  config: TerminalConfig;
}

interface UseTerminalControllerResult {
  containerRef: RefObject<HTMLDivElement | null>;
  envInfo?: TerminalEnvironmentInfo;
}

// 终端控制器：封装会话初始化、环境同步与终端尺寸联动。
export function useTerminalController({
  sessionId,
  config,
}: UseTerminalControllerParams): UseTerminalControllerResult {
  const containerRef = useRef<HTMLDivElement>(null);
  const resizeRafRef = useRef<number | null>(null);
  const [envInfo, setEnvInfo] = useState<TerminalEnvironmentInfo | undefined>(
    undefined,
  );

  // 监听后端环境变化并同步到前端状态。
  const handleEnvChange = useCallback((env: TerminalEnvironmentInfo) => {
    setEnvInfo(env);
  }, []);

  // 初始化终端会话，并绑定环境变化回调。
  useEffect(() => {
    let cancelled = false;

    const session = terminalSessionManager.getOrCreate(sessionId);
    terminalSessionManager.setEnvChangeCallback(sessionId, handleEnvChange);

    if (!session.isInitialized) {
      terminalSessionManager.initialize(sessionId, config).then((env) => {
        if (cancelled) return;
        setEnvInfo(env);
      });
    } else {
      setEnvInfo(session.environmentInfo);
    }

    return () => {
      cancelled = true;
      terminalSessionManager.setEnvChangeCallback(sessionId, undefined);
    };
  }, [sessionId, config, handleEnvChange]);

  // 监听容器尺寸变化并同步 rows/cols 到后端伪终端。
  useEffect(() => {
    const updateSize = () => {
      if (!containerRef.current) return;
      const { clientWidth, clientHeight } = containerRef.current;
      const charWidth = 8;
      const charHeight = 16;
      const cols = Math.max(20, Math.floor(clientWidth / charWidth));
      const rows = Math.max(6, Math.floor(clientHeight / charHeight));
      terminalSessionManager.resize(sessionId, cols, rows);
    };

    const scheduleResize = () => {
      if (resizeRafRef.current !== null) {
        cancelAnimationFrame(resizeRafRef.current);
      }
      resizeRafRef.current = requestAnimationFrame(() => {
        updateSize();
        resizeRafRef.current = null;
      });
    };

    scheduleResize();

    const observer = new ResizeObserver(() => {
      scheduleResize();
    });

    if (containerRef.current) {
      observer.observe(containerRef.current);
    }

    return () => {
      observer.disconnect();
      if (resizeRafRef.current !== null) {
        cancelAnimationFrame(resizeRafRef.current);
      }
    };
  }, [sessionId]);

  return {
    containerRef,
    envInfo,
  };
}
