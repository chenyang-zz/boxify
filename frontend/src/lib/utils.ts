import { QueryResult } from "@wails/connection";
import { clsx, type ClassValue } from "clsx";
import { toast } from "sonner";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// 封装一个函数用于调用Wails后端函数，并统一处理错误
export async function callWails<
  T extends (...args: any[]) => Promise<QueryResult | null>,
>(fn: T, ...args: Parameters<T>) {
  let timer: ReturnType<typeof setTimeout> | undefined;

  try {
    const res = await Promise.race([
      fn(...args),
      new Promise<QueryResult>((_, reject) => {
        timer = setTimeout(() => reject(new Error("请求超时")), 10000);
      }),
    ]);

    if (!res) {
      throw new Error("后端返回空结果");
    }

    if (!res.success) {
      throw new Error(res.message);
    }

    return res;
  } catch (e) {
    toast.error("发生错误", {
      description: (e as Error).message,
    });
    throw e;
  } finally {
    if (timer) {
      clearTimeout(timer);
    }
  }
}

export async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text);
    toast.success("复制成功");
  } catch (e) {
    toast.error("复制失败");
  }
}

// 获取当前页面ID（从meta标签中读取）
export function currentPageId() {
  return (
    document.querySelector('meta[name="page-id"]')?.getAttribute("content") ||
    "index"
  );
}

// 获取当前页面名称（从meta标签中读取）
export function currentWindowName() {
  return (
    document
      .querySelector('meta[name="window-name"]')
      ?.getAttribute("content") || "index"
  );
}
