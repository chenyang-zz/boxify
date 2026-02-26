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
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/creack/pty"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// 错误模式常量（预编译小写版本，避免重复分配）
var (
	// Unix/Linux/macOS 错误模式
	unixErrorPatterns = []string{
		"command not found",
		"no such file or directory",
		"permission denied",
		"cannot execute",
		"is not recognized",
		"invalid option",
		"illegal option",
		"error:",
		"failed",
		"unable to",
	}

	// Windows cmd.exe 错误模式
	windowsCmdErrorPatterns = []string{
		"'\"' is not recognized",
		"is not recognized as an internal or external command",
		"cannot find the path specified",
		"the system cannot find the file specified",
		"access is denied",
		"the syntax of the command is incorrect",
	}

	// PowerShell 错误模式
	powerShellErrorPatterns = []string{
		"term '",
		"' is not recognized",
		"cannot be found",
		"does not exist",
		"access to the path",
		"is denied",
		"error:",
		"exception",
	}

	// 根据平台合并的错误模式（初始化时计算一次）
	platformErrorPatterns = func() []string {
		switch runtime.GOOS {
		case "windows":
			// Windows: 合并所有模式
			patterns := make([]string, 0, len(windowsCmdErrorPatterns)+len(powerShellErrorPatterns)+len(unixErrorPatterns))
			patterns = append(patterns, windowsCmdErrorPatterns...)
			patterns = append(patterns, powerShellErrorPatterns...)
			patterns = append(patterns, unixErrorPatterns...) // 也包含 Unix 模式（WSL/Git Bash）
			return patterns
		default:
			// Unix/Linux/macOS
			return unixErrorPatterns
		}
	}()
)

type ShellType string

const (
	ShellTypeCmd        ShellType = "cmd"
	ShellTypePowershell ShellType = "powershell"
	ShellTypePwsh       ShellType = "pwsh"
	ShellTypeBash       ShellType = "bash"
	ShellTypeZsh        ShellType = "zsh"
	ShellTypeAuto       ShellType = "auto"
)

// TerminalConfig 终端配置
type TerminalConfig struct {
	ID             string    `json:"id"`                       // 会话 ID
	Shell          ShellType `json:"shell"`                    // shell 路径，"auto" 表示自动检测
	Rows           uint16    `json:"rows,omitempty"`           // 终端行数
	Cols           uint16    `json:"cols,omitempty"`           // 终端列数
	WorkPath       string    `json:"workPath,omitempty"`       // 工作路径
	InitialCommand string    `json:"initialCommand,omitempty"` // 初始命令
}

// TerminalSession 终端会话
type TerminalSession struct {
	ID        string
	Pty       *os.File
	Cmd       *exec.Cmd
	CreatedAt time.Time
}

// TerminalService 终端服务
type TerminalService struct {
	BaseService
	sessions   map[string]*TerminalSession
	mu         sync.RWMutex
	shellCache sync.Map // 缓存 shell 路径: cacheKey -> shellPath
}

// NewTerminalService 创建终端服务
func NewTerminalService(deps *ServiceDeps) *TerminalService {
	return &TerminalService{
		BaseService: NewBaseService(deps),
		sessions:    make(map[string]*TerminalSession),
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
func (ts *TerminalService) validateBasicConfig(config TerminalConfig) (bool, string, string) {
	// 验证终端尺寸（允许 0，表示使用默认值）
	if config.Rows < 0 || config.Rows > 300 {
		return false, "终端行数超出范围，支持 0-300 行（0 表示使用默认值）", ""
	}
	if config.Cols < 0 || config.Cols > 500 {
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
	shell := ts.detectShell(config.Shell)

	// 检查 shell 是否存在（使用缓存优化）
	cacheKey := "shell:" + shell
	var shellPath string
	if cached, ok := ts.shellCache.Load(cacheKey); ok {
		shellPath = cached.(string)
	} else {
		var err error
		shellPath, err = exec.LookPath(shell)
		if err != nil {
			return false, fmt.Sprintf("找不到 shell: %s (%v)", shell, err), ""
		}
		ts.shellCache.Store(cacheKey, shellPath)
	}

	return true, "", shellPath
}

// validateInitialCommand 验证初始命令是否能执行
// 返回: (是否有效, 错误信息, 输出内容)
func (ts *TerminalService) validateInitialCommand(shellPath string, config TerminalConfig) (bool, string, string) {
	// 检查命令是否为空或仅包含空白字符
	trimmed := strings.TrimSpace(config.InitialCommand)
	if trimmed == "" {
		return false, "初始命令不能为空或仅包含空白字符", ""
	}

	// 检查命令长度限制
	if len(config.InitialCommand) > 10000 {
		return false, "初始命令过长，最大支持 10000 字符", ""
	}

	// 设置工作目录（默认使用用户主目录）
	testCmd := exec.Command(shellPath)
	workPath := config.WorkPath
	if workPath == "" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			workPath = homeDir
		}
	}
	if workPath != "" {
		testCmd.Dir = workPath
	}

	// 设置环境变量
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
	_, _ = ptyFile.WriteString(initialCmd)

	// 写入 exit 命令让 shell 在执行完初始命令后退出
	_, _ = ptyFile.WriteString("exit\n")

	// 使用 context 实现超时控制（简化版）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 单一 goroutine 处理进程等待和输出读取
	type execResult struct {
		output string
		err    error
	}
	resultCh := make(chan execResult, 1)

	go func() {
		outputBuf := make([]byte, 4096)
		var outputBuilder strings.Builder

		// 在后台读取输出
		go func() {
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

		// 等待进程结束
		err := testCmd.Wait()

		// 等待一小段时间让输出读取完成
		time.Sleep(100 * time.Millisecond)

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
		if ts.hasCommandError(res.output) {
			return false, "初始命令执行出现错误，请检查命令是否正确", res.output
		}

		return true, "", res.output

	case <-ctx.Done():
		// 超时强制终止
		testCmd.Process.Kill()
		ptyFile.Close()
		return false, "初始命令执行超时（5秒）", ""
	}
}

// Create 创建新的终端会话
func (ts *TerminalService) Create(config TerminalConfig) *connection.QueryResult {
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

	// 创建命令
	cmd := exec.Command(shellPath)

	// 设置工作目录（默认使用用户主目录）
	workPath := config.WorkPath
	if workPath == "" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			workPath = homeDir
		}
	}
	if workPath != "" {
		cmd.Dir = workPath
	}

	// 设置环境变量以支持颜色输出
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	)

	// 启动 PTY
	ptyFile, err := pty.Start(cmd)
	if err != nil {
		ts.Logger().Error("创建 PTY 失败", "shell", shellPath, "error", err)
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
		_, _ = ptyFile.WriteString(initialCmd)
	}

	session := &TerminalSession{
		ID:        config.ID,
		Pty:       ptyFile,
		Cmd:       cmd,
		CreatedAt: time.Now(),
	}

	ts.mu.Lock()
	ts.sessions[config.ID] = session
	ts.mu.Unlock()

	// 启动输出读取 goroutine
	go ts.readOutputLoop(session)

	ts.Logger().Info("终端会话创建", "sessionId", config.ID, "shell", shellPath, "workPath", config.WorkPath, "initialCommand", config.InitialCommand)

	return &connection.QueryResult{
		Success: true,
		Message: "终端创建成功",
	}
}

// readOutputLoop 读取 PTY 输出并发送到前端
func (ts *TerminalService) readOutputLoop(session *TerminalSession) {
	buf := make([]byte, 1024)

	for {
		n, err := session.Pty.Read(buf)
		if err != nil {
			if err != io.EOF {
				ts.Logger().Error("读取 PTY 输出失败", "sessionId", session.ID, "error", err)
				ts.App().Event.Emit("terminal:error", map[string]interface{}{
					"sessionId": session.ID,
					"message":   err.Error(),
				})
			}
			break
		}

		// Base64 编码避免传输问题
		encoded := base64.StdEncoding.EncodeToString(buf[:n])
		ts.App().Event.Emit("terminal:output", map[string]interface{}{
			"sessionId": session.ID,
			"data":      encoded,
		})
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
	session, ok := ts.sessions[sessionID]
	ts.mu.RUnlock()

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
func (ts *TerminalService) closeSessionUnsafe(session *TerminalSession) {
	if session.Pty != nil {
		session.Pty.Close()
	}
	if session.Cmd != nil && session.Cmd.Process != nil {
		session.Cmd.Process.Kill()
		session.Cmd.Wait()
	}
}

// TestConfig 测试终端配置参数是否有效
func (ts *TerminalService) TestConfig(config TerminalConfig) *connection.QueryResult {
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

	shell := ts.detectShell(config.Shell)
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

// detectShell 检测系统默认 shell（带缓存优化）
func (ts *TerminalService) detectShell(preferred ShellType) string {
	if preferred != ShellTypeAuto {
		if preferred == ShellTypeCmd || preferred == ShellTypePowershell || preferred == ShellTypePwsh {
			return string(preferred) + ".exe"
		}
		return string(preferred)
	}

	// 检查缓存
	cacheKey := "default:" + runtime.GOOS
	if cached, ok := ts.shellCache.Load(cacheKey); ok {
		return cached.(string)
	}

	// 根据平台检测默认 shell
	var shell string
	switch runtime.GOOS {
	case "windows":
		// Windows: 优先 PowerShell，其次 cmd
		if path, err := exec.LookPath("pwsh"); err == nil {
			shell = path
		} else if path, err := exec.LookPath("powershell"); err == nil {
			shell = path
		} else {
			shell = "cmd.exe"
		}

	case "darwin", "linux":
		// Unix: 优先 zsh，其次 bash
		if path, err := exec.LookPath("zsh"); err == nil {
			shell = path
		} else if path, err := exec.LookPath("bash"); err == nil {
			shell = path
		} else {
			shell = "/bin/sh"
		}

	default:
		shell = "/bin/sh"
	}

	// 缓存结果
	ts.shellCache.Store(cacheKey, shell)
	return shell
}

// hasCommandError 检查命令输出中是否包含错误标志（优化版：使用预编译模式）
func (ts *TerminalService) hasCommandError(output string) bool {
	lowerOutput := strings.ToLower(output)

	// 使用预编译的平台特定错误模式
	for _, pattern := range platformErrorPatterns {
		if strings.Contains(lowerOutput, pattern) {
			return true
		}
	}

	return false
}
