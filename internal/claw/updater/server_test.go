package updater

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// newTestServer 创建测试专用更新服务实例。
func newTestServer(t *testing.T) *Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewServerWithLogger("1.0.0", t.TempDir(), t.TempDir(), 19521, logger)
}

// signedTokenAt 按指定时间戳生成签名令牌，便于覆盖过期场景。
func signedTokenAt(panelPort int, ts int64) string {
	payload := fmt.Sprintf("%d:%d", panelPort, ts)
	mac := hmac.New(sha256.New, []byte(TokenSecret))
	mac.Write([]byte(payload))
	return fmt.Sprintf("%s.%d", hex.EncodeToString(mac.Sum(nil)), ts)
}

func TestGenerateTokenAndValidateToken(t *testing.T) {
	t.Parallel()

	panelPort := 19001
	token := GenerateToken(panelPort)
	if token == "" {
		t.Fatal("GenerateToken should return non-empty token")
	}
	if !ValidateToken(token, panelPort) {
		t.Fatal("ValidateToken should accept token generated for same panel port")
	}
	if ValidateToken(token, panelPort+1) {
		t.Fatal("ValidateToken should reject token for different panel port")
	}
}

func TestValidateTokenRejectsExpiredAndMalformed(t *testing.T) {
	t.Parallel()

	panelPort := 19002

	expiredTS := time.Now().Add(-TokenValidDuration - time.Minute).Unix()
	expiredToken := signedTokenAt(panelPort, expiredTS)
	if ValidateToken(expiredToken, panelPort) {
		t.Fatal("ValidateToken should reject expired token")
	}

	cases := []string{
		"",
		"nosplit",
		"abc.def.ghi",
		"abc.not-a-number",
	}
	for _, tc := range cases {
		if ValidateToken(tc, panelPort) {
			t.Fatalf("ValidateToken should reject malformed token: %q", tc)
		}
	}
}

func TestCheckToken(t *testing.T) {
	t.Parallel()

	s := newTestServer(t)
	valid := GenerateToken(s.panelPort)

	t.Run("query token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/updater/api/check-version?token="+valid, nil)
		rec := httptest.NewRecorder()
		if !s.checkToken(rec, req) {
			t.Fatal("checkToken should pass with valid query token")
		}
	})

	t.Run("header token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/updater/api/check-version", nil)
		req.Header.Set("X-Update-Token", valid)
		rec := httptest.NewRecorder()
		if !s.checkToken(rec, req) {
			t.Fatal("checkToken should pass with valid header token")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/updater/api/check-version?token=bad", nil)
		rec := httptest.NewRecorder()
		if s.checkToken(rec, req) {
			t.Fatal("checkToken should fail with invalid token")
		}
		if rec.Code != 403 {
			t.Fatalf("status should be 403, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "授权令牌无效或已过期") {
			t.Fatalf("unexpected error body: %s", rec.Body.String())
		}
	})
}

func TestHandleValidate(t *testing.T) {
	t.Parallel()

	s := newTestServer(t)
	valid := GenerateToken(s.panelPort)

	req := httptest.NewRequest("GET", "/updater/api/validate?token="+valid, nil)
	rec := httptest.NewRecorder()
	s.handleValidate(rec, req)

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode validate response failed: %v", err)
	}
	if ok, _ := payload["ok"].(bool); !ok {
		t.Fatalf("validate should return ok=true, got payload=%v", payload)
	}
}

func TestHandleProgressReturnsIndependentCopy(t *testing.T) {
	t.Parallel()

	s := newTestServer(t)
	s.mu.Lock()
	s.state.Log = []string{"line-1"}
	s.state.Steps = []UpdateStep{{Name: "A", Status: "running", Message: "m1"}}
	s.mu.Unlock()

	req := httptest.NewRequest("GET", "/updater/api/progress", nil)

	rec1 := httptest.NewRecorder()
	s.handleProgress(rec1, req)
	var first struct {
		OK    bool        `json:"ok"`
		State UpdateState `json:"state"`
	}
	if err := json.Unmarshal(rec1.Body.Bytes(), &first); err != nil {
		t.Fatalf("decode first response failed: %v", err)
	}

	first.State.Log[0] = "mutated-log"
	first.State.Steps[0].Message = "mutated-message"

	rec2 := httptest.NewRecorder()
	s.handleProgress(rec2, req)
	var second struct {
		OK    bool        `json:"ok"`
		State UpdateState `json:"state"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &second); err != nil {
		t.Fatalf("decode second response failed: %v", err)
	}
	if second.State.Log[0] != "line-1" {
		t.Fatalf("state log should be copied, got=%s", second.State.Log[0])
	}
	if second.State.Steps[0].Message != "m1" {
		t.Fatalf("state steps should be copied, got=%s", second.State.Steps[0].Message)
	}
}

func TestSetPhase(t *testing.T) {
	t.Parallel()

	s := newTestServer(t)
	s.setPhase("checking")
	if s.state.FinishedAt != "" {
		t.Fatalf("FinishedAt should stay empty for non-terminal phase, got=%s", s.state.FinishedAt)
	}

	s.setPhase("done")
	if s.state.FinishedAt == "" {
		t.Fatal("FinishedAt should be set for done phase")
	}
	if _, err := time.Parse(time.RFC3339, s.state.FinishedAt); err != nil {
		t.Fatalf("FinishedAt should be RFC3339, got=%s err=%v", s.state.FinishedAt, err)
	}
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
		{name: "with prefix", latest: "v1.2.0", current: "v1.1.9", expect: true},
		{name: "longer segment", latest: "1.2.0.1", current: "1.2.0", expect: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isNewerVersion(tc.latest, tc.current)
			if got != tc.expect {
				t.Fatalf("isNewerVersion(%q,%q)=%v, expect=%v", tc.latest, tc.current, got, tc.expect)
			}
		})
	}
}

func TestUtilityFunctions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	dst := filepath.Join(dir, "dst.bin")
	content := []byte("boxify-updater-test")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatalf("write src failed: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst failed: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("copied content mismatch, got=%q expect=%q", string(got), string(content))
	}

	sum, err := fileSHA256(dst)
	if err != nil {
		t.Fatalf("fileSHA256 failed: %v", err)
	}
	if len(sum) != 64 {
		t.Fatalf("sha256 hex length should be 64, got=%d", len(sum))
	}

	expectedPlatform := runtime.GOOS + "_" + runtime.GOARCH
	if getPlatformKey() != expectedPlatform {
		t.Fatalf("getPlatformKey mismatch, got=%s expect=%s", getPlatformKey(), expectedPlatform)
	}

	if ternary(true, "a", "b") != "a" || ternary(false, "a", "b") != "b" {
		t.Fatal("ternary result mismatch")
	}
}
