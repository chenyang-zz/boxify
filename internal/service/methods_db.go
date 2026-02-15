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
	"fmt"
	"strings"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/chenyang-zz/boxify/internal/logger"
)

// 通用数据库方法

// DBConnect 连接数据库，成功则返回成功消息，失败则返回错误信息
func (a *DatabaseService) DBConnect(config *connection.ConnectionConfig) *connection.QueryResult {
	// 连接测试需要强制 ping，避免缓存命中但连接已失效时误判成功
	_, err := a.getDatabaseForcePing(config)
	if err != nil {
		logger.Error("DBConnect 连接失败：%s, err: %w", formatConnSummary(config), err)
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}
	logger.Info("DBConnect 连接成功：%s", formatConnSummary(config))

	return &connection.QueryResult{
		Success: true,
		Message: "连接成功",
	}
}

// TestConnection 测试数据库连接，成功则返回成功消息，失败则返回错误信息
func (a *DatabaseService) TestConnection(config *connection.ConnectionConfig) *connection.QueryResult {
	_, err := a.getDatabaseForcePing(config)
	if err != nil {
		logger.Error("TestConnection 连接失败：%s, err: %w", formatConnSummary(config), err)
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	logger.Info("TestConnection 连接成功：%s", formatConnSummary(config))
	return &connection.QueryResult{
		Success: true,
		Message: "连接成功",
	}
}

// CreateDatabase 创建一个新的数据库
func (a *DatabaseService) CreateDatabase(config *connection.ConnectionConfig, dbName string) *connection.QueryResult {
	runConfig := *config
	runConfig.Database = ""

	dbInst, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	escapedDbName := strings.ReplaceAll(dbName, "`", "``") // MySQL中使用反引号包裹数据库名，并对其中的反引号进行转义
	query := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", escapedDbName)
	dbType := strings.ToLower(strings.TrimSpace(config.Type))
	if dbType == "postgres" || dbType == "kingbase" || dbType == "highgo" || dbType == "vastbase" {
		escapedDbName = strings.ReplaceAll(dbName, `"`, `""`) // PostgreSQL及其衍生数据库中使用双引号包裹数据库名，并对其中的双引号进行转义
		query = fmt.Sprintf("CREATE DATABASE \"%s\"", escapedDbName)
	} else if dbType == "tdengine" {
		query = fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", quoteIdentByType(dbType, dbName))
	} else if dbType == "mariadb" {
		// MariaDB 支持 MYSQL 语法
	}

	_, err = dbInst.Exec(query)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "数据库创建成功",
	}
}
