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
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var (
	unixEnvAssignRe = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.+)$`)
	powerShellEnvRe = regexp.MustCompile(`(?i)^\$env:([A-Za-z_][A-Za-z0-9_]*)\s*(\+?=)\s*(.+)$`)
	cmdSetEnvRe     = regexp.MustCompile(`(?i)^set\s+([A-Za-z_][A-Za-z0-9_]*)=(.+)$`)
	unixBraceVarRe  = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
	unixSimpleVarRe = regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
	powerShellVarRe = regexp.MustCompile(`(?i)\$env:([A-Za-z_][A-Za-z0-9_]*)`)
	cmdPercentVarRe = regexp.MustCompile(`(?i)%([A-Za-z_][A-Za-z0-9_]*)%`)
)

// ShellEnvironmentLoader 负责按 shell 配置文件构建环境变量集合。
type ShellEnvironmentLoader struct {
	logger *slog.Logger
	goos   string
}

// NewShellEnvironmentLoader 创建环境加载器。
func NewShellEnvironmentLoader(logger *slog.Logger, goos string) *ShellEnvironmentLoader {
	if logger == nil {
		logger = slog.Default()
	}
	if strings.TrimSpace(goos) == "" {
		goos = runtime.GOOS
	}

	return &ShellEnvironmentLoader{
		logger: logger,
		goos:   goos,
	}
}

// LoadForShell 读取 shell 配置文件并应用环境变量覆盖。
func (l *ShellEnvironmentLoader) LoadForShell(shellType ShellType) map[string]string {
	env := l.envMapFromProcess()
	homeDir, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(homeDir) == "" {
		if err != nil {
			l.logger.Warn("读取用户目录失败，回退到当前进程环境", "error", err)
		}
		return env
	}

	for _, configPath := range l.shellConfigCandidates(shellType, homeDir) {
		content, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			l.logger.Warn("读取 shell 配置文件失败", "shellType", shellType, "configPath", configPath, "error", err)
			continue
		}

		l.applyShellEnvConfig(content, shellType, env)
		l.logger.Debug("已加载 shell 配置环境变量", "shellType", shellType, "configPath", configPath)
	}

	return env
}

// GetEnv 从环境变量 map 读取值，Windows 下大小写不敏感。
func (l *ShellEnvironmentLoader) GetEnv(env map[string]string, key string) string {
	if l.goos == "windows" {
		return env[strings.ToUpper(key)]
	}
	return env[key]
}

// SetEnv 写入环境变量 map，Windows 下大小写不敏感。
func (l *ShellEnvironmentLoader) SetEnv(env map[string]string, key, value string) {
	if l.goos == "windows" {
		env[strings.ToUpper(key)] = value
		return
	}
	env[key] = value
}

// shellConfigCandidates 返回指定 shell 可能使用的配置文件路径列表。
func (l *ShellEnvironmentLoader) shellConfigCandidates(shellType ShellType, homeDir string) []string {
	switch shellType {
	case ShellTypeZsh:
		return []string{
			filepath.Join(homeDir, ".zshenv"),
			filepath.Join(homeDir, ".zprofile"),
			filepath.Join(homeDir, ".zshrc"),
			filepath.Join(homeDir, ".zlogin"),
		}
	case ShellTypeBash:
		return []string{
			filepath.Join(homeDir, ".bash_profile"),
			filepath.Join(homeDir, ".bash_login"),
			filepath.Join(homeDir, ".profile"),
			filepath.Join(homeDir, ".bashrc"),
		}
	case ShellTypeSh:
		return []string{filepath.Join(homeDir, ".profile")}
	case ShellTypePowershell, ShellTypePwsh:
		if l.goos == "windows" {
			return []string{
				filepath.Join(homeDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"),
				filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"),
				filepath.Join(homeDir, "Documents", "PowerShell", "profile.ps1"),
			}
		}
		return []string{
			filepath.Join(homeDir, ".config", "powershell", "Microsoft.PowerShell_profile.ps1"),
			filepath.Join(homeDir, ".config", "powershell", "profile.ps1"),
		}
	default:
		return nil
	}
}

// applyShellEnvConfig 按 shell 语法解析配置文件中的环境变量赋值。
func (l *ShellEnvironmentLoader) applyShellEnvConfig(content []byte, shellType ShellType, env map[string]string) {
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		switch shellType {
		case ShellTypePowershell, ShellTypePwsh:
			l.parsePowerShellEnvLine(trimmed, env)
		case ShellTypeCmd:
			l.parseCmdEnvLine(trimmed, env)
		default:
			l.parseUnixEnvLine(trimmed, env)
		}
	}
}

// parseUnixEnvLine 解析 Unix shell 环境变量赋值语句。
func (l *ShellEnvironmentLoader) parseUnixEnvLine(line string, env map[string]string) {
	if strings.HasPrefix(line, "export ") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
	}

	match := unixEnvAssignRe.FindStringSubmatch(line)
	if len(match) != 3 {
		return
	}

	key := strings.TrimSpace(match[1])
	rawValue := strings.TrimSpace(match[2])
	value := normalizeAssignmentValue(rawValue)
	if !(strings.HasPrefix(rawValue, "'") && strings.HasSuffix(rawValue, "'")) {
		value = l.expandUnixVariables(value, env)
	}
	l.SetEnv(env, key, value)
}

// parsePowerShellEnvLine 解析 PowerShell 的 $env:KEY 赋值语句。
func (l *ShellEnvironmentLoader) parsePowerShellEnvLine(line string, env map[string]string) {
	match := powerShellEnvRe.FindStringSubmatch(line)
	if len(match) != 4 {
		return
	}

	key := strings.TrimSpace(match[1])
	op := strings.TrimSpace(match[2])
	rawValue := strings.TrimSpace(match[3])
	value := l.expandPowerShellVariables(normalizeAssignmentValue(rawValue), env)
	if op == "+=" {
		value = l.GetEnv(env, key) + value
	}
	l.SetEnv(env, key, value)
}

// parseCmdEnvLine 解析 CMD 的 set KEY=VALUE 赋值语句。
func (l *ShellEnvironmentLoader) parseCmdEnvLine(line string, env map[string]string) {
	match := cmdSetEnvRe.FindStringSubmatch(line)
	if len(match) != 3 {
		return
	}

	key := strings.TrimSpace(match[1])
	rawValue := strings.TrimSpace(match[2])
	value := l.expandCmdVariables(normalizeAssignmentValue(rawValue), env)
	l.SetEnv(env, key, value)
}

// normalizeAssignmentValue 去除常见赋值语句的包裹引号。
func normalizeAssignmentValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 2 {
		if strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"") {
			return trimmed[1 : len(trimmed)-1]
		}
		if strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'") {
			return trimmed[1 : len(trimmed)-1]
		}
	}
	return trimmed
}

// expandUnixVariables 展开 Unix 风格变量引用。
func (l *ShellEnvironmentLoader) expandUnixVariables(value string, env map[string]string) string {
	result := unixBraceVarRe.ReplaceAllStringFunc(value, func(expr string) string {
		match := unixBraceVarRe.FindStringSubmatch(expr)
		if len(match) != 2 {
			return expr
		}
		return l.GetEnv(env, match[1])
	})

	result = unixSimpleVarRe.ReplaceAllStringFunc(result, func(expr string) string {
		match := unixSimpleVarRe.FindStringSubmatch(expr)
		if len(match) != 2 {
			return expr
		}
		return l.GetEnv(env, match[1])
	})

	return result
}

// expandPowerShellVariables 展开 PowerShell 风格变量引用。
func (l *ShellEnvironmentLoader) expandPowerShellVariables(value string, env map[string]string) string {
	return powerShellVarRe.ReplaceAllStringFunc(value, func(expr string) string {
		match := powerShellVarRe.FindStringSubmatch(expr)
		if len(match) != 2 {
			return expr
		}
		return l.GetEnv(env, match[1])
	})
}

// expandCmdVariables 展开 CMD 风格变量引用。
func (l *ShellEnvironmentLoader) expandCmdVariables(value string, env map[string]string) string {
	return cmdPercentVarRe.ReplaceAllStringFunc(value, func(expr string) string {
		match := cmdPercentVarRe.FindStringSubmatch(expr)
		if len(match) != 2 {
			return expr
		}
		return l.GetEnv(env, match[1])
	})
}

// envMapFromProcess 把当前进程环境拷贝到 map，便于后续覆盖。
func (l *ShellEnvironmentLoader) envMapFromProcess() map[string]string {
	env := make(map[string]string)
	for _, item := range os.Environ() {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			continue
		}
		l.SetEnv(env, parts[0], parts[1])
	}
	return env
}
