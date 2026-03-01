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

package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chenyang-zz/boxify/internal/terminal"
	"github.com/chenyang-zz/boxify/internal/types"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// TerminalService 终端服务
type TerminalService struct {
	BaseService
	sessionManager  *terminal.SessionManager
	processManager  *terminal.ProcessManager
	outputHandler   *terminal.OutputHandler
	validator       *terminal.Validator
	shellDetector   *terminal.ShellDetector
	configGenerator *terminal.ShellConfigGenerator
}

// NewTerminalService 创建终端服务
func NewTerminalService(deps *ServiceDeps) *TerminalService {
	shellDetector := terminal.NewShellDetector()
	configGenerator := terminal.NewShellConfigGenerator()

	return &TerminalService{
		BaseService:     NewBaseService(deps),
		sessionManager:  terminal.NewSessionManager(),
		shellDetector:   shellDetector,
		configGenerator: configGenerator,
		validator:       terminal.NewValidator(shellDetector),
		processManager:  terminal.NewProcessManager(configGenerator, nil), // logger 稍后设置
	}
}

// ServiceStartup 服务启动
func (ts *TerminalService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	ts.SetContext(ctx)

	// 创建输出处理器（实现 EventEmitter 接口）
	ts.outputHandler = terminal.NewOutputHandler(ts, ts.Logger())

	// 更新 processManager 的 logger
	ts.processManager = terminal.NewProcessManager(ts.configGenerator, ts.Logger())

	ts.Logger().Info("服务启动", "service", "TerminalService")
	return nil
}

// ServiceShutdown 服务关闭
func (ts *TerminalService) ServiceShutdown() error {
	ts.Logger().Info("服务开始关闭，准备释放资源", "service", "TerminalService")
	ts.sessionManager.CloseAll(ts.configGenerator)
	ts.Logger().Info("服务关闭", "service", "TerminalService")
	return nil
}

// Emit 实现 EventEmitter 接口
func (ts *TerminalService) Emit(event string, data map[string]interface{}) {
	ts.App().Event.Emit(event, data)
}

// Create 创建新的终端会话
func (ts *TerminalService) Create(config terminal.TerminalConfig) *types.TerminalCreateResult {
	// 验证基本配置
	validationResult := ts.validator.ValidateBasicConfig(config)
	if !validationResult.Valid {
		return &types.TerminalCreateResult{
			BaseResult: types.BaseResult{
				Success: false,
				Message: validationResult.Message,
			},
		}
	}

	// 验证初始命令格式（不做完整执行测试，避免重复执行）
	if config.InitialCommand != "" {
		if err := ts.validator.ValidateInitialCommandFormat(config.InitialCommand); err != nil {
			return &types.TerminalCreateResult{
				BaseResult: types.BaseResult{
					Success: false,
					Message: err.Error(),
				},
			}
		}
	}

	// 规范化终端尺寸
	rows, cols := ts.validator.NormalizeSize(config.Rows, config.Cols)

	// 创建 PTY 进程
	process, err := ts.processManager.CreateProcess(&terminal.ProcessOptions{
		ShellPath: validationResult.ShellPath,
		ShellType: validationResult.ShellType,
		WorkPath:  validationResult.WorkPath,
		SessionID: config.ID,
		Rows:      rows,
		Cols:      cols,
	})
	if err != nil {
		ts.Logger().Error("创建 PTY 失败", "shell", validationResult.ShellPath, "error", err)
		return &types.TerminalCreateResult{
			BaseResult: types.BaseResult{
				Success: false,
				Message: err.Error(),
			},
		}
	}

	// 执行初始命令
	if config.InitialCommand != "" {
		if err := ts.processManager.WriteInitialCommand(process.Pty, config.InitialCommand); err != nil {
			ts.Logger().Warn("写入初始命令失败", "sessionId", config.ID, "error", err)
		}
	}

	// 创建会话
	session := terminal.NewSession(context.Background(), config.ID, process.Pty, process.Cmd, validationResult.ShellType, process.UseHooks)
	session.SetConfigPath(process.ConfigPath)

	ts.sessionManager.Add(session)

	// 启动输出读取 goroutine
	go ts.outputHandler.StartOutputLoop(session)

	ts.Logger().Info("终端会话创建",
		"sessionId", config.ID,
		"shell", validationResult.ShellPath,
		"shellType", validationResult.ShellType,
		"useHooks", process.UseHooks,
		"workPath", config.WorkPath,
		"initialCommand", config.InitialCommand)

	// 获取环境信息
	envInfo := ts.getEnvironmentInfo(validationResult.WorkPath)

	return &types.TerminalCreateResult{
		BaseResult: types.BaseResult{
			Success: true,
			Message: "终端创建成功",
		},
		Data: envInfo,
	}
}

// Write 向终端写入用户输入
func (ts *TerminalService) Write(sessionID, data string) error {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		ts.Logger().Error("Base64 解码失败", "sessionId", sessionID, "error", err)
		return fmt.Errorf("数据解码失败: %w", err)
	}

	session, ok := ts.sessionManager.Get(sessionID)
	if !ok {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	_, err = session.Pty.Write(decoded)
	if err != nil {
		ts.Logger().Error("写入 PTY 失败", "sessionId", sessionID, "error", err)
		return fmt.Errorf("写入终端失败: %w", err)
	}

	return nil
}

// WriteCommand 写入命令并返回 block ID
// 用于追踪命令输出，实现 block 关联
func (ts *TerminalService) WriteCommand(sessionID, command string) (string, error) {
	session, ok := ts.sessionManager.Get(sessionID)
	if !ok {
		return "", fmt.Errorf("会话不存在: %s", sessionID)
	}

	// 生成新的 block ID
	blockID := uuid.New().String()

	// 设置当前 block
	session.SetCurrentBlock(blockID)

	// 根据模式决定是否包装命令
	cmd := command
	if !session.UseHooks() {
		// 非 hooks 模式：使用命令包装器添加标记
		cmd = session.Wrapper().Wrap(command)
	}

	// 确保命令以换行符结尾
	if !strings.HasSuffix(cmd, "\n") && !strings.HasSuffix(cmd, "\r") {
		cmd += "\r"
	}

	_, err := session.Pty.WriteString(cmd)
	if err != nil {
		ts.Logger().Error("写入命令失败", "sessionId", sessionID, "command", command, "error", err)
		return "", fmt.Errorf("写入命令失败: %w", err)
	}

	ts.Logger().Debug("命令已写入", "sessionId", sessionID, "blockId", blockID, "command", command, "useHooks", session.UseHooks())

	return blockID, nil
}

// Resize 调整终端大小
func (ts *TerminalService) Resize(sessionID string, rows, cols uint16) error {
	session, ok := ts.sessionManager.Get(sessionID)
	if !ok {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	err := ts.processManager.Resize(session.Pty, rows, cols)
	if err != nil {
		ts.Logger().Error("调整终端大小失败", "sessionId", sessionID, "error", err)
		return err
	}

	ts.Logger().Debug("终端大小已调整", "sessionId", sessionID, "rows", rows, "cols", cols)
	return nil
}

// Close 关闭终端会话
func (ts *TerminalService) Close(sessionID string) error {
	err := ts.sessionManager.CloseSession(sessionID, ts.configGenerator)
	if err != nil {
		return err
	}

	ts.Logger().Info("终端会话已关闭", "sessionId", sessionID)
	return nil
}

// TestConfig 测试终端配置参数是否有效
func (ts *TerminalService) TestConfig(config terminal.TerminalConfig) *types.TerminalTestConfigResult {
	result := &types.TerminalTestConfigResult{
		BaseResult: types.BaseResult{
			Success: false,
			Message: "",
		},
		Data: &types.TerminalTestConfigData{},
	}

	// 步骤1：验证基本配置
	validationResult := ts.validator.ValidateBasicConfig(config)
	if !validationResult.Valid {
		result.Message = validationResult.Message
		return result
	}

	// 记录基本配置信息到 data
	result.Data.Rows = config.Rows
	result.Data.Cols = config.Cols
	if config.WorkPath != "" {
		result.Data.WorkPath = config.WorkPath
		result.Data.WorkPathValid = true
	}

	shell := ts.shellDetector.DetectShell(config.Shell)
	result.Data.RequestedShell = string(config.Shell)
	result.Data.DetectedShell = shell
	result.Data.ShellPath = validationResult.ShellPath
	result.Data.Available = true

	// 如果没有初始命令，直接返回成功
	if config.InitialCommand != "" {
		// 步骤2：验证初始命令
		cmdResult := ts.validator.ValidateInitialCommand(validationResult.ShellPath, config)
		if !cmdResult.Success {
			result.Message = cmdResult.Error
			return result
		}

		// 记录初始命令执行结果
		result.Data.InitialCommand = config.InitialCommand
		result.Data.InitialCommandExecuted = true
		result.Data.Output = cmdResult.Output
	}

	result.Success = true
	result.Message = "终端配置验证通过"

	ts.Logger().Info("终端配置测试成功",
		"shell", shell,
		"shellPath", validationResult.ShellPath,
		"rows", config.Rows,
		"cols", config.Cols,
		"workPath", config.WorkPath,
		"initialCommand", config.InitialCommand)

	return result
}

// Test 测试方法
func (ts *TerminalService) TestExample() {
	ts.Create(terminal.TerminalConfig{
		ID:             "123",
		Shell:          terminal.ShellTypeZsh,
		Rows:           0,
		Cols:           0,
		WorkPath:       "",
		InitialCommand: "ls",
	})
}

// getEnvironmentInfo 获取终端环境信息
func (ts *TerminalService) getEnvironmentInfo(workPath string) *types.TerminalEnvironmentInfo {
	info := &types.TerminalEnvironmentInfo{
		WorkPath: ts.shortenPath(workPath),
	}

	// 获取 Python 环境信息
	info.PythonEnv = ts.getPythonEnv(workPath)

	// 获取 Git 信息
	info.GitInfo = ts.getGitInfo(workPath)

	return info
}

// shortenPath 缩短路径，将用户目录替换为 ~
func (ts *TerminalService) shortenPath(path string) string {
	if path == "" {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	// 如果路径是用户目录或其子目录，替换为 ~
	if path == homeDir {
		return "~"
	}

	// 确保路径以分隔符结尾，避免部分匹配
	homeDirWithSlash := homeDir + string(filepath.Separator)
	if strings.HasPrefix(path, homeDirWithSlash) {
		return "~" + path[len(homeDir):]
	}

	return path
}

// getPythonEnv 获取 Python 环境信息
func (ts *TerminalService) getPythonEnv(workPath string) *types.PythonEnv {
	env := &types.PythonEnv{}

	// 检查 Python 是否安装
	cmd := exec.Command("python3", "--version")
	cmd.Dir = workPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 尝试 python 命令
		cmd = exec.Command("python", "--version")
		cmd.Dir = workPath
		output, err = cmd.CombinedOutput()
		if err != nil {
			return env
		}
	}

	env.HasPython = true
	env.Version = strings.TrimSpace(string(output))

	// 检测虚拟环境（按优先级检测）
	// 1. 检测 Conda 环境
	if condaEnv := os.Getenv("CONDA_DEFAULT_ENV"); condaEnv != "" {
		env.EnvActive = true
		env.EnvType = "conda"
		env.EnvName = condaEnv
		env.EnvPath = os.Getenv("CONDA_PREFIX")
		return env
	}

	// 2. 检测 Pipenv 环境
	if os.Getenv("PIPENV_ACTIVE") != "" {
		env.EnvActive = true
		env.EnvType = "pipenv"
		env.EnvName = filepath.Base(workPath)
		env.EnvPath = os.Getenv("VIRTUAL_ENV")
		return env
	}

	// 3. 检测 Poetry 环境
	if os.Getenv("POETRY_ACTIVE") != "" {
		env.EnvActive = true
		env.EnvType = "poetry"
		env.EnvName = filepath.Base(workPath)
		env.EnvPath = os.Getenv("VIRTUAL_ENV")
		return env
	}

	// 4. 检测 venv/virtualenv 环境
	if venvPath := os.Getenv("VIRTUAL_ENV"); venvPath != "" {
		env.EnvActive = true
		env.EnvType = "venv"
		env.EnvPath = venvPath
		env.EnvName = filepath.Base(venvPath)
		return env
	}

	return env
}

// getGitInfo 获取 Git 信息
func (ts *TerminalService) getGitInfo(workPath string) *types.GitInfo {
	info := &types.GitInfo{}

	// 检查是否是 Git 仓库
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = workPath
	output, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) != "true" {
		return info
	}

	info.IsRepo = true

	// 获取当前分支
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = workPath
	output, err = cmd.Output()
	if err == nil {
		info.Branch = strings.TrimSpace(string(output))
	}

	// 获取修改统计
	cmd = exec.Command("git", "diff", "--numstat")
	cmd.Dir = workPath
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			info.ModifiedFiles++
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				// 新增行数（第一列）
				if parts[0] != "-" {
					var added int
					fmt.Sscanf(parts[0], "%d", &added)
					info.AddedLines += added
				}
				// 删除行数（第二列）
				if parts[1] != "-" {
					var deleted int
					fmt.Sscanf(parts[1], "%d", &deleted)
					info.DeletedLines += deleted
				}
			}
		}
	}

	// 获取暂存区修改统计
	cmd = exec.Command("git", "diff", "--cached", "--numstat")
	cmd.Dir = workPath
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			info.ModifiedFiles++
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				if parts[0] != "-" {
					var added int
					fmt.Sscanf(parts[0], "%d", &added)
					info.AddedLines += added
				}
				if parts[1] != "-" {
					var deleted int
					fmt.Sscanf(parts[1], "%d", &deleted)
					info.DeletedLines += deleted
				}
			}
		}
	}

	return info
}
