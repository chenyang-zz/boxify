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

import { FC, JSX, lazy, useEffect } from "react";
import { dataSyncStoreMethods } from "./store/data-sync.store";
import { eventStoreMethods } from "./store/event.store";
import { currentPageId, currentWindowName } from "./lib/utils";
import { Toaster } from "./components/ui/sonner";
import { ThemeProvider } from "./providers/ThemeProvider";

// 获取当前页面 ID
const pageId = currentPageId();

// 动态页面映射
const pageComponents: Record<
  string,
  React.LazyExoticComponent<() => JSX.Element>
> = {
  index: lazy(() => import("./pages/main")),
  settings: lazy(() => import("./pages/settings")),
  "connection-edit": lazy(() => import("./pages/connectionEdit")),
};

const PageComponent = pageComponents[pageId] || pageComponents.index;

const App: FC = () => {
  useEffect(() => {
    console.log(`📄 当前页面ID: ${pageId}`);
    const windowName = currentWindowName();
    console.log(`📄  当前窗口名称: ${windowName}`);
    dataSyncStoreMethods.setCurrentWindow(windowName);
    // eventStoreMethods.initialize();

    // return () => {
    //   eventStoreMethods.dispose();
    // };
  }, []);

  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <Toaster />
      <PageComponent />
    </ThemeProvider>
  );
};

export default App;
