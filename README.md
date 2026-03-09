# Boxify

基于 Wails v3 的跨平台数据库管理桌面应用，提供 MySQL / PostgreSQL 可视化管理能力，并支持 SSH 隧道、文件管理与内置终端等开发辅助功能。

![Boxify](./boxify.png)

## 功能特性

- 多数据库支持：MySQL、PostgreSQL
- SSH 隧道连接数据库
- 数据表查询、编辑与批量修改
- 导入/导出：CSV、JSON、Markdown
- 表结构管理：字段、索引、外键、触发器
- 内置文件树与终端能力

## 技术栈

- 后端：Go、Wails v3、MySQL/PostgreSQL 驱动、SSH
- 前端：React 19、TypeScript、Vite、Tailwind CSS、shadcn/ui
- 包管理：Go Modules、pnpm

## 环境要求

- Go `1.25+`（当前 `go.mod` 为 `1.25`）
- Node.js `20+`（建议 LTS）
- pnpm `9+`（当前 lockfile 为 `v9`）
- Wails CLI（项目使用命令 `wails3`）

示例安装（按需）：

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

## 快速开始

```bash
# 1) 安装依赖（Go + 前端）
make init

# 2) 启动开发模式（前后端热重载）
make dev
```

## 一键安装（Linux/macOS）

```bash
curl -fsSLO https://raw.githubusercontent.com/chenyang-zz/boxify/main/scripts/install.sh && sudo bash install.sh
```

## 常用命令

```bash
# 开发与构建
make dev                # 启动开发模式
make build              # 构建生产版本
make git-release        # 兼容发布命令（默认 patch 自动升版本并推送 tag）
make git-release PART=minor  # 次版本 +1 并推送 tag
make git-release VERSION=0.0.18  # 手动指定版本号仅打 tag
make release-auto-tag PART=patch  # 自动升版本并推送 tag（默认 patch，由 GitHub Actions 发布）
make release-auto-tag PART=minor  # 次版本 +1 并推送 tag
make release-auto-tag PART=major  # 主版本 +1 并推送 tag

# 依赖与检查
make install            # 安装所有依赖
make check              # 检查 Go/前端/Git 状态
make tidy               # 整理 Go 依赖

# 质量保障
make test               # 运行测试
make test-coverage      # 生成覆盖率报告（coverage.html）
make format             # 格式化代码
make lint               # 代码检查

# 清理
make clean              # 清理构建产物
make clean-cache        # 清理缓存
```

仅前端开发：

```bash
cd frontend
pnpm install
pnpm run dev
pnpm run build
```

## 项目文档索引

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
├── frontend/               # React + TS 前端
├── agents/                 # 架构与协作规范文档
├── docs/                   # 详细设计文档
├── script/                 # 构建脚本
├── Makefile                # 开发与构建命令
└── README.md
```

## 开发注意事项

- MySQL 标识符使用反引号 `` ` ``，PostgreSQL 使用双引号 `"`。
- 前后端通信优先复用 `frontend/bindings` 生成类型，避免同义类型漂移。
- 错误信息统一使用中文。
- 连接池使用前需做有效性检查（Ping），应用退出时应清理连接。

## License

本项目采用 [MIT License](./LICENSE)。
