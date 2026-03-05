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
import { Events } from "@wailsio/runtime";
import { Terminal as XTerm } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { Unicode11Addon } from "@xterm/addon-unicode11";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { terminalSessionManager } from "../lib/session-manager";
import "@xterm/xterm/css/xterm.css";

interface FullscreenTerminalProps {
  sessionId: string;
}

// 解码后端 Base64 编码输出，保留 UTF-8 多字节字符。
function decodeOutput(encodedData: string): string {
  const binaryString = atob(encodedData);
  const bytes = Uint8Array.from(binaryString, (c) => c.charCodeAt(0));
  return new TextDecoder("utf-8").decode(bytes);
}

// 全屏终端组件：使用 xterm 接管输入输出交互。
export function FullscreenTerminal({ sessionId }: FullscreenTerminalProps) {
  const hostRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!hostRef.current) return;

    const term = new XTerm({
      allowProposedApi: true,
      cursorBlink: true,
      fontFamily: '"Sarasa Mono SC", "JetBrainsMono Nerd Font", monospace',
      fontSize: 13,
      lineHeight: 1.35,
      scrollback: 5000,
      allowTransparency: true,
      theme: {
        background: "#00000000",
      },
    });

    const fitAddon = new FitAddon();
    const unicodeAddon = new Unicode11Addon();
    const webLinksAddon = new WebLinksAddon();

    term.loadAddon(fitAddon);
    term.loadAddon(unicodeAddon);
    term.loadAddon(webLinksAddon);
    term.unicode.activeVersion = "11";

    term.open(hostRef.current);

    // 同步 xterm 计算出的真实 rows/cols 到后端 PTY。
    const syncTerminalSize = () => {
      fitAddon.fit();
      terminalSessionManager.resize(sessionId, term.cols, term.rows);
    };

    syncTerminalSize();
    term.focus();

    const observer = new ResizeObserver(() => {
      syncTerminalSize();
    });
    observer.observe(hostRef.current);

    const dataDisposable = term.onData((data) => {
      terminalSessionManager.write(sessionId, data);
    });

    const unbindOutput = Events.On(
      "terminal:output",
      (event: {
        data: { sessionId: string; data: string; blockId?: string };
      }) => {
        if (event.data.sessionId !== sessionId) return;
        try {
          term.write(decodeOutput(event.data.data));
        } catch (err) {
          console.error("xterm 输出解码失败:", err);
        }
      },
    );

    const unbindError = Events.On(
      "terminal:error",
      (event: { data: { sessionId: string; message: string } }) => {
        if (event.data.sessionId !== sessionId) return;
        term.writeln(`\r\n\x1b[31m错误: ${event.data.message}\x1b[0m`);
      },
    );

    return () => {
      unbindOutput();
      unbindError();
      dataDisposable.dispose();
      observer.disconnect();
      term.dispose();
    };
  }, [sessionId]);

  return <div ref={hostRef} className="h-full w-full overflow-hidden" />;
}
