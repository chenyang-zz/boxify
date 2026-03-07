package update

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newTestUpdater 创建测试专用 Updater，使用临时目录和静默日志器。
func newTestUpdater(t *testing.T) *Updater {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewUpdaterWithLogger("1.0.0", t.TempDir(), logger)
}

func TestIsNewerVersion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		latest  string
		current string
		expect  bool
	}{
		{name: "major higher", latest: "2.0.0", current: "1.9.9", expect: true},
		{name: "minor higher", latest: "1.2.0", current: "1.1.9", expect: true},
		{name: "patch higher", latest: "1.0.1", current: "1.0.0", expect: true},
		{name: "same version", latest: "1.0.0", current: "1.0.0", expect: false},
		{name: "older version", latest: "1.0.0", current: "1.1.0", expect: false},
		{name: "with v prefix", latest: "v1.2.0", current: "v1.1.9", expect: true},
		{name: "longer segment", latest: "1.2.0.1", current: "1.2.0", expect: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isNewerVersion(tc.latest, tc.current)
			if got != tc.expect {
				t.Fatalf("isNewerVersion(%q, %q) = %v, expect %v", tc.latest, tc.current, got, tc.expect)
			}
		})
	}
}

func TestShortHash(t *testing.T) {
	t.Parallel()

	if got := shortHash("abcd"); got != "abcd" {
		t.Fatalf("shortHash should keep short hash, got: %s", got)
	}

	if got := shortHash("1234567890abcdefxyz"); got != "1234567890abcdef..." {
		t.Fatalf("shortHash should truncate long hash, got: %s", got)
	}
}

func TestFileSHA256(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "sample.txt")
	content := []byte("boxify-test")
	if err := os.WriteFile(target, content, 0o644); err != nil {
		t.Fatalf("write test file failed: %v", err)
	}

	got, err := fileSHA256(target)
	if err != nil {
		t.Fatalf("fileSHA256 failed: %v", err)
	}

	sum := sha256.Sum256(content)
	expect := hex.EncodeToString(sum[:])
	if got != expect {
		t.Fatalf("fileSHA256 mismatch, got=%s expect=%s", got, expect)
	}
}

func TestCopyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	dst := filepath.Join(dir, "dst.bin")
	expect := []byte("copy-file-content")
	if err := os.WriteFile(src, expect, 0o644); err != nil {
		t.Fatalf("write src failed: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst failed: %v", err)
	}
	if string(got) != string(expect) {
		t.Fatalf("copied content mismatch, got=%q expect=%q", string(got), string(expect))
	}
}

func TestSaveAndMarkPopup(t *testing.T) {
	t.Parallel()

	u := newTestUpdater(t)
	info := &UpdateInfo{
		LatestVersion: "1.2.3",
		ReleaseNote:   "fixes",
	}
	u.saveUpdatePopup(info)

	popup := u.GetUpdatePopup()
	if popup == nil {
		t.Fatalf("popup should exist after save")
	}
	if !popup.Show || popup.Version != "1.2.3" || popup.ReleaseNote != "fixes" {
		t.Fatalf("unexpected popup content: %+v", *popup)
	}
	if popup.ShownAt != "" {
		t.Fatalf("ShownAt should be empty before mark shown")
	}

	u.MarkPopupShown()
	popup = u.GetUpdatePopup()
	if popup == nil {
		t.Fatalf("popup should exist after mark shown")
	}
	if popup.Show {
		t.Fatalf("popup show should be false after mark shown")
	}
	if popup.ShownAt == "" {
		t.Fatalf("ShownAt should be filled after mark shown")
	}
	if _, err := time.Parse(time.RFC3339, popup.ShownAt); err != nil {
		t.Fatalf("ShownAt should be RFC3339, got=%s err=%v", popup.ShownAt, err)
	}
}

func TestDoUpdateNilInfoSetsError(t *testing.T) {
	t.Parallel()

	u := newTestUpdater(t)
	u.DoUpdate(nil)

	p := u.GetProgress()
	if p.Status != "error" {
		t.Fatalf("status should be error, got=%s", p.Status)
	}
	if p.Error == "" || !strings.Contains(p.Error, "更新信息不能为空") {
		t.Fatalf("unexpected error message: %s", p.Error)
	}
	if p.FinishedAt == "" {
		t.Fatalf("FinishedAt should be set on error")
	}
}

func TestDoUpdateIgnoreWhenBusy(t *testing.T) {
	t.Parallel()

	u := newTestUpdater(t)
	u.mu.Lock()
	u.progress.Status = "downloading"
	u.progress.Progress = 35
	u.progress.Message = "busy"
	u.mu.Unlock()

	u.DoUpdate(&UpdateInfo{LatestVersion: "2.0.0"})
	p := u.GetProgress()
	if p.Status != "downloading" || p.Progress != 35 || p.Message != "busy" {
		t.Fatalf("progress should keep busy status, got=%+v", p)
	}
}

func TestGetProgressReturnsLogCopy(t *testing.T) {
	t.Parallel()

	u := newTestUpdater(t)
	u.mu.Lock()
	u.progress.Log = []string{"line1"}
	u.mu.Unlock()

	snapshot := u.GetProgress()
	snapshot.Log[0] = "changed"

	latest := u.GetProgress()
	if latest.Log[0] != "line1" {
		t.Fatalf("internal log should not be changed by snapshot mutation, got=%s", latest.Log[0])
	}
}
