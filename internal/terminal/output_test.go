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
	"encoding/base64"
	"os"
	"testing"
	"time"
)

// mockEventEmitter 用于测试的模拟事件发射器
type mockEventEmitter struct {
	events []mockEvent
}

type mockEvent struct {
	name string
	data map[string]interface{}
}

func (m *mockEventEmitter) Emit(event string, data map[string]interface{}) {
	m.events = append(m.events, mockEvent{name: event, data: data})
}

func TestNewOutputHandler(t *testing.T) {
	emitter := &mockEventEmitter{}

	handler := NewOutputHandler(emitter, testLogger)

	if handler == nil {
		t.Fatal("NewOutputHandler returned nil")
	}

	if handler.emitter == nil {
		t.Error("emitter should not be nil")
	}
}

func TestNewOutputHandler_NilDependencies(t *testing.T) {
	// nil emitter 应该是允许的
	handler := NewOutputHandler(nil, testLogger)
	if handler == nil {
		t.Fatal("NewOutputHandler returned nil")
	}

	// nil logger 应该是允许的
	handler = NewOutputHandler(&mockEventEmitter{}, nil)
	if handler == nil {
		t.Fatal("NewOutputHandler returned nil")
	}
}

func TestOutputHandler_emitOutput(t *testing.T) {
	emitter := &mockEventEmitter{}
	handler := NewOutputHandler(emitter, testLogger)

	// 测试发送输出
	output := []byte("hello world")
	handler.emitOutput("session-1", "block-1", output)

	if len(emitter.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(emitter.events))
	}

	event := emitter.events[0]
	if event.name != "terminal:output" {
		t.Errorf("event name = %s, want terminal:output", event.name)
	}

	// 验证数据
	if event.data["sessionId"] != "session-1" {
		t.Errorf("sessionId = %v, want session-1", event.data["sessionId"])
	}
	if event.data["blockId"] != "block-1" {
		t.Errorf("blockId = %v, want block-1", event.data["blockId"])
	}

	// 验证 base64 编码
	encoded, ok := event.data["data"].(string)
	if !ok {
		t.Fatal("data should be a string")
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("failed to decode base64: %v", err)
	}

	if string(decoded) != "hello world" {
		t.Errorf("decoded data = %s, want hello world", decoded)
	}
}

func TestOutputHandler_emitOutput_NilEmitter(t *testing.T) {
	handler := NewOutputHandler(nil, testLogger)

	// 不应该 panic
	handler.emitOutput("session-1", "block-1", []byte("test"))
}

func TestOutputHandler_emitError(t *testing.T) {
	emitter := &mockEventEmitter{}
	handler := NewOutputHandler(emitter, testLogger)

	handler.emitError("session-1", "test error message")

	if len(emitter.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(emitter.events))
	}

	event := emitter.events[0]
	if event.name != "terminal:error" {
		t.Errorf("event name = %s, want terminal:error", event.name)
	}

	if event.data["sessionId"] != "session-1" {
		t.Errorf("sessionId = %v, want session-1", event.data["sessionId"])
	}
	if event.data["message"] != "test error message" {
		t.Errorf("message = %v, want test error message", event.data["message"])
	}
}

func TestOutputHandler_emitError_NilEmitter(t *testing.T) {
	handler := NewOutputHandler(nil, testLogger)

	// 不应该 panic
	handler.emitError("session-1", "test error")
}

func TestOutputHandler_emitCommandEnd(t *testing.T) {
	emitter := &mockEventEmitter{}
	handler := NewOutputHandler(emitter, testLogger)

	handler.emitCommandEnd("session-1", "block-1", 0)

	if len(emitter.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(emitter.events))
	}

	event := emitter.events[0]
	if event.name != "terminal:command_end" {
		t.Errorf("event name = %s, want terminal:command_end", event.name)
	}

	if event.data["sessionId"] != "session-1" {
		t.Errorf("sessionId = %v, want session-1", event.data["sessionId"])
	}
	if event.data["blockId"] != "block-1" {
		t.Errorf("blockId = %v, want block-1", event.data["blockId"])
	}
	if event.data["exitCode"] != 0 {
		t.Errorf("exitCode = %v, want 0", event.data["exitCode"])
	}

	// 测试非零退出码
	emitter.events = nil
	handler.emitCommandEnd("session-2", "block-2", 1)

	if emitter.events[0].data["exitCode"] != 1 {
		t.Errorf("exitCode = %v, want 1", emitter.events[0].data["exitCode"])
	}
}

func TestOutputHandler_emitCommandEnd_NilEmitter(t *testing.T) {
	handler := NewOutputHandler(nil, testLogger)

	// 不应该 panic
	handler.emitCommandEnd("session-1", "block-1", 0)
}

func TestOutputHandler_StartOutputLoop_ContextCancellation(t *testing.T) {
	emitter := &mockEventEmitter{}
	handler := NewOutputHandler(emitter, testLogger)

	// 创建一个会话，使用 pipe 来模拟 PTY
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	session := NewSession(context.Background(), "test-session", r, nil, ShellTypeBash, false, testLogger)

	// 启动输出循环
	done := make(chan struct{})
	go func() {
		handler.StartOutputLoop(session)
		close(done)
	}()

	// 等待一小段时间让循环开始
	time.Sleep(50 * time.Millisecond)

	// 取消 context 并关闭写入端来触发读取结束
	session.Cancel()
	w.Close()

	// 等待循环结束
	select {
	case <-done:
		// 成功退出
	case <-time.After(2 * time.Second):
		t.Error("StartOutputLoop did not exit after context cancellation")
	}

	// 清理
	r.Close()
}

func TestOutputHandler_StartOutputLoop_DataProcessing(t *testing.T) {
	emitter := &mockEventEmitter{}
	handler := NewOutputHandler(emitter, testLogger)

	// 创建一个会话，使用 pipe 来模拟 PTY
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	session := NewSession(context.Background(), "test-session", r, nil, ShellTypeBash, false, testLogger)
	session.SetCurrentBlock("test-block")

	// 启动输出循环
	done := make(chan struct{})
	go func() {
		handler.StartOutputLoop(session)
		close(done)
	}()

	// 写入一些数据（包含 OSC 133 标记）
	testData := "\x1b]133;A;cmd=test\x07hello world\x1b]133;D;exit=0\x07"
	_, err = w.WriteString(testData)
	if err != nil {
		t.Fatalf("failed to write to pipe: %v", err)
	}

	// 等待事件被发送
	time.Sleep(100 * time.Millisecond)

	// 取消 context 并关闭写入端
	session.Cancel()
	w.Close()

	// 等待循环结束
	select {
	case <-done:
		// 成功退出
	case <-time.After(2 * time.Second):
		t.Error("StartOutputLoop did not exit")
	}

	// 清理
	r.Close()

	// 验证事件（根据过滤器的行为，可能不会发送所有数据）
	// 由于过滤器会处理 OSC 133 标记，我们主要验证循环正常工作
	t.Logf("received %d events", len(emitter.events))
}

func TestOutputHandler_StartOutputLoop_EmptyOutput(t *testing.T) {
	emitter := &mockEventEmitter{}
	handler := NewOutputHandler(emitter, testLogger)

	// 创建一个会话
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	session := NewSession(context.Background(), "test-session", r, nil, ShellTypeBash, false, testLogger)

	// 启动输出循环
	done := make(chan struct{})
	go func() {
		handler.StartOutputLoop(session)
		close(done)
	}()

	// 写入 OSC 133 标记但不包含实际输出
	testData := "\x1b]133;A;cmd=test\x07\x1b]133;D;exit=0\x07"
	_, err = w.WriteString(testData)
	if err != nil {
		t.Fatalf("failed to write to pipe: %v", err)
	}

	// 等待事件被发送
	time.Sleep(100 * time.Millisecond)

	// 取消 context 并关闭写入端
	session.Cancel()
	w.Close()

	// 等待循环结束
	select {
	case <-done:
		// 成功退出
	case <-time.After(2 * time.Second):
		t.Error("StartOutputLoop did not exit")
	}

	// 清理
	r.Close()
}

func TestEventEmitter_Interface(t *testing.T) {
	// 验证 mockEventEmitter 实现了 EventEmitter 接口
	var _ EventEmitter = (*mockEventEmitter)(nil)
}

func TestOutputHandler_MultipleEvents(t *testing.T) {
	emitter := &mockEventEmitter{}
	handler := NewOutputHandler(emitter, testLogger)

	// 发送多个事件
	handler.emitOutput("session-1", "block-1", []byte("output 1"))
	handler.emitOutput("session-1", "block-2", []byte("output 2"))
	handler.emitCommandEnd("session-1", "block-2", 0)
	handler.emitError("session-1", "error message")

	if len(emitter.events) != 4 {
		t.Errorf("expected 4 events, got %d", len(emitter.events))
	}

	// 验证事件顺序
	expectedNames := []string{"terminal:output", "terminal:output", "terminal:command_end", "terminal:error"}
	for i, event := range emitter.events {
		if event.name != expectedNames[i] {
			t.Errorf("event %d: name = %s, want %s", i, event.name, expectedNames[i])
		}
	}
}
