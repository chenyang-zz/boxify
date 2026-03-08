# GitHub Release 发布命令

发布 Boxify 新版本到 GitHub。

## 使用方式

```bash
/git-release [patch|minor|major]
```

- 不带参数：会询问用户版本更新类型
- `patch`：修订号 +1 (0.0.x) - bug 修复、小优化
- `minor`：次版本 +1 (0.x.0) - 新功能、功能增强
- `major`：主版本 +1 (x.0.0) - 重大更新、破坏性变更

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

4. **触发发布**
   - 使用 `make release-auto-tag PART=patch|minor|major` 自动递增版本并推送
   - 自动触发 GitHub Actions 构建与发布

5. **发布后验证**
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

## 发布产物

自动构建以下平台产物：
- macOS: `darwin-amd64`, `darwin-arm64`
- Linux: `linux-amd64`, `linux-arm64`
- Windows: `windows-amd64.exe`, `windows-amd64-setup.exe`

## 示例

```
/git-release           # 询问版本类型后发布
/git-release patch     # 修订号更新
/git-release minor     # 次版本更新
/git-release major     # 主版本更新
```

---

**参数**: $ARGUMENTS

请执行以下发布流程：

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

4. **提交发布说明（如有修改）**：
   - 如果 RELEASE_NOTES.md 有修改，执行 `/git-push` 提交

5. **执行发布**：
   - 执行 `make release-auto-tag PART=patch|minor|major`（使用步骤2确定的 PART 值）
   - 该命令会自动：递增版本号 → 提交版本变更 → 推送 → 创建并推送 tag

6. **发布完成后**：
   - 显示 workflow 运行状态：`gh run list --limit 3`
   - 提示用户可通过 `gh run view` 查看详情
