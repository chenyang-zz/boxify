package git

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	boxtypes "github.com/chenyang-zz/boxify/internal/types"
)

// StatusCollector 负责采集并组装仓库状态。
type StatusCollector struct {
	runner   *CommandRunner
	resolver *Resolver
	parser   *StatusParser
	logger   *slog.Logger
}

// NewStatusCollector 创建状态采集器。
func NewStatusCollector(runner *CommandRunner, resolver *Resolver, parser *StatusParser, logger *slog.Logger) *StatusCollector {
	if logger == nil {
		logger = slog.Default()
	}
	return &StatusCollector{
		runner:   runner,
		resolver: resolver,
		parser:   parser,
		logger:   logger,
	}
}

// CollectByPath 通过任意路径采集仓库状态。
func (c *StatusCollector) CollectByPath(ctx context.Context, path string) (*boxtypes.GitRepoStatus, *RepoLocation, error) {
	location, err := c.resolver.Resolve(ctx, path)
	if err != nil {
		return nil, nil, err
	}

	status, err := c.CollectByRepoRoot(ctx, location.Path, location.RepoRoot)
	if err != nil {
		return nil, nil, err
	}

	return status, location, nil
}

// CollectByRepoRoot 通过仓库根路径采集状态。
func (c *StatusCollector) CollectByRepoRoot(ctx context.Context, currentPath, repoRoot string) (*boxtypes.GitRepoStatus, error) {
	out, err := c.runner.Run(ctx, repoRoot, "status", "--porcelain=v2", "--branch")
	if err != nil {
		return nil, fmt.Errorf("执行 git status 失败: %w", err)
	}

	status, err := c.parser.ParsePorcelainV2(strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n"))
	if err != nil {
		return nil, err
	}

	status.RepositoryRoot = repoRoot
	status.CurrentPath = currentPath
	status.UpdatedAt = time.Now().Unix()
	status.IsClean = status.StagedCount == 0 && status.UnstagedCount == 0 && status.UntrackedCount == 0 && status.ConflictCount == 0
	return status, nil
}
