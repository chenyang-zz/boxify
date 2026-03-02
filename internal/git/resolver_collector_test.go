package git

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
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

func TestStatusCollectorCollectByPathLineStats(t *testing.T) {
	repoDir := mustInitGitRepo(t)
	filePath := filepath.Join(repoDir, "README.md")

	runGit(t, repoDir, "-c", "user.name=boxify-test", "-c", "user.email=boxify@test.local", "add", "README.md")
	runGit(t, repoDir, "-c", "user.name=boxify-test", "-c", "user.email=boxify@test.local", "commit", "-m", "init", "--no-gpg-sign")

	if err := os.WriteFile(filePath, []byte("line1\nline2\nline3\n"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	runGit(t, repoDir, "add", "README.md")
	if err := os.WriteFile(filePath, []byte("line1\nline3\nline4\n"), 0o644); err != nil {
		t.Fatalf("rewrite file failed: %v", err)
	}

	runner := NewCommandRunner(0, slog.Default())
	resolver := NewResolver(runner, slog.Default())
	parser := NewStatusParser(slog.Default())
	collector := NewStatusCollector(runner, resolver, parser, slog.Default())

	status, _, err := collector.CollectByPath(context.Background(), repoDir)
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}

	if status.AddedLines != 4 {
		t.Fatalf("unexpected added lines: %d", status.AddedLines)
	}
	if status.DeletedLines != 2 {
		t.Fatalf("unexpected deleted lines: %d", status.DeletedLines)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v, output=%s", args, err, string(out))
	}
}
