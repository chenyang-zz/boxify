# Boxify 后端架构（当前仓库）

更新时间：2026-03-07

```text
Boxify/
├── main.go                         # 应用入口（Wails 启动与服务装配）
├── go.mod                          # Go 依赖定义
├── internal/
│   ├── config/                     # 配置加载与解析（page config）
│   ├── connection/                 # 连接相关类型定义
│   ├── claw/                       # OpenClaw 相关能力（process 生命周期、monitor 监控）
│   ├── db/                         # 数据库抽象、连接管理与 MySQL 实现
│   ├── events/                     # 事件类型定义
│   ├── git/                        # Git 管理、解析、监听
│   ├── logger/                     # 日志能力
│   ├── redis/                      # Redis 相关模块（目录保留）
│   ├── service/                    # 应用服务层（DB/文件/Git/终端/窗口）
│   ├── ssh/                        # SSH 隧道能力
│   ├── terminal/                   # 终端会话与进程管理
│   ├── types/                      # 通用类型定义
│   ├── utils/                      # 工具函数
│   └── window/                     # 窗口注册与管理
├── docs/
│   ├── context-menu-guide.md
│   └── git-package-implementation.md
├── script/
│   └── build-release.sh
└── Makefile
```

## 分层摘要

1. 入口与装配：`main.go`
2. 服务编排：`internal/service`
3. 基础能力：`internal/db`、`internal/ssh`、`internal/terminal`、`internal/git`、`internal/claw`
4. 通用与支撑：`internal/types`、`internal/utils`、`internal/logger`
