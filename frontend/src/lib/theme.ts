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
import { parse, formatHex } from "culori";

// 读取全局主题变量，供 xterm 同步应用当前 Tailwind 主题配色。
export function resolveCssVar(name: string, fallback: string): string {
  const value = getComputedStyle(document.documentElement)
    .getPropertyValue(name)
    .trim();
  return oklchToHex(value) || fallback;
}

// 将 CSS oklch 颜色值转换为 hex
export function oklchToHex(color: string): string {
  let value = color.trim();
  const parsed = parse(value);

  if (!parsed) {
    console.warn("Invalid color:", value);
    return "#000000";
  }

  return formatHex(parsed);
}
