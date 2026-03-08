# Changelog

All notable changes to this project will be documented in this file.

## [v0.0.26] - 2026-03-08

### Added
- CI/CD 新增 Linux amd64 的 deb/rpm 打包
- CI/CD 新增 Windows arm64 安装包构建
- CI/CD 新增 macOS 多架构 DMG（x64/arm64/universal）
- CI/CD 添加发布产物完整性验证
- CI/CD 添加 checksums GPG 签名

### Changed
- 重构 GitHub Actions workflow 支持并行构建
- 增强 release-undo-version 命令支持撤销远程版本

---

## [v0.0.21] - 2026-03-08

### Changed
- release-undo-version 命令支持清空 RELEASE_NOTES.md
- release-undo-version 命令支持移除 CHANGELOG.md 对应条目
- CI 构建添加 CGO_ENABLED=0 环境变量
- 优化 CI 构建错误提示信息
- 修复 checksums 生成命令

---
