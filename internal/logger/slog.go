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
	"fmt"
	"log/slog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapHandler slog.Handler 实现，底层使用 zap logger
type zapHandler struct {
	base   *zap.Logger   // 底层 zap logger
	attrs  []slog.Attr // 预设属性
	groups []string     // 命名空间组
}

// newZapHandler 创建新的 zap handler
func newZapHandler(zapLogger *zap.Logger) *zapHandler {
	return &zapHandler{
		base:   zapLogger,
		attrs:  nil,
		groups: nil,
	}
}

// Enabled 检查日志级别是否启用
func (h *zapHandler) Enabled(ctx context.Context, level slog.Level) bool {
	zapLevel, err := slogLevelToZapLevel(level)
	if err != nil {
		return false
	}
	return h.base.Core().Enabled(zapLevel)
}

// Handle 处理日志记录
func (h *zapHandler) Handle(ctx context.Context, r slog.Record) error {
	// 转换日志级别
	zapLevel, err := slogLevelToZapLevel(r.Level)
	if err != nil {
		return err
	}

	// 构建 zap fields
	fields := make([]zap.Field, 0, len(h.attrs))

	// 构建命名空间前缀
	var namespacePrefix string
	if len(h.groups) > 0 {
		for _, group := range h.groups {
			if namespacePrefix == "" {
				namespacePrefix = group
			} else {
				namespacePrefix = group + "." + namespacePrefix
			}
		}
	}

	// 添加预设属性（应用命名空间前缀）
	for _, attr := range h.attrs {
		field := attrToField(attr)
		if namespacePrefix != "" {
			field.Key = namespacePrefix + "." + field.Key
		}
		fields = append(fields, field)
	}

	// 添加记录属性（应用命名空间前缀）
	r.Attrs(func(attr slog.Attr) bool {
		field := attrToField(attr)
		if namespacePrefix != "" {
			field.Key = namespacePrefix + "." + field.Key
		}
		fields = append(fields, field)
		return true
	})

	// 调用 zap logger
	if h.base.Core().Enabled(zapLevel) {
		h.base.Log(zapLevel, r.Message, fields...)
	}

	return nil
}

// WithAttrs 返回带有额外属性的 handler
func (h *zapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &zapHandler{
		base:   h.base,
		attrs:  append(h.attrs, attrs...),
		groups: h.groups,
	}
}

// WithGroup 返回带有命名空间的 handler
func (h *zapHandler) WithGroup(name string) slog.Handler {
	return &zapHandler{
		base:   h.base,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

// slogLevelToZapLevel 转换日志级别
func slogLevelToZapLevel(level slog.Level) (zapcore.Level, error) {
	switch level {
	case slog.LevelDebug:
		return zapcore.DebugLevel, nil
	case slog.LevelInfo:
		return zapcore.InfoLevel, nil
	case slog.LevelWarn:
		return zapcore.WarnLevel, nil
	case slog.LevelError:
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown slog level: %d", level)
	}
}

// attrToField 将 slog.Attr 转换为 zap.Field
func attrToField(attr slog.Attr) zap.Field {
	switch attr.Value.Kind() {
	case slog.KindString:
		return zap.String(attr.Key, attr.Value.String())
	case slog.KindInt64:
		return zap.Int64(attr.Key, attr.Value.Int64())
	case slog.KindUint64:
		return zap.Uint64(attr.Key, attr.Value.Uint64())
	case slog.KindFloat64:
		return zap.Float64(attr.Key, attr.Value.Float64())
	case slog.KindBool:
		return zap.Bool(attr.Key, attr.Value.Bool())
	case slog.KindTime:
		return zap.Time(attr.Key, attr.Value.Time())
	case slog.KindDuration:
		return zap.Duration(attr.Key, attr.Value.Duration())
	case slog.KindGroup:
		// 递归处理组 - 使用 Any 将整个组序列化
		groupMap := make(map[string]interface{})
		group := attr.Value.Group()
		for _, a := range group {
			groupMap[a.Key] = a.Value.Any()
		}
		return zap.Any(attr.Key, groupMap)
	default:
		return zap.Any(attr.Key, attr.Value.Any())
	}
}
