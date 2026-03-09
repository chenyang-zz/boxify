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

v0.0.17 - CI 构建流程优化

重构 CI 构建流程，使用 matrix 策略并行构建多平台产物，并修复若干构建问题。

Improvements
- CI: Linux 构建合并为 matrix 并行构建 (amd64/arm64)
- CI: Windows 构建合并为 matrix 并行构建 (amd64/arm64)
- CI: macOS DMG 打包前清理缓存释放磁盘空间
- CI: 使用 go env GOMODCACHE 获取正确的 Go 模块缓存路径
- Makefile: 优化 release-undo-version 版本查找逻辑

Fixes
- 修复 Windows 打包任务中 ARCH、CERT_PATH 等变量未正确传递到子任务的问题
- 修复 macOS 构建中 Go 模块缓存权限问题

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.15...v0.0.17
