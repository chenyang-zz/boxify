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

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// PageConfig 页面配置
type PageConfig struct {
	ID          string                            `json:"id"`
	Type        string                            `json:"type"` // "main", "singleton", "modal"
	Title       string                            `json:"title"`
	Entry       string                            `json:"entry"`
	ContainerID string                            `json:"containerId"`
	IsMain      bool                              `json:"isMain"`
	Parent      string                            `json:"parent"`
	Center      bool                              `json:"center"`
	Window      *application.WebviewWindowOptions `json:"window"`
}

// PageConfigFile 页面配置文件结构
type PageConfigFile struct {
	Pages []PageConfig `json:"pages"`
}

// LoadPageConfig 加载页面配置文件
func LoadPageConfig(configPath string) (*PageConfigFile, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config PageConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	for _, page := range config.Pages {
		if err := page.Validate(); err != nil {
			return nil, fmt.Errorf("页面 %s 配置无效: %w", page.ID, err)
		}
	}

	return &config, nil
}

// GetPageConfig 根据 pageId 获取页面配置
func (pc *PageConfigFile) GetPageConfig(pageId string) *PageConfig {
	for _, page := range pc.Pages {
		if page.ID == pageId {
			return &page
		}
	}
	return nil
}

// GetMainPageConfig 获取主页面配置
func (pc *PageConfigFile) GetMainPageConfig() *PageConfig {
	for _, page := range pc.Pages {
		if page.IsMain {
			return &page
		}
	}
	return nil
}

// Validate 验证页面配置
func (pc *PageConfig) Validate() error {
	if pc.ID == "" {
		return fmt.Errorf("页面 ID 不能为空")
	}

	if pc.Window == nil {
		return fmt.Errorf("窗口配置不能为空")
	}

	opts := pc.Window

	if opts.Name == "" {
		return fmt.Errorf("窗口名称不能为空")
	}

	if opts.Width <= 0 || opts.Height <= 0 {
		return fmt.Errorf("窗口宽度和高度必须大于 0")
	}

	return nil
}

// GetPageConfigPath 获取页面配置文件的默认路径
func GetPageConfigPath() string {
	return filepath.Join(".", "page.config.json")
}
