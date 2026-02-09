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
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"

	"github.com/joho/godotenv"
)

// getTestConfig 从环境变量获取测试配置
func getTestConfig() *connection.ConnectionConfig {
	// 加载 .env.local 文件
	_ = godotenv.Load("../../.env.local")

	host := os.Getenv("TEST_MYSQL_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 3306
	if portStr := os.Getenv("TEST_MYSQL_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	user := os.Getenv("TEST_MYSQL_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("TEST_MYSQL_PASSWORD")
	if password == "" {
		password = ""
	}

	database := os.Getenv("TEST_MYSQL_DB")
	if database == "" {
		database = "boxify_test"
	}

	return &connection.ConnectionConfig{
		Type:     "mysql",
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
		Timeout:  10,
		UseSSH:   false,
	}
}

// setupTestDB 创建测试数据库连接并准备测试数据
func setupTestDB(t *testing.T) *MySQLDB {
	config := getTestConfig()

	db := &MySQLDB{}
	if err := db.Connect(config); err != nil {
		t.Skipf("无法连接到测试数据库: %v\n请确保 .env.local 中配置了正确的测试数据库连接信息", err)
		return nil
	}

	// 创建测试表
	createTestTables(t, db)

	return db
}

// createTestTables 创建测试所需的表
func createTestTables(t *testing.T, db *MySQLDB) {
	// 创建用户表
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS test_users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) NOT NULL UNIQUE,
		email VARCHAR(100) NOT NULL,
		age INT DEFAULT 0,
		is_active BIT(1) DEFAULT b'1',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	// 创建订单表（用于测试外键）
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS test_orders (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		order_no VARCHAR(50) NOT NULL,
		amount DECIMAL(10,2) DEFAULT 0.00,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES test_users(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil && !strings.Contains(err.Error(), "foreign key constraint") {
		// 如果外键约束失败，继续测试（可能表已存在）
		t.Logf("创建订单表警告: %v", err)
	}

	// 清空测试数据
	db.Exec("DELETE FROM test_orders")
	db.Exec("DELETE FROM test_users")

	// 插入测试数据
	_, err = db.Exec(`INSERT INTO test_users (username, email, age, is_active) VALUES
		('alice', 'alice@example.com', 25, b'1'),
		('bob', 'bob@example.com', 30, b'0'),
		('charlie', 'charlie@example.com', 35, b'1')`)
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	_, err = db.Exec(`INSERT INTO test_orders (user_id, order_no, amount) VALUES
		(1, 'ORD001', 100.50),
		(1, 'ORD002', 200.00),
		(2, 'ORD003', 150.75)`)
	if err != nil && !strings.Contains(err.Error(), "foreign key constraint") {
		t.Logf("插入订单数据警告: %v", err)
	}
}

// cleanupTestDB 清理测试数据
func cleanupTestDB(t *testing.T, db *MySQLDB) {
	if db != nil {
		db.Exec("DROP TABLE IF EXISTS test_orders")
		db.Exec("DROP TABLE IF EXISTS test_users")
		db.Close()
	}
}

// TestMySQLDB_Connect 测试数据库连接
func TestMySQLDB_Connect(t *testing.T) {
	config := getTestConfig()

	t.Run("有效配置", func(t *testing.T) {
		db := &MySQLDB{}
		err := db.Connect(config)
		if err != nil {
			t.Skipf("无法连接到测试数据库: %v", err)
			return
		}
		defer db.Close()

		if db.conn == nil {
			t.Error("连接成功但 conn 为 nil")
		}
	})

	t.Run("无效配置", func(t *testing.T) {
		db := &MySQLDB{}
		invalidConfig := &connection.ConnectionConfig{
			Type:     "mysql",
			Host:     "invalid-host",
			Port:     9999,
			User:     "invalid",
			Password: "invalid",
			Database: "invalid",
			Timeout:  1,
		}
		err := db.Connect(invalidConfig)
		if err == nil {
			db.Close()
			t.Error("期望连接失败，但连接成功")
		}
	})
}

// TestMySQLDB_Ping 测试 Ping 功能
func TestMySQLDB_Ping(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	if err := db.Ping(); err != nil {
		t.Errorf("Ping() 失败: %v", err)
	}
}

// TestMySQLDB_Query 测试查询功能
func TestMySQLDB_Query(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	t.Run("简单查询", func(t *testing.T) {
		data, columns, err := db.Query("SELECT * FROM test_users ORDER BY id")
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		if len(data) != 3 {
			t.Errorf("期望返回 3 行数据，得到 %d 行", len(data))
		}

		if len(columns) == 0 {
			t.Error("没有返回列名")
		}
	})

	t.Run("带条件查询", func(t *testing.T) {
		data, _, err := db.Query("SELECT * FROM test_users WHERE username = 'alice'")
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		if len(data) != 1 {
			t.Errorf("期望返回 1 行数据，得到 %d 行", len(data))
		}

		if data[0]["username"] != "alice" {
			t.Errorf("期望用户名为 alice，得到 %v", data[0]["username"])
		}
	})

	t.Run("聚合查询", func(t *testing.T) {
		data, _, err := db.Query("SELECT COUNT(*) as count FROM test_users")
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		if len(data) != 1 {
			t.Errorf("期望返回 1 行数据，得到 %d 行", len(data))
		}
	})

	t.Run("无效查询", func(t *testing.T) {
		_, _, err := db.Query("SELECT * FROM nonexistent_table")
		if err == nil {
			t.Error("期望查询失败，但查询成功")
		}
	})
}

// TestMySQLDB_Exec 测试执行功能
func TestMySQLDB_Exec(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	t.Run("插入数据", func(t *testing.T) {
		affected, err := db.Exec("INSERT INTO test_users (username, email, age) VALUES ('david', 'david@example.com', 28)")
		if err != nil {
			t.Fatalf("插入失败: %v", err)
		}

		if affected != 1 {
			t.Errorf("期望影响 1 行，得到 %d 行", affected)
		}
	})

	t.Run("更新数据", func(t *testing.T) {
		affected, err := db.Exec("UPDATE test_users SET age = 26 WHERE username = 'alice'")
		if err != nil {
			t.Fatalf("更新失败: %v", err)
		}

		if affected != 1 {
			t.Errorf("期望影响 1 行，得到 %d 行", affected)
		}
	})

	t.Run("删除数据", func(t *testing.T) {
		affected, err := db.Exec("DELETE FROM test_users WHERE username = 'david'")
		if err != nil {
			t.Fatalf("删除失败: %v", err)
		}

		if affected != 1 {
			t.Errorf("期望影响 1 行，得到 %d 行", affected)
		}
	})
}

// TestMySQLDB_GetDatabases 测试获取数据库列表
func TestMySQLDB_GetDatabases(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	databases, err := db.GetDatabases()
	if err != nil {
		t.Fatalf("获取数据库列表失败: %v", err)
	}

	if len(databases) == 0 {
		t.Error("数据库列表为空")
	}

	// 检查是否包含测试数据库
	found := false
	testDB := os.Getenv("TEST_MYSQL_DB")
	if testDB == "" {
		testDB = "boxify_test"
	}
	for _, db := range databases {
		if db == testDB {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("数据库列表中未找到测试数据库 %s", testDB)
	}
}

// TestMySQLDB_GetTables 测试获取表列表
func TestMySQLDB_GetTables(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	t.Run("获取指定数据库的表", func(t *testing.T) {
		testDB := os.Getenv("TEST_MYSQL_DB")
		if testDB == "" {
			testDB = "boxify_test"
		}

		tables, err := db.GetTables(testDB)
		if err != nil {
			t.Fatalf("获取表列表失败: %v", err)
		}

		// 检查是否包含测试表
		found := false
		for _, table := range tables {
			if table == "test_users" {
				found = true
				break
			}
		}
		if !found {
			t.Error("表列表中未找到 test_users 表")
		}
	})
}

// TestMySQLDB_GetCreateStatement 测试获取建表语句
func TestMySQLDB_GetCreateStatement(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	testDB := os.Getenv("TEST_MYSQL_DB")
	if testDB == "" {
		testDB = "boxify_test"
	}

	sql, err := db.GetCreateStatement(testDB, "test_users")
	if err != nil {
		t.Fatalf("获取建表语句失败: %v", err)
	}

	if sql == "" {
		t.Error("建表语句为空")
	}

	if !strings.Contains(sql, "CREATE TABLE") {
		t.Error("建表语句不包含 CREATE TABLE")
	}

	if !strings.Contains(sql, "test_users") {
		t.Error("建表语句不包含表名 test_users")
	}
}

// TestMySQLDB_GetColumns 测试获取列信息
func TestMySQLDB_GetColumns(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	testDB := os.Getenv("TEST_MYSQL_DB")
	if testDB == "" {
		testDB = "boxify_test"
	}

	columns, err := db.GetColumns(testDB, "test_users")
	if err != nil {
		t.Fatalf("获取列信息失败: %v", err)
	}

	if len(columns) == 0 {
		t.Error("列信息为空")
	}

	// 检查是否包含必要的列
	columnMap := make(map[string]connection.ColumnDefinition)
	for _, col := range columns {
		columnMap[col.Name] = col
	}

	requiredColumns := []string{"id", "username", "email", "age", "is_active", "created_at", "updated_at"}
	for _, colName := range requiredColumns {
		if _, ok := columnMap[colName]; !ok {
			t.Errorf("列信息中未找到预期的列: %s", colName)
		}
	}

	// 检查 id 列的属性
	if col, ok := columnMap["id"]; ok {
		if col.Key != "PRI" {
			t.Errorf("id 列的 Key 属性应该是 'PRI'，得到 '%s'", col.Key)
		}
		if !strings.Contains(col.Extra, "auto_increment") {
			t.Error("id 列应该有 auto_increment 属性")
		}
	}
}

// TestMySQLDB_GetIndexes 测试获取索引信息
func TestMySQLDB_GetIndexes(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	testDB := os.Getenv("TEST_MYSQL_DB")
	if testDB == "" {
		testDB = "boxify_test"
	}

	indexes, err := db.GetIndexes(testDB, "test_users")
	if err != nil {
		t.Fatalf("获取索引信息失败: %v", err)
	}

	if len(indexes) == 0 {
		t.Error("索引信息为空")
	}

	// 检查是否有主键索引
	foundPK := false
	for _, idx := range indexes {
		if idx.Name == "PRIMARY" {
			foundPK = true
			if idx.ColumnName != "id" {
				t.Errorf("主键索引的列名应该是 'id'，得到 '%s'", idx.ColumnName)
			}
			break
		}
	}
	if !foundPK {
		t.Error("未找到主键索引")
	}
}

// TestMySQLDB_GetForeignKeys 测试获取外键信息
func TestMySQLDB_GetForeignKeys(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	testDB := os.Getenv("TEST_MYSQL_DB")
	if testDB == "" {
		testDB = "boxify_test"
	}

	fks, err := db.GetForeignKeys(testDB, "test_orders")
	if err != nil {
		t.Fatalf("获取外键信息失败: %v", err)
	}

	// 可能有外键，也可能没有（取决于创建表时是否成功）
	if len(fks) > 0 {
		// 检查外键属性
		fk := fks[0]
		if fk.RefTableName != "test_users" {
			t.Errorf("外键应该引用 test_users 表，得到 '%s'", fk.RefTableName)
		}
	}
}

// TestMySQLDB_GetTriggers 测试获取触发器信息
func TestMySQLDB_GetTriggers(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	testDB := os.Getenv("TEST_MYSQL_DB")
	if testDB == "" {
		testDB = "boxify_test"
	}

	triggers, err := db.GetTriggers(testDB, "test_users")
	if err != nil {
		t.Fatalf("获取触发器信息失败: %v", err)
	}

	// 测试表可能没有触发器
	if len(triggers) == 0 {
		t.Log("测试表没有触发器")
		return
	}

	// 如果有触发器，验证触发器属性
	for _, trig := range triggers {
		if trig.Name == "" {
			t.Error("触发器名称为空")
		}
		if trig.Timing != "BEFORE" && trig.Timing != "AFTER" {
			t.Errorf("触发器时机无效: %s", trig.Timing)
		}
	}
}

// TestMySQLDB_GetAllColumns 测试获取所有列信息
func TestMySQLDB_GetAllColumns(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	testDB := os.Getenv("TEST_MYSQL_DB")
	if testDB == "" {
		testDB = "boxify_test"
	}

	columns, err := db.GetAllColumns(testDB)
	if err != nil {
		t.Fatalf("获取所有列信息失败: %v", err)
	}

	if len(columns) == 0 {
		t.Error("列信息为空")
	}

	// 检查是否包含测试表的列
	found := false
	for _, col := range columns {
		if col.TableName == "test_users" && col.Name == "id" {
			found = true
			break
		}
	}
	if !found {
		t.Error("所有列信息中未找到 test_users.id 列")
	}
}

// TestMySQLDB_ApplyChanges 测试批量修改
func TestMySQLDB_ApplyChanges(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	t.Run("插入操作", func(t *testing.T) {
		changes := &connection.ChangeSet{
			Inserts: []map[string]interface{}{
				{
					"username": "test_user_1",
					"email":    "test1@example.com",
					"age":      20,
				},
				{
					"username": "test_user_2",
					"email":    "test2@example.com",
					"age":      22,
				},
			},
			Updates: []connection.UpdateRow{},
			Deletes: []map[string]interface{}{},
		}

		err := db.ApplyChanges("test_users", changes)
		if err != nil {
			t.Fatalf("批量插入失败: %v", err)
		}

		// 验证数据
		data, _, err := db.Query("SELECT * FROM test_users WHERE username IN ('test_user_1', 'test_user_2')")
		if err != nil {
			t.Fatalf("验证数据失败: %v", err)
		}

		if len(data) != 2 {
			t.Errorf("期望插入 2 行数据，实际插入 %d 行", len(data))
		}
	})

	t.Run("更新操作", func(t *testing.T) {
		changes := &connection.ChangeSet{
			Inserts: []map[string]interface{}{},
			Updates: []connection.UpdateRow{
				{
					Keys: map[string]interface{}{
						"id": int64(1),
					},
					Values: map[string]interface{}{
						"age": 99,
					},
				},
			},
			Deletes: []map[string]interface{}{},
		}

		err := db.ApplyChanges("test_users", changes)
		if err != nil {
			t.Fatalf("批量更新失败: %v", err)
		}

		// 验证数据
		data, _, err := db.Query("SELECT age FROM test_users WHERE id = 1")
		if err != nil {
			t.Fatalf("验证数据失败: %v", err)
		}

		if len(data) == 0 {
			t.Error("更新后未找到数据")
		} else if data[0]["age"] != int64(99) {
			t.Errorf("期望 age = 99，得到 %v", data[0]["age"])
		}
	})

	t.Run("删除操作", func(t *testing.T) {
		changes := &connection.ChangeSet{
			Inserts: []map[string]interface{}{},
			Updates: []connection.UpdateRow{},
			Deletes: []map[string]interface{}{
				{
					"id": int64(3),
				},
			},
		}

		err := db.ApplyChanges("test_users", changes)
		if err != nil {
			t.Fatalf("批量删除失败: %v", err)
		}

		// 验证数据
		data, _, err := db.Query("SELECT * FROM test_users WHERE id = 3")
		if err != nil {
			t.Fatalf("验证数据失败: %v", err)
		}

		if len(data) != 0 {
			t.Error("删除后仍存在数据")
		}
	})

	t.Run("混合操作", func(t *testing.T) {
		changes := &connection.ChangeSet{
			Inserts: []map[string]interface{}{
				{
					"username": "mixed_insert",
					"email":    "mixed@example.com",
					"age":      40,
				},
			},
			Updates: []connection.UpdateRow{
				{
					Keys: map[string]interface{}{
						"username": "alice",
					},
					Values: map[string]interface{}{
						"age": 27,
					},
				},
			},
			Deletes: []map[string]interface{}{
				{
					"username": "bob",
				},
			},
		}

		err := db.ApplyChanges("test_users", changes)
		if err != nil {
			t.Fatalf("混合批量操作失败: %v", err)
		}

		// 验证插入
		data, _, err := db.Query("SELECT * FROM test_users WHERE username = 'mixed_insert'")
		if err != nil {
			t.Fatalf("验证插入失败: %v", err)
		}
		if len(data) != 1 {
			t.Error("插入操作失败")
		}

		// 验证更新
		data, _, err = db.Query("SELECT age FROM test_users WHERE username = 'alice'")
		if err != nil {
			t.Fatalf("验证更新失败: %v", err)
		}
		if len(data) == 0 || data[0]["age"] != int64(27) {
			t.Error("更新操作失败")
		}

		// 验证删除
		data, _, err = db.Query("SELECT * FROM test_users WHERE username = 'bob'")
		if err != nil {
			t.Fatalf("验证删除失败: %v", err)
		}
		if len(data) != 0 {
			t.Error("删除操作失败")
		}
	})
}

// TestMySQLDB_Close 测试关闭连接
func TestMySQLDB_Close(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}

	err := db.Close()
	if err != nil {
		t.Errorf("关闭连接失败: %v", err)
	}

	// 关闭后 Ping 应该失败
	err = db.Ping()
	if err == nil {
		t.Error("关闭连接后 Ping 应该失败")
	}
}

// TestMySQLDB_getDSN 测试 DSN 生成
func TestMySQLDB_getDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   *connection.ConnectionConfig
		contains []string // DSN 应该包含的字符串
	}{
		{
			name: "基础配置",
			config: &connection.ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "password",
				Database: "testdb",
				Timeout:  30,
				UseSSH:   false,
			},
			contains: []string{"root:password@", "tcp", "localhost:3306", "testdb", "timeout=30"},
		},
		{
			name: "使用 SSH",
			config: &connection.ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "password",
				Database: "testdb",
				UseSSH:   true,
				SSH: &connection.SSHConfig{
					Host:     "ssh.example.com",
					Port:     22,
					User:     "sshuser",
					Password: "sshpass",
				},
			},
			contains: []string{"root:password@", "localhost:3306", "testdb"},
		},
		{
			name: "使用默认超时",
			config: &connection.ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "password",
				Database: "testdb",
				UseSSH:   false,
			},
			contains: []string{"root:password@", "timeout=30"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &MySQLDB{}
			dsn := db.getDSN(tt.config)

			for _, substr := range tt.contains {
				if !strings.Contains(dsn, substr) {
					t.Errorf("DSN 应该包含 '%s'，实际 DSN: %s", substr, dsn)
				}
			}

			// 不应该在日志中输出密码
			t.Logf("生成的 DSN: %s", strings.ReplaceAll(dsn, tt.config.Password, "***"))
		})
	}
}

// TestMySQLDB_Concurrency 测试并发操作
func TestMySQLDB_Concurrency(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	// 并发查询
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, _, err := db.Query("SELECT * FROM test_users")
			if err != nil {
				t.Errorf("并发查询失败: %v", err)
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("并发测试超时")
		}
	}
}

// TestMySQLDB_QueryContext 测试带上下文的查询
func TestMySQLDB_QueryContext(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	t.Run("正常查询", func(t *testing.T) {
		ctx, cancel := createTestContext(5 * time.Second)
		defer cancel()

		data, columns, err := db.QueryContext(ctx, "SELECT * FROM test_users")
		if err != nil {
			t.Fatalf("带上下文的查询失败: %v", err)
		}

		if len(data) == 0 {
			t.Error("查询结果为空")
		}

		if len(columns) == 0 {
			t.Error("列信息为空")
		}
	})

	t.Run("超时查询", func(t *testing.T) {
		ctx, cancel := createTestContext(1 * time.Millisecond)
		defer cancel()

		// 使用 SLEEP 函数模拟长时间查询
		_, _, err := db.QueryContext(ctx, "SELECT SLEEP(1)")
		if err == nil {
			t.Error("期望查询超时，但查询成功")
		}
	})
}

// TestMySQLDB_ExecContext 测试带上下文的执行
func TestMySQLDB_ExecContext(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	if db == nil {
		return
	}

	ctx, cancel := createTestContext(5 * time.Second)
	defer cancel()

	affected, err := db.ExecContext(ctx, "INSERT INTO test_users (username, email) VALUES ('context_test', 'context@example.com')")
	if err != nil {
		t.Fatalf("带上下文的执行失败: %v", err)
	}

	if affected != 1 {
		t.Errorf("期望影响 1 行，得到 %d 行", affected)
	}
}

// createTestContext 创建测试用的上下文
func createTestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// BenchmarkQuery 基准测试查询性能
func BenchmarkQuery(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("无法连接到测试数据库")
		return
	}
	defer cleanupTestDB(&testing.T{}, db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = db.Query("SELECT * FROM test_users")
	}
}

// BenchmarkExec 基准测试执行性能
func BenchmarkExec(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("无法连接到测试数据库")
		return
	}
	defer cleanupTestDB(&testing.T{}, db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = db.Exec("INSERT INTO test_users (username, email) VALUES ('bench', 'bench@example.com')")
	}
	b.StopTimer()

	// 清理基准测试数据
	_, _ = db.Exec("DELETE FROM test_users WHERE username = 'bench'")
}
