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
