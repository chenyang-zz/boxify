package git

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func canonicalPath(t *testing.T, path string) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(p)
}

func TestResolverResolveSuccess(t *testing.T) {
	repoDir := mustInitGitRepo(t)
	subDir := filepath.Join(repoDir, "sub", "nested")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	runner := NewCommandRunner(0, slog.Default())
	resolver := NewResolver(runner, slog.Default())

	location, err := resolver.Resolve(context.Background(), subDir)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if canonicalPath(t, location.RepoRoot) != canonicalPath(t, repoDir) {
		t.Fatalf("unexpected repo root: %s", location.RepoRoot)
	}
	if canonicalPath(t, location.Path) != canonicalPath(t, subDir) {
		t.Fatalf("unexpected path: %s", location.Path)
	}
	if location.GitDir == "" {
		t.Fatal("git dir should not be empty")
	}
}

func TestStatusCollectorCollectByPath(t *testing.T) {
	repoDir := mustInitGitRepo(t)
	filePath := filepath.Join(repoDir, "untracked.txt")
	if err := os.WriteFile(filePath, []byte("data\n"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	runner := NewCommandRunner(0, slog.Default())
	resolver := NewResolver(runner, slog.Default())
	parser := NewStatusParser(slog.Default())
	collector := NewStatusCollector(runner, resolver, parser, slog.Default())

	status, location, err := collector.CollectByPath(context.Background(), repoDir)
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}

	if canonicalPath(t, location.RepoRoot) != canonicalPath(t, repoDir) {
		t.Fatalf("unexpected repo root: %s", location.RepoRoot)
	}
	if canonicalPath(t, status.RepositoryRoot) != canonicalPath(t, repoDir) {
		t.Fatalf("unexpected status repo root: %s", status.RepositoryRoot)
	}
	if status.UntrackedCount < 1 {
		t.Fatalf("expected untracked files, got %d", status.UntrackedCount)
	}
	if status.IsClean {
		t.Fatal("status should not be clean")
	}
}
