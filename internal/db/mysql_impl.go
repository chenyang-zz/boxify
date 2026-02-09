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
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/chenyang-zz/boxify/internal/logger"
	"github.com/chenyang-zz/boxify/internal/ssh"
	"github.com/chenyang-zz/boxify/internal/utils"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLDB Database接口的MySQL实现
type MySQLDB struct {
	conn        *sql.DB
	pintTimeout time.Duration // 可配置的Ping超时
}

// getDSN 构建MySQL连接字符串，考虑SSH隧道
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
		} else {
			logger.Warnf("注册 SSH 网络失败，将尝试直连：地址=%s:%d 用户=%s，原因：%v", config.Host, config.Port, config.User, err)
		}
	}

	// 获取连接超时时间
	timeout := getConnectTimeoutSeconds(config)

	return fmt.Sprintf("%s:%s@%s(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%ds", config.User, config.Password, protocol, address, database, timeout)
}

// Connect建立数据库连接
func (m *MySQLDB) Connect(config *connection.ConnectionConfig) error {
	dsn := m.getDSN(config)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败：%w", err)
	}

	m.conn = db
	m.pintTimeout = getConnectTimeout(config)

	// 尝试Ping以验证连接
	if err := m.Ping(); err != nil {
		return fmt.Errorf("连接建立后验证失败：%w", err)
	}

	return nil
}

// Close关闭数据库连接
func (m *MySQLDB) Close() error {
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}

// Ping验证数据库连接是否可用
func (m *MySQLDB) Ping() error {
	if m.conn == nil {
		return fmt.Errorf("连接没有打开")
	}

	timeout := m.pintTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := utils.ContextWithTimeout(timeout)
	defer cancel()

	return m.conn.PingContext(ctx)
}

// QueryContext 执行带有上下文的查询并返回结果
func (m *MySQLDB) QueryContext(ctx context.Context, query string) ([]map[string]interface{}, []string, error) {
	if m.conn == nil {
		return nil, nil, fmt.Errorf("连接没有打开")
	}

	rows, err := m.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	return scanRows(rows)
}

// Query 执行查询并返回结果
func (m *MySQLDB) Query(query string) ([]map[string]interface{}, []string, error) {
	if m.conn == nil {
		return nil, nil, fmt.Errorf("连接没有打开")
	}

	rows, err := m.conn.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	return scanRows(rows)
}

// ExecContext 执行带有上下文的命令并返回受影响的行数
func (m *MySQLDB) ExecContext(ctx context.Context, query string) (int64, error) {
	if m.conn == nil {
		return 0, fmt.Errorf("连接没有打开")
	}

	res, err := m.conn.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

// Exec 执行命令并返回受影响的行数
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

// GetDatabases 返回数据库列表
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

// GetTables 返回指定数据库的表列表，如果dbName为空，则返回当前连接数据库的表
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

// GetCreateStatement 返回指定表的创建语句
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

// GetColumns 返回指定表的列定义
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

// GetAllColumns 返回指定数据库的所有列定义
// 包含表名以区分不同表的同名列
func (m *MySQLDB) GetAllColumns(dbName string) ([]connection.ColumnDefinitionWithTable, error) {
	query := fmt.Sprintf("SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '%s'", dbName)
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

// GetIndexes 返回指定表的索引定义
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

// GetForeignKeys 返回指定表的外键定义
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

// GetTriggers 返回指定表的触发器定义
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

// ApplyChanges 根据提供的ChangeSet对指定表应用批量更改（插入、更新、删除）
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
			wheres = append(wheres, fmt.Sprintf("`%s` = ?", k))
			args = append(args, v)
		}
		if len(wheres) == 0 {
			continue
		}
		query := fmt.Sprintf("DELETE FROM `%s` WHERE %s", tableName, strings.Join(wheres, " AND "))
		res, err := tx.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("删除错误：%w", err)
		}
		if affected, err := res.RowsAffected(); err == nil && affected == 0 {
			return fmt.Errorf("删除未生效：未匹配到任何行")
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
		res, err := tx.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("更新错误：%w", err)
		}
		if affected, err := res.RowsAffected(); err == nil && affected == 0 {
			return fmt.Errorf("更新未生效：未匹配到任何行")
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
		res, err := tx.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("插入错误：%w", err)
		}
		if affected, err := res.RowsAffected(); err == nil && affected == 0 {
			return fmt.Errorf("插入未生效：未插入任何行")
		}
	}

	return tx.Commit()
}

// normalizeMySQLDateTimeValue 处理MySQL可能返回的日期时间字符串，修复常见格式问题并尝试解析为标准格式
func normalizeMySQLDateTimeValue(value interface{}) interface{} {
	text, ok := value.(string)
	if !ok {
		return value
	}

	raw := strings.TrimSpace(text)
	if raw == "" {
		return value
	}

	cleaned := strings.ReplaceAll(raw, "+ ", "+")    // 修复MySQL在某些环境下输出的日期时间字符串中可能出现的空格问题
	cleaned = strings.ReplaceAll(cleaned, "- ", "-") // 修复MySQL在某些环境下输出的日期时间字符串中可能出现的空格问题

	if len(cleaned) >= 19 && cleaned[10] == 'T' {
		if strings.HasSuffix(cleaned, "Z") || hasTimezoneOffset(cleaned) {
			if t, err := time.Parse(time.RFC3339Nano, cleaned); err == nil {
				return formatMySQLDateTime(t)
			}
			if t, err := time.Parse(time.RFC3339, cleaned); err == nil {
				return formatMySQLDateTime(t)
			}
		}
		return strings.Replace(cleaned, "T", " ", 1)
	}

	if strings.Contains(cleaned, " ") && (strings.HasSuffix(cleaned, "Z") || hasTimezoneOffset(cleaned)) {
		candidate := strings.Replace(cleaned, " ", "T", 1)
		if t, err := time.Parse(time.RFC3339Nano, candidate); err == nil {
			return formatMySQLDateTime(t)
		}
		if t, err := time.Parse(time.RFC3339, candidate); err == nil {
			return formatMySQLDateTime(t)
		}
	}

	return value
}

// hasTimezoneOffset 检查字符串是否包含有效的时区偏移（+hh:mm, -hh:mm, +hhmm, -hhmm）
func hasTimezoneOffset(text string) bool {
	pos := strings.LastIndexAny(text, "+-")
	if pos < 0 || pos < 10 || pos+1 > len(text) {
		return false
	}

	offset := text[pos+1:]
	if len(offset) == 5 && offset[2] == ':' {
		return isAllDigits(offset[0:2]) && isAllDigits(offset[3:])
	}
	if len(offset) == 4 {
		return isAllDigits(offset)
	}

	return false
}

// isAllDigits 检查字符串是否仅包含数字字符
func isAllDigits(text string) bool {
	if text == "" {
		return false
	}
	for _, r := range text {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// formatMySQLDateTime 将time.Time格式化为MySQL DATETIME字符串，保留微秒部分
func formatMySQLDateTime(t time.Time) string {
	base := t.Format("2006-01-02 15:04:05")
	nanos := t.Nanosecond()
	if nanos == 0 {
		return base
	}
	micro := nanos / 1000
	return fmt.Sprintf("%s.%06d", base, micro)
}
