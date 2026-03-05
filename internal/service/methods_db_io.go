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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/chenyang-zz/boxify/internal/db"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// exportWriterContext 封装导出场景中的写入器状态。
type exportWriterContext struct {
	format         string
	csvWriter      *csv.Writer
	jsonEncoder    *json.Encoder
	isJSONFirstRow bool
}

// OpenSQLFile 选择 SQL 文件并返回内容。
func (a *DatabaseService) OpenSQLFile() *connection.QueryResult {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select SQL File",
		Filters: []runtime.FileFilter{
			{DisplayName: "SQL Files (*.sql)", Pattern: "*.sql"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}
	if selection == "" {
		return &connection.QueryResult{Success: false, Message: "Cancelled"}
	}

	content, err := os.ReadFile(selection)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}
	return &connection.QueryResult{Success: true, Message: "SQL文件加载成功", Data: string(content)}
}

// ImportData 选择 CSV/JSON 文件并导入到目标表。
func (a *DatabaseService) ImportData(config *connection.ConnectionConfig, dbName, tableName string) *connection.QueryResult {
	selection, err := selectImportDataFile(a.ctx, tableName)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}
	if selection == "" {
		return &connection.QueryResult{Success: false, Message: "Cancelled"}
	}

	rows, err := parseImportRows(selection)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}
	if len(rows) == 0 {
		return &connection.QueryResult{Success: true, Message: "没有数据可导入"}
	}

	runConfig := cloneConfigWithDatabase(config, dbName)
	dbInst, err := a.getDatabase(runConfig)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	successCount, errCount := applyImportRows(dbInst, runConfig.Type, tableName, rows)
	return &connection.QueryResult{Success: true, Message: fmt.Sprintf("导入完成，成功: %d, 失败: %d", successCount, errCount)}
}

// ApplyChanges 将更改集应用到数据库表中。
func (a *DatabaseService) ApplyChanges(config *connection.ConnectionConfig, dbName, tableName string, changes *connection.ChangeSet) *connection.QueryResult {
	runConfig := cloneConfigWithDatabase(config, dbName)
	dbInst, err := a.getDatabase(runConfig)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	if applier, ok := dbInst.(db.BatchApplier); ok {
		if err := applier.ApplyChanges(tableName, changes); err != nil {
			return &connection.QueryResult{Success: false, Message: err.Error()}
		}
		return &connection.QueryResult{Success: true, Message: "批量更改应用成功"}
	}
	return &connection.QueryResult{Success: false, Message: "数据库不支持批量更改"}
}

// ExportTable 导出表数据到 CSV、JSON 或 Markdown 文件。
func (a *DatabaseService) ExportTable(config *connection.ConnectionConfig, dbName, tableName string, format string) *connection.QueryResult {
	filename, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           fmt.Sprintf("导出 %s", tableName),
		DefaultFilename: fmt.Sprintf("%s.%s", tableName, format),
	})
	if err != nil || filename == "" {
		return &connection.QueryResult{Success: false, Message: "Cancelled"}
	}

	runConfig := cloneConfigWithDatabase(config, dbName)
	dbInst, err := a.getDatabase(runConfig)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	query := buildExportSelectQuery(runConfig.Type, tableName)
	data, columns, err := dbInst.Query(query)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	f, err := os.Create(filename)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}
	defer f.Close()

	writerCtx, err := initExportWriter(f, strings.ToLower(format), columns)
	if err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}
	if writerCtx.csvWriter != nil {
		defer writerCtx.csvWriter.Flush()
	}

	if err := writeExportRows(f, writerCtx, columns, data); err != nil {
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}
	if writerCtx.format == "json" {
		f.WriteString("]\n")
	}

	return &connection.QueryResult{Success: true, Message: "导出成功"}
}

// TypeOnly_ColumnDefinition 仅用于导出类型到前端绑定。
func (a *DatabaseService) TypeOnly_ColumnDefinition() *connection.ColumnDefinition {
	return &connection.ColumnDefinition{}
}

// cloneConfigWithDatabase 复制连接配置并按需覆盖数据库名。
func cloneConfigWithDatabase(config *connection.ConnectionConfig, dbName string) *connection.ConnectionConfig {
	runConfig := *config
	if dbName != "" {
		runConfig.Database = dbName
	}
	return &runConfig
}

// selectImportDataFile 弹出导入文件选择窗口。
func selectImportDataFile(ctx context.Context, tableName string) (string, error) {
	return runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title: fmt.Sprintf("Import into %s", tableName),
		Filters: []runtime.FileFilter{
			{DisplayName: "Data Files", Pattern: "*csv;*.json"},
		},
	})
}

// parseImportRows 从 CSV/JSON 文件解析出待导入数据行。
func parseImportRows(selection string) ([]map[string]interface{}, error) {
	f, err := os.Open(selection)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if strings.HasSuffix(strings.ToLower(selection), ".json") {
		var rows []map[string]interface{}
		decoder := json.NewDecoder(f)
		if err := decoder.Decode(&rows); err != nil {
			return nil, fmt.Errorf("Failed to parse JSON: %v", err)
		}
		return rows, nil
	}

	if strings.HasSuffix(strings.ToLower(selection), ".csv") {
		return parseCSVRows(f)
	}

	return nil, fmt.Errorf("不支持的文件类型")
}

// parseCSVRows 将 CSV 内容转换为行对象。
func parseCSVRows(f *os.File) ([]map[string]interface{}, error) {
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Failed to parse CSV: %v", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV是空的或没有头行")
	}

	headers := records[0]
	rows := make([]map[string]interface{}, 0, len(records)-1)
	for _, record := range records[1:] {
		row := make(map[string]interface{})
		for i, val := range record {
			if i >= len(headers) {
				continue
			}
			if val == "NULL" {
				row[headers[i]] = nil
			} else {
				row[headers[i]] = val
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// applyImportRows 执行逐行导入并返回成功/失败统计。
func applyImportRows(dbInst db.Database, dbType connection.ConnectionType, tableName string, rows []map[string]interface{}) (int, int) {
	successCount := 0
	errCount := 0
	cols := extractColumnOrder(rows[0])

	for _, row := range rows {
		query := buildImportInsertQuery(dbType, tableName, cols, row)
		if _, err := dbInst.Exec(query); err != nil {
			errCount++
			fmt.Printf("导入错误: %v\n", err)
		} else {
			successCount++
		}
	}

	return successCount, errCount
}

// extractColumnOrder 从首行提取列顺序。
func extractColumnOrder(firstRow map[string]interface{}) []string {
	cols := make([]string, 0, len(firstRow))
	for k := range firstRow {
		cols = append(cols, k)
	}
	return cols
}

// buildImportInsertQuery 按数据库类型构造插入 SQL。
func buildImportInsertQuery(dbType connection.ConnectionType, tableName string, cols []string, row map[string]interface{}) string {
	values := buildImportValueTokens(cols, row)
	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", tableName, strings.Join(cols, ", "), strings.Join(values, ", "))

	if dbType == "postgres" {
		pgCols := make([]string, len(cols))
		for i, c := range cols {
			pgCols[i] = fmt.Sprintf(`"%s"`, c)
		}
		query = fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`, tableName, strings.Join(pgCols, ", "), strings.Join(values, ", "))
	}

	return query
}

// buildImportValueTokens 将行数据转换为 SQL values token 列表。
func buildImportValueTokens(cols []string, row map[string]interface{}) []string {
	values := make([]string, 0, len(cols))
	for _, col := range cols {
		val := row[col]
		if val == nil {
			values = append(values, "NULL")
			continue
		}
		vStr := fmt.Sprintf("%v", val)
		vStr = strings.ReplaceAll(vStr, "'", "''")
		values = append(values, fmt.Sprintf("'%s'", vStr))
	}
	return values
}

// buildExportSelectQuery 构造导出使用的查询语句。
func buildExportSelectQuery(dbType connection.ConnectionType, tableName string) string {
	if dbType == "postgres" {
		return fmt.Sprintf(`SELECT * FROM "%s"`, tableName)
	}
	return fmt.Sprintf("SELECT * FROM `%s`", tableName)
}

// initExportWriter 初始化导出写入器并写入头信息。
func initExportWriter(f *os.File, format string, columns []string) (*exportWriterContext, error) {
	ctx := &exportWriterContext{format: format, isJSONFirstRow: true}

	switch format {
	case "csv", "xlsx":
		f.Write([]byte{0xEF, 0xBB, 0xBF})
		ctx.csvWriter = csv.NewWriter(f)
		if err := ctx.csvWriter.Write(columns); err != nil {
			return nil, err
		}
	case "json":
		f.WriteString("[\n")
		ctx.jsonEncoder = json.NewEncoder(f)
		ctx.jsonEncoder.SetIndent("  ", "  ")
	case "md":
		fmt.Fprintf(f, "| %s |\n", strings.Join(columns, " | "))
		seps := make([]string, len(columns))
		for i := range seps {
			seps[i] = "---"
		}
		fmt.Fprintf(f, "| %s |\n", strings.Join(seps, " | "))
	default:
		return nil, fmt.Errorf("不支持的导出格式")
	}

	return ctx, nil
}

// writeExportRows 逐行写入导出结果。
func writeExportRows(f *os.File, writerCtx *exportWriterContext, columns []string, data []map[string]interface{}) error {
	for _, rowMap := range data {
		record := buildExportRecord(columns, rowMap, writerCtx.format)
		if err := writeExportRow(f, writerCtx, record, rowMap); err != nil {
			return err
		}
	}
	return nil
}

// buildExportRecord 按导出格式将单行转为文本字段。
func buildExportRecord(columns []string, rowMap map[string]interface{}, format string) []string {
	record := make([]string, len(columns))
	for i, col := range columns {
		val := rowMap[col]
		if val == nil {
			record[i] = "NULL"
			continue
		}
		s := fmt.Sprintf("%v", val)
		if format == "md" {
			s = strings.ReplaceAll(s, "|", "\\|")
			s = strings.ReplaceAll(s, "\n", "<br>")
		}
		record[i] = s
	}
	return record
}

// writeExportRow 根据目标格式写入一行数据。
func writeExportRow(f *os.File, writerCtx *exportWriterContext, record []string, rowMap map[string]interface{}) error {
	switch writerCtx.format {
	case "csv", "xlsx":
		return writerCtx.csvWriter.Write(record)
	case "json":
		if !writerCtx.isJSONFirstRow {
			f.WriteString(",\n")
		}
		if err := writerCtx.jsonEncoder.Encode(rowMap); err != nil {
			return err
		}
		writerCtx.isJSONFirstRow = false
		return nil
	case "md":
		_, err := fmt.Fprintf(f, "| %s |\n", strings.Join(record, " | "))
		return err
	default:
		return fmt.Errorf("不支持的导出格式")
	}
}
