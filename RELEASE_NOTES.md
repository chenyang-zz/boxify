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

v0.0.21 - CI 构建流程重构

重构 CI 构建流程，使用 matrix 策略优化跨平台构建。

Improvements
- CI: Linux 构建合并为 matrix 并行构建 (amd64/arm64)
- CI: Windows 构建合并为 matrix 并行构建 (amd64/arm64)
- Makefile: 优化 release-undo-version 版本查找逻辑

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.20...v0.0.21
