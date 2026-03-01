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

package terminal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShortenPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("无法获取用户主目录")
	}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "home directory",
			path:     homeDir,
			expected: "~",
		},
		{
			name:     "subdirectory of home",
			path:     filepath.Join(homeDir, "Documents"),
			expected: "~/Documents",
		},
		{
			name:     "nested subdirectory of home",
			path:     filepath.Join(homeDir, "Documents", "Projects"),
			expected: "~/Documents/Projects",
		},
		{
			name:     "path outside home",
			path:     "/usr/local/bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "root directory",
			path:     "/",
			expected: "/",
		},
		{
			name:     "temp directory",
			path:     "/tmp",
			expected: "/tmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShortenPath(tt.path)
			if result != tt.expected {
				t.Errorf("ShortenPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestShortenPath_PartialMatch(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("无法获取用户主目录")
	}

	// 测试不会部分匹配的情况
	// 例如：homeDir = /Users/test，path = /Users/testing 不应该被缩短
	if homeDir != "/" {
		partialMatchPath := homeDir + "ing"
		result := ShortenPath(partialMatchPath)
		if result == "~ing" {
			t.Errorf("ShortenPath should not partially match home directory")
		}
	}
}

func TestGetPythonEnv(t *testing.T) {
	// 基本测试，不设置任何环境变量
	env := GetPythonEnv("/tmp")

	// 检查返回的结构体不为 nil
	if env == nil {
		t.Fatal("GetPythonEnv returned nil")
	}

	// 在没有 Python 的环境中，HasPython 可能为 false
	// 这里只检查字段可以访问
	t.Logf("Python env: HasPython=%v, Version=%s, EnvActive=%v, EnvType=%s, EnvName=%s",
		env.HasPython, env.Version, env.EnvActive, env.EnvType, env.EnvName)
}

func TestGetPythonEnv_WithVirtualEnv(t *testing.T) {
	// 注意：如果系统中有 Conda 激活，它会被优先检测
	// 所以这个测试主要验证函数不会崩溃

	env := GetPythonEnv("/tmp")

	if env == nil {
		t.Fatal("GetPythonEnv returned nil")
	}

	// 只检查返回值是有效的
	t.Logf("Python env: HasPython=%v, EnvActive=%v, EnvType=%s",
		env.HasPython, env.EnvActive, env.EnvType)
}

func TestGetPythonEnv_WithConda(t *testing.T) {
	// 保存原始环境变量
	originalConda := os.Getenv("CONDA_DEFAULT_ENV")
	originalCondaPrefix := os.Getenv("CONDA_PREFIX")
	originalVenv := os.Getenv("VIRTUAL_ENV")
	defer func() {
		os.Setenv("CONDA_DEFAULT_ENV", originalConda)
		os.Setenv("CONDA_PREFIX", originalCondaPrefix)
		os.Setenv("VIRTUAL_ENV", originalVenv)
	}()

	// 清除其他环境变量
	os.Setenv("VIRTUAL_ENV", "")
	// 设置 Conda 环境变量
	os.Setenv("CONDA_DEFAULT_ENV", "base")
	os.Setenv("CONDA_PREFIX", "/opt/anaconda3")

	env := GetPythonEnv("/tmp")

	if env == nil {
		t.Fatal("GetPythonEnv returned nil")
	}

	// 检查 Conda 环境被检测到（Conda 优先级高于 venv）
	if !env.EnvActive {
		t.Error("expected EnvActive to be true when CONDA_DEFAULT_ENV is set")
	}
	if env.EnvType != "conda" {
		t.Errorf("expected EnvType 'conda', got %s", env.EnvType)
	}
	if env.EnvName != "base" {
		t.Errorf("expected EnvName 'base', got %s", env.EnvName)
	}
}

func TestGetGitStatus(t *testing.T) {
	// 测试非 Git 目录
	status := GetGitStatus("/tmp")
	if status == nil {
		t.Fatal("GetGitStatus returned nil")
	}

	// /tmp 通常不是 Git 仓库
	t.Logf("Git status for /tmp: IsRepo=%v, Branch=%s", status.IsRepo, status.Branch)

	// 测试当前项目目录（应该是 Git 仓库）
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("无法获取当前工作目录")
	}

	status = GetGitStatus(cwd)
	if status == nil {
		t.Fatal("GetGitStatus returned nil")
	}

	t.Logf("Git status for %s: IsRepo=%v, Branch=%s, ModifiedFiles=%d, AddedLines=%d, DeletedLines=%d",
		cwd, status.IsRepo, status.Branch, status.ModifiedFiles, status.AddedLines, status.DeletedLines)
}

func TestGetGitStatus_EmptyPath(t *testing.T) {
	// 测试空路径
	// 注意：空路径时 git 命令会在当前目录执行，可能返回当前目录的 Git 状态
	status := GetGitStatus("")
	if status == nil {
		t.Fatal("GetGitStatus returned nil")
	}

	// 空路径时，结果取决于当前工作目录
	// 主要验证函数不会崩溃
	t.Logf("Git status for empty path: IsRepo=%v", status.IsRepo)
}

func TestGetEnvironmentInfo(t *testing.T) {
	// 测试获取环境信息
	info := GetEnvironmentInfo("/tmp")

	if info == nil {
		t.Fatal("GetEnvironmentInfo returned nil")
	}

	// 检查工作路径被缩短（如果 /tmp 不是用户主目录的子目录，则不会被缩短）
	t.Logf("Environment info: WorkPath=%s, PythonEnv=%v, GitInfo=%v",
		info.WorkPath, info.PythonEnv, info.GitInfo)

	// 检查 Python 环境信息
	if info.PythonEnv == nil {
		t.Error("PythonEnv should not be nil")
	}

	// 检查 Git 信息
	if info.GitInfo == nil {
		t.Error("GitInfo should not be nil")
	}
}

func TestGetEnvironmentInfo_WithHomePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("无法获取用户主目录")
	}

	info := GetEnvironmentInfo(homeDir)

	if info == nil {
		t.Fatal("GetEnvironmentInfo returned nil")
	}

	// 工作路径应该被缩短为 ~
	if info.WorkPath != "~" {
		t.Errorf("expected WorkPath '~', got %s", info.WorkPath)
	}
}
