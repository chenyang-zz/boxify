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
	"path/filepath"

	"log/slog"
)

// ShellConfigGenerator 生成 shell 临时配置文件
// 用于注入 Shell Hooks 以输出 OSC 133 标记
type ShellConfigGenerator struct {
	tempDir string
	logger  *slog.Logger
}

// NewShellConfigGenerator 创建配置生成器
func NewShellConfigGenerator(logger *slog.Logger) *ShellConfigGenerator {
	return &ShellConfigGenerator{
		tempDir: os.TempDir(),
		logger:  logger,
	}
}

// GenerateConfig 生成指定 shell 的临时配置文件
// 返回配置文件路径（zsh 返回 ZDOTDIR 目录路径），需要在会话结束时清理
func (g *ShellConfigGenerator) GenerateConfig(shellType ShellType, sessionID string) (string, error) {
	switch shellType {
	case ShellTypeZsh:
		// zsh 使用 ZDOTDIR 方式，创建临时目录和 .zshrc
		return g.generateZshConfig(sessionID)
	case ShellTypeBash:
		return g.generateBashConfig(sessionID)
	case ShellTypePowershell, ShellTypePwsh:
		return g.generatePowerShellConfig(sessionID)
	default:
		return "", fmt.Errorf("不支持的 shell 类型: %s", shellType)
	}
}

// generateZshConfig 为 zsh 创建临时目录和 .zshrc（使用 ZDOTDIR 机制）
func (g *ShellConfigGenerator) generateZshConfig(sessionID string) (string, error) {
	// 创建临时目录
	tempDir := filepath.Join(g.tempDir, fmt.Sprintf("boxify_zsh_%s", sessionID))
	if err := os.MkdirAll(tempDir, 0700); err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 生成 .zshrc 内容
	content := g.getZshRcContent()

	// 写入 .zshrc 文件
	rcPath := filepath.Join(tempDir, ".zshrc")
	if err := os.WriteFile(rcPath, []byte(content), 0600); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("写入 .zshrc 失败: %w", err)
	}

	g.logger.Info("shell配置信息加载成功", "shellType", ShellTypeZsh, "configPath", tempDir)

	return tempDir, nil
}

// generateBashConfig 为 bash 创建临时配置文件
func (g *ShellConfigGenerator) generateBashConfig(sessionID string) (string, error) {
	content := g.getBashConfig()
	filename := fmt.Sprintf("boxify_shell_%s.bash", sessionID)
	configPath := filepath.Join(g.tempDir, filename)

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("写入配置文件失败: %w", err)
	}

	g.logger.Info("shell配置信息加载成功", "shellType", ShellTypeBash, "configPath", configPath)

	return configPath, nil
}

// generatePowerShellConfig 为 PowerShell 创建临时配置文件
func (g *ShellConfigGenerator) generatePowerShellConfig(sessionID string) (string, error) {
	content := g.getPowerShellConfig()
	filename := fmt.Sprintf("boxify_shell_%s.ps1", sessionID)
	configPath := filepath.Join(g.tempDir, filename)

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("写入配置文件失败: %w", err)
	}

	g.logger.Info("shell配置信息加载成功", "shellType", ShellTypePowershell, "configPath", configPath)

	return configPath, nil
}

// Cleanup 清理临时配置文件或目录
func (g *ShellConfigGenerator) Cleanup(configPath string) error {
	if configPath == "" {
		return nil
	}

	// 检查是否是目录（zsh 的 ZDOTDIR）
	info, err := os.Stat(configPath)
	if err == nil && info.IsDir() {
		return os.RemoveAll(configPath)
	}

	// 单个文件清理
	return os.Remove(configPath)
}

// GetShellArgs 获取启动 shell 时需要的参数
// 返回: (shell参数, 是否需要source配置)
func (g *ShellConfigGenerator) GetShellArgs(shellType ShellType, configPath string) ([]string, bool) {
	switch shellType {
	case ShellTypeZsh:
		// zsh: 直接启动交互式 shell，依赖 ZDOTDIR 环境变量加载配置
		return []string{"-i"}, false
	case ShellTypeBash:
		// bash: 使用 --rcfile 指定自定义配置文件
		return []string{"--rcfile", configPath, "-i"}, false
	case ShellTypePowershell, ShellTypePwsh:
		// PowerShell: 使用 -NoExit -Command 加载配置
		return []string{"-NoExit", "-Command", fmt.Sprintf(". %s", configPath)}, false
	default:
		return nil, false
	}
}

// getZshRcContent 生成 zsh 的 .zshrc 内容
// 先加载用户的 .zshrc，再注册 boxify hooks
func (g *ShellConfigGenerator) getZshRcContent() string {
	return `# Boxify Shell Integration for Zsh
# 此文件由 Boxify 自动生成，请勿手动修改

# 先加载用户的 .zshrc（如果存在）
if [[ -f "$HOME/.zshrc" ]]; then
    source "$HOME/.zshrc"
fi

# 命令执行前调用
__boxify_preexec() {
    printf '\e]133;A\e\\'
}

# 命令执行后、下一个提示符前调用
__boxify_precmd() {
    # 输出当前工作路径（OSC 1337;Pwd 序列）
    local pwd="${PWD/#$HOME/~}"
    local encoded=$(printf '%s' "$pwd" | base64)
    printf '\e]1337;Pwd;%s\e\\' "$encoded"

    # 命令结束标记
    printf '\e]133;D;%s\e\\' "$?"
}

# 注册 hooks
autoload -Uz add-zsh-hook
add-zsh-hook preexec __boxify_preexec
add-zsh-hook precmd __boxify_precmd
`
}

func (g *ShellConfigGenerator) getBashConfig() string {
	return `# Boxify Shell Integration for Bash
# 此文件由 Boxify 自动生成，请勿手动修改

# 命令执行前调用（使用 DEBUG trap）
__boxify_preexec() {
    if [[ -z "$__boxify_in_command" ]]; then
        __boxify_in_command=1
        printf '\e]133;A\e\\'
    fi
}

# 命令执行后调用
__boxify_prompt_command() {
    __boxify_in_command=""

    # 输出当前工作路径（OSC 1337;Pwd 序列）
    local pwd="${PWD/#$HOME/~}"
    local encoded=$(printf '%s' "$pwd" | base64)
    printf '\e]1337;Pwd;%s\e\\' "$encoded"

    # 命令结束标记
    printf '\e]133;D;%s\e\\' "$?"
}

# 设置 PROMPT_COMMAND
PROMPT_COMMAND="__boxify_prompt_command${PROMPT_COMMAND:+;$PROMPT_COMMAND}"

# 注册 DEBUG trap
trap '__boxify_preexec' DEBUG
`
}

func (g *ShellConfigGenerator) getPowerShellConfig() string {
	// PowerShell 使用 ` (反引号) 作为转义字符，需要特殊处理
	// ESC 字符在 PowerShell 中是 `e (PowerShell 7+) 或 $([char]27)
	return "# Boxify Shell Integration for PowerShell\n" +
		"# 此文件由 Boxify 自动生成，请勿手动修改\n" +
		"\n" +
		"# 定义 ESC 字符\n" +
		"$__boxify_esc = [char]27\n" +
		"\n" +
		"# 保存原始 prompt 函数\n" +
		"$__boxify_original_prompt = ${function:prompt}\n" +
		"\n" +
		"# 定义新的 prompt 函数\n" +
		"function global:prompt {\n" +
		"    # 输出命令结束标记\n" +
		"    $exitCode = $LASTEXITCODE\n" +
		"    if ($null -eq $exitCode) { $exitCode = 0 }\n" +
		"    Write-Host \"$__boxify_esc]133;D;$exitCode$__boxify_esc\\\" -NoNewline\n" +
		"\n" +
		"    # 调用原始 prompt\n" +
		"    & $__boxify_original_prompt\n" +
		"}\n" +
		"\n" +
		"# 注册预执行钩子（PowerShell 7.2+）\n" +
		"if ($PSVersionTable.PSVersion.Major -ge 7) {\n" +
		"    $ExecutionContext.SessionState.InvokeCommand.PreCommandLookupAction = {\n" +
		"        param($commandName, $commandLookupEventArgs)\n" +
		"        Write-Host \"$__boxify_esc]133;A$__boxify_esc\\\" -NoNewline\n" +
		"    }\n" +
		"}\n"
}
