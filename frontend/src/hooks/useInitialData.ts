import { useEffect, useState } from "react";
import { Events } from "@wailsio/runtime";
import { InitialDataService } from "@wails/service";
import { callWails, currentWindowName } from "@/lib/utils";

/**
 * 初始数据传递 Hook
 *
 * 使用场景：
 * 1. 源窗口：保存数据并在打开新窗口前调用
 * 2. 目标窗口：在组件初始化时调用以获取数据
 *
 * @example
 * // 源窗口：保存数据
 * const { saveInitialData } = useInitialData();
 * await saveInitialData("settings", { theme: "dark" });
 * await WindowService.OpenPage("settings");
 *
 * @example
 * // 目标窗口：获取数据
 * const { initialData, isLoading } = useInitialData();
 * useEffect(() => {
 *   if (initialData && !isLoading) {
 *     console.log("收到初始数据:", initialData.data);
 *     // 更新UI状态
 *   }
 * }, [initialData, isLoading]);
 */
export const useInitialData = <T>() => {
  const [initialData, setInitialData] = useState<InitialDataEntry<T> | null>(
    null,
  );
  const [isLoading, setIsLoading] = useState(false);

  /**
   * 获取初始数据（供目标窗口使用）
   *
   * @param windowName - 窗口名称，如果不指定则使用当前窗口
   * @returns Promise<InitialDataEntry | null> - 初始数据或null
   */
  const getInitialData = async <T>(
    windowName?: string,
  ): Promise<InitialDataEntry<T> | null> => {
    setIsLoading(true);
    try {
      // 如果未指定窗口名称，使用当前窗口
      const targetWindow = windowName || currentWindowName();

      const result = await callWails(
        InitialDataService.GetInitialData,
        targetWindow,
      );

      if (result.success && result.data) {
        const data = result.data as InitialDataEntry<T>;
        setInitialData(data as any);
        console.log("[初始数据] 已加载:", data);
        return data;
      } else {
        console.log("[初始数据] 无数据或已过期:", result.message);
        return null;
      }
    } catch (error) {
      console.error("[初始数据] 获取失败:", error);
      return null;
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * 清除初始数据
   *
   * @param windowName - 窗口名称，如果不指定则使用当前窗口
   */
  const clearInitialData = async (windowName?: string): Promise<void> => {
    try {
      const targetWindow = windowName || currentWindowName();

      await callWails(InitialDataService.ClearInitialData, targetWindow);
      setInitialData(null);
      console.log("[初始数据] 已清除:", targetWindow);
    } catch (error) {
      console.error("[初始数据] 清除失败:", error);
    }
  };

  // 第一次加载时尝试获取初始数据
  useEffect(() => {
    getInitialData();
  }, []);

  // 监听初始数据接收事件（用于实时响应）
  useEffect(() => {
    const unbind = Events.On(
      "initial-data:received",
      (event: { data: InitialDataEntry<T> }) => {
        if (event.data.windowName === currentWindowName()) {
          console.log("[初始数据] 接收事件:", event.data);
          setInitialData(event.data);
        }
      },
    );

    return () => unbind();
  }, []);

  return {
    initialData,
    isLoading,
    getInitialData,
    clearInitialData,
  };
};
