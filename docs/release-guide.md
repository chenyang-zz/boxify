# Boxify 发布流程

本文档记录 Boxify 的标准发布方式，默认使用 GitHub Actions workflow `Release On Tag` 自动构建与发布。

## 推荐方式：自动升版本并发布

```bash
make release-auto-tag
```

可选参数：

```bash
make release-auto-tag PART=patch  # 默认，修订号 +1
make release-auto-tag PART=minor  # 次版本 +1，修订号归零
make release-auto-tag PART=major  # 主版本 +1，次版本/修订号归零
```

执行内容：

1. 检查工作区是否干净（有未提交改动会中断）。
2. 对比远端已有 tag 与 `frontend/package.json` 版本，取较大值后递增。
3. 更新 `frontend/package.json` 版本并提交（`chore(release): bump version`）。
4. 推送到 `main`。
5. 创建并推送 `vX.Y.Z` tag。
6. 触发 GitHub Actions `Release On Tag` 自动构建并发布 Release。

## 手动方式：只打 tag 触发发布

```bash
make release-tag VERSION=0.0.15
```

适用于已手动维护版本号、仅需触发发布的场景。

## 发布前检查

1. 本地分支已同步：`git pull --rebase`。
2. 工作区干净：`git status` 无未提交改动。
3. 已登录 GitHub CLI：`gh auth status`。
4. 确认目标版本号未被占用：`git tag -l "v*"`。

## 发布后验证

1. 查看 workflow：`gh run list --limit 5`。
2. 查看本次 run 详情：`gh run view <run-id>`。
3. 检查 Release 资产是否齐全：`gh release view vX.Y.Z --json assets,url`。

## 常见问题

### 1) `工作区有未提交改动，请先提交后再发布`

先提交或暂存当前改动，再执行 `make release-auto-tag`。

### 2) `tag already exists`

说明目标版本已存在。优先使用 `make release-auto-tag` 让脚本自动选择下一个版本。

### 3) Workflow 构建失败

优先查看失败步骤日志：

```bash
gh run view <run-id> --log-failed
```

当前发布 workflow 已临时关闭 `linux-arm64` 产物构建，待交叉编译链路修复后再恢复。
