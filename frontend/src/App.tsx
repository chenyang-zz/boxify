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

import { FC, JSX, lazy } from "react";
import { useWindowListener } from "./hooks/useWindowListener";

// 获取当前页面 ID
const pageId =
  document.querySelector('meta[name="page-id"]')?.getAttribute("content") ||
  "index";

// 动态页面映射
const pageComponents: Record<
  string,
  React.LazyExoticComponent<() => JSX.Element>
> = {
  index: lazy(() => import("./pages/main")),
  settings: lazy(() => import("./pages/settings")),
};

const PageComponent = pageComponents[pageId] || pageComponents.index;

console.log(pageId);

const App: FC = () => {
  useWindowListener();
  return (
    <>
      <PageComponent />
    </>
  );
};

export default App;
