package git

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestResolverNormalizePathExpandHome(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := filepath.Join(homeDir, "repo", "sub")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	t.Setenv("HOME", homeDir)

	resolver := NewResolver(NewCommandRunner(0, slog.Default()), slog.Default())
	normalized, err := resolver.normalizePath("~/repo/sub")
	if err != nil {
		t.Fatalf("normalize path failed: %v", err)
	}

	if filepath.Clean(normalized) != filepath.Clean(targetDir) {
		t.Fatalf("unexpected normalized path: %s", normalized)
	}
}
