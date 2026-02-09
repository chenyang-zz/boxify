package main

import (
	"Boxify/internal/connection"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"strings"
	"sync"
)

// App struct
type App struct {
	ctx     context.Context
	dbCache map[string]Database // 缓存数据库连接
	mu      sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		dbCache: make(map[string]Database),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app terminates
func (a *App) shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, db := range a.dbCache {
		db.Close()
	}
}

// 获取或创建一个数据库连接
func (a *App) getDatabase(config *connection.ConnectionConfig) (Database, error) {
	key := getCacheKey(config)

	a.mu.Lock()
	defer a.mu.Unlock()

	if db, ok := a.dbCache[key]; ok {
		// 检查连接是否还活着
		if err := db.Ping(); err == nil {
			return db, nil
		}

		// 连接不可用，关闭并删除缓存
		db.Close()
		delete(a.dbCache, key)
	}

	// 创建新的数据库连接
	db, err := NewDatabase(config.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %v", err)
	}

	if err := db.Connect(config); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// 缓存数据库连接
	a.dbCache[key] = db
	return db, nil
}

// 通用数据库方法

func (a *App) DBConnect(config *connection.ConnectionConfig) *connection.QueryResult {
	key := getCacheKey(config)
	a.mu.Lock()
	defer a.mu.Unlock()

	if oldDB, ok := a.dbCache[key]; ok {
		oldDB.Close()
		delete(a.dbCache, key)
	}

	_, err := a.getDatabase(config)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "连接成功",
	}
}

// 根据连接配置生成一个唯一的缓存键，以便在dbCache中存储和检索数据库连接
func getCacheKey(config *connection.ConnectionConfig) string {
	// 包括数据库类型、主机、端口、用户、数据库名称（如果相关的话，还有SSH参数）
	return fmt.Sprintf("%s|%s|%s:%d|%s|%s|%v", config.Type, config.User, config.Host, config.Port, config.Database, config.SSH.Host, config.UseSSH)
}

// 兼容性包装

func (a *App) MySQLConnect(config *connection.ConnectionConfig) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBConnect(config)
}

func (a *App) MySQLQuery(config *connection.ConnectionConfig, dbName, query string) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBQuery(config, dbName, query)
}

func (a *App) MySQLGetDatabases(config *connection.ConnectionConfig) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBGetDatabases(config)
}

func (a *App) MySQLGetTables(config *connection.ConnectionConfig, dbName string) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBGetTables(config, dbName)
}

func (a *App) MySQLShowCreateTable(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBShowCreateTable(config, dbName, tableName)
}

// CreateDatabase 创建一个新的数据库
func (a *App) CreateDatabase(config *connection.ConnectionConfig, dbName string) *connection.QueryResult {
	runConfig := *config
	runConfig.Database = ""

	db, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	query := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName)
	if runConfig.Type == "postgres" {
		query = fmt.Sprintf("CREATE DATABASE \"%s\"", dbName)
	}

	_, err = db.Exec(query)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "数据库创建成功",
	}
}

// DBQuery 执行一个查询并返回结果
func (a *App) DBQuery(config *connection.ConnectionConfig, dbName, query string) *connection.QueryResult {
	runConfig := *config

	if dbName != "" {
		runConfig.Database = dbName
	}

	db, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	// 检查是否是一个 SELECT 查询
	lowerQuery := strings.TrimSpace(strings.ToLower(query))
	if strings.HasPrefix(lowerQuery, "select") || strings.HasPrefix(lowerQuery, "show") || strings.HasPrefix(lowerQuery, "describe") || strings.HasPrefix(lowerQuery, "explain") {
		data, columns, err := db.Query(query)
		if err != nil {
			return &connection.QueryResult{
				Success: false,
				Message: err.Error(),
			}
		}
		return &connection.QueryResult{
			Success: true,
			Message: "查询成功",
			Data:    data,
			Fields:  columns,
		}
	} else {
		// Exec
		affected, err := db.Exec(query)
		if err != nil {
			return &connection.QueryResult{
				Success: false,
				Message: err.Error(),
			}
		}
		return &connection.QueryResult{
			Success: true,
			Message: fmt.Sprintf("执行成功，受影响的行数: %d", affected),
			Data: map[string]int64{
				"affectedRows": affected,
			},
		}
	}
}

// DBGetDatabases 获取数据库列表
func (a *App) DBGetDatabases(config *connection.ConnectionConfig) *connection.QueryResult {
	db, err := a.getDatabase(config)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	dbs, err := db.GetDatabases()
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	var resData []map[string]string
	for _, name := range dbs {
		resData = append(resData, map[string]string{
			"Database": name,
		})
	}

	return &connection.QueryResult{
		Success: true,
		Message: "获取数据库列表成功",
		Data:    resData,
	}
}

// DBGetTables 获取表列表
func (a *App) DBGetTables(config *connection.ConnectionConfig, dbName string) *connection.QueryResult {
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	db, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	tables, err := db.GetTables(dbName)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	var resData []map[string]string
	for _, name := range tables {
		resData = append(resData, map[string]string{
			"Table": name,
		})
	}

	return &connection.QueryResult{
		Success: true,
		Message: "获取表列表成功",
		Data:    resData,
	}
}

// DBShowCreateTable 获取建表语句
func (a *App) DBShowCreateTable(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
	runeConfig := *config
	if dbName != "" {
		runeConfig.Database = dbName
	}

	db, err := a.getDatabase(&runeConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	sqlStr, err := db.GetCreateStatement(dbName, tableName)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "获取建表语句成功",
		Data:    sqlStr,
	}
}

// DBGetColumns 获取列信息
func (a *App) DBGetColumns(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	db, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	columns, err := db.GetColumns(dbName, tableName)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "获取列信息成功",
		Data:    columns,
	}
}

// DBGetIndexes 获取索引信息
func (a *App) DBGetIndexes(config *connection.ConnectionConfig, dbName string, tableName string) *connection.QueryResult {
	runeConfig := *config
	if dbName != "" {
		runeConfig.Database = dbName
	}

	db, err := a.getDatabase(&runeConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	indexes, err := db.GetIndexes(dbName, tableName)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "获取索引信息成功",
		Data:    indexes,
	}
}

// DBGetForeignKeys 获取外键信息
func (a *App) DBGetForeignKeys(config *connection.ConnectionConfig, dbName string, tableName string) *connection.QueryResult {
	runeConfig := *config
	if dbName != "" {
		runeConfig.Database = dbName
	}

	db, err := a.getDatabase(&runeConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	fks, err := db.GetForeignKeys(dbName, tableName)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "获取外键信息成功",
		Data:    fks,
	}
}

// DBGetForeignKeys 获取外键信息
func (a *App) DBGetTriggers(config *connection.ConnectionConfig, dbName string, tableName string) *connection.QueryResult {
	runeConfig := *config
	if dbName != "" {
		runeConfig.Database = dbName
	}

	db, err := a.getDatabase(&runeConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	triggers, err := db.GetTriggers(dbName, tableName)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "获取触发器信息成功",
		Data:    triggers,
	}
}

// DBGetAllColumns 获取所有列信息（包含系统表）
func (a *App) DBGetAllColumns(config *connection.ConnectionConfig, dbName string) *connection.QueryResult {
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	db, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	columns, err := db.GetAllColumns(dbName)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "获取所有列信息成功",
		Data:    columns,
	}
}

// OpenSQLFile 打开一个文件选择对话框，允许用户选择一个SQL文件，并读取其内容返回
func (a *App) OpenSQLFile() *connection.QueryResult {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select SQL File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "SQL Files (*.sql)",
				Pattern:     "*.sql",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	if selection == "" {
		return &connection.QueryResult{
			Success: false,
			Message: "Cancelled",
		}
	}

	content, err := os.ReadFile(selection)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "SQL文件加载成功",
		Data:    string(content),
	}
}

// ImportData 打开一个文件选择对话框，允许用户选择一个CSV或JSON文件，并将其内容导入到指定的数据库表中
func (a *App) ImportData(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: fmt.Sprintf("Import into %s", tableName),
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Data Files",
				Pattern:     "*csv;*.json",
			},
		},
	})

	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	if selection == "" {
		return &connection.QueryResult{
			Success: false,
			Message: "Cancelled",
		}
	}

	// 读取文件
	f, err := os.Open(selection)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}
	defer f.Close()

	// 基于文件扩展名解析
	var rows []map[string]interface{}

	if strings.HasSuffix(strings.ToLower(selection), ".json") {
		decoder := json.NewDecoder(f)
		if err := decoder.Decode(&rows); err != nil {
			return &connection.QueryResult{
				Success: false,
				Message: fmt.Sprintf("Failed to parse JSON: %v", err),
			}
		}
	} else if strings.HasSuffix(strings.ToLower(selection), ".csv") {
		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		if err != nil {
			return &connection.QueryResult{
				Success: false,
				Message: fmt.Sprintf("Failed to parse CSV: %v", err),
			}
		}
		if len(records) < 2 {
			return &connection.QueryResult{
				Success: false,
				Message: "CSV是空的或没有头行",
			}
		}
		headers := records[0]
		for _, record := range records[1:] {
			row := make(map[string]interface{})
			for i, val := range record {
				if i < len(headers) {
					if val == "NULL" {
						row[headers[i]] = nil
					} else {
						row[headers[i]] = val
					}
				}
			}
			rows = append(rows, row)
		}
	} else {
		return &connection.QueryResult{
			Success: false,
			Message: "不支持的文件类型",
		}
	}

	if len(rows) == 0 {
		return &connection.QueryResult{
			Success: true,
			Message: "没有数据可导入",
		}
	}

	// 获取数据库连接
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	db, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	successCount := 0
	errCount := 0
	firstRow := rows[0]
	var cols []string
	for k := range firstRow {
		cols = append(cols, k)
	}

	for _, row := range rows {
		var values []string
		for _, col := range cols {
			val := row[col]
			if val == nil {
				values = append(values, "NULL")
			} else {
				vStr := fmt.Sprintf("%v", val)
				vStr = strings.ReplaceAll(vStr, "'", "''")
				values = append(values, fmt.Sprintf("'%s'", vStr))
			}
		}

		query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", tableName, strings.Join(cols, ", "), strings.Join(values, ", "))

		if runConfig.Type == "postgres" {
			pgCols := make([]string, len(cols))
			for i, c := range cols {
				pgCols[i] = fmt.Sprintf(`"%s"`, c)
			}
			query = fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`, tableName, strings.Join(pgCols, ", "), strings.Join(values, ", "))
		}

		_, err := db.Exec(query)
		if err != nil {
			errCount++
			fmt.Printf("导入错误: %v\n", err)
		} else {
			successCount++
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: fmt.Sprintf("导入完成，成功: %d, 失败: %d", successCount, errCount),
	}
}

// ApplyChanges 将更改集应用到数据库表中
func (a *App) ApplyChanges(config *connection.ConnectionConfig, dbName, tableName string, changes *connection.ChangeSet) *connection.QueryResult {
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	db, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	if applier, ok := db.(BatchApplier); ok {
		err := applier.ApplyChanges(tableName, changes)
		if err != nil {
			return &connection.QueryResult{
				Success: false,
				Message: err.Error(),
			}
		}
		return &connection.QueryResult{
			Success: true,
			Message: "批量更改应用成功",
		}
	}
	return &connection.QueryResult{
		Success: false,
		Message: "数据库不支持批量更改",
	}
}

// ExportTable 导出表数据到CSV、JSON或Markdown文件
func (a *App) ExportTable(config *connection.ConnectionConfig, dbName, tableName string, format string) *connection.QueryResult {
	filename, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           fmt.Sprintf("导出 ", tableName),
		DefaultFilename: fmt.Sprintf("%s.%s", tableName, format),
	})

	if err != nil || filename == "" {
		return &connection.QueryResult{
			Success: false,
			Message: "Cancelled",
		}
	}

	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	db, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	query := fmt.Sprintf("SELECT * FROM `%s`", tableName)
	if runConfig.Type == "postgres" {
		query = fmt.Sprintf(`SELECT * FROM "%s"`, tableName)
	}

	data, columns, err := db.Query(query)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}
	defer f.Close()

	format = strings.ToLower(format)
	var csvWriter *csv.Writer
	var jsonEncoder *json.Encoder
	var isJsonFirstRow = true

	switch format {
	case "csv", "xlsx":
		f.Write([]byte{0xEF, 0xBB, 0xBF})
		csvWriter = csv.NewWriter(f)
		defer csvWriter.Flush()
		if err := csvWriter.Write(columns); err != nil {
			return &connection.QueryResult{
				Success: false,
				Message: err.Error(),
			}
		}
	case "json":
		f.WriteString("[\n")
		jsonEncoder = json.NewEncoder(f)
		jsonEncoder.SetIndent("  ", "  ")
	case "md":
		fmt.Fprint(f, "| %s |\n", strings.Join(columns, " | "))
		seps := make([]string, len(columns))
		for i := range seps {
			seps[i] = "---"
		}
		fmt.Fprint(f, "| %s |\n", strings.Join(seps, " | "))
	default:
		return &connection.QueryResult{
			Success: false,
			Message: "不支持的导出格式",
		}
	}

	for _, rowMap := range data {
		record := make([]string, len(columns))
		for i, col := range columns {
			val := rowMap[col]
			if val == nil {
				record[i] = "NULL"
			} else {
				s := fmt.Sprintf("%v", val)
				if format == "md" {
					s = strings.ReplaceAll(s, "|", "\\|")
					s = strings.ReplaceAll(s, "\n", "<br>")
				}
				record[i] = s
			}
		}

		switch format {
		case "csv", "xlsx":
			if err := csvWriter.Write(record); err != nil {
				return &connection.QueryResult{
					Success: false,
					Message: err.Error(),
				}
			}
		case "json":
			if !isJsonFirstRow {
				f.WriteString(",\n")
			}
			if err := jsonEncoder.Encode(rowMap); err != nil {
				return &connection.QueryResult{
					Success: false,
					Message: err.Error(),
				}
			}
			isJsonFirstRow = false
		case "md":
			fmt.Fprintf(f, "| %s |\n", strings.Join(record, " | "))
		}
	}

	if format == "json" {
		f.WriteString("]\n")
	}

	return &connection.QueryResult{
		Success: true,
		Message: "导出成功",
	}
}
