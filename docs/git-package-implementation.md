# Git 包实现说明

本文档描述 `internal/git` 包在当前代码状态下的实现方式、调用链路和关键设计取舍，便于后续维护与重构。

## 1. 目标与职责

`internal/git` 负责 Git 状态能力的核心实现，主要职责如下：

- 路径解析：将任意输入路径解析为仓库根目录与 `.git` 目录。
- 状态采集：执行 `git status --porcelain=v2 --branch` 并解析结构化结果。
- 多仓管理：维护多个仓库的注册信息、激活状态与监听状态。
- 变更监听：基于 `fsnotify` + 轮询兜底，按变更事件推送状态。

`internal/service/git_service.go` 仅做服务层封装和事件转发，不承载核心 Git 逻辑。

## 2. 目录与模块划分

`internal/git` 包由以下模块组成：

- `command_runner.go`
- `resolver.go`
- `parser.go`
- `collector.go`
- `watcher.go`
- `manager.go`

建议按“从底到顶”的依赖关系理解：

1. `CommandRunner`：执行 Git 命令
2. `Resolver`：定位仓库
3. `StatusParser`：解析 porcelain v2
4. `StatusCollector`：采集并组装状态
5. `RepositoryWatcher`：监听并触发状态刷新
6. `Manager`：管理多仓与生命周期

## 3. 核心数据结构

核心对外模型定义在 `internal/types/git.go`：

- `GitRepoStatus`：仓库状态聚合（分支、ahead/behind、文件列表与计数）
- `GitFileStatus`：文件级状态（changed/renamed/untracked/unmerged）
- `GitStatusChangedEvent`：状态变化事件载体
- `GitRepoInfo`：管理器中的仓库元信息（监听状态、激活状态、最近错误等）

## 4. 状态采集流程

### 4.1 路径解析（Resolver）

`Resolver.Resolve(ctx, path)` 的流程：

1. `normalizePath` 规范化路径：
   - 空路径 -> `os.Getwd()`
   - 若传入文件路径 -> 转为所在目录
2. 执行 `git rev-parse --show-toplevel` 获取仓库根目录
3. 执行 `git rev-parse --absolute-git-dir` 获取绝对 `.git` 目录

输出 `RepoLocation{Path, RepoRoot, GitDir}`。

### 4.2 命令执行（CommandRunner）

`CommandRunner.Run` 使用 `exec.CommandContext` 执行 `git`，并带超时控制（默认 4 秒）。

- 成功返回 `CombinedOutput` 文本
- 失败时优先返回 Git 原始 stderr/stdout 文本，便于上层展示可读错误

### 4.3 状态解析（StatusParser）

解析输入为 `git status --porcelain=v2 --branch` 输出行，支持：

- `# branch.*`：分支元信息（`head/upstream/ahead/behind/oid`）
- `1 `：普通变更
- `2 `：重命名/复制变更
- `u `：冲突
- `? `：未跟踪

并计算：

- `StagedCount`
- `UnstagedCount`
- `UntrackedCount`
- `ConflictCount`

### 4.4 聚合（StatusCollector）

`CollectByRepoRoot(currentPath, repoRoot)`：

1. 运行 `git status --porcelain=v2 --branch`
2. 调用 `StatusParser` 解析
3. 回填 `RepositoryRoot`、`CurrentPath`、`UpdatedAt`
4. 计算 `IsClean`

`CollectByPath` 则先调用 `Resolver`，再转发到 `CollectByRepoRoot`。

## 5. 监听实现（RepositoryWatcher）

### 5.1 启动与循环

`RepositoryWatcher.Start` 会启动 `watchLoop`：

- 优先创建 `fsnotify.Watcher`
- 创建失败时降级为纯轮询 `pollingLoop`
- 启动后先发送一次 `startup` 触发的最新状态

循环中有两个 ticker：

- `debounceTicker`（180ms）：合并短时间内大量文件事件
- `fallbackTicker`（默认 2s，可配置）：周期性兜底采集

### 5.2 当前监听策略（最小监听集）

当前版本不再递归监听整个仓库目录树，而是监听以下最小集合：

- 工作区根目录：`repoRoot`
- `.git` 关键路径：
  - `HEAD`
  - `index`
  - `refs`
  - `logs`
  - `packed-refs`

设计目的：

- 大幅降低 watcher/FD 数量，避免大仓库监听成本过高
- 通过 `.git` 热点路径捕获绝大多数状态变化
- 通过 `fallbackTicker` 补齐未覆盖场景（例如部分工作区深层文件变化）

### 5.3 动态目录监听

当收到 `Create` 事件时，若创建目标是目录且不在 `.git` 下，则动态 `watcher.Add(新目录)`。

当前策略仅添加“新目录本身”，不递归整个子树，避免单次创建触发大量监听注册。

### 5.4 去重与事件发送

`emitLatest` 在发送前会构造快照摘要（`snapshot`），通过 `shouldEmit` 去重：

- 状态摘要未变化 -> 不发事件
- 状态摘要变化 -> 触发 `onStatus` 回调

## 6. 多仓管理（Manager）

`Manager` 负责仓库生命周期与并发安全（`sync.RWMutex`）：

- `RegisterRepo`：注册/覆盖仓库
- `RemoveRepo`：移除仓库并停止监听
- `StartWatch` / `StopWatch` / `StopAllWatches`
- `SetActiveRepo`：切换激活仓库，可选自动启动和停止其他仓库
- `GetStatus` / `Probe`：主动采集

监听间隔规则：

- 默认 2 秒
- `StartWatch(intervalMs)` 可指定
- 最小下限 800ms（防止过高频轮询）

## 7. Service 层对接

`GitService` 在 `ServiceStartup` 中创建 `Manager`，并将状态变化通过 Wails 事件系统向前端广播：

- 事件类型：`events.EventTypeGitStatusChanged`
- 事件负载：`GitStatusChangedEvent`

Service 层提供的前端接口主要是对 `Manager` 的一层转发：

- `RegisterRepo`
- `StartRepoWatch`
- `StopRepoWatch`
- `StopAllRepoWatch`
- `SetActiveRepo`
- `GetRepoStatus`
- `GetInitialStatusEvent`

## 8. 错误与并发策略

- 组件级错误（命令执行、解析、监听）不 panic，均向上返回或记录 `LastError`
- `RepositoryWatcher` 保存最近错误，`Manager.repoToInfo` 对外暴露
- `Manager` 使用 `RWMutex` 保护 `repos` map 和激活仓库字段
- `RepositoryWatcher` 内部用 `Mutex` 保护 `cancel/lastSnapshot/lastError`

## 9. 已知边界与后续优化点

- 最小监听集策略依赖 fallback 轮询补齐，极端场景下事件驱动实时性可能降低。
- `snapshot` 当前使用 SHA-1 对完整状态串做摘要；若后续状态字段继续增多，可考虑增量序列化优化。
- 目前每次采集都执行完整 `git status`，在超大仓库可评估缓存或分层采集策略。

