package git

import "testing"

func TestParseNumstat(t *testing.T) {
	output := "2\t1\tREADME.md\n-\t-\timage.png\n3\t0\tsrc/main.go\n"
	added, deleted := parseNumstat(output)
	if added != 5 {
		t.Fatalf("unexpected added lines: %d", added)
	}
	if deleted != 1 {
		t.Fatalf("unexpected deleted lines: %d", deleted)
	}
}
