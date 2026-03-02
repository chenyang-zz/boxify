package git

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

// CommandRunner 负责执行 Git 命令。
type CommandRunner struct {
	timeout time.Duration
	logger  *slog.Logger
}

// NewCommandRunner 创建命令执行器。
func NewCommandRunner(timeout time.Duration, logger *slog.Logger) *CommandRunner {
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &CommandRunner{
		timeout: timeout,
		logger:  logger,
	}
}

// Run 在指定目录执行 git 子命令。
func (r *CommandRunner) Run(ctx context.Context, dir string, args ...string) (string, error) {
	runCtx := ctx
	if runCtx == nil {
		runCtx = context.Background()
	}

	cmdCtx, cancel := context.WithTimeout(runCtx, r.timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}
	return string(out), nil
}
