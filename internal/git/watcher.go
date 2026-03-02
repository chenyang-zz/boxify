package git

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	boxtypes "github.com/chenyang-zz/boxify/internal/types"
	"github.com/fsnotify/fsnotify"
)

var gitHotPathCandidates = []string{
	"HEAD",
	"index",
	"refs",
	"logs",
	"packed-refs",
}

// RepositoryWatcher 负责单仓库监听与状态增量推送。
type RepositoryWatcher struct {
	key       string
	path      string
	repoRoot  string
	gitDir    string
	interval  time.Duration
	logger    *slog.Logger
	onStatus  func(boxtypes.GitStatusChangedEvent)
	collector *StatusCollector
	parentCtx context.Context

	mu           sync.Mutex
	cancel       context.CancelFunc
	lastSnapshot string
	lastError    string
}

// NewRepositoryWatcher 创建单仓库监听器。
func NewRepositoryWatcher(
	ctx context.Context,
	logger *slog.Logger,
	collector *StatusCollector,
	key, path, repoRoot, gitDir string,
	interval time.Duration,
	onStatus func(boxtypes.GitStatusChangedEvent),
) *RepositoryWatcher {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = slog.Default()
	}
	if interval <= 0 {
		interval = defaultWatchInterval
	}

	return &RepositoryWatcher{
		key:       key,
		path:      path,
		repoRoot:  repoRoot,
		gitDir:    gitDir,
		interval:  interval,
		logger:    logger,
		onStatus:  onStatus,
		collector: collector,
		parentCtx: ctx,
	}
}

// IsRunning 返回监听器是否处于运行状态。
func (w *RepositoryWatcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.cancel != nil
}

// LastError 返回最近一次监听/采集错误。
func (w *RepositoryWatcher) LastError() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastError
}

// Start 启动监听。
func (w *RepositoryWatcher) Start(interval time.Duration) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if interval > 0 {
		w.interval = interval
	}
	if w.cancel != nil {
		return nil
	}

	ctx, cancel := context.WithCancel(w.parentCtx)
	w.cancel = cancel
	w.lastSnapshot = ""
	go w.watchLoop(ctx)

	w.logger.Info("Git 仓库监听启动", "repoKey", w.key, "repo", w.repoRoot, "interval", w.interval.String())
	return nil
}

// Stop 停止监听。
func (w *RepositoryWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		w.cancel()
	}
	w.cancel = nil
	w.lastSnapshot = ""
	w.logger.Info("Git 仓库监听停止", "repoKey", w.key, "repo", w.repoRoot, "interval", w.interval.String())
}

// CollectStatus 主动采集一次当前仓库状态。
func (w *RepositoryWatcher) CollectStatus(ctx context.Context) (*boxtypes.GitRepoStatus, error) {
	return w.collector.CollectByRepoRoot(ctx, w.path, w.repoRoot)
}

// setLastError 记录最近一次错误文本。
func (w *RepositoryWatcher) setLastError(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err == nil {
		w.lastError = ""
		return
	}
	w.lastError = err.Error()
}

// watchLoop 运行 fsnotify 监听主循环，并在必要时触发状态刷新。
func (w *RepositoryWatcher) watchLoop(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.setLastError(err)
		w.logger.Error("创建 fsnotify watcher 失败，降级为轮询", "repoKey", w.key, "repo", w.repoRoot, "error", err)
		w.pollingLoop(ctx)
		return
	}
	defer watcher.Close()

	if err := w.addRepoWatchPaths(watcher); err != nil {
		w.setLastError(err)
		w.logger.Warn("添加监听目录部分失败", "repoKey", w.key, "repo", w.repoRoot, "error", err)
	}

	w.emitLatest(ctx, "startup")

	fallbackTicker := time.NewTicker(w.interval)
	debounceTicker := time.NewTicker(debounceInterval)
	defer fallbackTicker.Stop()
	defer debounceTicker.Stop()

	pendingRefresh := false

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename|fsnotify.Chmod) == 0 {
				continue
			}
			if event.Op&fsnotify.Create != 0 {
				w.tryAddDynamicWatch(watcher, event.Name)
			}
			pendingRefresh = true
		case err := <-watcher.Errors:
			if err != nil {
				w.setLastError(err)
				w.logger.Debug("fsnotify 监听错误", "repoKey", w.key, "repo", w.repoRoot, "error", err)
			}
		case <-debounceTicker.C:
			if pendingRefresh {
				pendingRefresh = false
				w.emitLatest(ctx, "fsnotify")
			}
		case <-fallbackTicker.C:
			w.emitLatest(ctx, "fallback")
		}
	}
}

// pollingLoop 在 fsnotify 不可用时使用轮询兜底监听。
func (w *RepositoryWatcher) pollingLoop(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.emitLatest(ctx, "startup")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.emitLatest(ctx, "fallback")
		}
	}
}

// emitLatest 采集最新状态并在有变化时发送事件。
func (w *RepositoryWatcher) emitLatest(ctx context.Context, trigger string) {
	status, err := w.CollectStatus(ctx)
	if err != nil {
		w.setLastError(err)
		w.logger.Debug("Git 状态采集失败", "repoKey", w.key, "repo", w.repoRoot, "trigger", trigger, "error", err)
		return
	}

	w.setLastError(nil)
	snapshot := w.snapshot(*status)
	if !w.shouldEmit(snapshot) {
		return
	}

	if w.onStatus != nil {
		w.onStatus(boxtypes.GitStatusChangedEvent{
			RepoKey:   w.key,
			Status:    *status,
			Timestamp: time.Now().Unix(),
		})
	}
}

// shouldEmit 判断本次快照是否需要对外发送。
func (w *RepositoryWatcher) shouldEmit(snapshot string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if snapshot == w.lastSnapshot {
		return false
	}
	w.lastSnapshot = snapshot
	return true
}

// snapshot 生成用于去重的状态快照摘要。
func (w *RepositoryWatcher) snapshot(status boxtypes.GitRepoStatus) string {
	var b strings.Builder
	b.WriteString(status.RepositoryRoot)
	b.WriteString("|")
	b.WriteString(status.Head)
	b.WriteString("|")
	b.WriteString(status.Upstream)
	b.WriteString("|")
	b.WriteString(strconv.Itoa(status.Ahead))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(status.Behind))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(status.StagedCount))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(status.UnstagedCount))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(status.UntrackedCount))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(status.ConflictCount))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(status.AddedLines))
	b.WriteString("|")
	b.WriteString(strconv.Itoa(status.DeletedLines))
	for _, file := range status.Files {
		b.WriteString("|")
		b.WriteString(file.Kind)
		b.WriteString(":")
		b.WriteString(file.IndexStatus)
		b.WriteString(file.WorkTreeStatus)
		b.WriteString(":")
		b.WriteString(file.Path)
		b.WriteString(":")
		b.WriteString(file.OriginalPath)
	}

	h := sha1.Sum([]byte(b.String()))
	return hex.EncodeToString(h[:])
}

// addRepoWatchPaths 注册仓库工作区与 .git 目录监听。
func (w *RepositoryWatcher) addRepoWatchPaths(watcher *fsnotify.Watcher) error {
	var errList []string
	seen := make(map[string]struct{})

	// 只监听工作区根目录而不是整棵目录树：
	// 这样可以在目录规模较大时显著降低 watcher/FD 消耗。
	// 工作区内深层变化依赖轮询兜底，以及新建目录时的动态补充监听。
	if err := w.addWatchPath(watcher, seen, w.repoRoot); err != nil {
		errList = append(errList, "workspaceRoot: "+err.Error())
	}

	// .git 是状态变化高频源，这里监听关键路径而不是完整递归树。
	// refs/logs 作为目录监听，HEAD/index/packed-refs 作为文件监听。
	for _, rel := range gitHotPathCandidates {
		candidate := filepath.Join(w.gitDir, rel)
		if err := w.addWatchPath(watcher, seen, candidate); err != nil {
			errList = append(errList, ".git/"+rel+": "+err.Error())
		}
	}

	if len(errList) > 0 {
		return errors.New(strings.Join(errList, "; "))
	}
	return nil
}

// addWatchPath 将单个文件或目录添加到监听列表。
// 如果路径不存在则忽略，避免在不同 git 实现/平台差异下导致启动失败。
func (w *RepositoryWatcher) addWatchPath(watcher *fsnotify.Watcher, seen map[string]struct{}, path string) error {
	clean := filepath.Clean(path)
	if _, ok := seen[clean]; ok {
		return nil
	}

	if _, err := os.Stat(clean); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := watcher.Add(clean); err != nil {
		return err
	}
	seen[clean] = struct{}{}
	return nil
}

// tryAddDynamicWatch 为新创建目录动态补充监听。
func (w *RepositoryWatcher) tryAddDynamicWatch(watcher *fsnotify.Watcher, createdPath string) {
	fi, err := os.Stat(createdPath)
	if err != nil || !fi.IsDir() {
		return
	}

	cleanPath := filepath.Clean(createdPath)
	gitPath := filepath.Join(w.repoRoot, ".git")
	if strings.HasPrefix(cleanPath, filepath.Clean(gitPath)+string(os.PathSeparator)) || cleanPath == filepath.Clean(gitPath) {
		return
	}

	// 动态目录只添加当前目录，不递归添加整棵子树，避免单次创建触发大量 watcher。
	// 后续子目录创建事件会继续触发本函数补充监听。
	_ = watcher.Add(cleanPath)
}
