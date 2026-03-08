# GitHub Release 发布命令

发布 Boxify 新版本到 GitHub。

## 使用方式

```bash
/git-release [patch|minor|major]
```

- 不带参数或 `patch`：修订号 +1 (0.0.x)
- `minor`：次版本 +1 (0.x.0)
- `major`：主版本 +1 (x.0.0)

## 执行步骤

1. **自动提交未提交内容**
   - 检查工作区是否有未提交的改动
   - 如果有未提交内容，自动执行 `/git-push` 命令提交
   - 确保工作区干净后再继续发布流程

2. **编写发布说明**
   - 查看本次更新内容，在根目录编辑 `RELEASE_NOTES.md`
   - 等待用户确认后再继续发布

3. **触发发布**
   - 使用 `make release-auto-tag` 自动递增版本并推送
   - 自动触发 GitHub Actions 构建与发布

4. **发布后验证**
   - 显示 workflow 运行状态

## 发布说明格式

RELEASE_NOTES.md 应遵循以下格式：

```markdown
vX.Y.Z - 发布主题/名称

简短介绍本次更新的核心内容。

Highlights
- 新增功能1
- 新增功能2

Stability
- 修复问题1
- 优化项1

Verification
- Passed pnpm run build
- Passed make dev
- Passed make build

Full Changelog: v旧版本...v新版本
```

试例

```markdown
v5.2.1 - OpenClaw 正式接入

这次更新，Boxify 正式加入 Workflow Center 1.0。

现在它不只是 OpenClaw 的管理面板，也开始具备“复杂任务自动接管 -> 分步骤推进 -> 产出真实文件 -> 自动回传用户”的完整工作流能力。

Highlights
- 新增完整工作流中心：设置、模板管理、AI 生成模板、运行列表、运行详情、步骤进度、事件流、删除实例
- 支持复杂任务自动转工作流：即时确认、进度回写、暂停、恢复、重试、审批、取消、多实例并发
- 工作流关键步骤支持自动落文件到 OpenClaw 工作区
- QQ 私聊 / 群聊支持工作流完成后自动回传最终文件
- 工作流详情页新增文件区：预览、下载、回传状态、单文件重发、批量重发最终文件
- 节点支持 skill，可加载对应 SKILL.md 参与执行
- 活动日志新增 Workflow / QQ / 飞书 / 企业微信 / OpenClaw 多来源消息流

Stability
- 修复复杂任务重复创建工作流实例
- 修复等待用户回复后仍无法继续执行的问题
- 修复审批继续后重新阻塞的问题
- 修复工作流模型协议兼容问题，支持 openai-responses
- 修复 AI 模板生成时 outputFile=true/false 导致的异常文件名问题
- 优化工作流上下文裁剪、超时后精简上下文重试与大步骤输出约束，降低中后段 504 超时概率

Verification
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
/git-release           # 自动递增修订号发布
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

2. **编写发布说明**：
   - 运行 `git log $(git describe --tags --abbrev=0)..HEAD --oneline` 查看本次更新内容
   - 在根目录编辑 `RELEASE_NOTES.md`，按上述格式编写发布说明
   - **暂停并等待用户确认**：询问用户发布说明是否满意，是否需要修改
   - 用户确认后继续

3. **提交发布说明（如有修改）**：
   - 如果 RELEASE_NOTES.md 有修改，执行 `/git-push` 提交

4. **执行发布**：
   - 解析参数 `$ARGUMENTS` 确定 PART 值（默认 patch）
   - 执行 `make release-auto-tag PART=patch|minor|major`
   - 该命令会自动：递增版本号 → 提交版本变更 → 推送 → 创建并推送 tag

5. **发布完成后**：
   - 显示 workflow 运行状态：`gh run list --limit 3`
   - 提示用户可通过 `gh run view` 查看详情
