package taskman

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	StatusPending  TaskStatus = "pending"
	StatusRunning  TaskStatus = "running"
	StatusSuccess  TaskStatus = "success"
	StatusFailed   TaskStatus = "failed"
	StatusCanceled TaskStatus = "canceled"
)

// Task 安装任务
type Task struct {
	ID        string     `json:"id"`              // 任务唯一标识。
	Name      string     `json:"name"`            // 任务名称。
	Type      string     `json:"type"`            // 任务类型，如 install_software/install_openclaw。
	Status    TaskStatus `json:"status"`          // 当前任务状态。
	Progress  int        `json:"progress"`        // 当前进度，范围 0-100。
	Log       []string   `json:"log"`             // 任务日志列表。
	Error     string     `json:"error,omitempty"` // 任务失败时的错误信息。
	CreatedAt time.Time  `json:"createdAt"`       // 任务创建时间。
	UpdatedAt time.Time  `json:"updatedAt"`       // 任务最近更新时间。
	cancel    func()     // 任务取消函数。
	mu        sync.Mutex // 保护任务可变字段的互斥锁。
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
		ID:        id,
		Name:      name,
		Type:      taskType,
		Status:    StatusPending,
		Progress:  0,
		Log:       []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

// SetStatus 设置状态
func (t *Task) SetStatus(s TaskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Status = s
	t.UpdatedAt = time.Now()
}

// RunCommand 运行命令并实时推送输出
func (m *Manager) RunCommand(task *Task, name string, args ...string) error {
	m.logger.Info("开始执行任务命令", "task_id", task.ID, "task_type", task.Type, "command", name, "args", args)
	task.SetStatus(StatusRunning)
	m.broadcastTaskUpdate(task)

	cmd := exec.Command(name, args...)
	env := cmd.Environ()

	if runtime.GOOS == "windows" {
		if os.Getenv("USERPROFILE") == "" {
			if home, err := os.UserHomeDir(); err == nil && home != "" {
				m.logger.Debug("检测到 USERPROFILE 缺失，使用用户目录补齐", "task_id", task.ID, "home", home)
				env = append(env, "USERPROFILE="+home)
			}
		}
	} else {
		home := os.Getenv("HOME")
		if home == "" {
			home, _ = os.UserHomeDir()
		}
		if home == "" {
			if runtime.GOOS == "darwin" {
				home = "/var/root"
			} else {
				home = "/root"
			}
		}
		if os.Getenv("HOME") == "" && home != "" {
			m.logger.Debug("检测到 HOME 缺失，使用回退目录补齐", "task_id", task.ID, "home", home)
			env = append(env, "HOME="+home)
		}
	}

	if os.Getenv("PATH") == "" {
		m.logger.Warn("检测到 PATH 缺失，使用默认 PATH 兜底", "task_id", task.ID, "os", runtime.GOOS)
		if runtime.GOOS == "windows" {
			env = append(env, "PATH=C:\\Windows\\System32;C:\\Windows;C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\")
		} else {
			env = append(env, "PATH=/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin")
		}
	}

	env = append(env,
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
		m.logger.Error("启动命令失败", "task_id", task.ID, "command", name, "error", err)
		return fmt.Errorf("启动命令失败: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*64), 1024*64)
	for scanner.Scan() {
		line := scanner.Text()
		task.AppendLog(line)
		m.broadcastTaskLog(task, line)
	}
	if err := scanner.Err(); err != nil {
		m.logger.Error("读取命令输出失败", "task_id", task.ID, "command", name, "error", err)
		return fmt.Errorf("读取命令输出失败: %w", err)
	}

	if err := cmd.Wait(); err != nil {
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

// broadcastTaskUpdate 广播任务状态更新
func (m *Manager) broadcastTaskUpdate(task *Task) {
	if m.hub == nil {
		return
	}
	task.mu.Lock()
	msg := map[string]interface{}{
		"type": "task_update",
		"task": map[string]interface{}{
			"id":        task.ID,
			"name":      task.Name,
			"type":      task.Type,
			"status":    task.Status,
			"progress":  task.Progress,
			"error":     task.Error,
			"createdAt": task.CreatedAt.Format(time.RFC3339),
			"updatedAt": task.UpdatedAt.Format(time.RFC3339),
			"logCount":  len(task.Log),
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
	if err != nil {
		task.SetStatus(StatusFailed)
		task.mu.Lock()
		task.Error = err.Error()
		task.mu.Unlock()
		task.AppendLog(fmt.Sprintf("❌ 失败: %v", err))
		m.logger.Error("任务执行失败", "task_id", task.ID, "task_type", task.Type, "error", err)
	} else {
		task.SetStatus(StatusSuccess)
		task.SetProgress(100)
		task.AppendLog("✅ 完成")
		m.logger.Info("任务执行完成", "task_id", task.ID, "task_type", task.Type)
	}
	m.broadcastTaskUpdate(task)
}
