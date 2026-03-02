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

package types

// GitFileStatus Git 文件状态（与 porcelain v2 对齐）。
type GitFileStatus struct {
	Path           string `json:"path"`                   // 文件路径
	OriginalPath   string `json:"originalPath,omitempty"` // 原始路径（重命名场景）
	IndexStatus    string `json:"indexStatus"`            // 暂存区状态
	WorkTreeStatus string `json:"workTreeStatus"`         // 工作区状态
	Kind           string `json:"kind"`                   // 变更类型: changed, renamed, untracked, unmerged
}

// GitRepoStatus Git 仓库状态。
type GitRepoStatus struct {
	RepositoryRoot string          `json:"repositoryRoot"`     // 仓库根目录
	CurrentPath    string          `json:"currentPath"`        // 当前工作路径
	Head           string          `json:"head"`               // 当前分支/HEAD
	Upstream       string          `json:"upstream,omitempty"` // 上游分支
	Oid            string          `json:"oid,omitempty"`      // HEAD 提交哈希
	Ahead          int             `json:"ahead"`              // 领先上游提交数
	Behind         int             `json:"behind"`             // 落后上游提交数
	Detached       bool            `json:"detached"`           // 是否为 detached HEAD
	IsClean        bool            `json:"isClean"`            // 是否工作区干净
	StagedCount    int             `json:"stagedCount"`        // 暂存区变更文件数
	UnstagedCount  int             `json:"unstagedCount"`      // 工作区未暂存变更文件数
	UntrackedCount int             `json:"untrackedCount"`     // 未跟踪文件数
	ConflictCount  int             `json:"conflictCount"`      // 冲突文件数
	AddedLines     int             `json:"addedLines"`         // 新增代码行数（暂存+未暂存）
	DeletedLines   int             `json:"deletedLines"`       // 删除代码行数（暂存+未暂存）
	Files          []GitFileStatus `json:"files"`              // 文件级状态明细
	UpdatedAt      int64           `json:"updatedAt"`          // 状态更新时间（Unix 秒）
}

// GitStatusChangedEvent 仓库状态变化事件。
type GitStatusChangedEvent struct {
	RepoKey   string        `json:"repoKey"`   // 仓库唯一标识
	Status    GitRepoStatus `json:"status"`    // 最新仓库状态
	Timestamp int64         `json:"timestamp"` // 事件时间（Unix 秒）
}

// GitRepoInfo Git 管理器中的仓库信息。
type GitRepoInfo struct {
	RepoKey      string `json:"repoKey"`             // 仓库唯一标识
	Path         string `json:"path"`                // 注册时路径
	RepoRoot     string `json:"repoRoot"`            // 仓库根目录
	GitDir       string `json:"gitDir"`              // Git 元数据目录
	Watching     bool   `json:"watching"`            // 是否正在监听
	Active       bool   `json:"active"`              // 是否激活仓库
	IntervalMs   int64  `json:"intervalMs"`          // 兜底轮询间隔（毫秒）
	LastError    string `json:"lastError,omitempty"` // 最近一次监听/采集错误
	RegisteredAt int64  `json:"registeredAt"`        // 注册时间（Unix 秒）
}

// GitRepoInfoResult Git 仓库信息返回结果。
type GitRepoInfoResult struct {
	BaseResult
	Data *GitRepoInfo `json:"data,omitempty"` // 仓库信息
}

// GitRepoListData Git 仓库列表数据。
type GitRepoListData struct {
	Repos []GitRepoInfo `json:"repos"` // 仓库列表
}

// GitRepoListResult Git 仓库列表返回结果。
type GitRepoListResult struct {
	BaseResult
	Data *GitRepoListData `json:"data,omitempty"` // 仓库列表数据
}

// GitRepoStatusData Git 仓库状态数据。
type GitRepoStatusData struct {
	Status *GitRepoStatus `json:"status,omitempty"` // 仓库状态
}

// GitRepoStatusResult Git 仓库状态返回结果。
type GitRepoStatusResult struct {
	BaseResult
	Data *GitRepoStatusData `json:"data,omitempty"` // 仓库状态数据
}

// GitActiveRepoData 当前激活仓库数据。
type GitActiveRepoData struct {
	RepoKey string `json:"repoKey"` // 当前激活仓库 key
}

// GitActiveRepoResult 当前激活仓库返回结果。
type GitActiveRepoResult struct {
	BaseResult
	Data *GitActiveRepoData `json:"data,omitempty"` // 激活仓库数据
}

// GitStatusEventData 首次状态事件数据。
type GitStatusEventData struct {
	Event *GitStatusChangedEvent `json:"event,omitempty"` // Git 状态事件
}

// GitStatusEventResult 首次状态事件返回结果。
type GitStatusEventResult struct {
	BaseResult
	Data *GitStatusEventData `json:"data,omitempty"` // 状态事件数据
}

// GitRemoveRepoData 移除仓库结果数据。
type GitRemoveRepoData struct {
	RepoKey string `json:"repoKey"` // 已移除仓库 key
}

// GitRemoveRepoResult 移除仓库返回结果。
type GitRemoveRepoResult struct {
	BaseResult
	Data *GitRemoveRepoData `json:"data,omitempty"` // 移除结果数据
}

// GitStopAllRepoWatchData 停止全部监听结果数据。
type GitStopAllRepoWatchData struct {
	StoppedCount int `json:"stoppedCount"` // 停止监听的仓库数量
}

// GitStopAllRepoWatchResult 停止全部监听返回结果。
type GitStopAllRepoWatchResult struct {
	BaseResult
	Data *GitStopAllRepoWatchData `json:"data,omitempty"` // 停止结果数据
}
