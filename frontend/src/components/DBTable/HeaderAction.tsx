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
import { Button } from "../ui/button";
import {
  ArrowUpNarrowWideIcon,
  FunnelPlusIcon,
  MinusIcon,
  PlusIcon,
  RotateCcwIcon,
  SaveIcon,
  TimerResetIcon,
  Undo2Icon,
} from "lucide-react";
import { Separator } from "../ui/separator";

const HeaderAction: FC = () => {
  return (
    <header className="shrink-0 px-2 py-1 flex items-center gap-0.5 shadow shadow-background">
      <Button variant="ghost" size="sm">
        <TimerResetIcon className="size-4" />
        开始事务
      </Button>
      <Separator orientation="vertical" />
      <Button className=" rounded-full" size="icon-sm" variant="ghost">
        <RotateCcwIcon />
      </Button>
      <Button className=" rounded-full" size="icon-sm" variant="ghost">
        <PlusIcon />
      </Button>
      <Button className=" rounded-full" size="icon-sm" variant="ghost">
        <MinusIcon />
      </Button>
      <Button className=" rounded-full" size="icon-sm" variant="ghost">
        <SaveIcon />
      </Button>
      <Button className=" rounded-full" size="icon-sm" variant="ghost">
        <Undo2Icon />
      </Button>
      <Separator orientation="vertical" />
      <Button className=" rounded-full" size="icon-sm" variant="ghost">
        <Undo2Icon />
      </Button>
      <Button className=" rounded-full" size="icon-sm" variant="ghost">
        <FunnelPlusIcon />
      </Button>
      <Button className=" rounded-full" size="icon-sm" variant="ghost">
        <ArrowUpNarrowWideIcon />
      </Button>
      <Separator orientation="vertical" />
      <Button className=" rounded-full" size="icon-sm" variant="ghost">
        <Undo2Icon />
      </Button>
      <Button className=" rounded-full" size="icon-sm" variant="ghost">
        <Undo2Icon />
      </Button>
    </header>
  );
};

export default HeaderAction;
