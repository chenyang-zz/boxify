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

package service

import (
	"context"
	"time"

	"github.com/chenyang-zz/boxify/internal/events"
	gitcore "github.com/chenyang-zz/boxify/internal/git"
	boxtypes "github.com/chenyang-zz/boxify/internal/types"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// GitService 仅提供前端可调用接口，核心逻辑在 internal/git。
type GitService struct {
	BaseService
	manager *gitcore.Manager
}

// NewGitService 创建 Git 服务
func NewGitService(deps *ServiceDeps) *GitService {
	return &GitService{BaseService: NewBaseService(deps)}
}

// ServiceStartup 服务启动
func (g *GitService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	g.SetContext(ctx)
	g.manager = gitcore.NewManager(ctx, g.Logger(), func(event boxtypes.GitStatusChangedEvent) {
		g.Logger().Info("Git 状态变化事件", "repoKey", event.RepoKey, "status", event.Status, "timestamp", event.Timestamp)
		g.App().Event.Emit(string(events.EventTypeGitStatusChanged), boxtypes.GitStatusChangedEvent{
			RepoKey:   event.RepoKey,
			Status:    event.Status,
			Timestamp: event.Timestamp,
		})
	})
	g.Logger().Info("服务启动", "service", "GitService")
	return nil
}

// ServiceShutdown 服务关闭
func (g *GitService) ServiceShutdown() error {
	if g.manager != nil {
		g.manager.Shutdown()
	}
	g.Logger().Info("服务关闭", "service", "GitService")
	return nil
}

// RegisterRepo 注册仓库到管理器 map。
func (g *GitService) RegisterRepo(repoKey, path string) *boxtypes.GitRepoInfoResult {
	info, err := g.manager.RegisterRepo(repoKey, path)
	if err != nil {
		return &boxtypes.GitRepoInfoResult{
			BaseResult: boxtypes.BaseResult{Success: false, Message: err.Error()},
		}
	}
	return &boxtypes.GitRepoInfoResult{
		BaseResult: boxtypes.BaseResult{Success: true, Message: "仓库注册成功"},
		Data:       info,
	}
}

// RemoveRepo 从管理器 map 中移除仓库。
func (g *GitService) RemoveRepo(repoKey string) *boxtypes.GitRemoveRepoResult {
	err := g.manager.RemoveRepo(repoKey)
	if err != nil {
		return &boxtypes.GitRemoveRepoResult{
			BaseResult: boxtypes.BaseResult{Success: false, Message: err.Error()},
		}
	}
	return &boxtypes.GitRemoveRepoResult{
		BaseResult: boxtypes.BaseResult{Success: true, Message: "仓库移除成功"},
		Data:       &boxtypes.GitRemoveRepoData{RepoKey: repoKey},
	}
}

// ListRepos 获取所有已注册仓库。
func (g *GitService) ListRepos() *boxtypes.GitRepoListResult {
	return &boxtypes.GitRepoListResult{
		BaseResult: boxtypes.BaseResult{Success: true, Message: "获取仓库列表成功"},
		Data:       &boxtypes.GitRepoListData{Repos: g.manager.ListRepos()},
	}
}

// GetRepoStatus 获取指定仓库状态。
func (g *GitService) GetRepoStatus(repoKey string) *boxtypes.GitRepoStatusResult {
	status, err := g.manager.GetStatus(repoKey)
	if err != nil {
		return &boxtypes.GitRepoStatusResult{
			BaseResult: boxtypes.BaseResult{Success: false, Message: err.Error()},
		}
	}
	return &boxtypes.GitRepoStatusResult{
		BaseResult: boxtypes.BaseResult{Success: true, Message: "获取 Git 状态成功"},
		Data:       &boxtypes.GitRepoStatusData{Status: status},
	}
}

// StartRepoWatch 启动指定仓库监听。
func (g *GitService) StartRepoWatch(repoKey string, intervalMs int) *boxtypes.GitRepoInfoResult {
	info, err := g.manager.StartWatch(repoKey, intervalMs)
	if err != nil {
		return &boxtypes.GitRepoInfoResult{
			BaseResult: boxtypes.BaseResult{Success: false, Message: err.Error()},
		}
	}
	return &boxtypes.GitRepoInfoResult{
		BaseResult: boxtypes.BaseResult{Success: true, Message: "仓库监听已启动"},
		Data:       info,
	}
}

// StopRepoWatch 停止指定仓库监听。
func (g *GitService) StopRepoWatch(repoKey string) *boxtypes.GitRepoInfoResult {
	info, err := g.manager.StopWatch(repoKey)
	if err != nil {
		return &boxtypes.GitRepoInfoResult{
			BaseResult: boxtypes.BaseResult{Success: false, Message: err.Error()},
		}
	}
	return &boxtypes.GitRepoInfoResult{
		BaseResult: boxtypes.BaseResult{Success: true, Message: "仓库监听已停止"},
		Data:       info,
	}
}

// StopAllRepoWatch 停止全部仓库监听。
func (g *GitService) StopAllRepoWatch() *boxtypes.GitStopAllRepoWatchResult {
	count := g.manager.StopAllWatches()
	return &boxtypes.GitStopAllRepoWatchResult{
		BaseResult: boxtypes.BaseResult{Success: true, Message: "全部仓库监听已停止"},
		Data:       &boxtypes.GitStopAllRepoWatchData{StoppedCount: count},
	}
}

// SetActiveRepo 设置激活仓库。autoStart=true 时自动开启该仓库监听，stopOthers=true 时停止其他仓库监听。
func (g *GitService) SetActiveRepo(repoKey string, autoStart, stopOthers bool) *boxtypes.GitRepoInfoResult {
	info, err := g.manager.SetActiveRepo(repoKey, autoStart, stopOthers)
	if err != nil {
		return &boxtypes.GitRepoInfoResult{
			BaseResult: boxtypes.BaseResult{Success: false, Message: err.Error()},
		}
	}
	return &boxtypes.GitRepoInfoResult{
		BaseResult: boxtypes.BaseResult{Success: true, Message: "激活仓库设置成功"},
		Data:       info,
	}
}

// GetActiveRepo 获取当前激活仓库 key。
func (g *GitService) GetActiveRepo() *boxtypes.GitActiveRepoResult {
	return &boxtypes.GitActiveRepoResult{
		BaseResult: boxtypes.BaseResult{Success: true, Message: "获取激活仓库成功"},
		Data:       &boxtypes.GitActiveRepoData{RepoKey: g.manager.ActiveRepoKey()},
	}
}

// GetInitialStatusEvent 首次手动获取 Git 状态事件数据（不依赖监听回调）。
// repoKey 为空时使用当前激活仓库。
func (g *GitService) GetInitialStatusEvent(repoKey string) *boxtypes.GitStatusEventResult {
	key := repoKey
	if key == "" {
		key = g.manager.ActiveRepoKey()
	}
	if key == "" {
		return &boxtypes.GitStatusEventResult{
			BaseResult: boxtypes.BaseResult{Success: false, Message: "未指定 repoKey 且当前无激活仓库"},
		}
	}

	status, err := g.manager.GetStatus(key)
	if err != nil {
		return &boxtypes.GitStatusEventResult{
			BaseResult: boxtypes.BaseResult{Success: false, Message: err.Error()},
		}
	}

	event := boxtypes.GitStatusChangedEvent{
		RepoKey:   key,
		Status:    *status,
		Timestamp: time.Now().Unix(),
	}

	return &boxtypes.GitStatusEventResult{
		BaseResult: boxtypes.BaseResult{Success: true, Message: "获取首次 Git 状态事件成功"},
		Data:       &boxtypes.GitStatusEventData{Event: &event},
	}
}
