/**
 * 初始数据条目
 *
 * 用于在窗口间传递初始数据
 */
interface InitialDataEntry<T> {
  /** 目标窗口名称 */
  windowName: string;
  /** 源窗口名称 */
  source: string;
  /** 实际数据 */
  data: T;
  /** 创建时间戳（Unix 时间戳，秒） */
  timestamp: number;
  /** 过期时间戳（Unix 时间戳，秒） */
  expiresAt: number;
}

/**
 * 初始数据保存选项
 */
interface SaveInitialDataOptions<T> {
  /** 目标窗口名称 */
  targetWindow: string;
  /** 要传递的数据 */
  data: T;
  /** 可选：存活时间（分钟），默认30分钟 */
  ttl?: number;
}

interface SettingsInitialData {
  title: string;
}
