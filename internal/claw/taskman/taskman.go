package taskman

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	clawprocess "github.com/chenyang-zz/boxify/internal/claw/process"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	StatusPending  TaskStatus = "pending"
	StatusRunning  TaskStatus = "running"
	StatusPaused   TaskStatus = "paused"
	StatusSuccess  TaskStatus = "success"
	StatusFailed   TaskStatus = "failed"
	StatusCanceled TaskStatus = "canceled"
)

var ErrTaskCanceled = errors.New("任务已取消")

// Task 安装任务
type Task struct {
	ID               string             `json:"id"`                        // 任务唯一标识。
	Name             string             `json:"name"`                      // 任务名称。
	Type             string             `json:"type"`                      // 任务类型，如 install_software/install_openclaw。
	Status           TaskStatus         `json:"status"`                    // 当前任务状态。
	Paused           bool               `json:"paused"`                    // 当前任务是否已暂停。
	Progress         int                `json:"progress"`                  // 当前进度，范围 0-100。
	Stage            string             `json:"stage,omitempty"`           // 当前执行阶段，如 node/openclaw/done。
	NodeProgress     int                `json:"nodeProgress"`              // Node.js 安装阶段进度，范围 0-100。
	NodeMessage      string             `json:"nodeMessage,omitempty"`     // Node.js 安装阶段说明。
	OpenClawProgress int                `json:"openClawProgress"`          // OpenClaw 安装阶段进度，范围 0-100。
	OpenClawMessage  string             `json:"openClawMessage,omitempty"` // OpenClaw 安装阶段说明。
	Log              []string           `json:"log"`                       // 任务日志列表。
	Error            string             `json:"error,omitempty"`           // 任务失败时的错误信息。
	CreatedAt        time.Time          `json:"createdAt"`                 // 任务创建时间。
	UpdatedAt        time.Time          `json:"updatedAt"`                 // 任务最近更新时间。
	cancel           context.CancelFunc // 任务取消函数。
	cmd              *exec.Cmd          // 当前正在执行的命令。
	mu               sync.Mutex         // 保护任务可变字段的互斥锁。
}

// Broadcaster 用于向外广播任务事件。
type Broadcaster interface {
	Broadcast(data []byte)
}

// Manager 任务管理器
type Manager struct {
	tasks  map[string]*Task // 全量任务缓存，key 为任务 ID。
	hub    Broadcaster      // 任务事件广播器，可为空。
	logger *slog.Logger     // 任务管理日志器。
	mu     sync.RWMutex     // 保护任务缓存的读写锁。
}

// NewManager 创建任务管理器
func NewManager(hub Broadcaster, logger *slog.Logger) *Manager {
	return NewManagerWithLogger(hub, logger)
}

// NewManagerWithLogger 创建带日志器的任务管理器。
func NewManagerWithLogger(hub Broadcaster, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("module", "claw.taskman")

	return &Manager{
		tasks:  make(map[string]*Task),
		hub:    hub,
		logger: logger,
	}
}

// CreateTask 创建新任务
func (m *Manager) CreateTask(name, taskType string) *Task {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("task-%d", time.Now().UnixMilli())
	task := &Task{
		ID:               id,
		Name:             name,
		Type:             taskType,
		Status:           StatusPending,
		Paused:           false,
		Progress:         0,
		Stage:            "",
		NodeProgress:     0,
		NodeMessage:      "",
		OpenClawProgress: 0,
		OpenClawMessage:  "",
		Log:              []string{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	m.tasks[id] = task
	m.logger.Info("创建任务", "task_id", task.ID, "task_name", task.Name, "task_type", task.Type)
	m.broadcastTaskUpdate(task)
	return task
}

// GetTask 获取任务
func (m *Manager) GetTask(id string) *Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tasks[id]
}

// GetAllTasks 获取所有任务
func (m *Manager) GetAllTasks() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		result = append(result, t)
	}
	return result
}

// GetRecentTasks 获取最近的任务（最多50个）
func (m *Manager) GetRecentTasks() []*Task {
	tasks := m.GetAllTasks()
	// 按创建时间倒序排列
	for i := 0; i < len(tasks); i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[j].CreatedAt.After(tasks[i].CreatedAt) {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}
	if len(tasks) > 50 {
		tasks = tasks[:50]
	}
	return tasks
}

// HasRunningTask 检查是否有正在运行的同类型任务
func (m *Manager) HasRunningTask(taskType string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.tasks {
		if t.Type == taskType && t.Status == StatusRunning {
			return true
		}
	}
	return false
}

// AppendLog 追加日志
func (t *Task) AppendLog(line string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Log = append(t.Log, line)
	t.UpdatedAt = time.Now()
}

// SetProgress 设置进度
func (t *Task) SetProgress(p int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Progress = p
	t.UpdatedAt = time.Now()
}

// BindCommand 绑定当前执行命令与取消函数。
func (t *Task) BindCommand(cmd *exec.Cmd, cancel context.CancelFunc) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cmd = cmd
	t.cancel = cancel
	t.Paused = false
	t.UpdatedAt = time.Now()
}

// ClearCommand 清理当前执行命令上下文。
func (t *Task) ClearCommand() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cmd = nil
	t.cancel = nil
	t.Paused = false
	t.UpdatedAt = time.Now()
}

// MarkPaused 更新暂停状态。
func (t *Task) MarkPaused(paused bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Paused = paused
	t.UpdatedAt = time.Now()
}

// SetStageProgress 设置安装阶段进度与说明。
func (t *Task) SetStageProgress(stage string, progress int, message string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Stage = stage
	switch stage {
	case "node":
		t.NodeProgress = progress
		t.NodeMessage = message
	case "openclaw", "done":
		t.OpenClawProgress = progress
		t.OpenClawMessage = message
	}
	t.UpdatedAt = time.Now()
}

// SetStatus 设置状态
func (t *Task) SetStatus(s TaskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Status = s
	t.UpdatedAt = time.Now()
}

// Snapshot 返回任务状态快照，避免外部直接读取内部运行字段。
func (t *Task) Snapshot() Task {
	t.mu.Lock()
	defer t.mu.Unlock()

	return Task{
		ID:               t.ID,
		Name:             t.Name,
		Type:             t.Type,
		Status:           t.Status,
		Paused:           t.Paused,
		Progress:         t.Progress,
		Stage:            t.Stage,
		NodeProgress:     t.NodeProgress,
		NodeMessage:      t.NodeMessage,
		OpenClawProgress: t.OpenClawProgress,
		OpenClawMessage:  t.OpenClawMessage,
		Log:              append([]string(nil), t.Log...),
		Error:            t.Error,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}

// IsCanceled 判断任务是否已取消。
func (t *Task) IsCanceled() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Status == StatusCanceled
}

// currentCommand 返回当前命令与取消函数快照。
func (t *Task) currentCommand() (*exec.Cmd, context.CancelFunc, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cmd, t.cancel, t.Paused
}

// RunCommand 运行命令并实时推送输出
func (m *Manager) RunCommand(task *Task, name string, args ...string) error {
	if task.IsCanceled() {
		return ErrTaskCanceled
	}
	m.logger.Info("开始执行任务命令", "task_id", task.ID, "task_type", task.Type, "command", name, "args", args)
	task.SetStatus(StatusRunning)
	task.MarkPaused(false)
	m.broadcastTaskUpdate(task)

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, name, args...)
	configureTaskCommand(cmd)
	env := append(clawprocess.BuildExecEnv(),
		"DEBIAN_FRONTEND=noninteractive",
		"LANG=en_US.UTF-8",
	)

	cmd.Env = dedupeEnv(env)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.logger.Error("创建命令输出管道失败", "task_id", task.ID, "command", name, "error", err)
		return fmt.Errorf("创建命令输出管道失败: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		cancel()
		m.logger.Error("启动命令失败", "task_id", task.ID, "command", name, "error", err)
		return fmt.Errorf("启动命令失败: %w", err)
	}
	task.BindCommand(cmd, cancel)
	defer task.ClearCommand()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*64), 1024*64)
	for scanner.Scan() {
		line := scanner.Text()
		task.AppendLog(line)
		m.broadcastTaskLog(task, line)
	}
	if err := scanner.Err(); err != nil {
		if task.IsCanceled() || errors.Is(ctx.Err(), context.Canceled) {
			m.logger.Warn("命令读取因任务取消而结束", "task_id", task.ID, "command", name)
			return ErrTaskCanceled
		}
		m.logger.Error("读取命令输出失败", "task_id", task.ID, "command", name, "error", err)
		return fmt.Errorf("读取命令输出失败: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		if task.IsCanceled() || errors.Is(ctx.Err(), context.Canceled) {
			m.logger.Warn("命令因任务取消而终止", "task_id", task.ID, "command", name)
			return ErrTaskCanceled
		}
		m.logger.Error("命令执行失败", "task_id", task.ID, "command", name, "error", err)
		return fmt.Errorf("命令执行失败: %w", err)
	}
	m.logger.Info("任务命令执行完成", "task_id", task.ID, "task_type", task.Type, "command", name)
	return nil
}

// dedupeEnv 对环境变量按 key 去重，后出现的值覆盖先出现的值。
func dedupeEnv(env []string) []string {
	seen := make(map[string]int)
	out := make([]string, 0, len(env))

	for _, kv := range env {
		eq := strings.IndexByte(kv, '=')
		if eq <= 0 {
			out = append(out, kv)
			continue
		}
		key := kv[:eq]
		if idx, ok := seen[key]; ok {
			out[idx] = kv
			continue
		}
		seen[key] = len(out)
		out = append(out, kv)
	}

	return out
}

// RunScript 运行脚本并实时推送输出（Windows 用 PowerShell，其他平台用 bash）
func (m *Manager) RunScript(task *Task, script string) error {
	m.logger.Info("开始执行任务脚本", "task_id", task.ID, "task_type", task.Type, "os", runtime.GOOS)
	if runtime.GOOS == "windows" {
		return m.RunCommand(task, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script)
	}
	return m.RunCommand(task, "bash", "-c", script)
}

// RunScriptWithSudo 使用 sudo 运行脚本
func (m *Manager) RunScriptWithSudo(task *Task, sudoPass, script string) error {
	m.logger.Info("开始执行 sudo 任务脚本", "task_id", task.ID, "task_type", task.Type)
	// 使用 base64 包裹脚本，减少转义导致的执行错误。
	encoded := base64.StdEncoding.EncodeToString([]byte(script))
	fullScript := fmt.Sprintf("echo '%s' | sudo -S bash -c \"$(echo %s | base64 -d)\"", sudoPass, encoded)
	return m.RunCommand(task, "bash", "-c", fullScript)
}

// PauseTask 暂停正在执行的任务。
func (m *Manager) PauseTask(id string) error {
	task := m.GetTask(strings.TrimSpace(id))
	if task == nil {
		return fmt.Errorf("任务不存在")
	}

	cmd, _, paused := task.currentCommand()
	if cmd == nil || cmd.Process == nil {
		return fmt.Errorf("当前任务没有可暂停的下载进程")
	}
	if paused {
		return fmt.Errorf("任务已暂停")
	}
	if err := pauseProcess(cmd); err != nil {
		return fmt.Errorf("暂停任务失败: %w", err)
	}

	task.MarkPaused(true)
	task.SetStatus(StatusPaused)
	task.AppendLog("⏸️ 已暂停当前下载")
	m.broadcastTaskUpdate(task)
	m.logger.Info("任务已暂停", "task_id", task.ID, "task_type", task.Type)
	return nil
}

// ResumeTask 恢复已暂停的任务。
func (m *Manager) ResumeTask(id string) error {
	task := m.GetTask(strings.TrimSpace(id))
	if task == nil {
		return fmt.Errorf("任务不存在")
	}

	cmd, _, paused := task.currentCommand()
	if cmd == nil || cmd.Process == nil {
		return fmt.Errorf("当前任务没有可恢复的下载进程")
	}
	if !paused {
		return fmt.Errorf("任务当前未暂停")
	}
	if err := resumeProcess(cmd); err != nil {
		return fmt.Errorf("恢复任务失败: %w", err)
	}

	task.MarkPaused(false)
	task.SetStatus(StatusRunning)
	task.AppendLog("▶️ 已恢复当前下载")
	m.broadcastTaskUpdate(task)
	m.logger.Info("任务已恢复", "task_id", task.ID, "task_type", task.Type)
	return nil
}

// CancelTask 取消正在执行的任务。
func (m *Manager) CancelTask(id string) error {
	task := m.GetTask(strings.TrimSpace(id))
	if task == nil {
		return fmt.Errorf("任务不存在")
	}

	cmd, cancel, _ := task.currentCommand()
	if cancel == nil && (cmd == nil || task.Status != StatusPending) {
		return fmt.Errorf("当前任务不可取消")
	}

	task.SetStatus(StatusCanceled)
	task.MarkPaused(false)
	task.AppendLog("🛑 用户取消了当前任务")
	if cancel != nil {
		cancel()
	}
	if cmd != nil {
		if err := terminateProcess(cmd); err != nil {
			m.logger.Warn("终止任务进程失败", "task_id", task.ID, "error", err)
		}
	}
	m.broadcastTaskUpdate(task)
	m.logger.Info("任务已取消", "task_id", task.ID, "task_type", task.Type)
	return nil
}

// broadcastTaskUpdate 广播任务状态更新
func (m *Manager) broadcastTaskUpdate(task *Task) {
	if m.hub == nil {
		return
	}
	task.mu.Lock()
	msg := map[string]interface{}{
		"type": "task_update",
		"task": map[string]interface{}{
			"id":               task.ID,
			"name":             task.Name,
			"type":             task.Type,
			"status":           task.Status,
			"paused":           task.Paused,
			"progress":         task.Progress,
			"stage":            task.Stage,
			"nodeProgress":     task.NodeProgress,
			"nodeMessage":      task.NodeMessage,
			"openClawProgress": task.OpenClawProgress,
			"openClawMessage":  task.OpenClawMessage,
			"error":            task.Error,
			"createdAt":        task.CreatedAt.Format(time.RFC3339),
			"updatedAt":        task.UpdatedAt.Format(time.RFC3339),
			"logCount":         len(task.Log),
		},
	}
	task.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		m.logger.Warn("序列化任务更新事件失败", "task_id", task.ID, "error", err)
		return
	}
	m.hub.Broadcast(data)
}

// broadcastTaskLog 广播任务日志行
func (m *Manager) broadcastTaskLog(task *Task, line string) {
	if m.hub == nil {
		return
	}
	msg := map[string]interface{}{
		"type":   "task_log",
		"taskId": task.ID,
		"line":   line,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		m.logger.Warn("序列化任务日志事件失败", "task_id", task.ID, "error", err)
		return
	}
	m.hub.Broadcast(data)
}

// FinishTask 完成任务
func (m *Manager) FinishTask(task *Task, err error) {
	if errors.Is(err, ErrTaskCanceled) || task.IsCanceled() {
		task.SetStatus(StatusCanceled)
		task.MarkPaused(false)
		task.AppendLog("🛑 已取消")
		m.logger.Warn("任务已取消结束", "task_id", task.ID, "task_type", task.Type)
	} else if err != nil {
		task.SetStatus(StatusFailed)
		task.mu.Lock()
		task.Error = err.Error()
		task.mu.Unlock()
		task.AppendLog(fmt.Sprintf("❌ 失败: %v", err))
		m.logger.Error("任务执行失败", "task_id", task.ID, "task_type", task.Type, "error", err)
	} else {
		task.SetStatus(StatusSuccess)
		task.MarkPaused(false)
		task.SetProgress(100)
		task.AppendLog("✅ 完成")
		m.logger.Info("任务执行完成", "task_id", task.ID, "task_type", task.Type)
	}
	m.broadcastTaskUpdate(task)
}
