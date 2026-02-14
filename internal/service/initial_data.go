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

package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/chenyang-zz/boxify/internal/window"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	// maxDataSize 数据大小限制（1MB）
	maxDataSize = 1024 * 1024
)

// InitialDataEntry 初始数据条目
type InitialDataEntry struct {
	WindowName string                 `json:"windowName"` // 目标窗口名称
	Source     string                 `json:"source"`     // 源窗口名称
	Data       map[string]interface{} `json:"data"`       // 实际数据
	Timestamp  int64                  `json:"timestamp"`  // 创建时间戳
	ExpiresAt  int64                  `json:"expiresAt"`  // 过期时间戳
}

// InitialDataService 初始数据服务
type InitialDataService struct {
	am        *window.AppManager
	data      map[string]*InitialDataEntry // windowName -> 数据条目
	mu        sync.RWMutex                 // 读写锁
	logger    *slog.Logger
	maxAge    time.Duration // 最大存活时间
	cleanChan chan struct{} // 清理通道
}

// NewInitialDataService 创建初始数据服务
func NewInitialDataService(am *window.AppManager) *InitialDataService {

	service := &InitialDataService{
		am:        am,
		data:      make(map[string]*InitialDataEntry),
		maxAge:    30 * time.Minute, // 默认30分钟过期
		cleanChan: make(chan struct{}),
		logger:    application.Get().Logger,
	}

	// 启动后台清理协程
	go service.cleanupExpiredData()

	return service
}

// SaveInitialData 保存窗口初始数据
func (ids *InitialDataService) SaveInitialData(sourceWindow, targetWindow string, data map[string]interface{}, ttlMinutes int) *connection.QueryResult {
	// 验证输入
	if sourceWindow == "" {
		return &connection.QueryResult{
			Success: false,
			Message: "源窗口名称不能为空",
		}
	}
	if targetWindow == "" {
		return &connection.QueryResult{
			Success: false,
			Message: "目标窗口名称不能为空",
		}
	}

	// 目标窗口不能打开
	if ids.am.GetRegistry().IsRegistered(targetWindow) {
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("目标窗口已打开: %s", targetWindow),
		}
	}

	// 检查数据大小
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("数据序列化失败: %s", err.Error()),
		}
	}
	if len(dataBytes) > maxDataSize {
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("数据过大: %d bytes (最大 %d bytes)", len(dataBytes), maxDataSize),
		}
	}

	// 计算过期时间
	expiresAt := time.Now().Add(ids.maxAge)
	if ttlMinutes > 0 {
		expiresAt = time.Now().Add(time.Duration(ttlMinutes) * time.Minute)
	}

	entry := &InitialDataEntry{
		WindowName: targetWindow,
		Source:     sourceWindow,
		Data:       data,
		Timestamp:  time.Now().Unix(),
		ExpiresAt:  expiresAt.Unix(),
	}

	ids.mu.Lock()
	ids.data[targetWindow] = entry
	ids.mu.Unlock()

	if ids.logger != nil {
		ids.logger.Info("初始数据已保存",
			"source", sourceWindow,
			"target", targetWindow,
			"expiresAt", expiresAt,
		)
	}

	// 发送事件通知目标窗口
	ids.emitWindowInitialData(ids.GetInitialData(targetWindow).Data.(*InitialDataEntry))

	return &connection.QueryResult{
		Success: true,
		Message: fmt.Sprintf("初始数据已保存: %s -> %s", sourceWindow, targetWindow),
		Data: map[string]interface{}{
			"windowName": targetWindow,
			"expiresAt":  expiresAt.Unix(),
		},
	}
}

// GetInitialData 获取窗口初始数据
func (ids *InitialDataService) GetInitialData(windowName string) *connection.QueryResult {
	if windowName == "" {
		return &connection.QueryResult{
			Success: false,
			Message: "窗口名称不能为空",
		}
	}

	ids.mu.RLock()
	entry, exists := ids.data[windowName]
	ids.mu.RUnlock()

	if !exists {
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("无初始数据: %s", windowName),
		}
	}

	// 检查是否过期
	if time.Now().Unix() > entry.ExpiresAt {
		ids.mu.Lock()
		delete(ids.data, windowName)
		ids.mu.Unlock()

		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("初始数据已过期: %s", windowName),
		}
	}

	if ids.logger != nil {
		ids.logger.Info("初始数据已获取",
			"window", windowName,
			"source", entry.Source,
		)
	}

	return &connection.QueryResult{
		Success: true,
		Message: "初始数据获取成功",
		Data:    entry,
	}
}

// ClearInitialData 清除窗口初始数据
func (ids *InitialDataService) ClearInitialData(windowName string) *connection.QueryResult {
	if windowName == "" {
		return &connection.QueryResult{
			Success: false,
			Message: "窗口名称不能为空",
		}
	}

	ids.mu.Lock()
	defer ids.mu.Unlock()

	if _, exists := ids.data[windowName]; exists {
		delete(ids.data, windowName)

		if ids.logger != nil {
			ids.logger.Info("初始数据已清除",
				"window", windowName,
			)
		}

		return &connection.QueryResult{
			Success: true,
			Message: fmt.Sprintf("初始数据已清除: %s", windowName),
		}
	}

	return &connection.QueryResult{
		Success: false,
		Message: fmt.Sprintf("无初始数据可清除: %s", windowName),
	}
}

// cleanupExpiredData 清理过期数据（后台协程）
func (ids *InitialDataService) cleanupExpiredData() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ids.mu.Lock()
			now := time.Now().Unix()
			cleanedCount := 0
			for name, entry := range ids.data {
				if now > entry.ExpiresAt {
					delete(ids.data, name)
					cleanedCount++
				}
			}
			ids.mu.Unlock()

			if cleanedCount > 0 && ids.logger != nil {
				ids.logger.Info("过期数据已清理",
					"count", cleanedCount,
				)
			}
		case <-ids.cleanChan:
			return
		}
	}
}

// Shutdown 关闭服务
func (ids *InitialDataService) Shutdown() {
	close(ids.cleanChan)
}

// emitWindowEvent 发送窗口事件
func (ids *InitialDataService) emitWindowInitialData(entry *InitialDataEntry) {
	application.Get().Event.Emit("initial-data:received", *entry)
}
