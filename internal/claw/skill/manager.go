package skill

import (
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	clawplugin "github.com/chenyang-zz/boxify/internal/claw/plugin"
	"github.com/chenyang-zz/boxify/internal/types"
)

// Manager 负责技能扫描与启停状态管理。
type Manager struct {
	cfg         *clawplugin.Config // OpenClaw 配置读写器。
	openClawDir string             // OpenClaw 配置目录。
	openClawApp string             // OpenClaw 应用目录。
	logger      *slog.Logger       // 日志记录器。
}

// NewManager 创建技能管理器。
func NewManager(cfg *clawplugin.Config, openClawDir, openClawApp string, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	return &Manager{
		cfg:         cfg,
		openClawDir: strings.TrimSpace(openClawDir),
		openClawApp: strings.TrimSpace(openClawApp),
		logger:      logger.With("module", "claw.skill.manager"),
	}
}

// List 返回技能中心所需的技能列表。
func (m *Manager) List() []types.ClawSkill {
	if m == nil {
		return []types.ClawSkill{}
	}
	m.logger.Debug("开始扫描技能列表", "open_claw_dir", m.openClawDir, "open_claw_app", m.openClawApp)

	ocConfig, _ := m.readOpenClawConfig()
	blockSet := buildBlockSet(ocConfig)

	skills := make([]types.ClawSkill, 0)
	seenSkills := make(map[string]bool)
	for _, dir := range m.searchDirs() {
		m.scanDir(dir, blockSet, &skills, seenSkills)
	}

	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})
	m.logger.Debug("技能列表扫描完成", "count", len(skills))
	return skills
}

// Toggle 更新技能启用状态。
func (m *Manager) Toggle(id string, enabled bool) error {
	if m == nil || m.cfg == nil {
		return nil
	}

	id = strings.TrimSpace(id)
	ocConfig, err := m.readOpenClawConfig()
	if err != nil {
		return err
	}
	skillsCfg, _ := ocConfig["skills"].(map[string]interface{})
	if skillsCfg == nil {
		skillsCfg = map[string]interface{}{}
	}

	blockSet := buildBlockSet(ocConfig)
	if enabled {
		delete(blockSet, id)
	} else {
		blockSet[id] = true
	}

	blocklist := make([]interface{}, 0, len(blockSet))
	keys := make([]string, 0, len(blockSet))
	for key := range blockSet {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		blocklist = append(blocklist, key)
	}

	skillsCfg["blocklist"] = blocklist
	ocConfig["skills"] = skillsCfg
	m.logger.Info("更新技能状态", "skill_id", id, "enabled", enabled, "blocklist_size", len(blocklist))
	return m.cfg.WriteOpenClawJSON(ocConfig)
}

// readOpenClawConfig 读取 openclaw.json，并在缺失时返回空配置。
func (m *Manager) readOpenClawConfig() (map[string]interface{}, error) {
	if m == nil || m.cfg == nil {
		return map[string]interface{}{}, nil
	}
	ocConfig, err := m.cfg.ReadOpenClawJSON()
	if ocConfig == nil {
		ocConfig = map[string]interface{}{}
	}
	return ocConfig, err
}

// searchDirs 返回技能扫描目录，兼容 OpenClaw 与应用内置技能目录。
func (m *Manager) searchDirs() []string {
	parentDir := filepath.Dir(m.openClawDir)
	workDir := firstNonEmptyString(
		existingDir(filepath.Join(parentDir, "work")),
		existingDir(filepath.Join(parentDir, "openclaw", "work")),
		existingDir(filepath.Join(detectGlobalOpenClawDir(), "agents")),
	)
	appDir := firstNonEmptyString(
		existingDir(m.openClawApp),
		existingDir(filepath.Join(parentDir, "app")),
		existingDir(filepath.Join(parentDir, "openclaw", "app")),
		existingDir(detectGlobalOpenClawDir()),
	)

	dirs := []string{
		filepath.Join(m.openClawDir, "skills"),
		filepath.Join(workDir, "skills"),
	}
	if appDir != "" {
		dirs = append(dirs, filepath.Join(appDir, "skills"))
	}
	dirs = append(dirs, filepath.Join(parentDir, "app", "skills"))
	return uniqueExistingDirs(dirs)
}
