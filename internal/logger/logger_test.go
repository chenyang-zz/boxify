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

package logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// resetLogger 重置全局 logger 状态（仅用于测试）
func resetLogger() {
	mu.Lock()
	defer mu.Unlock()
	logger = nil
	sugar = nil
	logPath = ""
	initialized = false
}

// TestInit 测试日志系统的初始化
func TestInit(t *testing.T) {
	// 保存原始环境变量
	oldEnv := os.Getenv(envLogDir)
	oldLevel := os.Getenv("BOXIFY_LOG_LEVEL")
	defer func() {
		if oldEnv == "" {
			os.Unsetenv(envLogDir)
		} else {
			os.Setenv(envLogDir, oldEnv)
		}
		if oldLevel == "" {
			os.Unsetenv("BOXIFY_LOG_LEVEL")
		} else {
			os.Setenv("BOXIFY_LOG_LEVEL", oldLevel)
		}
		resetLogger()
	}()

	// 创建临时目录用于测试
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)

	// 调用 Init
	Init(context.Background())

	// 验证日志文件已创建
	expectedPath := filepath.Join(tempDir, "boxify.log")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("日志文件未被创建: %s", expectedPath)
	}

	// 验证 Path() 返回正确的路径
	if path := Path(); path != expectedPath {
		t.Errorf("Path() 返回错误的路径，得到: %s，期望: %s", path, expectedPath)
	}
}

// TestInitMultipleCalls 测试多次调用 Init 是否安全
func TestInitMultipleCalls(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	// 多次调用 Init 应该是安全的
	for i := 0; i < 10; i++ {
		Init(context.Background())
	}

	// 验证只有一个日志文件被创建
	expectedPath := filepath.Join(tempDir, "boxify.log")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("日志文件未被创建: %s", expectedPath)
	}
}

// TestPath 测试 Path 函数
func TestPath(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	path := Path()
	expectedPath := filepath.Join(tempDir, "boxify.log")

	if path != expectedPath {
		t.Errorf("Path() 返回错误的路径，得到: %s，期望: %s", path, expectedPath)
	}
}

// TestInfof 测试 Info 日志级别
func TestInfof(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	testMsg := "这是一条测试信息"
	Infof(testMsg)

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	// 读取日志文件并验证内容
	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "INFO") {
		t.Errorf("日志内容缺少 INFO 标记")
	}
	if !strings.Contains(logContent, testMsg) {
		t.Errorf("日志内容缺少测试信息: %s", testMsg)
	}
}

// TestInfofWithFormat 测试 Infof 的格式化功能
func TestInfofWithFormat(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	name := "测试用户"
	count := 42
	Infof("用户 %s 登录了系统，计数: %d", name, count)

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "INFO") {
		t.Errorf("日志内容缺少 INFO 标记")
	}
	if !strings.Contains(logContent, name) {
		t.Errorf("日志内容缺少用户名: %s", name)
	}
	if !strings.Contains(logContent, fmt.Sprintf("%d", count)) {
		t.Errorf("日志内容缺少计数: %d", count)
	}
}

// TestInfoStructured 测试结构化 Info 日志
func TestInfoStructured(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	testMsg := "结构化日志测试"
	Info(testMsg, String("user", "test"), Int("count", 123))

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "INFO") {
		t.Errorf("日志内容缺少 INFO 标记")
	}
	if !strings.Contains(logContent, testMsg) {
		t.Errorf("日志内容缺少测试信息: %s", testMsg)
	}
	if !strings.Contains(logContent, "user") || !strings.Contains(logContent, "test") {
		t.Errorf("日志内容缺少结构化字段")
	}
	if !strings.Contains(logContent, "count") || !strings.Contains(logContent, "123") {
		t.Errorf("日志内容缺少计数字段")
	}
}

// TestWarnf 测试 Warn 日志级别
func TestWarnf(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	testMsg := "这是一条警告信息"
	Warnf(testMsg)

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "WARN") {
		t.Errorf("日志内容缺少 WARN 标记")
	}
	if !strings.Contains(logContent, testMsg) {
		t.Errorf("日志内容缺少警告信息: %s", testMsg)
	}
}

// TestDebugf 测试 Debug 日志级别
func TestDebugf(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	os.Setenv("BOXIFY_LOG_LEVEL", "DEBUG")
	defer os.Unsetenv(envLogDir)
	defer os.Unsetenv("BOXIFY_LOG_LEVEL")
	defer resetLogger()

	testMsg := "这是一条调试信息"
	Debugf(testMsg)

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "DEBUG") {
		t.Errorf("日志内容缺少 DEBUG 标记")
	}
	if !strings.Contains(logContent, testMsg) {
		t.Errorf("日志内容缺少调试信息: %s", testMsg)
	}
}

// TestErrorf 测试 Error 日志级别
func TestErrorf(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	testMsg := "这是一条错误信息"
	Errorf(testMsg)

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "ERROR") {
		t.Errorf("日志内容缺少 ERROR 标记")
	}
	if !strings.Contains(logContent, testMsg) {
		t.Errorf("日志内容缺少错误信息: %s", testMsg)
	}
}

// TestErrorStructured 测试结构化 Error 日志
func TestErrorStructured(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	testErr := errors.New("测试错误")
	testMsg := "操作失败"
	Error(testMsg, Err(testErr), String("operation", "test"))

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "ERROR") {
		t.Errorf("日志内容缺少 ERROR 标记")
	}
	if !strings.Contains(logContent, testMsg) {
		t.Errorf("日志内容缺少消息: %s", testMsg)
	}
	if !strings.Contains(logContent, "测试错误") {
		t.Errorf("日志内容缺少错误信息")
	}
}

// TestErrorfWithTrace 测试带错误链的错误日志
func TestErrorfWithTrace(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	// 创建一个错误链
	baseErr := errors.New("基础错误")
	midErr := fmt.Errorf("中间错误: %w", baseErr)
	topErr := fmt.Errorf("顶层错误: %w", midErr)

	ErrorfWithTrace(topErr, "操作失败")

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "ERROR") {
		t.Errorf("日志内容缺少 ERROR 标记")
	}
	if !strings.Contains(logContent, "操作失败") {
		t.Errorf("日志内容缺少消息")
	}
	if !strings.Contains(logContent, "错误链：") {
		t.Errorf("日志内容缺少错误链标记")
	}
	// 验证错误链的各个层级都被记录
	if !strings.Contains(logContent, "顶层错误") {
		t.Errorf("日志内容缺少顶层错误")
	}
	if !strings.Contains(logContent, "中间错误") {
		t.Errorf("日志内容缺少中间错误")
	}
	if !strings.Contains(logContent, "基础错误") {
		t.Errorf("日志内容缺少基础错误")
	}
}

// TestErrorfWithTraceNilError 测试传入 nil 错误的情况
func TestErrorfWithTraceNilError(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	ErrorfWithTrace(nil, "没有错误的消息")

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "ERROR") {
		t.Errorf("日志内容缺少 ERROR 标记")
	}
	if !strings.Contains(logContent, "没有错误的消息") {
		t.Errorf("日志内容缺少消息")
	}
	// 不应该包含错误链
	if strings.Contains(logContent, "错误链：") {
		t.Errorf("nil 错误不应该包含错误链")
	}
}

// TestErrorChain 测试错误链格式化
func TestErrorChain(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil 错误",
			err:      nil,
			expected: "<nil>",
		},
		{
			name:     "单个错误",
			err:      errors.New("错误1"),
			expected: "错误1",
		},
		{
			name:     "错误链",
			err:      fmt.Errorf("错误1: %w", errors.New("错误2")),
			expected: "错误1: 错误2 -> 错误2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ErrorChain(tt.err)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("ErrorChain() = %s, 期望包含: %s", result, tt.expected)
			}
		})
	}
}

// TestErrorChainDeduplication 测试错误链去重功能
func TestErrorChainDeduplication(t *testing.T) {
	// 创建一个会完全重复相同错误消息的错误链
	baseMsg := "连接失败"
	baseErr := errors.New(baseMsg)

	// 创建多个包装，但都使用相同的错误消息字符串
	err1 := baseErr
	_ = fmt.Errorf("操作失败: %w", err1) // 这个错误不会用在最终链中
	err3 := fmt.Errorf("任务失败: %w", err1) // 再次包装 err1

	result := ErrorChain(err3)

	// 验证错误链包含所有层级的消息
	if !strings.Contains(result, "任务失败") {
		t.Errorf("错误链缺少顶层错误")
	}
	if !strings.Contains(result, "连接失败") {
		t.Errorf("错误链缺少底层错误")
	}

	// 验证错误链的格式正确（使用 -> 分隔）
	if !strings.Contains(result, " -> ") {
		t.Errorf("错误链应该包含箭头分隔符")
	}
}

// TestErrorChainTruncation 测试错误链截断
func TestErrorChainTruncation(t *testing.T) {
	// 创建一个超过20层的错误链
	err := errors.New("底层错误")
	for i := 0; i < 25; i++ {
		err = fmt.Errorf("第%d层: %w", i, err)
	}

	result := ErrorChain(err)
	if !strings.Contains(result, "错误链过长，已截断") {
		t.Errorf("超长错误链未被截断")
	}
}

// TestHelperMethods 测试便捷字段方法
func TestHelperMethods(t *testing.T) {
	tests := []struct {
		name   string
		field  zap.Field
		key    string
		verify func(string) bool
	}{
		{
			name:   "String",
			field:  String("key", "value"),
			key:    "key",
			verify: func(s string) bool { return strings.Contains(s, "value") },
		},
		{
			name:   "Int",
			field:  Int("count", 42),
			key:    "count",
			verify: func(s string) bool { return strings.Contains(s, "42") },
		},
		{
			name:   "Err",
			field:  Err(errors.New("test error")),
			key:    "error",
			verify: func(s string) bool { return strings.Contains(s, "test error") },
		},
		{
			name:   "Duration",
			field:  Duration("elapsed", time.Second),
			key:    "elapsed",
			verify: func(s string) bool { return strings.Contains(s, "1s") || strings.Contains(s, "1000") },
		},
		{
			name:   "Any",
			field:  Any("data", map[string]int{"a": 1}),
			key:    "data",
			verify: func(s string) bool { return strings.Contains(s, "map") || strings.Contains(s, "a") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证字段可以正常创建
			if tt.field.Key != tt.key {
				t.Errorf("字段 key 不匹配，得到: %s，期望: %s", tt.field.Key, tt.key)
			}
		})
	}
}

// TestGetZapLogger 测试获取底层 zap logger
func TestGetZapLogger(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	zapLogger := GetZapLogger()
	if zapLogger == nil {
		t.Error("GetZapLogger 返回 nil")
	}

	// 验证返回的 logger 可以正常使用
	zapLogger.Info("test", zap.String("key", "value"))
}

// TestClose 测试关闭日志系统
func TestClose(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	// 初始化并记录一些日志
	Init(context.Background())
	Infof("关闭前的日志")

	// 关闭日志系统
	Close()

	// 验证日志文件仍然存在
	logFilePath := Path()
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("日志文件在 Close 后不应该被删除")
	}
}

// TestConcurrentLogging 测试并发日志记录
func TestConcurrentLogging(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	// 使用 WaitGroup 等待所有 goroutine 完成
	var wg sync.WaitGroup
	numGoroutines := 50
	logsPerGoroutine := 20

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				Infof("Goroutine %d, 日志 %d", id, j)
			}
		}(i)
	}

	wg.Wait()

	// 等待日志写入
	time.Sleep(200 * time.Millisecond)

	// 验证日志文件存在且不为空
	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	// 验证至少有一些日志被记录
	infoCount := strings.Count(logContent, "INFO")
	if infoCount < numGoroutines*logsPerGoroutine/2 {
		t.Errorf("日志记录不完整，期望至少 %d 条，实际得到 %d 条", numGoroutines*logsPerGoroutine/2, infoCount)
	}
}

// TestNewLogger 测试 NewLogger 函数
func TestNewLogger(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "基本配置",
			config: &Config{
				Level:      "INFO",
				Format:     "text",
				OutputPath:  logPath,
				MaxSize:     1,
				MaxBackups:  3,
				MaxAge:      7,
				Compress:    false,
			},
			wantErr: false,
		},
		{
			name: "JSON 格式",
			config: &Config{
				Level:      "DEBUG",
				Format:     "json",
				OutputPath:  logPath + ".json",
				MaxSize:     1,
				MaxBackups:  3,
				MaxAge:      7,
				Compress:    false,
			},
			wantErr: false,
		},
		{
			name: "无效日志级别",
			config: &Config{
				Level:      "INVALID",
				Format:     "text",
				OutputPath:  logPath + ".invalid",
				MaxSize:     1,
				MaxBackups:  3,
				MaxAge:      7,
				Compress:    false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if logger == nil {
					t.Error("NewLogger() 返回 nil logger")
				}
				// 测试 logger 可以正常使用
				logger.Info("test", zap.String("message", "test message"))
				logger.Sync()
			}
		})
	}
}

// TestLoadConfig 测试 loadConfig 函数
func TestLoadConfig(t *testing.T) {
	oldLevel := os.Getenv("BOXIFY_LOG_LEVEL")
	defer func() {
		if oldLevel == "" {
			os.Unsetenv("BOXIFY_LOG_LEVEL")
		} else {
			os.Setenv("BOXIFY_LOG_LEVEL", oldLevel)
		}
	}()

	tests := []struct {
		name           string
		envLevel       string
		expectedLevel  string
		expectedFormat string
	}{
		{
			name:           "默认配置",
			envLevel:       "",
			expectedLevel:  "INFO", // 默认生产环境
			expectedFormat: "text",
		},
		{
			name:           "环境变量覆盖",
			envLevel:       "WARN",
			expectedLevel:  "WARN",
			expectedFormat: "text",
		},
		{
			name:           "环境变量 INFO",
			envLevel:       "INFO",
			expectedLevel:  "INFO",
			expectedFormat: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envLevel != "" {
				os.Setenv("BOXIFY_LOG_LEVEL", tt.envLevel)
			} else {
				os.Unsetenv("BOXIFY_LOG_LEVEL")
			}

			config := loadConfig(context.Background())
			if config.Level != tt.expectedLevel {
				t.Errorf("loadConfig() Level = %s，期望 %s", config.Level, tt.expectedLevel)
			}
			if config.Format != tt.expectedFormat {
				t.Errorf("loadConfig() Format = %s，期望 %s", config.Format, tt.expectedFormat)
			}
		})
	}
}

// TestGetLogPath 测试 getLogPath 函数
func TestGetLogPath(t *testing.T) {
	oldEnv := os.Getenv(envLogDir)
	defer func() {
		if oldEnv == "" {
			os.Unsetenv(envLogDir)
		} else {
			os.Setenv(envLogDir, oldEnv)
		}
	}()

	tests := []struct {
		name          string
		envDir        string
		expectedInDir string
	}{
		{
			name:          "使用环境变量",
			envDir:        "/tmp/test_logs",
			expectedInDir: "/tmp/test_logs",
		},
		{
			name:          "使用默认路径",
			envDir:        "",
			expectedInDir: "boxify",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envDir != "" {
				os.Setenv(envLogDir, tt.envDir)
			} else {
				os.Unsetenv(envLogDir)
			}

			path := getLogPath()
			if !strings.Contains(path, tt.expectedInDir) {
				t.Errorf("getLogPath() = %s，期望包含 %s", path, tt.expectedInDir)
			}
		})
	}
}

// TestInitWithInvalidDir 测试使用无效目录初始化
func TestInitWithInvalidDir(t *testing.T) {
	// 设置一个无效的环境变量（例如：一个无法创建的路径）
	// 在 Unix 系统上，使用 /root 可能需要权限
	// 这里我们使用一个不存在的路径
	invalidPath := "/this/path/does/not/exist/and/cannot/be/created/12345"
	os.Setenv(envLogDir, invalidPath)
	defer os.Unsetenv(envLogDir)
	defer resetLogger()

	// Init 应该能够处理错误
	// 不应该 panic
	Init(context.Background())
	Infof("测试无效目录的情况")
}
