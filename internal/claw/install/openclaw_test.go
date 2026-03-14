package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   [3]int
		wantOK bool
	}{
		{name: "full version", input: "v24.1.2", want: [3]int{24, 1, 2}, wantOK: true},
		{name: "missing patch", input: "22.16", want: [3]int{22, 16, 0}, wantOK: true},
		{name: "major only", input: "24", want: [3]int{24, 0, 0}, wantOK: true},
		{name: "invalid", input: "nightly", want: [3]int{}, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch, ok := parseVersion(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("parseVersion(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			got := [3]int{major, minor, patch}
			if got != tt.want {
				t.Fatalf("parseVersion(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsSupportedNodeVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{version: "v22.15.9", want: false},
		{version: "v22.16.0", want: true},
		{version: "v22.16.3", want: true},
		{version: "v23.0.0", want: true},
		{version: "v24.0.0", want: true},
		{version: "invalid", want: false},
	}

	for _, tt := range tests {
		if got := isSupportedNodeVersion(tt.version); got != tt.want {
			t.Fatalf("isSupportedNodeVersion(%q) = %v, want %v", tt.version, got, tt.want)
		}
	}
}

func TestResolveReusableOpenClawConfigFallsBackToHomeDotOpenClaw(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	customDir := filepath.Join(t.TempDir(), "custom-openclaw")
	homeDir := filepath.Join(home, ".openclaw")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(homeDir) error = %v", err)
	}
	homeConfigPath := filepath.Join(homeDir, "openclaw.json")
	if err := os.WriteFile(homeConfigPath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("WriteFile(homeConfigPath) error = %v", err)
	}

	gotDir, gotPath := resolveReusableOpenClawConfig(customDir)
	if gotDir != homeDir {
		t.Fatalf("resolveReusableOpenClawConfig() dir = %q, want %q", gotDir, homeDir)
	}
	if gotPath != homeConfigPath {
		t.Fatalf("resolveReusableOpenClawConfig() path = %q, want %q", gotPath, homeConfigPath)
	}
}

func TestResolveReusableOpenClawConfigPrefersExplicitDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	customDir := filepath.Join(t.TempDir(), "custom-openclaw")
	if err := os.MkdirAll(customDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(customDir) error = %v", err)
	}
	customConfigPath := filepath.Join(customDir, "openclaw.json")
	if err := os.WriteFile(customConfigPath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("WriteFile(customConfigPath) error = %v", err)
	}

	homeDir := filepath.Join(home, ".openclaw")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(homeDir) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(homeDir, "openclaw.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("WriteFile(home openclaw.json) error = %v", err)
	}

	gotDir, gotPath := resolveReusableOpenClawConfig(customDir)
	if gotDir != customDir {
		t.Fatalf("resolveReusableOpenClawConfig() dir = %q, want %q", gotDir, customDir)
	}
	if gotPath != customConfigPath {
		t.Fatalf("resolveReusableOpenClawConfig() path = %q, want %q", gotPath, customConfigPath)
	}
}
