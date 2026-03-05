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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/pkg/errors"
)

// DefaultCachePingInterval 是缓存连接的默认探活间隔。
const DefaultCachePingInterval = 30 * time.Second

// cacheEntry 描述一个已缓存的数据库连接及其最近探活时间。
type cacheEntry struct {
	inst     Database
	lastPing time.Time
}

// ConnectionManager 管理数据库连接缓存、探活和重建。
type ConnectionManager struct {
	mu           sync.RWMutex
	logger       *slog.Logger
	pingInterval time.Duration
	cache        map[string]cacheEntry
}

// NewConnectionManager 创建数据库连接管理器。
func NewConnectionManager(logger *slog.Logger) *ConnectionManager {
	return &ConnectionManager{
		logger:       logger,
		pingInterval: DefaultCachePingInterval,
		cache:        make(map[string]cacheEntry),
	}
}

// Get 返回可用数据库连接；forcePing=true 时会强制探活。
func (m *ConnectionManager) Get(config *connection.ConnectionConfig, forcePing bool) (Database, error) {
	key := cacheKey(config)
	shortKey := shortCacheKey(key)

	m.mu.RLock()
	entry, ok := m.cache[key]
	m.mu.RUnlock()

	if ok {
		needPing := forcePing
		if !needPing && (entry.lastPing.IsZero() || time.Since(entry.lastPing) >= m.pingInterval) {
			needPing = true
		}

		if !needPing {
			return entry.inst, nil
		}

		if err := entry.inst.Ping(); err == nil {
			m.mu.Lock()
			if cur, exists := m.cache[key]; exists && cur.inst == entry.inst {
				cur.lastPing = time.Now()
				m.cache[key] = cur
			}
			m.mu.Unlock()
			return entry.inst, nil
		}

		m.logError("缓存连接不可用，准备重建", "summary", FormatConnSummary(config), "key", shortKey)
		m.removeCacheEntry(key, entry.inst)
	}

	m.logInfo("获取数据库连接", "summary", FormatConnSummary(config), "key", shortKey)
	dbInst, err := NewDatabase(config.Type)
	if err != nil {
		m.logError("创建数据库驱动实例失败", "type", config.Type, "key", shortKey, "error", err)
		return nil, err
	}

	if err = dbInst.Connect(config); err != nil {
		wrapped := wrapConnectError(config, err)
		m.logError("建立数据库连接失败", "summary", FormatConnSummary(config), "key", shortKey, "error", wrapped)
		return nil, wrapped
	}

	now := time.Now()
	m.mu.Lock()
	if existing, exists := m.cache[key]; exists && existing.inst != nil {
		m.mu.Unlock()
		_ = dbInst.Close()
		return existing.inst, nil
	}
	m.cache[key] = cacheEntry{inst: dbInst, lastPing: now}
	m.mu.Unlock()

	m.logInfo("数据库连接成功并写入缓存", "summary", FormatConnSummary(config), "key", shortKey)
	return dbInst, nil
}

// CloseAll 关闭并清空所有缓存连接。
func (m *ConnectionManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var closeErr error
	for key, entry := range m.cache {
		if entry.inst == nil {
			delete(m.cache, key)
			continue
		}
		if err := entry.inst.Close(); err != nil && closeErr == nil {
			closeErr = err
			m.logError("关闭数据库连接失败", "key", shortCacheKey(key), "error", err)
		}
		delete(m.cache, key)
	}
	return closeErr
}

func (m *ConnectionManager) removeCacheEntry(key string, expected Database) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cur, exists := m.cache[key]
	if !exists || cur.inst != expected {
		return
	}
	if err := cur.inst.Close(); err != nil {
		m.logError("关闭失效缓存连接失败", "key", shortCacheKey(key), "error", err)
	}
	delete(m.cache, key)
}

func (m *ConnectionManager) logInfo(msg string, args ...any) {
	if m.logger != nil {
		m.logger.Info(msg, args...)
	}
}

func (m *ConnectionManager) logError(msg string, args ...any) {
	if m.logger != nil {
		m.logger.Error(msg, args...)
	}
}

func cacheKey(config *connection.ConnectionConfig) string {
	normalized := normalizedConfig(config)
	b, _ := json.Marshal(normalized)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func normalizedConfig(config *connection.ConnectionConfig) *connection.ConnectionConfig {
	runConfig := *config

	if !runConfig.UseSSH {
		runConfig.SSH = &connection.SSHConfig{}
	}

	// 保持与历史行为一致，避免同一连接生成不同缓存 key。
	if (runConfig.Type == "postgres" || runConfig.Type == connection.ConnectionTypePostgreSQL) && runConfig.Database == "" {
		runConfig.Database = "postgres"
	}

	return &runConfig
}

func shortCacheKey(key string) string {
	if len(key) <= 12 {
		return key
	}
	return key[:12]
}

func wrapConnectError(config *connection.ConnectionConfig, err error) error {
	if err == nil {
		return nil
	}

	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		dbName := config.Database
		if dbName == "" {
			dbName = "<default>"
		}
		err = fmt.Errorf("数据库连接超时：%s %s:%d/%s：%w", config.Type, config.Host, config.Port, dbName, err)
	}

	return withLogHint{
		err:     err,
		logPath: "",
	}
}

// FormatConnSummary 输出连接信息摘要，供日志复用。
func FormatConnSummary(config *connection.ConnectionConfig) string {
	timeoutSeconds := config.Timeout
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	dbName := config.Database
	if dbName == "" {
		dbName = "<default>"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("类型=%s 地址=%s:%d 数据库=%s 用户=%s 超时=%ds",
		config.Type, config.Host, config.Port, dbName, config.User, timeoutSeconds))
	if config.UseSSH && config.SSH != nil {
		b.WriteString(fmt.Sprintf(" SSH=%s:%d 用户=%s", config.SSH.Host, config.SSH.Port, config.SSH.User))
	}
	if config.Type == "custom" {
		driver := strings.TrimSpace(config.Driver)
		if driver == "" {
			driver = "<未配置>"
		}
		dsnState := "<未配置>"
		if strings.TrimSpace(config.DSN) != "" {
			dsnState = fmt.Sprintf("已配置(长度=%d)", len(config.DSN))
		}
		b.WriteString(fmt.Sprintf(" 驱动=%s DSN=%s", driver, dsnState))
	}
	return b.String()
}

type withLogHint struct {
	err     error
	logPath string
}

func (e withLogHint) Error() string {
	if strings.TrimSpace(e.logPath) == "" {
		return e.err.Error()
	}
	return fmt.Sprintf("%s（详细日志：%s）", e.err.Error(), e.logPath)
}

func (e withLogHint) Unwrap() error {
	return e.err
}
