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
	"fmt"
	"testing"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
)

// TestGetConnectTimeoutSeconds 测试获取连接超时时间
func TestGetConnectTimeoutSeconds(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		expected int
	}{
		{
			name:     "使用配置的超时时间",
			timeout:  60,
			expected: 60,
		},
		{
			name:     "使用默认超时时间（未设置）",
			timeout:  0,
			expected: defaultConnectTimeoutSeconds,
		},
		{
			name:     "使用默认超时时间（负值）",
			timeout:  -10,
			expected: defaultConnectTimeoutSeconds,
		},
		{
			name:     "使用配置的超时时间（大值）",
			timeout:  300,
			expected: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &connection.ConnectionConfig{
				Timeout: tt.timeout,
			}
			result := getConnectTimeoutSeconds(config)
			if result != tt.expected {
				t.Errorf("getConnectTimeoutSeconds() = %d, 期望 %d", result, tt.expected)
			}
		})
	}
}

// TestGetConnectTimeout 测试获取连接超时 Duration
func TestGetConnectTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		expected time.Duration
	}{
		{
			name:     "60秒超时",
			timeout:  60,
			expected: 60 * time.Second,
		},
		{
			name:     "使用默认值",
			timeout:  0,
			expected: defaultConnectTimeoutSeconds * time.Second,
		},
		{
			name:     "120秒超时",
			timeout:  120,
			expected: 120 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &connection.ConnectionConfig{
				Timeout: tt.timeout,
			}
			result := getConnectTimeout(config)
			if result != tt.expected {
				t.Errorf("getConnectTimeout() = %v, 期望 %v", result, tt.expected)
			}
		})
	}
}

// TestBytesToReadableString 测试字节数组转可读字符串
func TestBytesToReadableString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected interface{}
	}{
		{
			name:     "nil 字节数组",
			input:    nil,
			expected: nil,
		},
		{
			name:     "空字节数组",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "普通字节数组",
			input:    []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f},
			expected: "0x48656c6c6f",
		},
		{
			name:     "二进制数据",
			input:    []byte{0x00, 0xFF, 0x10, 0x20},
			expected: "0x00ff1020",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bytesToReadableString(tt.input)
			if result != tt.expected {
				t.Errorf("bytesToReadableString() = %v, 期望 %v", result, tt.expected)
			}
		})
	}
}

// TestIsBitLikeDBType 测试 BIT 类型检查
func TestIsBitLikeDBType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		expected bool
	}{
		{
			name:     "BIT 类型",
			typeName: "BIT",
			expected: true,
		},
		{
			name:     "VARBIT 类型",
			typeName: "VARBIT",
			expected: true,
		},
		{
			name:     "BIT(1) 类型",
			typeName: "BIT(1)",
			expected: true,
		},
		{
			name:     "BIT(64) 类型",
			typeName: "BIT(64)",
			expected: true,
		},
		{
			name:     "VARCHAR 类型",
			typeName: "VARCHAR",
			expected: false,
		},
		{
			name:     "INT 类型",
			typeName: "INT",
			expected: false,
		},
		{
			name:     "空字符串",
			typeName: "",
			expected: false,
		},
		{
			name:     "TEXT 类型",
			typeName: "TEXT",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBitLikeDBType(tt.typeName)
			if result != tt.expected {
				t.Errorf("isBitLikeDBType(%q) = %v, 期望 %v", tt.typeName, result, tt.expected)
			}
		})
	}
}

// TestBytesToUint64 测试字节数组转 uint64
func TestBytesToUint64(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint64
		valid    bool
	}{
		{
			name:     "空字节数组",
			input:    []byte{},
			expected: 0,
			valid:    false,
		},
		{
			name:     "单字节",
			input:    []byte{0xFF},
			expected: 0xFF,
			valid:    true,
		},
		{
			name:     "双字节",
			input:    []byte{0x01, 0x02},
			expected: 0x0102,
			valid:    true,
		},
		{
			name:     "四字节",
			input:    []byte{0x01, 0x02, 0x03, 0x04},
			expected: 0x01020304,
			valid:    true,
		},
		{
			name:     "八字节（最大）",
			input:    []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			expected: 0xFFFFFFFFFFFFFFFF,
			valid:    true,
		},
		{
			name:     "超过八字节",
			input:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
			expected: 0,
			valid:    false,
		},
		{
			name:     "全零",
			input:    []byte{0x00, 0x00, 0x00, 0x00},
			expected: 0,
			valid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := bytesToUint64(tt.input)
			if valid != tt.valid {
				t.Errorf("bytesToUint64(%v) valid = %v, 期望 %v", tt.input, valid, tt.valid)
			}
			if valid && result != tt.expected {
				t.Errorf("bytesToUint64(%v) = %d, 期望 %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsMostlyPrintable 测试字符串可打印性检查
func TestIsMostlyPrintable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "空字符串",
			input:    "",
			expected: true,
		},
		{
			name:     "纯ASCII文本",
			input:    "Hello, World!",
			expected: true,
		},
		{
			name:     "包含换行符",
			input:    "Line 1\nLine 2\nLine 3",
			expected: true,
		},
		{
			name:     "包含制表符",
			input:    "Column1\tColumn2\tColumn3",
			expected: true,
		},
		{
			name:     "包含回车符",
			input:    "Text\r\n",
			expected: true,
		},
		{
			name:     "Unicode文本",
			input:    "你好，世界！🌍",
			expected: true,
		},
		{
			name:     "90%可打印（边界情况）",
			input:    "aaaaaaaaaab", // 10个字符，10%不可打印
			expected: true,
		},
		{
			name:     "少量不可打印字符",
			input:    "Hello\x00World",
			expected: true, // 11个字符，1个不可打印，约91%可打印
		},
		{
			name:     "大量不可打印字符",
			input:    "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09",
			expected: false, // 除了\t，都是不可打印的
		},
		{
			name:     "混合可打印和控制字符（低于90%）",
			input:    "Text\x00\x01\x02\x03End",
			expected: false, // 11个字符，7个可打印，约64%可打印，低于90%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMostlyPrintable(tt.input)
			if result != tt.expected {
				t.Errorf("isMostlyPrintable(%q) = %v, 期望 %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizeQueryValueWithDBType 测试查询值规范化
func TestNormalizeQueryValueWithDBType(t *testing.T) {
	tests := []struct {
		name             string
		value            interface{}
		databaseTypeName string
		expected         interface{}
	}{
		{
			name:             "字符串类型",
			value:            "hello",
			databaseTypeName: "VARCHAR",
			expected:         "hello",
		},
		{
			name:             "整数类型",
			value:            int64(42),
			databaseTypeName: "INT",
			expected:         int64(42),
		},
		{
			name:             "nil值",
			value:            nil,
			databaseTypeName: "VARCHAR",
			expected:         nil,
		},
		{
			name:             "字节数组（UTF-8文本）",
			value:            []byte("hello"),
			databaseTypeName: "VARCHAR",
			expected:         "hello",
		},
		{
			name:             "空字节数组",
			value:            []byte{},
			databaseTypeName: "VARCHAR",
			expected:         "",
		},
		{
			name:             "nil字节数组",
			value:            ([]byte)(nil),
			databaseTypeName: "VARCHAR",
			expected:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeQueryValueWithDBType(tt.value, tt.databaseTypeName)
			if result != tt.expected {
				t.Errorf("normalizeQueryValueWithDBType(%v, %q) = %v, 期望 %v",
					tt.value, tt.databaseTypeName, result, tt.expected)
			}
		})
	}
}

// TestBytesToDisplayValue 测试字节数组转显示值
func TestBytesToDisplayValue(t *testing.T) {
	tests := []struct {
		name             string
		input            []byte
		databaseTypeName string
		wantType         string
	}{
		{
			name:             "nil字节数组",
			input:            nil,
			databaseTypeName: "VARCHAR",
			wantType:         "nil",
		},
		{
			name:             "空字节数组",
			input:            []byte{},
			databaseTypeName: "VARCHAR",
			wantType:         "string",
		},
		{
			name:             "BIT类型 - 单字节0",
			input:            []byte{0x00},
			databaseTypeName: "BIT",
			wantType:         "int64",
		},
		{
			name:             "BIT类型 - 单字节1",
			input:            []byte{0x01},
			databaseTypeName: "BIT",
			wantType:         "int64",
		},
		{
			name:             "BIT类型 - 多字节",
			input:            []byte{0x01, 0x02, 0x03, 0x04},
			databaseTypeName: "BIT",
			wantType:         "int64",
		},
		{
			name:             "UTF-8文本",
			input:            []byte("Hello, 世界!"),
			databaseTypeName: "VARCHAR",
			wantType:         "string",
		},
		{
			name:             "二进制数据",
			input:            []byte{0x00, 0xFF, 0x10, 0x20},
			databaseTypeName: "BLOB",
			wantType:         "string",
		},
		{
			name:             "BIT(1) 类型 - 0",
			input:            []byte{0x00},
			databaseTypeName: "",
			wantType:         "int64",
		},
		{
			name:             "BIT(1) 类型 - 1",
			input:            []byte{0x01},
			databaseTypeName: "",
			wantType:         "int64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bytesToDisplayValue(tt.input, tt.databaseTypeName)
			if result == nil {
				if tt.wantType != "nil" {
					t.Errorf("bytesToDisplayValue() = nil, 期望类型 %s", tt.wantType)
				}
			} else {
				typeMatch := false
				switch tt.wantType {
				case "string":
					_, typeMatch = result.(string)
				case "int64":
					_, typeMatch = result.(int64)
				case "nil":
					typeMatch = (result == nil)
				}
				if !typeMatch {
					t.Errorf("bytesToDisplayValue() = %v (类型 %T), 期望类型 %s",
						result, result, tt.wantType)
				}
			}
		})
	}
}

// TestScanRows 测试扫描SQL行（需要模拟 rows）
func TestScanRows(t *testing.T) {
	// 这个测试需要一个模拟的 sql.Rows
	// 由于 sql.Rows 是接口，但实际使用是具体的实现，这里只测试空情况
	t.Skip("需要模拟数据库连接来完整测试 scanRows")
}

// BenchmarkBytesToUint64 基准测试
func BenchmarkBytesToUint64(b *testing.B) {
	input := []byte{0x01, 0x02, 0x03, 0x04}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bytesToUint64(input)
	}
}

// BenchmarkIsMostlyPrintable 基准测试
func BenchmarkIsMostlyPrintable(b *testing.B) {
	input := "Hello, World! 你好，世界！"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isMostlyPrintable(input)
	}
}

// BenchmarkBytesToDisplayValue 基准测试
func BenchmarkBytesToDisplayValue(b *testing.B) {
	input := []byte("Hello, World!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bytesToDisplayValue(input, "VARCHAR")
	}
}

// ExampleConnectionConfig 示例：创建数据库连接配置
func ExampleConnectionConfig() {
	config := &connection.ConnectionConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "testdb",
		Timeout:  30,
		UseSSH:   false,
	}
	fmt.Printf("数据库类型: %s\n", config.Type)
	fmt.Printf("主机: %s:%d\n", config.Host, config.Port)
}

// ExampleColumnDefinition 示例：列定义
func ExampleColumnDefinition() {
	col := &connection.ColumnDefinition{
		Name:     "id",
		Type:     "INT",
		Nullable: "NO",
		Key:      "PRI",
		Extra:    "auto_increment",
		Comment:  "主键ID",
	}
	fmt.Printf("列名: %s\n", col.Name)
	fmt.Printf("类型: %s\n", col.Type)
	fmt.Printf("可空: %s\n", col.Nullable)
}
