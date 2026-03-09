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

v0.0.18 - 发布流程优化

优化发布流程，支持自动生成规范化发布说明。

Improvements
- Workflow: 新增自动生成 RELEASE_NOTES.md 步骤
- Makefile: 添加 git-release 兼容命令
- 脚本: 重构 generate-release-notes.sh，增加 Release Info / Chores 分类
- 文档: 更新 README 和 release-guide 发布说明

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v0.0.17...v0.0.18
