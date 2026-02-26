import { ConnectionEnum } from "@/common/constrains";
import type { TagColorValue } from "@/constants/color-constants";

/**
 * 初始数据条目
 *
 * 用于在窗口间传递初始数据
 */
export interface InitialDataEntry<T> {
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
export interface SaveInitialDataOptions<T> {
  /** 目标窗口名称 */
  targetWindow: string;
  /** 要传递的数据 */
  data: T;
  /** 可选：存活时间（分钟），默认30分钟 */
  ttl?: number;
}

export interface SettingsInitialData {
  title: string;
}

export interface TerminalStandard {
  tagColor?: TagColorValue;
  name: string;
  shell?: string;
  workpath?: string;
  initialCommand?: string;
  remark?: string;
}

export interface CommonStandard {
  tagColor: TagColorValue | "";
  environment: string;
  name: string;
  host: string;
  user: string;
  port: number;
  authMethod: string;
  password?: string;
  remark?: string;
}

export interface ConnectionAdvanced {
  useSSH: boolean;
  sshName?: string;
  defaultDatabase?: string;
  timeout?: number;
  expiredAt?: number;
}

export interface ConnectionParameters {
  parameters: Record<string, string>;
}

export type ConnectionStandard = TerminalStandard | CommonStandard;

export interface ConnectionEditInitialData {
  uuid?: string;
  type?: ConnectionEnum;
  title?: string;
  standard?: ConnectionStandard;
  advanced?: ConnectionAdvanced;
  parameters?: ConnectionParameters;
}
