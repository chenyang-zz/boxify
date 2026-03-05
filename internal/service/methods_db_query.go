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
	"fmt"
	"strings"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/chenyang-zz/boxify/internal/db"
	"github.com/chenyang-zz/boxify/internal/utils"
)

// DBQuery 执行 SQL 并返回查询结果或受影响行数。
func (a *DatabaseService) DBQuery(config *connection.ConnectionConfig, dbName, query string, args []any) *connection.QueryResult {
	runConfig := normalizeRunConfig(config, dbName)

	dbInst, err := a.getDatabase(runConfig)
	if err != nil {
		a.Logger().Error("DBQuery 获取连接失败", "error", err, "summary", db.FormatConnSummary(runConfig))
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	query = sanitizeSQLForPgLike(runConfig.Type, query)
	timeoutSeconds := runConfig.Timeout
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	ctx, cancel := utils.ContextWithTimeout(time.Duration(timeoutSeconds) * time.Second)
	defer cancel()

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
			a.Logger().Error("DBQuery 查询失败", "error", err, "summary", db.FormatConnSummary(runConfig), "snippet", sqlSnippet(query))
			return &connection.QueryResult{Success: false, Message: err.Error()}
		}
		return &connection.QueryResult{Success: true, Message: "查询成功", Data: data, Fields: columns}
	}

	var affected int64
	if e, ok := dbInst.(interface {
		ExecContext(context.Context, string) (int64, error)
	}); ok {
		affected, err = e.ExecContext(ctx, query)
	} else {
		affected, err = dbInst.Exec(query)
	}
	if err != nil {
		a.Logger().Error("DBQuery 执行失败", "error", err, "summary", db.FormatConnSummary(runConfig), "snippet", sqlSnippet(query))
		return &connection.QueryResult{Success: false, Message: err.Error()}
	}

	return &connection.QueryResult{
		Success: true,
		Message: fmt.Sprintf("执行成功，受影响的行数: %d", affected),
		Data:    map[string]int64{"affectedRows": affected},
	}
}

// sqlSnippet 返回SQL查询的简短片段，用于日志输出，限制长度以避免过长。
func sqlSnippet(query string) string {
	q := strings.TrimSpace(query)
	const max = 200
	if len(q) <= max {
		return q
	}
	return q[:max] + "..."
}
