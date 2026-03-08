v0.0.24 - 发布流程优化

本次更新优化了发布流程和 CI 构建配置。

Improvements
- 发布命令支持自动更新 CHANGELOG.md
- 新增 release-undo-version 命令用于版本回退
- 优化 Linux 构建依赖配置

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.23...v0.0.24
