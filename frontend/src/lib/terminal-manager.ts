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

import { Terminal as XTerminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { Events } from "@wailsio/runtime";
import { TerminalConfig, TerminalService } from "@wails/service";

function normalizeTerminalOutput(data: string) {
  return data
    .replace(/ð/g, "") // python icon
    .replace(/î°/g, "")
    .replace(/î /g, "") // git branch
    .replace(/Â/g, "")
    .replace(/â±/g, "")
    .replace(/â/g, "x");
}

interface CachedTerminal {
  xterm: XTerminal;
  fitAddon: FitAddon;
  isInitialized: boolean;
  container: HTMLElement | null;
}

class TerminalManager {
  private cache = new Map<string, CachedTerminal>();

  // 获取或创建终端实例
  getOrCreate(sessionId: string): CachedTerminal {
    let cached = this.cache.get(sessionId);

    if (!cached) {
      const xterm = new XTerminal({
        cursorBlink: true,
        cursorStyle: "bar",
        lineHeight: 1.2,
        cursorWidth: 2,
        fontSize: 12,
        fontFamily: `
"JetBrainsMono Nerd",
monospace
  `,
        theme: {
          background: "#1e1e1e",
          foreground: "#c9d1d9",
          cursor: "#58a6ff",
          selectionBackground: "#264f78",
          black: "#484f58",
          red: "#ff7b72",
          green: "#3fb950",
          yellow: "#d29922",
          blue: "#58a6ff",
          magenta: "#bc8cff",
          cyan: "#39c5cf",
          white: "#b1bac4",
          brightBlack: "#6e7681",
          brightRed: "#ffa198",
          brightGreen: "#56d364",
          brightYellow: "#e3b341",
          brightBlue: "#79c0ff",
          brightMagenta: "#d2a8ff",
          brightCyan: "#56d4dd",
          brightWhite: "#ffffff",
        },
        scrollback: 1000,
        convertEol: true,
      });

      const fitAddon = new FitAddon();
      xterm.loadAddon(fitAddon);
      xterm.loadAddon(new WebLinksAddon());

      cached = {
        xterm,
        fitAddon,
        isInitialized: false,
        container: null,
      };

      this.cache.set(sessionId, cached);
      this.setupEventListeners(sessionId, xterm);
    }

    return cached;
  }

  // 设置事件监听器（只设置一次）
  private setupEventListeners(sessionId: string, xterm: XTerminal): void {
    // 监听用户输入并发送到后端
    xterm.onData((data) => {
      const encoded = btoa(data);
      TerminalService.Write(sessionId, encoded).catch((err) => {
        console.error("写入终端失败:", err);
      });
    });

    // 监听后端输出事件
    const unbindOutput = Events.On(
      "terminal:output",
      (event: { data: { sessionId: string; data: string } }) => {
        if (event.data.sessionId === sessionId) {
          try {
            const decoded = atob(event.data.data);
            console.log(decoded);
            xterm.write(normalizeTerminalOutput(decoded));
          } catch (e) {
            console.error("解码终端输出失败:", e);
          }
        }
      },
    );

    // 监听终端错误事件
    const unbindError = Events.On(
      "terminal:error",
      (event: { data: { sessionId: string; message: string } }) => {
        if (event.data.sessionId === sessionId) {
          xterm.write(`\r\n\x1b[31m错误: ${event.data.message}\x1b[0m\r\n`);
        }
      },
    );

    // 保存清理函数到 xterm 实例上
    (xterm as any)._cleanup = () => {
      unbindOutput();
      unbindError();
    };
  }

  // 初始化终端会话（调用后端 Create）
  async initialize(sessionId: string, shell: string): Promise<void> {
    const cached = this.cache.get(sessionId);
    if (!cached || cached.isInitialized) return;

    try {
      const res = await TerminalService.Create(
        TerminalConfig.createFrom({
          id: sessionId,
          shell,
          rows: cached.xterm.rows,
          cols: cached.xterm.cols,
        }),
      );

      if (res && res.success) {
        cached.isInitialized = true;
        console.log("终端会话创建成功:", sessionId);
      } else {
        cached.xterm.write(
          `\r\n\x1b[31m创建终端失败: ${res?.message || "未知错误"}\x1b[0m\r\n`,
        );
      }
    } catch (err) {
      console.error("创建终端失败:", err);
      cached.xterm.write(
        `\r\n\x1b[31m创建终端失败: ${(err as Error).message}\x1b[0m\r\n`,
      );
    }
  }

  // 将终端打开到容器（只调用一次）
  open(sessionId: string, container: HTMLElement): void {
    const cached = this.cache.get(sessionId);
    if (!cached) return;

    // 如果已经打开过，就不再重复打开
    if (cached.container) {
      return;
    }

    cached.xterm.open(container);
    cached.container = container;
    cached.fitAddon.fit();
  }

  // 调整终端大小
  resize(sessionId: string): void {
    const cached = this.cache.get(sessionId);
    if (!cached || !cached.container) return;

    cached.fitAddon.fit();
    const { rows, cols } = cached.xterm;
    TerminalService.Resize(sessionId, rows, cols).catch((err) => {
      console.error("调整终端大小失败:", err);
    });
  }

  // 销毁终端实例
  destroy(sessionId: string): void {
    const cached = this.cache.get(sessionId);
    if (!cached) return;

    // 调用清理函数
    if (typeof (cached.xterm as any)._cleanup === "function") {
      (cached.xterm as any)._cleanup();
    }

    // 关闭后端会话
    TerminalService.Close(sessionId).catch((err) => {
      console.error("关闭终端失败:", err);
    });

    // 销毁 xterm 实例
    cached.xterm.dispose();

    this.cache.delete(sessionId);
  }

  // 检查终端是否存在
  has(sessionId: string): boolean {
    return this.cache.has(sessionId);
  }

  // 获取终端实例（不创建）
  get(sessionId: string): CachedTerminal | undefined {
    return this.cache.get(sessionId);
  }
}

// 单例导出
export const terminalManager = new TerminalManager();
