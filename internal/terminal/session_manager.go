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
	"sync"
)

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// Get 获取会话
func (sm *SessionManager) Get(sessionID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, ok := sm.sessions[sessionID]
	return session, ok
}

// Add 添加会话
func (sm *SessionManager) Add(session *Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[session.ID] = session
}

// Remove 移除会话（返回被移除的会话）
func (sm *SessionManager) Remove(sessionID string) (*Session, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	session, ok := sm.sessions[sessionID]
	if ok {
		delete(sm.sessions, sessionID)
	}
	return session, ok
}

// CloseSession 关闭指定会话
func (sm *SessionManager) CloseSession(sessionID string, configGenerator *ShellConfigGenerator) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil
	}

	sm.closeSessionUnsafe(session, configGenerator)
	delete(sm.sessions, sessionID)

	return nil
}

// CloseAll 关闭所有会话
func (sm *SessionManager) CloseAll(configGenerator *ShellConfigGenerator) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, session := range sm.sessions {
		sm.closeSessionUnsafe(session, configGenerator)
	}

	// 清空 map
	sm.sessions = make(map[string]*Session)
}

// closeSessionUnsafe 内部方法：关闭会话（不加锁）
func (sm *SessionManager) closeSessionUnsafe(session *Session, configGenerator *ShellConfigGenerator) {
	// 关闭会话资源
	session.Close()

	// 终止进程并等待
	if err := session.KillProcess(); err != nil {
		// 忽略错误，进程可能已结束
	}
	if err := session.WaitProcess(); err != nil {
		// 忽略错误
	}

	// 清理临时配置文件
	if session.ConfigPath() != "" && configGenerator != nil {
		configGenerator.Cleanup(session.ConfigPath())
	}
}

// Count 获取会话数量
func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// ForEach 遍历所有会话
func (sm *SessionManager) ForEach(fn func(session *Session)) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, session := range sm.sessions {
		fn(session)
	}
}

// IDs 获取所有会话 ID
func (sm *SessionManager) IDs() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ids := make([]string, 0, len(sm.sessions))
	for id := range sm.sessions {
		ids = append(ids, id)
	}
	return ids
}
