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

	// 验证 Title 字段
	if pc.Title == "" {
		return fmt.Errorf("页面标题不能为空")
	}

	// 验证 Entry 字段
	if pc.Entry == "" {
		return fmt.Errorf("页面入口文件不能为空")
	}

	// 验证 Type 字段
	validTypes := map[string]bool{
		"main":      true,
		"singleton": true,
		"modal":     true,
	}
	if pc.Type != "" && !validTypes[pc.Type] {
		return fmt.Errorf("无效的页面类型: %s，必须是 main、singleton 或 modal", pc.Type)
	}

	// 验证 IsMain 和 Type 的一致性
	if pc.IsMain && pc.Type != "" && pc.Type != "main" {
		return fmt.Errorf("主页面（isMain=true）的类型必须是 main，当前为: %s", pc.Type)
	}

	// 验证 modal 类型必须有 parent
	if pc.Type == "modal" && pc.Parent == "" {
		return fmt.Errorf("模态窗口（modal）必须指定父窗口（parent）")
	}

	if pc.Window == nil {
		return fmt.Errorf("窗口配置不能为空")
	}

	opts := pc.Window

	if opts.Name == "" {
		return fmt.Errorf("窗口名称不能为空")
	}

	if opts.Title == "" {
		opts.Title = pc.Title
	}

	// 验证窗口宽度和高度
	if opts.Width <= 0 {
		return fmt.Errorf("窗口宽度必须大于 0，当前为: %d", opts.Width)
	}

	if opts.Height <= 0 {
		return fmt.Errorf("窗口高度必须大于 0，当前为: %d", opts.Height)
	}

	// 验证窗口大小合理性（最大不超过 10000）
	if opts.Width > 10000 {
		return fmt.Errorf("窗口宽度超出合理范围（最大 10000），当前为: %d", opts.Width)
	}

	if opts.Height > 10000 {
		return fmt.Errorf("窗口高度超出合理范围（最大 10000），当前为: %d", opts.Height)
	}

	// 设置默认 URL
	if opts.URL == "" {
		if pc.IsMain {
			opts.URL = "/"
		} else {
			opts.URL = fmt.Sprintf("/%s", pc.ID)
		}
	}

	return nil
}

// GetPageConfigPath 获取页面配置文件的默认路径
func GetPageConfigPath() string {
	return filepath.Join(".", "page.config.json")
}
