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
	"fmt"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/chenyang-zz/boxify/internal/window"
)

// WindowService struct
type WindowService struct {
	am *window.AppManager
}

// NewWindowService 新建一个WindowService实例
func NewWindowService(am *window.AppManager) *WindowService {
	return &WindowService{
		am: am,
	}
}

// OpenPage 打开页面（统一 API）
func (ws *WindowService) OpenPage(pageId string) *connection.QueryResult {
	err := ws.am.OpenPage(pageId)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("打开页面失败: %s", err.Error()),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: fmt.Sprintf("页面已打开: %s", pageId),
	}
}

// ClosePage 关闭页面
func (ws *WindowService) ClosePage(pageId string) *connection.QueryResult {
	err := ws.am.ClosePage(pageId)
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("关闭页面失败: %s", err.Error()),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: fmt.Sprintf("页面已关闭: %s", pageId),
	}
}

// GetPageList 获取所有可用页面列表
func (ws *WindowService) GetPageList() *connection.QueryResult {
	pages := ws.am.GetPageConfig()

	pageList := make([]map[string]interface{}, 0)
	for _, page := range pages.Pages {
		pageInfo := map[string]interface{}{
			"id":     page.ID,
			"title":  page.Title,
			"isMain": page.IsMain,
		}
		pageList = append(pageList, pageInfo)
	}

	return &connection.QueryResult{
		Success: true,
		Data:    pageList,
	}
}
