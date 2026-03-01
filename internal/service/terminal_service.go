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
	configGenerator := terminal.NewShellConfigGenerator(deps.app.Logger)

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
	session := terminal.NewSession(context.Background(), config.ID, process.Pty, process.Cmd, validationResult.ShellType, process.UseHooks, ts.Logger())
	session.SetConfigPath(process.ConfigPath)
	session.SetWorkPath(validationResult.WorkPath)
	session.SetEmitter(ts) // 设置事件发射器
	session.SetLogger(ts.Logger())

	ts.sessionManager.Add(session)

	// 启动输出读取 goroutine
	go ts.outputHandler.StartOutputLoop(session)

	// 启动 Git 监听（Session 自管理）
	gitStatus := session.StartGitWatcher()

	ts.Logger().Info("终端会话创建",
		"sessionId", config.ID,
		"shell", validationResult.ShellPath,
		"shellType", validationResult.ShellType,
		"useHooks", process.UseHooks,
		"workPath", config.WorkPath,
		"initialCommand", config.InitialCommand)

	// 获取环境信息
	envInfo := terminal.GetEnvironmentInfo(validationResult.WorkPath)

	// 使用 Git 监听器获取的 Git 信息（更准确）
	if gitStatus != nil {
		envInfo.GitInfo = gitStatus
	}

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

// UpdateWorkPath 更新工作路径（由 shell hook 触发）
func (ts *TerminalService) UpdateWorkPath(sessionID, newPwd string) {
	session, ok := ts.sessionManager.Get(sessionID)
	if !ok {
		return
	}

	// 使用 Session 的方法更新工作路径和 Git 监听器
	session.UpdateGitWorkPath(newPwd)
}
