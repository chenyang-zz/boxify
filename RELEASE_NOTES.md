v0.0.21 - 发布流程优化

本次更新主要优化了 GitHub Release 发布流程和 CI 工作流配置。

Highlights
- 新增 RELEASE_NOTES.md 支持自定义发布说明
- 简化 git-release 命令，移除手动指定版本号功能
- 优化 CI 工作流配置，增加完整 git 历史获取

Stability
- 为 bindings 生成任务添加 CGO_ENABLED=0 环境变量，避免 CI 环境缺少依赖时的告警
- 格式化 CI 工作流 YAML 配置，提升可读性

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.20...v0.0.21
