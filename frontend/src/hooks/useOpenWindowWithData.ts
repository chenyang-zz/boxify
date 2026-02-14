import { useCallback } from "react";
import { InitialDataService, WindowService } from "@wails/service";
import { useInitialData } from "./useInitialData";
import { callWails, currentPageId } from "@/lib/utils";

/**
 * 保存初始数据（供源窗口使用）
 *
 * @param targetWindow - 目标窗口名称
 * @param data - 要传递的数据
 * @param ttl - 可选：存活时间（分钟），默认30分钟
 * @returns Promise<boolean> - 是否保存成功
 */
const saveInitialData = async <T>(
  targetWindow: string,
  data: T,
  ttl?: number,
): Promise<boolean> => {
  try {
    // 源窗口名称使用当前窗口
    const pageId = currentPageId();

    const result = await InitialDataService.SaveInitialData(
      pageId,
      targetWindow,
      data as any,
      ttl || 30,
    );

    if (result && result.success) {
      console.log("[初始数据] 已保存:", {
        source: pageId,
        targetWindow,
        data,
        expiresAt: result.data?.expiresAt,
      });
      return true;
    } else if (result) {
      console.error("[初始数据] 保存失败:", result.message);
      return false;
    }
    return false;
  } catch (error) {
    console.error("[初始数据] 保存失败:", error);
    return false;
  }
};

/**
 * 打开窗口并传递数据的 Hook
 *
 * 封装了"保存初始数据 + 打开窗口"的组合操作，简化使用
 *
 * @example
 * const { openWindowWithData } = useOpenWindowWithData();
 * await openWindowWithData("settings", { theme: "dark", userId: 123 });
 */
export const useOpenWindowWithData = () => {
  /**
   * 打开窗口并传递数据
   *
   * @param pageId - 页面ID（如 "settings", "connection-edit"）
   * @param data - 要传递的数据
   * @param options - 可选配置
   * @returns Promise<void>
   */
  const openWindowWithData = useCallback(
    async <T>(
      pageId: string,
      data: T,
      options?: {
        /** 存活时间（分钟），默认30分钟 */
        ttl?: number;
      },
    ) => {
      try {
        // 1. 获取目标窗口的窗口名称
        const windowNameResult = await callWails(
          WindowService.GetWindowNameByPageID,
          pageId,
        );

        if (!windowNameResult.success) {
          throw new Error(`获取窗口名称失败: ${windowNameResult.message}`);
        }

        const targetWindow = windowNameResult.data as string;

        // 2. 保存初始数据
        const saved = await saveInitialData(targetWindow, data, options?.ttl);

        if (!saved) {
          throw new Error("保存初始数据失败");
        }

        // 3. 打开窗口
        await callWails(WindowService.OpenPage, pageId);

        console.log("[窗口] 已打开并传递数据:", {
          pageId,
          targetWindow,
          data,
        });
      } catch (error) {
        console.error("[窗口] 打开失败:", error);
        throw error;
      }
    },
    [saveInitialData],
  );

  return { openWindowWithData };
};
