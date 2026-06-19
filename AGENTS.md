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

格式：`<图标> <类型>(<范围>): <简短描述>`

| 类型 | 图标 | 说明 |
|------|------|------|
| `feat` | ✨ | 新功能 |
| `fix` | 🐛 | 修复 bug |
| `refactor` | ♻️ | 重构代码 |
| `docs` | 📝 | 文档更新 |
| `test` | ✅ | 测试相关 |
| `chore` | 🔧 | 构建/工具链 |
| `perf` | ⚡ | 性能优化 |

示例：`✨ feat(terminal): 添加目录选择器搜索功能`

详细规范见 `agents/git-commit-convention.md`

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
   - 类型优先复用 `frontend/bindings` 生成的后端类型（如 `@wails/types/models`）
   - 后端已定义返回结构时，前端不要新增同义接口，避免类型漂移
   - 错误信息使用中文

5. **后端实现风格**：
   - 优先采用“结构体 + 方法”组织可复用能力，避免核心逻辑散落在包级函数
   - 导出符号及关键内部函数需补齐职责注释
   - 结构体字段注释统一使用“右侧行尾注释”，避免字段上方堆叠注释
   - 后端注释统一使用中文表达（保留必要技术名词）
   - 需要日志的后端管理类结构体应显式持有 `logger` 字段，并通过构造函数注入
   - 关键流程需打印“开始/完成”，异常与降级路径需打印 `Warn`/`Error`
   - 详细规则见 `agents/backend-code-organization.md`

6. **前端实现风格**：
   - 前端新增或修改代码时，函数需补充简洁职责注释（工具函数、组件函数、关键回调函数）
   - 注释强调“做什么/为什么”，避免无信息量描述
   - 开发 `ClawContent` 组件时，可参考 `/Users/sheepzhao/WorkSpace/ClawPanel` 目录中的项目实现
   - 终端输入组件需支持命令行分词展示，不同 token 使用不同颜色
   - 命令 token 需基于当前 session 的可执行命令缓存校验：有效命令显示绿色；不存在命令显示红色虚线下划线


## 测试策略

1. **单元测试**：测试核心函数和工具方法
2. **集成测试**：测试数据库连接和操作
3. **端到端测试**：测试完整用户流程

## 资源链接

- [Wails 文档](https://v3alpha.wails.io/)
- [React 文档](https://react.dev/)
- [shadcn/ui 组件](https://ui.shadcn.com/)
- [Tailwind CSS](https://tailwindcss.com/)

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **boxify** (7980 symbols, 20849 relationships, 300 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> If any GitNexus tool warns the index is stale, run `npx gitnexus analyze` in terminal first.

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `gitnexus_impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `gitnexus_detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `gitnexus_query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `gitnexus_context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `gitnexus_impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `gitnexus_rename` which understands the call graph.
- NEVER commit changes without running `gitnexus_detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/boxify/context` | Codebase overview, check index freshness |
| `gitnexus://repo/boxify/clusters` | All functional areas |
| `gitnexus://repo/boxify/processes` | All execution flows |
| `gitnexus://repo/boxify/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
