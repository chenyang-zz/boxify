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
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chenyang-zz/boxify/internal/types"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// FilesystemService 文件系统服务
type FilesystemService struct {
	BaseService
}

// NewFilesystemService 创建文件系统服务
func NewFilesystemService(deps *ServiceDeps) *FilesystemService {
	return &FilesystemService{
		BaseService: NewBaseService(deps),
	}
}

// ServiceStartup 服务启动
func (s *FilesystemService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.SetContext(ctx)
	s.Logger().Info("服务启动", "service", "FilesystemService")
	return nil
}

// ServiceShutdown 服务关闭
func (s *FilesystemService) ServiceShutdown() error {
	s.Logger().Info("服务关闭", "service", "FilesystemService")
	return nil
}

// ListDirectories 列出指定路径下的所有目录
func (s *FilesystemService) ListDirectories(path string) *types.ListDirectoryResult {
	result := &types.ListDirectoryResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取目录列表成功"},
		Data: &types.ListDirectoryData{
			CurrentPath: path,
			Directories: make([]*types.DirectoryInfo, 0),
		},
	}

	// 如果路径为空，使用当前工作目录
	if path == "" {
		wd, err := os.Getwd()
		if err != nil {
			result.Success = false
			result.Message = "获取当前工作目录失败: " + err.Error()
			s.Logger().Error("获取当前工作目录失败", "error", err)
			return result
		}
		path = wd
		result.Data.CurrentPath = path
	}

	// 规范化路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		result.Success = false
		result.Message = "获取绝对路径失败: " + err.Error()
		s.Logger().Error("获取绝对路径失败", "path", path, "error", err)
		return result
	}
	result.Data.CurrentPath = absPath

	// 获取父目录路径
	parentPath := filepath.Dir(absPath)
	if parentPath != absPath {
		result.Data.ParentPath = parentPath
	}

	// 读取目录
	entries, err := os.ReadDir(absPath)
	if err != nil {
		result.Success = false
		result.Message = "读取目录失败: " + err.Error()
		s.Logger().Error("读取目录失败", "path", absPath, "error", err)
		return result
	}

	// 筛选目录并排序
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 跳过隐藏目录（以 . 开头）
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(absPath, entry.Name())
		result.Data.Directories = append(result.Data.Directories, &types.DirectoryInfo{
			Name:        entry.Name(),
			Path:        fullPath,
			DisplayName: entry.Name(),
			IsParent:    false,
		})
	}

	// 按名称排序
	sort.Slice(result.Data.Directories, func(i, j int) bool {
		return result.Data.Directories[i].Name < result.Data.Directories[j].Name
	})

	return result
}

// ExpandPath 展开路径，将 ~ 替换为用户目录
func (s *FilesystemService) ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[1:])
	}
	return path
}
