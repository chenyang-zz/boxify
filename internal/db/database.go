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

package db

import (
	"fmt"

	"github.com/chenyang-zz/boxify/internal/connection"
)

// Database 定义数据库驱动需要实现的统一能力。
type Database interface {
	Connect(config *connection.ConnectionConfig) error
	Close() error
	Ping() error
	Query(query string, args ...any) ([]map[string]interface{}, []string, error)
	Exec(query string, args ...any) (int64, error)
	GetDatabases() ([]string, error)
	GetTables(dbName string) ([]string, error)
	GetCreateStatement(dbName, tableName string) (string, error)
	GetColumns(dbName, tableName string) ([]*connection.ColumnDefinition, error)
	GetAllColumns(dbName string) ([]*connection.ColumnDefinitionWithTable, error)
	GetIndexes(dbName, tableName string) ([]*connection.IndexDefinition, error)
	GetForeignKeys(dbName, tableName string) ([]*connection.ForeignKeyDefinition, error)
	GetTriggers(dbName, tableName string) ([]*connection.TriggerDefinition, error)
}

// BatchApplier 定义批量数据变更能力。
type BatchApplier interface {
	ApplyChanges(tableName string, changes *connection.ChangeSet) error
}

// DatabaseFactory 负责根据数据库类型创建驱动实例。
type DatabaseFactory struct{}

// NewDatabaseFactory 创建数据库工厂。
func NewDatabaseFactory() *DatabaseFactory {
	return &DatabaseFactory{}
}

// Create 根据数据库类型创建对应驱动实例。
func (f *DatabaseFactory) Create(dbType connection.ConnectionType) (Database, error) {
	switch dbType {
	case connection.ConnectionTypeMySQL:
		return &MySQLDB{}, nil
	case connection.ConnectionTypePostgreSQL:
		return nil, fmt.Errorf("暂不支持的数据库类型: %s", dbType)
	case connection.ConnectionTypeSQLite:
		return nil, fmt.Errorf("暂不支持的数据库类型: %s", dbType)
	default:
		// Default to MySQL for backward compatibility if empty
		if dbType == "" {
			return &MySQLDB{}, nil
		}
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
}

// NewDatabase 是兼容历史调用的工厂入口。
func NewDatabase(dbType connection.ConnectionType) (Database, error) {
	return NewDatabaseFactory().Create(dbType)
}
