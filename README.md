# Boxify
<p align="center">
    <picture>
        <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/chenyang-zz/boxify/main/boxify-logo-text.png">
        <img src="https://raw.githubusercontent.com/chenyang-zz/boxify/main/boxify-logo-text.png" alt="Boxify" width="500">
    </picture>
</p>


<p align="center">
  <a href="https://github.com/chenyang-zz/boxify/actions/workflows/ci.yml?branch=main"><img src="https://img.shields.io/github/actions/workflow/status/chenyang-zz/boxify/ci.yml?branch=main&style=for-the-badge" alt="CI status"></a>
  <a href="https://github.com/chenyang-zz/boxify/releases"><img src="https://img.shields.io/github/v/release/chenyang-zz/boxify?include_prereleases&style=for-the-badge" alt="GitHub release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge" alt="MIT License"></a>
</p>

基于 Wails v3 的跨平台数据库管理桌面应用，面向 MySQL / PostgreSQL 提供可视化连接、查询、结构管理与开发辅助能力。

## 项目定位

Boxify 不是只做 SQL 查询窗口的轻量工具，它同时覆盖数据库日常操作和本地开发协作场景：

- 管理 MySQL 与 PostgreSQL 连接
- 通过 SSH 隧道访问远程数据库
- 查询、编辑、批量修改表数据
- 管理字段、索引、外键、触发器等结构对象
- 导入/导出 CSV、JSON、Markdown
- 使用内置文件树与终端处理开发辅助任务

## 核心特性

- 多数据库支持：MySQL、PostgreSQL
- SSH 隧道连接数据库
- 数据查询、编辑与批量修改
- 表结构管理：字段、索引、外键、触发器
- 导入/导出：CSV、JSON、Markdown
- 内置终端与文件树能力
- 基于 Wails bindings 的前后端类型安全调用

## 技术栈

- 后端：Go `1.25`、Wails v3、`go-sql-driver/mysql`、`lib/pq`、`golang.org/x/crypto/ssh`
- 前端：React `19`、TypeScript、Vite、Tailwind CSS、shadcn/ui
- 包管理：Go Modules、pnpm

## 环境要求

- Go `1.25+`
- Node.js `20+`
- pnpm `9+`
- Wails CLI `wails3`

安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

## 快速开始

```bash
# 安装依赖
make install

# 启动前后端热重载开发
make dev
```

如果只需要前端独立开发：

```bash
cd frontend
pnpm install
pnpm run dev
```

## 常用命令

```bash
# 开发与构建
make dev
make build
make frontend-dev
make frontend-build

# 质量保障
make test
make format
make tidy
make lint

# macOS 打包
make refresh-icons
make build-macos-app
make build-macos-app-universal
make run-macos-app
```

## 架构概览

Boxify 采用 Wails 前后端一体化桌面架构：

1. 后端负责数据库连接、终端、文件系统、窗口与系统能力编排。
2. 前端负责界面渲染、交互逻辑与状态管理。
3. Wails bindings 负责前后端 RPC 调用和事件通信。

核心入口：

- 应用入口：`main.go`
- 后端服务：`internal/service/*`
- 窗口管理：`internal/window/*`
- 前端入口：`frontend/src/main.tsx`

## 目录概览

```text
Boxify/
├── main.go
├── internal/      # Go 后端核心逻辑
├── frontend/      # React + TypeScript 前端
├── docs/          # 设计与流程文档
├── agents/        # 架构与协作规范
├── scripts/       # 构建与安装脚本
└── Makefile
```

## 文档索引

- 项目架构索引：`agents/project-architecture.md`
- 后端架构：`agents/backend-architecture.md`
- 前端架构：`agents/frontend-architecture.md`
- 后端代码组织规范：`agents/backend-code-organization.md`
- Git 提交规范：`agents/git-commit-convention.md`
- 发布流程：`docs/release-guide.md`
- 终端组件架构：`docs/terminal-component-architecture.md`

## 开发约定

- MySQL 标识符使用反引号，PostgreSQL 使用双引号。
- 前端优先复用 `frontend/bindings` 生成类型，避免同义类型漂移。
- 错误信息统一使用中文。
- 连接使用前需要探活，应用退出时需要清理连接。

更细的实现规范请查看 [AGENTS.md](./AGENTS.md) 和 `agents/` 目录文档。

## 安装脚本

```bash
curl -fsSL https://raw.githubusercontent.com/chenyang-zz/boxify/main/scripts/install.sh | sudo bash
```

## License

本项目采用 [MIT License](./LICENSE)。
