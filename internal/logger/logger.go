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
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/chenyang-zz/boxify/internal/utils"
)

const (
	envLogDir = "Boxify_LOG_DIR"
)

var (
	mu         sync.RWMutex
	logger     *zap.Logger
	sugar      *zap.SugaredLogger
	logPath    string
	initialized bool
)

// Init 初始化日志系统。根据环境变量或系统默认路径创建日志文件，并设置日志输出。
// 支持使用新的 context 重新初始化。
func Init(ctx context.Context) {
	mu.Lock()
	defer mu.Unlock()

	// 加载配置
	config := loadConfig(ctx)

	// 创建 zap logger
	newLogger, err := NewLogger(config)
	if err != nil {
		panic(fmt.Sprintf("初始化 logger 失败: %v", err))
	}

	// 如果已经初始化过，先关闭旧的 logger
	if initialized && logger != nil {
		_ = logger.Sync()
	}

	logger = newLogger
	sugar = logger.Sugar()
	logPath = config.OutputPath
	initialized = true

	logger.Info("日志初始化完成",
		zap.String("file", logPath),
		zap.String("level", config.Level),
	)
}

// loadConfig 加载日志配置
func loadConfig(ctx context.Context) *Config {
	// 确定日志级别
	level := "INFO"
	if utils.IsDev(ctx) {
		level = "DEBUG"
	}

	// 从环境变量读取日志级别（如果设置）
	if envLevel := strings.TrimSpace(os.Getenv("BOXIFY_LOG_LEVEL")); envLevel != "" {
		level = envLevel
	}

	// 构建配置
	return &Config{
		Level:      level,
		Format:      "text", // 默认文本格式
		OutputPath:  getLogPath(),
		MaxSize:     10,  // 10 MB
		MaxBackups:  10,  // 保留 10 个
		MaxAge:      30,  // 30 天
		Compress:    false,
	}
}

// Path 返回当前日志文件的路径。如果日志系统尚未初始化，则会先进行初始化。
func Path() string {
	Init(context.Background())
	return logPath
}

// Close 关闭日志系统，释放资源。
func Close() {
	Init(context.Background())
	if logger != nil {
		_ = logger.Sync()
	}
}

// GetZapLogger 返回底层 zap logger 实例（供 Wails 等外部使用）
func GetZapLogger() *zap.Logger {
	Init(context.Background())
	return logger
}

// GetSlogLogger 返回 slog.Logger 实例（底层使用 zap）
// 注意：每次调用都会创建新的 slog logger 实例，如果 logger 已重新初始化，
// 需要重新调用 GetSlogLogger() 获取新的实例。
func GetSlogLogger() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()

	// 确保 logger 已初始化
	if !initialized || logger == nil {
		mu.RUnlock()
		Init(context.Background())
		mu.RLock()
	}

	// 创建 zap handler
	handler := newZapHandler(logger)

	// 返回 slog logger
	return slog.New(handler)
}

// Debug 输出 DEBUG 级别的日志消息（结构化 API）
func Debug(msg string, fields ...zap.Field) {
	Init(context.Background())
	logger.Debug(msg, fields...)
}

// Info 输出 INFO 级别的日志消息（结构化 API）
func Info(msg string, fields ...zap.Field) {
	Init(context.Background())
	logger.Info(msg, fields...)
}

// Warn 输出 WARN 级别的日志消息（结构化 API）
func Warn(msg string, fields ...zap.Field) {
	Init(context.Background())
	logger.Warn(msg, fields...)
}

// Error 输出 ERROR 级别的日志消息（结构化 API）
func Error(msg string, fields ...zap.Field) {
	Init(context.Background())
	logger.Error(msg, fields...)
}

// Fatal 输出 FATAL 级别的日志消息，然后退出程序
func Fatal(msg string, fields ...zap.Field) {
	Init(context.Background())
	logger.Fatal(msg, fields...)
}

// Debugf 输出 DEBUG 级别的日志消息（格式化 API）
func Debugf(format string, args ...any) {
	Init(context.Background())
	sugar.Debugf(format, args...)
}

// Infof 输出 INFO 级别的日志消息（格式化 API）
func Infof(format string, args ...any) {
	Init(context.Background())
	sugar.Infof(format, args...)
}

// Warnf 输出 WARN 级别的日志消息（格式化 API）
func Warnf(format string, args ...any) {
	Init(context.Background())
	sugar.Warnf(format, args...)
}

// Errorf 输出 ERROR 级别的日志消息（格式化 API）
func Errorf(format string, args ...any) {
	Init(context.Background())
	sugar.Errorf(format, args...)
}

// ErrorfWithTrace 输出错误级别的日志消息，包含错误链信息。
func ErrorfWithTrace(err error, format string, args ...any) {
	Init(context.Background())
	msg := fmt.Sprintf(format, args...)
	if err == nil {
		Errorf("%s", msg)
		return
	}
	Errorf("%s；错误链：%s", msg, ErrorChain(err))
}

// ErrorChain 返回错误链的字符串表示，包含所有独特的错误消息，按顺序连接。
func ErrorChain(err error) string {
	if err == nil {
		return "<nil>"
	}

	var parts []string
	seen := make(map[string]struct{})
	cur := err
	truncated := false
	for i := 0; cur != nil && i < 20; i++ {
		s := cur.Error()
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			parts = append(parts, s)
		}
		cur = errors.Unwrap(cur)
	}
	if cur != nil {
		truncated = true
	}

	if len(parts) == 0 {
		return err.Error()
	}

	if truncated {
		parts = append(parts, "（错误链过长，已截断）")
	}
	return strings.Join(parts, " -> ")
}

// String 创建字符串字段（便捷方法）
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

// Int 创建整数字段（便捷方法）
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Err 创建错误字段（便捷方法）
func Err(err error) zap.Field {
	return zap.Error(err)
}

// Duration 创建时间段字段（便捷方法）
func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

// Any 创建任意类型字段（便捷方法）
func Any(key string, val any) zap.Field {
	return zap.Any(key, val)
}
