import { DEFAULT_ACCOUNT_ID } from "openclaw/plugin-sdk";

import { DEFAULT_LISTEN_URL } from "./constants.js";

/**
 * 规整账号 ID，保证默认账号语义稳定。
 */
export function normalizeAccountId(accountId) {
  const trimmed = String(accountId || "").trim();
  return trimmed || DEFAULT_ACCOUNT_ID;
}

/**
 * 读取字符串配置并提供默认值。
 */
function stringOrDefault(value, fallback) {
  const trimmed = String(value ?? "").trim();
  return trimmed || fallback;
}

/**
 * 规整监听地址，仅保留基础地址。
 */
function normalizeListenURL(listenURL) {
  const trimmed = String(listenURL || "").trim();
  if (!trimmed) {
    return DEFAULT_LISTEN_URL;
  }
  return trimmed.replace(/\/$/, "");
}

/**
 * 返回原生 channel 配置，并兼容单账号/多账号结构。
 */
export function resolveChannelConfig(cfg, accountId) {
  const root = cfg?.channels?.boxify && typeof cfg.channels.boxify === "object" ? cfg.channels.boxify : {};
  const accounts = root.accounts && typeof root.accounts === "object" ? root.accounts : {};
  const resolvedAccountId = normalizeAccountId(accountId);
  const accountOverrides =
    resolvedAccountId !== DEFAULT_ACCOUNT_ID &&
    accounts[resolvedAccountId] &&
    typeof accounts[resolvedAccountId] === "object"
      ? accounts[resolvedAccountId]
      : {};
  const merged = { ...root, ...accountOverrides };
  const rawListenURL = stringOrDefault(merged.listenUrl, DEFAULT_LISTEN_URL);

  return {
    accountId: resolvedAccountId,
    name: stringOrDefault(merged.name, resolvedAccountId === DEFAULT_ACCOUNT_ID ? "Boxify" : resolvedAccountId),
    enabled: merged.enabled !== false,
    listenUrl: normalizeListenURL(rawListenURL),
    sharedToken: stringOrDefault(merged.sharedToken, ""),
    defaultAgent: stringOrDefault(merged.defaultAgent, "main"),
    configured: Boolean(rawListenURL.trim()),
  };
}

/**
 * 列出当前配置可见的账号 ID。
 */
export function listAccountIds(cfg) {
  const root = cfg?.channels?.boxify && typeof cfg.channels.boxify === "object" ? cfg.channels.boxify : {};
  const ids = Object.keys(root.accounts && typeof root.accounts === "object" ? root.accounts : {});
  return ids.length > 0 ? ids : [DEFAULT_ACCOUNT_ID];
}
