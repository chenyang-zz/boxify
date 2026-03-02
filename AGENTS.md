# Boxify 项目指南

## 项目概述

Boxify 是一个基于 Wails 框架的跨平台数据库管理桌面应用，支持 MySQL 和 PostgreSQL 数据库的可视化管理。

### 核心功能
- 多数据库支持：MySQL、PostgreSQL
- SSH 隧道连接
- 数据查询与编辑
- 数据导入/导出（CSV、JSON、Markdown）
- 批量数据修改
- 表结构管理（查看、创建、修改）
- 索引、外键、触发器管理

## 技术栈

### 后端
- **Go** - 主要后端语言
- **Wails v3** - 桌面应用框架
- **数据库驱动**:
  - MySQL: `go-sql-driver/mysql`
  - PostgreSQL: `lib/pq`
- **SSH** - golang.org/x/crypto/ssh

### 前端
- **React 18** - UI 框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **Tailwind CSS** - 样式框架
- **shadcn/ui** - UI 组件库
- **pnpm** - 包管理器

## 项目结构

项目架构已迁移为独立文档，请查看：

- `agents/backend-architecture.md`
- `agents/frontend-architecture.md`
- `agents/project-architecture.md`（索引）
- `agents/backend-code-organization.md`（后端代码组织规则）

说明：
- `AGENTS.md` 保留规范与流程说明，避免在此维护易过期的静态目录树
- 后端结构变更时更新 `agents/backend-architecture.md`
- 前端结构变更时更新 `agents/frontend-architecture.md`
- 后端模块分层/职责调整时更新 `agents/backend-code-organization.md`


## 常用开发命令

### 前端开发
```bash
cd frontend
pnpm install          # 安装依赖
pnpm run dev          # 启动开发服务器
pnpm run build        # 构建生产版本
```

### 应用开发
```bash
make dev             # 启动热重载开发模式（包括前后端）
make build           # 构建生产版本
```

### Go 代码
```bash
go fmt ./...          # 格式化代码
go mod tidy           # 整理依赖
go test ./...         # 运行测试
```

## Git 提交规范

提交规范已迁移为独立文档，请查看：

- `agents/git-commit-convention.md`

## 重要注意事项

1. **数据库兼容性**：
   - MySQL 使用反引号 ``` ` ``` 包裹标识符
   - PostgreSQL 使用双引号 `"` 包裹标识符
   - 注意不同数据库的 SQL 语法差异

2. **连接缓存**：
   - 连接基于配置参数生成唯一 key
   - 使用前检查连接是否有效（Ping）
   - 应用关闭时清理所有连接

3. **字符编码**：
   - 默认使用 `utf8mb4` 字符集（MySQL）
   - 支持 emoji 和特殊字符

4. **前端-后端通信**：
   - 通过 Wails 绑定 Go 方法到前端
   - 通过callWails调用
   - 错误信息使用中文


## 测试策略

1. **单元测试**：测试核心函数和工具方法
2. **集成测试**：测试数据库连接和操作
3. **端到端测试**：测试完整用户流程

## 资源链接

- [Wails 文档](https://v3alpha.wails.io/)
- [React 文档](https://react.dev/)
- [shadcn/ui 组件](https://ui.shadcn.com/)
- [Tailwind CSS](https://tailwindcss.com/)
