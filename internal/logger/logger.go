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
	"os"
	"sync"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

var (
	logger *slog.Logger
	mu     sync.RWMutex
)

func Init(level slog.Leveler) {
	mu.Lock()
	defer mu.Unlock()
	if logger != nil {
		return
	}

	logger = DefaultLogger(level)
}

func DefaultLogger(level slog.Leveler) *slog.Logger {
	return slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		TimeFormat: time.Kitchen,
		NoColor:    !isatty.IsTerminal(os.Stderr.Fd()),
		Level:      level,
	}))
}

func GetDefaultLogger() *slog.Logger {
	mu.RLock()
	if logger != nil {
		mu.RUnlock()
		return logger
	}
	mu.RUnlock()

	// 需要初始化
	Init(slog.LevelInfo)

	mu.RLock()
	defer mu.RUnlock()
	return logger
}

func Debug(msg string, args ...any) {
	GetDefaultLogger().Debug(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	GetDefaultLogger().DebugContext(ctx, msg, args...)
}

func Info(msg string, args ...any) {
	GetDefaultLogger().Info(msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	GetDefaultLogger().InfoContext(ctx, msg, args...)
}

func Warn(msg string, args ...any) {
	GetDefaultLogger().Warn(msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	GetDefaultLogger().WarnContext(ctx, msg, args...)
}

func Error(msg string, args ...any) {
	GetDefaultLogger().Error(msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	GetDefaultLogger().ErrorContext(ctx, msg, args...)
}
