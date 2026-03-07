import { QueryResult } from "@wails/connection";
import { clsx, type ClassValue } from "clsx";
import { toast } from "sonner";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface BaseResult {
  success: boolean;
  message: string;
}

interface CallWailsOptions {
  timeoutMs?: number;
  timeoutMessage?: string;
}

const DEFAULT_WAILS_TIMEOUT_MS = 30000;

// 统一打印后端查询函数执行的 SQL，便于排查筛选与查询问题。
function logBackendQuerySQL(fnName: string, args: unknown[]) {
  const isQueryCall = fnName === "DBQuery" || fnName === "MySQLQuery" || fnName.endsWith("Query");
  if (!isQueryCall) {
    return;
  }
  const sql = args[2];
  if (typeof sql !== "string" || !sql.trim()) {
    return;
  }
  console.info(`[Wails:${fnName}] SQL => ${sql}`);
}

// 使用默认超时调用 Wails 后端函数，并统一处理错误。
export async function callWails<T extends BaseResult | null>(
  fn: (...args: any[]) => Promise<T>,
  ...args: Parameters<typeof fn>
) {
  return callWailsWithOptions(fn, args);
}

// 按选项调用 Wails 后端函数，允许按场景覆盖超时。
export async function callWailsWithOptions<T extends BaseResult | null>(
  fn: (...args: any[]) => Promise<T>,
  args: Parameters<typeof fn>,
  options: CallWailsOptions = {},
) {
  let timer: ReturnType<typeof setTimeout> | undefined;
  const timeoutMs = options.timeoutMs ?? DEFAULT_WAILS_TIMEOUT_MS;
  const timeoutMessage =
    options.timeoutMessage ?? `请求超时（>${Math.ceil(timeoutMs / 1000)}秒）`;
  logBackendQuerySQL(fn.name, args as unknown[]);

  try {
    const res = await Promise.race([
      fn(...args),
      new Promise<T>((_, reject) => {
        timer = setTimeout(() => reject(new Error(timeoutMessage)), timeoutMs);
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
      style: {
        textAlign: "left",
      },
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
    toast.success("复制成功", {
      style: {
        textAlign: "left",
      },
    });
  } catch (e) {
    toast.error("复制失败", {
      style: {
        textAlign: "left",
      },
    });
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
