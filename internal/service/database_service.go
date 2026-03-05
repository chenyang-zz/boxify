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

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/chenyang-zz/boxify/internal/db"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// DatabaseService 负责前端服务编排，连接管理由 db.ConnectionManager 承担。
type DatabaseService struct {
	BaseService
	manager *db.ConnectionManager
}

// NewDatabaseService 创建 DatabaseService（使用依赖注入）。
func NewDatabaseService(deps *ServiceDeps) *DatabaseService {
	return &DatabaseService{
		BaseService: NewBaseService(deps),
		manager:     db.NewConnectionManager(deps.app.Logger),
	}
}

// ServiceStartup 在应用启动时初始化数据库服务状态。
func (a *DatabaseService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	a.SetContext(ctx)
	if a.manager == nil {
		a.manager = db.NewConnectionManager(a.Logger())
	}
	a.Logger().Info("服务启动", "service", "DatabaseService")
	return nil
}

// ServiceShutdown 在应用关闭时释放数据库连接资源。
func (a *DatabaseService) ServiceShutdown() error {
	a.Logger().Info("服务开始关闭，准备释放资源", "service", "DatabaseService")
	if a.manager != nil {
		if err := a.manager.CloseAll(); err != nil {
			a.Logger().Error("关闭数据库连接失败", "error", err)
		}
	}
	a.Logger().Info("服务关闭", "service", "DatabaseService")
	return nil
}

// getDatabaseForcePing 强制探活后返回数据库连接。
func (a *DatabaseService) getDatabaseForcePing(config *connection.ConnectionConfig) (db.Database, error) {
	return a.getDatabaseWithPing(config, true)
}

// getDatabase 返回可用数据库连接。
func (a *DatabaseService) getDatabase(config *connection.ConnectionConfig) (db.Database, error) {
	return a.getDatabaseWithPing(config, false)
}

// getDatabaseWithPing 按需探活并返回数据库连接。
func (a *DatabaseService) getDatabaseWithPing(config *connection.ConnectionConfig, forcePing bool) (db.Database, error) {
	if a.manager == nil {
		a.manager = db.NewConnectionManager(a.Logger())
	}
	return a.manager.Get(config, forcePing)
}
