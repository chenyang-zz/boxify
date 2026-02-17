# 上下文菜单系统使用文档

## 概述

Boxify 的上下文菜单系统是一个基于 Wails v3 的跨平台原生菜单解决方案，由后端服务（Go）和前端 Hook（TypeScript）组成，支持动态菜单创建、更新和事件处理。

## 架构设计

```
┌─────────────────────────────────────────────────────────────────┐
│                         前端 (React + TypeScript)                │
├─────────────────────────────────────────────────────────────────┤
│  useContextMenu Hook                                             │
│  ├── 自动创建/更新菜单                                            │
│  ├── 管理回调映射                                                 │
│  └── 监听 "menu:clicked" 事件                                    │
└─────────────────────────────────────────────────────────────────┘
                            ↕ Wails Bindings
┌─────────────────────────────────────────────────────────────────┐
│                         后端 (Go Service)                        │
├─────────────────────────────────────────────────────────────────┤
│  MenuService                                                     │
│  ├── CreateContextMenu()     创建原生菜单                        │
│  ├── UpdateMenu()            动态更新菜单项                      │
│  ├── UnregisterContextMenu() 注销菜单                            │
│  └── sendMenuEvent()         发送点击事件到前端                  │
└─────────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────────┐
│                      原生系统菜单 (macOS/Win/Linux)              │
└─────────────────────────────────────────────────────────────────┘
```

---

## 后端 API

### MenuService 方法

#### 1. CreateContextMenu

创建一个新的上下文菜单。

```go
func (ms *MenuService) CreateContextMenu(definition MenuDefinition) *connection.QueryResult
```

**参数：**
- `definition.MenuID` (string): 菜单唯一标识符，必须唯一
- `definition.Label` (string): 菜单标签（可选）
- `definition.Window` (string): 关联窗口名称（可选，默认当前窗口）
- `definition.Items` ([]MenuItemDefinition): 菜单项列表
- `definition.ContextData` (map[string]interface{}): 上下文数据（可选）

**返回值：**
```go
type QueryResult struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}
```

#### 2. UpdateMenu

动态更新已存在的菜单。

```go
func (ms *MenuService) UpdateMenu(request MenuUpdateRequest) *connection.QueryResult
```

**参数：**
- `request.MenuID` (string): 要更新的菜单 ID
- `request.Items` ([]MenuItemDefinition): 新的菜单项列表
- `request.ContextData` (map[string]interface{}): 新的上下文数据

#### 3. UnregisterContextMenu

注销并销毁指定菜单。

```go
func (ms *MenuService) UnregisterContextMenu(menuID string) *connection.QueryResult
```

#### 4. GetMenuList

获取所有已注册的菜单列表。

```go
func (ms *MenuService) GetMenuList() *connection.QueryResult
```

---

### 菜单项类型

| 类型 | 常量 | 描述 |
|------|------|------|
| 普通项 | `MenuItemTypeItem` | 可点击的菜单项 |
| 复选框 | `MenuItemTypeCheckbox` | 带选中状态的切换项 |
| 单选框 | `MenuItemTypeRadio` | 单选项（需要配合使用） |
| 分隔符 | `MenuItemTypeSeparator` | 视觉分隔线 |
| 子菜单 | `MenuItemTypeSubmenu` | 嵌套菜单 |

---

### 数据结构

#### MenuItemDefinition

```go
type MenuItemDefinition struct {
    ID          string                 // 菜单项 ID（唯一）
    Type        MenuItemType           // 类型
    Label       string                 // 显示文本
    Checked     *bool                  // 选中状态（checkbox/radio）
    Shortcut    *string                // 快捷键（如 "CmdOrCtrl+S"）
    Enabled     *bool                  // 是否启用
    Items       []MenuItemDefinition   // 子菜单项（submenu）
    ContextData map[string]interface{} // 上下文数据
}
```

#### MenuClickEvent

```go
type MenuClickEvent struct {
    MenuID      string                 // 菜单 ID
    ItemID      string                 // 菜单项 ID
    Type        MenuItemType           // 类型
    Label       string                 // 标签
    Checked     bool                   // 选中状态
    ContextData map[string]interface{} // 菜单的上下文数据
    ItemData    map[string]interface{} // 菜单项的上下文数据
    Timestamp   int64                  // 点击时间戳
    Window      string                 // 触发窗口
}
```

---

## 前端 API

### useContextMenu Hook

```typescript
function useContextMenu(menuConfig: MenuConfig): ContextMenuInstance
```

**参数：**
- `menuConfig.items` (MenuItemDefinition[]): 菜单项配置
- `menuConfig.contextData` (Record<string, any>): 上下文数据
- `menuConfig.window` (string): 窗口名称（可选）

**返回值：**
```typescript
interface ContextMenuInstance {
    open: (pos?: { x: number; y: number }) => void;
    update: (menuConfig: MenuConfig) => void;
}
```

---

### 前端数据类型

#### MenuItemDefinition

```typescript
interface MenuItemDefinition {
    id?: string;                      // 自动生成
    type: MenuItemType;
    label: string;
    checked?: boolean;
    shortcut?: string;
    enabled?: boolean;
    items?: MenuItemDefinition[];
    contextData?: Record<string, any>;
    onClick?: (payload: MenuClickPayload) => void;
}
```

#### MenuClickPayload

```typescript
interface MenuClickPayload {
    itemId: string;
    label: string;
    checked: boolean;
    contextData: Record<string, any>;
    itemData: Record<string, any>;
}
```

---

## 使用示例

### 基础示例

```tsx
import { useContextMenu } from "@/hooks/use-context-menu";
import { MenuItemType } from "@wails/service";

const MyComponent = () => {
    const menu = useContextMenu({
        items: [
            {
                type: MenuItemType.MenuItemTypeItem,
                label: "刷新",
                onClick: (payload) => {
                    console.log("刷新点击", payload);
                },
            },
            {
                type: MenuItemType.MenuItemTypeSeparator,
            },
            {
                type: MenuItemType.MenuItemTypeCheckbox,
                label: "显示隐藏文件",
                checked: false,
                onClick: (payload) => {
                    console.log("切换显示:", payload.checked);
                },
            },
        ],
        contextData: { source: "file-manager" },
    });

    return <button onClick={() => menu.open()}>打开菜单</button>;
};
```

### 子菜单示例

```tsx
const contextMenu = useContextMenu({
    items: [
        {
            type: MenuItemType.MenuItemTypeSubmenu,
            label: "新建",
            items: [
                {
                    type: MenuItemType.MenuItemTypeItem,
                    label: "文件夹",
                    onClick: () => console.log("创建文件夹"),
                },
                {
                    type: MenuItemType.MenuItemTypeItem,
                    label: "文件",
                    onClick: () => console.log("创建文件"),
                },
            ],
        },
    ],
});
```

### 动态更新菜单

```tsx
const menu = useContextMenu(initialConfig);

// 更新菜单
menu.update({
    items: [
        {
            type: MenuItemType.MenuItemTypeItem,
            label: "新选项",
            onClick: () => { /* ... */ },
        },
    ],
});
```

### 在按钮点击位置打开

```tsx
const ButtonWithMenu = () => {
    const menu = useContextMenu({
        items: [
            {
                type: MenuItemType.MenuItemTypeItem,
                label: "操作",
                onClick: (payload) => console.log(payload),
            },
        ],
    });

    return (
        <button onClick={(e) => menu.open({ x: e.clientX, y: e.clientY })}>
            点击打开菜单
        </button>
    );
};
```

---

## 完整示例：TreeHeader 组件

```tsx
import { useContextMenu } from "@/hooks/use-context-menu";
import { MenuItemType } from "@wails/service";
import { PlusIcon } from "lucide-react";

const TreeHeader = () => {
    const addMenu = useContextMenu({
        items: [
            {
                label: "目录",
                type: MenuItemType.MenuItemTypeItem,
                onClick: async (payload) => {
                    console.log("创建目录:", payload);
                },
            },
            {
                label: "数据库",
                type: MenuItemType.MenuItemTypeSubmenu,
                items: [
                    {
                        label: "MySQL",
                        type: MenuItemType.MenuItemTypeItem,
                        onClick: async (payload) => {
                            console.log("创建 MySQL 连接:", payload);
                        },
                    },
                    {
                        label: "PostgreSQL",
                        type: MenuItemType.MenuItemTypeItem,
                        onClick: async (payload) => {
                            console.log("创建 PostgreSQL 连接:", payload);
                        },
                    },
                ],
            },
            {
                label: "远程连接",
                type: MenuItemType.MenuItemTypeSubmenu,
                items: [
                    { label: "SSH", type: MenuItemType.MenuItemTypeItem, onClick: /* ... */ },
                    { label: "RDP", type: MenuItemType.MenuItemTypeItem, onClick: /* ... */ },
                    { label: "Telnet", type: MenuItemType.MenuItemTypeItem, onClick: /* ... */ },
                ],
            },
        ],
        contextData: { source: "property-tree" },
    });

    return (
        <button onClick={(e) => addMenu.open({ x: e.clientX, y: e.clientY })}>
            <PlusIcon className="size-4" />
            创建资产
        </button>
    );
};

export default TreeHeader;
```

---

## 快捷键格式

快捷键使用 Wails 格式，支持以下修饰键：

| 修饰键 | 格式 |
|--------|------|
| Ctrl | `Ctrl` 或 `CmdOrCtrl` |
| Command (macOS) | `Cmd` 或 `CmdOrCtrl` |
| Alt/Option | `Alt` 或 `Option` |
| Shift | `Shift` |

**示例：**
```typescript
{
    type: MenuItemType.MenuItemTypeItem,
    label: "保存",
    shortcut: "CmdOrCtrl+S",
    onClick: () => { /* ... */ },
}
```

---

## 生命周期管理

### 创建流程

1. 前端调用 `useContextMenu`
2. Hook 生成唯一 `menuId`
3. 调用 `MenuService.CreateContextMenu`
4. 后端创建原生菜单并缓存
5. 注册全局事件监听器

### 更新流程

1. 调用 `menu.update()`
2. Hook 提取新的回调映射
3. 调用 `MenuService.UpdateMenu`
4. 后端清空旧菜单项并重建

### 销毁流程

1. 组件卸载触发 cleanup
2. 移除回调映射
3. 注销后端菜单
4. 清理全局监听器

---

## 注意事项

### 后端开发

1. **并发安全**：`MenuService` 使用 `sync.RWMutex` 保护菜单缓存
2. **资源清理**：服务关闭时会自动清理所有菜单
3. **事件发送**：使用 `ms.App().Event.Emit()` 发送事件到前端

### 前端开发

1. **回调管理**：所有菜单共享一个全局事件监听器，提高性能
2. **UUID 前缀**：菜单项 ID 自动添加 `menu-item-` 前缀避免冲突
3. **位置记忆**：`lastPositionRef` 记录上次打开位置

### 最佳实践

1. **菜单 ID 唯一性**：不要手动指定可能冲突的 ID
2. **上下文数据**：善用 `contextData` 传递菜单相关的业务数据
3. **快捷键**：为常用操作添加快捷键
4. **子菜单层级**：避免过深的嵌套（建议不超过 3 层）
5. **复选框状态**：checkbox/radio 状态在点击时自动切换

---

## 错误处理

### 常见错误

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| "菜单 ID 不能为空" | `MenuID` 为空 | 确保提供有效的 `menuId` |
| "菜单已存在" | ID 冲突 | 使用自动生成的 ID 或注销旧菜单 |
| "菜单不存在" | 更新/注销不存在的菜单 | 检查菜单是否已创建 |
| "未知的菜单项类型" | Type 值错误 | 使用 `MenuItemType` 枚举 |

---

## 类型参考

### MenuItemType 枚举

```typescript
enum MenuItemType {
    MenuItemTypeItem = "item",
    MenuItemTypeCheckbox = "checkbox",
    MenuItemTypeRadio = "radio",
    MenuItemTypeSeparator = "separator",
    MenuItemTypeSubmenu = "submenu",
}
```

---

## 相关文件

- 后端服务：`internal/service/menu_service.go`
- 前端 Hook：`frontend/src/hooks/use-context-menu.ts`
- 类型定义：`frontend/src/types/menu.ts`
- 示例组件：`frontend/src/components/PropertyTree/TreeHeader.tsx`
