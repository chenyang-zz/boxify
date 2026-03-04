# Boxify 终端组件架构文档

更新时间：2026-03-04

## 1. 文档目标

本文档用于说明 `frontend/src/components/Terminal` 的当前架构设计、职责边界和演进规范，帮助后续开发在不破坏行为的前提下持续扩展。

适用范围：

1. 终端会话创建与销毁
2. 命令提交流程
3. 输出渲染与事件分发
4. 输入编辑器与命令校验
5. Git 状态联动与审查面板

## 2. 模块结构

```text
frontend/src/components/Terminal/
├── index.tsx                         # 终端入口（布局+会话绑定）
├── TerminalCore.tsx                  # 终端核心视图（输出区+输入区）
├── app/
│   ├── index.ts                      # app 层导出
│   ├── terminal-application.ts       # 用例编排（提交命令/事件落库）
│   ├── use-terminal-controller.ts    # 会话初始化+环境同步+尺寸联动
│   └── use-input-editor-controller.ts# 输入控制器（快捷键/高亮/命令校验）
├── lib/
│   ├── session-manager.ts            # Wails 终端适配与事件汇聚（infra）
│   └── ansi-parser.ts                # ANSI 解析器（输出格式化）
├── store/
│   ├── terminal.store.ts             # 终端 Zustand 状态
│   ├── selectors.ts                  # 选择器
│   └── index.ts
├── domain/
│   ├── command-parser.ts             # 命令分词/分类（纯函数）
│   ├── block-reducer.ts              # block 变更 reducer（纯函数）
│   ├── command-parser.unit.ts        # 领域单测（可独立执行）
│   ├── block-reducer.unit.ts         # 领域单测（可独立执行）
│   └── index.ts
├── components/
│   ├── InputEditor.tsx               # 输入区视图
│   ├── OutputRenderer.tsx            # 输出渲染视图
│   ├── TerminalBlock.tsx             # 命令块视图
│   ├── DirectorySelector.tsx
│   └── GitReviewPanel.tsx
└── types/
    ├── block.ts
    ├── terminal.ts
    └── index.ts
```

## 3. 分层职责

### 3.1 Container 层

文件：`index.tsx`

职责：

1. 根据 `sessionId` 读取终端配置并校验有效性
2. 调用 `terminalSessionManager.getOrCreate` 预创建会话缓存
3. 调用 `terminalApplication.bindSession` 绑定事件消费
4. 组织布局（终端区 + GitReviewPanel 可伸缩分栏）

### 3.2 Application 层

文件：`app/terminal-application.ts`

职责：

1. `submitCommand`：统一处理空命令、命令写入、block 创建、历史入库
2. 消费 `session-manager` 事件并落地到 store
3. 管理会话绑定状态，避免重复绑定

该层是 UI 与 Infra 的唯一业务编排入口。

### 3.3 Infra 层

文件：`lib/session-manager.ts`

职责：

1. 调用 Wails `TerminalService/GitService` 完成 IO
2. 监听后端事件（output/error/command_end/pwd_update/git_update）
3. 进行输出队列合并（`requestAnimationFrame` 批量刷新）
4. 将结果以 `TerminalSessionEvent` 发给 application

约束：

1. Infra 不直接写 Zustand
2. Infra 不承载 UI 逻辑

### 3.4 Domain 层

文件：`domain/command-parser.ts`、`domain/block-reducer.ts`

职责：

1. 命令分词、命令/参数/路径/操作符分类
2. block 生命周期变更（创建、追加输出、结束、状态更新）

约束：

1. 纯函数，无 React / Wails 依赖
2. 可独立测试

### 3.5 UI 层

文件：`TerminalCore.tsx`、`components/*`

职责：

1. 只负责渲染和交互事件绑定
2. 通过 controller/app 层触发行为，不直接处理底层协议

## 4. 核心数据模型

### 4.1 会话 ID

`sessionId` 是终端模块的主键，贯穿：

1. Wails 终端会话
2. Zustand 分片状态
3. 事件路由

### 4.2 Block 模型

每条命令对应一个 block，关键字段：

1. `id`：与后端命令执行 ID 对齐
2. `command`：原始命令文本
3. `output[]`：输出行
4. `status`：`running/success/error/pending`
5. `startTime/endTime/exitCode`

## 5. 关键流程

### 5.1 初始化流程

1. `Terminal/index.tsx` 校验配置
2. `session-manager.getOrCreate(sessionId)`
3. `terminalApplication.bindSession(sessionId)`
4. `useTerminalController.initialize(...)` 创建后端终端
5. `setEnvChangeCallback` 将环境信息同步到 `envInfo`

### 5.2 命令提交流程

1. `InputEditor` 触发 `onSubmit`
2. `TerminalCore` 调 `terminalApplication.submitCommand`
3. Application 调 `writeCommand`
4. 返回 `blockId` 后写入 store：`createBlock + addToHistory`

### 5.3 输出回流流程

1. 后端推送 `terminal:output`
2. `session-manager` 解码 Base64 + ANSI 解析
3. 输出进入队列并在帧级批量合并
4. 发出 `output_batch` 事件
5. Application 调 `store.appendOutputBatch`

### 5.4 错误与结束流程

1. `terminal:error` -> 发 `error` 事件 -> block 标红并 finalize
2. `terminal:command_end` -> 发 `command_end` -> finalizeBlock(exitCode)

### 5.5 环境与 Git 联动

1. `terminal:pwd_update` 更新工作目录
2. `GitService.RegisterRepo/SetActiveRepo/GetInitialStatusEvent`
3. `terminal:git_update` 同步分支和变更统计
4. 输入区与 GitReviewPanel 使用 event/store 展示

## 6. 状态管理设计

文件：`store/terminal.store.ts`

状态分片：

1. `sessionBlocks`：命令块列表
2. `sessionHistory/historyIndexes`：历史命令导航
3. `reviewPanelOpenBySession`：审查面板开关

核心 actions：

1. `createBlock`
2. `appendOutputBatch`
3. `finalizeBlock`
4. `addToHistory/navigateHistory`
5. `openReviewPanel/closeReviewPanel`
6. `clearSession`

## 7. 输入编辑器设计

文件：`app/use-input-editor-controller.ts` + `components/InputEditor.tsx`

能力点：

1. 键盘快捷键（Enter、历史导航、Ctrl+C/L/A/E/U/K）
2. 自适应输入高度
3. 命令缓存校验（来自 `session-manager.getExecutableCommandCache`）
4. token 高亮（命令、选项、变量、字符串、路径、操作符）
5. Git 状态入口与目录选择器联动

## 8. 输出渲染设计

文件：`lib/ansi-parser.ts` + `components/OutputRenderer.tsx`

策略：

1. 解析 ANSI SGR 序列，输出 `FormattedChar[]`
2. 处理控制字符（回车、换行、退格、制表）
3. UI 按字符样式渲染（粗体/斜体/下划线/反色等）
4. 默认 ANSI 色板内置于 parser，Terminal 不再维护主题系统

## 9. 性能与稳定性策略

1. 输出批处理：`requestAnimationFrame` 合并输出，减少高频 setState
2. 选择器分片：按 `sessionId` 读取状态，降低无关重渲染
3. 纯函数下沉：`domain` 层可单测，降低回归风险
4. 生命周期解耦：Tab 切换不销毁会话，关闭 Tab 时统一清理

## 10. 扩展规范

新增能力时遵循：

1. 新增协议调用：优先放 `lib/session-manager.ts`
2. 新增业务流程：放 `app/terminal-application.ts` 或 controller hook
3. 新增纯逻辑：放 `domain/*` 并补 `.unit.ts`
4. UI 组件避免直接访问 Wails API
5. store 只做状态落地，不做协议编排

## 11. 测试建议

当前已有：

1. `domain/command-parser.unit.ts`
2. `domain/block-reducer.unit.ts`

建议继续补充：

1. `terminal-application` 事件映射测试
2. `session-manager` 事件适配测试（mock Wails events）
3. `InputEditor` 快捷键行为测试

## 12. 维护清单

当以下内容变化时，需要同步更新本文档：

1. `TerminalSessionEvent` 事件类型
2. store 状态结构
3. 命令提交与输出回流流程
4. Terminal 目录结构
5. session 生命周期约束（谁负责销毁）
