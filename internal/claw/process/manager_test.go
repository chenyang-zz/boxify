package process

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestGatewayListening(t *testing.T) {
	openclawDir := newOpenClawDir(t)
	ln, port := listenTCP(t)
	defer ln.Close()
	writeGatewayConfig(t, openclawDir, port)

	mgr := NewManager(ManagerConfig{OpenClawDir: openclawDir}, slog.Default())
	mgr.gatewayProbe = listeningProbe(port)
	if !mgr.GatewayListening() {
		t.Fatalf("expected GatewayListening to detect active gateway port %d", port)
	}
}

func TestGatewayListeningFalseWhenPortClosed(t *testing.T) {
	openclawDir := newOpenClawDir(t)
	ln, port := listenTCP(t)
	_ = ln.Close()
	writeGatewayConfig(t, openclawDir, port)

	mgr := NewManager(ManagerConfig{OpenClawDir: openclawDir}, slog.Default())
	mgr.gatewayProbe = listeningProbe(port)
	if mgr.GatewayListening() {
		t.Fatalf("expected GatewayListening to be false once port %d is closed", port)
	}
}

func TestGetStatusReportsExternallyManagedGateway(t *testing.T) {
	openclawDir := newOpenClawDir(t)
	ln, port := listenTCP(t)
	defer ln.Close()
	writeGatewayConfig(t, openclawDir, port)

	mgr := NewManager(ManagerConfig{OpenClawDir: openclawDir}, slog.Default())
	mgr.gatewayProbe = listeningProbe(port)
	status := mgr.GetStatus()
	if !status.Running {
		t.Fatalf("expected external gateway to be reported as running")
	}
	if !status.ManagedExternally {
		t.Fatalf("expected external gateway to be marked as managed externally")
	}
}

func TestStartRejectsExternallyManagedGateway(t *testing.T) {
	openclawDir := newOpenClawDir(t)
	ln, port := listenTCP(t)
	defer ln.Close()
	writeGatewayConfig(t, openclawDir, port)

	mgr := NewManager(ManagerConfig{OpenClawDir: openclawDir}, slog.Default())
	mgr.gatewayProbe = listeningProbe(port)
	err := mgr.Start()
	if err == nil || !strings.Contains(err.Error(), "外部进程管理") {
		t.Fatalf("expected Start to reject externally managed gateway, got %v", err)
	}
}

func TestGatewayPortCheckTargetsLoopbackBindUsesLoopbackOnly(t *testing.T) {
	allTargets := []string{"127.0.0.1", "localhost", "::1", "10.0.0.2", "fd00:1234:ffff::10"}
	got := gatewayPortCheckTargets("loopback", "", allTargets)
	want := []string{"127.0.0.1", "localhost", "::1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected loopback-only targets, got %#v", got)
	}
}

func TestGatewayPortCheckTargetsCustomHostUsesCustomTarget(t *testing.T) {
	allTargets := []string{"127.0.0.1", "localhost", "::1", "10.0.0.2", "fd00:1234:ffff::10"}
	got := gatewayPortCheckTargets("custom", "10.0.0.2", allTargets)
	want := []string{"10.0.0.2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected custom bind host only, got %#v", got)
	}
}

func TestGatewayPortCheckTargetsTailnetUsesTailnetAddresses(t *testing.T) {
	allTargets := []string{"127.0.0.1", "localhost", "::1", "10.0.0.2", "100.100.100.1", "fd7a:115c:a1e0::1"}
	got := gatewayPortCheckTargets("tailnet", "", allTargets)
	want := []string{"100.100.100.1", "fd7a:115c:a1e0::1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected tailnet-only targets, got %#v", got)
	}
}

func TestGetGatewayPortCheckTargetsUsesRuntimeFallbackWhenLoopbackUnavailable(t *testing.T) {
	openclawDir := newOpenClawDir(t)
	cfgPath := filepath.Join(openclawDir, "openclaw.json")
	if err := os.WriteFile(cfgPath, []byte(`{"gateway":{"bind":"loopback"}}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	mgr := NewManager(ManagerConfig{OpenClawDir: openclawDir}, slog.Default())
	mgr.bindHostCheck = func(string) bool { return false }
	got := mgr.getGatewayPortCheckTargets()
	want := collectGatewayCandidateTargets()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected port-check targets to follow runtime fallback when loopback is unavailable, got %#v", got)
	}
}

func newOpenClawDir(t *testing.T) string {
	t.Helper()
	openclawDir := filepath.Join(t.TempDir(), ".openclaw")
	if err := os.MkdirAll(openclawDir, 0o755); err != nil {
		t.Fatalf("mkdir openclaw dir: %v", err)
	}
	return openclawDir
}

func writeGatewayConfig(t *testing.T, openclawDir string, port int) {
	t.Helper()
	cfgPath := filepath.Join(openclawDir, "openclaw.json")
	if err := os.WriteFile(cfgPath, []byte(fmt.Sprintf(`{"gateway":{"port":%d}}`, port)), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func listenTCP(t *testing.T) (net.Listener, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	return ln, port
}

func listeningProbe(port int) func(string, string) bool {
	expectedPort := strconv.Itoa(port)
	return func(host, actualPort string) bool {
		if actualPort != expectedPort {
			return false
		}
		if host != "localhost" {
			ip := net.ParseIP(host)
			if ip == nil || !ip.IsLoopback() {
				return false
			}
		}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, actualPort), 200000000)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}
}
