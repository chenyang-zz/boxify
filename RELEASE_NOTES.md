v0.0.22 - 发布工作流重构

本次更新重构了 GitHub Release 发布工作流，新增 tag 触发的独立 CI 配置。

Improvements
- 新增 release-on-tag.yml 工作流，支持 tag 推送时自动构建和发布
- 支持多平台构建: darwin-amd64/arm64, linux-amd64/arm64, windows-amd64
- 更新 git-release 命令支持版本类型询问 (patch/minor/major)
- 添加不同版本类型的发布说明格式模板

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.21...v0.0.22
