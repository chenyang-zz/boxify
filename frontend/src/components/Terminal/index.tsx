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

import { useEffect, useRef } from "react";
import { terminalManager } from "@/lib/terminal-manager";
import "./style.css";

interface TerminalComponentProps {
  sessionId: string;
  shell?: string;
}

export default function Terminal({
  sessionId,
  shell = "auto",
}: TerminalComponentProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const isOpenedRef = useRef(false);

  useEffect(() => {
    if (!containerRef.current || isOpenedRef.current) return;

    // 获取或创建缓存的终端实例
    const cached = terminalManager.getOrCreate(sessionId);

    // 如果还没初始化，初始化后端会话
    if (!cached.isInitialized) {
      terminalManager.initialize(sessionId, shell);
    }

    // 打开终端到容器（只执行一次）
    terminalManager.open(sessionId, containerRef.current);
    isOpenedRef.current = true;

    // 调整大小以适应容器
    const resizeTimer = setTimeout(() => {
      terminalManager.resize(sessionId);
    }, 0);

    return () => {
      clearTimeout(resizeTimer);
    };
  }, [sessionId, shell]);

  // 监听窗口大小变化
  useEffect(() => {
    const handleResize = () => {
      if (isOpenedRef.current) {
        terminalManager.resize(sessionId);
      }
    };

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, [sessionId]);

  return (
    <div
      ref={containerRef}
      className="terminal-wrapper h-full w-full bg-background"
    />
  );
}
