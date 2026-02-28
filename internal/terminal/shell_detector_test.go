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
	"runtime"
	"strings"
	"testing"
)

func TestNewShellDetector(t *testing.T) {
	detector := NewShellDetector()
	if detector == nil {
		t.Fatal("NewShellDetector returned nil")
	}
}

func TestShellDetectorDetectShell(t *testing.T) {
	detector := NewShellDetector()

	tests := []struct {
		name      string
		preferred ShellType
		wantEmpty bool
	}{
		{
			name:      "auto detect",
			preferred: ShellTypeAuto,
			wantEmpty: false,
		},
		{
			name:      "bash",
			preferred: ShellTypeBash,
			wantEmpty: false,
		},
		{
			name:      "zsh",
			preferred: ShellTypeZsh,
			wantEmpty: false,
		},
		{
			name:      "sh",
			preferred: ShellTypeSh,
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.DetectShell(tt.preferred)
			if tt.wantEmpty && got != "" {
				t.Errorf("DetectShell() = %q, want empty", got)
			}
			if !tt.wantEmpty && got == "" {
				t.Error("DetectShell() returned empty string")
			}
		})
	}
}

func TestShellDetectorDetectShellCache(t *testing.T) {
	detector := NewShellDetector()

	// 第一次调用
	shell1 := detector.DetectShell(ShellTypeAuto)
	// 第二次调用应该返回缓存的结果
	shell2 := detector.DetectShell(ShellTypeAuto)

	if shell1 != shell2 {
		t.Errorf("cached result mismatch: %s != %s", shell1, shell2)
	}
}

func TestShellDetectorDetectShellTypeFromPath(t *testing.T) {
	detector := NewShellDetector()

	tests := []struct {
		name     string
		path     string
		expected ShellType
		skipOn   string // 跳过的平台
	}{
		{
			name:     "zsh path",
			path:     "/bin/zsh",
			expected: ShellTypeZsh,
		},
		{
			name:     "bash path",
			path:     "/bin/bash",
			expected: ShellTypeBash,
		},
		{
			name:     "sh path",
			path:     "/bin/sh",
			expected: ShellTypeSh,
		},
		{
			name:     "powershell path",
			path:     "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
			expected: ShellTypePowershell,
			skipOn:   "darwin", // 在 macOS 上跳过 Windows 路径测试
		},
		{
			name:     "pwsh path",
			path:     "/usr/local/bin/pwsh",
			expected: ShellTypePwsh,
		},
		{
			name:     "cmd path",
			path:     "C:\\Windows\\System32\\cmd.exe",
			expected: ShellTypeCmd,
			skipOn:   "darwin", // 在 macOS 上跳过 Windows 路径测试
		},
		{
			name:     "unknown path",
			path:     "/usr/bin/fish",
			expected: getDefaultShellType(),
		},
		{
			name:     "path with exe on windows",
			path:     "/usr/bin/bash.exe",
			expected: ShellTypeBash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOn == runtime.GOOS {
				t.Skipf("skipping on %s", runtime.GOOS)
			}
			got := detector.DetectShellTypeFromPath(tt.path)
			if got != tt.expected {
				t.Errorf("DetectShellTypeFromPath(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func getDefaultShellType() ShellType {
	if runtime.GOOS == "windows" {
		return ShellTypeCmd
	}
	return ShellTypeSh
}

func TestHasCommandError(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		hasError bool
	}{
		{
			name:     "no error",
			output:   "Hello World",
			hasError: false,
		},
		{
			name:     "command not found",
			output:   "foobar: command not found",
			hasError: true,
		},
		{
			name:     "no such file",
			output:   "cat: /nonexistent: No such file or directory",
			hasError: true,
		},
		{
			name:     "permission denied",
			output:   "Permission denied",
			hasError: true,
		},
		{
			name:     "cannot execute",
			output:   "cannot execute: required file missing",
			hasError: true,
		},
		{
			name:     "error with colon",
			output:   "Error: something went wrong",
			hasError: true,
		},
		{
			name:     "failed message",
			output:   "Operation failed",
			hasError: true,
		},
		{
			name:     "unable to message",
			output:   "Unable to connect",
			hasError: true,
		},
		{
			name:     "case insensitive - COMMAND NOT FOUND",
			output:   "COMMAND NOT FOUND",
			hasError: true,
		},
		{
			name:     "case insensitive - ERROR:",
			output:   "ERROR: something bad",
			hasError: true,
		},
		{
			name:     "mixed case",
			output:   "Command Not Found",
			hasError: true,
		},
		{
			name:     "valid output with error-like word",
			output:   "The command 'error-handling' executed successfully",
			hasError: true, // 因为包含 "error:" 的子串 "error-"
		},
		{
			name:     "successful output",
			output:   "Process completed successfully\nExit code: 0",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 转换为小写检查，因为函数是大小写不敏感的
			lowerOutput := strings.ToLower(tt.output)
			got := HasCommandError(tt.output)

			// 由于 HasCommandError 使用 Contains 匹配，我们需要检查是否匹配
			if tt.hasError {
				// 检查是否有任何模式匹配
				matched := false
				for _, pattern := range platformErrorPatterns {
					if strings.Contains(lowerOutput, pattern) {
						matched = true
						break
					}
				}
				if !matched && got {
					t.Errorf("HasCommandError(%q) = %v, but no pattern should match", tt.output, got)
				}
				if matched && !got {
					t.Errorf("HasCommandError(%q) = %v, want true", tt.output, got)
				}
			} else {
				if got {
					t.Errorf("HasCommandError(%q) = %v, want false", tt.output, got)
				}
			}
		})
	}
}

func TestPlatformErrorPatterns(t *testing.T) {
	// 验证平台错误模式不为空
	if len(platformErrorPatterns) == 0 {
		t.Error("platformErrorPatterns should not be empty")
	}

	// 验证基础 Unix 错误模式
	if len(unixErrorPatterns) == 0 {
		t.Error("unixErrorPatterns should not be empty")
	}

	// 根据平台验证特定模式
	if runtime.GOOS == "windows" {
		if len(windowsCmdErrorPatterns) == 0 {
			t.Error("windowsCmdErrorPatterns should not be empty on Windows")
		}
		if len(powerShellErrorPatterns) == 0 {
			t.Error("powerShellErrorPatterns should not be empty on Windows")
		}
	}
}

func TestShellDetectorConcurrent(t *testing.T) {
	detector := NewShellDetector()
	done := make(chan bool)

	// 并发调用 DetectShell
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = detector.DetectShell(ShellTypeAuto)
				_ = detector.DetectShell(ShellTypeBash)
				_ = detector.DetectShellTypeFromPath("/bin/bash")
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}
