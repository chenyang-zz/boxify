package git

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
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
	status.AddedLines, status.DeletedLines = c.collectLineStats(ctx, repoRoot)
	status.UpdatedAt = time.Now().Unix()
	status.IsClean = status.StagedCount == 0 && status.UnstagedCount == 0 && status.UntrackedCount == 0 && status.ConflictCount == 0
	return status, nil
}

// collectLineStats 汇总暂存区与工作区的新增/删除行数。
func (c *StatusCollector) collectLineStats(ctx context.Context, repoRoot string) (int, int) {
	totalAdded, totalDeleted := 0, 0
	commands := [][]string{
		{"diff", "--numstat"},
		{"diff", "--cached", "--numstat"},
	}

	for _, args := range commands {
		out, err := c.runner.Run(ctx, repoRoot, args...)
		if err != nil {
			c.logger.Debug("采集 Git 行数统计失败", "repo", repoRoot, "args", strings.Join(args, " "), "error", err)
			continue
		}
		added, deleted := parseNumstat(out)
		totalAdded += added
		totalDeleted += deleted
	}

	return totalAdded, totalDeleted
}

// parseNumstat 解析 git diff --numstat 输出。
func parseNumstat(output string) (int, int) {
	added, deleted := 0, 0
	for _, raw := range strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		if parts[0] != "-" {
			if n, err := strconv.Atoi(parts[0]); err == nil {
				added += n
			}
		}
		if parts[1] != "-" {
			if n, err := strconv.Atoi(parts[1]); err == nil {
				deleted += n
			}
		}
	}
	return added, deleted
}
