import { connection } from "@wails/models";
import { clsx, type ClassValue } from "clsx";
import { toast } from "sonner";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// 封装一个函数用于调用Wails后端函数，并统一处理错误
export async function callWails(
  fn: (...args: any) => Promise<connection.QueryResult>,
  ...args: any
) {
  return new Promise<connection.QueryResult>(async (resolve, reject) => {
    const res = await fn(...args);
    if (!res.success) {
      toast.error("发生错误", {
        description: res.message,
      });
      reject(new Error(res.message));
      return;
    }
    resolve(res);
  });
}
