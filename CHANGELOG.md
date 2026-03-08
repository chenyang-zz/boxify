# Changelog

All notable changes to this project will be documented in this file.

## [v0.0.26] - 2026-03-08

### Changed
- CI: 启用 CGO_ENABLED=1 支持 Linux/macOS 构建
- CI: 更新 Windows 构建依赖（GTK/WebKit/libsoup）
- Makefile: release-undo-version 自动删除远程 Release

---

## [v0.0.21] - 2026-03-08

### Changed
- release-undo-version 命令支持清空 RELEASE_NOTES.md
- release-undo-version 命令支持移除 CHANGELOG.md 对应条目
- CI 构建添加 CGO_ENABLED=0 环境变量
- 优化 CI 构建错误提示信息
- 修复 checksums 生成命令

---
