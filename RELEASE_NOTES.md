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

v0.0.26 - CI 构建优化与版本撤销增强

优化跨平台 CI 构建流程，增强版本撤销命令功能。

Improvements
- CI: Windows 安装包查找改用 find_installer 函数，兼容命名差异
- CI: macOS 构建改为 matrix 并行构建 (amd64/arm64/universal)，提升构建效率
- Makefile: release-undo-version 支持批量删除高于目标版本的 Release/Tag
- Makefile: CHANGELOG 处理改为移除高于目标版本的所有条目

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.25...v0.0.26
