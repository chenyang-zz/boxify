// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with License.
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
	"fmt"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
)

// DataSyncEvent 数据同步事件
type DataSyncEvent struct {
	Source    string                 `json:"source"`    // 发送窗口名称
	Target    string                 `json:"target"`    // 目标窗口（空 = 广播）
	Channel   string                 `json:"channel"`   // 数据频道
	DataType  string                 `json:"dataType"`  // 数据类型标识
	Data      map[string]interface{} `json:"data"`      // 实际数据
	Timestamp int64                  `json:"timestamp"` // 时间戳
	ID        string                 `json:"id"`        // 唯一消息ID
}

// 预定义的数据频道
const (
	ChannelConfig     = "config"     // 配置同步
	ChannelConnection = "connection" // 连接配置
	ChannelSettings   = "settings"   // 应用设置
	ChannelCustom     = "custom"     // 自定义频道
)

// 预定义的数据类型
const (
	DataTypePropertyUpdate   = "property:update"
	DataTypeConnectionAdd    = "connection:add"
	DataTypeConnectionUpdate = "connection:update"
	DataTypeConnectionDelete = "connection:delete"
	DataTypeSettingsUpdate   = "settings:update"
	DataTypeThemeChanged     = "theme:changed"
	DataTypeConnectionState  = "connection:state"
)

// DataSyncService 数据同步服务
type DataSyncService struct {
	BaseService
	lastEventTime map[string]time.Time // 消息去重
}

// NewDataSyncService 创建数据同步服务
func NewDataSyncService(deps *ServiceDeps) *DataSyncService {
	return &DataSyncService{
		BaseService:   NewBaseService(deps),
		lastEventTime: make(map[string]time.Time),
	}
}

// Broadcast 广播消息到所有窗口
func (ds *DataSyncService) Broadcast(channel, dataType string, data map[string]interface{}, source string) error {
	return ds.Emit(DataSyncEvent{
		Source:    source,
		Target:    "", // 空表示广播
		Channel:   channel,
		DataType:  dataType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// SendTo 发送消息到指定窗口
func (ds *DataSyncService) SendTo(targetWindow, channel, dataType string, data map[string]interface{}, source string) error {
	// 验证目标窗口是否存在
	registry := ds.Registry()
	if registry == nil || registry.Get(targetWindow) == nil {
		return fmt.Errorf("目标窗口不存在: %s", targetWindow)
	}

	return ds.Emit(DataSyncEvent{
		Source:    source,
		Target:    targetWindow,
		Channel:   channel,
		DataType:  dataType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// Emit 发送事件（内部方法）
func (ds *DataSyncService) Emit(event DataSyncEvent) error {
	// 消息去重：1秒内相同事件去重
	key := fmt.Sprintf("%s:%s:%s", event.Source, event.Channel, event.DataType)

	ds.mu.Lock()
	defer ds.mu.Unlock()

	if lastTime, exists := ds.lastEventTime[key]; exists {
		if time.Since(lastTime) < time.Second {
			ds.Logger().Debug("数据同步事件去重", "key", key)
			return nil
		}
	}

	ds.lastEventTime[key] = time.Now()

	// 生成唯一消息ID
	event.ID = ds.generateMessageID()

	// 根据目标获取事件名称
	eventName := ds.getEventName(event.Target)

	// 发送事件
	ds.App().Event.Emit(eventName, event)

	ds.Logger().Info("数据同步事件已发送",
		"source", event.Source,
		"target", event.Target,
		"channel", event.Channel,
		"dataType", event.DataType,
	)

	return nil
}

// getEventName 根据目标获取事件名称
func (ds *DataSyncService) getEventName(target string) string {
	if target == "" {
		return "data-sync:broadcast"
	}
	return "data-sync:targeted"
}

// generateMessageID 生成唯一消息ID
func (ds *DataSyncService) generateMessageID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetWindowsList 获取所有窗口列表
func (ds *DataSyncService) GetWindowsList() *connection.QueryResult {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	registry := ds.Registry()
	if registry == nil {
		return &connection.QueryResult{
			Success: false,
			Message: "窗口注册表未初始化",
		}
	}

	windowInfos := registry.GetAllWindowInfos()
	result := make([]map[string]interface{}, 0, len(windowInfos))

	for _, info := range windowInfos {
		result = append(result, map[string]interface{}{
			"name":  info.Name,
			"type":  info.Type,
			"title": info.Title,
			"id":    info.ID,
		})
	}

	return &connection.QueryResult{
		Success: true,
		Message: "获取窗口列表成功",
		Data:    result,
	}
}
