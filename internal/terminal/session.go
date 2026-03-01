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
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chenyang-zz/boxify/internal/types"
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

// Session 终端会话
type Session struct {
	ID           string
	Pty          *os.File
	Cmd          *exec.Cmd
	CreatedAt    time.Time
	ctx          context.Context    // 用于控制读取循环退出
	cancel       context.CancelFunc // 取消函数
	currentBlock string             // 当前活动的 block ID
	blockMutex   sync.RWMutex       // 保护 currentBlock

	// Shell Hooks 相关字段
	filter     *MarkerFilter   // 输出过滤器
	wrapper    *CommandWrapper // 命令包装器
	shellType  ShellType       // shell 类型
	useHooks   bool            // 是否使用 hooks 模式
	configPath string          // 临时配置文件路径
	workPath   string          // 当前工作路径

	// Git 监听器（Session 自管理）
	gitWatcher *GitWatcher
	emitter    EventEmitter // 事件发射器（用于 Git 状态更新）
	logger     *slog.Logger
}

// NewSession 创建新的终端会话
func NewSession(ctx context.Context, id string, pty *os.File, cmd *exec.Cmd, shellType ShellType, useHooks bool, logger *slog.Logger) *Session {
	sessionCtx, sessionCancel := context.WithCancel(ctx)

	return &Session{
		ID:        id,
		Pty:       pty,
		Cmd:       cmd,
		CreatedAt: time.Now(),
		ctx:       sessionCtx,
		cancel:    sessionCancel,
		filter:    NewMarkerFilter(logger),
		wrapper:   NewCommandWrapper(shellType, testLogger),
		shellType: shellType,
		useHooks:  useHooks,
		logger:    logger,
	}
}

// SetEmitter 设置事件发射器
func (s *Session) SetEmitter(emitter EventEmitter) {
	s.emitter = emitter
}

// SetLogger 设置日志器
func (s *Session) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

// StartGitWatcher 启动 Git 监听器
// 返回初始 Git 状态
func (s *Session) StartGitWatcher() *types.GitInfo {
	if s.gitWatcher != nil {
		s.gitWatcher.Stop()
	}

	s.gitWatcher = NewGitWatcher(s.emitter, s.logger)
	status, err := s.gitWatcher.Start(s.ID, s.workPath)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("启动 Git 监听失败", "sessionId", s.ID, "error", err)
		}
		return nil
	}

	return status
}

// StopGitWatcher 停止 Git 监听器
func (s *Session) StopGitWatcher() {
	if s.gitWatcher != nil {
		s.gitWatcher.Stop()
		s.gitWatcher = nil
	}
}

// UpdateGitWorkPath 更新 Git 监听器的工作目录
func (s *Session) UpdateGitWorkPath(newPwd string) *types.GitInfo {
	// 转换 ~ 为用户目录
	workPath := newPwd
	if strings.HasPrefix(newPwd, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			workPath = filepath.Join(homeDir, newPwd[1:])
		}
	}

	// 更新 session 的工作路径
	s.SetWorkPath(workPath)

	// 更新 Git 监听器
	if s.gitWatcher != nil {
		status, err := s.gitWatcher.UpdateWorkPath(workPath)
		if err != nil && s.logger != nil {
			s.logger.Warn("更新 Git 监听路径失败", "sessionId", s.ID, "error", err)
		}
		return status
	}

	return nil
}

// GitWatcher 返回 Git 监听器
func (s *Session) GitWatcher() *GitWatcher {
	return s.gitWatcher
}

// Context 返回会话的 context
func (s *Session) Context() context.Context {
	return s.ctx
}

// Cancel 取消会话的 context
func (s *Session) Cancel() {
	if s.cancel != nil {
		s.cancel()
	}
}

// Filter 返回输出过滤器
func (s *Session) Filter() *MarkerFilter {
	return s.filter
}

// Wrapper 返回命令包装器
func (s *Session) Wrapper() *CommandWrapper {
	return s.wrapper
}

// ShellType 返回 shell 类型
func (s *Session) ShellType() ShellType {
	return s.shellType
}

// UseHooks 返回是否使用 hooks 模式
func (s *Session) UseHooks() bool {
	return s.useHooks
}

// ConfigPath 返回临时配置文件路径
func (s *Session) ConfigPath() string {
	return s.configPath
}

// SetConfigPath 设置临时配置文件路径
func (s *Session) SetConfigPath(path string) {
	s.configPath = path
}

// CurrentBlock 返回当前活动的 block ID
func (s *Session) CurrentBlock() string {
	s.blockMutex.RLock()
	defer s.blockMutex.RUnlock()
	return s.currentBlock
}

// SetCurrentBlock 设置当前活动的 block ID
func (s *Session) SetCurrentBlock(blockID string) {
	s.blockMutex.Lock()
	defer s.blockMutex.Unlock()
	s.currentBlock = blockID
}

// WorkPath 返回当前工作路径
func (s *Session) WorkPath() string {
	return s.workPath
}

// SetWorkPath 设置当前工作路径
func (s *Session) SetWorkPath(path string) {
	s.workPath = path
}

// Close 关闭会话资源（不含 Process.Wait）
func (s *Session) Close() {
	// 先取消 context，通知读取循环退出
	s.Cancel()

	// 停止 Git 监听器
	s.StopGitWatcher()

	if s.Pty != nil {
		s.Pty.Close()
	}
}

// KillProcess 终止会话进程
func (s *Session) KillProcess() error {
	if s.Cmd != nil && s.Cmd.Process != nil {
		return s.Cmd.Process.Kill()
	}
	return nil
}

// WaitProcess 等待进程结束
func (s *Session) WaitProcess() error {
	if s.Cmd != nil {
		return s.Cmd.Wait()
	}
	return nil
}
