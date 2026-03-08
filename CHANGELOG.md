# Changelog

All notable changes to this project will be documented in this file.

## [v0.0.26] - 2026-03-08

### Changed
- CI: Windows 安装包查找改用 find_installer 函数兼容命名差异
- CI: macOS 构建改为 matrix 并行构建 (amd64/arm64/universal)
- Makefile: release-undo-version 支持批量删除高于目标版本的 Release/Tag
- Makefile: CHANGELOG 处理改为移除高于目标版本的所有条目

---

## [v0.0.21] - 2026-03-08

### Changed
- release-undo-version 命令支持清空 RELEASE_NOTES.md
- release-undo-version 命令支持移除 CHANGELOG.md 对应条目
- CI 构建添加 CGO_ENABLED=0 环境变量
- 优化 CI 构建错误提示信息
- 修复 checksums 生成命令

---
