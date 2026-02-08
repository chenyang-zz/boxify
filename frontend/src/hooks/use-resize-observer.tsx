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

import { useEffect, useRef, useState } from "react";

export interface Size {
  width: number;
  height: number;
}

export interface ResizeObserverOptions {
  delay?: number; // 节流延迟，单位毫秒，默认为 200ms
  onResize?: (size: Size) => void; // 大小变化时的回调函数
}

export function useResizeObserver<T extends HTMLElement>(
  options: ResizeObserverOptions = {},
) {
  const { delay = 200, onResize = () => {} } = options;
  const ref = useRef<T | null>(null);
  const [size, setSize] = useState<Size>({ width: 0, height: 0 });

  const lastExecRef = useRef(0);
  const trailingTimerRef = useRef<number | null>(null);

  const updateSize = (newSize: Size) => {
    const { width, height } = newSize;
    const prev = { ...size };

    if (prev.width === width && prev.height === height) {
      return;
    }
    setSize(newSize);
    onResize(newSize);
  };

  useEffect(() => {
    if (!ref.current) return;

    const element = ref.current;

    const observer = new ResizeObserver((entries) => {
      const { width, height } = entries[0].contentRect;
      const now = Date.now();

      const remaining = delay - (now - lastExecRef.current);

      if (remaining <= 0) {
        // 立即执行（leading）
        lastExecRef.current = now;
        updateSize({ width, height });
      } else if (!trailingTimerRef.current) {
        // trailing 执行
        trailingTimerRef.current = window.setTimeout(() => {
          lastExecRef.current = Date.now();
          trailingTimerRef.current = null;

          updateSize({ width, height });
        }, remaining);
      }
    });

    observer.observe(element);

    return () => {
      observer.disconnect();
      if (trailingTimerRef.current) {
        clearTimeout(trailingTimerRef.current);
      }
    };
  }, [delay]);

  return { ref, size };
}
