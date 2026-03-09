# GitHub Release 发布命令

发布 Boxify 新版本到 GitHub，或撤销最近一次版本。

## 使用方式

```bash
/git-release [patch|minor|major|undo]
```

- 不带参数：会询问用户版本更新类型
- `patch`：修订号 +1 (0.0.x) - bug 修复、小优化
- `minor`：次版本 +1 (0.x.0) - 新功能、功能增强
- `major`：主版本 +1 (x.0.0) - 重大更新、破坏性变更
- `undo`：撤销最近一次版本号修改（回退版本并清空发布说明）

## 执行步骤

1. **自动提交未提交内容**
   - 检查工作区是否有未提交的改动
   - 如果有未提交内容，自动执行 `/git-push` 命令提交
   - 确保工作区干净后再继续发布流程

2. **确定版本类型**
   - 如果 `$ARGUMENTS` 为空，使用 AskUserQuestion 询问用户：
     - patch: 修复 bug 或小优化
     - minor: 新增功能或功能增强
     - major: 重大更新或破坏性变更
   - 根据用户选择或参数确定 PART 值

3. **编写发布说明**
   - 查看本次更新内容，在根目录编辑 `RELEASE_NOTES.md`
   - 根据 PART 类型选择合适的发布说明格式
   - 编写完成后自动继续发布流程

4. **更新 CHANGELOG.md**
   - 在 `CHANGELOG.md` 文件顶部添加新版本的更新日志
   - 格式参考下方「CHANGELOG 格式」章节
   - 根据版本类型选择合适的条目分类

5. **提交变更**
   - 如果 RELEASE_NOTES.md 或 CHANGELOG.md 有修改，执行 `/git-push` 提交

6. **触发发布**
   - 使用 `make release-auto-tag PART=patch|minor|major` 自动递增版本并推送
   - 自动触发 GitHub Actions 构建与发布

7. **发布后验证**
   - 显示 workflow 运行状态

## 发布说明格式

根据版本类型选择不同格式：

### patch 版本格式

```markdown
vX.Y.Z - 发布主题

简短描述本次修复/优化的内容。

Fixes
- 修复问题1
- 修复问题2

Improvements
- 优化项1

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v旧版本...v新版本
```

### minor 版本格式

```markdown
vX.Y.Z - 发布主题/名称

简短介绍本次更新的核心内容。

Highlights
- 新增功能1
- 新增功能2
- 功能增强1

Fixes
- 修复问题1

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v旧版本...v新版本
```

### major 版本格式

```markdown
vX.Y.0 - 发布主题/名称

## 概述

详细介绍本次重大更新的背景和目标。

## Breaking Changes

- 破坏性变更1及迁移指南
- 破坏性变更2及迁移指南

## New Features

- 重要新功能1
- 重要新功能2

## Improvements

- 优化项1
- 优化项2

## Fixes

- 修复问题1

## Upgrade Guide

升级步骤和注意事项。

## Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v旧版本...v新版本
```

## CHANGELOG 格式

CHANGELOG.md 记录所有版本的历史变更，新版本信息添加到文件顶部。

### 基本格式

```markdown
# Changelog

All notable changes to this project will be documented in this file.

## [vX.Y.Z] - YYYY-MM-DD

### Added
- 新增功能描述

### Changed
- 变更/优化描述

### Fixed
- 修复问题描述

### Removed
- 移除功能描述（如有）

---

## [v0.0.1] - 2024-01-01
...
```

### patch 版本 CHANGELOG 示例

```markdown
## [v0.0.24] - 2024-03-08

### Fixed
- 修复登录页面在移动端的显示问题
- 修复文件上传时的内存泄漏

### Changed
- 优化首页加载速度
```

### minor 版本 CHANGELOG 示例

```markdown
## [v0.1.0] - 2024-03-08

### Added
- 新增用户权限管理模块
- 新增数据导出功能（支持 CSV/JSON）
- 新增深色模式支持

### Changed
- 重构 API 响应结构，统一错误处理

### Fixed
- 修复分页组件在数据为空时的显示问题
```

### major 版本 CHANGELOG 示例

```markdown
## [v1.0.0] - 2024-03-08

### Added
- 全新的插件系统
- 支持多语言（i18n）
- 完整的 CLI 工具链

### Changed
- **BREAKING**: API 接口全面升级，不再兼容 v0.x
- **BREAKING**: 配置文件格式变更，需迁移

### Fixed
- 修复所有已知的稳定性问题

### Removed
- 移除已废弃的旧版 API

### Migration Guide
- 配置文件迁移：运行 `boxify migrate config`
- API 升级：参考 [升级指南](./docs/upgrade-v1.md)
```

## 发布产物

自动构建并发布以下产物（`VERSION` 为 `vX.Y.Z`）：
- macOS: `boxify-${VERSION}-darwin-amd64`、`boxify-${VERSION}-darwin-arm64`
- Linux: `boxify-${VERSION}-linux-amd64`、`boxify-${VERSION}-linux-arm64`
- Windows: `boxify-${VERSION}-windows-amd64.exe`、`boxify-${VERSION}-windows-amd64-setup.exe`
- 校验文件: `checksums.txt`

## 示例

```
/git-release           # 询问版本类型后发布
/git-release patch     # 修订号更新
/git-release minor     # 次版本更新
/git-release major     # 主版本更新
/git-release undo      # 撤销最近一次版本
```

---

## 撤销版本

当需要撤销最近一次版本发布时：

```bash
/git-release undo
```

撤销操作会：
1. 回退 `frontend/package.json` 的版本号到上一个版本
2. 清空 `RELEASE_NOTES.md` 文件内容
3. 撤回 `CHANGELOG.md` 文件中的上一次更改记录
4. 删除被撤销版本对应的远程 Release（若存在）
5. 删除被撤销版本对应的远程 Tag（若存在）
6. 删除被撤销版本对应的本地 Tag（若存在）
7. 提示用户检查后手动提交变更

---

**参数**: $ARGUMENTS

请根据参数执行对应流程：

### 如果参数为 `undo`（撤销版本）

1. **执行撤销命令**：
   - 运行 `make release-undo-version`
   - 该命令会自动回退版本号、清空 RELEASE_NOTES.md、移除 CHANGELOG 对应条目，并清理被撤销版本的 Release/Tag 信息

2. **显示结果**：
   - 告知用户版本回退结果
   - 提示用户检查变更后手动提交

### 否则（发布新版本）

1. **检查并提交未提交内容**：
   - 运行 `git status --porcelain` 检查工作区状态
   - 如果有未提交内容，执行 `/git-push` 命令自动提交
   - 等待提交完成后再继续

2. **确定版本类型**：
   - 检查 `$ARGUMENTS` 是否有值
   - 如果 `$ARGUMENTS` 为空，使用 AskUserQuestion 询问用户版本更新类型
   - 将确定的版本类型赋值给 PART（patch/minor/major）

3. **编写发布说明**：
   - 运行 `git log $(git describe --tags --abbrev=0)..HEAD --oneline` 查看本次更新内容
   - 在根目录编辑 `RELEASE_NOTES.md`
   - 根据 PART 类型（patch/minor/major）选择对应的发布说明格式
   - 编写完成后直接继续，无需等待用户确认

4. **更新 CHANGELOG.md**：
   - 读取当前 `CHANGELOG.md` 文件内容
   - 根据本次更新内容，按照「CHANGELOG 格式」章节编写新版本的变更日志
   - 将新版本条目插入到文件顶部（在 `# Changelog` 标题和分隔线之后）
   - 日期使用当前日期（格式：YYYY-MM-DD）

5. **提交变更**：
   - 如果 RELEASE_NOTES.md 或 CHANGELOG.md 有修改，执行 `/git-push` 提交

6. **执行发布**：
   - 执行 `make release-auto-tag PART=patch|minor|major`（使用步骤2确定的 PART 值）
   - 该命令会自动：递增版本号 → 提交版本变更 → 推送 → 创建并推送 tag

7. **发布完成后**：
   - 显示 workflow 运行状态：`gh run list --limit 3`
   - 提示用户可通过 `gh run view` 查看详情
