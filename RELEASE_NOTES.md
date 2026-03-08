v0.0.23 - Linux 构建依赖优化

优化 GitHub Release 工作流中的 Linux 构建依赖配置。

Improvements
- 简化 Linux 依赖安装步骤名称
- 移除 pkg-config 和 gcc（Go 工具链已包含）
- 使用 libglib2.0-dev 替代 libsoup-3.0-dev

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.22...v0.0.23
