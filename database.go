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

import "fmt"

type ColumnDefinition struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Nullable string  `json:"nullable"` // "YES" or "NO"
	Key      string  `json:"key"`      // "PRI" for primary key, "MUL" for foreign key, "UNI" for others
	Default  *string `json:"default"`
	Extra    string  `json:"extra"` // auto_increment
	Comment  string  `json:"comment"`
}

type IndexDefinition struct {
	Name       string `json:"name"`
	ColumnName string `json:"columnName"`
	NonUnique  int    `json:"nonUnique"`
	SeqInIndex int    `json:"seqInIndex"`
	IndexType  string `json:"indexType"`
}

type ForeignKeyDefinition struct {
	Name          string `json:"name"`
	ColumnName    string `json:"columnName"`
	RefTableName  string `json:"refTableName"`
	RefColumnName string `json:"refColumnName"`
	ConstrainName string `json:"constrainName"`
}

type TriggerDefinition struct {
	Name      string `json:"name"`
	Timing    string `json:"timing"` // BEFORE or AFTER
	Event     string `json:"event"`  // INSERT, UPDATE, DELETE
	Statement string `json:"statement"`
}

type ColumnDefinitionWithTable struct {
	TableName string `json:"tableName"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}

type Database interface {
	Connect(config *ConnectionConfig) error
	Close() error
	Ping() error
	Query(query string) ([]map[string]interface{}, []string, error)
	Exec(query string) (int64, error)
	GetDatabases() ([]string, error)
	GetTables(dbName string) ([]string, error)
	GetCreateStatement(dbName, tableName string) (string, error)
	GetColumns(dbName, tableName string) ([]ColumnDefinition, error)
	GetAllColumns(dbName string) ([]ColumnDefinitionWithTable, error)
	GetIndexes(dbName, tableName string) ([]IndexDefinition, error)
	GetForeignKeys(dbName, tableName string) ([]ForeignKeyDefinition, error)
	GetTriggers(dbName, tableName string) ([]TriggerDefinition, error)
}

type BatchApplier interface {
	ApplyChanges(tableName string, changes *ChangeSet) error
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
