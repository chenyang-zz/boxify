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
	"testing"
)

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager()

	if sm == nil {
		t.Fatal("NewSessionManager returned nil")
	}

	if sm.sessions == nil {
		t.Error("sessions map should be initialized")
	}

	if sm.Count() != 0 {
		t.Errorf("new SessionManager should have 0 sessions, got %d", sm.Count())
	}
}

func TestSessionManager_AddAndGet(t *testing.T) {
	sm := NewSessionManager()

	// 创建测试会话
	session := createTestSession(t, "test-1")

	// 添加会话
	sm.Add(session)

	// 验证数量
	if sm.Count() != 1 {
		t.Errorf("expected 1 session, got %d", sm.Count())
	}

	// 获取会话
	got, ok := sm.Get("test-1")
	if !ok {
		t.Fatal("expected to find session")
	}

	if got.ID != session.ID {
		t.Errorf("got session ID %s, want %s", got.ID, session.ID)
	}

	// 获取不存在的会话
	_, ok = sm.Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent session")
	}
}

func TestSessionManager_Remove(t *testing.T) {
	sm := NewSessionManager()

	// 添加会话
	session := createTestSession(t, "test-1")
	sm.Add(session)

	// 移除会话
	removed, ok := sm.Remove("test-1")
	if !ok {
		t.Fatal("expected to remove session")
	}

	if removed.ID != session.ID {
		t.Errorf("removed session ID %s, want %s", removed.ID, session.ID)
	}

	// 验证数量
	if sm.Count() != 0 {
		t.Errorf("expected 0 sessions after remove, got %d", sm.Count())
	}

	// 再次移除应该失败
	_, ok = sm.Remove("test-1")
	if ok {
		t.Error("expected not to remove nonexistent session")
	}
}

func TestSessionManager_Count(t *testing.T) {
	sm := NewSessionManager()

	// 初始为 0
	if sm.Count() != 0 {
		t.Errorf("expected 0 sessions, got %d", sm.Count())
	}

	// 添加多个会话
	for i := 0; i < 5; i++ {
		sm.Add(createTestSession(t, string(rune('a'+i))))
	}

	if sm.Count() != 5 {
		t.Errorf("expected 5 sessions, got %d", sm.Count())
	}

	// 移除一个
	sm.Remove("a")
	if sm.Count() != 4 {
		t.Errorf("expected 4 sessions after remove, got %d", sm.Count())
	}
}

func TestSessionManager_IDs(t *testing.T) {
	sm := NewSessionManager()

	// 空时应该返回空切片
	ids := sm.IDs()
	if len(ids) != 0 {
		t.Errorf("expected empty IDs, got %v", ids)
	}

	// 添加会话
	expectedIDs := []string{"id-1", "id-2", "id-3"}
	for _, id := range expectedIDs {
		sm.Add(createTestSession(t, id))
	}

	ids = sm.IDs()
	if len(ids) != 3 {
		t.Errorf("expected 3 IDs, got %d", len(ids))
	}

	// 验证所有 ID 都存在
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for _, expected := range expectedIDs {
		if !idMap[expected] {
			t.Errorf("expected ID %s not found", expected)
		}
	}
}

func TestSessionManager_ForEach(t *testing.T) {
	sm := NewSessionManager()

	// 添加会话
	for i := 0; i < 3; i++ {
		sm.Add(createTestSession(t, string(rune('a'+i))))
	}

	// 遍历并收集 ID
	visitedIDs := make(map[string]bool)
	sm.ForEach(func(session *Session) {
		visitedIDs[session.ID] = true
	})

	// 验证所有会话都被访问
	if len(visitedIDs) != 3 {
		t.Errorf("expected to visit 3 sessions, visited %d", len(visitedIDs))
	}

	for _, id := range []string{"a", "b", "c"} {
		if !visitedIDs[id] {
			t.Errorf("session %s was not visited", id)
		}
	}
}

func TestSessionManager_CloseSession(t *testing.T) {
	sm := NewSessionManager()

	// 添加会话
	session := createTestSession(t, "test-1")
	sm.Add(session)

	// 关闭会话
	err := sm.CloseSession("test-1", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 验证会话被移除
	if sm.Count() != 0 {
		t.Errorf("expected 0 sessions after close, got %d", sm.Count())
	}

	// 关闭不存在的会话应该返回 nil
	err = sm.CloseSession("nonexistent", nil)
	if err != nil {
		t.Errorf("expected nil for nonexistent session, got %v", err)
	}
}

func TestSessionManager_CloseAll(t *testing.T) {
	sm := NewSessionManager()

	// 添加多个会话
	for i := 0; i < 5; i++ {
		sm.Add(createTestSession(t, string(rune('a'+i))))
	}

	// 关闭所有
	sm.CloseAll(nil)

	// 验证所有会话被移除
	if sm.Count() != 0 {
		t.Errorf("expected 0 sessions after CloseAll, got %d", sm.Count())
	}

	// 再次关闭应该安全
	sm.CloseAll(nil)
}

func TestSessionManager_Concurrent(t *testing.T) {
	sm := NewSessionManager()
	var wg sync.WaitGroup

	// 并发添加
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id := string(rune('a' + idx%26))
			sm.Add(createTestSession(t, id))
		}(i)
	}

	// 并发读取
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = sm.Count()
			_, _ = sm.Get(string(rune('a' + idx%26)))
			_ = sm.IDs()
		}(i)
	}

	wg.Wait()

	// 应该有 26 个不同的会话（a-z）
	count := sm.Count()
	if count != 26 {
		t.Errorf("expected 26 sessions, got %d", count)
	}
}

func TestSessionManager_CloseAllWithConfigGenerator(t *testing.T) {
	sm := NewSessionManager()
	configGenerator := NewShellConfigGenerator(testLogger)

	// 创建一个带有配置路径的会话
	session := NewSession(context.Background(), "test-1", os.Stdout, nil, ShellTypeBash, false, testLogger)
	session.SetConfigPath("/tmp/test-config") // 设置一个假的配置路径
	sm.Add(session)

	// 关闭所有会话
	sm.CloseAll(configGenerator)

	// 验证会话被移除
	if sm.Count() != 0 {
		t.Errorf("expected 0 sessions after CloseAll, got %d", sm.Count())
	}
}

// 辅助函数：创建测试会话
func createTestSession(t *testing.T, id string) *Session {
	t.Helper()
	cmd := exec.Command("echo", "test")
	return NewSession(context.Background(), id, os.Stdout, cmd, ShellTypeBash, false, testLogger)
}
