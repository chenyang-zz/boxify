package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func mustInitGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v, output=%s", args, err, string(out))
		}
	}

	run("init", "-q")

	filePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(filePath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write seed file failed: %v", err)
	}

	return dir
}
