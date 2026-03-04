# Boxify 前端架构（当前仓库）

更新时间：2026-03-04

```text
Boxify/frontend/
├── package.json
├── pnpm-lock.yaml
├── vite.config.ts
├── tsconfig.json
├── index.html
├── src/
│   ├── main.tsx                    # 前端入口
│   ├── App.tsx                     # 主应用组件
│   ├── style.css
│   ├── assets/                     # 字体、图片等资源
│   ├── common/                     # 通用常量/枚举
│   ├── constants/                  # 业务常量
│   ├── components/                 # 功能组件与 UI 组件
│   │   ├── ui/
│   │   ├── TitleBar/
│   │   ├── PropertyTree/
│   │   ├── UtilBar/
│   │   ├── DBTable/
│   │   ├── FileTree/
│   │   ├── Terminal/
│   │   └── WindowHeader/
│   ├── hooks/                      # React Hooks
│   ├── lib/                        # 前端业务库（连接、SQL、属性、同步）
│   ├── pages/                      # 页面级模块（main/settings/connectionEdit）
│   ├── providers/                  # Provider（如 ThemeProvider）
│   ├── store/                      # 状态管理
│   └── types/                      # 前端类型定义
├── bindings/                       # Wails 生成绑定
├── public/                         # 静态资源
├── dist/                           # 构建产物
└── plugins/                        # 前端插件
```

## 分层摘要

1. 启动层：`src/main.tsx`
2. 页面层：`src/pages`
3. 组件层：`src/components`
4. 状态与逻辑层：`src/store`、`src/hooks`、`src/lib`
5. 桥接层：`bindings`（Wails 前后端调用绑定）

## 类型约束

1. 前后端交互数据结构优先使用 `frontend/bindings` 中生成的类型（例如 `@wails/types/models`）。
2. 若后端 `internal/types` 已有对应定义，前端禁止在 `src/types` 或业务代码中新增同义接口。
3. 仅在纯前端 UI 局部状态且无后端对应模型时，才新增前端自定义类型。

## 代码注释规范

1. 前端函数必须补充简洁职责注释，至少覆盖：
   - 工具函数（如解析、格式化、转换）
   - 组件函数（说明组件核心职责）
   - 关键回调函数（如键盘事件、提交流程、状态同步）
2. 注释内容聚焦“做什么/为什么”，避免“给变量赋值”这类低价值描述。
3. 注释保持简短，优先一行说明；复杂逻辑可在代码块前补充 1-2 行解释。

## Terminal 输入组件约定

1. `InputEditor` 需要对命令行输入做分词渲染，按 token 类型区分颜色（命令、选项、变量、字符串、路径、操作符等）。
2. 主命令 token 需要基于当前 `session` 的可执行命令缓存（`terminalSessionManager.getExecutableCommandCache`）进行有效性校验。
3. 有效命令显示绿色，不存在命令显示红色虚线下划线。
4. 当 session 命令缓存异步刷新后，输入区应同步刷新校验状态，避免用户需要再次输入才触发更新。

## Terminal 详细架构文档

终端组件详细分层、数据流和扩展规范请查看：

- `docs/terminal-component-architecture.md`
