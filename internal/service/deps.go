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
	"github.com/chenyang-zz/boxify/internal/window"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// ServiceDeps 服务依赖容器
type ServiceDeps struct {
	app        *application.App
	appManager *window.AppManager
	registry   *window.WindowRegistry
}

// NewServiceDeps 创建依赖容器
func NewServiceDeps(app *application.App, am *window.AppManager) *ServiceDeps {
	deps := &ServiceDeps{
		app:        app,
		appManager: am,
	}
	if am != nil {
		deps.registry = am.GetRegistry()
	}
	return deps
}

// App 获取应用实例
func (d *ServiceDeps) App() *application.App {
	return d.app
}

// AppManager 获取窗口管理器
func (d *ServiceDeps) AppManager() *window.AppManager {
	return d.appManager
}

// Registry 获取窗口注册表
func (d *ServiceDeps) Registry() *window.WindowRegistry {
	return d.registry
}
