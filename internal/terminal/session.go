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
	"os"
	"os/exec"
	"sync"
	"time"
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
}

// NewSession 创建新的终端会话
func NewSession(ctx context.Context, id string, pty *os.File, cmd *exec.Cmd, shellType ShellType, useHooks bool) *Session {
	sessionCtx, sessionCancel := context.WithCancel(ctx)

	return &Session{
		ID:        id,
		Pty:       pty,
		Cmd:       cmd,
		CreatedAt: time.Now(),
		ctx:       sessionCtx,
		cancel:    sessionCancel,
		filter:    NewMarkerFilter(),
		wrapper:   NewCommandWrapper(shellType),
		shellType: shellType,
		useHooks:  useHooks,
	}
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

// Close 关闭会话资源（不含 Process.Wait）
func (s *Session) Close() {
	// 先取消 context，通知读取循环退出
	s.Cancel()

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
