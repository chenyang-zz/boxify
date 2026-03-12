package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	RegistryURL       = "https://raw.githubusercontent.com/zhaoxinyi02/ClawPanel-Plugins/main/registry.json" // 官方插件仓库地址。
	RegistryMirrorURL = "http://39.102.53.188:16198/clawpanel/plugins/registry.json"                         // 国内加速镜像地址。
)

// 插件元数据（openclaw.plugin.json + package.json 聚合结果）。
type PluginMeta struct {
	ID           string            `json:"id"`                     // 插件唯一标识。
	Name         string            `json:"name"`                   // 插件展示名称。
	Version      string            `json:"version"`                // 插件版本号。
	Author       string            `json:"author"`                 // 插件作者信息。
	Description  string            `json:"description"`            // 插件功能简介。
	Homepage     string            `json:"homepage,omitempty"`     // 插件主页地址。
	Repository   string            `json:"repository,omitempty"`   // 插件源码仓库地址。
	License      string            `json:"license,omitempty"`      // 插件许可证标识。
	Category     string            `json:"category,omitempty"`     // 插件分类（如 basic/ai/message/fun/tool）。
	Tags         []string          `json:"tags,omitempty"`         // 插件标签列表。
	Icon         string            `json:"icon,omitempty"`         // 插件图标地址或路径。
	MinOpenClaw  string            `json:"minOpenClaw,omitempty"`  // 要求的最小 OpenClaw 版本。
	MinPanel     string            `json:"minPanel,omitempty"`     // 要求的最小面板版本。
	EntryPoint   string            `json:"entryPoint,omitempty"`   // 插件主入口脚本文件。
	ConfigSchema json.RawMessage   `json:"configSchema,omitempty"` // 插件配置 JSON Schema。
	Dependencies map[string]string `json:"dependencies,omitempty"` // 插件依赖及版本约束。
	Permissions  []string          `json:"permissions,omitempty"`  // 插件请求的权限列表。
}

// 已安装到磁盘的插件。
type InstalledPlugin struct {
	PluginMeta                         // 内嵌插件基础元数据。
	Enabled     bool                   `json:"enabled"`             // 插件当前是否启用。
	InstalledAt string                 `json:"installedAt"`         // 首次安装时间（RFC3339）。
	UpdatedAt   string                 `json:"updatedAt,omitempty"` // 最近更新时间（RFC3339）。
	Source      string                 `json:"source"`              // 安装来源（registry/local/github/custom/npm）。
	Dir         string                 `json:"dir"`                 // 插件实际安装目录。
	Config      map[string]interface{} `json:"config,omitempty"`    // 插件运行配置。
	LogLines    []string               `json:"logLines,omitempty"`  // 最近插件日志行缓存。
}

// 插件仓库中的插件项。
type RegistryPlugin struct {
	PluginMeta           // 内嵌仓库中的插件元数据。
	Downloads     int    `json:"downloads,omitempty"`     // 下载次数统计。
	Stars         int    `json:"stars,omitempty"`         // 收藏/点赞统计。
	DownloadURL   string `json:"downloadUrl,omitempty"`   // 插件归档下载地址。
	GitURL        string `json:"gitUrl,omitempty"`        // 插件源码仓库地址。
	InstallSubDir string `json:"installSubDir,omitempty"` // 仓库内需要安装的子目录。
	NpmPackage    string `json:"npmPackage,omitempty"`    // 可直接安装的 npm 包名。
	Screenshot    string `json:"screenshot,omitempty"`    // 插件截图地址。
	Readme        string `json:"readme,omitempty"`        // 插件说明文档地址或内容索引。
}

// 插件仓库数据结构。
type Registry struct {
	Version   string           `json:"version"`   // 仓库索引版本号。
	UpdatedAt string           `json:"updatedAt"` // 仓库更新时间（RFC3339）。
	Plugins   []RegistryPlugin `json:"plugins"`   // 仓库内插件列表。
}

// 插件生命周期管理器。
type Manager struct {
	cfg        *Config                     // 插件管理依赖的路径配置。
	plugins    map[string]*InstalledPlugin // 已安装插件的内存索引（key=pluginID）。
	registry   *Registry                   // 最近一次加载的插件仓库缓存。
	mu         sync.RWMutex                // 插件状态与仓库缓存的并发访问锁。
	pluginsDir string                      // 本地插件目录（OpenClaw/extensions）。
	configFile string                      // 插件状态持久化文件路径（plugins.json）。
	logger     *slog.Logger                // 日志记录器。
}

// SkillCenterPlugin 表示技能中心中的插件列表项。
type SkillCenterPlugin struct {
	ID          string // 插件唯一标识。
	Name        string // 插件展示名称。
	Description string // 插件说明。
	Version     string // 插件版本。
	Enabled     bool   // 是否启用。
	Source      string // 插件来源。
	InstalledAt string // 安装时间。
	Path        string // 安装目录。
}

type skillCenterPluginSearchDir struct {
	dir    string
	source string
}

// 创建插件管理器。
func NewManager(cfg *Config, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("module", "plugin.manager")

	pluginsDir := filepath.Join(cfg.OpenClawDir, "extensions")
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(pluginsDir, 0o755); mkErr != nil {
			logger.Warn("创建插件目录失败", "path", pluginsDir, "error", mkErr)
		}
	}
	m := &Manager{
		cfg:        cfg,
		plugins:    make(map[string]*InstalledPlugin),
		pluginsDir: pluginsDir,
		configFile: filepath.Join(cfg.DataDir, "plugins.json"),
		logger:     logger,
	}
	m.loadPluginsState()
	m.scanInstalledPlugins()
	m.logger.Info("插件管理器初始化完成", "plugins_dir", m.pluginsDir, "state_file", m.configFile, "installed_count", len(m.plugins))
	return m
}

// 返回插件目录路径。
func (m *Manager) GetPluginsDir() string {
	return m.pluginsDir
}

// 返回已安装插件列表。
func (m *Manager) ListInstalled() []*InstalledPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*InstalledPlugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		result = append(result, p)
	}
	return result
}

// ListSkillCenterPlugins 返回技能中心所需的已安装插件视图。
func (m *Manager) ListSkillCenterPlugins() []SkillCenterPlugin {
	if m == nil || m.cfg == nil {
		return []SkillCenterPlugin{}
	}

	ocConfig, _ := m.cfg.ReadOpenClawJSON()
	if ocConfig == nil {
		ocConfig = map[string]interface{}{}
	}

	pluginEntries, pluginInstalls := extractOpenClawPluginState(ocConfig)
	pluginEntries, pluginInstalls = filterSkillPluginState(m.cfg.OpenClawDir, pluginEntries, pluginInstalls)

	plugins := make([]SkillCenterPlugin, 0)
	seenPlugins := make(map[string]bool)
	for _, sourceDir := range m.skillCenterPluginSearchDirs() {
		scanSkillCenterPluginDir(sourceDir.dir, pluginEntries, pluginInstalls, &plugins, seenPlugins, sourceDir.source)
	}
	scanConfiguredPluginInstalls(pluginEntries, pluginInstalls, &plugins, seenPlugins)
	appendConfigOnlyPlugins(pluginEntries, pluginInstalls, &plugins, seenPlugins)

	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})
	return plugins
}

// skillCenterPluginSearchDirs 返回技能中心插件扫描目录，当前仅扫描当前 OpenClaw 实例。
func (m *Manager) skillCenterPluginSearchDirs() []skillCenterPluginSearchDir {
	if m == nil {
		return nil
	}

	candidates := []skillCenterPluginSearchDir{
		{dir: m.pluginsDir, source: "installed"},
	}
	seen := make(map[string]struct{})
	result := make([]skillCenterPluginSearchDir, 0, len(candidates))
	for _, candidate := range candidates {
		candidate.dir = strings.TrimSpace(candidate.dir)
		if candidate.dir == "" {
			continue
		}
		candidate.dir = filepath.Clean(candidate.dir)
		if _, err := os.Stat(candidate.dir); err != nil {
			continue
		}
		if _, ok := seen[candidate.dir]; ok {
			continue
		}
		seen[candidate.dir] = struct{}{}
		result = append(result, candidate)
	}
	return result
}

// 返回指定插件信息。
func (m *Manager) GetPlugin(id string) *InstalledPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins[id]
}

// 从远端拉取插件仓库数据。
func (m *Manager) FetchRegistry() (*Registry, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	// 先尝试镜像地址，再回退官方地址。
	urls := []string{RegistryMirrorURL, RegistryURL}
	var lastErr error
	for _, url := range urls {
		m.logger.Debug("开始拉取插件仓库", "url", url)
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			m.logger.Warn("拉取插件仓库失败", "url", url, "error", err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
			m.logger.Warn("插件仓库返回异常状态码", "url", url, "status", resp.StatusCode)
			continue
		}
		var reg Registry
		if err := json.NewDecoder(resp.Body).Decode(&reg); err != nil {
			lastErr = fmt.Errorf("parse registry: %v", err)
			m.logger.Warn("解析插件仓库数据失败", "url", url, "error", err)
			continue
		}
		m.mu.Lock()
		m.registry = &reg
		m.mu.Unlock()
		// 缓存到本地磁盘。
		m.cacheRegistry(&reg)
		m.logger.Info("拉取插件仓库成功", "url", url, "plugin_count", len(reg.Plugins), "version", reg.Version)
		return &reg, nil
	}

	// 网络失败时回退本地缓存。
	if cached := m.loadCachedRegistry(); cached != nil {
		m.logger.Warn("插件仓库网络拉取失败，已回退本地缓存", "plugin_count", len(cached.Plugins), "error", lastErr)
		return cached, nil
	}

	m.logger.Error("获取插件仓库失败且无本地缓存", "error", lastErr)
	return nil, fmt.Errorf("获取插件仓库失败: %v", lastErr)
}

// 返回内存缓存仓库，缺失时尝试读取磁盘缓存。
func (m *Manager) GetRegistry() *Registry {
	m.mu.RLock()
	reg := m.registry
	m.mu.RUnlock()
	if reg != nil {
		return reg
	}
	if cached := m.loadCachedRegistry(); cached != nil {
		m.mu.Lock()
		m.registry = cached
		m.mu.Unlock()
		return cached
	}
	return &Registry{Plugins: []RegistryPlugin{}}
}

// 一次安装应采用的来源类型与目标地址。
type pluginInstallStrategy struct {
	kind   string // 安装方式（如 npm/download）。
	target string // 安装目标（包名或下载地址）。
}

// 根据显式来源与仓库元数据推断安装策略。
func resolvePluginInstallStrategy(regPlugin *RegistryPlugin, source string) pluginInstallStrategy {
	source = strings.TrimSpace(source)
	if source != "" {
		if strings.HasPrefix(source, "@") || (!strings.Contains(source, "/") && !strings.HasSuffix(source, ".git") && source != "") {
			return pluginInstallStrategy{kind: "npm", target: source}
		}
		return pluginInstallStrategy{kind: "download", target: source}
	}
	if regPlugin == nil {
		return pluginInstallStrategy{}
	}
	if regPlugin.DownloadURL != "" {
		return pluginInstallStrategy{kind: "download", target: regPlugin.DownloadURL}
	}
	if regPlugin.GitURL != "" {
		return pluginInstallStrategy{kind: "download", target: regPlugin.GitURL}
	}
	if regPlugin.NpmPackage != "" {
		return pluginInstallStrategy{kind: "npm", target: regPlugin.NpmPackage}
	}
	return pluginInstallStrategy{}
}

// 安装插件（支持仓库源、git、npm 与归档）。
func (m *Manager) Install(pluginID string, source string) error {
	m.logger.Info("开始安装插件", "plugin_id", pluginID, "source", source)

	// 在仓库中查找插件。
	reg := m.GetRegistry()
	var regPlugin *RegistryPlugin
	for i := range reg.Plugins {
		if reg.Plugins[i].ID == pluginID {
			regPlugin = &reg.Plugins[i]
			break
		}
	}

	if regPlugin == nil && source == "" {
		m.logger.Warn("插件不在仓库且未提供安装源", "plugin_id", pluginID)
		return fmt.Errorf("插件 %s 不在仓库中，请提供安装源", pluginID)
	}

	strategy := resolvePluginInstallStrategy(regPlugin, source)
	m.logger.Debug("已解析插件安装策略", "plugin_id", pluginID, "kind", strategy.kind, "target", strategy.target)
	if strategy.kind == "npm" {
		npmPkg := strategy.target
		// 通过 npm 全局安装。
		if err := m.installFromNpm(npmPkg); err != nil {
			m.logger.Warn("npm 安装插件失败", "plugin_id", pluginID, "package", npmPkg, "error", err)
			return fmt.Errorf("npm 安装失败: %v", err)
		}
		// 查找 npm 安装目录。
		npmRoot := ""
		if out, err := exec.Command("npm", "root", "-g").Output(); err == nil {
			npmRoot = strings.TrimSpace(string(out))
		}
		pkgName := npmPkg
		if idx := strings.LastIndex(pkgName, "/"); idx >= 0 {
			pkgName = pkgName[idx+1:]
		}
		installedDir := ""
		if npmRoot != "" {
			// 对于作用域包（如 @openclaw/feishu），目录名包含作用域前缀。
			installedDir = filepath.Join(npmRoot, npmPkg)
			if _, err := os.Stat(installedDir); err != nil {
				installedDir = filepath.Join(npmRoot, pkgName)
			}
		}
		meta := &PluginMeta{ID: pluginID, Name: pluginID}
		if regPlugin != nil {
			meta = &regPlugin.PluginMeta
		}
		if installedDir == "" {
			installedDir = npmPkg
		}
		installed := &InstalledPlugin{
			PluginMeta:  *meta,
			Enabled:     true,
			InstalledAt: time.Now().Format(time.RFC3339),
			Source:      "npm",
			Dir:         installedDir,
		}
		m.mu.Lock()
		m.plugins[meta.ID] = installed
		m.mu.Unlock()
		m.savePluginsState()
		if err := m.syncOpenClawPluginState(meta.ID, installedDir, installed.Enabled, installed.Source, meta.Version); err != nil {
			m.logger.Warn("同步 OpenClaw 插件状态失败", "plugin_id", meta.ID, "error", err)
			return err
		}
		m.logger.Info("插件安装完成", "plugin_id", meta.ID, "source", installed.Source, "dir", installed.Dir)
		return nil
	}

	// 确定下载地址（git/压缩包）。
	downloadURL := strategy.target

	if downloadURL == "" {
		m.logger.Warn("无法确定插件安装方式", "plugin_id", pluginID)
		return fmt.Errorf("无法确定插件 %s 的安装方式，请提供 npm 包名或下载地址", pluginID)
	}

	pluginDir := filepath.Join(m.pluginsDir, pluginID)

	// 检查是否已安装。
	if _, err := os.Stat(pluginDir); err == nil {
		m.logger.Warn("插件已安装，拒绝重复安装", "plugin_id", pluginID, "dir", pluginDir)
		return fmt.Errorf("插件 %s 已安装，请先卸载或使用更新功能", pluginID)
	}

	// 从仓库元数据确定安装子目录。
	installSubDir := ""
	if regPlugin != nil {
		installSubDir = regPlugin.InstallSubDir
	}

	// 根据来源类型执行安装。
	if strings.HasSuffix(downloadURL, ".git") || strings.Contains(downloadURL, "github.com") || strings.Contains(downloadURL, "gitee.com") {
		if installSubDir != "" {
			// 先克隆完整仓库到临时目录，再复制子目录。
			tmpDir, err := os.MkdirTemp("", "clawpanel-plugin-*")
			if err != nil {
				return fmt.Errorf("创建临时目录失败: %v", err)
			}
			defer os.RemoveAll(tmpDir)
			if err := m.installFromGit(downloadURL, tmpDir); err != nil {
				return fmt.Errorf("Git 安装失败: %v", err)
			}
			subPath := filepath.Join(tmpDir, filepath.FromSlash(installSubDir))
			if _, err := os.Stat(subPath); err != nil {
				return fmt.Errorf("子目录 %s 在仓库中不存在", installSubDir)
			}
			if err := copyDir(subPath, pluginDir); err != nil {
				os.RemoveAll(pluginDir)
				return fmt.Errorf("复制插件目录失败: %v", err)
			}
		} else {
			if err := m.installFromGit(downloadURL, pluginDir); err != nil {
				os.RemoveAll(pluginDir)
				return fmt.Errorf("Git 安装失败: %v", err)
			}
		}
	} else if strings.HasSuffix(downloadURL, ".zip") || strings.HasSuffix(downloadURL, ".tar.gz") {
		// 下载压缩包安装。
		if err := m.installFromArchive(downloadURL, pluginDir); err != nil {
			os.RemoveAll(pluginDir)
			return fmt.Errorf("下载安装失败: %v", err)
		}
	} else {
		// 兜底尝试使用 git clone。
		if err := m.installFromGit(downloadURL, pluginDir); err != nil {
			os.RemoveAll(pluginDir)
			return fmt.Errorf("安装失败: %v", err)
		}
	}

	// 读取插件元数据。
	meta, err := m.readPluginMeta(pluginDir)
	if err != nil {
		// 若缺少 openclaw.plugin.json，则构造最小元数据。
		meta = &PluginMeta{
			ID:   pluginID,
			Name: pluginID,
		}
		if regPlugin != nil {
			meta = &regPlugin.PluginMeta
		}
	}

	// 若存在 package.json，安装 npm 生产依赖。
	if _, err := os.Stat(filepath.Join(pluginDir, "package.json")); err == nil {
		cmd := exec.Command("npm", "install", "--production", "--registry=https://registry.npmmirror.com")
		cmd.Dir = pluginDir
		cmd.Run()
	}

	// 注册已安装插件。
	installed := &InstalledPlugin{
		PluginMeta:  *meta,
		Enabled:     true,
		InstalledAt: time.Now().Format(time.RFC3339),
		Source:      "registry",
		Dir:         pluginDir,
	}
	if source != "" {
		installed.Source = "custom"
	}

	m.mu.Lock()
	m.plugins[meta.ID] = installed
	m.mu.Unlock()
	m.savePluginsState()
	if err := m.syncOpenClawPluginState(meta.ID, pluginDir, installed.Enabled, installed.Source, meta.Version); err != nil {
		m.logger.Warn("同步 OpenClaw 插件状态失败", "plugin_id", meta.ID, "error", err)
		return err
	}
	m.logger.Info("插件安装完成", "plugin_id", meta.ID, "source", installed.Source, "dir", installed.Dir)

	return nil
}

// 从本地目录安装插件。
func (m *Manager) InstallLocal(srcDir string) error {
	m.logger.Info("开始本地安装插件", "source_dir", srcDir)
	meta, err := m.readPluginMeta(srcDir)
	if err != nil {
		m.logger.Warn("读取本地插件信息失败", "source_dir", srcDir, "error", err)
		return fmt.Errorf("读取插件信息失败: %v", err)
	}

	pluginDir := filepath.Join(m.pluginsDir, meta.ID)
	if _, err := os.Stat(pluginDir); err == nil {
		return fmt.Errorf("插件 %s 已安装", meta.ID)
	}

	// 复制插件目录。
	if err := copyDir(srcDir, pluginDir); err != nil {
		return fmt.Errorf("复制插件失败: %v", err)
	}

	installed := &InstalledPlugin{
		PluginMeta:  *meta,
		Enabled:     true,
		InstalledAt: time.Now().Format(time.RFC3339),
		Source:      "local",
		Dir:         pluginDir,
	}

	m.mu.Lock()
	m.plugins[meta.ID] = installed
	m.mu.Unlock()
	m.savePluginsState()
	m.logger.Info("本地插件安装完成", "plugin_id", meta.ID, "dir", pluginDir)

	return nil
}

// 卸载插件。
func (m *Manager) Uninstall(pluginID string) error {
	m.logger.Info("开始卸载插件", "plugin_id", pluginID)
	m.mu.Lock()
	p, ok := m.plugins[pluginID]
	if !ok {
		m.mu.Unlock()
		m.logger.Warn("卸载失败，插件未安装", "plugin_id", pluginID)
		return fmt.Errorf("插件 %s 未安装", pluginID)
	}
	delete(m.plugins, pluginID)
	m.mu.Unlock()

	// 删除插件目录。
	if p.Dir != "" {
		if err := os.RemoveAll(p.Dir); err != nil {
			m.logger.Warn("删除插件目录失败", "plugin_id", pluginID, "dir", p.Dir, "error", err)
		}
	}

	m.savePluginsState()
	if err := m.removeOpenClawPluginState(pluginID); err != nil {
		m.logger.Warn("移除 OpenClaw 插件状态失败", "plugin_id", pluginID, "error", err)
		return err
	}
	m.logger.Info("插件卸载完成", "plugin_id", pluginID)
	return nil
}

// 启用插件。
func (m *Manager) Enable(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.plugins[pluginID]
	if !ok {
		return fmt.Errorf("插件 %s 未安装", pluginID)
	}
	p.Enabled = true
	m.savePluginsStateUnlocked()
	if err := m.syncOpenClawPluginState(p.ID, p.Dir, true, p.Source, p.Version); err != nil {
		m.logger.Warn("启用插件后同步 OpenClaw 状态失败", "plugin_id", p.ID, "error", err)
		return err
	}
	m.logger.Info("插件已启用", "plugin_id", p.ID)
	return nil
}

// 禁用插件。
func (m *Manager) Disable(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.plugins[pluginID]
	if !ok {
		return fmt.Errorf("插件 %s 未安装", pluginID)
	}
	p.Enabled = false
	m.savePluginsStateUnlocked()
	if err := m.syncOpenClawPluginState(p.ID, p.Dir, false, p.Source, p.Version); err != nil {
		m.logger.Warn("禁用插件后同步 OpenClaw 状态失败", "plugin_id", p.ID, "error", err)
		return err
	}
	m.logger.Info("插件已禁用", "plugin_id", p.ID)
	return nil
}

// 更新插件配置。
func (m *Manager) UpdateConfig(pluginID string, cfg map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.plugins[pluginID]
	if !ok {
		return fmt.Errorf("插件 %s 未安装", pluginID)
	}
	p.Config = cfg

	// 同步写入插件目录下的 config.json。
	configPath := filepath.Join(p.Dir, "config.json")
	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		m.logger.Warn("写入插件配置文件失败", "plugin_id", pluginID, "path", configPath, "error", err)
	}

	m.savePluginsStateUnlocked()
	m.logger.Debug("插件配置更新完成", "plugin_id", pluginID, "config_path", configPath)
	return nil
}

// 获取插件配置。
func (m *Manager) GetConfig(pluginID string) (map[string]interface{}, json.RawMessage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[pluginID]
	if !ok {
		return nil, nil, fmt.Errorf("插件 %s 未安装", pluginID)
	}

	// 兜底从插件目录读取配置文件。
	cfg := p.Config
	if cfg == nil {
		configPath := filepath.Join(p.Dir, "config.json")
		if data, err := os.ReadFile(configPath); err == nil {
			json.Unmarshal(data, &cfg)
		}
	}
	if cfg == nil {
		cfg = map[string]interface{}{}
	}

	return cfg, p.ConfigSchema, nil
}

// 获取插件最近日志。
func (m *Manager) GetPluginLogs(pluginID string) ([]string, error) {
	m.mu.RLock()
	p, ok := m.plugins[pluginID]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("插件 %s 未安装", pluginID)
	}

	// 优先读取插件目录中的日志文件。
	logPath := filepath.Join(p.Dir, "plugin.log")
	if data, err := os.ReadFile(logPath); err == nil {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 200 {
			lines = lines[len(lines)-200:]
		}
		return lines, nil
	}

	return p.LogLines, nil
}

// 更新插件到最新版本。
func (m *Manager) Update(pluginID string) error {
	m.logger.Info("开始更新插件", "plugin_id", pluginID)
	m.mu.RLock()
	p, ok := m.plugins[pluginID]
	m.mu.RUnlock()
	if !ok {
		m.logger.Warn("更新失败，插件未安装", "plugin_id", pluginID)
		return fmt.Errorf("插件 %s 未安装", pluginID)
	}

	if p.Dir == "" {
		return fmt.Errorf("插件目录未知")
	}

	// 若是 Git 仓库，执行 git pull 更新。
	gitDir := filepath.Join(p.Dir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		cmd := exec.Command("git", "pull", "--rebase")
		cmd.Dir = p.Dir
		if out, err := cmd.CombinedOutput(); err != nil {
			m.logger.Warn("git pull 失败", "plugin_id", pluginID, "dir", p.Dir, "error", err)
			return fmt.Errorf("git pull 失败: %s %v", string(out), err)
		}

		// 重新读取元数据。
		if meta, err := m.readPluginMeta(p.Dir); err == nil {
			m.mu.Lock()
			p.PluginMeta = *meta
			p.UpdatedAt = time.Now().Format(time.RFC3339)
			m.mu.Unlock()
			m.savePluginsState()
		}

		// 重新安装 npm 生产依赖。
		if _, err := os.Stat(filepath.Join(p.Dir, "package.json")); err == nil {
			cmd := exec.Command("npm", "install", "--production", "--registry=https://registry.npmmirror.com")
			cmd.Dir = p.Dir
			cmd.Run()
		}

		m.logger.Info("插件更新完成（Git）", "plugin_id", pluginID, "dir", p.Dir)
		return nil
	}

	// 否则卸载后按仓库方式重装。
	source := p.Source
	if err := m.Uninstall(pluginID); err != nil {
		return err
	}
	if source == "registry" {
		err := m.Install(pluginID, "")
		if err == nil {
			m.logger.Info("插件更新完成（重装）", "plugin_id", pluginID)
		}
		return err
	}
	m.logger.Warn("插件更新失败，非 Git 仓库插件不支持自动更新", "plugin_id", pluginID, "source", source)
	return fmt.Errorf("非 Git 仓库插件无法自动更新，请手动卸载重装")
}

// 检查安装前冲突。
func (m *Manager) CheckConflicts(pluginID string) []string {
	var conflicts []string
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.plugins[pluginID]; exists {
		conflicts = append(conflicts, fmt.Sprintf("插件 %s 已安装", pluginID))
	}

	return conflicts
}

// 以下为内部实现方法。

// 扫描插件目录并与内存状态及 OpenClaw 配置对齐。
func (m *Manager) scanInstalledPlugins() {
	scannedDirs := map[string]struct{}{}
	m.scanPluginDir(m.pluginsDir, scannedDirs)
	m.pruneDetachedPlugins()
}

// 扫描单个插件目录并回填内存状态。
func (m *Manager) scanPluginDir(baseDir string, scannedDirs map[string]struct{}) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		m.logger.Debug("跳过不可用插件目录", "dir", baseDir, "error", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pluginDir := filepath.Join(baseDir, entry.Name())
		if _, ok := scannedDirs[pluginDir]; ok {
			continue
		}
		scannedDirs[pluginDir] = struct{}{}

		meta, err := m.readPluginMeta(pluginDir)
		if err != nil {
			m.logger.Debug("跳过无效插件目录", "dir", pluginDir, "error", err)
			continue
		}
		m.upsertScannedPlugin(meta, pluginDir)
	}
}

// 清理不属于当前 OpenClaw 实例插件目录的历史插件状态。
func (m *Manager) pruneDetachedPlugins() {
	baseDir := filepath.Clean(m.pluginsDir)
	m.mu.Lock()
	defer m.mu.Unlock()

	for pluginID, plugin := range m.plugins {
		pluginDir := filepath.Clean(strings.TrimSpace(plugin.Dir))
		if pluginDir == "" {
			delete(m.plugins, pluginID)
			continue
		}
		rel, err := filepath.Rel(baseDir, pluginDir)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			delete(m.plugins, pluginID)
			continue
		}
		if !pathExists(pluginDir) {
			delete(m.plugins, pluginID)
		}
	}
}

// 将扫描到的插件元数据合并到当前状态，并同步回 openclaw.json。
func (m *Manager) upsertScannedPlugin(meta *PluginMeta, pluginDir string) {
	enabled := true
	source := "local"
	version := meta.Version

	m.mu.Lock()
	if _, exists := m.plugins[meta.ID]; !exists {
		m.plugins[meta.ID] = &InstalledPlugin{
			PluginMeta:  *meta,
			Enabled:     true,
			InstalledAt: time.Now().Format(time.RFC3339),
			Source:      "local",
			Dir:         pluginDir,
		}
	} else {
		// 用磁盘数据更新目录路径与元数据。
		m.plugins[meta.ID].Dir = pluginDir
		m.plugins[meta.ID].PluginMeta = *meta
		enabled = m.plugins[meta.ID].Enabled
		source = m.plugins[meta.ID].Source
	}
	m.mu.Unlock()

	if err := m.syncOpenClawPluginState(meta.ID, pluginDir, enabled, source, version); err != nil {
		m.logger.Warn("扫描插件后同步 OpenClaw 状态失败", "plugin_id", meta.ID, "error", err)
	}
}

// 从 openclaw.json 恢复插件索引，避免目录布局变化时列表为空。
func (m *Manager) hydratePluginsFromOpenClawConfig() {
	ocConfig, err := m.cfg.ReadOpenClawJSON()
	if err != nil || ocConfig == nil {
		return
	}
	plugins, _ := ocConfig["plugins"].(map[string]interface{})
	entries, _ := plugins["entries"].(map[string]interface{})
	installs, _ := plugins["installs"].(map[string]interface{})
	if len(entries) == 0 && len(installs) == 0 {
		return
	}

	for pluginID, rawEntry := range entries {
		entry, _ := rawEntry.(map[string]interface{})
		enabled, hasEnabled := entry["enabled"].(bool)
		if !hasEnabled {
			enabled = true
		}

		install, _ := installs[pluginID].(map[string]interface{})
		installPath, _ := install["installPath"].(string)
		version, _ := install["version"].(string)
		source, _ := install["source"].(string)
		installedAt, _ := install["installedAt"].(string)
		resolvedDir := m.resolvePluginInstallPath(pluginID, installPath)

		m.mu.Lock()
		plugin, exists := m.plugins[pluginID]
		if !exists {
			plugin = &InstalledPlugin{
				PluginMeta: PluginMeta{
					ID:      pluginID,
					Name:    pluginID,
					Version: version,
				},
				Enabled:     enabled,
				InstalledAt: installedAt,
				Source:      source,
				Dir:         resolvedDir,
			}
			if plugin.InstalledAt == "" {
				plugin.InstalledAt = time.Now().Format(time.RFC3339)
			}
			if plugin.Source == "" {
				plugin.Source = "config"
			}
			m.plugins[pluginID] = plugin
		} else {
			plugin.Enabled = enabled
			if plugin.Version == "" {
				plugin.Version = version
			}
			if plugin.Source == "" {
				plugin.Source = source
			}
			if plugin.Dir == "" {
				plugin.Dir = resolvedDir
			}
		}
		m.mu.Unlock()

		if resolvedDir != "" {
			if meta, metaErr := m.readPluginMeta(resolvedDir); metaErr == nil {
				m.mu.Lock()
				current := m.plugins[pluginID]
				current.PluginMeta = *meta
				current.Enabled = enabled
				if current.Source == "" {
					current.Source = source
				}
				if current.Source == "" {
					current.Source = "config"
				}
				current.Dir = resolvedDir
				m.mu.Unlock()
			}
		}
	}
}

// 解析插件真实安装目录，仅接受显式 installPath。
func (m *Manager) resolvePluginInstallPath(pluginID, installPath string) string {
	if pathExists(installPath) {
		return filepath.Clean(installPath)
	}

	if strings.TrimSpace(installPath) != "" {
		return filepath.Clean(installPath)
	}
	return ""
}

// 判断路径是否存在。
func pathExists(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// 读取插件元数据，优先 openclaw.plugin.json，并补充 package.json 中的通用字段。
func (m *Manager) readPluginMeta(dir string) (*PluginMeta, error) {
	var meta PluginMeta

	metaPath := filepath.Join(dir, "openclaw.plugin.json")
	metaData, metaErr := os.ReadFile(metaPath)
	if metaErr == nil {
		if err := json.Unmarshal(metaData, &meta); err != nil {
			return nil, fmt.Errorf("parse openclaw.plugin.json failed: %w", err)
		}
	}

	pkgPath := filepath.Join(dir, "package.json")
	pkgData, pkgErr := os.ReadFile(pkgPath)
	if pkgErr == nil {
		var pkgMeta PluginMeta
		if err := json.Unmarshal(pkgData, &pkgMeta); err != nil {
			return nil, fmt.Errorf("parse package.json failed: %w", err)
		}
		mergePluginMeta(&meta, &pkgMeta)
	}

	if metaErr != nil && pkgErr != nil {
		return nil, fmt.Errorf("no openclaw.plugin.json or package.json found")
	}
	if meta.ID == "" {
		meta.ID = filepath.Base(dir)
	}
	if meta.Name == "" {
		meta.Name = meta.ID
	}
	return &meta, nil
}

// mergePluginMeta 用 package.json 中的通用字段补齐扩展元数据。
func mergePluginMeta(dst, src *PluginMeta) {
	if dst == nil || src == nil {
		return
	}
	if strings.TrimSpace(dst.ID) == "" {
		dst.ID = src.ID
	}
	if strings.TrimSpace(dst.Name) == "" {
		dst.Name = src.Name
	}
	if strings.TrimSpace(dst.Version) == "" {
		dst.Version = src.Version
	}
	if strings.TrimSpace(dst.Author) == "" {
		dst.Author = src.Author
	}
	if strings.TrimSpace(dst.Description) == "" {
		dst.Description = src.Description
	}
	if strings.TrimSpace(dst.Homepage) == "" {
		dst.Homepage = src.Homepage
	}
	if strings.TrimSpace(dst.Repository) == "" {
		dst.Repository = src.Repository
	}
	if strings.TrimSpace(dst.License) == "" {
		dst.License = src.License
	}
	if strings.TrimSpace(dst.Category) == "" {
		dst.Category = src.Category
	}
	if len(dst.Tags) == 0 {
		dst.Tags = src.Tags
	}
	if strings.TrimSpace(dst.Icon) == "" {
		dst.Icon = src.Icon
	}
	if strings.TrimSpace(dst.MinOpenClaw) == "" {
		dst.MinOpenClaw = src.MinOpenClaw
	}
	if strings.TrimSpace(dst.MinPanel) == "" {
		dst.MinPanel = src.MinPanel
	}
	if strings.TrimSpace(dst.EntryPoint) == "" {
		dst.EntryPoint = src.EntryPoint
	}
	if len(dst.ConfigSchema) == 0 {
		dst.ConfigSchema = src.ConfigSchema
	}
	if len(dst.Dependencies) == 0 {
		dst.Dependencies = src.Dependencies
	}
	if len(dst.Permissions) == 0 {
		dst.Permissions = src.Permissions
	}
}

// 从 plugins.json 恢复已安装插件状态。
func (m *Manager) loadPluginsState() {
	data, err := os.ReadFile(m.configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			m.logger.Warn("读取插件状态文件失败", "path", m.configFile, "error", err)
		}
		return
	}
	var plugins map[string]*InstalledPlugin
	if json.Unmarshal(data, &plugins) == nil {
		m.plugins = plugins
		m.logger.Debug("已加载插件状态文件", "path", m.configFile, "count", len(m.plugins))
		return
	}
	m.logger.Warn("解析插件状态文件失败", "path", m.configFile)
}

// 在持有读锁的情况下持久化插件状态。
func (m *Manager) savePluginsState() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.savePluginsStateUnlocked()
}

// 将当前插件状态写入 plugins.json（调用方需自行保证并发安全）。
func (m *Manager) savePluginsStateUnlocked() {
	data, _ := json.MarshalIndent(m.plugins, "", "  ")
	if err := os.WriteFile(m.configFile, data, 0o644); err != nil {
		m.logger.Warn("写入插件状态文件失败", "path", m.configFile, "error", err)
	}
}

// 将仓库数据缓存到本地，供离线回退使用。
func (m *Manager) cacheRegistry(reg *Registry) {
	data, _ := json.MarshalIndent(reg, "", "  ")
	cachePath := filepath.Join(m.cfg.DataDir, "plugin-registry-cache.json")
	if err := os.WriteFile(cachePath, data, 0o644); err != nil {
		m.logger.Warn("写入插件仓库缓存失败", "path", cachePath, "error", err)
	}
}

// 读取本地仓库缓存，失败时返回 nil。
func (m *Manager) loadCachedRegistry() *Registry {
	cachePath := filepath.Join(m.cfg.DataDir, "plugin-registry-cache.json")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if !os.IsNotExist(err) {
			m.logger.Warn("读取插件仓库缓存失败", "path", cachePath, "error", err)
		}
		return nil
	}
	var reg Registry
	if json.Unmarshal(data, &reg) == nil {
		m.logger.Debug("已加载插件仓库缓存", "path", cachePath, "plugin_count", len(reg.Plugins))
		return &reg
	}
	m.logger.Warn("解析插件仓库缓存失败", "path", cachePath)
	return nil
}

// 将插件启用状态与安装信息同步到 openclaw.json。
func (m *Manager) syncOpenClawPluginState(pluginID, installPath string, enabled bool, source string, version string) error {
	ocConfig, err := m.cfg.ReadOpenClawJSON()
	if err != nil || ocConfig == nil {
		ocConfig = map[string]interface{}{}
	}
	pl, _ := ocConfig["plugins"].(map[string]interface{})
	if pl == nil {
		pl = map[string]interface{}{}
		ocConfig["plugins"] = pl
	}
	ent, _ := pl["entries"].(map[string]interface{})
	if ent == nil {
		ent = map[string]interface{}{}
		pl["entries"] = ent
	}
	ent[pluginID] = map[string]interface{}{"enabled": enabled}

	ins, _ := pl["installs"].(map[string]interface{})
	if ins == nil {
		ins = map[string]interface{}{}
		pl["installs"] = ins
	}
	item, _ := ins[pluginID].(map[string]interface{})
	if item == nil {
		item = map[string]interface{}{}
		ins[pluginID] = item
	}
	if installPath != "" {
		item["installPath"] = installPath
	}
	if normalized := normalizeOpenClawInstallSource(source); normalized != "" {
		item["source"] = normalized
	}
	if version != "" {
		item["version"] = version
	}
	if _, ok := item["installedAt"]; !ok {
		item["installedAt"] = time.Now().UTC().Format(time.RFC3339)
	}
	return m.cfg.WriteOpenClawJSON(ocConfig)
}

// 将内部来源枚举映射为 OpenClaw 兼容值。
func normalizeOpenClawInstallSource(source string) string {
	switch strings.TrimSpace(strings.ToLower(source)) {
	case "npm":
		return "npm"
	case "archive":
		return "archive"
	case "path", "local", "registry", "custom", "github", "git":
		return "path"
	default:
		return "path"
	}
}

// 从 openclaw.json 中移除指定插件的状态记录。
func (m *Manager) removeOpenClawPluginState(pluginID string) error {
	ocConfig, err := m.cfg.ReadOpenClawJSON()
	if err != nil || ocConfig == nil {
		return err
	}
	pl, _ := ocConfig["plugins"].(map[string]interface{})
	if pl == nil {
		return nil
	}
	if ent, ok := pl["entries"].(map[string]interface{}); ok {
		delete(ent, pluginID)
	}
	if ins, ok := pl["installs"].(map[string]interface{}); ok {
		delete(ins, pluginID)
	}
	return m.cfg.WriteOpenClawJSON(ocConfig)
}

// 通过 npm 全局安装插件包，并在镜像失败时回退官方源。
func (m *Manager) installFromNpm(pkgName string) error {
	cmd := exec.Command("npm", "install", "-g", pkgName+"@latest", "--registry=https://registry.npmmirror.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// 镜像失败时回退官方源重试。
		cmd2 := exec.Command("npm", "install", "-g", pkgName+"@latest")
		out2, err2 := cmd2.CombinedOutput()
		if err2 != nil {
			return fmt.Errorf("%s\n%s", string(out), string(out2))
		}
	}
	return nil
}

// 使用浅克隆将插件仓库拉取到目标目录。
func (m *Manager) installFromGit(gitURL, dest string) error {
	cmd := exec.Command("git", "clone", "--depth=1", gitURL, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(out))
	}
	return nil
}

// 下载并解压 zip/tar.gz 插件归档到目标目录。
func (m *Manager) installFromArchive(url, dest string) error {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	os.MkdirAll(dest, 0755)
	tmpFile := filepath.Join(dest, "plugin-archive.tmp")
	f, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	io.Copy(f, resp.Body)
	f.Close()

	// 按文件后缀选择解压方式。
	if strings.HasSuffix(url, ".zip") {
		cmd := exec.Command("unzip", "-o", tmpFile, "-d", dest)
		cmd.Run()
	} else if strings.HasSuffix(url, ".tar.gz") {
		cmd := exec.Command("tar", "-xzf", tmpFile, "-C", dest)
		cmd.Run()
	}

	os.Remove(tmpFile)
	return nil
}

// 递归复制目录。
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}
