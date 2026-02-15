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
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/chenyang-zz/boxify/internal/utils"
)

// Config 定义 zap logger 的配置
type Config struct {
	Level      string // DEBUG, INFO, WARN, ERROR
	Format     string // json, text
	OutputPath string // 日志文件路径
	MaxSize    int    // 单文件最大大小（MB）
	MaxBackups int    // 保留旧文件数量
	MaxAge     int    // 保留天数
	Compress   bool   // 是否压缩旧日志
}

// NewLogger 创建一个新的 zap logger 实例
func NewLogger(config *Config) (*zap.Logger, error) {
	// 解析日志级别
	level := zapcore.InfoLevel
	if err := level.Set(config.Level); err != nil {
		return nil, err
	}

	// 创建 lumberjack writer（支持日志轮转）
	fileWriter := &lumberjack.Logger{
		Filename:   config.OutputPath,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
		LocalTime:  true,
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "func",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 根据格式选择编码器
	var encoder zapcore.Encoder
	if strings.ToLower(config.Format) == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建核心（文件输出）
	fileCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(fileWriter),
		level,
	)

	// 开发模式添加控制台输出
	core := fileCore
	if utils.IsDev(nil) {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
		core = zapcore.NewTee(fileCore, consoleCore)
	}

	// 创建 logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

// getLogPath 获取日志文件路径
func getLogPath() string {
	// 优先使用环境变量指定的目录
	dir := strings.TrimSpace(os.Getenv("Boxify_LOG_DIR"))

	if dir == "" {
		// 否则使用系统默认配置目录
		base, err := os.UserConfigDir()
		if err != nil || strings.TrimSpace(base) == "" {
			base = os.TempDir()
		}
		dir = filepath.Join(base, "boxify", "logs")
	}

	return filepath.Join(dir, "boxify.log")
}
