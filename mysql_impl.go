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
	"Boxify/internal/ssh"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLDB struct {
	conn *sql.DB
}

func (m *MySQLDB) getDSN(config *connection.ConnectionConfig) string {
	database := config.Database
	protocol := "tcp"
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// 重用app.go SSH中的SSH逻辑如果全局可用或复制逻辑，则执行
	// 目前假设RegisterSSHNetwork是全局的
	if config.UseSSH {
		netName, err := ssh.RegisterSSHNetwork(config.SSH)
		if err == nil {
			protocol = netName
			address = fmt.Sprintf("%s:%d", config.Host, config.Port)
		}
	}

	return fmt.Sprintf("%s:%s@%s(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.User, config.Password, protocol, address, database)
}

func (m *MySQLDB) Connect(config *connection.ConnectionConfig) error {
	dsn := m.getDSN(config)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	m.conn = db
	return nil
}

func (m *MySQLDB) Close() error {
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}

func (m *MySQLDB) Ping() error {
	if m.conn == nil {
		return fmt.Errorf("连接没有打开")
	}

	ctx, cancel := contextWithTimeout(5 * time.Second)
	defer cancel()

	return m.conn.PingContext(ctx)
}

func (m *MySQLDB) Query(query string) ([]map[string]interface{}, []string, error) {
	if m.conn == nil {
		return nil, nil, fmt.Errorf("连接没有打开")
	}

	rows, err := m.conn.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var resultData []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err = rows.Scan(valuePtrs...); err != nil {
			continue
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		resultData = append(resultData, entry)
	}

	return resultData, columns, nil
}

func (m *MySQLDB) Exec(query string) (int64, error) {
	if m.conn == nil {
		return 0, fmt.Errorf("连接没有打开")
	}

	res, err := m.conn.Exec(query)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (m *MySQLDB) GetDatabases() ([]string, error) {
	data, _, err := m.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}

	var dbs []string
	for _, row := range data {
		if val, ok := row["Database"]; ok {
			dbs = append(dbs, fmt.Sprintf("%v", val))
		} else if val, ok := row["database"]; ok {
			dbs = append(dbs, fmt.Sprintf("%v", val))
		}
	}
	return dbs, nil
}

func (m *MySQLDB) GetTables(dbName string) ([]string, error) {
	// MySQL连接通常绑定到一个数据库，但我们可能需要查询另一个数据库或只是SHOW TABLES
	// 如果当前conn绑定到dbName，没问题。如果不是，SHOW TABLES FROM dbName
	query := "SHOW TABLES"
	if dbName != "" {
		query = fmt.Sprintf("SHOW TABLES FROM `%s`", dbName)
	}

	data, _, err := m.Query(query)
	if err != nil {
		return nil, err
	}

	var tables []string
	for _, row := range data {
		// 列名通常是 Tables_in_dbname。
		for _, val := range row {
			tables = append(tables, fmt.Sprintf("%v", val))
			break
		}
	}

	return tables, nil
}

func (m *MySQLDB) GetCreateStatement(dbName, tableName string) (string, error) {
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", dbName, tableName)
	// 如果dbName已被选中或为空，则只使用表名
	if dbName == "" {
		query = fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)
	}

	data, _, err := m.Query(query)
	if err != nil {
		return "", err
	}

	if len(data) > 0 {
		if val, ok := data[0]["Create Table"]; ok {
			return fmt.Sprintf("%v", val), nil
		}
	}

	return "", fmt.Errorf("未找到创建语句")
}

func (m *MySQLDB) GetColumns(dbName, tableName string) ([]connection.ColumnDefinition, error) {
	query := fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`.`%s`", dbName, tableName)
	if dbName == "" {
		query = fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`", tableName)
	}

	data, _, err := m.Query(query)
	if err != nil {
		return nil, err
	}

	var columns []connection.ColumnDefinition
	for _, row := range data {
		col := connection.ColumnDefinition{
			Name:     fmt.Sprintf("%v", row["Field"]),
			Type:     fmt.Sprintf("%v", row["Type"]),
			Nullable: fmt.Sprintf("%v", row["Null"]),
			Key:      fmt.Sprintf("%v", row["Key"]),
			Extra:    fmt.Sprintf("%v", row["Extra"]),
			Comment:  fmt.Sprintf("%v", row["Comment"]),
		}

		if row["Default"] != nil {
			d := fmt.Sprintf("%v", row["Default"])
			col.Default = &d
		}

		columns = append(columns, col)
	}

	return columns, nil
}

func (m *MySQLDB) GetAllColumns(dbName string) ([]connection.ColumnDefinitionWithTable, error) {
	query := fmt.Sprintf("SELECt TABLE_NAME, COLUMN_NAME, COLUMN_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = `%s`", dbName)
	if dbName == "" {
		// 如果dbName为空，我们可能需要使用connection
		// 但是information_schema通常需要一个模式过滤器，否则它返回所有
		// 假设提供了dbName，或者我们尝试获取它。
		// 对于MVP，如果为空，则返回空或尝试“SELECT DATABASE()”
		return nil, fmt.Errorf("dbName必传")
	}

	data, _, err := m.Query(query)
	if err != nil {
		return nil, err
	}

	var cols []connection.ColumnDefinitionWithTable
	for _, row := range data {
		col := connection.ColumnDefinitionWithTable{
			TableName: fmt.Sprintf("%v", row["TABLE_NAME"]),
			Name:      fmt.Sprintf("%v", row["COLUMN_NAME"]),
			Type:      fmt.Sprintf("%v", row["COLUMN_TYPE"]),
		}
		cols = append(cols, col)
	}

	return cols, nil
}

func (m *MySQLDB) GetIndexes(dbName, tableName string) ([]connection.IndexDefinition, error) {
	query := fmt.Sprintf("SHOW INDEX FROM `%s`.`%s`", dbName, tableName)
	if dbName == "" {
		query = fmt.Sprintf("SHOW INDEX FROM `%s`", tableName)
	}

	data, _, err := m.Query(query)
	if err != nil {
		return nil, err
	}

	var indexs []connection.IndexDefinition
	for _, row := range data {
		// 需要小心处理类型 Non_unique通常是int
		nonUnique := 0
		if val, ok := row["Non_unique"]; ok {
			// 处理各种数字类型（json解码可能是float64）
			if f, ok := val.(float64); ok {
				nonUnique = int(f)
			} else if i, ok := val.(int64); ok {
				nonUnique = int(i)
			}
		}

		seq := 0
		if val, ok := row["Seq_in_index"]; ok {
			if f, ok := val.(float64); ok {
				seq = int(f)
			} else if i, ok := val.(int64); ok {
				seq = int(i)
			}
		}

		idx := connection.IndexDefinition{
			Name:       fmt.Sprintf("%v", row["Key_name"]),
			ColumnName: fmt.Sprintf("%v", row["Column_name"]),
			NonUnique:  nonUnique,
			SeqInIndex: seq,
			IndexType:  fmt.Sprintf("%v", row["Index_type"]),
		}
		indexs = append(indexs, idx)
	}

	return indexs, nil
}

func (m *MySQLDB) GetForeignKeys(dbName, tableName string) ([]connection.ForeignKeyDefinition, error) {
	query := fmt.Sprintf(`SELECT CONSTRAINT_NAME, COLUMN_NAME, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME 
	FROM information_schema.KEY_COLUMN_USAGE 
	WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s' AND REFERENCED_TABLE_NAME IS NOT NULL`, dbName, tableName)

	data, _, err := m.Query(query)
	if err != nil {
		return nil, err
	}

	var fks []connection.ForeignKeyDefinition
	for _, row := range data {
		fk := connection.ForeignKeyDefinition{
			Name:          fmt.Sprintf("%v", row["CONSTRAINT_NAME"]),
			ColumnName:    fmt.Sprintf("%v", row["COLUMN_NAME"]),
			RefTableName:  fmt.Sprintf("%v", row["REFERENCED_TABLE_NAME"]),
			RefColumnName: fmt.Sprintf("%v", row["REFERENCED_COLUMN_NAME"]),
			ConstrainName: fmt.Sprintf("%v", row["CONSTRAINT_NAME"]),
		}
		fks = append(fks, fk)
	}
	return fks, nil
}

func (m *MySQLDB) GetTriggers(dbName, tableName string) ([]connection.TriggerDefinition, error) {
	query := fmt.Sprintf("SHOW TRIGGERS FROM `%s` WHERE `Table` = '%s'", dbName, tableName)
	data, _, err := m.Query(query)
	if err != nil {
		return nil, err
	}

	var triggers []connection.TriggerDefinition
	for _, row := range data {
		trig := connection.TriggerDefinition{
			Name:      fmt.Sprintf("%v", row["Trigger"]),
			Timing:    fmt.Sprintf("%v", row["Timing"]),
			Event:     fmt.Sprintf("%v", row["Event"]),
			Statement: fmt.Sprintf("%v", row["Statement"]),
		}
		triggers = append(triggers, trig)
	}

	return triggers, nil
}

func (m *MySQLDB) ApplyChanges(tableName string, changes *connection.ChangeSet) error {
	if m.conn == nil {
		return fmt.Errorf("连接没有打开")
	}

	tx, err := m.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // 确保在出错时回滚

	// 1， 删除
	for _, pk := range changes.Deletes {
		// 构建DELETE语句
		var wheres []string
		var args []interface{}
		for k, v := range pk {
			wheres = append(wheres, fmt.Sprintf("`%s` = ?"), k)
			args = append(args, v)
		}
		if len(wheres) == 0 {
			continue
		}
		query := fmt.Sprintf("DELETE FROM `%s` WHERE %s", tableName, strings.Join(wheres, " AND "))
		if _, err := tx.Exec(query, args...); err != nil {
			return fmt.Errorf("删除错误：%w", err)
		}
	}

	// 2. 更新
	for _, update := range changes.Updates {
		var sets []string
		var args []interface{}

		for k, v := range update.Values {
			sets = append(sets, fmt.Sprintf("`%s` = ?", k))
			args = append(args, v)
		}

		if len(sets) == 0 {
			continue
		}

		var wheres []string
		for k, v := range update.Keys {
			wheres = append(wheres, fmt.Sprintf("`%s` = ?", k))
			args = append(args, v)
		}

		if len(wheres) == 0 {
			return fmt.Errorf("更新缺少主键条件")
		}

		query := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", tableName, strings.Join(sets, ", "), strings.Join(wheres, " AND "))
		if _, err = tx.Exec(query, args...); err != nil {
			return fmt.Errorf("更新错误：%w", err)
		}
	}

	// 3. 插入
	for _, row := range changes.Inserts {
		var cols []string
		var placeholders []string
		var args []interface{}

		for k, v := range row {
			cols = append(cols, fmt.Sprintf("`%s`", k))
			placeholders = append(placeholders, "?")
			args = append(args, v)
		}

		if len(cols) == 0 {
			continue
		}

		query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", tableName, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
		if _, err := tx.Exec(query, args...); err != nil {
			return fmt.Errorf("插入错误：%w", err)
		}
	}

	return tx.Commit()
}
