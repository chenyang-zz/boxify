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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chenyang-zz/boxify/internal/types"
)

// GetEnvironmentInfo 获取终端环境信息
func GetEnvironmentInfo(workPath string) *types.TerminalEnvironmentInfo {
	info := &types.TerminalEnvironmentInfo{
		WorkPath: ShortenPath(workPath),
	}

	// 获取 Python 环境信息
	info.PythonEnv = GetPythonEnv(workPath)

	// 获取 Git 信息（注意：这里只获取初始信息，实时更新由 GitWatcher 负责）
	gitStatus := GetGitStatus(workPath)
	info.GitInfo = &types.GitInfo{
		IsRepo:        gitStatus.IsRepo,
		Branch:        gitStatus.Branch,
		ModifiedFiles: gitStatus.ModifiedFiles,
		AddedLines:    gitStatus.AddedLines,
		DeletedLines:  gitStatus.DeletedLines,
	}

	return info
}

// ShortenPath 缩短路径，将用户目录替换为 ~
func ShortenPath(path string) string {
	if path == "" {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	// 如果路径是用户目录或其子目录，替换为 ~
	if path == homeDir {
		return "~"
	}

	// 确保路径以分隔符结尾，避免部分匹配
	homeDirWithSlash := homeDir + string(filepath.Separator)
	if strings.HasPrefix(path, homeDirWithSlash) {
		return "~" + path[len(homeDir):]
	}

	return path
}

// GetPythonEnv 获取 Python 环境信息
func GetPythonEnv(workPath string) *types.PythonEnv {
	env := &types.PythonEnv{}

	// 检查 Python 是否安装
	cmd := exec.Command("python3", "--version")
	cmd.Dir = workPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 尝试 python 命令
		cmd = exec.Command("python", "--version")
		cmd.Dir = workPath
		output, err = cmd.CombinedOutput()
		if err != nil {
			return env
		}
	}

	env.HasPython = true
	env.Version = strings.TrimSpace(string(output))

	// 检测虚拟环境（按优先级检测）
	// 1. 检测 Conda 环境
	if condaEnv := os.Getenv("CONDA_DEFAULT_ENV"); condaEnv != "" {
		env.EnvActive = true
		env.EnvType = "conda"
		env.EnvName = condaEnv
		env.EnvPath = os.Getenv("CONDA_PREFIX")
		return env
	}

	// 2. 检测 Pipenv 环境
	if os.Getenv("PIPENV_ACTIVE") != "" {
		env.EnvActive = true
		env.EnvType = "pipenv"
		env.EnvName = filepath.Base(workPath)
		env.EnvPath = os.Getenv("VIRTUAL_ENV")
		return env
	}

	// 3. 检测 Poetry 环境
	if os.Getenv("POETRY_ACTIVE") != "" {
		env.EnvActive = true
		env.EnvType = "poetry"
		env.EnvName = filepath.Base(workPath)
		env.EnvPath = os.Getenv("VIRTUAL_ENV")
		return env
	}

	// 4. 检测 venv/virtualenv 环境
	if venvPath := os.Getenv("VIRTUAL_ENV"); venvPath != "" {
		env.EnvActive = true
		env.EnvType = "venv"
		env.EnvPath = venvPath
		env.EnvName = filepath.Base(venvPath)
		return env
	}

	return env
}

// GetGitStatus 获取 Git 状态信息（静态方法，不监听）
func GetGitStatus(workPath string) *types.GitInfo {
	status := &types.GitInfo{}

	// 检查是否是 Git 仓库
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = workPath
	output, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) != "true" {
		return status
	}

	status.IsRepo = true

	// 获取当前分支
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = workPath
	output, err = cmd.Output()
	if err == nil {
		status.Branch = strings.TrimSpace(string(output))
	}

	// 获取修改统计
	cmd = exec.Command("git", "diff", "--numstat")
	cmd.Dir = workPath
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			status.ModifiedFiles++
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				// 新增行数（第一列）
				if parts[0] != "-" {
					var added int
					fmt.Sscanf(parts[0], "%d", &added)
					status.AddedLines += added
				}
				// 删除行数（第二列）
				if parts[1] != "-" {
					var deleted int
					fmt.Sscanf(parts[1], "%d", &deleted)
					status.DeletedLines += deleted
				}
			}
		}
	}

	// 获取暂存区修改统计
	cmd = exec.Command("git", "diff", "--cached", "--numstat")
	cmd.Dir = workPath
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			status.ModifiedFiles++
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				if parts[0] != "-" {
					var added int
					fmt.Sscanf(parts[0], "%d", &added)
					status.AddedLines += added
				}
				if parts[1] != "-" {
					var deleted int
					fmt.Sscanf(parts[1], "%d", &deleted)
					status.DeletedLines += deleted
				}
			}
		}
	}

	return status
}
