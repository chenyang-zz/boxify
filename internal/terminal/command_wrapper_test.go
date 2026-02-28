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
	"strings"
	"testing"
)

func TestShellType_SupportsHooks(t *testing.T) {
	tests := []struct {
		name     string
		shell    ShellType
		expected bool
	}{
		{"zsh supports hooks", ShellTypeZsh, true},
		{"bash supports hooks", ShellTypeBash, true},
		{"powershell supports hooks", ShellTypePowershell, true},
		{"pwsh supports hooks", ShellTypePwsh, true},
		{"cmd does not support hooks", ShellTypeCmd, false},
		{"sh does not support hooks", ShellTypeSh, false},
		{"auto does not support hooks", ShellTypeAuto, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.shell.SupportsHooks(); got != tt.expected {
				t.Errorf("ShellType(%q).SupportsHooks() = %v, want %v", tt.shell, got, tt.expected)
			}
		})
	}
}

func TestCommandWrapper_Wrap(t *testing.T) {
	tests := []struct {
		name      string
		shellType ShellType
		command   string
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:      "empty command returns empty",
			shellType: ShellTypeBash,
			command:   "",
			checkFunc: func(t *testing.T, result string) {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
				}
			},
		},
		{
			name:      "whitespace only command returns empty",
			shellType: ShellTypeBash,
			command:   "   \t\n  ",
			checkFunc: func(t *testing.T, result string) {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
				}
			},
		},
		{
			name:      "bash wraps with OSC 133 markers",
			shellType: ShellTypeBash,
			command:   "echo hello",
			checkFunc: func(t *testing.T, result string) {
				// 检查包含开始标记
				if !strings.Contains(result, "'\\e]133;A\\e\\\\'") {
					t.Errorf("expected start marker in result, got %q", result)
				}
				// 检查包含结束标记
				if !strings.Contains(result, "'\\e]133;D;%s\\e\\\\'") {
					t.Errorf("expected end marker in result, got %q", result)
				}
				// 检查包含原始命令
				if !strings.Contains(result, "echo hello") {
					t.Errorf("expected original command in result, got %q", result)
				}
			},
		},
		{
			name:      "zsh wraps with OSC 133 markers",
			shellType: ShellTypeZsh,
			command:   "ls -la",
			checkFunc: func(t *testing.T, result string) {
				// zsh 使用相同的 Unix shell 包装逻辑
				if !strings.Contains(result, "'\\e]133;A\\e\\\\'") {
					t.Errorf("expected start marker in result, got %q", result)
				}
				if !strings.Contains(result, "ls -la") {
					t.Errorf("expected original command in result, got %q", result)
				}
			},
		},
		{
			name:      "sh wraps with OSC 133 markers",
			shellType: ShellTypeSh,
			command:   "pwd",
			checkFunc: func(t *testing.T, result string) {
				if !strings.Contains(result, "'\\e]133;A\\e\\\\'") {
					t.Errorf("expected start marker in result, got %q", result)
				}
				if !strings.Contains(result, "pwd") {
					t.Errorf("expected original command in result, got %q", result)
				}
			},
		},
		{
			name:      "powershell wraps with Write-Host markers",
			shellType: ShellTypePowershell,
			command:   "Get-Process",
			checkFunc: func(t *testing.T, result string) {
				// 检查包含 PowerShell 风格的标记
				if !strings.Contains(result, "Write-Host") {
					t.Errorf("expected Write-Host in result, got %q", result)
				}
				if !strings.Contains(result, "[char]27") {
					t.Errorf("expected [char]27 in result, got %q", result)
				}
				if !strings.Contains(result, "Get-Process") {
					t.Errorf("expected original command in result, got %q", result)
				}
			},
		},
		{
			name:      "pwsh wraps same as powershell",
			shellType: ShellTypePwsh,
			command:   "Write-Output test",
			checkFunc: func(t *testing.T, result string) {
				if !strings.Contains(result, "Write-Host") {
					t.Errorf("expected Write-Host in result, got %q", result)
				}
				if !strings.Contains(result, "Write-Output test") {
					t.Errorf("expected original command in result, got %q", result)
				}
			},
		},
		{
			name:      "cmd wraps with echo markers",
			shellType: ShellTypeCmd,
			command:   "dir",
			checkFunc: func(t *testing.T, result string) {
				// 检查包含 CMD 风格的标记
				if !strings.Contains(result, "echo BOXIFY_CMD_START") {
					t.Errorf("expected BOXIFY_CMD_START marker in result, got %q", result)
				}
				if !strings.Contains(result, "echo BOXIFY_CMD_END") {
					t.Errorf("expected BOXIFY_CMD_END marker in result, got %q", result)
				}
				if !strings.Contains(result, "dir") {
					t.Errorf("expected original command in result, got %q", result)
				}
			},
		},
		{
			name:      "command with leading/trailing spaces is trimmed",
			shellType: ShellTypeBash,
			command:   "  echo test  ",
			checkFunc: func(t *testing.T, result string) {
				// 命令应该被 trim 后再包装
				if !strings.Contains(result, "echo test") {
					t.Errorf("expected trimmed command in result, got %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := NewCommandWrapper(tt.shellType)
			result := wrapper.Wrap(tt.command)
			tt.checkFunc(t, result)
		})
	}
}

func TestNewCommandWrapper(t *testing.T) {
	shellType := ShellTypeBash
	wrapper := NewCommandWrapper(shellType)

	if wrapper == nil {
		t.Fatal("NewCommandWrapper returned nil")
	}

	if wrapper.shellType != shellType {
		t.Errorf("expected shellType %q, got %q", shellType, wrapper.shellType)
	}
}

func TestOSC133Constants(t *testing.T) {
	// 测试 OSC 133 标记常量
	if !strings.Contains(StartMarker, "\x1b]133;A") {
		t.Errorf("StartMarker should contain OSC 133;A sequence")
	}

	if !strings.Contains(EndMarkerTemplate, "\x1b]133;D;") {
		t.Errorf("EndMarkerTemplate should contain OSC 133;D sequence")
	}
}

func TestCommandWrapper_ComplexCommands(t *testing.T) {
	tests := []struct {
		name      string
		shellType ShellType
		command   string
	}{
		{"bash with pipes", ShellTypeBash, "cat file.txt | grep pattern | wc -l"},
		{"bash with quotes", ShellTypeBash, `echo "hello world"`},
		{"bash with semicolons", ShellTypeBash, "cd /tmp; ls -la"},
		{"bash with &&", ShellTypeBash, "mkdir test && cd test"},
		{"zsh with redirection", ShellTypeZsh, "echo test > /tmp/test.txt"},
		{"powershell with pipeline", ShellTypePowershell, "Get-ChildItem | Where-Object {$_.Length -gt 1MB}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := NewCommandWrapper(tt.shellType)
			result := wrapper.Wrap(tt.command)

			// 确保原始命令在包装后的结果中
			if !strings.Contains(result, tt.command) {
				t.Errorf("wrapped command should contain original command %q, got %q", tt.command, result)
			}
		})
	}
}
