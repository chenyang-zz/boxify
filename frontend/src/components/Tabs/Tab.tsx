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

import { FC, useEffect, useRef, useState } from "react";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";
import { TabProps } from "./types";
import TabContextMenu from "./TabContextMenu";

const Tab: FC<TabProps> = ({ tab, isActive, onSelect, onClose }) => {
  const labelRef = useRef<HTMLSpanElement>(null);
  const [isLabelOverflow, setIsLabelOverflow] = useState(false);

  // 标签名溢出时启用左侧渐变，优先保留尾部关键信息可见。
  useEffect(() => {
    const element = labelRef.current;
    if (!element) return;

    const checkOverflow = () => {
      setIsLabelOverflow(element.scrollWidth > element.clientWidth);
    };

    checkOverflow();

    const resizeObserver = new ResizeObserver(checkOverflow);
    resizeObserver.observe(element);

    return () => resizeObserver.disconnect();
  }, [tab.label]);

  return (
    <TabContextMenu tab={tab}>
      <div
        style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
        className={cn(
          "group relative flex h-full select-none items-center",
          "border-l border-border/30 cursor-default flex-1 max-w-50 px-3 justify-center overflow-hidden",
          isActive
            ? "bg-background text-foreground"
            : "bg-transparent text-muted-foreground  hover:bg-muted hover:text-foreground",
        )}
        onClick={() => onSelect(tab.id)}
      >
        <span
          className={cn(
            " absolute left-3 w-4 h-full bg-linear-to-r ",
            isActive
              ? "from-background to-transparent"
              : "group-hover:from-muted from-card to-transparent",
          )}
        />
        <span
          ref={labelRef}
          className={cn(
            "overflow-hidden whitespace-nowrap text-xs",
            "max-w-full text-right ",
          )}
        >
          {tab.label}
        </span>
        <button
          type="button"
          aria-label={`关闭 ${tab.label}`}
          title="关闭标签"
          className={cn(
            "absolute right-2 top-1/2 -translate-y-1/2 inline-flex size-7 items-center justify-center rounded-md",
            "opacity-0 pointer-events-none transition-all duration-150",
            "text-muted-foreground group-hover:text-foreground",
            "group-hover:opacity-100 group-hover:pointer-events-auto group-hover:bg-muted/45",
          )}
          onClick={(event) => {
            event.stopPropagation();
            onClose(tab.id);
          }}
        >
          <X className="size-4" />
        </button>
      </div>
    </TabContextMenu>
  );
};

export default Tab;
