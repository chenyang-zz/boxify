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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/chenyang-zz/boxify/internal/types"
	"github.com/fsnotify/fsnotify"
)

// GitWatcher Git 文件系统监听器
type GitWatcher struct {
	mu          sync.Mutex
	watcher     *fsnotify.Watcher
	sessionID   string
	workPath    string
	gitDir      string
	emitter     EventEmitter
	logger      *slog.Logger
	debounce    *time.Timer
	debounceMux sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewGitWatcher 创建 Git 监听器
func NewGitWatcher(emitter EventEmitter, logger *slog.Logger) *GitWatcher {
	return &GitWatcher{
		emitter: emitter,
		logger:  logger,
	}
}

// Start 启动监听
// sessionID: 会话 ID
// workPath: 工作目录
// 返回初始 Git 状态
func (w *GitWatcher) Start(sessionID, workPath string) (*types.GitInfo, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 先停止之前的监听
	w.stopLocked()

	w.sessionID = sessionID
	w.workPath = workPath
	w.ctx, w.cancel = context.WithCancel(context.Background())

	// 检查是否是 Git 仓库
	gitDir := filepath.Join(workPath, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		// 不是 Git 仓库
		return &types.GitInfo{IsRepo: false}, nil
	}

	w.gitDir = gitDir
	if info.IsDir() {
		// 如果 .git 是目录，直接使用
	} else {
		// 如果 .git 是文件（git worktree 或 submodule），读取内容获取真实路径
		content, err := os.ReadFile(gitDir)
		if err == nil {
			// 内容格式: gitdir: /path/to/.git/worktrees/xxx
			if strings.HasPrefix(string(content), "gitdir: ") {
				w.gitDir = strings.TrimSpace(strings.TrimPrefix(string(content), "gitdir: "))
			}
		}
	}

	// 创建 fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.logger.Error("创建 fsnotify watcher 失败", "error", err)
		return nil, fmt.Errorf("创建文件监听器失败: %w", err)
	}
	w.watcher = watcher

	// 监听 .git 目录
	if err := w.addWatch(w.gitDir); err != nil {
		w.watcher.Close()
		w.watcher = nil
		return nil, err
	}

	// 启动事件处理 goroutine
	go w.watchLoop()

	// 获取初始状态
	status := w.getGitStatus()

	w.logger.Info("Git 监听已启动", "workPath", workPath, "gitDir", w.gitDir)

	return status, nil
}

// addWatch 添加监听目录
func (w *GitWatcher) addWatch(gitDir string) error {
	// 监听核心文件
	paths := []string{
		filepath.Join(gitDir, "HEAD"),
		filepath.Join(gitDir, "index"),
		filepath.Join(gitDir, "refs"),
	}

	for _, path := range paths {
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				if err := w.watcher.Add(path); err != nil {
					w.logger.Warn("添加目录监听失败", "path", path, "error", err)
				}
			} else {
				if err := w.watcher.Add(path); err != nil {
					w.logger.Warn("添加文件监听失败", "path", path, "error", err)
				}
			}
		}
	}

	return nil
}

// watchLoop 监听循环
func (w *GitWatcher) watchLoop() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			// 只关心写入和创建事件
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				w.logger.Debug("Git 文件变化", "event", event.String())
				w.triggerDebounce()
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.logger.Error("Git 监听错误", "error", err)
		}
	}
}

// triggerDebounce 触发防抖更新
func (w *GitWatcher) triggerDebounce() {
	w.debounceMux.Lock()
	defer w.debounceMux.Unlock()

	// 如果已有定时器，先取消
	if w.debounce != nil {
		w.debounce.Stop()
	}

	// 设置新的定时器（500ms 防抖）
	w.debounce = time.AfterFunc(500*time.Millisecond, func() {
		w.emitGitStatusUpdate()
	})
}

// emitGitStatusUpdate 发送 Git 状态更新事件
func (w *GitWatcher) emitGitStatusUpdate() {
	status := w.getGitStatus()

	if w.emitter != nil {
		w.emitter.Emit("terminal:git_update", map[string]interface{}{
			"sessionId": w.sessionID,
			"git":       status,
		})
	}
}

// getGitStatus 获取当前 Git 状态
func (w *GitWatcher) getGitStatus() *types.GitInfo {
	status := &types.GitInfo{}

	// 检查是否是 Git 仓库
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = w.workPath
	output, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) != "true" {
		return status
	}

	status.IsRepo = true

	// 获取当前分支
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = w.workPath
	output, err = cmd.Output()
	if err == nil {
		status.Branch = strings.TrimSpace(string(output))
	}

	// 获取修改统计
	cmd = exec.Command("git", "diff", "--numstat")
	cmd.Dir = w.workPath
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			status.ModifiedFiles++
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				if parts[0] != "-" {
					var added int
					fmt.Sscanf(parts[0], "%d", &added)
					status.AddedLines += added
				}
				if parts[1] != "-" {
					var deleted int
					fmt.Sscanf(parts[1], "%d", &deleted)
					status.DeletedLines += deleted
				}
			}
		}
	}

	// 获取暂存区修改统计
	cmd = exec.Command("git", "diff", "--cached", "--numstat")
	cmd.Dir = w.workPath
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			status.ModifiedFiles++
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				if parts[0] != "-" {
					var added int
					fmt.Sscanf(parts[0], "%d", &added)
					status.AddedLines += added
				}
				if parts[1] != "-" {
					var deleted int
					fmt.Sscanf(parts[1], "%d", &deleted)
					status.DeletedLines += deleted
				}
			}
		}
	}

	return status
}

// UpdateWorkPath 更新工作目录（切换目录时调用）
func (w *GitWatcher) UpdateWorkPath(newWorkPath string) (*types.GitInfo, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 如果目录没变，不需要重新监听
	if w.workPath == newWorkPath {
		return w.getGitStatus(), nil
	}

	// 停止之前的监听
	w.stopLocked()

	// 重新启动监听（使用已有的 sessionID）
	return w.Start(w.sessionID, newWorkPath)
}

// Stop 停止监听
func (w *GitWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stopLocked()
}

// stopLocked 停止监听（需要持有锁）
func (w *GitWatcher) stopLocked() {
	// 取消 context
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}

	// 取消防抖定时器
	w.debounceMux.Lock()
	if w.debounce != nil {
		w.debounce.Stop()
		w.debounce = nil
	}
	w.debounceMux.Unlock()

	// 关闭 watcher
	if w.watcher != nil {
		w.watcher.Close()
		w.watcher = nil
	}

	w.gitDir = ""
}
