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

package connection

// ConnectionType 数据库连接类型
type ConnectionType string

const (
	ConnectionTypeMySQL      ConnectionType = "mysql"      // MySQL 数据库
	ConnectionTypePostgreSQL ConnectionType = "postgresql" // PostgreSQL 数据库
	ConnectionTypeKingbase   ConnectionType = "kingbase"   // Kingbase 数据库
	ConnectionTypeHighGo     ConnectionType = "highgo"     // HighGo 数据库
	ConnectionTypeVastBase   ConnectionType = "vastbase"   // VastBase 数据库
	ConnectionTypeTDengine   ConnectionType = "tdengine"   // TDengine 数据库
	ConnectionTypeMariaDB    ConnectionType = "mariadb"    // MariaDB 数据库
	ConnectionTypeMongoDB    ConnectionType = "mongodb"    // MongoDB 数据库
	ConnectionTypeDameng     ConnectionType = "dameng"     // 达梦数据库
	ConnectionTypeSQLServer  ConnectionType = "sqlserver"  // SQL Server 数据库
	ConnectionTypeSQLite     ConnectionType = "sqlite"     // SQLite 数据库
	ConnectionTypeCustom     ConnectionType = "custom"     // 自定义连接
)

// SSHConfig 是SSH连接的配置结构体
// 包含主机、端口、用户、密码和密钥路径等信息
type SSHConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	KeyPath  string `json:"keyPath"`
}

// ConnectionConfig 是数据库连接的配置结构体
// 包含连接类型、主机、端口、用户、密码、数据库名称以及SSH配置等信息
type ConnectionConfig struct {
	Type     ConnectionType `json:"type"`
	Host     string         `json:"host"`
	Port     int            `json:"port"`
	User     string         `json:"user"`
	Password string         `json:"password"`
	Database string         `json:"database,omitempty"`
	UseSSH   bool           `json:"useSSH"`
	SSH      *SSHConfig     `json:"ssh"`
	Driver   string         `json:"driver,omitempty"`  // 用于自定义连接
	DSN      string         `json:"dsn,omitempty"`     // 用于自定义连接
	Timeout  int            `json:"timeout,omitempty"` // 连接超时时间，单位秒
}

// QueryResult 是查询结果的结构体
// 包含查询是否成功、消息、数据和字段列表等信息
type QueryResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Fields  []string    `json:"fields"`
}

// ColumnDefinition 是数据库列的定义结构体
// 包含列名、类型、是否可空、键类型、默认值、额外信息和注释等信息
type ColumnDefinition struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Nullable string  `json:"nullable"` // "YES" or "NO"
	Key      string  `json:"key"`      // "PRI" for primary key, "MUL" for foreign key, "UNI" for others
	Default  *string `json:"default"`
	Extra    string  `json:"extra"` // auto_increment
	Comment  string  `json:"comment"`
}

// IndexDefinition 是数据库索引的定义结构体
// 包含索引名、列名、是否唯一、在索引中的位置和索引类型等信息
type IndexDefinition struct {
	Name       string `json:"name"`
	ColumnName string `json:"columnName"`
	NonUnique  int    `json:"nonUnique"`
	SeqInIndex int    `json:"seqInIndex"`
	IndexType  string `json:"indexType"`
}

// ForeignKeyDefinition 是数据库外键的定义结构体
// 包含外键名、列名、引用的表名、引用的列名和约束名等信息
type ForeignKeyDefinition struct {
	Name          string `json:"name"`
	ColumnName    string `json:"columnName"`
	RefTableName  string `json:"refTableName"`
	RefColumnName string `json:"refColumnName"`
	ConstrainName string `json:"constrainName"`
}

// TriggerDefinition 是数据库触发器的定义结构体
// 包含触发器名、触发时机、事件类型和触发器语句等信息
type TriggerDefinition struct {
	Name      string `json:"name"`
	Timing    string `json:"timing"` // BEFORE or AFTER
	Event     string `json:"event"`  // INSERT, UPDATE, DELETE
	Statement string `json:"statement"`
}

// ColumnDefinitionWithTable 是包含表名的列定义结构体
// 用于查询整个数据库的列信息时，包含所属表名以区分不同表的同名列
type ColumnDefinitionWithTable struct {
	TableName string `json:"tableName"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}

// UpdateRow 是更新操作中包含主键和值的结构体
// 用于表示一行更新的数据，其中Keys包含主键列和值，Values包含要更新的列和值
type UpdateRow struct {
	Keys   map[string]interface{} `json:"keys"`
	Values map[string]interface{} `json:"values"`
}

// ChangeSet 是一组数据库变更的结构体
// 包含插入、更新和删除的变更数据，用于批量应用变更时传递数据
type ChangeSet struct {
	Inserts []map[string]interface{} `json:"inserts"`
	Updates []UpdateRow              `json:"updates"`
	Deletes []map[string]interface{} `json:"deletes"`
}
