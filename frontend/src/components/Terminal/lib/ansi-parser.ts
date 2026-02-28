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

import type { TextStyle, FormattedChar } from "../types/block";
import type { TerminalTheme } from "../types/theme";

type ParserState = "normal" | "escape" | "csi" | "osc";

interface ParserContext {
  state: ParserState;
  params: number[];
  currentParam: string;
  oscBuffer: string;
}

// ANSI 颜色索引
const ANSI_COLORS = [
  "black",
  "red",
  "green",
  "yellow",
  "blue",
  "magenta",
  "cyan",
  "white",
] as const;

/**
 * 终端缓冲区类 - 处理控制字符和光标移动
 * 模拟真实终端的行为，正确处理回车符、退格符等控制字符
 */
class TerminalBuffer {
  private lines: FormattedChar[][] = [[]];
  private cursorX = 0;
  private cursorY = 0;

  /**
   * 写入字符到缓冲区
   * @param char - 要写入的字符
   * @param style - 字符样式
   */
  write(char: string, style: TextStyle): void {
    switch (char) {
      case "\r": // 回车 - 移动光标到行首
        this.cursorX = 0;
        break;

      case "\n": // 换行 - 移动到下一行
        this.cursorY++;
        while (this.lines.length <= this.cursorY) {
          this.lines.push([]);
        }
        break;

      case "\b": // 退格 - 光标左移一位
        if (this.cursorX > 0) {
          this.cursorX--;
        }
        break;

      case "\t": // 制表符 - 移动到下一个制表位（每8字符）
        this.cursorX = Math.floor(this.cursorX / 8 + 1) * 8;
        break;

      default:
        // 只处理可打印字符
        if (char >= " ") {
          // 确保有足够的行
          while (this.lines.length <= this.cursorY) {
            this.lines.push([]);
          }

          const line = this.lines[this.cursorY];

          if (this.cursorX < line.length) {
            // 覆盖已存在的字符（回车符的行为）
            line[this.cursorX] = { char, style: { ...style } };
          } else {
            // 在行末尾添加字符，填充空格
            while (line.length < this.cursorX) {
              line.push({ char: " ", style: {} });
            }
            line.push({ char, style: { ...style } });
          }
          this.cursorX++;
        }
    }
  }

  /**
   * 获取所有行的内容
   */
  getLines(): FormattedChar[][] {
    return this.lines;
  }

  /**
   * 将缓冲区内容扁平化为 FormattedChar 数组
   * 每行之间插入换行符
   */
  toFlatArray(): FormattedChar[] {
    const result: FormattedChar[] = [];

    for (let i = 0; i < this.lines.length; i++) {
      result.push(...this.lines[i]);
      // 行间添加换行符（最后一行不加）
      if (i < this.lines.length - 1) {
        result.push({ char: "\n", style: {} });
      }
    }

    return result;
  }
}

export class AnsiParser {
  private context: ParserContext;
  private theme: TerminalTheme;

  constructor(theme: TerminalTheme) {
    this.context = this.createInitialContext();
    this.theme = theme;
  }

  private createInitialContext(): ParserContext {
    return {
      state: "normal",
      params: [],
      currentParam: "",
      oscBuffer: "",
    };
  }

  setTheme(theme: TerminalTheme) {
    this.theme = theme;
  }

  // 解析 ANSI 序列并返回格式化字符
  parse(input: string): FormattedChar[] {
    // 使用 TerminalBuffer 处理控制字符
    const buffer = new TerminalBuffer();
    let currentStyle: TextStyle = {};
    this.context = this.createInitialContext();

    for (let i = 0; i < input.length; i++) {
      const char = input[i];

      switch (this.context.state) {
        case "normal":
          if (char === "\x1b") {
            this.context.state = "escape";
          } else {
            // 所有字符（包括控制字符）都通过 buffer 处理
            buffer.write(char, currentStyle);
          }
          break;

        case "escape":
          if (char === "[") {
            this.context.state = "csi";
            this.context.params = [];
            this.context.currentParam = "";
          } else if (char === "]") {
            this.context.state = "osc";
            this.context.oscBuffer = "";
          } else {
            this.context.state = "normal";
          }
          break;

        case "csi":
          if (char >= "0" && char <= "9") {
            this.context.currentParam += char;
          } else if (char === ";") {
            this.context.params.push(
              parseInt(this.context.currentParam, 10) || 0,
            );
            this.context.currentParam = "";
          } else if (char === "?") {
            // Private mode indicator, skip
          } else if (char >= "@" && char <= "~") {
            // 终止字符
            this.context.params.push(
              parseInt(this.context.currentParam, 10) || 0,
            );
            currentStyle = this.applySgr(this.context.params, currentStyle);
            this.context.state = "normal";
          } else {
            // 未知序列，重置
            this.context.state = "normal";
          }
          break;

        case "osc":
          if (char === "\x07" || (char === "\\" && input[i - 1] === "\x1b")) {
            // OSC 结束
            this.handleOsc(this.context.oscBuffer);
            this.context.state = "normal";
          } else if (char !== "\x1b") {
            this.context.oscBuffer += char;
          }
          break;
      }
    }

    // 返回扁平化的结果
    return buffer.toFlatArray();
  }

  // 应用 SGR (Select Graphic Rendition) 参数
  private applySgr(params: number[], currentStyle: TextStyle): TextStyle {
    const style = { ...currentStyle };

    for (let i = 0; i < params.length; i++) {
      const param = params[i];

      switch (param) {
        case 0: // 重置所有样式
          return {};

        case 1: // 粗体
          style.bold = true;
          break;

        case 2: // 暗淡
          style.dim = true;
          break;

        case 3: // 斜体
          style.italic = true;
          break;

        case 4: // 下划线
          style.underline = true;
          break;

        case 5: // 慢闪烁
        case 6: // 快闪烁
          style.blink = true;
          break;

        case 7: // 反色
          style.inverse = true;
          break;

        case 8: // 隐藏
          style.hidden = true;
          break;

        case 9: // 删除线
          style.strikethrough = true;
          break;

        case 22: // 正常亮度
          style.bold = false;
          style.dim = false;
          break;

        case 23: // 非斜体
          style.italic = false;
          break;

        case 24: // 非下划线
          style.underline = false;
          break;

        case 27: // 非反色
          style.inverse = false;
          break;

        case 28: // 非隐藏
          style.hidden = false;
          break;

        case 29: // 非删除线
          style.strikethrough = false;
          break;

        // 前景色 30-37
        case 30:
        case 31:
        case 32:
        case 33:
        case 34:
        case 35:
        case 36:
        case 37:
          style.fg = this.getAnsiColor(param - 30, false);
          break;

        // 38: 256色或真彩色
        case 38:
          if (params[i + 1] === 5 && params[i + 2] !== undefined) {
            // 256色
            style.fg = this.get256Color(params[i + 2]);
            i += 2;
          } else if (params[i + 1] === 2 && params[i + 4] !== undefined) {
            // 真彩色 RGB
            style.fg = this.getRgbColor(
              params[i + 2],
              params[i + 3],
              params[i + 4],
            );
            i += 4;
          }
          break;

        // 39: 默认前景色
        case 39:
          style.fg = undefined;
          break;

        // 背景色 40-47
        case 40:
        case 41:
        case 42:
        case 43:
        case 44:
        case 45:
        case 46:
        case 47:
          style.bg = this.getAnsiColor(param - 40, false);
          break;

        // 48: 256色或真彩色背景
        case 48:
          if (params[i + 1] === 5 && params[i + 2] !== undefined) {
            style.bg = this.get256Color(params[i + 2]);
            i += 2;
          } else if (params[i + 1] === 2 && params[i + 4] !== undefined) {
            style.bg = this.getRgbColor(
              params[i + 2],
              params[i + 3],
              params[i + 4],
            );
            i += 4;
          }
          break;

        // 49: 默认背景色
        case 49:
          style.bg = undefined;
          break;

        // 亮前景色 90-97
        case 90:
        case 91:
        case 92:
        case 93:
        case 94:
        case 95:
        case 96:
        case 97:
          style.fg = this.getAnsiColor(param - 90, true);
          break;

        // 亮背景色 100-107
        case 100:
        case 101:
        case 102:
        case 103:
        case 104:
        case 105:
        case 106:
        case 107:
          style.bg = this.getAnsiColor(param - 100, true);
          break;
      }
    }

    return style;
  }

  private handleOsc(buffer: string): void {
    // 处理 OSC 序列（如窗口标题）
    // 目前忽略，但保留接口
    const [code, ...data] = buffer.split(";");
    // OSC 0 或 2: 设置窗口标题
    // OSC 7: 设置当前目录
    // 暂不处理
  }

  private getAnsiColor(index: number, bright: boolean): string {
    const colorName = ANSI_COLORS[index];
    if (bright) {
      return this.theme[
        `bright${colorName.charAt(0).toUpperCase() + colorName.slice(1)}` as keyof TerminalTheme
      ] as string;
    }
    return this.theme[colorName as keyof TerminalTheme] as string;
  }

  private get256Color(index: number): string {
    // 标准 16 色
    if (index < 16) {
      const bright = index >= 8;
      const colorIndex = bright ? index - 8 : index;
      return this.getAnsiColor(colorIndex, bright);
    }

    // 216 色调色板 (16-231)
    if (index < 232) {
      const n = index - 16;
      const r = Math.floor(n / 36) * 51;
      const g = (Math.floor(n / 6) % 6) * 51;
      const b = (n % 6) * 51;
      return this.getRgbColor(r, g, b);
    }

    // 灰度 (232-255)
    const gray = (index - 232) * 10 + 8;
    return this.getRgbColor(gray, gray, gray);
  }

  private getRgbColor(r: number, g: number, b: number): string {
    return `rgb(${r}, ${g}, ${b})`;
  }
}

// 检测命令边界（简化版本）
export function detectCommandBoundary(data: string): {
  type: "prompt" | "output" | "command_end";
  command?: string;
} {
  // 简单检测：如果包含常见的 prompt 模式
  const promptPatterns = [
    /\$ $/, // bash/zsh $
    /# $/, // root #
    /> $/, // Windows cmd
    /❯ $/, // starship/fish
    /➜ $/, // agnoster
  ];

  for (const pattern of promptPatterns) {
    if (pattern.test(data)) {
      return { type: "prompt" };
    }
  }

  return { type: "output" };
}
