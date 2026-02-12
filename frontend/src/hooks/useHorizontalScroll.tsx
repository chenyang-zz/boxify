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

import { useEffect, useRef } from "react";

interface UseHorizontalScrollOptions {
  hideScrollbar?: boolean;
  scrollSpeed?: number;
  preventDefault?: boolean;
}

export function useHorizontalScroll<T extends HTMLElement = HTMLDivElement>(
  options: UseHorizontalScrollOptions = {},
) {
  const {
    hideScrollbar = false,
    scrollSpeed = 1,
    preventDefault = true,
  } = options;

  const ref = useRef<T>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    // 横向滚轮适配
    const onWheel = (e: WheelEvent) => {
      if (e.deltaY === 0) return;

      if (preventDefault) {
        e.preventDefault();
      }

      el.scrollLeft += e.deltaY * scrollSpeed;
    };

    el.addEventListener("wheel", onWheel, { passive: false });

    // 隐藏滚动条
    if (hideScrollbar) {
      el.style.scrollbarWidth = "none"; // Firefox
      // @ts-ignore
      el.style.msOverflowStyle = "none"; // IE

      const style = document.createElement("style");
      style.innerHTML = `
        .__hide-scrollbar::-webkit-scrollbar {
          display: none;
        }
      `;
      document.head.appendChild(style);

      el.classList.add("__hide-scrollbar");
    }

    return () => {
      el.removeEventListener("wheel", onWheel);
    };
  }, [hideScrollbar, scrollSpeed, preventDefault]);

  return ref;
}
