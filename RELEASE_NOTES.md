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

v0.0.23 - macOS 构建磁盘空间优化

在 macOS DMG 打包前清理缓存，解决磁盘空间不足问题。

Improvements
- CI: macOS DMG 打包前清理 node_modules、pnpm、Go、Homebrew、Xcode 缓存释放磁盘空间

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.22...v0.0.23
