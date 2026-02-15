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
	"log/slog"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestSlogLevelToZapLevel 测试日志级别转换
func TestSlogLevelToZapLevel(t *testing.T) {
	tests := []struct {
		name     string
		slogLevel slog.Level
		zapLevel zapcore.Level
		wantErr  bool
	}{
		{
			name:     "Debug level",
			slogLevel: slog.LevelDebug,
			zapLevel: zapcore.DebugLevel,
			wantErr:  false,
		},
		{
			name:     "Info level",
			slogLevel: slog.LevelInfo,
			zapLevel: zapcore.InfoLevel,
			wantErr:  false,
		},
		{
			name:     "Warn level",
			slogLevel: slog.LevelWarn,
			zapLevel: zapcore.WarnLevel,
			wantErr:  false,
		},
		{
			name:     "Error level",
			slogLevel: slog.LevelError,
			zapLevel: zapcore.ErrorLevel,
			wantErr:  false,
		},
		{
			name:     "Unknown level",
			slogLevel: slog.Level(100),
			zapLevel: zapcore.InfoLevel,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zapLevel, err := slogLevelToZapLevel(tt.slogLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("slogLevelToZapLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && zapLevel != tt.zapLevel {
				t.Errorf("slogLevelToZapLevel() = %v, want %v", zapLevel, tt.zapLevel)
			}
		})
	}
}

// TestAttrToField 测试 Attr 转 Field
func TestAttrToField(t *testing.T) {
	tests := []struct {
		name  string
		attr  slog.Attr
		check func(zap.Field) bool
	}{
		{
			name: "String attribute",
			attr: slog.String("key", "value"),
			check: func(f zap.Field) bool {
				return f.Key == "key" && f.Type == zapcore.StringType && f.String == "value"
			},
		},
		{
			name: "Int64 attribute",
			attr: slog.Int64("count", 42),
			check: func(f zap.Field) bool {
				return f.Key == "count" && f.Type == zapcore.Int64Type && f.Integer == 42
			},
		},
		{
			name: "Float64 attribute",
			attr: slog.Float64("price", 3.14),
			check: func(f zap.Field) bool {
				return f.Key == "price" && f.Type == zapcore.Float64Type
			},
		},
		{
			name: "Bool attribute",
			attr: slog.Bool("enabled", true),
			check: func(f zap.Field) bool {
				return f.Key == "enabled" && f.Type == zapcore.BoolType && f.Integer == 1
			},
		},
		{
			name: "Time attribute",
			attr: slog.Time("timestamp", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
			check: func(f zap.Field) bool {
				return f.Key == "timestamp" && f.Type == zapcore.TimeType
			},
		},
		{
			name: "Duration attribute",
			attr: slog.Duration("elapsed", time.Second),
			check: func(f zap.Field) bool {
				return f.Key == "elapsed" && f.Type == zapcore.DurationType
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := attrToField(tt.attr)
			if !tt.check(field) {
				t.Errorf("attrToField() = %+v, check failed", field)
			}
		})
	}
}

// TestZapHandlerEnabled 测试 Enabled 方法
func TestZapHandlerEnabled(t *testing.T) {
	observedZapCore, logs := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(observedZapCore)
	handler := newZapHandler(zapLogger)

	tests := []struct {
		name    string
		level   slog.Level
		enabled bool
	}{
		{
			name:    "Info level enabled",
			level:   slog.LevelInfo,
			enabled: true,
		},
		{
			name:    "Debug level disabled",
			level:   slog.LevelDebug,
			enabled: false,
		},
		{
			name:    "Warn level enabled",
			level:   slog.LevelWarn,
			enabled: true,
		},
		{
			name:    "Error level enabled",
			level:   slog.LevelError,
			enabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enabled := handler.Enabled(context.Background(), tt.level)
			if enabled != tt.enabled {
				t.Errorf("Enabled() = %v, want %v", enabled, tt.enabled)
			}
		})
	}

	_ = logs // 避免未使用警告
}

// TestZapHandlerHandle 测试 Handle 方法
func TestZapHandlerHandle(t *testing.T) {
	observedZapCore, logs := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(observedZapCore)
	handler := newZapHandler(zapLogger)

	// 创建日志记录
	now := time.Now()
	record := slog.NewRecord(now, slog.LevelInfo, "test message", 0)

	// 添加属性
	record.AddAttrs(slog.String("key1", "value1"))
	record.AddAttrs(slog.Int64("key2", 42))

	// 处理日志
	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	// 验证日志输出
	allLogs := logs.TakeAll()
	if len(allLogs) != 1 {
		t.Fatalf("got %d logs, want 1", len(allLogs))
	}

	logEntry := allLogs[0]
	if logEntry.Message != "test message" {
		t.Errorf("message = %v, want 'test message'", logEntry.Message)
	}

	// 验证日志级别
	if logEntry.Level != zapcore.InfoLevel {
		t.Errorf("level = %v, want InfoLevel", logEntry.Level)
	}
}

// TestZapHandlerWithAttrs 测试 WithAttrs 方法
func TestZapHandlerWithAttrs(t *testing.T) {
	observedZapCore, logs := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(observedZapCore)
	baseHandler := newZapHandler(zapLogger)

	// 添加属性
	handlerWithAttrs := baseHandler.WithAttrs([]slog.Attr{
		slog.String("base", "value"),
	})

	// 创建日志记录
	now := time.Now()
	record := slog.NewRecord(now, slog.LevelInfo, "test message", 0)
	record.AddAttrs(slog.String("extra", "data"))

	// 处理日志
	err := handlerWithAttrs.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	// 验证日志输出
	if len(logs.All()) != 1 {
		t.Fatalf("got %d logs, want 1", len(logs.All()))
	}

	logEntry := logs.All()[0]
	// 验证 base 属性存在
	foundBase := false
	for _, field := range logEntry.Context {
		if field.Key == "base" && field.String == "value" {
			foundBase = true
			break
		}
	}
	if !foundBase {
		t.Error("base attribute not found in log entry")
	}
}

// TestZapHandlerWithGroup 测试 WithGroup 方法
func TestZapHandlerWithGroup(t *testing.T) {
	observedZapCore, logs := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(observedZapCore)
	baseHandler := newZapHandler(zapLogger)

	// 添加组
	handlerWithGroup := baseHandler.WithGroup("user")

	// 创建日志记录
	now := time.Now()
	record := slog.NewRecord(now, slog.LevelInfo, "test message", 0)
	record.AddAttrs(slog.String("name", "john"))

	// 处理日志
	err := handlerWithGroup.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	// 验证日志输出
	if len(logs.All()) != 1 {
		t.Fatalf("got %d logs, want 1", len(logs.All()))
	}

	logEntry := logs.All()[0]
	// 验证命名空间存在
	foundNamespace := false
	for _, field := range logEntry.Context {
		if field.Type == zapcore.StringType && field.Key == "user.name" {
			foundNamespace = true
			break
		}
	}
	if !foundNamespace {
		t.Error("group namespace not found in log entry")
	}
}

// TestGetSlogLogger 测试 GetSlogLogger 方法
func TestGetSlogLogger(t *testing.T) {
	resetLogger()
	defer resetLogger()

	// 初始化 logger
	Init(context.Background())

	// 获取 slog logger
	slogLogger := GetSlogLogger()
	if slogLogger == nil {
		t.Fatal("GetSlogLogger() returned nil")
	}

	// 验证可以记录日志
	observedZapCore, logs := observer.New(zapcore.InfoLevel)
	testLogger := zap.New(observedZapCore)

	// 重新初始化 logger 使用测试 core
	mu.Lock()
	logger = testLogger
	sugar = logger.Sugar()
	mu.Unlock()

	// 获取新的 slog logger
	slogLogger = GetSlogLogger()

	// 记录日志
	slogLogger.Info("test message", "key", "value")

	// 验证日志输出
	allLogs := logs.TakeAll()
	if len(allLogs) != 1 {
		t.Fatalf("got %d logs, want 1", len(allLogs))
	}

	logEntry := allLogs[0]
	if logEntry.Message != "test message" {
		t.Errorf("message = %v, want 'test message'", logEntry.Message)
	}
}
