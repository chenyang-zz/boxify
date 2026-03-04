package git

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// RepoLocation 表示路径解析出的仓库定位信息。
type RepoLocation struct {
	Path     string
	RepoRoot string
	GitDir   string
	logger   *slog.Logger
}

// Resolver 负责将任意路径解析为仓库定位信息。
type Resolver struct {
	runner *CommandRunner
	logger *slog.Logger
}

// NewResolver 创建路径解析器。
func NewResolver(runner *CommandRunner, logger *slog.Logger) *Resolver {
	if logger == nil {
		logger = slog.Default()
	}
	return &Resolver{
		runner: runner,
		logger: logger,
	}
}

// Resolve 解析路径并返回仓库信息。
func (r *Resolver) Resolve(ctx context.Context, path string) (*RepoLocation, error) {
	normalizedPath, err := r.normalizePath(path)
	if err != nil {
		return nil, err
	}

	out, err := r.runner.Run(ctx, normalizedPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("路径不在 Git 仓库中: %s", normalizedPath)
	}
	repoRoot := strings.TrimSpace(out)

	out, err = r.runner.Run(ctx, repoRoot, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return nil, fmt.Errorf("获取 Git 元数据目录失败: %w", err)
	}

	return &RepoLocation{
		Path:     normalizedPath,
		RepoRoot: repoRoot,
		GitDir:   strings.TrimSpace(out),
		logger:   r.logger,
	}, nil
}

// normalizePath 规范化输入路径，确保返回目录路径。
func (r *Resolver) normalizePath(path string) (string, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("获取当前目录失败: %w", err)
		}
		p = cwd
	}

	p, err := expandUserHome(p)
	if err != nil {
		return "", err
	}

	p = filepath.Clean(p)
	fi, err := os.Stat(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("路径不存在: %s", p)
		}
		return "", fmt.Errorf("无法访问路径 %s: %w", p, err)
	}
	if !fi.IsDir() {
		p = filepath.Dir(p)
	}
	return p, nil
}

func expandUserHome(path string) (string, error) {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("获取用户目录失败: %w", err)
		}
		return home, nil
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("获取用户目录失败: %w", err)
		}
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}
