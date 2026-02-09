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
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/chenyang-zz/boxify/internal/connection"
)

// 默认连接超时时间（秒）
const defaultConnectTimeoutSeconds = 30

// getConnectTimeoutSeconds从连接配置中获取连接超时时间
// 如果未设置或无效，则返回默认值
func getConnectTimeoutSeconds(config *connection.ConnectionConfig) int {
	timeoutSeconds := config.Timeout
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultConnectTimeoutSeconds
	}
	return timeoutSeconds
}

// getConnectTimeout返回连接超时时间的Duration表示
func getConnectTimeout(config *connection.ConnectionConfig) time.Duration {
	return time.Duration(getConnectTimeoutSeconds(config)) * time.Second
}

// scanRows是一个实用函数，用于将sql.Rows转换为更通用的格式，适用于不同数据库类型
func scanRows(rows *sql.Rows) ([]map[string]interface{}, []string, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	colTypes, err := rows.ColumnTypes()
	if err != nil || len(colTypes) != len(columns) {
		colTypes = nil // 如果无法获取列类型，继续但不使用类型信息
	}

	resultData := make([]map[string]interface{}, 0)

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err = rows.Scan(valuePtrs...); err != nil {
			continue
		}

		entry := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			dbTypeName := ""
			if colTypes != nil && i < len(colTypes) && colTypes[i] != nil {
				dbTypeName = colTypes[i].DatabaseTypeName()
			}
			entry[col] = normalizeQueryValueWithDBType(values[i], dbTypeName)
		}
		resultData = append(resultData, entry)
	}

	if err := rows.Err(); err != nil {
		return resultData, columns, err
	}

	return resultData, columns, nil
}

// normalizeQueryValueWithDBType 根据数据库类型对查询结果中的值进行规范化处理
func normalizeQueryValueWithDBType(v interface{}, databaseTypeName string) interface{} {
	if b, ok := v.([]byte); ok {
		return bytesToDisplayValue(b, databaseTypeName)
	}
	return v
}

// bytesToDisplayValue 将字节数组转换为适合显示的值，考虑数据库类型和内容
func bytesToDisplayValue(b []byte, databaseTypeName string) interface{} {
	if b == nil {
		return nil
	}
	if len(b) == 0 {
		return ""
	}

	dbType := strings.ToUpper(strings.TrimSpace(databaseTypeName))
	if isBitLikeDBType(dbType) {
		if u, ok := bytesToUint64(b); ok {
			// JS编号精度有限；
			// 保持大的位掩码为字符串
			const maxSafeInteger = 9007199254740991 // 2^53 - 1
			if u <= maxSafeInteger {
				return int64(u)
			}
			return fmt.Sprintf("%d", u)
		}
	}

	if utf8.Valid(b) {
		s := string(b)
		if isMostlyPrintable(s) {
			return s
		}
	}

	// 回退：一些驱动返回BIT(1)为[]byte{0} / []byte{1}，没有类型info
	if dbType == "" && len(b) == 1 && (b[0] == 0 || b[0] == 1) {
		return int64(b[0])
	}

	return bytesToReadableString(b)
}

// bytesToReadableString 将字节数组转换为可读字符串，非UTF-8内容以十六进制表示
func bytesToReadableString(b []byte) interface{} {
	if b == nil {
		return nil
	}
	if len(b) == 0 {
		return ""
	}
	return "0x" + hex.EncodeToString(b)
}

// isBitLikeDBType 检查数据库类型名称是否表示类似于BIT的类型
func isBitLikeDBType(typeName string) bool {
	if typeName == "" {
		return false
	}
	switch typeName {
	case "BIT", "VARBIT":
		return true
	default:
	}
	return strings.HasPrefix(typeName, "BIT")
}

// bytesToUint64 尝试将字节数组转换为uint64，适用于BIT类型数据
func bytesToUint64(b []byte) (uint64, bool) {
	if len(b) == 0 || len(b) > 8 {
		return 0, false
	}
	var u uint64
	for _, v := range b {
		u = (u << 8) | uint64(v)
	}
	return u, true
}

// isMostlyPrintable检 查字符串中是否大部分字符都是可打印的，允许少量控制字符
func isMostlyPrintable(s string) bool {
	if s == "" {
		return true
	}

	total := 0
	printable := 0
	for _, r := range s {
		total++
		switch r {
		case '\n', '\r', '\t':
			printable++
			continue
		default:
		}
		if unicode.IsPrint(r) {
			printable++
		}
	}

	// 允许少量不可见字符，避免把正常文本误判为二进制。
	return printable*100 >= total*90
}
