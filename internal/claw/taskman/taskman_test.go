package taskman

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type testBroadcaster struct {
	mu     sync.Mutex
	events [][]byte
}

func (b *testBroadcaster) Broadcast(data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	cp := make([]byte, len(data))
	copy(cp, data)
	b.events = append(b.events, cp)
}

func (b *testBroadcaster) countByType(t *testing.T, eventType string) int {
	t.Helper()

	b.mu.Lock()
	defer b.mu.Unlock()

	count := 0
	for _, event := range b.events {
		var msg map[string]interface{}
		if err := json.Unmarshal(event, &msg); err != nil {
			t.Fatalf("解析广播消息失败: %v", err)
		}
		if msg["type"] == eventType {
			count++
		}
	}
	return count
}

func newTestManager(b Broadcaster) *Manager {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewManagerWithLogger(b, logger)
}

func testCommandForSuccess() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", "echo line1&&echo line2"}
	}
	return "sh", []string{"-c", "printf 'line1\nline2\n'"}
}

func testCommandForFailure() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", "exit 3"}
	}
	return "sh", []string{"-c", "exit 3"}
}

func TestCreateTaskShouldBroadcastUpdate(t *testing.T) {
	b := &testBroadcaster{}
	m := newTestManager(b)

	task := m.CreateTask("安装 OpenClaw", "install_openclaw")

	if task == nil {
		t.Fatal("创建任务返回 nil")
	}
	if task.Status != StatusPending {
		t.Fatalf("创建后任务状态错误: got=%s want=%s", task.Status, StatusPending)
	}
	if b.countByType(t, "task_update") != 1 {
		t.Fatalf("创建任务应广播 1 次 task_update，实际=%d", b.countByType(t, "task_update"))
	}
}

func TestGetRecentTasksShouldLimitAndSortByCreatedAtDesc(t *testing.T) {
	m := newTestManager(nil)
	now := time.Now()

	for i := 0; i < 55; i++ {
		id := "task-test-" + time.Now().Add(time.Duration(i)*time.Nanosecond).Format("150405.000000000")
		m.tasks[id] = &Task{
			ID:        id,
			Name:      "t",
			Type:      "x",
			Status:    StatusPending,
			Progress:  0,
			Log:       []string{},
			CreatedAt: now.Add(time.Duration(i) * time.Second),
			UpdatedAt: now.Add(time.Duration(i) * time.Second),
		}
	}

	recent := m.GetRecentTasks()
	if len(recent) != 50 {
		t.Fatalf("最近任务数量错误: got=%d want=50", len(recent))
	}

	for i := 1; i < len(recent); i++ {
		if recent[i-1].CreatedAt.Before(recent[i].CreatedAt) {
			t.Fatalf("任务未按创建时间倒序排列: idx=%d", i)
		}
	}
}

func TestHasRunningTaskShouldMatchType(t *testing.T) {
	m := newTestManager(nil)
	task := m.CreateTask("安装", "install_openclaw")
	task.SetStatus(StatusRunning)

	if !m.HasRunningTask("install_openclaw") {
		t.Fatal("应存在同类型运行中任务")
	}
	if m.HasRunningTask("install_wechat") {
		t.Fatal("不应匹配不同类型任务")
	}
}

func TestDedupeEnvShouldKeepLastValue(t *testing.T) {
	env := []string{"A=1", "B=2", "A=3", "NOEQ", "B=4"}
	got := dedupeEnv(env)

	expect := map[string]string{
		"A": "A=3",
		"B": "B=4",
	}
	for _, kv := range got {
		if strings.Contains(kv, "=") {
			parts := strings.SplitN(kv, "=", 2)
			if want, ok := expect[parts[0]]; ok && kv != want {
				t.Fatalf("去重结果错误: key=%s got=%s want=%s", parts[0], kv, want)
			}
		}
	}
}

func TestRunCommandShouldAppendLogsAndBroadcast(t *testing.T) {
	b := &testBroadcaster{}
	m := newTestManager(b)
	task := m.CreateTask("执行命令", "install_software")
	name, args := testCommandForSuccess()

	err := m.RunCommand(task, name, args...)
	if err != nil {
		t.Fatalf("运行命令失败: %v", err)
	}
	if task.Status != StatusRunning {
		t.Fatalf("RunCommand 后状态应为 running: got=%s", task.Status)
	}
	if len(task.Log) < 2 {
		t.Fatalf("应至少采集 2 行日志，实际=%d", len(task.Log))
	}
	if b.countByType(t, "task_log") < 2 {
		t.Fatalf("应至少广播 2 次 task_log，实际=%d", b.countByType(t, "task_log"))
	}
}

func TestRunCommandShouldReturnChineseWrappedErrorOnFailure(t *testing.T) {
	m := newTestManager(nil)
	task := m.CreateTask("失败命令", "install_software")
	name, args := testCommandForFailure()

	err := m.RunCommand(task, name, args...)
	if err == nil {
		t.Fatal("预期命令失败，但返回 nil")
	}
	if !strings.Contains(err.Error(), "命令执行失败") {
		t.Fatalf("错误信息应包含中文上下文: %v", err)
	}
}

func TestFinishTaskShouldSetStatusAndBroadcast(t *testing.T) {
	b := &testBroadcaster{}
	m := newTestManager(b)
	task := m.CreateTask("结束任务", "install_napcat")

	m.FinishTask(task, nil)
	if task.Status != StatusSuccess {
		t.Fatalf("成功结束后状态错误: got=%s want=%s", task.Status, StatusSuccess)
	}
	if task.Progress != 100 {
		t.Fatalf("成功结束后进度错误: got=%d want=100", task.Progress)
	}

	task2 := m.CreateTask("结束失败任务", "install_napcat")
	failErr := errors.New("mock err")
	m.FinishTask(task2, failErr)
	if task2.Status != StatusFailed {
		t.Fatalf("失败结束后状态错误: got=%s want=%s", task2.Status, StatusFailed)
	}
	if task2.Error != "mock err" {
		t.Fatalf("失败结束后错误字段错误: got=%s", task2.Error)
	}
	if b.countByType(t, "task_update") < 4 {
		t.Fatalf("应有多次 task_update 广播，实际=%d", b.countByType(t, "task_update"))
	}
}
