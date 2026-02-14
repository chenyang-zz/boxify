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

import { FC } from "react";

interface WindowHeaderProps {
  title?: string;
}

const WindowHeader: FC<WindowHeaderProps> = ({ title }) => {
  return (
    <header
      className="w-full h-10 cursor-default shrink-0 flex items-center justify-center text-foreground text-xs"
      style={{ "--wails-draggable": "drag" } as React.CSSProperties}
    >
      {title}
    </header>
  );
};

export default WindowHeader;
