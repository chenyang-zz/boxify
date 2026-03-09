package skill

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	clawplugin "github.com/chenyang-zz/boxify/internal/claw/plugin"
	"github.com/chenyang-zz/boxify/internal/types"
)

// TestManagerListScansSkills 验证管理器会聚合多个技能目录并应用 blocklist。
func TestManagerListScansSkills(t *testing.T) {
	t.Parallel()

	openClawDir := t.TempDir()
	appDir := t.TempDir()
	writeSkillFixture(t, filepath.Join(openClawDir, "skills", "alpha"), "Alpha Skill", "alpha desc")
	writeSkillFixture(t, filepath.Join(filepath.Dir(openClawDir), "work", "skills", "beta"), "Beta Skill", "beta desc")
	writeSkillFixture(t, filepath.Join(appDir, "skills", "gamma"), "Gamma Skill", "gamma desc")

	cfg := &clawplugin.Config{OpenClawDir: openClawDir, DataDir: t.TempDir()}
	if err := cfg.WriteOpenClawJSON(map[string]interface{}{
		"skills": map[string]interface{}{
			"blocklist": []interface{}{"beta"},
		},
	}); err != nil {
		t.Fatalf("WriteOpenClawJSON: %v", err)
	}

	manager := newSkillManagerForTest(cfg, openClawDir, appDir)
	items := manager.List()

	if len(items) != 3 {
		t.Fatalf("expected 3 skills, got %#v", items)
	}
	if !containsSkill(items, "alpha", true) {
		t.Fatalf("expected alpha enabled, got %#v", items)
	}
	if !containsSkill(items, "beta", false) {
		t.Fatalf("expected beta disabled, got %#v", items)
	}
	if !containsSkill(items, "gamma", true) {
		t.Fatalf("expected gamma enabled, got %#v", items)
	}
}

// TestManagerToggleUpdatesBlocklist 验证启停状态会持久化回 openclaw.json。
func TestManagerToggleUpdatesBlocklist(t *testing.T) {
	t.Parallel()

	openClawDir := t.TempDir()
	cfg := &clawplugin.Config{OpenClawDir: openClawDir, DataDir: t.TempDir()}
	if err := cfg.WriteOpenClawJSON(map[string]interface{}{
		"skills": map[string]interface{}{
			"blocklist": []interface{}{"alpha"},
		},
	}); err != nil {
		t.Fatalf("WriteOpenClawJSON: %v", err)
	}

	manager := newSkillManagerForTest(cfg, openClawDir, "")
	if err := manager.Toggle("beta", false); err != nil {
		t.Fatalf("Toggle disable beta: %v", err)
	}
	if err := manager.Toggle("alpha", true); err != nil {
		t.Fatalf("Toggle enable alpha: %v", err)
	}

	ocConfig, err := cfg.ReadOpenClawJSON()
	if err != nil {
		t.Fatalf("ReadOpenClawJSON: %v", err)
	}
	skillsCfg, _ := ocConfig["skills"].(map[string]interface{})
	blocklist, _ := skillsCfg["blocklist"].([]interface{})
	if len(blocklist) != 1 || blocklist[0] != "beta" {
		t.Fatalf("expected blocklist [beta], got %#v", blocklist)
	}
}

// newSkillManagerForTest 创建测试用技能管理器。
func newSkillManagerForTest(cfg *clawplugin.Config, openClawDir, appDir string) *Manager {
	return NewManager(cfg, openClawDir, appDir, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
}

// writeSkillFixture 写入技能测试目录。
func writeSkillFixture(t *testing.T, dir, name, description string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll skill dir: %v", err)
	}
	content := "---\nname: " + name + "\ndescription: " + description + "\n---\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile SKILL.md: %v", err)
	}
}

// containsSkill 判断结果中是否包含指定技能状态。
func containsSkill(items []types.ClawSkill, id string, enabled bool) bool {
	for _, item := range items {
		if item.ID == id && item.Enabled == enabled {
			return true
		}
	}
	return false
}
