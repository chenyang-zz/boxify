# Changelog

All notable changes to this project will be documented in this file.

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
