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
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// 错误模式常量（预编译小写版本，避免重复分配）
var (
	// Unix/Linux/macOS 错误模式
	unixErrorPatterns = []string{
		"command not found",
		"no such file or directory",
		"permission denied",
		"cannot execute",
		"is not recognized",
		"invalid option",
		"illegal option",
		"error:",
		"failed",
		"unable to",
	}

	// Windows cmd.exe 错误模式
	windowsCmdErrorPatterns = []string{
		"'\"' is not recognized",
		"is not recognized as an internal or external command",
		"cannot find the path specified",
		"the system cannot find the file specified",
		"access is denied",
		"the syntax of the command is incorrect",
	}

	// PowerShell 错误模式
	powerShellErrorPatterns = []string{
		"term '",
		"' is not recognized",
		"cannot be found",
		"does not exist",
		"access to the path",
		"is denied",
		"error:",
		"exception",
	}

	// 根据平台合并的错误模式（初始化时计算一次）
	platformErrorPatterns = func() []string {
		switch runtime.GOOS {
		case "windows":
			// Windows: 合并所有模式
			patterns := make([]string, 0, len(windowsCmdErrorPatterns)+len(powerShellErrorPatterns)+len(unixErrorPatterns))
			patterns = append(patterns, windowsCmdErrorPatterns...)
			patterns = append(patterns, powerShellErrorPatterns...)
			patterns = append(patterns, unixErrorPatterns...) // 也包含 Unix 模式（WSL/Git Bash）
			return patterns
		default:
			// Unix/Linux/macOS
			return unixErrorPatterns
		}
	}()
)

// ShellDetector shell 检测器
type ShellDetector struct {
	cache sync.Map // 缓存 shell 路径: cacheKey -> shellPath
}

// NewShellDetector 创建 shell 检测器
func NewShellDetector() *ShellDetector {
	return &ShellDetector{}
}

// DetectShell 检测系统默认 shell 并返回路径（带缓存优化）
func (d *ShellDetector) DetectShell(preferred ShellType) string {
	if preferred != ShellTypeAuto {
		// 对于指定的 shell 类型，查找其路径
		cacheKey := "shell:" + string(preferred)
		if cached, ok := d.cache.Load(cacheKey); ok {
			return cached.(string)
		}

		shellName := string(preferred)
		if preferred == ShellTypePowershell {
			shellName = "powershell"
		}

		// Windows shell 需要加 .exe
		if runtime.GOOS == "windows" && !strings.HasSuffix(shellName, ".exe") {
			if preferred != ShellTypeCmd {
				shellName += ".exe"
			} else {
				shellName = "cmd.exe"
			}
		}

		path, err := exec.LookPath(shellName)
		if err != nil {
			return shellName
		}
		d.cache.Store(cacheKey, path)
		return path
	}

	// 检查缓存
	cacheKey := "default:" + runtime.GOOS
	if cached, ok := d.cache.Load(cacheKey); ok {
		return cached.(string)
	}

	// 根据平台检测默认 shell
	var shellPath string
	switch runtime.GOOS {
	case "windows":
		// Windows: 优先 PowerShell，其次 cmd
		if path, err := exec.LookPath("pwsh"); err == nil {
			shellPath = path
		} else if path, err := exec.LookPath("powershell"); err == nil {
			shellPath = path
		} else {
			shellPath = "cmd.exe"
		}

	case "darwin", "linux":
		// Unix: 优先 zsh，其次 bash
		if path, err := exec.LookPath("zsh"); err == nil {
			shellPath = path
		} else if path, err := exec.LookPath("bash"); err == nil {
			shellPath = path
		} else {
			shellPath = "/bin/sh"
		}

	default:
		shellPath = "/bin/sh"
	}

	// 缓存结果
	d.cache.Store(cacheKey, shellPath)
	return shellPath
}

// DetectShellTypeFromPath 从 shell 路径检测 shell 类型
func (d *ShellDetector) DetectShellTypeFromPath(shellPath string) ShellType {
	// 提取文件名（不含路径和扩展名）
	base := strings.ToLower(filepath.Base(shellPath))
	base = strings.TrimSuffix(base, ".exe")

	switch base {
	case "zsh":
		return ShellTypeZsh
	case "bash":
		return ShellTypeBash
	case "powershell":
		return ShellTypePowershell
	case "pwsh":
		return ShellTypePwsh
	case "cmd":
		return ShellTypeCmd
	case "sh":
		return ShellTypeSh
	default:
		// 默认根据平台返回
		if runtime.GOOS == "windows" {
			return ShellTypeCmd
		}
		return ShellTypeSh
	}
}

// HasCommandError 检查命令输出中是否包含错误标志（优化版：使用预编译模式）
func HasCommandError(output string) bool {
	lowerOutput := strings.ToLower(output)

	// 使用预编译的平台特定错误模式
	for _, pattern := range platformErrorPatterns {
		if strings.Contains(lowerOutput, pattern) {
			return true
		}
	}

	return false
}
