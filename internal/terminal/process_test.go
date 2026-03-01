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
	"testing"
)

func TestNewProcessManager(t *testing.T) {
	generator := NewShellConfigGenerator(testLogger)

	pm := NewProcessManager(generator, testLogger)

	if pm == nil {
		t.Fatal("NewProcessManager returned nil")
	}

	if pm.configGenerator == nil {
		t.Error("configGenerator should not be nil")
	}
}

func TestNewProcessManager_NilLogger(t *testing.T) {
	generator := NewShellConfigGenerator(testLogger)

	// nil logger 应该是允许的
	pm := NewProcessManager(generator, nil)

	if pm == nil {
		t.Fatal("NewProcessManager returned nil")
	}
}

func TestProcessManager_CreateProcess(t *testing.T) {
	// 获取可用的 shell
	detector := NewShellDetector()
	shellPath := detector.DetectShell(ShellTypeAuto)

	generator := NewShellConfigGenerator(testLogger)
	pm := NewProcessManager(generator, testLogger)

	tests := []struct {
		name    string
		opts    *ProcessOptions
		wantErr bool
	}{
		{
			name: "valid options with bash",
			opts: &ProcessOptions{
				ShellPath: shellPath,
				ShellType: ShellTypeBash,
				WorkPath:  "/tmp",
				SessionID: "test-1",
				Rows:      24,
				Cols:      80,
			},
			wantErr: false,
		},
		{
			name: "valid options with auto shell type",
			opts: &ProcessOptions{
				ShellPath: shellPath,
				ShellType: detector.DetectShellTypeFromPath(shellPath),
				WorkPath:  "",
				SessionID: "test-2",
				Rows:      0,
				Cols:      0,
			},
			wantErr: false,
		},
		{
			name: "invalid shell path",
			opts: &ProcessOptions{
				ShellPath: "/nonexistent/shell",
				ShellType: ShellTypeBash,
				SessionID: "test-3",
				Rows:      24,
				Cols:      80,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			process, err := pm.CreateProcess(tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProcess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if process == nil {
					t.Fatal("expected non-nil process")
				}

				// 验证进程字段
				if process.Pty == nil {
					t.Error("Pty should not be nil")
				}
				if process.Cmd == nil {
					t.Error("Cmd should not be nil")
				}

				// 清理
				pm.Cleanup(process)
			}
		})
	}
}

func TestProcessManager_CreateProcess_HooksMode(t *testing.T) {
	// 获取 zsh 路径
	detector := NewShellDetector()
	shellPath := detector.DetectShell(ShellTypeZsh)

	// 如果没有 zsh，跳过测试
	if shellPath == "" || shellPath == "/bin/sh" {
		t.Skip("zsh not available")
	}

	generator := NewShellConfigGenerator(testLogger)
	pm := NewProcessManager(generator, testLogger)

	opts := &ProcessOptions{
		ShellPath: shellPath,
		ShellType: ShellTypeZsh,
		WorkPath:  "/tmp",
		SessionID: "test-hooks",
		Rows:      24,
		Cols:      80,
	}

	process, err := pm.CreateProcess(opts)
	if err != nil {
		t.Fatalf("CreateProcess() error = %v", err)
	}
	defer pm.Cleanup(process)

	// zsh 支持 hooks，应该尝试使用 hooks 模式
	if process.UseHooks {
		if process.ConfigPath == "" {
			t.Error("ConfigPath should not be empty when using hooks")
		}
	}

	// 检查日志
	// foundHooksLog := false
	// for _, msg := range testLogger.() {
	// 	if strings.Contains(msg, "hooks") {
	// 		foundHooksLog = true
	// 		break
	// 	}
	// }
	// if !foundHooksLog {
	// 	t.Log("Warning: expected hooks mode log message")
	// }
}

func TestProcessManager_Resize(t *testing.T) {
	generator := NewShellConfigGenerator(testLogger)
	pm := NewProcessManager(generator, testLogger)

	// 创建一个进程用于测试
	detector := NewShellDetector()
	shellPath := detector.DetectShell(ShellTypeAuto)

	process, err := pm.CreateProcess(&ProcessOptions{
		ShellPath: shellPath,
		ShellType: ShellTypeBash,
		SessionID: "test-resize",
		Rows:      24,
		Cols:      80,
	})
	if err != nil {
		t.Fatalf("CreateProcess() error = %v", err)
	}
	defer pm.Cleanup(process)

	// 测试调整大小
	err = pm.Resize(process.Pty, 40, 120)
	if err != nil {
		t.Errorf("Resize() error = %v", err)
	}

	// 测试极端值
	err = pm.Resize(process.Pty, 1, 1)
	if err != nil {
		t.Errorf("Resize() with min values error = %v", err)
	}

	err = pm.Resize(process.Pty, MaxRows, MaxCols)
	if err != nil {
		t.Errorf("Resize() with max values error = %v", err)
	}
}

func TestProcessManager_WriteInitialCommand(t *testing.T) {
	generator := NewShellConfigGenerator(testLogger)
	pm := NewProcessManager(generator, testLogger)

	// 创建一个进程
	detector := NewShellDetector()
	shellPath := detector.DetectShell(ShellTypeAuto)

	process, err := pm.CreateProcess(&ProcessOptions{
		ShellPath: shellPath,
		ShellType: ShellTypeBash,
		SessionID: "test-write",
		Rows:      24,
		Cols:      80,
	})
	if err != nil {
		t.Fatalf("CreateProcess() error = %v", err)
	}
	defer pm.Cleanup(process)

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "simple command",
			command: "echo hello",
			wantErr: false,
		},
		{
			name:    "empty command",
			command: "",
			wantErr: false, // 空命令应该直接返回 nil
		},
		{
			name:    "command with newline",
			command: "echo hello\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.WriteInitialCommand(process.Pty, tt.command)

			if (err != nil) != tt.wantErr {
				t.Errorf("WriteInitialCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessManager_Cleanup(t *testing.T) {
	generator := NewShellConfigGenerator(testLogger)
	pm := NewProcessManager(generator, testLogger)

	t.Run("nil process", func(t *testing.T) {
		err := pm.Cleanup(nil)
		if err != nil {
			t.Errorf("Cleanup(nil) should return nil, got %v", err)
		}
	})

	t.Run("process without config path", func(t *testing.T) {
		detector := NewShellDetector()
		shellPath := detector.DetectShell(ShellTypeAuto)

		process, err := pm.CreateProcess(&ProcessOptions{
			ShellPath: shellPath,
			ShellType: ShellTypeBash,
			SessionID: "test-cleanup",
			Rows:      24,
			Cols:      80,
		})
		if err != nil {
			t.Fatalf("CreateProcess() error = %v", err)
		}

		err = pm.Cleanup(process)
		if err != nil {
			t.Errorf("Cleanup() error = %v", err)
		}
	})

	t.Run("process with config path", func(t *testing.T) {
		detector := NewShellDetector()
		shellPath := detector.DetectShell(ShellTypeZsh)

		if shellPath == "" || shellPath == "/bin/sh" {
			t.Skip("zsh not available")
		}

		process, err := pm.CreateProcess(&ProcessOptions{
			ShellPath: shellPath,
			ShellType: ShellTypeZsh,
			SessionID: "test-cleanup-config",
			Rows:      24,
			Cols:      80,
		})
		if err != nil {
			t.Fatalf("CreateProcess() error = %v", err)
		}

		// 如果使用了 hooks，ConfigPath 应该被清理
		configPath := process.ConfigPath

		err = pm.Cleanup(process)
		if err != nil {
			t.Errorf("Cleanup() error = %v", err)
		}

		// 验证配置文件被删除（如果存在）
		if configPath != "" {
			if _, err := os.Stat(configPath); !os.IsNotExist(err) {
				t.Logf("Warning: config path %s might not be cleaned up", configPath)
			}
		}
	})
}

func TestProcessOptions(t *testing.T) {
	opts := &ProcessOptions{
		ShellPath: "/bin/bash",
		ShellType: ShellTypeBash,
		WorkPath:  "/tmp",
		SessionID: "test-123",
		Rows:      24,
		Cols:      80,
	}

	if opts.ShellPath != "/bin/bash" {
		t.Errorf("ShellPath = %s, want /bin/bash", opts.ShellPath)
	}
	if opts.ShellType != ShellTypeBash {
		t.Errorf("ShellType = %s, want %s", opts.ShellType, ShellTypeBash)
	}
	if opts.WorkPath != "/tmp" {
		t.Errorf("WorkPath = %s, want /tmp", opts.WorkPath)
	}
	if opts.SessionID != "test-123" {
		t.Errorf("SessionID = %s, want test-123", opts.SessionID)
	}
	if opts.Rows != 24 {
		t.Errorf("Rows = %d, want 24", opts.Rows)
	}
	if opts.Cols != 80 {
		t.Errorf("Cols = %d, want 80", opts.Cols)
	}
}

func TestProcess(t *testing.T) {
	process := &Process{
		Pty:        nil,
		Cmd:        &exec.Cmd{},
		ConfigPath: "/tmp/config",
		UseHooks:   true,
	}

	if process.Cmd == nil {
		t.Error("Cmd should not be nil")
	}
	if process.ConfigPath != "/tmp/config" {
		t.Errorf("ConfigPath = %s, want /tmp/config", process.ConfigPath)
	}
	if !process.UseHooks {
		t.Error("UseHooks should be true")
	}
}
