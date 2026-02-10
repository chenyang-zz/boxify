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

package app

import (
	"strings"

	"github.com/chenyang-zz/boxify/internal/connection"
)

// normalizeRunConfig 根据连接配置和用户输入的 dbName 生成最终的运行配置
// 对于大多数数据库类型，dbName 被视为要连接的数据库名称
// 并覆盖连接配置中的 Database 字段。
func normalizeRunConfig(config *connection.ConnectionConfig, dbName string) *connection.ConnectionConfig {
	runConfig := *config
	name := strings.TrimSpace(dbName)
	if name == "" {
		return &runConfig
	}

	switch strings.ToLower(strings.TrimSpace(config.Type)) {
	case "mysql", "mariadb", "postgres", "kingbase", "highgo", "vastbase", "sqlserver", "mongodb", "tdengine":
		// 这些类型的 dbName 表示"数据库"，需要写入连接配置以选择目标库
		runConfig.Database = name
	case "dameng":
		// 达梦使用 schema 参数，沿用现有行为：dbName 表示 schema。
		runConfig.Database = name
	default:
		// oracle: dbName 表示 schema/owner，不能覆盖 config.Database（服务名）
		// sqlite: 无需设置 Database
		// custom: 语义不明确，避免污染缓存 key
	}

	return &runConfig
}

// normalizeSchemaAndTable 根据数据库类型和用户输入的 dbName/tableName 规范化 schema 和 table 名称
func normalizeSchemaAndTable(config *connection.ConnectionConfig, dbName string, tableName string) (string, string) {
	rawTable := strings.TrimSpace(tableName)
	rawDB := strings.TrimSpace(dbName)
	if rawTable == "" {
		return rawDB, rawTable
	}

	if parts := strings.SplitN(rawTable, ".", 2); len(parts) == 2 {
		schema := strings.TrimSpace(parts[0])
		table := strings.TrimSpace(parts[1])
		if schema != "" && table != "" {
			return schema, table
		}
	}

	switch strings.ToLower(strings.TrimSpace(config.Type)) {
	case "postgres", "kingbase", "highgo", "vastbase":
		// PG/金仓/瀚高/海量：dbName 在 UI 里是"数据库"，schema 需从 tableName 或使用默认 public。
		return "public", rawTable
	case "sqlserver":
		// SQL Server：dbName 表示数据库，schema 默认 dbo
		return "dbo", rawTable
	default:
		// MySQL：dbName 表示数据库；Oracle/达梦：dbName 表示 schema/owner。
		return rawDB, rawTable
	}
}
