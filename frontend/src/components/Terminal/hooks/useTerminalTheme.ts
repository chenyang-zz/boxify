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

import { useTerminalStore } from "../store/terminal.store";
import {
  availableThemes,
  type TerminalTheme,
} from "../types/theme";

export function useTerminalTheme() {
  const currentTheme = useTerminalStore((state) => state.currentTheme);
  const setTheme = useTerminalStore((state) => state.setTheme);

  // 按名称设置主题
  const setThemeByName = (name: string) => {
    const theme = availableThemes.find((t) => t.name === name);
    if (theme) {
      setTheme(theme);
    }
  };

  // 获取所有可用主题
  const getAvailableThemes = () => availableThemes;

  return {
    theme: currentTheme,
    setTheme,
    setThemeByName,
    getAvailableThemes,
  };
}

// 导出主题相关
export { defaultTheme, draculaTheme, oneDarkTheme, availableThemes } from "../types/theme";
export type { TerminalTheme } from "../types/theme";
