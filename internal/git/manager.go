package git

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	boxtypes "github.com/chenyang-zz/boxify/internal/types"
)

const (
	defaultWatchInterval = 2 * time.Second
	debounceInterval     = 180 * time.Millisecond
)

// repoEntry 是管理器内部维护的仓库条目。
type repoEntry struct {
	key          string
	path         string
	repoRoot     string
	gitDir       string
	registeredAt time.Time
	watcher      *RepositoryWatcher
	interval     time.Duration
	logger       *slog.Logger
}

// Manager 维护多仓库监听 map，并支持激活仓库切换。
type Manager struct {
	mu              sync.RWMutex
	ctx             context.Context
	logger          *slog.Logger
	repos           map[string]*repoEntry
	activeRepoKey   string
	defaultInterval time.Duration
	onStatus        func(boxtypes.GitStatusChangedEvent)

	runner    *CommandRunner
	resolver  *Resolver
	parser    *StatusParser
	collector *StatusCollector
}

// NewManager 创建 Git 仓库管理器。
func NewManager(ctx context.Context, logger *slog.Logger, onStatus func(boxtypes.GitStatusChangedEvent)) *Manager {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = slog.Default()
	}

	runner := NewCommandRunner(4*time.Second, logger)
	resolver := NewResolver(runner, logger)
	parser := NewStatusParser(logger)
	collector := NewStatusCollector(runner, resolver, parser, logger)

	return &Manager{
		ctx:             ctx,
		logger:          logger,
		repos:           make(map[string]*repoEntry),
		defaultInterval: defaultWatchInterval,
		onStatus:        onStatus,
		runner:          runner,
		resolver:        resolver,
		parser:          parser,
		collector:       collector,
	}
}

// Shutdown 停止所有仓库监听。
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.repos {
		entry.watcher.Stop()
	}
}

// Probe 对任意路径做一次状态探测，不注册到管理器。
func (m *Manager) Probe(path string) (*boxtypes.GitRepoStatus, error) {
	status, _, err := m.collector.CollectByPath(m.ctx, path)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// RegisterRepo 注册仓库到管理器 map。
func (m *Manager) RegisterRepo(repoKey, path string) (*boxtypes.GitRepoInfo, error) {
	location, err := m.resolver.Resolve(m.ctx, path)
	if err != nil {
		return nil, err
	}
	if repoKey == "" {
		repoKey = location.RepoRoot
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	old, exists := m.repos[repoKey]
	interval := m.defaultInterval
	if exists {
		interval = old.interval
		if old.watcher.IsRunning() {
			old.watcher.Stop()
		}
	}

	entry := &repoEntry{
		key:          repoKey,
		path:         location.Path,
		repoRoot:     location.RepoRoot,
		gitDir:       location.GitDir,
		registeredAt: time.Now(),
		interval:     interval,
		logger:       m.logger,
	}
	entry.watcher = NewRepositoryWatcher(m.ctx, m.logger, m.collector, repoKey, location.Path, location.RepoRoot, location.GitDir, interval, m.onStatus)

	m.repos[repoKey] = entry
	info := m.repoToInfo(entry)
	return &info, nil
}

// RemoveRepo 从管理器中移除仓库。
func (m *Manager) RemoveRepo(repoKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.repos[repoKey]
	if !ok {
		return fmt.Errorf("仓库未注册: %s", repoKey)
	}
	entry.watcher.Stop()
	delete(m.repos, repoKey)
	if m.activeRepoKey == repoKey {
		m.activeRepoKey = ""
	}
	return nil
}

// StartWatch 启动指定仓库监听。
func (m *Manager) StartWatch(repoKey string, intervalMs int) (*boxtypes.GitRepoInfo, error) {
	m.mu.Lock()
	entry, ok := m.repos[repoKey]
	if !ok {
		m.mu.Unlock()
		return nil, fmt.Errorf("仓库未注册: %s", repoKey)
	}

	interval := entry.interval
	if intervalMs > 0 {
		interval = time.Duration(intervalMs) * time.Millisecond
		if interval < 800*time.Millisecond {
			interval = 800 * time.Millisecond
		}
		entry.interval = interval
	}
	m.mu.Unlock()

	if err := entry.watcher.Start(interval); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	info := m.repoToInfo(entry)
	return &info, nil
}

// StopWatch 停止指定仓库监听。
func (m *Manager) StopWatch(repoKey string) (*boxtypes.GitRepoInfo, error) {
	m.mu.RLock()
	entry, ok := m.repos[repoKey]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("仓库未注册: %s", repoKey)
	}
	entry.watcher.Stop()

	m.mu.RLock()
	defer m.mu.RUnlock()
	info := m.repoToInfo(entry)
	return &info, nil
}

// StopAllWatches 停止所有仓库监听，返回停止数量。
func (m *Manager) StopAllWatches() int {
	m.mu.RLock()
	entries := make([]*repoEntry, 0, len(m.repos))
	for _, entry := range m.repos {
		entries = append(entries, entry)
	}
	m.mu.RUnlock()

	for _, entry := range entries {
		entry.watcher.Stop()
	}

	return len(entries)
}

// SetActiveRepo 设置当前激活仓库。
func (m *Manager) SetActiveRepo(repoKey string, autoStart bool, stopOthers bool) (*boxtypes.GitRepoInfo, error) {
	m.mu.Lock()
	_, ok := m.repos[repoKey]
	if !ok {
		m.mu.Unlock()
		return nil, fmt.Errorf("仓库未注册: %s", repoKey)
	}

	var otherEntries []*repoEntry
	if stopOthers {
		for key, repo := range m.repos {
			if key != repoKey {
				otherEntries = append(otherEntries, repo)
			}
		}
	}
	m.mu.Unlock()

	if autoStart {
		if _, err := m.StartWatch(repoKey, 0); err != nil {
			return nil, err
		}
	}
	if stopOthers {
		for _, repo := range otherEntries {
			repo.watcher.Stop()
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.repos[repoKey]
	if !ok {
		return nil, fmt.Errorf("仓库未注册: %s", repoKey)
	}
	m.activeRepoKey = repoKey
	info := m.repoToInfo(entry)
	return &info, nil
}

// ActiveRepoKey 返回当前激活仓库 key。
func (m *Manager) ActiveRepoKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeRepoKey
}

// GetStatus 获取指定仓库当前状态。
func (m *Manager) GetStatus(repoKey string) (*boxtypes.GitRepoStatus, error) {
	m.mu.RLock()
	entry, ok := m.repos[repoKey]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("仓库未注册: %s", repoKey)
	}
	return entry.watcher.CollectStatus(m.ctx)
}

// GetRepoInfo 获取单仓库元信息。
func (m *Manager) GetRepoInfo(repoKey string) (*boxtypes.GitRepoInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.repos[repoKey]
	if !ok {
		return nil, fmt.Errorf("仓库未注册: %s", repoKey)
	}
	info := m.repoToInfo(entry)
	return &info, nil
}

// ListRepos 获取所有已注册仓库。
func (m *Manager) ListRepos() []boxtypes.GitRepoInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]boxtypes.GitRepoInfo, 0, len(m.repos))
	for _, entry := range m.repos {
		result = append(result, m.repoToInfo(entry))
	}
	return result
}

// repoToInfo 将内部仓库条目转换为对外信息结构。
func (m *Manager) repoToInfo(entry *repoEntry) boxtypes.GitRepoInfo {
	return boxtypes.GitRepoInfo{
		RepoKey:      entry.key,
		Path:         entry.path,
		RepoRoot:     entry.repoRoot,
		GitDir:       entry.gitDir,
		Watching:     entry.watcher.IsRunning(),
		Active:       m.activeRepoKey == entry.key,
		IntervalMs:   entry.interval.Milliseconds(),
		LastError:    entry.watcher.LastError(),
		RegisteredAt: entry.registeredAt.Unix(),
	}
}
