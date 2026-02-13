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
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/chenyang-zz/boxify/internal/db"
	"github.com/chenyang-zz/boxify/internal/logger"
	"github.com/chenyang-zz/boxify/internal/utils"

	"github.com/pkg/errors"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wailsapp/wails/v3/pkg/application"

	"strings"
	"sync"
)

// 数据库缓存的ping间隔，超过这个时间会自动ping一次以保持连接活跃
const dbCachePingInterval = 30 * time.Second

// cachedDatabase 包含数据库实例和上次ping的时间戳，用于连接池管理
type cachedDatabase struct {
	inst     db.Database
	lastPing time.Time
}

// AppService struct
type AppService struct {
	ctx     context.Context
	app     *application.App
	dbCache map[string]cachedDatabase // 缓存数据库连接
	mu      sync.RWMutex
}

// NewApp 新建一个App实例
func NewService(app *application.App) *AppService {
	return &AppService{
		app:     app,
		dbCache: make(map[string]cachedDatabase),
	}
}

// Startup 是在应用程序启动时调用的函数
func (a *AppService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	buildType := "prod"
	if a.app.Env.Info().Debug {
		buildType = "dev"
	}
	a.ctx = context.WithValue(ctx, "buildType", buildType)

	logger.Init(a.ctx)
	logger.Infof("服务启动完成")
	return nil
}

// Shutdown 是在应用程序关闭时调用的函数
func (a *AppService) ServiceShutdown() error {
	logger.Infof("服务开始关闭，准备释放资源")
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, dbInst := range a.dbCache {
		if err := dbInst.inst.Close(); err != nil {
			logger.Errorf("关闭数据库连接失败: %v", err)
		}
	}

	logger.Infof("资源释放完成，服务已关闭")
	logger.Close()

	return nil
}

// getDatabaseForcePing 强制ping数据库连接，适用于需要确保连接可用的场景
func (a *AppService) getDatabaseForcePing(config *connection.ConnectionConfig) (db.Database, error) {
	return a.getDatabaseWithPing(config, true)
}

// 获取或创建一个数据库连接
func (a *AppService) getDatabase(config *connection.ConnectionConfig) (db.Database, error) {
	return a.getDatabaseWithPing(config, false)
}

// getDatabaseWithPing 在获取数据库连接时增加ping检查，确保连接可用
func (a *AppService) getDatabaseWithPing(config *connection.ConnectionConfig, forcePing bool) (db.Database, error) {
	key := getCacheKey(config)
	shortKey := key
	if len(shortKey) > 12 {
		shortKey = key[:12]
	}

	a.mu.RLock()
	entry, ok := a.dbCache[key]
	a.mu.RUnlock()

	if ok {
		needPing := forcePing
		if !needPing {
			lastPing := entry.lastPing
			if lastPing.IsZero() || time.Since(lastPing) >= dbCachePingInterval {
				needPing = true
			}
		}

		if !needPing {
			return entry.inst, nil
		}

		if err := entry.inst.Ping(); err == nil {
			// 更新ping时间戳
			a.mu.Lock()
			if cur, exist := a.dbCache[key]; exist && cur.inst == entry.inst {
				cur.lastPing = time.Now()
				a.dbCache[key] = cur
			}
			a.mu.Unlock()
			return entry.inst, nil
		} else {
			logger.Errorf("缓存连接不可用，准备重建：%s 缓存Key=%s， 错误：%v", formatConnSummary(config), shortKey, err)
		}

		// ping失败，关闭连接并删除缓存
		a.mu.Lock()
		if cur, exists := a.dbCache[key]; exists && cur.inst == entry.inst {
			if err := cur.inst.Close(); err != nil {
				logger.Errorf("关闭失效缓存连接失败：缓存Key=%s, 错误：%v", shortKey, err)
			}
			delete(a.dbCache, key)
		}
		a.mu.Unlock()
	}

	logger.Infof("获取数据库连接：%s 缓存Key=%s", formatConnSummary(config), shortKey)
	logger.Infof("创建数据库驱动实例：类型=%s 缓存Key=%s", config.Type, shortKey)
	dbInst, err := db.NewDatabase(config.Type)
	if err != nil {
		logger.Errorf("创建数据库驱动实例失败：类型=%s 缓存Key=%s, 错误：%v", config.Type, shortKey, err)
		return nil, err
	}

	if err = dbInst.Connect(config); err != nil {
		wrapped := wrapConnectError(config, err)
		logger.Errorf("建立数据库连接失败：%s 缓存Key=%s, 错误：%v", formatConnSummary(config), shortKey, wrapped)
		return nil, wrapped
	}

	now := time.Now()
	a.mu.Lock()
	if existing, exist := a.dbCache[key]; exist && existing.inst != nil {
		a.mu.Unlock()
		_ = dbInst.Close()
		return existing.inst, nil
	}
	a.dbCache[key] = cachedDatabase{inst: dbInst, lastPing: now}
	a.mu.Unlock()

	logger.Infof("数据库连接成功并写入缓存：%s 缓存Key=%s", formatConnSummary(config), shortKey)
	return dbInst, nil
}

// 根据连接配置生成一个唯一的缓存键，以便在dbCache中存储和检索数据库连接
func getCacheKey(config *connection.ConnectionConfig) string {
	if !config.UseSSH {
		config.SSH = &connection.SSHConfig{}
	}

	// 保持与驱动默认一致，避免同一连接被重复缓存
	if config.Type == "postgres" && config.Database == "" {
		config.Database = "postgres"
	}

	b, _ := json.Marshal(config)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// wrapConnectError 包装连接错误，如果是网络超时错误则添加更多上下文信息，并附加日志文件路径提示
func wrapConnectError(config *connection.ConnectionConfig, err error) error {
	if err == nil {
		return nil
	}

	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		dbName := config.Database
		if dbName == "" {
			dbName = "<default>"
		}
		err = fmt.Errorf("数据库连接超时：%s %s:%d/%s：%w", config.Type, config.Host, config.Port, dbName, err)
	}

	return withLogHint{
		err:     err,
		logPath: logger.Path(),
	}
}

// formatConnSummary 格式化连接信息摘要，用于日志输出，包含类型、主机、端口和数据库名
func formatConnSummary(config *connection.ConnectionConfig) string {
	timeoutSeconds := config.Timeout
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	dbName := config.Database
	if dbName == "" {
		dbName = "<default>"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("类型=%s 地址=%s:%d 数据库=%s 用户=%s 超时=%ds",
		config.Type, config.Host, config.Port, dbName, config.User, timeoutSeconds))

	if config.UseSSH {
		b.WriteString(fmt.Sprintf(" SSH=%s:%d 用户=%s", config.SSH.Host, config.SSH.Port, config.SSH.User))
	}

	if config.Type == "custom" {
		driver := strings.TrimSpace(config.Driver)
		if driver == "" {
			driver = "<未配置>"
		}
		dsnState := "<未配置>"
		if strings.TrimSpace(config.DSN) != "" {
			dsnState = fmt.Sprintf("已配置(长度=%d)", len(config.DSN))
		}
		b.WriteString(fmt.Sprintf(" 驱动=%s DSN=%s", driver, dsnState))
	}

	return b.String()
}

type withLogHint struct {
	err     error
	logPath string
}

func (e withLogHint) Error() string {
	if strings.TrimSpace(e.logPath) == "" {
		return e.err.Error()
	}
	return fmt.Sprintf("%s（详细日志：%s）", e.err.Error(), e.logPath)
}

func (e withLogHint) Unwrap() error {
	return e.err
}

// 兼容性包装

func (a *AppService) MySQLConnect(config *connection.ConnectionConfig) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBConnect(config)
}

func (a *AppService) MySQLQuery(config *connection.ConnectionConfig, dbName, query string, args []any) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBQuery(config, dbName, query, args)
}

func (a *AppService) MySQLGetDatabases(config *connection.ConnectionConfig) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBGetDatabases(config)
}

func (a *AppService) MySQLGetTables(config *connection.ConnectionConfig, dbName string) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBGetTables(config, dbName)
}

func (a *AppService) MySQLShowCreateTable(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
	config.Type = "mysql"
	return a.DBShowCreateTable(config, dbName, tableName)
}

// DBQuery 执行一个查询并返回结果
func (a *AppService) DBQuery(config *connection.ConnectionConfig, dbName, query string, args []any) *connection.QueryResult {
	runConfig := normalizeRunConfig(config, dbName)

	dbInst, err := a.getDatabase(runConfig)
	if err != nil {
		logger.ErrorfWithTrace(err, "DBQuery 获取连接失败：%s", formatConnSummary(runConfig))
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	query = sanitizeSQLForPgLike(runConfig.Type, query)
	timeoutSeconds := runConfig.Timeout
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	ctx, cancel := utils.ContextWithTimeout(time.Duration(timeoutSeconds) * time.Second)
	defer cancel()

	// 检查是否是一个 SELECT 查询
	lowerQuery := strings.TrimSpace(strings.ToLower(query))
	if strings.HasPrefix(lowerQuery, "select") || strings.HasPrefix(lowerQuery, "show") || strings.HasPrefix(lowerQuery, "describe") || strings.HasPrefix(lowerQuery, "explain") {
		var data []map[string]interface{}
		var columns []string

		if q, ok := dbInst.(interface {
			QueryContext(context.Context, string, ...any) ([]map[string]interface{}, []string, error)
		}); ok {
			data, columns, err = q.QueryContext(ctx, query, args...)
		} else {
			data, columns, err = dbInst.Query(query, args...)
		}
		if err != nil {
			logger.ErrorfWithTrace(err, "DBQuery 查询失败：%s SQL片段=%q", formatConnSummary(runConfig), sqlSnippet(query))
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

		var affected int64
		if e, ok := dbInst.(interface {
			ExecContext(context.Context, string) (int64, error)
		}); ok {
			affected, err = e.ExecContext(ctx, query)
		} else {
			affected, err = dbInst.Exec(query)
		}
		if err != nil {
			logger.ErrorfWithTrace(err, "DBQuery 执行失败：%s SQL片段=%q", formatConnSummary(runConfig), sqlSnippet(query))
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
func (a *AppService) DBGetDatabases(config *connection.ConnectionConfig) *connection.QueryResult {
	dbInst, err := a.getDatabase(config)
	if err != nil {
		logger.ErrorfWithTrace(err, "DBGetDatabases 获取连接失败：%s", formatConnSummary(config))
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	dbs, err := dbInst.GetDatabases()
	if err != nil {
		logger.ErrorfWithTrace(err, "DBGetDatabases 获取数据库列表失败：%s", formatConnSummary(config))
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
func (a *AppService) DBGetTables(config *connection.ConnectionConfig, dbName string) *connection.QueryResult {
	runConfig := normalizeRunConfig(config, dbName)

	dbInst, err := a.getDatabase(runConfig)
	if err != nil {
		logger.ErrorfWithTrace(err, "DBGetTables 获取连接失败：%s", formatConnSummary(runConfig))
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	tables, err := dbInst.GetTables(dbName)
	if err != nil {
		logger.ErrorfWithTrace(err, "DBGetTables 获取表列表失败：%s", formatConnSummary(runConfig))
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
func (a *AppService) DBShowCreateTable(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
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
func (a *AppService) DBGetColumns(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
	runConfig := normalizeRunConfig(config, dbName)

	db, err := a.getDatabase(runConfig)
	if err != nil {
		logger.ErrorfWithTrace(err, "DBGetColumns 获取连接失败：%s", formatConnSummary(runConfig))
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	schemaName, pureTableName := normalizeSchemaAndTable(config, dbName, tableName)
	columns, err := db.GetColumns(schemaName, pureTableName)
	if err != nil {
		logger.ErrorfWithTrace(err, "DBGetColumns 获取列信息失败：%s.%s.%s", formatConnSummary(runConfig), schemaName, pureTableName)
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
func (a *AppService) DBGetIndexes(config *connection.ConnectionConfig, dbName string, tableName string) *connection.QueryResult {
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
func (a *AppService) DBGetForeignKeys(config *connection.ConnectionConfig, dbName string, tableName string) *connection.QueryResult {
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
func (a *AppService) DBGetTriggers(config *connection.ConnectionConfig, dbName string, tableName string) *connection.QueryResult {
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
func (a *AppService) DBGetAllColumns(config *connection.ConnectionConfig, dbName string) *connection.QueryResult {
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
func (a *AppService) OpenSQLFile() *connection.QueryResult {
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
func (a *AppService) ImportData(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
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
func (a *AppService) ApplyChanges(config *connection.ConnectionConfig, dbName, tableName string, changes *connection.ChangeSet) *connection.QueryResult {
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}

	dbInst, err := a.getDatabase(&runConfig)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}

	if applier, ok := dbInst.(db.BatchApplier); ok {
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
func (a *AppService) ExportTable(config *connection.ConnectionConfig, dbName, tableName string, format string) *connection.QueryResult {
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

func (a *AppService) TypeOnly_ColumnDefinition() *connection.ColumnDefinition {
	return &connection.ColumnDefinition{}
}

// sqlSnippet 返回SQL查询的简短片段，用于日志输出，限制长度以避免过长
func sqlSnippet(query string) string {
	q := strings.TrimSpace(query)
	const max = 200
	if len(q) <= max {
		return q
	}
	return q[:max] + "..."
}
