# Changelog

All notable changes to this project will be documented in this file.

## [v0.0.24] - 2026-03-09

### Changed
- CI: 使用 go env GOMODCACHE 获取正确的 Go 模块缓存路径
- CI: 添加 chmod 处理缓存目录权限问题
- CI: 使用 go clean -cache 清理构建缓存

---

## [v0.0.23] - 2026-03-09

### Changed
- CI: macOS DMG 打包前清理 node_modules、pnpm、Go、Homebrew、Xcode 缓存释放磁盘空间

---

## [v0.0.22] - 2026-03-09

### Fixed
- 修复 Windows 打包任务中 ARCH、CERT_PATH、PUBLISHER、USE_MSIX_TOOL 变量未正确传递到子任务的问题

---

## [v0.0.21] - 2026-03-09

### Changed
- CI: Linux 构建合并为 matrix 并行构建 (amd64/arm64)
- CI: Windows 构建合并为 matrix 并行构建 (amd64/arm64)
- Makefile: 优化 release-undo-version 版本查找逻辑

---
