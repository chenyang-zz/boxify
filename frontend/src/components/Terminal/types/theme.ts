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

export interface TerminalTheme {
  name: string;
  displayName: string;

  // 基础颜色
  background: string;
  foreground: string;
  cursor: string;
  cursorAccent: string;
  selectionBackground: string;
  selectionForeground: string;

  // ANSI 16 色
  black: string;
  red: string;
  green: string;
  yellow: string;
  blue: string;
  magenta: string;
  cyan: string;
  white: string;
  brightBlack: string;
  brightRed: string;
  brightGreen: string;
  brightYellow: string;
  brightBlue: string;
  brightMagenta: string;
  brightCyan: string;
  brightWhite: string;

  // 字体设置
  fontFamily: string;
  fontSize: number;
  lineHeight: number;
  letterSpacing: number;

  // Block 样式
  blockStyle: {
    background: string;
    borderColor: string;
    borderRadius: number;
  };

  // 输入框样式
  inputStyle: {
    background: string;
    borderColor: string;
    focusBorderColor: string;
  };
}

// 默认主题（兼容现有 xterm 配色）
export const defaultTheme: TerminalTheme = {
  name: "default",
  displayName: "默认主题",

  background: "#1e1e1e",
  foreground: "#c9d1d9",
  cursor: "#58a6ff",
  cursorAccent: "#1e1e1e",
  selectionBackground: "#264f78",
  selectionForeground: "#c9d1d9",

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

  fontFamily: '"Sarasa Mono SC", "JetBrainsMono Nerd", "Fira Code", "Consolas", monospace',
  fontSize: 13,
  lineHeight: 1.4,
  letterSpacing: 0,

  blockStyle: {
    background: "rgba(255, 255, 255, 0.02)",
    borderColor: "rgba(255, 255, 255, 0.1)",
    borderRadius: 8,
  },

  inputStyle: {
    background: "rgba(0, 0, 0, 0.3)",
    borderColor: "rgba(255, 255, 255, 0.1)",
    focusBorderColor: "#58a6ff",
  },
};

// Dracula 主题
export const draculaTheme: TerminalTheme = {
  name: "dracula",
  displayName: "Dracula",

  background: "#282a36",
  foreground: "#f8f8f2",
  cursor: "#f8f8f2",
  cursorAccent: "#282a36",
  selectionBackground: "#44475a",
  selectionForeground: "#f8f8f2",

  black: "#21222c",
  red: "#ff5555",
  green: "#50fa7b",
  yellow: "#f1fa8c",
  blue: "#bd93f9",
  magenta: "#ff79c6",
  cyan: "#8be9fd",
  white: "#f8f8f2",
  brightBlack: "#6272a4",
  brightRed: "#ff6e6e",
  brightGreen: "#69ff94",
  brightYellow: "#ffffa5",
  brightBlue: "#d6acff",
  brightMagenta: "#ff92df",
  brightCyan: "#a4ffff",
  brightWhite: "#ffffff",

  fontFamily: '"Sarasa Mono SC", "JetBrainsMono Nerd", "Fira Code", "Consolas", monospace',
  fontSize: 13,
  lineHeight: 1.4,
  letterSpacing: 0,

  blockStyle: {
    background: "rgba(255, 255, 255, 0.03)",
    borderColor: "rgba(255, 255, 255, 0.08)",
    borderRadius: 8,
  },

  inputStyle: {
    background: "rgba(0, 0, 0, 0.2)",
    borderColor: "rgba(255, 255, 255, 0.1)",
    focusBorderColor: "#bd93f9",
  },
};

// One Dark 主题
export const oneDarkTheme: TerminalTheme = {
  name: "one-dark",
  displayName: "One Dark",

  background: "#282c34",
  foreground: "#abb2bf",
  cursor: "#528bff",
  cursorAccent: "#282c34",
  selectionBackground: "#3e4451",
  selectionForeground: "#abb2bf",

  black: "#282c34",
  red: "#e06c75",
  green: "#98c379",
  yellow: "#e5c07b",
  blue: "#61afef",
  magenta: "#c678dd",
  cyan: "#56b6c2",
  white: "#abb2bf",
  brightBlack: "#5c6370",
  brightRed: "#e06c75",
  brightGreen: "#98c379",
  brightYellow: "#e5c07b",
  brightBlue: "#61afef",
  brightMagenta: "#c678dd",
  brightCyan: "#56b6c2",
  brightWhite: "#ffffff",

  fontFamily: '"Sarasa Mono SC", "JetBrainsMono Nerd", "Fira Code", "Consolas", monospace',
  fontSize: 13,
  lineHeight: 1.4,
  letterSpacing: 0,

  blockStyle: {
    background: "rgba(255, 255, 255, 0.02)",
    borderColor: "rgba(255, 255, 255, 0.08)",
    borderRadius: 8,
  },

  inputStyle: {
    background: "rgba(0, 0, 0, 0.2)",
    borderColor: "rgba(255, 255, 255, 0.1)",
    focusBorderColor: "#61afef",
  },
};

// 所有可用主题
export const availableThemes: TerminalTheme[] = [
  defaultTheme,
  draculaTheme,
  oneDarkTheme,
];
