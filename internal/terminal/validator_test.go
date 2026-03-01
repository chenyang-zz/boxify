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
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	detector := NewShellDetector()
	validator := NewValidator(detector)

	if validator == nil {
		t.Fatal("NewValidator returned nil")
	}

	if validator.shellDetector == nil {
		t.Error("shellDetector should not be nil")
	}
}

func TestValidator_ValidateBasicConfig(t *testing.T) {
	detector := NewShellDetector()
	validator := NewValidator(detector)

	tests := []struct {
		name        string
		config      TerminalConfig
		wantValid   bool
		wantMsgPart string
	}{
		{
			name: "valid minimal config",
			config: TerminalConfig{
				ID:    "test-1",
				Shell: ShellTypeAuto,
			},
			wantValid: true,
		},
		{
			name: "valid full config",
			config: TerminalConfig{
				ID:       "test-2",
				Shell:    ShellTypeAuto,
				Rows:     24,
				Cols:     80,
				WorkPath: "/tmp",
			},
			wantValid: true,
		},
		{
			name: "rows exceeds max",
			config: TerminalConfig{
				ID:    "test-3",
				Shell: ShellTypeAuto,
				Rows:  500,
			},
			wantValid:   false,
			wantMsgPart: "行数超出范围",
		},
		{
			name: "cols exceeds max",
			config: TerminalConfig{
				ID:    "test-4",
				Shell: ShellTypeAuto,
				Cols:  600,
			},
			wantValid:   false,
			wantMsgPart: "列数超出范围",
		},
		{
			name: "invalid work path",
			config: TerminalConfig{
				ID:       "test-5",
				Shell:    ShellTypeAuto,
				WorkPath: "/nonexistent/path/that/does/not/exist",
			},
			wantValid:   false,
			wantMsgPart: "工作路径不存在",
		},
		{
			name: "work path is file not directory",
			config: TerminalConfig{
				ID:       "test-6",
				Shell:    ShellTypeAuto,
				WorkPath: "/etc/passwd", // 通常存在且是文件
			},
			wantValid:   false,
			wantMsgPart: "工作路径不是目录",
		},
		{
			name: "zero rows and cols should be valid",
			config: TerminalConfig{
				ID:    "test-7",
				Shell: ShellTypeAuto,
				Rows:  0,
				Cols:  0,
			},
			wantValid: true,
		},
		{
			name: "max rows and cols should be valid",
			config: TerminalConfig{
				ID:    "test-8",
				Shell: ShellTypeAuto,
				Rows:  MaxRows,
				Cols:  MaxCols,
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateBasicConfig(tt.config)

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateBasicConfig() valid = %v, want %v, message: %s", result.Valid, tt.wantValid, result.Message)
			}

			if tt.wantMsgPart != "" && !strings.Contains(result.Message, tt.wantMsgPart) {
				t.Errorf("ValidateBasicConfig() message = %s, want to contain %s", result.Message, tt.wantMsgPart)
			}

			// 验证成功时检查返回值
			if result.Valid {
				if result.ShellPath == "" {
					t.Error("ShellPath should not be empty for valid config")
				}
				if result.ShellType == ShellTypeAuto {
					t.Error("ShellType should be resolved, not Auto")
				}
				if result.WorkPath == "" {
					t.Error("WorkPath should not be empty for valid config")
				}
			}
		})
	}
}

func TestValidator_ValidateInitialCommandFormat(t *testing.T) {
	validator := NewValidator(NewShellDetector())

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "valid command",
			command: "ls -la",
			wantErr: false,
		},
		{
			name:    "empty command",
			command: "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			command: "   \t\n  ",
			wantErr: true,
		},
		{
			name:    "command too long",
			command: strings.Repeat("a", MaxCommandLength+1),
			wantErr: true,
		},
		{
			name:    "command at max length",
			command: strings.Repeat("a", MaxCommandLength),
			wantErr: false,
		},
		{
			name:    "multiline command",
			command: "echo hello\necho world",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateInitialCommandFormat(tt.command)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInitialCommandFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_GetWorkPath(t *testing.T) {
	validator := NewValidator(NewShellDetector())

	tests := []struct {
		name     string
		config   TerminalConfig
		wantPath string
	}{
		{
			name: "specified work path",
			config: TerminalConfig{
				WorkPath: "/tmp",
			},
			wantPath: "/tmp",
		},
		{
			name:     "empty work path uses home",
			config:   TerminalConfig{},
			wantPath: "", // 无法确定具体值，只检查不为空
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.GetWorkPath(tt.config)

			if tt.wantPath != "" {
				if got != tt.wantPath {
					t.Errorf("GetWorkPath() = %v, want %v", got, tt.wantPath)
				}
			} else {
				// 对于空配置，应该返回用户主目录
				homeDir, err := os.UserHomeDir()
				if err == nil && got != homeDir {
					t.Errorf("GetWorkPath() = %v, want %v", got, homeDir)
				}
			}
		})
	}
}

func TestValidator_NormalizeSize(t *testing.T) {
	validator := NewValidator(NewShellDetector())

	tests := []struct {
		name    string
		rows    uint16
		cols    uint16
		wantRow uint16
		wantCol uint16
	}{
		{
			name:    "zero values use defaults",
			rows:    0,
			cols:    0,
			wantRow: DefaultRows,
			wantCol: DefaultCols,
		},
		{
			name:    "non-zero values preserved",
			rows:    40,
			cols:    120,
			wantRow: 40,
			wantCol: 120,
		},
		{
			name:    "mixed zero and non-zero",
			rows:    0,
			cols:    100,
			wantRow: DefaultRows,
			wantCol: 100,
		},
		{
			name:    "max values preserved",
			rows:    MaxRows,
			cols:    MaxCols,
			wantRow: MaxRows,
			wantCol: MaxCols,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRows, gotCols := validator.NormalizeSize(tt.rows, tt.cols)

			if gotRows != tt.wantRow {
				t.Errorf("NormalizeSize() rows = %v, want %v", gotRows, tt.wantRow)
			}
			if gotCols != tt.wantCol {
				t.Errorf("NormalizeSize() cols = %v, want %v", gotCols, tt.wantCol)
			}
		})
	}
}

func TestValidator_ValidateInitialCommand(t *testing.T) {
	validator := NewValidator(NewShellDetector())

	// 先获取一个有效的 shell 路径
	result := validator.ValidateBasicConfig(TerminalConfig{Shell: ShellTypeAuto})
	if !result.Valid {
		t.Skip("无法获取有效的 shell 路径")
	}

	tests := []struct {
		name      string
		config    TerminalConfig
		wantError bool
	}{
		{
			name: "simple echo command",
			config: TerminalConfig{
				InitialCommand: "echo hello",
			},
			wantError: false,
		},
		{
			name: "empty command should fail",
			config: TerminalConfig{
				InitialCommand: "",
			},
			wantError: true,
		},
		{
			name: "invalid command",
			config: TerminalConfig{
				InitialCommand: "nonexistent_command_xyz123",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdResult := validator.ValidateInitialCommand(result.ShellPath, tt.config)

			if (cmdResult.Error != "") != tt.wantError {
				t.Errorf("ValidateInitialCommand() error = %v, wantError %v", cmdResult.Error, tt.wantError)
			}

			// 成功时应该有输出
			if !tt.wantError && cmdResult.Output == "" {
				t.Log("Warning: expected some output for successful command")
			}
		})
	}
}

func TestConstants(t *testing.T) {
	// 验证常量值合理
	if MaxRows <= DefaultRows {
		t.Errorf("MaxRows (%d) should be greater than DefaultRows (%d)", MaxRows, DefaultRows)
	}
	if MaxCols <= DefaultCols {
		t.Errorf("MaxCols (%d) should be greater than DefaultCols (%d)", MaxCols, DefaultCols)
	}
	if MaxCommandLength <= 0 {
		t.Errorf("MaxCommandLength (%d) should be positive", MaxCommandLength)
	}
	if CommandTimeout <= 0 {
		t.Errorf("CommandTimeout (%v) should be positive", CommandTimeout)
	}
}
