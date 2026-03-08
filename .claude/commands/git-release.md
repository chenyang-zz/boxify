# GitHub Release 发布命令

发布 Boxify 新版本到 GitHub。

## 使用方式

```bash
/git-release [patch|minor|major|VERSION]
```

- 不带参数：自动递增修订号 (patch)
- `patch`：修订号 +1 (0.0.x)
- `minor`：次版本 +1 (x.0.0)
- `major`：主版本 +1 (x.0.0)
- 指定版本号：如 `0.1.0`，手动指定版本

## 执行步骤

1. **发布前检查**
   - 检查工作区是否干净（无未提交改动）
   - 检查是否已登录 GitHub CLI
   - 检查本地分支是否与远端同步

2. **版本管理**
   - 自动模式：递增版本号并更新 `frontend/package.json`
   - 手动模式：使用指定的版本号创建 tag

3. **触发发布**
   - 推送 tag 到 GitHub
   - 自动触发 GitHub Actions 构建与发布

4. **发布后验证**
   - 显示 workflow 运行状态
   - 提供查看发布详情的命令

## 发布产物

自动构建以下平台产物：
- macOS: `darwin-amd64`, `darwin-arm64`
- Linux: `linux-amd64`, `linux-arm64`
- Windows: `windows-amd64.exe`, `windows-amd64-setup.exe`

## 示例

```
/git-release           # 自动递增修订号发布
/git-release minor     # 次版本更新
/git-release 0.2.0     # 手动指定版本
```

---

**参数**: $ARGUMENTS

请执行以下发布流程：

1. 首先解析参数 `$ARGUMENTS`：
   - 空或 `patch` → 执行 `make release-auto-tag PART=patch`
   - `minor` → 执行 `make release-auto-tag PART=minor`
   - `major` → 执行 `make release-auto-tag PART=major`
   - 版本号格式 (如 `0.1.0`) → 执行 `make release-tag VERSION=0.1.0`

2. 执行发布前检查：
   - 运行 `git status --porcelain` 检查工作区
   - 运行 `gh auth status` 检查 GitHub CLI 登录状态
   - 如有问题，提示用户修复后重试

3. 执行发布命令

4. 发布完成后：
   - 显示 workflow 运行状态：`gh run list --limit 3`
   - 提示用户可通过 `gh run view` 查看详情
