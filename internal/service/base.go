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
	"log/slog"
	"reflect"
	"sync"

	"github.com/chenyang-zz/boxify/internal/window"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// BaseService 所有服务的基础结构体
type BaseService struct {
	app        *application.App
	ctx        context.Context
	logger     *slog.Logger
	mu         sync.RWMutex
	appManager *window.AppManager
	registry   *window.WindowRegistry
}

// NewBaseService 使用依赖注入创建基础服务
func NewBaseService(deps *ServiceDeps) BaseService {
	return BaseService{
		app:        deps.app,
		ctx:        context.Background(),
		logger:     deps.app.Logger,
		appManager: deps.appManager,
		registry:   deps.registry,
	}
}

// NewBaseServiceSimple 简化创建方式
func NewBaseServiceSimple(app *application.App) BaseService {
	return BaseService{
		app:    app,
		ctx:    context.Background(),
		logger: app.Logger,
	}
}

// App 获取应用实例
func (b *BaseService) App() *application.App {
	return b.app
}

// Context 获取上下文
func (b *BaseService) Context() context.Context {
	return b.ctx
}

// SetContext 设置上下文
func (b *BaseService) SetContext(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ctx = ctx
}

// Logger 获取日志记录器
func (b *BaseService) Logger() *slog.Logger {
	return b.logger
}

// AppManager 获取窗口管理器
func (b *BaseService) AppManager() *window.AppManager {
	return b.appManager
}

// SetAppManager 设置窗口管理器
func (b *BaseService) SetAppManager(am *window.AppManager) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.appManager = am
}

// Registry 获取窗口注册表
func (b *BaseService) Registry() *window.WindowRegistry {
	return b.registry
}

// SetRegistry 设置窗口注册表
func (b *BaseService) SetRegistry(registry *window.WindowRegistry) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.registry = registry
}

// DefaultServiceStartup 默认启动实现
func (b *BaseService) DefaultServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	b.SetContext(ctx)
	b.Logger().Info("服务启动", "service", getServiceName(b))
	return nil
}

// DefaultServiceShutdown 默认关闭实现
func (b *BaseService) DefaultServiceShutdown() error {
	b.Logger().Info("服务关闭", "service", getServiceName(b))
	return nil
}

// getServiceName 获取服务名称（用于日志）
func getServiceName(service any) string {
	t := reflect.TypeOf(service)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
