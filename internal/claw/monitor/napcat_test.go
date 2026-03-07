package monitor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf16"
)

func TestDecodeUTF16LE(t *testing.T) {
	t.Parallel()

	want := "UserName=DESKTOP\\alice😀\r\n"
	got := decodeUTF16LE(utf16LEWithBOM(want))
	if got != want {
		t.Fatalf("decodeUTF16LE mismatch: got %q want %q", got, want)
	}
}

func TestReadTokenFromNapCatLogs(t *testing.T) {
	t.Parallel()

	t.Run("finds token from newest log under logs dir", func(t *testing.T) {
		bootDir := t.TempDir()
		logsDir := filepath.Join(bootDir, "runtime", "logs")
		if err := os.MkdirAll(logsDir, 0o755); err != nil {
			t.Fatalf("mkdir logs: %v", err)
		}
		if err := os.WriteFile(filepath.Join(logsDir, "20260101.log"), []byte("[WebUi] WebUi Token: old-token\n"), 0o644); err != nil {
			t.Fatalf("write old log: %v", err)
		}
		if err := os.WriteFile(filepath.Join(logsDir, "20260102.log"), []byte("line\n[WebUi] WebUi Token: new-token\n"), 0o644); err != nil {
			t.Fatalf("write new log: %v", err)
		}

		got := readTokenFromNapCatLogs(bootDir)
		if got != "new-token" {
			t.Fatalf("expected newest token, got %q", got)
		}
	})

	t.Run("falls back to parent logs dir", func(t *testing.T) {
		root := t.TempDir()
		parentLogs := filepath.Join(root, "logs")
		if err := os.MkdirAll(parentLogs, 0o755); err != nil {
			t.Fatalf("mkdir parent logs: %v", err)
		}
		if err := os.WriteFile(filepath.Join(parentLogs, "main.log"), []byte("[WebUi] WebUi Token: parent-token\n"), 0o644); err != nil {
			t.Fatalf("write parent log: %v", err)
		}

		bootDir := filepath.Join(root, "nested", "bootmain")
		if err := os.MkdirAll(bootDir, 0o755); err != nil {
			t.Fatalf("mkdir boot dir: %v", err)
		}

		got := readTokenFromNapCatLogs(bootDir)
		if got != "parent-token" {
			t.Fatalf("expected fallback token, got %q", got)
		}
	})
}

func TestEnsureNapCatNetworkConfigWritesDefaultWhenNoUIN(t *testing.T) {
	t.Parallel()

	shellDir := t.TempDir()
	ensureNapCatNetworkConfig(shellDir)

	cfgPath := filepath.Join(shellDir, "config", "onebot11.json")
	cfg := readJSONMap(t, cfgPath)
	assertNetworkDefaults(t, cfg)
}

func TestEnsureNapCatNetworkConfigWritesByUINAndPreservesNonEmptyNetwork(t *testing.T) {
	t.Parallel()

	shellDir := t.TempDir()
	cfgDir := filepath.Join(shellDir, "config")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	// UIN from onebot11_<uin>.json should be updated when network is empty.
	if err := os.WriteFile(filepath.Join(cfgDir, "onebot11_10001.json"), []byte(`{"network":{}}`), 0o644); err != nil {
		t.Fatalf("seed onebot11_10001: %v", err)
	}
	// UIN from napcat_<uin>.json should get a new onebot11_<uin>.json file.
	if err := os.WriteFile(filepath.Join(cfgDir, "napcat_10002.json"), []byte(`{"uin":"10002"}`), 0o644); err != nil {
		t.Fatalf("seed napcat_10002: %v", err)
	}
	// napcat_protocol_<uin>.json must be ignored.
	if err := os.WriteFile(filepath.Join(cfgDir, "napcat_protocol_10003.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatalf("seed napcat_protocol_10003: %v", err)
	}

	preservedPath := filepath.Join(cfgDir, "onebot11_10004.json")
	preservedCfg := map[string]any{
		"network": map[string]any{
			"httpServers":      []any{map[string]any{"port": 9999}},
			"websocketServers": []any{},
		},
	}
	preservedBytes, _ := json.Marshal(preservedCfg)
	if err := os.WriteFile(preservedPath, preservedBytes, 0o644); err != nil {
		t.Fatalf("seed onebot11_10004: %v", err)
	}

	ensureNapCatNetworkConfig(shellDir)

	cfg10001 := readJSONMap(t, filepath.Join(cfgDir, "onebot11_10001.json"))
	assertNetworkDefaults(t, cfg10001)

	cfg10002 := readJSONMap(t, filepath.Join(cfgDir, "onebot11_10002.json"))
	assertNetworkDefaults(t, cfg10002)

	cfg10004 := readJSONMap(t, preservedPath)
	network10004, _ := cfg10004["network"].(map[string]any)
	httpServers, _ := network10004["httpServers"].([]any)
	if len(httpServers) != 1 {
		t.Fatalf("expected preserved config to remain unchanged")
	}
	server0, _ := httpServers[0].(map[string]any)
	if port, _ := server0["port"].(float64); port != 9999 {
		t.Fatalf("expected preserved HTTP port 9999, got %v", server0["port"])
	}

	if _, err := os.Stat(filepath.Join(cfgDir, "onebot11_10003.json")); !os.IsNotExist(err) {
		t.Fatalf("expected no file generated for napcat_protocol_10003")
	}
}

func readJSONMap(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return m
}

func assertNetworkDefaults(t *testing.T, cfg map[string]any) {
	t.Helper()
	network, ok := cfg["network"].(map[string]any)
	if !ok {
		t.Fatalf("missing network config")
	}

	httpServers, _ := network["httpServers"].([]any)
	if len(httpServers) == 0 {
		t.Fatalf("expected at least one http server")
	}
	http0, _ := httpServers[0].(map[string]any)
	if port, _ := http0["port"].(float64); port != 3000 {
		t.Fatalf("expected http port 3000, got %v", http0["port"])
	}

	wsServers, _ := network["websocketServers"].([]any)
	if len(wsServers) == 0 {
		t.Fatalf("expected at least one websocket server")
	}
	ws0, _ := wsServers[0].(map[string]any)
	if port, _ := ws0["port"].(float64); port != 3001 {
		t.Fatalf("expected websocket port 3001, got %v", ws0["port"])
	}
}

func utf16LEWithBOM(s string) []byte {
	runes := []rune(s)
	encoded := utf16.Encode(runes)
	out := make([]byte, 0, 2+len(encoded)*2)
	out = append(out, 0xFF, 0xFE)
	for _, r := range encoded {
		out = append(out, byte(r), byte(r>>8))
	}
	return out
}
