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
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/chenyang-zz/boxify/internal/config"
	"github.com/chenyang-zz/boxify/internal/logger"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppManager 管理多个窗口的应用程序
type AppManager struct {
	app        *application.App
	registry   *WindowRegistry
	logger     *slog.Logger
	pageConfig *config.PageConfigFile // 页面配置
	ctx        context.Context        // 应用上下文，包含 buildType
}

func InitApplication(assets fs.FS) *AppManager {
	// 创建临时应用以获取环境信息
	tempApp := application.New(application.Options{
		Name:     "Boxify",
		LogLevel: slog.LevelInfo,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	// 设置应用上下文，包含 buildType
	buildType := "prod"
	if tempApp.Env.Info().Debug {
		buildType = "dev"
	}
	ctx := context.WithValue(context.Background(), "buildType", buildType)

	// 使用正确的上下文初始化 logger
	logger.Init(ctx)
	slogLogger := logger.GetSlogLogger()

	// 使用初始化后的 logger 创建应用
	app := application.New(application.Options{
		Name:     "Boxify",
		LogLevel: slog.LevelInfo,
		Logger:   slogLogger,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	am := &AppManager{
		app:    app,
		logger: slogLogger,
		ctx:    ctx,
	}

	// 创建窗口注册表
	am.registry = NewWindowRegistry(am.app, am.logger)

	// 加载页面配置
	pageConfig, err := config.LoadPageConfig(config.GetPageConfigPath())
	if err != nil {
		am.logger.LogAttrs(context.Background(), slog.LevelWarn,
			"无法加载页面配置，使用默认配置",
			slog.String("error", err.Error()))
		pageConfig = &config.PageConfigFile{Pages: []config.PageConfig{}}
	}
	am.pageConfig = pageConfig

	// 从配置创建主窗口
	am.CreateMainWindowFromConfig()

	// 加载保存的布局
	am.LoadLayout()

	return am
}

func (am *AppManager) RegisterService(registers ...func(app *application.App) application.Service) {
	for _, register := range registers {
		am.app.RegisterService(register(am.app))
	}
}

func (am *AppManager) Run() error {
	return am.app.Run()
}

// CreateMainWindowFromConfig 从配置创建主窗口
func (am *AppManager) CreateMainWindowFromConfig() {
	mainConfig := am.pageConfig.GetMainPageConfig()

	if mainConfig == nil {
		panic("主窗口配置不存在")
	}

	am.registry.Register(mainConfig)
}

// OpenPage 打开页面（统一 API）
func (am *AppManager) OpenPage(pageId string) error {
	pageConfig := am.pageConfig.GetPageConfig(pageId)
	if pageConfig == nil || pageConfig.Window == nil {
		return fmt.Errorf("页面不存在: %s", pageId)
	}

	am.registry.Register(pageConfig)
	return nil
}

// ClosePage 关闭页面
func (am *AppManager) ClosePage(pageId string) error {
	pageConfig := am.pageConfig.GetPageConfig(pageId)
	if pageConfig == nil || pageConfig.Window == nil {
		return fmt.Errorf("页面不存在: %s", pageId)
	}

	windowName := pageConfig.Window.Name
	if window := am.registry.Get(windowName); window != nil {
		window.Hide()
	}

	return nil
}

// GetWindow 根据名称获取窗口
func (am *AppManager) GetWindow(name string) *application.WebviewWindow {
	return am.registry.Get(name)
}

func (am *AppManager) GetWindowID(name string) uint {
	if window := am.registry.Get(name); window != nil {
		return window.ID()
	}
	return 0
}

// GetPageConfig 获取页面配置
func (am *AppManager) GetPageConfig() *config.PageConfigFile {
	return am.pageConfig
}

// GetRegistry 获取窗口注册表引用
func (am *AppManager) GetRegistry() *WindowRegistry {
	return am.registry
}

// App 获取应用实例
func (am *AppManager) App() *application.App {
	return am.app
}

// GetContext 获取应用上下文（包含 buildType）
func (am *AppManager) GetContext() context.Context {
	return am.ctx
}

// generateModalName 生成模态窗口唯一名称
func generateModalName() string {
	return fmt.Sprintf("modal-%d", time.Now().UnixNano())
}

// SaveLayout 保存窗口布局
func (am *AppManager) SaveLayout() {
	layout := make(map[string]WindowState)

	// 使用注册表的方法
	for name := range am.registry.GetAllWindowNames() {
		if window := am.registry.Get(name); window != nil {
			x, y := window.Position()
			width, height := window.Size()

			layout[name] = WindowState{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
			}
		}
	}

	data, _ := json.Marshal(layout)
	os.WriteFile("layout.json", data, 0644)
}

// LoadLayout 加载窗口布局
func (am *AppManager) LoadLayout() {
	data, err := os.ReadFile("layout.json")
	if err != nil {
		return
	}

	var layout map[string]WindowState
	if err := json.Unmarshal(data, &layout); err != nil {
		return
	}

	for name, state := range layout {
		if window := am.registry.Get(name); window != nil {
			window.SetPosition(state.X, state.Y)
			window.SetSize(state.Width, state.Height)
		}
	}
}

type WindowState struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
