# Boxify DBTable 组件架构文档

更新时间：2026-03-06

## 1. 目标

定义 `frontend/src/components/DBTable` 的分层职责，避免“视图 + 协议 + 状态”耦合在单个组件中，方便后续扩展筛选、排序、批量编辑与导入导出能力。

## 2. 目录结构

```text
frontend/src/components/DBTable/
├── index.tsx                        # 容器组件（布局、表格渲染、交互绑定）
├── HeaderAction.tsx                 # 兼容导出（re-export）
├── app/
│   └── use-db-table-controller.ts   # 应用层编排（加载/事务/保存/撤销重做）
├── domain/
│   └── draft.ts                     # 领域纯函数（草稿行、变更集、筛选排序）
├── components/
│   └── HeaderAction.tsx             # 头部操作栏视图
└── types/
    └── index.ts                     # 组件内部类型定义
```

## 3. 分层职责

1. `index.tsx`：只负责 UI 展示与事件绑定，不直接调用 Wails API。
2. `app/use-db-table-controller.ts`：负责动作编排，管理编辑事务、历史栈、选中态、异步加载和保存流程。
3. `domain/draft.ts`：负责纯逻辑计算，包括：
   - 后端数据转草稿行
   - 单元格变更、行删除切换
   - ChangeSet 构建
   - 关键字筛选和列排序
4. `components/HeaderAction.tsx`：负责按钮渲染、状态展示、事件透传。
5. `src/lib/db-table.ts`：作为 DBTable 的 infra 入口，封装 `DBGetColumns/DBQuery/ApplyChanges/ImportData/ExportTable`。

## 4. 核心流程

1. 初始化：`index.tsx` 挂载后调用 `controller.load()` 拉取列定义与表数据。
2. 开始事务：进入前端草稿编辑模式，允许增删改。
3. 保存：`domain.buildChangeSet` 生成变更集，`lib/db-table.ts` 调 `DatabaseService.ApplyChanges` 提交。
4. 刷新：重置草稿态并重新拉取数据。
5. 导入/导出：controller 调用 `importDBTableByUUID/exportDBTableByUUID`，导入后自动刷新。

## 5. 扩展约定

1. 新协议调用优先添加到 `src/lib/db-table.ts`。
2. 新业务动作优先添加到 controller。
3. 可测试的纯计算放到 `domain`，避免直接依赖 React/Wails。
4. HeaderAction 保持无状态渲染，避免在视图层累积业务逻辑。
