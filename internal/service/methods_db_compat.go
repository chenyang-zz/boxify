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

import "github.com/chenyang-zz/boxify/internal/connection"

// MySQLConnect 兼容历史 MySQL 入口，委托通用连接逻辑。
func (a *DatabaseService) MySQLConnect(config *connection.ConnectionConfig) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBConnect(config)
}

// MySQLQuery 兼容历史 MySQL 查询入口，委托通用查询逻辑。
func (a *DatabaseService) MySQLQuery(config *connection.ConnectionConfig, dbName, query string, args []any) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBQuery(config, dbName, query, args)
}

// MySQLGetDatabases 兼容历史 MySQL 库列表入口。
func (a *DatabaseService) MySQLGetDatabases(config *connection.ConnectionConfig) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBGetDatabases(config)
}

// MySQLGetTables 兼容历史 MySQL 表列表入口。
func (a *DatabaseService) MySQLGetTables(config *connection.ConnectionConfig, dbName string) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBGetTables(config, dbName)
}

// MySQLShowCreateTable 兼容历史 MySQL 建表语句入口。
func (a *DatabaseService) MySQLShowCreateTable(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBShowCreateTable(config, dbName, tableName)
}
