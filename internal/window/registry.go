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

package window

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/chenyang-zz/boxify/internal/config"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// WindowEntry 窗口注册表条目
type WindowEntry struct {
	Window *application.WebviewWindow
	Config *config.PageConfig
}

// WindowRegistry 窗口注册表
type WindowRegistry struct {
	windows map[string]*WindowEntry
	mu      sync.RWMutex
	app     *application.App
	logger  *slog.Logger
}

// NewWindowRegistry 创建窗口注册表
func NewWindowRegistry(app *application.App, logger *slog.Logger) *WindowRegistry {
	return &WindowRegistry{
		windows: make(map[string]*WindowEntry),
		app:     app,
		logger:  logger,
	}
}

// Register 注册窗口并设置生命周期钩子
func (wr *WindowRegistry) Register(config *config.PageConfig) *application.WebviewWindow {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	// 对于单例和主窗口，检查是否已存在
	windowType := ParseWindowType(config.Type)
	if windowType == WindowTypeSingleton || windowType == WindowTypeMain {
		if entry, exists := wr.windows[config.Window.Name]; exists {
			if entry.Window.IsVisible() {
				wr.logger.Info("窗口已存在且可见，聚焦现有窗口", "name", config.Window.Name, "type", windowType)
				entry.Window.Focus()
				return nil
			}
			wr.logger.Info("窗口已存在，聚焦现有窗口", "name", config.Window.Name, "type", windowType)
			entry.Window.Show()
			entry.Window.Focus()

			// 发送窗口打开事件
			wr.emitWindowEvent("window:opened", config)
			return entry.Window
		}
	}

	// 创建窗口
	window := wr.createWindow(config)

	entry := &WindowEntry{
		Window: window,
		Config: config,
	}

	wr.windows[config.Window.Name] = entry

	// 设置生命周期钩子
	wr.setupLifecycleHooks(entry)

	// 发送窗口打开事件
	wr.emitWindowEvent("window:opened", config)

	wr.logger.Info("窗口已注册", "name", config.Window.Name, "type", windowType, "url", config.Window.URL)

	return window
}

// Unregister 注销窗口
func (wr *WindowRegistry) Unregister(name string) {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	if _, exists := wr.windows[name]; exists {
		delete(wr.windows, name)

		wr.logger.Info("窗口已注销", "name", name)
	}
}

// Get 获取窗口
func (wr *WindowRegistry) Get(name string) *application.WebviewWindow {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	if entry, exists := wr.windows[name]; exists {
		return entry.Window
	}
	return nil
}

// setupLifecycleHooks 设置窗口生命周期钩子
func (wr *WindowRegistry) setupLifecycleHooks(entry *WindowEntry) {
	window := entry.Window
	config := entry.Config

	// 窗口关闭事件
	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		wr.logger.Info("窗口关闭事件", "name", config.Window.Name, "type", config.Type)

		switch ParseWindowType(config.Type) {
		case WindowTypeMain, WindowTypeSingleton:
			// 主窗口和单例窗口：隐藏而非关闭
			window.Hide()
			event.Cancel()

		case WindowTypeModal:
			// 模态窗口：重新启用父窗口并允许关闭
			if config.Parent != "" {
				if parentWindow := wr.Get(config.Parent); parentWindow != nil {
					parentWindow.SetEnabled(true)
				}
			}
			// 注销窗口
			wr.Unregister(config.Window.Name)
			// 不调用 Cancel()，允许关闭
		}

		wr.emitWindowEvent("window:closed", entry.Config)
	})
}

// createWindow 创建窗口
func (wr *WindowRegistry) createWindow(config *config.PageConfig) *application.WebviewWindow {
	opts := *config.Window
	if opts.Title == "" {
		opts.Title = config.Title
	}

	if config.IsMain && config.Window.URL == "" {
		opts.URL = "/"
	}

	if config.Window.URL == "" {
		opts.URL = fmt.Sprintf("/%s", config.ID)
	}

	opts.URL = config.Window.URL

	window := wr.app.Window.NewWithOptions(opts)

	// 居中
	if config.Center {
		window.Center()
	}

	return window
}

// GetAllWindowNames 获取所有窗口名称
func (wr *WindowRegistry) GetAllWindowNames() map[string]bool {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	names := make(map[string]bool)
	for name := range wr.windows {
		names[name] = true
	}
	return names
}

// WindowInfo 窗口信息
type WindowInfo struct {
	Name  string
	Type  string
	Title string
	ID    uint
}

// GetWindowInfo 获取窗口详细信息
func (wr *WindowRegistry) GetWindowInfo(name string) *WindowInfo {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	entry, exists := wr.windows[name]
	if !exists {
		return nil
	}

	return &WindowInfo{
		Name:  name,
		Type:  entry.Config.Type,
		Title: entry.Config.Title,
		ID:    entry.Window.ID(),
	}
}

// GetAllWindowInfos 获取所有窗口信息
func (wr *WindowRegistry) GetAllWindowInfos() []*WindowInfo {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	infos := make([]*WindowInfo, 0, len(wr.windows))
	for name, entry := range wr.windows {
		infos = append(infos, &WindowInfo{
			Name:  name,
			Type:  entry.Config.Type,
			Title: entry.Config.Title,
			ID:    entry.Window.ID(),
		})
	}
	return infos
}

// emitWindowEvent 发送窗口事件
func (wr *WindowRegistry) emitWindowEvent(eventType string, config *config.PageConfig) {
	wr.app.Event.Emit(eventType, map[string]interface{}{
		"name":  config.Window.Name,
		"type":  config.Type,
		"title": config.Title,
	})
}
