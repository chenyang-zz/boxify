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
	"log/slog"
	"strings"
)

// OSC 133 标记常量
const (
	// StartMarker 命令开始标记
	StartMarker = "\x1b]133;A\x1b\\"
	// EndMarkerTemplate 命令结束标记模板（需要填充退出码）
	EndMarkerTemplate = "\x1b]133;D;%s\x1b\\"
)

// ShellType 定义 shell 类型
type ShellType string

const (
	ShellTypeCmd        ShellType = "cmd"
	ShellTypePowershell ShellType = "powershell"
	ShellTypePwsh       ShellType = "pwsh"
	ShellTypeBash       ShellType = "bash"
	ShellTypeZsh        ShellType = "zsh"
	ShellTypeSh         ShellType = "sh"
	ShellTypeAuto       ShellType = "auto"
)

// SupportsHooks 检查 shell 是否支持 hooks
func (s ShellType) SupportsHooks() bool {
	switch s {
	case ShellTypeZsh, ShellTypeBash, ShellTypePowershell, ShellTypePwsh:
		return true
	default:
		return false
	}
}

// CommandWrapper 命令包装器
// 为命令添加 OSC 133 标记，用于在不支持 hooks 的 shell 中识别命令边界
type CommandWrapper struct {
	shellType ShellType
	logger    *slog.Logger
}

// NewCommandWrapper 创建命令包装器
func NewCommandWrapper(shellType ShellType, logger *slog.Logger) *CommandWrapper {
	return &CommandWrapper{
		shellType: shellType,
		logger:    logger,
	}
}

// Wrap 包装命令，添加开始和结束标记
func (w *CommandWrapper) Wrap(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return command
	}

	switch w.shellType {
	case ShellTypeCmd:
		return w.wrapForCmd(command)
	case ShellTypePowershell, ShellTypePwsh:
		return w.wrapForPowerShell(command)
	default:
		// Unix shells (bash, zsh, sh)
		return w.wrapForUnixShell(command)
	}
}

// wrapForUnixShell 为 Unix shell 包装命令
func (w *CommandWrapper) wrapForUnixShell(command string) string {
	// 使用 printf 输出标记，避免 echo 的平台差异
	// 将 OSC 133 标记常量转换为 shell printf 格式
	startMarkerPrintf := "'\\e]133;A\\e\\\\'"
	endMarkerPrintf := "'\\e]133;D;%s\\e\\\\'"
	return fmt.Sprintf("printf %s; %s; printf %s \"$?\"", startMarkerPrintf, command, endMarkerPrintf)
}

// wrapForPowerShell 为 PowerShell 包装命令
func (w *CommandWrapper) wrapForPowerShell(command string) string {
	// PowerShell 使用 Write-Host 输出标记
	// $LASTEXITCODE 用于获取退出码
	// ESC 字符: $([char]27)
	return fmt.Sprintf("Write-Host \"$([char]27]133;A$([char]27\\\" -NoNewline; %s; Write-Host \"$([char]27]133;D;$($LASTEXITCODE ?? 0)$([char]27\\\" -NoNewline", command)
}

// wrapForCmd 为 CMD 包装命令
func (w *CommandWrapper) wrapForCmd(command string) string {
	// CMD 的处理比较复杂，因为：
	// 1. echo 不支持转义序列
	// 2. 需要使用 prompt 或其他方式
	// 这里使用一种简化的方式：通过特殊标记识别
	// 实际上 CMD 很难完美支持，建议用户使用 PowerShell

	// 使用特殊的可见标记（因为 CMD 处理 ANSI 序列有限）
	// 用户应该优先使用 PowerShell
	return fmt.Sprintf("echo BOXIFY_CMD_START & %s & echo BOXIFY_CMD_END", command)
}
