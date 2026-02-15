/**
 * 颜色标签配置
 *
 * 定义了 8 种状态指示颜色，用于连接标签的颜色选择
 */

export const TAG_COLORS = [
  {
    value: "red",
    label: "红色",
    bgClass: "bg-red-500",
    hoverClass: "hover:bg-red-600",
    ringClass: "focus-visible:ring-red-500/50",
    lightBg: "bg-red-50 dark:bg-red-950",
    textColor: "text-red-700 dark:text-red-300",
  },
  {
    value: "orange",
    label: "橙色",
    bgClass: "bg-orange-500",
    hoverClass: "hover:bg-orange-600",
    ringClass: "focus-visible:ring-orange-500/50",
    lightBg: "bg-orange-50 dark:bg-orange-950",
    textColor: "text-orange-700 dark:text-orange-300",
  },
  {
    value: "yellow",
    label: "黄色",
    bgClass: "bg-yellow-500",
    hoverClass: "hover:bg-yellow-600",
    ringClass: "focus-visible:ring-yellow-500/50",
    lightBg: "bg-yellow-50 dark:bg-yellow-950",
    textColor: "text-yellow-700 dark:text-yellow-300",
  },
  {
    value: "lime",
    label: "青柠",
    bgClass: "bg-lime-500",
    hoverClass: "hover:bg-lime-600",
    ringClass: "focus-visible:ring-lime-500/50",
    lightBg: "bg-lime-50 dark:bg-lime-950",
    textColor: "text-lime-700 dark:text-lime-300",
  },
  {
    value: "green",
    label: "绿色",
    bgClass: "bg-green-500",
    hoverClass: "hover:bg-green-600",
    ringClass: "focus-visible:ring-green-500/50",
    lightBg: "bg-green-50 dark:bg-green-950",
    textColor: "text-green-700 dark:text-green-300",
  },
  {
    value: "teal",
    label: "青色",
    bgClass: "bg-teal-500",
    hoverClass: "hover:bg-teal-600",
    ringClass: "focus-visible:ring-teal-500/50",
    lightBg: "bg-teal-50 dark:bg-teal-950",
    textColor: "text-teal-700 dark:text-teal-300",
  },
  {
    value: "blue",
    label: "蓝色",
    bgClass: "bg-blue-500",
    hoverClass: "hover:bg-blue-600",
    ringClass: "focus-visible:ring-blue-500/50",
    lightBg: "bg-blue-50 dark:bg-blue-950",
    textColor: "text-blue-700 dark:text-blue-300",
  },
  {
    value: "violet",
    label: "紫色",
    bgClass: "bg-violet-500",
    hoverClass: "hover:bg-violet-600",
    ringClass: "focus-visible:ring-violet-500/50",
    lightBg: "bg-violet-50 dark:bg-violet-950",
    textColor: "text-violet-700 dark:text-violet-300",
  },
] as const;

/**
 * 颜色值类型
 */
export type TagColorValue = (typeof TAG_COLORS)[number]["value"];
