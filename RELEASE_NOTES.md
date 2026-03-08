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

v0.0.21 - 版本回退功能增强

本次更新增强了版本回退命令的功能，并优化了 CI 构建配置。

Improvements
- release-undo-version 命令现在支持自动清空 RELEASE_NOTES.md
- release-undo-version 命令现在支持自动移除 CHANGELOG.md 对应条目
- CI 构建添加 CGO_ENABLED=0 环境变量解决交叉编译问题
- 优化 CI 构建错误提示信息
- 修复 checksums 生成命令

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.20...v0.0.21
