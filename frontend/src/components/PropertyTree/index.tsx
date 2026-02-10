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

import { FC } from "react";
import TreeHeader from "./TreeHeader";
import { useResizeObserver } from "@/hooks/use-resize-observer";
import { useAppStore } from "@/store/app.store";
import FileTree from "../FileTree";

const MIN_WIDTH = 100;

const PropertyTree: FC = () => {
  const { ref } = useResizeObserver<HTMLDivElement>({
    onResize: ({ width }) => {
      if (width < MIN_WIDTH) {
        useAppStore.getState().setIsPropertyOpen(false);
      }
    },
    delay: 0,
  });

  return (
    <>
      <div ref={ref} className="h-0"></div>
      <div className="h-full w-full rounded-lg bg-card p-0.5 flex flex-col">
        <TreeHeader />
        <FileTree />
      </div>
    </>
  );
};

export default PropertyTree;
