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
- **Wails v2** - 桌面应用框架
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

```
Boxify/
├── main.go              # 应用入口点
├── app.go               # 核心应用逻辑和数据库操作
├── database.go          # 数据库接口定义
├── mysql_impl.go        # MySQL 实现
├── postgres_impl.go     # PostgreSQL 实现（如存在）
├── changeset.go         # 批量修改数据结构
├── ssh.go               # SSH 隧道连接
├── utils.go             # 工具函数
├── internal/            # 内部包（新开发）
├── frontend/            # 前端代码
│   ├── src/
│   │   ├── main.tsx     # React 入口
│   │   ├── App.tsx      # 主应用组件
│   │   ├── components/  # React 组件
│   │   │   ├── ui/      # shadcn/ui 组件
│   │   │   ├── TitleBar/    # 标题栏
│   │   │   ├── PropertyTree/ # 属性树
│   │   │   └── UtilBar/  # 工具栏
│   │   └── hooks/       # React Hooks
│   └── package.json
├── wails.json           # Wails 配置
├── go.mod               # Go 依赖
└── Makefile             # 构建脚本
```

## 开发规范

### 代码风格

#### Go 代码
1. **遵循 Go 官方代码规范**：使用 `gofmt` 格式化代码
2. **错误处理**：所有错误必须处理，不要忽略
3. **命名规范**：
   - 导出函数/变量使用大驼峰（PascalCase）
   - 私有函数/变量使用小驼峰（camelCase）
   - 常量使用全大写下划线分隔
4. **注释**：导出的函数必须有注释说明
5. **并发安全**：使用 `sync.Mutex` 保护共享资源（如 `dbCache`）

#### TypeScript/React 代码
1. **组件命名**：使用大驼峰（PascalCase）
2. **文件命名**：组件文件使用小写短横线（kebab-case）或小驼峰
3. **类型定义**：优先使用 TypeScript 类型，避免 `any`
4. **状态管理**：使用 React Hooks（useState、useEffect 等）
5. **样式**：使用 Tailwind CSS 工具类

### 文件组织规则

1. **后端文件**：
   - 每个主要功能模块一个文件
   - 接口定义在 `database.go`
   - 具体实现在 `*_impl.go`

2. **前端组件**：
   - 可复用组件放在 `components/ui/`
   - 功能组件放在对应的功能目录下
   - 每个组件目录包含 `index.tsx` 和相关文件

### 函数/方法设计规则

1. **返回值规范**：
   ```go
   // 所有数据库操作返回 QueryResult
   type QueryResult struct {
       Success bool        `json:"success"`
       Message string      `json:"message"`
       Data    interface{} `json:"data"`
       Fields  []string    `json:"fields"`
   }
   ```

2. **错误消息**：使用中文描述错误信息

3. **数据库连接管理**：
   - 使用连接缓存（`dbCache`）
   - 实现连接池
   - 自动重连机制

### 安全规则

1. **SQL 注入防护**：
   - 使用参数化查询
   - 对表名、字段名进行转义
   - 使用反引号（MySQL）或双引号（PostgreSQL）包裹标识符

2. **敏感信息**：
   - 不要在日志中输出密码
   - 使用环境变量或配置文件管理敏感信息

3. **输入验证**：
   - 验证所有用户输入
   - 限制查询执行时间
   - 防止 DoS 攻击

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
wails dev             # 启动热重载开发模式（包括前后端）
wails build           # 构建生产版本
```

### Go 代码
```bash
go fmt ./...          # 格式化代码
go mod tidy           # 整理依赖
go test ./...         # 运行测试
```

## Git 提交规范

使用中文提交信息，格式：

```
<类型>: <简短描述>

<详细描述（可选）>
```

类型：
- `feat`: 新功能
- `fix`: 修复 bug
- `refactor`: 重构代码
- `docs`: 文档更新
- `style`: 代码格式调整
- `test`: 测试相关
- `chore`: 构建/工具链相关

示例：
```
feat: 添加 PostgreSQL 数据库支持

实现了 PostgreSQL 的连接、查询、表结构查看等基础功能
```

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
   - 所有返回数据使用 JSON 序列化
   - 错误信息使用中文

## 扩展指南

### 添加新的数据库支持

1. 在 `database.go` 中实现 `Database` 接口
2. 创建新的 `*_impl.go` 文件
3. 在 `app.go` 中添加相应的连接和操作方法
4. 更新前端界面以支持新数据库类型

### 添加新的前端功能

1. 在 `frontend/src/components/` 创建新组件
2. 使用现有的 UI 组件库（shadcn/ui）
3. 通过 Wails 调用后端 Go 方法
4. 遵循 React Hooks 模式

## 测试策略

1. **单元测试**：测试核心函数和工具方法
2. **集成测试**：测试数据库连接和操作
3. **端到端测试**：测试完整用户流程

## 资源链接

- [Wails 文档](https://wails.io/docs/introduction)
- [React 文档](https://react.dev/)
- [shadcn/ui 组件](https://ui.shadcn.com/)
- [Tailwind CSS](https://tailwindcss.com/)
