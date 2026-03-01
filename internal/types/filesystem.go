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

package types

// ListDirectoryResult 列出目录结果
type ListDirectoryResult struct {
	BaseResult
	Data *ListDirectoryData `json:"data,omitempty"` // 目录数据
}

// ListDirectoryData 目录列表数据
type ListDirectoryData struct {
	CurrentPath string           `json:"currentPath"`           // 当前路径
	ParentPath  string           `json:"parentPath,omitempty"`  // 父目录路径
	Directories []*DirectoryInfo `json:"directories"`           // 目录列表
}

// DirectoryInfo 目录信息
type DirectoryInfo struct {
	Name        string           `json:"name"`                  // 目录名称
	Path        string           `json:"path"`                  // 完整路径
	DisplayName string           `json:"displayName"`           // 显示名称（用于 UI）
	IsParent    bool             `json:"isParent"`              // 是否是返回上一级
	Children    []*DirectoryInfo `json:"children,omitempty"`    // 子目录（可选，用于展开）
}
