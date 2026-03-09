package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePluginInstallStrategyPrefersRegistryGitOverNpm(t *testing.T) {
	t.Parallel()

	strategy := resolvePluginInstallStrategy(&RegistryPlugin{
		GitURL:     "https://github.com/example/repo.git",
		NpmPackage: "@openclaw/wecom",
	}, "")

	if strategy.kind != "download" || strategy.target != "https://github.com/example/repo.git" {
		t.Fatalf("expected git/download strategy, got %#v", strategy)
	}
}

func TestResolvePluginInstallStrategyUsesExplicitNpmSource(t *testing.T) {
	t.Parallel()

	strategy := resolvePluginInstallStrategy(&RegistryPlugin{
		GitURL:     "https://github.com/example/repo.git",
		NpmPackage: "@openclaw/wecom",
	}, "@openclaw/custom")

	if strategy.kind != "npm" || strategy.target != "@openclaw/custom" {
		t.Fatalf("expected explicit npm strategy, got %#v", strategy)
	}
}

func TestNormalizeOpenClawInstallSource(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"npm":      "npm",
		"archive":  "archive",
		"path":     "path",
		"local":    "path",
		"registry": "path",
		"custom":   "path",
		"github":   "path",
		"git":      "path",
		"":         "path",
	}

	for input, want := range tests {
		if got := normalizeOpenClawInstallSource(input); got != want {
			t.Fatalf("normalizeOpenClawInstallSource(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSyncOpenClawPluginStateWritesEntriesAndInstalls(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Config{OpenClawDir: dir}
	m := &Manager{cfg: cfg}

	if err := m.syncOpenClawPluginState("dingtalk", dir+"/extensions/dingtalk", true, "registry", "0.2.0"); err != nil {
		t.Fatalf("syncOpenClawPluginState: %v", err)
	}

	saved, err := cfg.ReadOpenClawJSON()
	if err != nil {
		t.Fatalf("ReadOpenClawJSON: %v", err)
	}
	pl, _ := saved["plugins"].(map[string]interface{})
	ent, _ := pl["entries"].(map[string]interface{})
	ins, _ := pl["installs"].(map[string]interface{})
	entry, _ := ent["dingtalk"].(map[string]interface{})
	install, _ := ins["dingtalk"].(map[string]interface{})
	if enabled, _ := entry["enabled"].(bool); !enabled {
		t.Fatalf("expected dingtalk entry enabled, got %#v", entry)
	}
	if got, _ := install["installPath"].(string); got == "" {
		t.Fatalf("expected installPath, got %#v", install)
	}
	if got, _ := install["version"].(string); got != "0.2.0" {
		t.Fatalf("expected version 0.2.0, got %#v", install)
	}
}

func TestRemoveOpenClawPluginStateDeletesEntriesAndInstalls(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Config{OpenClawDir: dir}
	m := &Manager{cfg: cfg}
	if err := m.syncOpenClawPluginState("wecom", dir+"/extensions/wecom", true, "registry", "latest"); err != nil {
		t.Fatalf("seed syncOpenClawPluginState: %v", err)
	}
	if err := m.removeOpenClawPluginState("wecom"); err != nil {
		t.Fatalf("removeOpenClawPluginState: %v", err)
	}
	saved, err := cfg.ReadOpenClawJSON()
	if err != nil {
		t.Fatalf("ReadOpenClawJSON: %v", err)
	}
	pl, _ := saved["plugins"].(map[string]interface{})
	ent, _ := pl["entries"].(map[string]interface{})
	ins, _ := pl["installs"].(map[string]interface{})
	if _, ok := ent["wecom"]; ok {
		t.Fatalf("expected wecom entry removed")
	}
	if _, ok := ins["wecom"]; ok {
		t.Fatalf("expected wecom install removed")
	}
}

func TestHydratePluginsFromOpenClawConfigRestoresConfiguredPlugin(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Config{OpenClawDir: dir}
	m := &Manager{
		cfg:     cfg,
		plugins: make(map[string]*InstalledPlugin),
	}

	if err := cfg.WriteOpenClawJSON(map[string]interface{}{
		"plugins": map[string]interface{}{
			"entries": map[string]interface{}{
				"feishu": map[string]interface{}{"enabled": false},
			},
			"installs": map[string]interface{}{
				"feishu": map[string]interface{}{
					"installPath": filepath.Join(dir, "extensions", "feishu"),
					"version":     "2026.2.25",
					"source":      "npm",
					"installedAt": "2026-02-28T02:53:12Z",
				},
			},
		},
	}); err != nil {
		t.Fatalf("WriteOpenClawJSON: %v", err)
	}

	m.hydratePluginsFromOpenClawConfig()

	got := m.GetPlugin("feishu")
	if got == nil {
		t.Fatalf("expected configured plugin to be restored")
	}
	if got.Enabled {
		t.Fatalf("expected feishu disabled, got %#v", got)
	}
	if got.Source != "npm" {
		t.Fatalf("expected source restored, got %#v", got)
	}
	if got.Dir == "" {
		t.Fatalf("expected install dir restored, got %#v", got)
	}
}

func TestHydratePluginsFromOpenClawConfigLoadsMetadataFromResolvedDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "extensions", "feishu")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	metaData, err := json.Marshal(map[string]interface{}{
		"id":          "feishu",
		"name":        "Feishu",
		"description": "飞书插件",
		"version":     "2026.2.25",
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "package.json"), metaData, 0o644); err != nil {
		t.Fatalf("WriteFile package.json: %v", err)
	}

	cfg := &Config{OpenClawDir: dir}
	m := &Manager{
		cfg:     cfg,
		plugins: make(map[string]*InstalledPlugin),
	}
	if err := cfg.WriteOpenClawJSON(map[string]interface{}{
		"plugins": map[string]interface{}{
			"entries": map[string]interface{}{
				"feishu": map[string]interface{}{"enabled": true},
			},
			"installs": map[string]interface{}{
				"feishu": map[string]interface{}{
					"installPath": pluginDir,
					"version":     "2026.2.25",
					"source":      "path",
				},
			},
		},
	}); err != nil {
		t.Fatalf("WriteOpenClawJSON: %v", err)
	}

	m.hydratePluginsFromOpenClawConfig()

	got := m.GetPlugin("feishu")
	if got == nil {
		t.Fatalf("expected configured plugin to be restored")
	}
	if got.Name != "Feishu" || got.Description != "飞书插件" {
		t.Fatalf("expected metadata loaded from plugin dir, got %#v", got)
	}
	if got.Dir != pluginDir {
		t.Fatalf("expected install dir restored, got %#v", got)
	}
}

func TestListSkillCenterPluginsIncludesDiskAndConfigOnlyPlugins(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Config{OpenClawDir: dir, DataDir: t.TempDir()}
	writeSkillCenterPluginFixture(t, filepath.Join(dir, "extensions", "feishu"), "Feishu Plugin", "plugin desc", "1.2.3")

	if err := cfg.WriteOpenClawJSON(map[string]interface{}{
		"plugins": map[string]interface{}{
			"entries": map[string]interface{}{
				"feishu": map[string]interface{}{"enabled": false},
				"ghost":  map[string]interface{}{"enabled": true},
			},
			"installs": map[string]interface{}{
				"feishu": map[string]interface{}{
					"version":     "9.9.9",
					"installedAt": "2026-03-01T00:00:00Z",
				},
				"ghost": map[string]interface{}{
					"version":     "0.0.1",
					"installedAt": "2026-03-02T00:00:00Z",
					"installPath": "/tmp/ghost",
				},
			},
		},
	}); err != nil {
		t.Fatalf("WriteOpenClawJSON: %v", err)
	}

	m := NewManager(cfg, nil)
	plugins := m.ListSkillCenterPlugins()

	if !containsSkillCenterPlugin(plugins, "feishu", true, "1.2.3") {
		t.Fatalf("expected feishu plugin from disk/config, got %#v", plugins)
	}
	if !containsSkillCenterPlugin(plugins, "ghost", true, "0.0.1") {
		t.Fatalf("expected ghost config-only plugin, got %#v", plugins)
	}
}

func TestListSkillCenterPluginsScansConfiguredInstallPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	externalPluginDir := filepath.Join(t.TempDir(), "external-plugin")
	writeSkillCenterPluginFixture(t, externalPluginDir, "External Plugin", "external desc", "2.0.0")

	cfg := &Config{OpenClawDir: dir, DataDir: t.TempDir()}
	if err := cfg.WriteOpenClawJSON(map[string]interface{}{
		"plugins": map[string]interface{}{
			"entries": map[string]interface{}{
				"external": map[string]interface{}{"enabled": false},
			},
			"installs": map[string]interface{}{
				"external": map[string]interface{}{
					"version":     "2.0.0",
					"installedAt": "2026-03-03T00:00:00Z",
					"installPath": externalPluginDir,
				},
			},
		},
	}); err != nil {
		t.Fatalf("WriteOpenClawJSON: %v", err)
	}

	m := NewManager(cfg, nil)
	plugins := m.ListSkillCenterPlugins()

	if !containsSkillCenterPlugin(plugins, "external", false, "2.0.0") {
		t.Fatalf("expected configured installPath plugin, got %#v", plugins)
	}
}

func writeSkillCenterPluginFixture(t *testing.T, dir, name, description, version string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll plugin dir: %v", err)
	}
	pkg, err := json.Marshal(map[string]interface{}{
		"name":        name,
		"description": description,
		"version":     version,
	})
	if err != nil {
		t.Fatalf("Marshal package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), pkg, 0o644); err != nil {
		t.Fatalf("WriteFile package.json: %v", err)
	}
}

func containsSkillCenterPlugin(items []SkillCenterPlugin, id string, enabled bool, version string) bool {
	for _, item := range items {
		if item.ID == id && item.Enabled == enabled && item.Version == version {
			return true
		}
	}
	return false
}
