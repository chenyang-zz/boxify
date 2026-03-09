# Changelog

All notable changes to this project will be documented in this file.

## [v0.0.18] - 2026-03-09

### Changed
- Workflow: 新增自动生成 RELEASE_NOTES.md 步骤
- Makefile: 添加 git-release 兼容命令
- 重构 generate-release-notes.sh，增加 Release Info / Chores 分类
- 更新 README 和 release-guide 文档

---

## [v0.0.17] - 2026-03-09

### Added
- CI: Linux 构建使用 matrix 并行构建 (amd64/arm64)
- CI: Windows 构建使用 matrix 并行构建 (amd64/arm64)

### Changed
- CI: macOS DMG 打包前清理缓存释放磁盘空间
- CI: 使用 go env GOMODCACHE 获取正确的 Go 模块缓存路径
- Makefile: 优化 release-undo-version 版本查找逻辑

### Fixed
- 修复 Windows 打包任务中变量未正确传递到子任务的问题
- 修复 macOS 构建中 Go 模块缓存权限问题

---
