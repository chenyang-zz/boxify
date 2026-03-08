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

v0.0.26 - CI 构建修复与发布流程优化

修复跨平台构建配置问题，增强版本撤销命令功能。

Improvements
- CI: 启用 CGO_ENABLED=1 支持 Linux/macOS 构建
- CI: 更新 Windows 构建依赖（GTK/WebKit/libsoup）
- Makefile: release-undo-version 自动删除远程 Release

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.25...v0.0.26
