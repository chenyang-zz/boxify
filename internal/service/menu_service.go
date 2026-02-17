// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// MenuItemType 菜单项类型枚举
type MenuItemType string

const (
	MenuItemTypeItem      MenuItemType = "item"      // 普通菜单项
	MenuItemTypeCheckbox  MenuItemType = "checkbox"  // 复选框
	MenuItemTypeRadio     MenuItemType = "radio"     // 单选框
	MenuItemTypeSeparator MenuItemType = "separator" // 分隔符
	MenuItemTypeSubmenu   MenuItemType = "submenu"   // 子菜单
)

// MenuClickEvent 菜单点击事件
type MenuClickEvent struct {
	MenuID      string                 `json:"menuId"`      // 菜单 ID
	ItemID      string                 `json:"itemId"`      // 菜单项 ID
	Type        MenuItemType           `json:"type"`        // 菜单项类型
	Label       string                 `json:"label"`       // 菜单项标签
	Checked     bool                   `json:"checked"`     // 是否选中（checkbox/radio）
	ContextData map[string]interface{} `json:"contextData"` // 菜单的上下文数据
	ItemData    map[string]interface{} `json:"itemData"`    // 菜单项的上下文数据
	Timestamp   int64                  `json:"timestamp"`   // 点击时间戳
	Window      string                 `json:"window"`      // 触发窗口
}

// MenuService 菜单管理服务
type MenuService struct {
	BaseService
	menus map[string]*MenuWrapper // 菜单缓存: menuID -> MenuWrapper
	mu    sync.RWMutex            // 保护 menus 的并发访问
}

// MenuWrapper 封装 Wails 菜单和元数据
type MenuWrapper struct {
	menuID    string                   // 菜单唯一标识
	context   *application.ContextMenu // Wails 菜单实例
	metadata  MenuMetadata             // 菜单元数据
	createdAt int64                    // 创建时间戳
}

// MenuMetadata 菜单元数据
type MenuMetadata struct {
	Label       string                 `json:"label"`       // 菜单标签
	Description string                 `json:"description"` // 菜单描述
	Window      string                 `json:"window"`      // 关联窗口（空表示全局）
	ContextData map[string]interface{} `json:"contextData"` // 上下文数据
}

// MenuDefinition 菜单定义（用于前端创建菜单）
type MenuDefinition struct {
	MenuID      string                 `json:"menuId"`      // 菜单 ID
	Label       string                 `json:"label"`       // 菜单标签（可选）
	Window      string                 `json:"window"`      // 关联窗口
	Items       []MenuItemDefinition   `json:"items"`       // 菜单项列表
	ContextData map[string]interface{} `json:"contextData"` // 上下文数据（可选）

}

// MenuItemDefinition 菜单项定义
type MenuItemDefinition struct {
	ID          string                 `json:"id"`          // 菜单项 ID
	Type        MenuItemType           `json:"type"`        // 类型: item|checkbox|radio|separator|submenu
	Label       string                 `json:"label"`       // 标签
	Checked     *bool                  `json:"checked"`     // 是否选中（checkbox/radio）
	Shortcut    *string                `json:"shortcut"`    // 快捷键
	Enabled     *bool                  `json:"enabled"`     // 是否启用
	Items       []MenuItemDefinition   `json:"items"`       // 子菜单项（submenu）
	ContextData map[string]interface{} `json:"contextData"` // 上下文数据
}

// MenuUpdateRequest 菜单更新请求
type MenuUpdateRequest struct {
	MenuID      string                 `json:"menuId"`      // 菜单 ID
	Items       []MenuItemDefinition   `json:"items"`       // 新的菜单项列表
	ContextData map[string]interface{} `json:"contextData"` // 新的上下文数据
}

// NewMenuService 创建 MenuService
func NewMenuService(deps *ServiceDeps) *MenuService {
	return &MenuService{
		BaseService: NewBaseService(deps),
		menus:       make(map[string]*MenuWrapper),
	}
}

// ServiceStartup 服务启动
func (ms *MenuService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	ms.SetContext(ctx)
	ms.Logger().Info("服务启动", "service", "MenuService")
	return nil
}

// ServiceShutdown 服务关闭
func (ms *MenuService) ServiceShutdown() error {
	ms.Logger().Info("服务开始关闭，准备释放资源", "service", "MenuService")

	ms.mu.Lock()
	defer ms.mu.Unlock()

	// 清理所有菜单
	for menuID, wrapper := range ms.menus {
		wrapper.context.Destroy()
		ms.Logger().Debug("销毁菜单", "menuId", menuID)
	}

	ms.menus = nil

	ms.Logger().Info("服务关闭", "service", "MenuService")
	return nil
}

// CreateContextMenu 创建上下文菜单
func (ms *MenuService) CreateContextMenu(definition MenuDefinition) *connection.QueryResult {
	if definition.MenuID == "" {
		ms.Logger().Error("菜单 ID 不能为空",
			"label", definition.Label,
			"window", definition.Window)
		return &connection.QueryResult{
			Success: false,
			Message: "菜单 ID 不能为空",
		}
	}

	// 检查是否可以复用现有菜单
	if _, existed := ms.menus[definition.MenuID]; existed {
		return &connection.QueryResult{
			Success: false,
			Message: "菜单已存在",
			Data: map[string]interface{}{
				"menuId": definition.MenuID,
			},
		}
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	// 创建 Wails 菜单
	contextMenu := application.NewContextMenu(definition.MenuID)

	// 构建菜单项
	if err := ms.buildMenuItems(contextMenu, definition.Items, definition.ContextData, definition.MenuID); err != nil {
		contextMenu.Destroy()
		ms.Logger().Error("构建菜单失败",
			"menuId", definition.MenuID,
			"label", definition.Label,
			"error", err)
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("构建菜单失败: %v", err),
		}
	}

	// 更新菜单
	contextMenu.Update()

	// 缓存菜单
	wrapper := &MenuWrapper{
		menuID:  definition.MenuID,
		context: contextMenu,
		metadata: MenuMetadata{
			Label:       definition.Label,
			Window:      definition.Window,
			ContextData: definition.ContextData,
		},
		createdAt: time.Now().Unix(),
	}
	ms.menus[definition.MenuID] = wrapper

	ms.Logger().Info("菜单创建成功",
		"menuId", definition.MenuID,
		"label", definition.Label,
		"window", definition.Window)

	return &connection.QueryResult{
		Success: true,
		Message: "菜单创建成功",
		Data: map[string]interface{}{
			"menuId": definition.MenuID,
		},
	}
}

// buildMenuItems 递归构建菜单项
func (ms *MenuService) buildMenuItems(
	menu *application.ContextMenu,
	items []MenuItemDefinition,
	contextData map[string]interface{},
	menuID string,
) error {
	return ms.buildMenuItemsInternal(menu, items, contextData, menuID)
}

// buildMenuItemsInternal 递归构建菜单项（内部实现）
func (ms *MenuService) buildMenuItemsInternal(
	menu interface {
		Add(label string) *application.MenuItem
		AddCheckbox(label string, enabled bool) *application.MenuItem
		AddRadio(label string, enabled bool) *application.MenuItem
		AddSeparator()
		AddSubmenu(label string) *application.Menu
	},
	items []MenuItemDefinition,
	contextData map[string]interface{},
	menuID string,
) error {
	for _, itemDef := range items {
		switch itemDef.Type {
		case MenuItemTypeSeparator:
			menu.AddSeparator()

		case MenuItemTypeItem:
			item := menu.Add(itemDef.Label)
			if itemDef.Shortcut != nil && *itemDef.Shortcut != "" {
				item.SetAccelerator(*itemDef.Shortcut)
			}
			if itemDef.Enabled != nil && !*itemDef.Enabled {
				item.SetEnabled(false)
			}
			// 设置事件发送处理器
			item.OnClick(func(ctx *application.Context) {
				ms.sendMenuEvent(menuID, itemDef, contextData)
			})

		case MenuItemTypeCheckbox:
			// 注意：AddCheckbox 的第二个参数是 enabled，不是 checked
			checked := false
			if itemDef.Checked != nil {
				checked = *itemDef.Checked
			}
			item := menu.AddCheckbox(itemDef.Label, checked)
			if itemDef.Shortcut != nil && *itemDef.Shortcut != "" {
				item.SetAccelerator(*itemDef.Shortcut)
			}
			if itemDef.Enabled != nil && !*itemDef.Enabled {
				item.SetEnabled(false)
			}

			item.OnClick(func(ctx *application.Context) {
				// 切换选中状态
				checked = !checked
				itemDef.Checked = &checked
				item.SetChecked(checked)
				ms.sendMenuEvent(menuID, itemDef, contextData)
			})

		case MenuItemTypeRadio:
			// 注意：AddRadio 的第二个参数是 enabled，不是 checked
			checked := false
			if itemDef.Checked != nil {
				checked = *itemDef.Checked
			}
			item := menu.AddRadio(itemDef.Label, checked)
			if itemDef.Shortcut != nil && *itemDef.Shortcut != "" {
				item.SetAccelerator(*itemDef.Shortcut)
			}
			if itemDef.Enabled != nil && !*itemDef.Enabled {
				item.SetEnabled(false)
			}

			item.OnClick(func(ctx *application.Context) {
				// 切换选中状态
				checked = !checked
				itemDef.Checked = &checked
				item.SetChecked(checked)
				ms.sendMenuEvent(menuID, itemDef, contextData)
			})

		case MenuItemTypeSubmenu:
			submenu := menu.AddSubmenu(itemDef.Label)
			if err := ms.buildMenuItemsInternal(submenu, itemDef.Items, contextData, menuID); err != nil {
				return err
			}

		default:
			err := fmt.Errorf("未知的菜单项类型: %s", itemDef.Type)
			ms.Logger().Error("未知的菜单项类型",
				"menuId", menuID,
				"itemType", itemDef.Type,
				"label", itemDef.Label,
				"error", err)
			return err
		}
	}
	return nil
}

// UpdateMenu 更新菜单
func (ms *MenuService) UpdateMenu(request MenuUpdateRequest) *connection.QueryResult {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	wrapper, exists := ms.menus[request.MenuID]
	if !exists {
		ms.Logger().Error("菜单不存在",
			"menuId", request.MenuID)
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("菜单不存在: %s", request.MenuID),
		}
	}

	// 更新上下文数据
	if request.ContextData != nil {
		wrapper.metadata.ContextData = request.ContextData
	}

	// 清空旧菜单项
	wrapper.context.Clear()

	// 重建菜单项
	if err := ms.buildMenuItems(wrapper.context, request.Items, request.ContextData, request.MenuID); err != nil {
		ms.Logger().Error("重建菜单失败",
			"menuId", request.MenuID,
			"error", err)
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("重建菜单失败: %v", err),
		}
	}

	// 更新菜单
	wrapper.context.Update()

	ms.Logger().Info("菜单更新成功", "menuId", request.MenuID)

	return &connection.QueryResult{
		Success: true,
		Message: "菜单更新成功",
	}
}

// UnregisterContextMenu 注销菜单
func (ms *MenuService) UnregisterContextMenu(menuID string) *connection.QueryResult {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	wrapper, exists := ms.menus[menuID]
	if !exists {
		ms.Logger().Error("菜单不存在",
			"menuId", menuID)
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("菜单不存在: %s", menuID),
		}
	}

	wrapper.context.Destroy()
	delete(ms.menus, menuID)

	ms.Logger().Info("菜单已注销", "menuId", menuID)

	return &connection.QueryResult{
		Success: true,
		Message: "菜单已注销",
	}
}

// GetMenuList 获取所有已注册的菜单列表
func (ms *MenuService) GetMenuList() *connection.QueryResult {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	menus := make([]map[string]interface{}, 0, len(ms.menus))
	for menuID, wrapper := range ms.menus {
		menuInfo := map[string]interface{}{
			"menuId":    menuID,
			"label":     wrapper.metadata.Label,
			"window":    wrapper.metadata.Window,
			"createdAt": wrapper.createdAt,
		}
		menus = append(menus, menuInfo)
	}

	return &connection.QueryResult{
		Success: true,
		Message: "获取菜单列表成功",
		Data:    menus,
	}
}

// sendMenuEvent 发送菜单点击事件到前端
func (ms *MenuService) sendMenuEvent(
	menuID string,
	itemDef MenuItemDefinition,
	contextData map[string]interface{},
) {
	// 从菜单元数据获取窗口名称
	windowName := ""
	ms.mu.RLock()
	if wrapper, exists := ms.menus[menuID]; exists {
		windowName = wrapper.metadata.Window
	}
	ms.mu.RUnlock()

	// 构建事件
	event := MenuClickEvent{
		MenuID:      menuID,
		ItemID:      itemDef.ID,
		Type:        itemDef.Type,
		Label:       itemDef.Label,
		Checked:     *itemDef.Checked,
		ContextData: contextData,
		ItemData:    itemDef.ContextData,
		Timestamp:   time.Now().Unix(),
		Window:      windowName,
	}

	// 发送事件
	ms.App().Event.Emit("menu:clicked", event)
	ms.Logger().Debug("菜单点击事件已发送",
		"menuId", menuID,
		"itemId", itemDef.ID)
}
