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
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name   string
		config TerminalConfig
	}{
		{
			name: "empty config",
			config: TerminalConfig{
				ID: "test-1",
			},
		},
		{
			name: "full config",
			config: TerminalConfig{
				ID:             "test-2",
				Shell:          ShellTypeZsh,
				Rows:           40,
				Cols:           120,
				WorkPath:       "/tmp",
				InitialCommand: "echo hello",
			},
		},
		{
			name: "auto shell",
			config: TerminalConfig{
				ID:    "test-3",
				Shell: ShellTypeAuto,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.ID == "" {
				t.Error("Config ID should not be empty")
			}
		})
	}
}

func TestNewSession(t *testing.T) {
	ctx := context.Background()
	id := "test-session-1"

	// 创建一个简单的命令（使用 echo）
	cmd := exec.Command("echo", "test")
	ptyFile, err := os.Stdout.Stat() // 使用 stdout 作为占位符
	if err != nil {
		t.Skip("无法获取 stdout")
	}

	session := NewSession(ctx, id, os.Stdout, cmd, ShellTypeBash, false)

	if session == nil {
		t.Fatal("NewSession returned nil")
	}

	if session.ID != id {
		t.Errorf("expected ID %s, got %s", id, session.ID)
	}

	if session.shellType != ShellTypeBash {
		t.Errorf("expected shell type %s, got %s", ShellTypeBash, session.shellType)
	}

	if session.useHooks {
		t.Error("expected useHooks to be false")
	}

	if session.filter == nil {
		t.Error("filter should not be nil")
	}

	if session.wrapper == nil {
		t.Error("wrapper should not be nil")
	}

	if session.ctx == nil {
		t.Error("ctx should not be nil")
	}

	_ = ptyFile // 使用变量避免编译错误
}

func TestSessionContext(t *testing.T) {
	ctx := context.Background()
	session := NewSession(ctx, "test", os.Stdout, nil, ShellTypeBash, false)

	// 测试 Context 方法
	if session.Context() == nil {
		t.Error("Context() returned nil")
	}

	// 测试取消前 context 未完成
	if session.Context().Err() != nil {
		t.Errorf("expected context to be active, got err: %v", session.Context().Err())
	}

	// 取消 context
	session.Cancel()

	// 测试取消后 context 已完成
	time.Sleep(10 * time.Millisecond) // 给一点时间让取消生效
	if session.Context().Err() == nil {
		t.Error("expected context to be cancelled")
	}
}

func TestSessionAccessors(t *testing.T) {
	session := NewSession(context.Background(), "test", os.Stdout, nil, ShellTypeZsh, true)
	session.SetConfigPath("/tmp/test-config")

	// 测试 ShellType
	if session.ShellType() != ShellTypeZsh {
		t.Errorf("expected ShellTypeZsh, got %s", session.ShellType())
	}

	// 测试 UseHooks
	if !session.UseHooks() {
		t.Error("expected UseHooks to be true")
	}

	// 测试 ConfigPath
	if session.ConfigPath() != "/tmp/test-config" {
		t.Errorf("expected /tmp/test-config, got %s", session.ConfigPath())
	}

	// 测试 Filter
	if session.Filter() == nil {
		t.Error("Filter() returned nil")
	}

	// 测试 Wrapper
	if session.Wrapper() == nil {
		t.Error("Wrapper() returned nil")
	}
}

func TestSessionCurrentBlock(t *testing.T) {
	session := NewSession(context.Background(), "test", os.Stdout, nil, ShellTypeBash, false)

	// 初始 block 应为空
	if session.CurrentBlock() != "" {
		t.Errorf("expected empty block, got %s", session.CurrentBlock())
	}

	// 设置 block
	testBlockID := "block-123"
	session.SetCurrentBlock(testBlockID)

	if session.CurrentBlock() != testBlockID {
		t.Errorf("expected %s, got %s", testBlockID, session.CurrentBlock())
	}

	// 并发测试
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			session.SetCurrentBlock("block-concurrent")
			_ = session.CurrentBlock()
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSessionCreatedAt(t *testing.T) {
	before := time.Now()
	session := NewSession(context.Background(), "test", os.Stdout, nil, ShellTypeBash, false)
	after := time.Now()

	if session.CreatedAt.Before(before) {
		t.Error("CreatedAt should be after session creation started")
	}

	if session.CreatedAt.After(after) {
		t.Error("CreatedAt should be before session creation ended")
	}
}

func TestSessionKillProcess(t *testing.T) {
	// 测试 nil Cmd 的情况
	session := NewSession(context.Background(), "test", os.Stdout, nil, ShellTypeBash, false)
	err := session.KillProcess()
	if err != nil {
		t.Errorf("KillProcess with nil Cmd should return nil, got %v", err)
	}

	// 测试 nil Process 的情况
	cmd := &exec.Cmd{}
	session = NewSession(context.Background(), "test", os.Stdout, cmd, ShellTypeBash, false)
	err = session.KillProcess()
	if err != nil {
		t.Errorf("KillProcess with nil Process should return nil, got %v", err)
	}
}

func TestSessionWaitProcess(t *testing.T) {
	// 测试 nil Cmd 的情况
	session := NewSession(context.Background(), "test", os.Stdout, nil, ShellTypeBash, false)
	err := session.WaitProcess()
	if err != nil {
		t.Errorf("WaitProcess with nil Cmd should return nil, got %v", err)
	}
}

func TestSessionClose(t *testing.T) {
	session := NewSession(context.Background(), "test", os.Stdout, nil, ShellTypeBash, false)

	// Close 不应该 panic
	session.Close()

	// 等待 context 取消生效
	time.Sleep(10 * time.Millisecond)

	// Context 应该被取消
	if session.Context().Err() == nil {
		t.Error("expected context to be cancelled after Close")
	}

	// 多次 Close 应该是安全的
	session.Close()
}
