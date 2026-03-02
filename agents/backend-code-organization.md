# 后端代码组织规则（Git 模块）

更新时间：2026-03-02

本文档基于当前实现（`internal/service/git_service.go` 与 `internal/git/*`）沉淀后端代码组织规则（当前以 Git 模块为样例），作为后续扩展与重构的约束。

## 1. 分层边界

### 1.1 Service 层（`internal/service/git_service.go`）

职责仅限：

1. 暴露给前端的 Wails 服务方法；
2. 参数透传与返回结构封装（`internal/types` 中的 `*Result`）；
3. 事件转发（`events.EventTypeGitStatusChanged`）。

禁止：

1. 在 Service 层实现 Git 命令、路径解析、状态解析、监听策略等核心逻辑；
2. 在 Service 层直接操作仓库 map 或并发控制。

### 1.2 Core 层（`internal/git`）

职责集中：

1. Git 命令执行；
2. 仓库路径解析；
3. 状态解析与聚合；
4. 监听与去重；
5. 多仓库生命周期管理。

规则：所有 Git 业务能力必须先落在 `internal/git`，再由 Service 层暴露。

## 2. 包内模块拆分规则（`internal/git`）

保持“一文件一职责”：

1. `command_runner.go`：命令执行与超时控制；
2. `resolver.go`：路径归一化与仓库定位；
3. `parser.go`：`porcelain v2` 文本解析；
4. `collector.go`：状态采集与字段补全；
5. `watcher.go`：单仓监听、去重、回调派发；
6. `manager.go`：多仓管理与生命周期编排。

新增能力时优先判断归属，避免在单个文件堆叠多类职责。

## 3. 依赖方向规则

依赖必须单向，推荐链路：

`Service -> Manager -> Watcher/Collector -> Resolver/Parser/CommandRunner`

约束：

1. `manager.go` 可组合其他组件，但底层组件不能反向依赖 `Manager`；
2. `watcher.go` 不直接执行 `git`，通过 `StatusCollector` 获取状态；
3. `parser.go` 不依赖 I/O（命令执行、文件系统、事件系统）。

## 4. 类型与返回值规则

1. Git 对外数据结构统一放在 `internal/types/git.go`；
2. `internal/git` 返回领域对象与 `error`；
3. `internal/service` 负责将 `error` 转换为 `BaseResult` / `*Result`；
4. 错误信息保持中文，便于前端直接展示。

## 5. 并发与生命周期规则

1. 多仓状态（`repos`、`activeRepoKey`）由 `Manager` 用 `sync.RWMutex` 保护；
2. 单仓监听状态（`cancel`、`lastSnapshot`、`lastError`）由 `RepositoryWatcher` 用 `sync.Mutex` 保护；
3. 生命周期入口固定为：
   - 服务级：`ServiceStartup` / `ServiceShutdown`
   - 管理器级：`RegisterRepo` / `RemoveRepo` / `StartWatch` / `StopWatch` / `Shutdown`
4. 停止动作必须幂等（重复 Stop 不报错）。

## 6. 事件与状态同步规则

1. 状态变更事件统一由 `RepositoryWatcher` 触发；
2. 事件发送前必须做快照去重，避免重复推送；
3. 首次状态获取通过显式查询接口（如 `GetInitialStatusEvent`）补齐，而不是依赖监听时序。

## 7. 测试组织规则

1. 测试文件与实现文件同目录共置（`*_test.go`）；
2. 优先覆盖纯逻辑模块（`parser`、`resolver`、`manager`）；
3. 涉及 Git 仓库场景时复用 `test_helpers_test.go`，避免重复搭建测试仓库。

## 8. 扩展落地清单

新增 Git 功能时按以下顺序实现：

1. 在 `internal/types/git.go` 增加或调整领域模型；
2. 在 `internal/git` 对应职责文件实现核心逻辑；
3. 在 `internal/service/git_service.go` 增加前端可调用接口；
4. 添加/更新对应单元测试；
5. 若模块边界变化，同步更新本文件与 `docs/git-package-implementation.md`。
