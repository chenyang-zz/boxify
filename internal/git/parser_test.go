package git

import (
	"log/slog"
	"testing"
)

func TestStatusParserParsePorcelainV2(t *testing.T) {
	parser := NewStatusParser(slog.Default())

	lines := []string{
		"# branch.oid 1234567890abcdef",
		"# branch.head main",
		"# branch.upstream origin/main",
		"# branch.ab +2 -1",
		"1 M. N... 100644 100644 100644 abcdef1 abcdef2 file1.txt",
		"2 R. N... 100644 100644 100644 abcdef3 abcdef4 R100 file2-new.txt\tfile2-old.txt",
		"u UU N... 100644 100644 100644 100644 abcdef5 abcdef6 abcdef7 conflict.txt",
		"? untracked.txt",
	}

	status, err := parser.ParsePorcelainV2(lines)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if status.Head != "main" {
		t.Fatalf("unexpected head: %s", status.Head)
	}
	if status.Upstream != "origin/main" {
		t.Fatalf("unexpected upstream: %s", status.Upstream)
	}
	if status.Ahead != 2 || status.Behind != 1 {
		t.Fatalf("unexpected ahead/behind: +%d -%d", status.Ahead, status.Behind)
	}
	if status.StagedCount != 2 {
		t.Fatalf("unexpected staged count: %d", status.StagedCount)
	}
	if status.UntrackedCount != 1 {
		t.Fatalf("unexpected untracked count: %d", status.UntrackedCount)
	}
	if status.ConflictCount != 1 {
		t.Fatalf("unexpected conflict count: %d", status.ConflictCount)
	}
	if len(status.Files) != 4 {
		t.Fatalf("unexpected files length: %d", len(status.Files))
	}
}
