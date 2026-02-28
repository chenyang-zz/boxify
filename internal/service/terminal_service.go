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
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/chenyang-zz/boxify/internal/terminal"
	"github.com/creack/pty"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// TerminalService 终端服务
type TerminalService struct {
	BaseService
	sessions        map[string]*terminal.Session
	mu              sync.RWMutex
	shellDetector   *terminal.ShellDetector
	configGenerator *terminal.ShellConfigGenerator
}

// NewTerminalService 创建终端服务
func NewTerminalService(deps *ServiceDeps) *TerminalService {
	return &TerminalService{
		BaseService:     NewBaseService(deps),
		sessions:        make(map[string]*terminal.Session),
		shellDetector:   terminal.NewShellDetector(),
		configGenerator: terminal.NewShellConfigGenerator(),
	}
}

// Startup 服务启动
func (ts *TerminalService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	ts.SetContext(ctx)
	ts.Logger().Info("服务启动", "service", "TerminalService")
	return nil
}

// Shutdown 服务关闭
func (ts *TerminalService) ServiceShutdown() error {
	ts.Logger().Info("服务开始关闭，准备释放资源", "service", "TerminalService")

	ts.mu.Lock()
	defer ts.mu.Unlock()

	// 关闭所有终端会话
	for _, session := range ts.sessions {
		ts.closeSessionUnsafe(session)
	}

	ts.Logger().Info("服务关闭", "service", "TerminalService")
	return nil
}

// validateBasicConfig 验证基本配置（不包含初始命令）
// 返回: (是否有效, 错误信息, 检测到的 shell 路径)
func (ts *TerminalService) validateBasicConfig(config terminal.TerminalConfig) (bool, string, string) {
	// 验证终端尺寸（允许 0，表示使用默认值）
	// 注意：Rows 和 Cols 是 uint16 类型，不会为负数
	if config.Rows > 300 {
		return false, "终端行数超出范围，支持 0-300 行（0 表示使用默认值）", ""
	}
	if config.Cols > 500 {
		return false, "终端列数超出范围，支持 0-500 列（0 表示使用默认值）", ""
	}

	// 验证工作路径
	if config.WorkPath != "" {
		info, err := os.Stat(config.WorkPath)
		if err != nil {
			return false, fmt.Sprintf("工作路径不存在或无法访问: %s (%v)", config.WorkPath, err), ""
		}
		if !info.IsDir() {
			return false, fmt.Sprintf("工作路径不是目录: %s", config.WorkPath), ""
		}
	}

	// 检测并验证 shell
	shellPath := ts.shellDetector.DetectShell(config.Shell)

	// 检查 shell 是否存在
	if _, err := exec.LookPath(shellPath); err != nil {
		return false, fmt.Sprintf("找不到 shell: %s (%v)", shellPath, err), ""
	}

	return true, "", shellPath
}

// getWorkPath 获取工作路径，如果配置中未指定则使用用户主目录
func (ts *TerminalService) getWorkPath(config terminal.TerminalConfig) string {
	if config.WorkPath != "" {
		return config.WorkPath
	}
	if homeDir, err := os.UserHomeDir(); err == nil {
		return homeDir
	}
	return ""
}

// createShellCommand 创建 shell 命令并设置环境变量
func (ts *TerminalService) createShellCommand(shellPath, workPath string, shellType terminal.ShellType, sessionID string) (*exec.Cmd, string, bool) {
	var cmd *exec.Cmd
	var configPath string
	var useHooks bool

	// 尝试为支持 hooks 的 shell 注入配置
	if shellType.SupportsHooks() {
		var err error
		configPath, err = ts.configGenerator.GenerateConfig(shellType, sessionID)
		if err != nil {
			ts.Logger().Warn("生成 shell 配置失败，使用命令包装模式", "error", err)
		} else {
			useHooks = true
		}
	}

	if useHooks && configPath != "" {
		// 使用 hooks 模式
		args, _ := ts.configGenerator.GetShellArgs(shellType, configPath)
		cmd = exec.Command(shellPath, args...)
		ts.Logger().Info("终端使用hooks模式")
	} else {
		// 使用命令包装模式或默认模式
		cmd = exec.Command(shellPath)
		ts.Logger().Info("终端使用包装模式")
	}

	if workPath != "" {
		cmd.Dir = workPath
	}
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"BOXIFY_SESSION_ID="+sessionID,
	)

	// 对于 zsh，设置 ZDOTDIR 环境变量让 shell 从临时目录加载配置
	if useHooks && shellType == terminal.ShellTypeZsh && configPath != "" {
		cmd.Env = append(cmd.Env, "ZDOTDIR="+configPath)
	}

	return cmd, configPath, useHooks
}

// validateInitialCommand 验证初始命令是否能执行
// 返回: (是否有效, 错误信息, 输出内容)
func (ts *TerminalService) validateInitialCommand(shellPath string, config terminal.TerminalConfig) (bool, string, string) {
	// 检查命令是否为空或仅包含空白字符
	trimmed := strings.TrimSpace(config.InitialCommand)
	if trimmed == "" {
		return false, "初始命令不能为空或仅包含空白字符", ""
	}

	// 检查命令长度限制
	if len(config.InitialCommand) > 10000 {
		return false, "初始命令过长，最大支持 10000 字符", ""
	}

	// 使用辅助方法创建命令（验证时不需要 hooks，使用简单模式）
	workPath := ts.getWorkPath(config)
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
		return false, fmt.Sprintf("PTY 创建测试失败: %v", err), ""
	}

	// 写入初始命令
	initialCmd := config.InitialCommand
	if !strings.HasSuffix(initialCmd, "\n") {
		initialCmd += "\n"
	}
	if _, err := ptyFile.WriteString(initialCmd); err != nil {
		ts.Logger().Warn("写入初始命令失败", "error", err)
	}

	// 写入 exit 命令让 shell 在执行完初始命令后退出
	if _, err := ptyFile.WriteString("exit\n"); err != nil {
		ts.Logger().Warn("写入 exit 命令失败", "error", err)
	}

	// 使用 context 实现超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
			return false, fmt.Sprintf("初始命令执行失败: %v", res.err), res.output
		}

		// 检查输出中的错误标志
		if terminal.HasCommandError(res.output) {
			return false, "初始命令执行出现错误，请检查命令是否正确", res.output
		}

		return true, "", res.output

	case <-ctx.Done():
		// 超时强制终止
		testCmd.Process.Kill()
		ptyFile.Close()
		// 等待读取 goroutine 退出
		<-readDone
		return false, "初始命令执行超时（5秒）", ""
	}
}

// Create 创建新的终端会话
func (ts *TerminalService) Create(config terminal.TerminalConfig) *connection.QueryResult {
	// 验证基本配置
	valid, errMsg, shellPath := ts.validateBasicConfig(config)
	if !valid {
		return &connection.QueryResult{
			Success: false,
			Message: errMsg,
		}
	}

	// 验证初始命令格式（不做完整执行测试，避免重复执行）
	if config.InitialCommand != "" {
		trimmed := strings.TrimSpace(config.InitialCommand)
		if trimmed == "" {
			return &connection.QueryResult{
				Success: false,
				Message: "初始命令不能为空或仅包含空白字符",
			}
		}
		if len(config.InitialCommand) > 10000 {
			return &connection.QueryResult{
				Success: false,
				Message: "初始命令过长，最大支持 10000 字符",
			}
		}
	}

	// 设置默认终端尺寸
	rows := config.Rows
	cols := config.Cols
	if rows == 0 {
		rows = 24
	}
	if cols == 0 {
		cols = 80
	}

	// 检测 shell 类型
	shellType := config.Shell
	if shellType == terminal.ShellTypeAuto {
		shellType = ts.shellDetector.DetectShellTypeFromPath(shellPath)
	}

	// 使用辅助方法创建命令（可能注入 hooks 配置）
	workPath := ts.getWorkPath(config)
	cmd, configPath, useHooks := ts.createShellCommand(shellPath, workPath, shellType, config.ID)

	// 启动 PTY
	ptyFile, err := pty.Start(cmd)
	if err != nil {
		ts.Logger().Error("创建 PTY 失败", "shell", shellPath, "error", err)
		// 清理可能生成的配置文件
		if configPath != "" {
			ts.configGenerator.Cleanup(configPath)
		}
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("创建终端失败: %v", err),
		}
	}

	// 设置终端大小
	if err := pty.Setsize(ptyFile, &pty.Winsize{Rows: rows, Cols: cols}); err != nil {
		ts.Logger().Warn("设置终端大小失败", "error", err)
	}

	// 执行初始命令
	if config.InitialCommand != "" {
		initialCmd := config.InitialCommand
		if !strings.HasSuffix(initialCmd, "\n") {
			initialCmd += "\n"
		}
		if _, err := ptyFile.WriteString(initialCmd); err != nil {
			ts.Logger().Warn("写入初始命令失败", "sessionId", config.ID, "error", err)
		}
	}

	// 创建会话
	session := terminal.NewSession(context.Background(), config.ID, ptyFile, cmd, shellType, useHooks)
	session.SetConfigPath(configPath)

	ts.mu.Lock()
	ts.sessions[config.ID] = session
	ts.mu.Unlock()

	// 启动输出读取 goroutine
	go ts.readOutputLoop(session)

	ts.Logger().Info("终端会话创建", "sessionId", config.ID, "shell", shellPath, "shellType", shellType, "useHooks", useHooks, "workPath", config.WorkPath, "initialCommand", config.InitialCommand)

	return &connection.QueryResult{
		Success: true,
		Message: "终端创建成功",
	}
}

// readOutputLoop 读取 PTY 输出并发送到前端
func (ts *TerminalService) readOutputLoop(session *terminal.Session) {
	buf := make([]byte, 1024)

	for {
		select {
		case <-session.Context().Done():
			// 收到退出信号
			return
		default:
			n, err := session.Pty.Read(buf)
			if err != nil {
				if err != io.EOF && session.Context().Err() == nil {
					// 只有在 context 未取消时才报告错误
					ts.Logger().Error("读取 PTY 输出失败", "sessionId", session.ID, "error", err)
					ts.App().Event.Emit("terminal:error", map[string]interface{}{
						"sessionId": session.ID,
						"message":   err.Error(),
					})
				}
				return
			}

			// 使用过滤器处理输出
			result := session.Filter().Process(buf[:n])

			// 获取当前 block ID
			blockID := session.CurrentBlock()

			// 只有有过滤后输出时才发送
			if len(result.Output) > 0 {
				ts.Logger().Info("提取过滤后终端输出", "text", string(result.Output))
				encoded := base64.StdEncoding.EncodeToString(result.Output)
				ts.App().Event.Emit("terminal:output", map[string]interface{}{
					"sessionId": session.ID,
					"blockId":   blockID,
					"data":      encoded,
				})
			}

			// 命令结束时发送事件
			if result.CommandEnded {
				ts.App().Event.Emit("terminal:command_end", map[string]interface{}{
					"sessionId": session.ID,
					"blockId":   blockID,
					"exitCode":  result.ExitCode,
				})
			}
		}
	}
}

// Write 向终端写入用户输入
func (ts *TerminalService) Write(sessionID, data string) error {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		ts.Logger().Error("Base64 解码失败", "sessionId", sessionID, "error", err)
		return fmt.Errorf("数据解码失败: %w", err)
	}

	ts.mu.RLock()
	defer ts.mu.RUnlock()

	session, ok := ts.sessions[sessionID]
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
	ts.mu.RLock()
	session, ok := ts.sessions[sessionID]
	ts.mu.RUnlock()

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
	ts.mu.RLock()
	session, ok := ts.sessions[sessionID]
	ts.mu.RUnlock()

	if !ok {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	err := pty.Setsize(session.Pty, &pty.Winsize{Rows: rows, Cols: cols})
	if err != nil {
		ts.Logger().Error("调整终端大小失败", "sessionId", sessionID, "error", err)
		return fmt.Errorf("调整终端大小失败: %w", err)
	}

	ts.Logger().Debug("终端大小已调整", "sessionId", sessionID, "rows", rows, "cols", cols)
	return nil
}

// Close 关闭终端会话
func (ts *TerminalService) Close(sessionID string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	session, ok := ts.sessions[sessionID]
	if !ok {
		return nil
	}

	ts.closeSessionUnsafe(session)
	delete(ts.sessions, sessionID)

	ts.Logger().Info("终端会话已关闭", "sessionId", sessionID)
	return nil
}

// closeSessionUnsafe 内部方法：关闭会话（不加锁）
func (ts *TerminalService) closeSessionUnsafe(session *terminal.Session) {
	// 关闭会话资源
	session.Close()

	// 终止进程并等待
	if err := session.KillProcess(); err != nil {
		ts.Logger().Debug("终止进程失败", "sessionId", session.ID, "error", err)
	}
	if err := session.WaitProcess(); err != nil {
		ts.Logger().Debug("进程等待结束", "sessionId", session.ID, "error", err)
	}

	// 清理临时配置文件
	if session.ConfigPath() != "" {
		if err := ts.configGenerator.Cleanup(session.ConfigPath()); err != nil {
			ts.Logger().Warn("清理临时配置文件失败", "sessionId", session.ID, "error", err)
		}
	}
}

// TestConfig 测试终端配置参数是否有效
func (ts *TerminalService) TestConfig(config terminal.TerminalConfig) *connection.QueryResult {
	result := &connection.QueryResult{
		Success: false,
		Message: "",
		Data:    make(map[string]interface{}),
	}

	data := result.Data.(map[string]interface{})

	// 步骤1：验证基本配置
	valid, errMsg, shellPath := ts.validateBasicConfig(config)
	if !valid {
		result.Message = errMsg
		return result
	}

	// 记录基本配置信息到 data
	data["rows"] = config.Rows
	data["cols"] = config.Cols
	if config.WorkPath != "" {
		data["workPath"] = config.WorkPath
		data["workPathValid"] = true
	}

	shell := ts.shellDetector.DetectShell(config.Shell)
	data["requestedShell"] = string(config.Shell)
	data["detectedShell"] = shell
	data["shellPath"] = shellPath
	data["available"] = true

	// 如果没有初始命令，直接返回成功
	if config.InitialCommand != "" {
		// 步骤2：验证初始命令
		valid, errMsg, output := ts.validateInitialCommand(shellPath, config)
		if !valid {
			result.Message = errMsg
			return result
		}

		// 记录初始命令执行结果
		data["initialCommand"] = config.InitialCommand
		data["initialCommandExecuted"] = true
		data["output"] = output
	}

	result.Success = true
	result.Message = "终端配置验证通过"

	ts.Logger().Info("终端配置测试成功",
		"shell", shell,
		"shellPath", shellPath,
		"rows", config.Rows,
		"cols", config.Cols,
		"workPath", config.WorkPath,
		"initialCommand", config.InitialCommand)

	return result
}

// Test
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
