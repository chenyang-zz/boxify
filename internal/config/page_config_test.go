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
	"os"
	"path/filepath"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// createValidPageConfig 创建一个有效的页面配置用于测试
func createValidPageConfig() *PageConfig {
	return &PageConfig{
		ID:     "test-page",
		Type:   "main",
		Title:  "测试页面",
		Entry:  "index.html",
		IsMain: true,
		Window: &application.WebviewWindowOptions{
			Name:   "test-window",
			Width:  800,
			Height: 600,
		},
	}
}

// TestPageConfigValidate 测试页面配置验证
func TestPageConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *PageConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "有效的配置",
			config:  createValidPageConfig(),
			wantErr: false,
		},
		{
			name: "ID 为空",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.ID = ""
				return cfg
			}(),
			wantErr: true,
			errMsg:  "页面 ID 不能为空",
		},
		{
			name: "Title 为空",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Title = ""
				return cfg
			}(),
			wantErr: true,
			errMsg:  "页面标题不能为空",
		},
		{
			name: "Entry 为空",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Entry = ""
				return cfg
			}(),
			wantErr: true,
			errMsg:  "页面入口文件不能为空",
		},
		{
			name: "无效的 Type",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Type = "invalid"
				return cfg
			}(),
			wantErr: true,
			errMsg:  "无效的页面类型",
		},
		{
			name: "Type 为 singleton",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Type = "singleton"
				cfg.IsMain = false
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "Type 为 modal 且有 parent",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Type = "modal"
				cfg.IsMain = false
				cfg.Parent = "main-window"
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "Type 为 modal 但 parent 为空",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Type = "modal"
				cfg.Parent = ""
				cfg.IsMain = false
				return cfg
			}(),
			wantErr: true,
			errMsg:  "模态窗口（modal）必须指定父窗口",
		},
		{
			name: "IsMain 为 true 但 Type 不为 main",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Type = "singleton"
				cfg.IsMain = true
				return cfg
			}(),
			wantErr: true,
			errMsg:  "主页面（isMain=true）的类型必须是 main",
		},
		{
			name: "Window 为空",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Window = nil
				return cfg
			}(),
			wantErr: true,
			errMsg:  "窗口配置不能为空",
		},
		{
			name: "窗口名称为空",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Window.Name = ""
				return cfg
			}(),
			wantErr: true,
			errMsg:  "窗口名称不能为空",
		},
		{
			name: "窗口宽度为 0",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Window.Width = 0
				return cfg
			}(),
			wantErr: true,
			errMsg:  "窗口宽度必须大于 0",
		},
		{
			name: "窗口宽度为负数",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Window.Width = -100
				return cfg
			}(),
			wantErr: true,
			errMsg:  "窗口宽度必须大于 0",
		},
		{
			name: "窗口高度为 0",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Window.Height = 0
				return cfg
			}(),
			wantErr: true,
			errMsg:  "窗口高度必须大于 0",
		},
		{
			name: "窗口高度为负数",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Window.Height = -100
				return cfg
			}(),
			wantErr: true,
			errMsg:  "窗口高度必须大于 0",
		},
		{
			name: "窗口宽度超出合理范围",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Window.Width = 10001
				return cfg
			}(),
			wantErr: true,
			errMsg:  "窗口宽度超出合理范围",
		},
		{
			name: "窗口高度超出合理范围",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Window.Height = 10001
				return cfg
			}(),
			wantErr: true,
			errMsg:  "窗口高度超出合理范围",
		},
		{
			name: "最小窗口尺寸",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Window.Width = 1
				cfg.Window.Height = 1
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "最大窗口尺寸",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Window.Width = 10000
				cfg.Window.Height = 10000
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "Type 为空（允许）",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Type = ""
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "ContainerID 可以为空",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.ContainerID = ""
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "Parent 可以为空（非 modal）",
			config: func() *PageConfig {
				cfg := createValidPageConfig()
				cfg.Parent = ""
				return cfg
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				// 检查错误消息是否包含预期内容
				found := false
				for _, msg := range []string{
					tt.errMsg,
					// 允许错误消息的其他可能形式
				} {
					if containsString(err.Error(), msg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Validate() error message = %q, 期望包含 %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// containsString 检查字符串是否包含子字符串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsInString(s, substr))
}

func containsInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestPageConfigFileGetPageConfig 测试获取页面配置
func TestPageConfigFileGetPageConfig(t *testing.T) {
	config := &PageConfigFile{
		Pages: []PageConfig{
			{
				ID:    "page1",
				Title: "页面 1",
				Entry: "page1.html",
				Window: &application.WebviewWindowOptions{
					Name:   "window1",
					Width:  800,
					Height: 600,
				},
			},
			{
				ID:    "page2",
				Title: "页面 2",
				Entry: "page2.html",
				Window: &application.WebviewWindowOptions{
					Name:   "window2",
					Width:  1024,
					Height: 768,
				},
			},
		},
	}

	tests := []struct {
		name     string
		pageId   string
		wantNil  bool
		expected string
	}{
		{
			name:     "存在的页面",
			pageId:   "page1",
			wantNil:  false,
			expected: "page1",
		},
		{
			name:     "存在的另一个页面",
			pageId:   "page2",
			wantNil:  false,
			expected: "page2",
		},
		{
			name:     "不存在的页面",
			pageId:   "page3",
			wantNil:  true,
		},
		{
			name:     "空字符串 ID",
			pageId:   "",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.GetPageConfig(tt.pageId)
			if (result == nil) != tt.wantNil {
				t.Errorf("GetPageConfig() = %v, wantNil %v", result, tt.wantNil)
				return
			}
			if !tt.wantNil && result.ID != tt.expected {
				t.Errorf("GetPageConfig().ID = %s, 期望 %s", result.ID, tt.expected)
			}
		})
	}
}

// TestPageConfigFileGetMainPageConfig 测试获取主页面配置
func TestPageConfigFileGetMainPageConfig(t *testing.T) {
	tests := []struct {
		name     string
		pages    []PageConfig
		wantNil  bool
		expected string
	}{
		{
			name: "存在主页面",
			pages: []PageConfig{
				{
					ID:     "page1",
					Title:  "页面 1",
					Entry:  "page1.html",
					IsMain: true,
					Window: &application.WebviewWindowOptions{
						Name:   "main",
						Width:  800,
						Height: 600,
					},
				},
				{
					ID:    "page2",
					Title: "页面 2",
					Entry: "page2.html",
					Window: &application.WebviewWindowOptions{
						Name:   "window2",
						Width:  1024,
						Height: 768,
					},
				},
			},
			wantNil:  false,
			expected: "page1",
		},
		{
			name: "没有主页面",
			pages: []PageConfig{
				{
					ID:    "page1",
					Title: "页面 1",
					Entry: "page1.html",
					Window: &application.WebviewWindowOptions{
						Name:   "window1",
						Width:  800,
						Height: 600,
					},
				},
			},
			wantNil: true,
		},
		{
			name:    "空配置",
			pages:   []PageConfig{},
			wantNil: true,
		},
		{
			name: "多个主页面（返回第一个）",
			pages: []PageConfig{
				{
					ID:     "page1",
					Title:  "页面 1",
					Entry:  "page1.html",
					IsMain: true,
					Window: &application.WebviewWindowOptions{
						Name:   "main1",
						Width:  800,
						Height: 600,
					},
				},
				{
					ID:     "page2",
					Title:  "页面 2",
					Entry:  "page2.html",
					IsMain: true,
					Window: &application.WebviewWindowOptions{
						Name:   "main2",
						Width:  1024,
						Height: 768,
					},
				},
			},
			wantNil:  false,
			expected: "page1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &PageConfigFile{Pages: tt.pages}
			result := config.GetMainPageConfig()
			if (result == nil) != tt.wantNil {
				t.Errorf("GetMainPageConfig() = %v, wantNil %v", result, tt.wantNil)
				return
			}
			if !tt.wantNil && result.ID != tt.expected {
				t.Errorf("GetMainPageConfig().ID = %s, 期望 %s", result.ID, tt.expected)
			}
		})
	}
}

// TestLoadPageConfig 测试加载页面配置文件
func TestLoadPageConfig(t *testing.T) {
	// 创建临时测试目录
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		config    *PageConfigFile
		wantErr   bool
		errMsg    string
		validate  bool
	}{
		{
			name: "有效的配置文件",
			config: &PageConfigFile{
				Pages: []PageConfig{
					{
						ID:     "main",
						Type:   "main",
						Title:  "主页面",
						Entry:  "index.html",
						IsMain: true,
						Window: &application.WebviewWindowOptions{
							Name:   "main-window",
							Width:  1024,
							Height: 768,
						},
					},
				},
			},
			wantErr:  false,
			validate: true,
		},
		{
			name: "多个页面的配置文件",
			config: &PageConfigFile{
				Pages: []PageConfig{
					{
						ID:     "main",
						Type:   "main",
						Title:  "主页面",
						Entry:  "index.html",
						IsMain: true,
						Window: &application.WebviewWindowOptions{
							Name:   "main",
							Width:  1024,
							Height: 768,
						},
					},
					{
						ID:     "settings",
						Type:   "modal",
						Title:  "设置",
						Entry:  "settings.html",
						Parent: "main",
						Window: &application.WebviewWindowOptions{
							Name:   "settings",
							Width:  600,
							Height: 400,
						},
					},
				},
			},
			wantErr:  false,
			validate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建配置文件
			configPath := filepath.Join(tempDir, tt.name+".json")
			data, err := json.MarshalIndent(tt.config, "", "  ")
			if err != nil {
				t.Fatalf("无法序列化配置: %v", err)
			}
			if err := os.WriteFile(configPath, data, 0644); err != nil {
				t.Fatalf("无法写入配置文件: %v", err)
			}

			// 加载配置
			result, err := LoadPageConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadPageConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate && err == nil {
				if len(result.Pages) != len(tt.config.Pages) {
					t.Errorf("LoadPageConfig() 页面数量 = %d, 期望 %d",
						len(result.Pages), len(tt.config.Pages))
				}
			}
		})
	}
}

// TestLoadPageConfigInvalid 测试加载无效配置文件
func TestLoadPageConfigInvalid(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		content  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "文件不存在",
			content: "",
			wantErr: true,
			errMsg:  "读取配置文件失败",
		},
		{
			name:    "无效的 JSON",
			content: `{invalid json}`,
			wantErr: true,
			errMsg:  "解析配置文件失败",
		},
		{
			name: "页面验证失败 - ID 为空",
			content: `{
				"pages": [
					{
						"id": "",
						"title": "测试",
						"entry": "test.html",
						"window": {
							"name": "test",
							"width": 800,
							"height": 600
						}
					}
				]
			}`,
			wantErr: true,
			errMsg:  "页面 ID 不能为空",
		},
		{
			name: "页面验证失败 - Title 为空",
			content: `{
				"pages": [
					{
						"id": "test",
						"title": "",
						"entry": "test.html",
						"window": {
							"name": "test",
							"width": 800,
							"height": 600
						}
					}
				]
			}`,
			wantErr: true,
			errMsg:  "页面标题不能为空",
		},
		{
			name: "页面验证失败 - Entry 为空",
			content: `{
				"pages": [
					{
						"id": "test",
						"title": "测试",
						"entry": "",
						"window": {
							"name": "test",
							"width": 800,
							"height": 600
						}
					}
				]
			}`,
			wantErr: true,
			errMsg:  "页面入口文件不能为空",
		},
		{
			name: "页面验证失败 - Window 为空",
			content: `{
				"pages": [
					{
						"id": "test",
						"title": "测试",
						"entry": "test.html",
						"window": null
					}
				]
			}`,
			wantErr: true,
			errMsg:  "窗口配置不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configPath string
			if tt.name == "文件不存在" {
				configPath = filepath.Join(tempDir, "nonexistent.json")
			} else {
				configPath = filepath.Join(tempDir, tt.name+".json")
				if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
					t.Fatalf("无法写入测试文件: %v", err)
				}
			}

			_, err := LoadPageConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadPageConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("LoadPageConfig() error = %q, 期望包含 %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// TestGetPageConfigPath 测试获取页面配置文件路径
func TestGetPageConfigPath(t *testing.T) {
	result := GetPageConfigPath()
	if result == "" {
		t.Error("GetPageConfigPath() 返回空字符串")
	}
	// 检查路径是否以 page.config.json 结尾（跨平台兼容）
	expectedFile := "page.config.json"
	if len(result) < len(expectedFile) || result[len(result)-len(expectedFile):] != expectedFile {
		t.Errorf("GetPageConfigPath() = %s, 期望以 %s 结尾", result, expectedFile)
	}
}

// BenchmarkPageConfigValidate 性能测试
func BenchmarkPageConfigValidate(b *testing.B) {
	config := createValidPageConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

// BenchmarkPageConfigFileGetPageConfig 性能测试
func BenchmarkPageConfigFileGetPageConfig(b *testing.B) {
	config := &PageConfigFile{
		Pages: []PageConfig{
			{ID: "page1", Title: "页面1", Entry: "page1.html", Window: &application.WebviewWindowOptions{Name: "w1", Width: 800, Height: 600}},
			{ID: "page2", Title: "页面2", Entry: "page2.html", Window: &application.WebviewWindowOptions{Name: "w2", Width: 800, Height: 600}},
			{ID: "page3", Title: "页面3", Entry: "page3.html", Window: &application.WebviewWindowOptions{Name: "w3", Width: 800, Height: 600}},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.GetPageConfig("page2")
	}
}
