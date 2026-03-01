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
	"os/exec"
	"strings"

	"log/slog"

	"github.com/creack/pty"
)

// ProcessOptions 进程创建选项
type ProcessOptions struct {
	ShellPath string
	ShellType ShellType
	WorkPath  string
	SessionID string
	Rows      uint16
	Cols      uint16
}

// Process 创建的进程信息
type Process struct {
	Pty        *os.File
	Cmd        *exec.Cmd
	ConfigPath string
	UseHooks   bool
}

// ProcessManager PTY 进程管理器
type ProcessManager struct {
	configGenerator *ShellConfigGenerator
	logger          *slog.Logger
}

// NewProcessManager 创建进程管理器
func NewProcessManager(generator *ShellConfigGenerator, logger *slog.Logger) *ProcessManager {
	return &ProcessManager{
		configGenerator: generator,
		logger:          logger,
	}
}

// CreateProcess 创建 PTY 进程
func (pm *ProcessManager) CreateProcess(opts *ProcessOptions) (*Process, error) {
	var cmd *exec.Cmd
	var configPath string
	var useHooks bool

	// 尝试为支持 hooks 的 shell 注入配置
	if opts.ShellType.SupportsHooks() {
		var err error
		configPath, err = pm.configGenerator.GenerateConfig(opts.ShellType, opts.SessionID)
		if err != nil {
			pm.logger.Warn("生成 shell 配置失败，使用命令包装模式", "error", err)
		} else {
			useHooks = true
		}
	}

	if useHooks && configPath != "" {
		// 使用 hooks 模式
		args, _ := pm.configGenerator.GetShellArgs(opts.ShellType, configPath)
		cmd = exec.Command(opts.ShellPath, args...)
		pm.logger.Info("终端使用hooks模式")
	} else {
		// 使用命令包装模式或默认模式
		cmd = exec.Command(opts.ShellPath)
		pm.logger.Info("终端使用包装模式")
	}

	if opts.WorkPath != "" {
		cmd.Dir = opts.WorkPath
	}
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"BOXIFY_SESSION_ID="+opts.SessionID,
	)

	// 对于 zsh，设置 ZDOTDIR 环境变量让 shell 从临时目录加载配置
	if useHooks && opts.ShellType == ShellTypeZsh && configPath != "" {
		cmd.Env = append(cmd.Env, "ZDOTDIR="+configPath)
	}

	// 启动 PTY
	ptyFile, err := pty.Start(cmd)
	if err != nil {
		// 清理可能生成的配置文件
		if configPath != "" {
			pm.configGenerator.Cleanup(configPath)
		}
		return nil, fmt.Errorf("创建 PTY 失败: %w", err)
	}

	// 设置终端大小
	if err := pty.Setsize(ptyFile, &pty.Winsize{Rows: opts.Rows, Cols: opts.Cols}); err != nil {
		pm.logger.Warn("设置终端大小失败", "error", err)
	}

	return &Process{
		Pty:        ptyFile,
		Cmd:        cmd,
		ConfigPath: configPath,
		UseHooks:   useHooks,
	}, nil
}

// Resize 调整终端大小
func (pm *ProcessManager) Resize(ptyFile *os.File, rows, cols uint16) error {
	err := pty.Setsize(ptyFile, &pty.Winsize{Rows: rows, Cols: cols})
	if err != nil {
		return fmt.Errorf("调整终端大小失败: %w", err)
	}
	return nil
}

// WriteInitialCommand 写入初始命令
func (pm *ProcessManager) WriteInitialCommand(ptyFile *os.File, command string) error {
	if command == "" {
		return nil
	}

	initialCmd := command
	if !strings.HasSuffix(initialCmd, "\n") {
		initialCmd += "\n"
	}

	_, err := ptyFile.WriteString(initialCmd)
	return err
}

// Cleanup 清理进程资源
func (pm *ProcessManager) Cleanup(process *Process) error {
	if process == nil {
		return nil
	}

	// 关闭 PTY
	if process.Pty != nil {
		process.Pty.Close()
	}

	// 终止进程
	if process.Cmd != nil && process.Cmd.Process != nil {
		process.Cmd.Process.Kill()
		process.Cmd.Wait()
	}

	// 清理临时配置文件
	if process.ConfigPath != "" {
		return pm.configGenerator.Cleanup(process.ConfigPath)
	}

	return nil
}
