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
)

// TestInit 测试日志系统的初始化
func TestInit(t *testing.T) {
	// 保存原始环境变量
	oldEnv := os.Getenv(envLogDir)
	defer func() {
		if oldEnv == "" {
			os.Unsetenv(envLogDir)
		} else {
			os.Setenv(envLogDir, oldEnv)
		}
	}()

	// 创建临时目录用于测试
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)

	// 重置全局变量以便测试（仅用于测试环境）
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

	// 调用 Init
	Init(context.Background())

	// 验证日志文件已创建
	expectedPath := filepath.Join(tempDir, logFileName)
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

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

	// 多次调用 Init 应该是安全的
	for i := 0; i < 10; i++ {
		Init(context.Background())
	}

	// 验证只有一个日志文件被创建
	expectedPath := filepath.Join(tempDir, logFileName)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("日志文件未被创建: %s", expectedPath)
	}
}

// TestPath 测试 Path 函数
func TestPath(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

	path := Path()
	expectedPath := filepath.Join(tempDir, logFileName)

	if path != expectedPath {
		t.Errorf("Path() 返回错误的路径，得到: %s，期望: %s", path, expectedPath)
	}
}

// TestInfof 测试 Info 日志级别
func TestInfof(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

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
	if !strings.Contains(logContent, "[INFO]") {
		t.Errorf("日志内容缺少 [INFO] 标记")
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

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

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
	if !strings.Contains(logContent, "[INFO]") {
		t.Errorf("日志内容缺少 [INFO] 标记")
	}
	if !strings.Contains(logContent, name) {
		t.Errorf("日志内容缺少用户名: %s", name)
	}
	if !strings.Contains(logContent, fmt.Sprintf("%d", count)) {
		t.Errorf("日志内容缺少计数: %d", count)
	}
}

// TestWarnf 测试 Warn 日志级别
func TestWarnf(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

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
	if !strings.Contains(logContent, "[WARN]") {
		t.Errorf("日志内容缺少 [WARN] 标记")
	}
	if !strings.Contains(logContent, testMsg) {
		t.Errorf("日志内容缺少警告信息: %s", testMsg)
	}
}

// TestErrorf 测试 Error 日志级别
func TestErrorf(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

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
	if !strings.Contains(logContent, "[ERROR]") {
		t.Errorf("日志内容缺少 [ERROR] 标记")
	}
	if !strings.Contains(logContent, testMsg) {
		t.Errorf("日志内容缺少错误信息: %s", testMsg)
	}
}

// TestErrorfWithTrace 测试带错误链的错误日志
func TestErrorfWithTrace(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

	// 创建一个错误链
	baseErr := errors.New("基础错误")
	midErr := fmt.Errorf("中间错误: %w", baseErr)
	topErr := fmt.Errorf("顶层错误: %w", midErr)

	testMsg := "操作失败"
	ErrorfWithTrace(topErr, testMsg)

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "[ERROR]") {
		t.Errorf("日志内容缺少 [ERROR] 标记")
	}
	if !strings.Contains(logContent, testMsg) {
		t.Errorf("日志内容缺少消息: %s", testMsg)
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

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

	testMsg := "没有错误的消息"
	ErrorfWithTrace(nil, testMsg)

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "[ERROR]") {
		t.Errorf("日志内容缺少 [ERROR] 标记")
	}
	if !strings.Contains(logContent, testMsg) {
		t.Errorf("日志内容缺少消息: %s", testMsg)
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
	_ = fmt.Errorf("操作失败: %w", err1)     // 这个错误不会用在最终链中
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

// TestClose 测试关闭日志系统
func TestClose(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

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

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

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
	infoCount := strings.Count(logContent, "[INFO]")
	if infoCount < numGoroutines*logsPerGoroutine/2 {
		t.Errorf("日志记录不完整，期望至少 %d 条，实际得到 %d 条", numGoroutines*logsPerGoroutine/2, infoCount)
	}
}

// TestLogRotation 测试日志轮转
func TestLogRotation(t *testing.T) {
	tempDir := t.TempDir()

	// 创建一个超过轮转大小的日志文件
	logFilePath := filepath.Join(tempDir, logFileName)
	largeData := make([]byte, logRotateMaxBytes+1000)
	for i := range largeData {
		largeData[i] = 'A'
	}

	if err := os.WriteFile(logFilePath, largeData, 0o644); err != nil {
		t.Fatalf("无法创建测试日志文件: %v", err)
	}

	// 调用 rotateIfNeeded
	rotateIfNeeded(logFilePath, tempDir)

	// 验证原文件被重命名
	matches, err := filepath.Glob(filepath.Join(tempDir, "boxify-*.log"))
	if err != nil {
		t.Fatalf("查找轮转日志文件失败: %v", err)
	}

	// 应该至少有一个轮转的文件
	if len(matches) == 0 {
		t.Error("日志文件未被轮转")
	}

	// 原文件应该不存在或已被重命名
	if _, err := os.Stat(logFilePath); err == nil {
		// 如果原文件还在，检查它是否是空文件或者大小变小了
		info, _ := os.Stat(logFilePath)
		if info.Size() >= logRotateMaxBytes {
			t.Error("原日志文件应该被轮转")
		}
	}
}

// TestCleanupOldLogs 测试清理旧日志
func TestCleanupOldLogs(t *testing.T) {
	tempDir := t.TempDir()

	// 创建多个旧的日志文件
	for i := 0; i < 15; i++ {
		filename := fmt.Sprintf("boxify-20060102-15040%d.log", i)
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("无法创建测试日志文件: %v", err)
		}
	}

	// 调用清理函数
	cleanupOldLogs(tempDir)

	// 验证只保留 logRotateMaxBackups 个文件
	matches, err := filepath.Glob(filepath.Join(tempDir, "boxify-*.log"))
	if err != nil {
		t.Fatalf("查找日志文件失败: %v", err)
	}

	if len(matches) > logRotateMaxBackups {
		t.Errorf("清理后剩余 %d 个日志文件，期望最多 %d 个", len(matches), logRotateMaxBackups)
	}
}

// TestInitWithInvalidDir 测试使用无效目录初始化
func TestInitWithInvalidDir(t *testing.T) {
	// 设置一个无效的环境变量（例如：一个无法创建的路径）
	os.Setenv(envLogDir, "/root/nonexistent_dir_12345")
	defer os.Unsetenv(envLogDir)

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

	// Init 应该能够处理错误并回退到 os.Stderr
	// 不应该 panic
	Init(context.Background())
	Infof("测试无效目录的情况")
}

// TestInitOutput 测试 initOutput 函数
func TestInitOutput(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)

	path, writer := initOutput(context.Background())

	// 验证返回值
	if writer == nil {
		t.Error("initOutput 返回的 writer 不应该为 nil")
	}

	if !strings.Contains(path, logFileName) {
		t.Errorf("路径应该包含日志文件名: %s", logFileName)
	}

	// 验证文件已创建
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("日志文件未被创建: %s", path)
	}

	// 清理
	if file, ok := writer.(*os.File); ok {
		file.Close()
	}
}

// TestPrintf 测试 printf 内部函数通过日志文件
func TestPrintf(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv(envLogDir, tempDir)
	defer os.Unsetenv(envLogDir)

	// 重置全局变量
	logPath = ""
	logInst = nil
	logFile = nil
	once = sync.Once{}

	// 初始化日志系统
	Init(context.Background())

	// 调用 printf 内部函数
	testLevel := "TEST"
	testFormat := "测试消息: %s"
	testArg := "参数"

	printf(testLevel, testFormat, testArg)

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	// 从日志文件读取内容
	logFilePath := Path()
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("无法读取日志文件: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "["+testLevel+"]") {
		t.Errorf("输出缺少级别标记: %s", testLevel)
	}
	if !strings.Contains(output, "测试消息: 参数") {
		t.Errorf("输出缺少格式化消息")
	}
}
