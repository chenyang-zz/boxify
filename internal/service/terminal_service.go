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
	"sync"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/creack/pty"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// TerminalConfig 终端配置
type TerminalConfig struct {
	ID    string `json:"id"`    // 会话 ID
	Shell string `json:"shell"` // shell 路径，"auto" 表示自动检测
	Rows  uint16 `json:"rows"`  // 终端行数
	Cols  uint16 `json:"cols"`  // 终端列数
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
	sessions map[string]*TerminalSession
	mu       sync.RWMutex
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

// Create 创建新的终端会话
func (ts *TerminalService) Create(config TerminalConfig) *connection.QueryResult {
	shell := ts.detectShell(config.Shell)

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
	cmd := exec.Command(shell)

	// 设置环境变量以支持颜色输出
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	)

	// 启动 PTY
	ptyFile, err := pty.Start(cmd)
	if err != nil {
		ts.Logger().Error("创建 PTY 失败", "shell", shell, "error", err)
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("创建终端失败: %v", err),
		}
	}

	// 设置终端大小
	if err := pty.Setsize(ptyFile, &pty.Winsize{Rows: rows, Cols: cols}); err != nil {
		ts.Logger().Warn("设置终端大小失败", "error", err)
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

	ts.Logger().Info("终端会话创建", "sessionId", config.ID, "shell", shell)

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

// detectShell 检测系统默认 shell
func (ts *TerminalService) detectShell(preferred string) string {
	if preferred != "" && preferred != "auto" {
		return preferred
	}

	// 根据平台检测默认 shell
	switch runtime.GOOS {
	case "windows":
		// Windows: 优先 PowerShell，其次 cmd
		if _, err := exec.LookPath("pwsh"); err == nil {
			return "pwsh.exe"
		}
		if _, err := exec.LookPath("powershell"); err == nil {
			return "powershell.exe"
		}
		return "cmd.exe"

	case "darwin", "linux":
		// Unix: 优先 zsh，其次 bash
		if _, err := exec.LookPath("zsh"); err == nil {
			return "zsh"
		}
		if _, err := exec.LookPath("bash"); err == nil {
			return "bash"
		}
		return "/bin/sh"

	default:
		return "/bin/sh"
	}
}
