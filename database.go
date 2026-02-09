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

package main

import (
	"Boxify/internal/connection"
	"fmt"
)

type Database interface {
	Connect(config *connection.ConnectionConfig) error
	Close() error
	Ping() error
	Query(query string) ([]map[string]interface{}, []string, error)
	Exec(query string) (int64, error)
	GetDatabases() ([]string, error)
	GetTables(dbName string) ([]string, error)
	GetCreateStatement(dbName, tableName string) (string, error)
	GetColumns(dbName, tableName string) ([]connection.ColumnDefinition, error)
	GetAllColumns(dbName string) ([]connection.ColumnDefinitionWithTable, error)
	GetIndexes(dbName, tableName string) ([]connection.IndexDefinition, error)
	GetForeignKeys(dbName, tableName string) ([]connection.ForeignKeyDefinition, error)
	GetTriggers(dbName, tableName string) ([]connection.TriggerDefinition, error)
}

type BatchApplier interface {
	ApplyChanges(tableName string, changes *connection.ChangeSet) error
}

// Factory
func NewDatabase(dbType string) (Database, error) {
	switch dbType {
	case "mysql":
		return nil, nil
	case "postgres":
		return nil, nil
	case "sqlite":
		return nil, nil
	default:
		// Default to MySQL for backward compatibility if empty
		if dbType == "" {
			return nil, nil
		}
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
}
