# Boxify

基于 Wails v3 的跨平台数据库管理桌面应用，提供 MySQL / PostgreSQL 可视化管理能力，并支持 SSH 隧道、文件管理与内置终端等开发辅助功能。

![Boxify](./boxify.png)

## 功能特性

- 多数据库支持：MySQL、PostgreSQL
- SSH 隧道连接数据库
- 数据查询、编辑与批量修改
- 导入/导出：CSV、JSON、Markdown
- 表结构管理：字段、索引、外键、触发器
- 内置文件树与终端能力

## 技术栈

- 后端：Go、Wails v3、`go-sql-driver/mysql`、`lib/pq`、`golang.org/x/crypto/ssh`
- 前端：React 19、TypeScript、Vite、Tailwind CSS、shadcn/ui
- 包管理：Go Modules、pnpm

## 系统架构

Boxify 采用 Wails 的前后端一体化桌面架构：

1. 后端（Go）负责数据库、终端、窗口、文件系统、Git、OpenClaw 等系统能力。
2. 前端（React）负责 UI 渲染、交互编排和状态管理。
3. 桥接层（Wails bindings）负责类型安全的 RPC 调用与事件通信。

核心入口：

- 应用启动：`main.go`
- 服务装配：`internal/service/*`
- 窗口管理：`internal/window/*`
- 前端入口：`frontend/src/main.tsx`、`frontend/src/App.tsx`

## 核心模块说明

### 后端模块

- `internal/service/database_service.go`：数据库服务编排入口，向前端暴露查询/结构/导入导出等能力。
- `internal/db/connection_manager.go`：连接缓存、Ping 探活、失效重建与关闭清理。
- `internal/service/terminal_service.go`：PTY 会话管理、命令写入、输出事件、会话生命周期。
- `internal/service/window_service.go` + `internal/window/*`：多窗口注册、打开/关闭、生命周期事件。
- `internal/service/git_service.go`：Git 仓库注册、状态查询、监听与状态事件推送。
- `internal/service/filesystem_service.go`：目录读取与路径展开能力。
- `internal/service/claw_service.go`：OpenClaw 进程/插件/任务/更新/监控能力聚合。

### 前端模块

- `frontend/src/pages/*`：页面级入口（main/settings/connection-edit）。
- `frontend/src/components/*`：业务组件（DBTable、Terminal、FileTree、ClawContent 等）。
- `frontend/src/lib/utils.ts`：`callWails` 统一调用封装（超时、错误提示、日志）。
- `frontend/src/store/event.store.ts`：Wails 事件订阅与状态缓存。
- `frontend/src/store/data-sync.store.ts`：窗口间广播/定向同步通道。

## 关键流程

### 1. 应用启动流程

1. `main.go` 注册窗口/数据同步/终端/Git 等事件类型。
2. 初始化 `AppManager`，加载页面配置并创建主窗口。
3. 注入 `ServiceDeps` 并注册各业务 Service。
4. 前端 `App.tsx` 根据页面 ID 加载页面并初始化全局事件订阅。

### 2. 数据库查询流程

1. 前端调用 `@wails/service` 的数据库接口。
2. 后端 `DatabaseService` 通过 `ConnectionManager` 获取可用连接。
3. 缓存连接按策略探活（Ping），失效则自动重建。
4. 执行 SQL 并返回统一 `QueryResult`（成功标志 + 消息 + 数据）。

### 3. 终端执行流程

1. 前端创建终端会话并传入 shell/workdir 配置。
2. 后端创建 PTY 进程并启动输出读取循环。
3. 前端写入命令（支持 blockID 关联）。
4. 后端通过事件实时推送输出，前端按 block 聚合渲染。

### 4. 多窗口与数据同步

1. `WindowService` 打开/关闭页面窗口。
2. `WindowRegistry` 管理窗口生命周期（主窗口/单例窗口通常隐藏而非销毁）。
3. 通过 `data-sync:broadcast` 与 `data-sync:targeted` 进行窗口间数据同步。

## 环境要求

- Go `1.25+`（当前 `go.mod` 为 `1.25`）
- Node.js `20+`（建议 LTS）
- pnpm `9+`
- Wails CLI：`wails3`

安装 Wails CLI（按需）：

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

## 快速开始

```bash
# 1) 初始化依赖（Go + 前端）
make init

# 2) 启动开发模式（前后端热重载）
make dev
```

## 常用命令

```bash
# 开发与构建
make dev                         # 启动开发模式
make build                       # 构建生产版本
make frontend-dev                # 仅启动前端开发服务器
make frontend-build              # 仅构建前端

# macOS 打包
make refresh-icons               # 根据 build/appicon.png 生成图标
make build-macos-app             # 生成图标并打包 macOS .app
make build-macos-app-universal   # 生成图标并打包 Universal .app
make run-macos-app               # 运行已打包的 .app

# 依赖与检查
make install                     # 安装全部依赖
make check                       # 检查 Go/前端/Git 状态
make tidy                        # 整理 Go 依赖
make clean                       # 清理构建产物
make clean-cache                 # 清理缓存

# 质量保障
make test                        # 运行 Go 测试
make test-coverage               # 生成覆盖率报告（coverage.html）
make format                      # 格式化代码
make lint                        # 代码检查

# 发布相关
make release-tag VERSION=0.0.19              # 创建并推送指定版本 tag
make release-auto-tag PART=patch             # 自动递增版本并发布（patch/minor/major）
make git-release PART=minor                  # 兼容命令（等价于 release-auto-tag）
make release-undo-version                    # 回退最近一次版本号与发布信息
```

## 前端独立开发

```bash
cd frontend
pnpm install
pnpm run dev
pnpm run build
```

## 文档索引

- 项目架构索引：`agents/project-architecture.md`
- 后端架构：`agents/backend-architecture.md`
- 前端架构：`agents/frontend-architecture.md`
- 后端代码组织规范：`agents/backend-code-organization.md`
- Git 提交规范：`agents/git-commit-convention.md`
- 发布流程：`docs/release-guide.md`
- 终端组件架构：`docs/terminal-component-architecture.md`

## 目录概览

```text
Boxify/
├── main.go                 # 应用入口
├── internal/               # 后端核心代码
├── frontend/               # React + TypeScript 前端
├── agents/                 # 架构与协作规范文档
├── docs/                   # 设计与流程文档
├── scripts/                # 构建与工具脚本
├── Makefile                # 开发/构建/发布命令
└── README.md
```

## 开发注意事项

- MySQL 标识符使用反引号 `` ` ``，PostgreSQL 使用双引号 `"`。
- 前后端通信优先复用 `frontend/bindings` 生成类型，避免同义类型漂移。
- 后端已定义返回结构时，前端不要新增同义接口。
- 错误信息统一使用中文。
- 连接使用前需做有效性检查（Ping），应用退出时应清理连接。

## 工程规范

- 后端推荐使用“结构体 + 方法”组织能力，减少包级散落函数。
- 后端管理类结构体需显式持有 `logger`，关键流程输出开始/完成日志。
- 前端函数（组件函数、工具函数、关键回调）需添加简洁职责注释。
- 终端输入组件需支持命令 token 分词渲染与可执行命令有效性校验。
- Git 提交规范：`<图标> <类型>(<范围>): <简短描述>`（详见 `agents/git-commit-convention.md`）。

## 一键安装（Linux/macOS）

```bash
curl -fsSLO https://raw.githubusercontent.com/chenyang-zz/boxify/main/scripts/install.sh && sudo bash install.sh
```

## License

本项目采用 [MIT License](./LICENSE)。
