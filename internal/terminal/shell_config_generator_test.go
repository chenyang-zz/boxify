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
	"strings"
	"testing"
)

func TestNewShellConfigGenerator(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	if gen == nil {
		t.Fatal("NewShellConfigGenerator returned nil")
	}
	if gen.tempDir != os.TempDir() {
		t.Errorf("expected tempDir %s, got %s", os.TempDir(), gen.tempDir)
	}
}

func TestShellConfigGenerator_GenerateConfig_Zsh(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	sessionID := "test123"

	configPath, err := gen.GenerateConfig(ShellTypeZsh, sessionID)
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}
	defer os.RemoveAll(configPath) // 清理目录

	// zsh 现在返回的是 ZDOTDIR 目录路径
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat config path: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected zsh config path to be a directory (ZDOTDIR), got %s", configPath)
	}

	// 检查目录名包含 sessionID
	if !strings.Contains(configPath, sessionID) {
		t.Errorf("expected config path to contain sessionID %s, got %s", sessionID, configPath)
	}

	// 检查 .zshrc 文件存在
	rcPath := filepath.Join(configPath, ".zshrc")
	content, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("failed to read .zshrc file: %v", err)
	}

	// 检查 zsh 配置内容
	contentStr := string(content)
	if !strings.Contains(contentStr, "source \"$HOME/.zshrc\"") {
		t.Error("zsh config should source user's .zshrc first")
	}
	if !strings.Contains(contentStr, "__boxify_preexec") {
		t.Error("zsh config should contain __boxify_preexec function")
	}
	if !strings.Contains(contentStr, "__boxify_precmd") {
		t.Error("zsh config should contain __boxify_precmd function")
	}
	if !strings.Contains(contentStr, "add-zsh-hook") {
		t.Error("zsh config should contain add-zsh-hook")
	}
	if !strings.Contains(contentStr, "\\e]133;A\\e\\\\") {
		t.Error("zsh config should contain OSC 133;A marker")
	}
	if !strings.Contains(contentStr, "\\e]133;D;") {
		t.Error("zsh config should contain OSC 133;D marker")
	}
}

func TestShellConfigGenerator_GenerateConfig_Bash(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	sessionID := "bash456"

	configPath, err := gen.GenerateConfig(ShellTypeBash, sessionID)
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}
	defer os.Remove(configPath)

	// 检查文件扩展名
	if !strings.HasSuffix(configPath, ".bash") {
		t.Errorf("expected .bash extension, got %s", configPath)
	}

	// 读取并检查内容
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "__boxify_preexec") {
		t.Error("bash config should contain __boxify_preexec function")
	}
	if !strings.Contains(contentStr, "PROMPT_COMMAND") {
		t.Error("bash config should set PROMPT_COMMAND")
	}
	if !strings.Contains(contentStr, "trap") {
		t.Error("bash config should use trap for DEBUG")
	}
}

func TestShellConfigGenerator_GenerateConfig_PowerShell(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	sessionID := "ps789"

	configPath, err := gen.GenerateConfig(ShellTypePowershell, sessionID)
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}
	defer os.Remove(configPath)

	// 检查文件扩展名
	if !strings.HasSuffix(configPath, ".ps1") {
		t.Errorf("expected .ps1 extension, got %s", configPath)
	}

	// 读取并检查内容
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "[char]27") {
		t.Error("powershell config should define ESC character")
	}
	if !strings.Contains(contentStr, "function global:prompt") {
		t.Error("powershell config should define prompt function")
	}
	if !strings.Contains(contentStr, "LASTEXITCODE") {
		t.Error("powershell config should handle LASTEXITCODE")
	}
}

func TestShellConfigGenerator_GenerateConfig_Pwsh(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	sessionID := "pwsh123"

	configPath, err := gen.GenerateConfig(ShellTypePwsh, sessionID)
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}
	defer os.Remove(configPath)

	// pwsh 应该使用与 powershell 相同的配置
	if !strings.HasSuffix(configPath, ".ps1") {
		t.Errorf("expected .ps1 extension for pwsh, got %s", configPath)
	}
}

func TestShellConfigGenerator_GenerateConfig_UnsupportedShell(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)

	unsupportedShells := []ShellType{ShellTypeCmd, ShellTypeSh, ShellTypeAuto}

	for _, shell := range unsupportedShells {
		t.Run(string(shell), func(t *testing.T) {
			_, err := gen.GenerateConfig(shell, "test")
			if err == nil {
				t.Errorf("expected error for unsupported shell type %s", shell)
			}
			if !strings.Contains(err.Error(), "不支持的 shell 类型") {
				t.Errorf("error message should mention unsupported shell type, got: %v", err)
			}
		})
	}
}

func TestShellConfigGenerator_Cleanup(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)

	// 创建临时文件
	configPath, err := gen.GenerateConfig(ShellTypeBash, "cleanup_test")
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file should exist")
	}

	// 清理文件
	if err := gen.Cleanup(configPath); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// 验证文件已删除
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("config file should be deleted after cleanup")
	}
}

func TestShellConfigGenerator_Cleanup_EmptyPath(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)

	// 空路径应该不报错
	err := gen.Cleanup("")
	if err != nil {
		t.Errorf("Cleanup with empty path should not return error, got: %v", err)
	}
}

func TestShellConfigGenerator_Cleanup_NonExistentFile(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)

	// 不存在的文件应该返回错误
	err := gen.Cleanup("/non/existent/file.bash")
	if err == nil {
		t.Error("Cleanup should return error for non-existent file")
	}
}

func TestShellConfigGenerator_GetShellArgs_Zsh(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	configPath := "/tmp/test_zdotdir"

	args, needsSource := gen.GetShellArgs(ShellTypeZsh, configPath)

	if needsSource {
		t.Error("zsh should not need additional source")
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg (-i), got %d: %v", len(args), args)
	}
	if args[0] != "-i" {
		t.Errorf("expected first arg -i, got %s", args[0])
	}
	// 注意：zsh 现在使用 ZDOTDIR 环境变量来加载配置，而不是通过命令行参数
}

func TestShellConfigGenerator_GetShellArgs_Bash(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	configPath := "/tmp/test.bash"

	args, needsSource := gen.GetShellArgs(ShellTypeBash, configPath)

	if needsSource {
		t.Error("bash should not need additional source")
	}
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
	if args[0] != "--rcfile" {
		t.Errorf("expected first arg --rcfile, got %s", args[0])
	}
	if args[1] != configPath {
		t.Errorf("expected config path as second arg, got %s", args[1])
	}
	if args[2] != "-i" {
		t.Errorf("expected -i as third arg, got %s", args[2])
	}
}

func TestShellConfigGenerator_GetShellArgs_PowerShell(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	configPath := "/tmp/test.ps1"

	args, needsSource := gen.GetShellArgs(ShellTypePowershell, configPath)

	if needsSource {
		t.Error("powershell should not need additional source")
	}
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
	if args[0] != "-NoExit" {
		t.Errorf("expected -NoExit, got %s", args[0])
	}
	if args[1] != "-Command" {
		t.Errorf("expected -Command, got %s", args[1])
	}
	if !strings.Contains(args[2], configPath) {
		t.Errorf("expected config path in command, got %s", args[2])
	}
}

func TestShellConfigGenerator_GetShellArgs_Pwsh(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	configPath := "/tmp/test.ps1"

	args, _ := gen.GetShellArgs(ShellTypePwsh, configPath)

	// pwsh 应该使用与 powershell 相同的参数
	if args[0] != "-NoExit" {
		t.Errorf("expected -NoExit for pwsh, got %s", args[0])
	}
}

func TestShellConfigGenerator_GetShellArgs_UnsupportedShell(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)

	unsupportedShells := []ShellType{ShellTypeCmd, ShellTypeSh, ShellTypeAuto}

	for _, shell := range unsupportedShells {
		t.Run(string(shell), func(t *testing.T) {
			args, needsSource := gen.GetShellArgs(shell, "/tmp/config")
			if args != nil {
				t.Errorf("expected nil args for unsupported shell %s, got %v", shell, args)
			}
			if needsSource {
				t.Errorf("expected false needsSource for unsupported shell %s", shell)
			}
		})
	}
}

func TestShellConfigGenerator_FilePermissions(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)

	configPath, err := gen.GenerateConfig(ShellTypeBash, "perm_test")
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}
	defer os.Remove(configPath)

	// 检查文件权限 (0600)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}

	expectedPerm := os.FileMode(0600)
	if info.Mode().Perm() != expectedPerm {
		t.Errorf("expected file permission %o, got %o", expectedPerm, info.Mode().Perm())
	}
}

func TestShellConfigGenerator_UniqueFilenames(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)

	// 生成多个配置文件
	paths := make([]string, 5)
	for i := 0; i < 5; i++ {
		sessionID := string(rune('a' + i))
		path, err := gen.GenerateConfig(ShellTypeBash, sessionID)
		if err != nil {
			t.Fatalf("GenerateConfig failed: %v", err)
		}
		paths[i] = path
		defer os.Remove(path)
	}

	// 检查所有路径都是唯一的
	for i := 0; i < len(paths); i++ {
		for j := i + 1; j < len(paths); j++ {
			if paths[i] == paths[j] {
				t.Errorf("config paths should be unique: %s", paths[i])
			}
		}
	}
}

func TestShellConfigGenerator_ZshConfigContent(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	content := gen.getZshRcContent()

	// 详细检查 zsh 配置内容
	requiredElements := []string{
		`source "$HOME/.zshrc"`, // 必须先加载用户配置
		"autoload -Uz add-zsh-hook",
		"add-zsh-hook preexec __boxify_preexec",
		"add-zsh-hook precmd __boxify_precmd",
		`printf '\e]133;A\e\\'`,
		`printf '\e]133;D;%s\e\\' "$?"`,
	}

	for _, elem := range requiredElements {
		if !strings.Contains(content, elem) {
			t.Errorf("zsh config missing required element: %s", elem)
		}
	}
}

func TestShellConfigGenerator_BashConfigContent(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	content := gen.getBashConfig()

	requiredElements := []string{
		"__boxify_in_command",
		"__boxify_preexec",
		"__boxify_prompt_command",
		"PROMPT_COMMAND",
		"trap '__boxify_preexec' DEBUG",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(content, elem) {
			t.Errorf("bash config missing required element: %s", elem)
		}
	}
}

func TestShellConfigGenerator_PowerShellConfigContent(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	content := gen.getPowerShellConfig()

	requiredElements := []string{
		"$__boxify_esc",
		"[char]27",
		"$__boxify_original_prompt",
		"function global:prompt",
		"$LASTEXITCODE",
		"PreCommandLookupAction",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(content, elem) {
			t.Errorf("powershell config missing required element: %s", elem)
		}
	}
}

func TestShellConfigGenerator_FileLocation(t *testing.T) {
	gen := NewShellConfigGenerator(testLogger)
	sessionID := "location_test"

	configPath, err := gen.GenerateConfig(ShellTypeBash, sessionID)
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}
	defer os.Remove(configPath)

	// 检查文件在 temp 目录中
	expectedDir := filepath.Clean(os.TempDir())
	actualDir := filepath.Dir(configPath)

	if actualDir != expectedDir {
		t.Errorf("expected config in %s, got %s", expectedDir, actualDir)
	}
}
