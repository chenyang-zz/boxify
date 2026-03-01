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
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
)

// 配置常量
const (
	MaxRows           uint16 = 300
	MaxCols           uint16 = 500
	DefaultRows       uint16 = 24
	DefaultCols       uint16 = 80
	MaxCommandLength  int    = 10000
	CommandTimeout           = 5 * time.Second
)

// ValidationResult 基本配置验证结果
type ValidationResult struct {
	Valid     bool
	Message   string
	ShellPath string
	ShellType ShellType
	WorkPath  string
}

// CommandTestResult 命令测试结果
type CommandTestResult struct {
	Success bool
	Output  string
	Error   string
}

// Validator 终端配置验证器
type Validator struct {
	shellDetector *ShellDetector
}

// NewValidator 创建验证器
func NewValidator(detector *ShellDetector) *Validator {
	return &Validator{
		shellDetector: detector,
	}
}

// ValidateBasicConfig 验证基本配置（不包含初始命令执行）
func (v *Validator) ValidateBasicConfig(config TerminalConfig) *ValidationResult {
	// 验证终端尺寸
	if config.Rows > MaxRows {
		return &ValidationResult{
			Valid:   false,
			Message: fmt.Sprintf("终端行数超出范围，支持 0-%d 行（0 表示使用默认值）", MaxRows),
		}
	}
	if config.Cols > MaxCols {
		return &ValidationResult{
			Valid:   false,
			Message: fmt.Sprintf("终端列数超出范围，支持 0-%d 列（0 表示使用默认值）", MaxCols),
		}
	}

	// 验证工作路径
	workPath := v.GetWorkPath(config)
	if config.WorkPath != "" {
		info, err := os.Stat(config.WorkPath)
		if err != nil {
			return &ValidationResult{
				Valid:   false,
				Message: fmt.Sprintf("工作路径不存在或无法访问: %s (%v)", config.WorkPath, err),
			}
		}
		if !info.IsDir() {
			return &ValidationResult{
				Valid:   false,
				Message: fmt.Sprintf("工作路径不是目录: %s", config.WorkPath),
			}
		}
	}

	// 检测并验证 shell
	shellPath := v.shellDetector.DetectShell(config.Shell)

	// 检查 shell 是否存在
	if _, err := exec.LookPath(shellPath); err != nil {
		return &ValidationResult{
			Valid:   false,
			Message: fmt.Sprintf("找不到 shell: %s (%v)", shellPath, err),
		}
	}

	// 检测 shell 类型
	shellType := config.Shell
	if shellType == ShellTypeAuto {
		shellType = v.shellDetector.DetectShellTypeFromPath(shellPath)
	}

	return &ValidationResult{
		Valid:     true,
		ShellPath: shellPath,
		ShellType: shellType,
		WorkPath:  workPath,
	}
}

// ValidateInitialCommandFormat 验证初始命令格式（不执行）
func (v *Validator) ValidateInitialCommandFormat(command string) error {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return fmt.Errorf("初始命令不能为空或仅包含空白字符")
	}
	if len(command) > MaxCommandLength {
		return fmt.Errorf("初始命令过长，最大支持 %d 字符", MaxCommandLength)
	}
	return nil
}

// ValidateInitialCommand 验证初始命令是否能执行
func (v *Validator) ValidateInitialCommand(shellPath string, config TerminalConfig) *CommandTestResult {
	// 检查命令格式
	if err := v.ValidateInitialCommandFormat(config.InitialCommand); err != nil {
		return &CommandTestResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	// 使用辅助方法创建命令（验证时不需要 hooks，使用简单模式）
	workPath := v.GetWorkPath(config)
	testCmd := exec.Command(shellPath)
	if workPath != "" {
		testCmd.Dir = workPath
	}
	testCmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	)

	// 创建 PTY 测试
	ptyFile, err := pty.Start(testCmd)
	if err != nil {
		return &CommandTestResult{
			Success: false,
			Error:   fmt.Sprintf("PTY 创建测试失败: %v", err),
		}
	}

	// 写入初始命令
	initialCmd := config.InitialCommand
	if !strings.HasSuffix(initialCmd, "\n") {
		initialCmd += "\n"
	}
	if _, err := ptyFile.WriteString(initialCmd); err != nil {
		// 记录警告但不中断
	}

	// 写入 exit 命令让 shell 在执行完初始命令后退出
	if _, err := ptyFile.WriteString("exit\n"); err != nil {
		// 记录警告但不中断
	}

	// 使用 context 实现超时控制
	ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
	defer cancel()

	// 使用 done channel 控制读取 goroutine 退出
	readDone := make(chan struct{})

	// 执行结果
	type execResult struct {
		output string
		err    error
	}
	resultCh := make(chan execResult, 1)

	// 输出读取 goroutine
	outputBuf := make([]byte, 4096)
	var outputBuilder strings.Builder
	go func() {
		defer close(readDone)
		for {
			n, err := ptyFile.Read(outputBuf)
			if n > 0 {
				outputBuilder.Write(outputBuf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	// 进程等待 goroutine
	go func() {
		err := testCmd.Wait()
		// 等待读取 goroutine 完成
		<-readDone

		resultCh <- execResult{
			output: outputBuilder.String(),
			err:    err,
		}
	}()

	// 等待执行完成或超时
	select {
	case res := <-resultCh:
		ptyFile.Close()

		// 检查执行结果
		if res.err != nil {
			return &CommandTestResult{
				Success: false,
				Error:   fmt.Sprintf("初始命令执行失败: %v", res.err),
				Output:  res.output,
			}
		}

		// 检查输出中的错误标志
		if HasCommandError(res.output) {
			return &CommandTestResult{
				Success: false,
				Error:   "初始命令执行出现错误，请检查命令是否正确",
				Output:  res.output,
			}
		}

		return &CommandTestResult{
			Success: true,
			Output:  res.output,
		}

	case <-ctx.Done():
		// 超时强制终止
		testCmd.Process.Kill()
		ptyFile.Close()
		// 等待读取 goroutine 退出
		<-readDone
		return &CommandTestResult{
			Success: false,
			Error:   fmt.Sprintf("初始命令执行超时（%v）", CommandTimeout),
		}
	}
}

// GetWorkPath 获取工作路径，如果配置中未指定则使用用户主目录
func (v *Validator) GetWorkPath(config TerminalConfig) string {
	if config.WorkPath != "" {
		return config.WorkPath
	}
	if homeDir, err := os.UserHomeDir(); err == nil {
		return homeDir
	}
	return ""
}

// NormalizeSize 规范化终端尺寸，返回 (rows, cols)
func (v *Validator) NormalizeSize(rows, cols uint16) (uint16, uint16) {
	if rows == 0 {
		rows = DefaultRows
	}
	if cols == 0 {
		cols = DefaultCols
	}
	return rows, cols
}
