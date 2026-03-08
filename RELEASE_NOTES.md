<!--
 Copyright 2026 chenyang

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

v0.0.26 - 多平台发布流程重构

重构 GitHub Actions 发布流程，支持完整的跨平台构建与签名验证。

Improvements
- 重构 CI/CD workflow 支持并行构建（Linux/Windows/macOS）
- 新增 Linux amd64 的 deb/rpm 打包格式
- 新增 Windows arm64 安装包构建
- 新增 macOS 多架构 DMG（x64/arm64/universal）
- 添加发布产物完整性验证
- 添加 checksums GPG 签名
- 增强 release-undo-version 命令支持撤销远程版本

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.25...v0.0.26
