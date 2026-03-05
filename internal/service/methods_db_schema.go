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
	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/chenyang-zz/boxify/internal/db"
)

// DBGetDatabases 获取数据库列表。
func (a *DatabaseService) DBGetDatabases(config *connection.ConnectionConfig) *connection.QueryResult {
	dbInst, err := a.getDatabase(config)
	if err != nil {
		a.Logger().Error("DBGetDatabases 获取连接失败", "error", err, "summary", db.FormatConnSummary(config))
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	dbs, err := dbInst.GetDatabases()
	if err != nil {
		a.Logger().Error("DBGetDatabases 获取数据库列表失败", "error", err, "summary", db.FormatConnSummary(config))
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	var resData []map[string]string
	for _, name := range dbs {
		resData = append(resData, map[string]string{"Database": name})
	}

	return &connection.QueryResult{Success: true, Message: "获取数据库列表成功", Data: resData}
}

// DBGetTables 获取表列表。
func (a *DatabaseService) DBGetTables(config *connection.ConnectionConfig, dbName string) *connection.QueryResult {
	runConfig := normalizeRunConfig(config, dbName)

	dbInst, err := a.getDatabase(runConfig)
	if err != nil {
		a.Logger().Error("DBGetTables 获取连接失败", "error", err, "summary", db.FormatConnSummary(runConfig))
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	tables, err := dbInst.GetTables(dbName)
	if err != nil {
		a.Logger().Error("DBGetTables 获取表列表失败", "error", err, "summary", db.FormatConnSummary(runConfig))
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	var resData []map[string]string
	for _, name := range tables {
		resData = append(resData, map[string]string{"Table": name})
	}

	return &connection.QueryResult{Success: true, Message: "获取表列表成功", Data: resData}
}

// DBShowCreateTable 获取建表语句。
func (a *DatabaseService) DBShowCreateTable(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	dbInst, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	sqlStr, err := dbInst.GetCreateStatement(dbName, tableName)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	return &connection.QueryResult{Success: true, Message: "获取建表语句成功", Data: sqlStr}
}

// DBGetColumns 获取列信息。
func (a *DatabaseService) DBGetColumns(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
	runConfig := normalizeRunConfig(config, dbName)

	dbInst, err := a.getDatabase(runConfig)
	if err != nil {
		a.Logger().Error("DBGetColumns 获取连接失败", "error", err, "summary", db.FormatConnSummary(runConfig))
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	schemaName, pureTableName := normalizeSchemaAndTable(config, dbName, tableName)
	columns, err := dbInst.GetColumns(schemaName, pureTableName)
	if err != nil {
		a.Logger().Error("DBGetColumns 获取列信息失败", "error", err, "summary", db.FormatConnSummary(runConfig), "schema", schemaName, "table", pureTableName)
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	return &connection.QueryResult{Success: true, Message: "获取列信息成功", Data: columns}
}

// DBGetIndexes 获取索引信息。
func (a *DatabaseService) DBGetIndexes(config *connection.ConnectionConfig, dbName string, tableName string) *connection.QueryResult {
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	dbInst, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	indexes, err := dbInst.GetIndexes(dbName, tableName)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	return &connection.QueryResult{Success: true, Message: "获取索引信息成功", Data: indexes}
}

// DBGetForeignKeys 获取外键信息。
func (a *DatabaseService) DBGetForeignKeys(config *connection.ConnectionConfig, dbName string, tableName string) *connection.QueryResult {
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	dbInst, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	fks, err := dbInst.GetForeignKeys(dbName, tableName)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	return &connection.QueryResult{Success: true, Message: "获取外键信息成功", Data: fks}
}

// DBGetTriggers 获取触发器信息。
func (a *DatabaseService) DBGetTriggers(config *connection.ConnectionConfig, dbName string, tableName string) *connection.QueryResult {
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	dbInst, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	triggers, err := dbInst.GetTriggers(dbName, tableName)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	return &connection.QueryResult{Success: true, Message: "获取触发器信息成功", Data: triggers}
}

// DBGetAllColumns 获取所有列信息（包含系统表）。
func (a *DatabaseService) DBGetAllColumns(config *connection.ConnectionConfig, dbName string) *connection.QueryResult {
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	dbInst, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	columns, err := dbInst.GetAllColumns(dbName)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	return &connection.QueryResult{Success: true, Message: "获取所有列信息成功", Data: columns}
}
