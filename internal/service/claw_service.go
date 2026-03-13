package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	clawchat "github.com/chenyang-zz/boxify/internal/claw/chat"
	clawmonitor "github.com/chenyang-zz/boxify/internal/claw/monitor"
	clawplugin "github.com/chenyang-zz/boxify/internal/claw/plugin"
	clawprocess "github.com/chenyang-zz/boxify/internal/claw/process"
	clawskill "github.com/chenyang-zz/boxify/internal/claw/skill"
	clawtaskman "github.com/chenyang-zz/boxify/internal/claw/taskman"
	clawupdate "github.com/chenyang-zz/boxify/internal/claw/update"
	"github.com/chenyang-zz/boxify/internal/types"
	"github.com/chenyang-zz/boxify/pkg/conversationstore"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const defaultClawManagerPort = 19527
const boxifyConfigFileName = "boxify.json"

type boxifyLocalConfig struct {
	ChatSharedToken string `json:"chatSharedToken,omitempty"` // Boxify 本地聊天共享令牌。
}

// ClawService 提供 OpenClaw 相关能力（进程、配置、插件、任务、更新、NapCat 监控）。
type ClawService struct {
	BaseService
	manager         *clawprocess.Manager         // OpenClaw 进程管理器
	pluginCfg       *clawplugin.Config           // OpenClaw 配置读写器
	pluginManager   *clawplugin.Manager          // 插件管理器
	skillManager    *clawskill.Manager           // 技能管理器
	taskManager     *clawtaskman.Manager         // 任务管理器
	updater         *clawupdate.Updater          // 面板更新器
	napcatMonitor   *clawmonitor.NapCatMonitor   // NapCat 监控器
	chatCoordinator *clawchat.ChannelCoordinator // Boxify 聊天 channel 协调器

	openClawDir        string // OpenClaw 配置目录
	openClawApp        string // OpenClaw 应用目录
	dataDir            string // Boxify 数据目录
	managerPort        int    // 本地管理端口
	pluginPort         int    // 预期插件入站监听端口
	chatToken          string // 插件 inbox 共享令牌
	chatTokenGenerated bool   // 是否在本次启动中首次生成了共享令牌
}

// NewClawService 创建 Claw 服务。
func NewClawService(deps *ServiceDeps) *ClawService {
	return &ClawService{BaseService: NewBaseService(deps)}
}

// ServiceStartup 服务启动。
func (s *ClawService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.SetContext(ctx)
	s.initRuntimeContext()
	s.rebuildManagers()
	if s.napcatMonitor != nil {
		s.napcatMonitor.Start()
	}
	s.Logger().Info("服务启动", "service", "ClawService", "open_claw_dir", s.openClawDir, "data_dir", s.dataDir)
	return nil
}

// ServiceShutdown 服务关闭。
func (s *ClawService) ServiceShutdown() error {
	if s.napcatMonitor != nil {
		s.napcatMonitor.Stop()
	}
	if s.manager != nil {
		s.manager.StopAll()
	}
	s.Logger().Info("服务关闭", "service", "ClawService")
	return nil
}

// Configure 更新 Claw 管理配置并重建依赖管理器。
func (s *ClawService) Configure(cfg types.ClawManagerConfig) *types.BaseResult {
	s.openClawDir = strings.TrimSpace(cfg.OpenClawDir)
	s.openClawApp = strings.TrimSpace(cfg.OpenClawApp)
	s.managerPort = cfg.ManagerPort
	if s.openClawDir == "" {
		s.openClawDir = defaultOpenClawDir()
	}
	if s.managerPort <= 0 {
		s.managerPort = defaultClawManagerPort
	}
	s.rebuildManagers()
	return &types.BaseResult{Success: true, Message: "Claw 配置已更新"}
}

// Start 启动 OpenClaw。
func (s *ClawService) Start() *types.BaseResult {
	if err := s.manager.Start(); err != nil {
		return &types.BaseResult{Success: false, Message: err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "OpenClaw 启动成功"}
}

// Stop 停止 OpenClaw。
func (s *ClawService) Stop() *types.BaseResult {
	if err := s.manager.Stop(); err != nil {
		return &types.BaseResult{Success: false, Message: err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "OpenClaw 已停止"}
}

// Restart 重启 OpenClaw。
func (s *ClawService) Restart() *types.BaseResult {
	if err := s.manager.Restart(); err != nil {
		return &types.BaseResult{Success: false, Message: err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "OpenClaw 重启成功"}
}

// StartProcess 启动 OpenClaw 进程（兼容 ClawPanel 接口名）。
func (s *ClawService) StartProcess() *types.BaseResult {
	return s.Start()
}

// StopProcess 停止 OpenClaw 进程（兼容 ClawPanel 接口名）。
func (s *ClawService) StopProcess() *types.BaseResult {
	return s.Stop()
}

// RestartProcess 重启 OpenClaw 进程（兼容 ClawPanel 接口名）。
func (s *ClawService) RestartProcess() *types.BaseResult {
	return s.Restart()
}

// GetStatus 获取 OpenClaw 状态。
func (s *ClawService) GetStatus() *types.ClawStatusResult {
	status := s.manager.GetStatus()
	return &types.ClawStatusResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取状态成功"},
		Data: &types.ClawStatus{
			Running:           status.Running,
			PID:               status.PID,
			StartedAt:         status.StartedAt,
			Uptime:            status.Uptime,
			ExitCode:          status.ExitCode,
			Daemonized:        status.Daemonized,
			ManagedExternally: status.ManagedExternally,
		},
	}
}

// GetOverview 获取 OpenClaw 概览信息。
func (s *ClawService) GetOverview() *types.ClawOverviewResult {
	status := s.manager.GetStatus()
	ocConfig, _ := s.pluginCfg.ReadOpenClawJSON()
	if ocConfig == nil {
		ocConfig = map[string]interface{}{}
	}

	pluginEnabledMap := buildInstalledPluginEnabledMap(s.pluginManager.ListInstalled())
	channels := buildOverviewChannels(ocConfig, pluginEnabledMap)
	connectedChannels := filterEnabledOverviewChannels(channels)
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)

	data := &types.ClawOverviewData{
		SystemStatus:   resolveOverviewSystemStatus(status),
		ActiveChannels: len(connectedChannels),
		AIModel:        resolveOverviewModel(ocConfig),
		Uptime:         formatOverviewUptime(status.Running, status.Uptime),
		MemoryUsage:    formatOverviewMemory(mem.Alloc),
		TodayMessages:  countTodayTasks(s.taskManager.GetRecentTasks()),
		Channels:       connectedChannels,
	}
	return &types.ClawOverviewResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取概览成功"},
		Data:       data,
	}
}

// GetSkillPlugins 获取技能中心中的已安装插件列表。
func (s *ClawService) GetSkillPlugins() *types.ClawSkillPluginsResult {
	if s.pluginManager == nil {
		return &types.ClawSkillPluginsResult{
			BaseResult: types.BaseResult{Success: true, Message: "获取插件列表成功"},
			Plugins:    []types.ClawSkillPlugin{},
		}
	}

	plugins := s.pluginManager.ListSkillCenterPlugins()
	result := make([]types.ClawSkillPlugin, 0, len(plugins))
	for _, plugin := range plugins {
		result = append(result, types.ClawSkillPlugin{
			ID:          plugin.ID,
			Name:        plugin.Name,
			Description: plugin.Description,
			Version:     plugin.Version,
			Enabled:     plugin.Enabled,
			Source:      plugin.Source,
			InstalledAt: plugin.InstalledAt,
			Path:        plugin.Path,
		})
	}

	return &types.ClawSkillPluginsResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取插件列表成功"},
		Plugins:    result,
	}
}

// GetSkills 获取技能中心所需的技能列表。
func (s *ClawService) GetSkills() *types.ClawSkillsResult {
	if s.skillManager == nil {
		return &types.ClawSkillsResult{
			BaseResult: types.BaseResult{Success: true, Message: "获取技能列表成功"},
			Skills:     []types.ClawSkill{},
		}
	}

	items := s.skillManager.List()
	return &types.ClawSkillsResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取技能列表成功"},
		Skills:     items,
	}
}

// ToggleSkill 切换技能启用状态。
func (s *ClawService) ToggleSkill(id string, enabled bool) *types.BaseResult {
	id = strings.TrimSpace(id)
	if id == "" {
		return &types.BaseResult{Success: false, Message: "技能 ID 不能为空"}
	}
	if s.skillManager == nil {
		s.skillManager = clawskill.NewManager(s.pluginCfg, s.openClawDir, s.openClawApp, s.Logger())
	}
	if err := s.skillManager.Toggle(id, enabled); err != nil {
		return &types.BaseResult{Success: false, Message: "更新技能状态失败: " + err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "技能状态已更新"}
}

// CheckOpenClaw 检查 OpenClaw 是否已安装并完成基础配置。
func (s *ClawService) CheckOpenClaw() *types.ClawOpenClawCheckResult {
	bin := strings.TrimSpace(clawprocess.DetectOpenClawBinaryPath())
	cfgPath := filepath.Join(s.openClawDir, "openclaw.json")
	configured := false
	if st, err := os.Stat(cfgPath); err == nil && !st.IsDir() {
		if s.pluginCfg != nil {
			if cfg, readErr := s.pluginCfg.ReadOpenClawJSON(); readErr == nil && cfg != nil {
				configured = true
			}
		}
	}
	return &types.ClawOpenClawCheckResult{
		BaseResult: types.BaseResult{Success: true, Message: "OpenClaw 检查完成"},
		Installed:  bin != "",
		Configured: configured,
		BinaryPath: bin,
		ConfigPath: cfgPath,
	}
}

// resolveOverviewSystemStatus 根据进程状态推导概览状态标签。
func resolveOverviewSystemStatus(status clawprocess.Status) string {
	if status.Running {
		return "normal"
	}
	return "warning"
}

// resolveOverviewModel 从 openclaw 配置中提取默认模型名。
func resolveOverviewModel(ocConfig map[string]interface{}) string {
	candidates := []string{
		readNestedString(ocConfig, "agents", "defaults", "model", "primary"),
		readNestedString(ocConfig, "agents", "defaults", "chatModel"),
		readNestedString(ocConfig, "models", "default"),
		readNestedString(ocConfig, "models", "defaults", "model"),
		readNestedString(ocConfig, "models", "current"),
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) != "" {
			return strings.TrimSpace(candidate)
		}
	}

	models, _ := ocConfig["models"].(map[string]interface{})
	if models == nil {
		return "-"
	}
	providers, _ := models["providers"].(map[string]interface{})
	if providers == nil || len(providers) == 0 {
		return "-"
	}

	providerKeys := sortedMapKeys(providers)
	for _, providerKey := range providerKeys {
		provider, _ := providers[providerKey].(map[string]interface{})
		if provider == nil {
			continue
		}
		model := strings.TrimSpace(readFirstString(provider, "defaultModel", "model", "chatModel"))
		if model != "" {
			return fmt.Sprintf("%s/%s", providerKey, model)
		}
	}

	return providerKeys[0]
}

// buildOverviewChannels 聚合 channels/plugins 配置为概览卡片列表。
func buildOverviewChannels(ocConfig map[string]interface{}, pluginEnabledMap map[string]bool) []types.ClawOverviewChannel {
	channels, _ := ocConfig["channels"].(map[string]interface{})
	channelKeys := sortedMapKeys(channels)
	builtInByCanonicalID := make(map[string]types.ClawOverviewChannel, len(channelKeys))
	for _, id := range channelKeys {
		cfg, _ := channels[id].(map[string]interface{})
		item := types.ClawOverviewChannel{
			ID:        id,
			Name:      resolveOverviewChannelName(id, cfg),
			Type:      "built-in",
			Status:    resolveOverviewChannelStatus(cfg),
			ManagedBy: "由网关管理",
		}
		builtInByCanonicalID[normalizeOverviewChannelID(id)] = item
	}

	pluginsRoot, _ := ocConfig["plugins"].(map[string]interface{})
	pluginEntries, _ := pluginsRoot["entries"].(map[string]interface{})
	pluginKeys := sortedMapKeys(pluginEntries)
	pluginByCanonicalID := make(map[string]types.ClawOverviewChannel, len(pluginKeys))
	for _, id := range pluginKeys {
		cfg, _ := pluginEntries[id].(map[string]interface{})
		candidate := types.ClawOverviewChannel{
			ID:        id,
			Name:      resolveOverviewChannelName(id, cfg),
			Type:      "plugin",
			Status:    resolveOverviewPluginStatus(id, cfg, pluginEnabledMap),
			ManagedBy: "由插件管理",
		}
		canonicalID := normalizeOverviewChannelID(id)
		current, exists := pluginByCanonicalID[canonicalID]
		if !exists || shouldReplacePluginChannel(current, candidate) {
			pluginByCanonicalID[canonicalID] = candidate
		}
	}

	canonicalIDs := make([]string, 0, len(builtInByCanonicalID)+len(pluginByCanonicalID))
	seen := make(map[string]struct{}, len(builtInByCanonicalID)+len(pluginByCanonicalID))
	for id := range builtInByCanonicalID {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		canonicalIDs = append(canonicalIDs, id)
	}
	for id := range pluginByCanonicalID {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		canonicalIDs = append(canonicalIDs, id)
	}
	sort.Strings(canonicalIDs)

	items := make([]types.ClawOverviewChannel, 0, len(canonicalIDs))
	for _, canonicalID := range canonicalIDs {
		// 规则：只要 channel 中存在该频道，就视为内置频道，状态仅由 channel.enabled 控制。
		if builtIn, exists := builtInByCanonicalID[canonicalID]; exists {
			items = append(items, builtIn)
			continue
		}
		// 规则：仅当不存在内置 channel 时，才将插件实现展示为插件频道，状态由插件配置控制。
		if plugin, exists := pluginByCanonicalID[canonicalID]; exists {
			items = append(items, plugin)
		}
	}

	return items
}

// buildInstalledPluginEnabledMap 构建插件启用状态索引（key=pluginID）。
func buildInstalledPluginEnabledMap(installed []*clawplugin.InstalledPlugin) map[string]bool {
	if len(installed) == 0 {
		return map[string]bool{}
	}
	statuses := make(map[string]bool, len(installed))
	for _, plugin := range installed {
		if plugin == nil {
			continue
		}
		id := strings.TrimSpace(plugin.ID)
		if id == "" {
			continue
		}
		statuses[id] = plugin.Enabled
	}
	return statuses
}

// filterEnabledOverviewChannels 仅保留已启用的频道用于“已连接频道”展示。
func filterEnabledOverviewChannels(channels []types.ClawOverviewChannel) []types.ClawOverviewChannel {
	if len(channels) == 0 {
		return channels
	}
	enabled := make([]types.ClawOverviewChannel, 0, len(channels))
	for _, channel := range channels {
		if channel.Status == "enabled" {
			enabled = append(enabled, channel)
		}
	}
	return enabled
}

// shouldReplacePluginChannel 判断插件候选是否应覆盖当前插件实现。
func shouldReplacePluginChannel(current types.ClawOverviewChannel, candidate types.ClawOverviewChannel) bool {
	currentEnabled := current.Status == "enabled"
	candidateEnabled := candidate.Status == "enabled"
	if currentEnabled != candidateEnabled {
		return candidateEnabled
	}
	return false
}

// normalizeOverviewChannelID 将别名归并到同一规范通道 ID。
func normalizeOverviewChannelID(id string) string {
	normalized := strings.ToLower(strings.TrimSpace(id))
	switch normalized {
	case "lark", "feishu-openclaw-plugin":
		return "feishu"
	default:
		return normalized
	}
}

// resolveOverviewPluginStatus 将插件通道状态优先映射为插件管理器状态，再回退插件配置。
func resolveOverviewPluginStatus(id string, cfg map[string]interface{}, pluginEnabledMap map[string]bool) string {
	normalizedID := normalizeOverviewChannelID(id)
	matched := false
	hasEnabled := false
	for pluginID, pluginEnabled := range pluginEnabledMap {
		if normalizeOverviewChannelID(pluginID) != normalizedID {
			continue
		}
		matched = true
		if pluginEnabled {
			hasEnabled = true
		}
	}
	if matched {
		if hasEnabled {
			return "enabled"
		}
		return "disabled"
	}
	return resolveOverviewChannelStatus(cfg)
}

// resolveOverviewChannelName 优先读取通道配置中的名称字段。
func resolveOverviewChannelName(id string, cfg map[string]interface{}) string {
	if cfg == nil {
		return id
	}
	nameKeys := []string{"name", "title", "displayName"}
	for _, key := range nameKeys {
		if raw, ok := cfg[key]; ok {
			if name, castOK := raw.(string); castOK && strings.TrimSpace(name) != "" {
				return strings.TrimSpace(name)
			}
		}
	}
	return id
}

// resolveOverviewChannelStatus 将 enabled 字段转换为 enabled/disabled。
func resolveOverviewChannelStatus(cfg map[string]interface{}) string {
	if cfg == nil {
		return "enabled"
	}
	enabled := true
	if raw, ok := cfg["enabled"]; ok {
		if v, castOK := raw.(bool); castOK {
			enabled = v
		}
	}
	if enabled {
		return "enabled"
	}
	return "disabled"
}

// sortedMapKeys 返回 map 的稳定升序 key 列表。
func sortedMapKeys(m map[string]interface{}) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// countEnabledChannels 统计已启用通道数量。
func countEnabledChannels(channels []types.ClawOverviewChannel) int {
	count := 0
	for _, channel := range channels {
		if channel.Status == "enabled" {
			count++
		}
	}
	return count
}

// formatOverviewUptime 将秒数转成概览文案。
func formatOverviewUptime(running bool, seconds int64) string {
	if seconds <= 0 {
		if running {
			return "运行中"
		}
		return "未运行"
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	totalHours := seconds / 3600
	totalMinutes := seconds / 60

	if days > 0 {
		return fmt.Sprintf("%d 天", days)
	}
	if totalHours > 0 || hours > 0 {
		return fmt.Sprintf("%d 时", totalHours)
	}
	if totalMinutes > 0 {
		return fmt.Sprintf("%d 分", totalMinutes)
	}
	return "1 分内"
}

// readNestedString 按层级读取字符串值，任意层失败时返回空串。
func readNestedString(root map[string]interface{}, keys ...string) string {
	if root == nil || len(keys) == 0 {
		return ""
	}
	var current interface{} = root
	for _, key := range keys {
		nextMap, ok := current.(map[string]interface{})
		if !ok || nextMap == nil {
			return ""
		}
		current = nextMap[key]
	}
	v, ok := current.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(v)
}

// readFirstString 读取候选字段中第一个非空字符串值。
func readFirstString(m map[string]interface{}, keys ...string) string {
	if m == nil {
		return ""
	}
	for _, key := range keys {
		if raw, ok := m[key]; ok {
			if v, castOK := raw.(string); castOK && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
}

// formatOverviewMemory 将字节数转成 MB 文案。
func formatOverviewMemory(bytes uint64) string {
	mb := float64(bytes) / 1024 / 1024
	return fmt.Sprintf("%.1f MB", mb)
}

// countTodayTasks 统计今日创建任务数，用于概览“今日消息”展示。
func countTodayTasks(tasks []*clawtaskman.Task) int {
	if len(tasks) == 0 {
		return 0
	}
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	count := 0
	for _, task := range tasks {
		if task == nil {
			continue
		}
		if !task.CreatedAt.Before(start) && task.CreatedAt.Before(end) {
			count++
		}
	}
	return count
}

// ProcessStatus 获取进程状态（兼容 ClawPanel 接口名）。
func (s *ClawService) ProcessStatus() *types.ClawProcessStatusResult {
	return &types.ClawProcessStatusResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取进程状态成功"},
		Status:     s.manager.GetStatus(),
	}
}

// GetLogs 获取最近日志。
func (s *ClawService) GetLogs(n int) *types.ClawLogsResult {
	lines := s.manager.GetLogs(n)
	return &types.ClawLogsResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取日志成功"},
		Data:       &types.ClawLogsData{Lines: lines},
	}
}

// GetOpenClawConfig 获取 OpenClaw 配置。
func (s *ClawService) GetOpenClawConfig() *types.ClawOpenConfigResult {
	cfg, err := s.pluginCfg.ReadOpenClawJSON()
	if err != nil {
		return &types.ClawOpenConfigResult{
			BaseResult: types.BaseResult{Success: true, Message: "获取 OpenClaw 配置成功"},
			Config:     map[string]interface{}{},
		}
	}
	return &types.ClawOpenConfigResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取 OpenClaw 配置成功"},
		Config:     cfg,
	}
}

// SaveOpenClawConfig 保存 OpenClaw 配置。
func (s *ClawService) SaveOpenClawConfig(cfg map[string]interface{}) *types.BaseResult {
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	if err := s.pluginCfg.WriteOpenClawJSON(cfg); err != nil {
		return &types.BaseResult{Success: false, Message: "保存 OpenClaw 配置失败: " + err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "OpenClaw 配置已保存"}
}

// GetModels 获取模型配置。
func (s *ClawService) GetModels() *types.ClawModelsResult {
	ocConfig, err := s.pluginCfg.ReadOpenClawJSON()
	if err != nil {
		return &types.ClawModelsResult{
			BaseResult: types.BaseResult{Success: true, Message: "获取模型配置成功"},
			Providers:  map[string]interface{}{},
			Defaults:   map[string]interface{}{},
		}
	}
	models, _ := ocConfig["models"].(map[string]interface{})
	if models == nil {
		models = map[string]interface{}{}
	}
	providers, _ := models["providers"].(map[string]interface{})
	if providers == nil {
		providers = map[string]interface{}{}
	}
	agents, _ := ocConfig["agents"].(map[string]interface{})
	defaults := map[string]interface{}{}
	if agents != nil {
		if v, ok := agents["defaults"].(map[string]interface{}); ok && v != nil {
			defaults = v
		}
	}
	return &types.ClawModelsResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取模型配置成功"},
		Providers:  providers,
		Defaults:   defaults,
	}
}

// SaveModels 保存模型配置。
func (s *ClawService) SaveModels(providers map[string]interface{}) *types.BaseResult {
	ocConfig, _ := s.pluginCfg.ReadOpenClawJSON()
	if ocConfig == nil {
		ocConfig = map[string]interface{}{}
	}
	models, _ := ocConfig["models"].(map[string]interface{})
	if models == nil {
		models = map[string]interface{}{}
	}
	models["providers"] = providers
	ocConfig["models"] = models
	if err := s.pluginCfg.WriteOpenClawJSON(ocConfig); err != nil {
		return &types.BaseResult{Success: false, Message: "保存模型配置失败: " + err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "模型配置已保存"}
}

// GetChannels 获取通道与插件配置。
func (s *ClawService) GetChannels() *types.ClawChannelsResult {
	ocConfig, err := s.pluginCfg.ReadOpenClawJSON()
	if err != nil {
		return &types.ClawChannelsResult{
			BaseResult: types.BaseResult{Success: true, Message: "获取通道配置成功"},
			Channels:   map[string]interface{}{},
			Plugins:    map[string]interface{}{},
		}
	}
	channels, _ := ocConfig["channels"].(map[string]interface{})
	plugins, _ := ocConfig["plugins"].(map[string]interface{})
	if channels == nil {
		channels = map[string]interface{}{}
	}
	if plugins == nil {
		plugins = map[string]interface{}{}
	}
	return &types.ClawChannelsResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取通道配置成功"},
		Channels:   channels,
		Plugins:    plugins,
	}
}

// SaveChannel 保存指定通道配置。
func (s *ClawService) SaveChannel(id string, payload map[string]interface{}) *types.BaseResult {
	id = strings.TrimSpace(id)
	if id == "" {
		return &types.BaseResult{Success: false, Message: "通道 ID 不能为空"}
	}
	ocConfig, _ := s.pluginCfg.ReadOpenClawJSON()
	if ocConfig == nil {
		ocConfig = map[string]interface{}{}
	}
	channels, _ := ocConfig["channels"].(map[string]interface{})
	if channels == nil {
		channels = map[string]interface{}{}
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	channels[id] = payload
	ocConfig["channels"] = channels
	if err := s.pluginCfg.WriteOpenClawJSON(ocConfig); err != nil {
		return &types.BaseResult{Success: false, Message: "保存通道配置失败: " + err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "通道配置已保存"}
}

// SavePlugin 保存指定插件配置。
func (s *ClawService) SavePlugin(id string, payload map[string]interface{}) *types.BaseResult {
	id = strings.TrimSpace(id)
	if id == "" {
		return &types.BaseResult{Success: false, Message: "插件 ID 不能为空"}
	}
	ocConfig, _ := s.pluginCfg.ReadOpenClawJSON()
	if ocConfig == nil {
		ocConfig = map[string]interface{}{}
	}
	plugins, _ := ocConfig["plugins"].(map[string]interface{})
	if plugins == nil {
		plugins = map[string]interface{}{}
	}
	entries, _ := plugins["entries"].(map[string]interface{})
	if entries == nil {
		entries = map[string]interface{}{}
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	entries[id] = payload
	plugins["entries"] = entries
	ocConfig["plugins"] = plugins
	if err := s.pluginCfg.WriteOpenClawJSON(ocConfig); err != nil {
		return &types.BaseResult{Success: false, Message: "保存插件配置失败: " + err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "插件配置已保存"}
}

// ToggleChannel 切换通道启用状态。
func (s *ClawService) ToggleChannel(id string, enabled bool) *types.BaseResult {
	id = strings.TrimSpace(id)
	if id == "" {
		return &types.BaseResult{Success: false, Message: "通道 ID 不能为空"}
	}
	ocConfig, _ := s.pluginCfg.ReadOpenClawJSON()
	if ocConfig == nil {
		ocConfig = map[string]interface{}{}
	}
	channels, _ := ocConfig["channels"].(map[string]interface{})
	if channels == nil {
		channels = map[string]interface{}{}
	}
	channelCfg, _ := channels[id].(map[string]interface{})
	if channelCfg == nil {
		channelCfg = map[string]interface{}{}
	}
	channelCfg["enabled"] = enabled
	channels[id] = channelCfg
	ocConfig["channels"] = channels
	if err := s.pluginCfg.WriteOpenClawJSON(ocConfig); err != nil {
		return &types.BaseResult{Success: false, Message: "更新通道状态失败: " + err.Error()}
	}
	if id == "qq" {
		if enabled {
			s.napcatMonitor.Resume()
		} else {
			s.napcatMonitor.Pause()
		}
	}
	return &types.BaseResult{Success: true, Message: "通道状态已更新"}
}

// GetPluginList 获取已安装插件与仓库插件。
func (s *ClawService) GetPluginList() *types.ClawPluginListResult {
	installed := s.pluginManager.ListInstalled()
	reg := s.pluginManager.GetRegistry()
	if len(reg.Plugins) == 0 {
		if fetched, err := s.pluginManager.FetchRegistry(); err == nil && fetched != nil {
			reg = fetched
		}
	}
	return &types.ClawPluginListResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取插件列表成功"},
		Installed:  installed,
		Registry:   reg.Plugins,
	}
}

// GetInstalledPlugins 获取已安装插件。
func (s *ClawService) GetInstalledPlugins() *types.ClawInstalledPluginsResult {
	return &types.ClawInstalledPluginsResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取已安装插件成功"},
		Plugins:    s.pluginManager.ListInstalled(),
	}
}

// GetPluginDetail 获取插件详情。
func (s *ClawService) GetPluginDetail(id string) *types.ClawPluginDetailResult {
	p := s.pluginManager.GetPlugin(strings.TrimSpace(id))
	if p == nil {
		return &types.ClawPluginDetailResult{BaseResult: types.BaseResult{Success: false, Message: "插件未安装"}}
	}
	return &types.ClawPluginDetailResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取插件详情成功"},
		Plugin:     p,
	}
}

// RefreshPluginRegistry 刷新插件仓库。
func (s *ClawService) RefreshPluginRegistry() *types.ClawRegistryResult {
	reg, err := s.pluginManager.FetchRegistry()
	if err != nil {
		return &types.ClawRegistryResult{BaseResult: types.BaseResult{Success: false, Message: err.Error()}}
	}
	return &types.ClawRegistryResult{
		BaseResult: types.BaseResult{Success: true, Message: "刷新插件仓库成功"},
		Registry:   reg,
	}
}

// InstallPlugin 安装插件。
func (s *ClawService) InstallPlugin(pluginID, source string) *types.BaseResult {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return &types.BaseResult{Success: false, Message: "pluginId 不能为空"}
	}
	if conflicts := s.pluginManager.CheckConflicts(pluginID); len(conflicts) > 0 {
		return &types.BaseResult{Success: false, Message: conflicts[0]}
	}
	if err := s.pluginManager.Install(pluginID, strings.TrimSpace(source)); err != nil {
		return &types.BaseResult{Success: false, Message: err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "插件安装成功"}
}

// UninstallPlugin 卸载插件。
func (s *ClawService) UninstallPlugin(id string) *types.BaseResult {
	if err := s.pluginManager.Uninstall(strings.TrimSpace(id)); err != nil {
		return &types.BaseResult{Success: false, Message: err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "插件已卸载"}
}

// TogglePlugin 切换插件启用状态。
func (s *ClawService) TogglePlugin(id string, enabled bool) *types.BaseResult {
	id = strings.TrimSpace(id)
	var err error
	if enabled {
		err = s.pluginManager.Enable(id)
	} else {
		err = s.pluginManager.Disable(id)
	}
	if err != nil {
		return &types.BaseResult{Success: false, Message: err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "插件状态已更新"}
}

// GetPluginConfig 获取插件配置与 schema。
func (s *ClawService) GetPluginConfig(id string) *types.ClawPluginConfigResult {
	cfg, schema, err := s.pluginManager.GetConfig(strings.TrimSpace(id))
	if err != nil {
		return &types.ClawPluginConfigResult{BaseResult: types.BaseResult{Success: false, Message: err.Error()}}
	}
	return &types.ClawPluginConfigResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取插件配置成功"},
		Config:     cfg,
		Schema:     schema,
	}
}

// UpdatePluginConfig 更新插件配置。
func (s *ClawService) UpdatePluginConfig(id string, cfg map[string]interface{}) *types.BaseResult {
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	if err := s.pluginManager.UpdateConfig(strings.TrimSpace(id), cfg); err != nil {
		return &types.BaseResult{Success: false, Message: err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "插件配置已更新"}
}

// GetPluginLogs 获取插件日志。
func (s *ClawService) GetPluginLogs(id string) *types.ClawPluginLogsResult {
	logs, err := s.pluginManager.GetPluginLogs(strings.TrimSpace(id))
	if err != nil {
		return &types.ClawPluginLogsResult{BaseResult: types.BaseResult{Success: false, Message: err.Error()}}
	}
	return &types.ClawPluginLogsResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取插件日志成功"},
		Logs:       logs,
	}
}

// UpdatePluginVersion 更新插件版本。
func (s *ClawService) UpdatePluginVersion(id string) *types.BaseResult {
	if err := s.pluginManager.Update(strings.TrimSpace(id)); err != nil {
		return &types.BaseResult{Success: false, Message: err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "插件更新成功"}
}

// GetTasks 获取最近任务列表。
func (s *ClawService) GetTasks() *types.ClawTasksResult {
	return &types.ClawTasksResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取任务列表成功"},
		Tasks:      s.taskManager.GetRecentTasks(),
	}
}

// GetTaskDetail 获取任务详情。
func (s *ClawService) GetTaskDetail(id string) *types.ClawTaskDetailResult {
	task := s.taskManager.GetTask(strings.TrimSpace(id))
	if task == nil {
		return &types.ClawTaskDetailResult{BaseResult: types.BaseResult{Success: false, Message: "任务不存在"}}
	}
	return &types.ClawTaskDetailResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取任务详情成功"},
		Task:       task,
	}
}

// CheckPanelUpdate 检查面板更新。
func (s *ClawService) CheckPanelUpdate() *types.ClawCheckUpdateResult {
	info, hasUpdate, err := s.updater.CheckUpdate()
	if err != nil {
		return &types.ClawCheckUpdateResult{BaseResult: types.BaseResult{Success: false, Message: err.Error()}}
	}
	return &types.ClawCheckUpdateResult{
		BaseResult:    types.BaseResult{Success: true, Message: "检查更新成功"},
		HasUpdate:     hasUpdate,
		LatestVersion: info.LatestVersion,
		ReleaseTime:   info.ReleaseTime,
		ReleaseNote:   info.ReleaseNote,
	}
}

// DoPanelUpdate 执行面板更新。
func (s *ClawService) DoPanelUpdate() *types.BaseResult {
	p := s.updater.GetProgress()
	if p.Status == "downloading" || p.Status == "verifying" || p.Status == "replacing" {
		return &types.BaseResult{Success: false, Message: "更新正在进行中"}
	}
	info, hasUpdate, err := s.updater.CheckUpdate()
	if err != nil {
		return &types.BaseResult{Success: false, Message: err.Error()}
	}
	if !hasUpdate {
		return &types.BaseResult{Success: false, Message: "当前已是最新版本"}
	}
	s.updater.DoUpdate(info)
	return &types.BaseResult{Success: true, Message: "更新已开始"}
}

// PanelUpdateProgress 获取面板更新进度。
func (s *ClawService) PanelUpdateProgress() *types.ClawPanelUpdateProgressResult {
	p := s.updater.GetProgress()
	return &types.ClawPanelUpdateProgressResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取更新进度成功"},
		Status:     p.Status,
		Progress:   p.Progress,
		Message:    p.Message,
		Log:        p.Log,
		Error:      p.Error,
	}
}

// GetUpdatePopup 获取更新完成弹窗信息。
func (s *ClawService) GetUpdatePopup() *types.ClawUpdatePopupResult {
	popup := s.updater.GetUpdatePopup()
	if popup == nil {
		return &types.ClawUpdatePopupResult{
			BaseResult: types.BaseResult{Success: true, Message: "获取更新弹窗成功"},
			Show:       false,
		}
	}
	return &types.ClawUpdatePopupResult{
		BaseResult:  types.BaseResult{Success: true, Message: "获取更新弹窗成功"},
		Show:        popup.Show,
		Version:     popup.Version,
		ReleaseNote: popup.ReleaseNote,
	}
}

// MarkUpdatePopupShown 标记更新弹窗已展示。
func (s *ClawService) MarkUpdatePopupShown() *types.BaseResult {
	s.updater.MarkPopupShown()
	return &types.BaseResult{Success: true, Message: "更新弹窗状态已记录"}
}

// GetNapCatStatus 获取 NapCat 状态。
func (s *ClawService) GetNapCatStatus() *types.ClawNapCatStatusResult {
	return &types.ClawNapCatStatusResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取 NapCat 状态成功"},
		Status:     s.napcatMonitor.GetStatus(),
	}
}

// GetNapCatReconnectLogs 获取 NapCat 重连日志。
func (s *ClawService) GetNapCatReconnectLogs() *types.ClawNapCatReconnectLogsResult {
	return &types.ClawNapCatReconnectLogsResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取 NapCat 重连日志成功"},
		Logs:       s.napcatMonitor.GetLogs(),
	}
}

// NapCatReconnect 手动触发 NapCat 重连。
func (s *ClawService) NapCatReconnect() *types.BaseResult {
	if err := s.napcatMonitor.Reconnect(); err != nil {
		return &types.BaseResult{Success: false, Message: err.Error()}
	}
	return &types.BaseResult{Success: true, Message: "重连请求已发送"}
}

// NapCatMonitorConfig 更新 NapCat 监控配置。
func (s *ClawService) NapCatMonitorConfig(payload map[string]interface{}) *types.BaseResult {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	if raw, ok := payload["autoReconnect"]; ok {
		if v, castOK := raw.(bool); castOK {
			s.napcatMonitor.SetAutoReconnect(v)
		}
	}
	if raw, ok := payload["maxReconnect"]; ok {
		switch v := raw.(type) {
		case int:
			s.napcatMonitor.SetMaxReconnect(v)
		case int32:
			s.napcatMonitor.SetMaxReconnect(int(v))
		case int64:
			s.napcatMonitor.SetMaxReconnect(int(v))
		case float64:
			s.napcatMonitor.SetMaxReconnect(int(v))
		case string:
			parsed, err := strconv.Atoi(strings.TrimSpace(v))
			if err == nil {
				s.napcatMonitor.SetMaxReconnect(parsed)
			}
		}
	}
	return &types.BaseResult{Success: true, Message: "NapCat 监控配置已更新"}
}

// GetChatChannelInfo 返回 Boxify 与原生 channel inbox 通信所需的连接信息。
func (s *ClawService) GetChatChannelInfo() *types.ClawChatChannelInfoResult {
	return &types.ClawChatChannelInfoResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取聊天通道信息成功"},
		Data: &types.ClawChatChannelInfo{
			ChannelInboxURL: fmt.Sprintf("http://127.0.0.1:%d", s.pluginPort),
			SharedToken:     s.chatToken,
		},
	}
}

// CreateChatConversation 创建聊天会话。
func (s *ClawService) CreateChatConversation(agentID string) *types.ClawChatConversationResult {
	if s.chatCoordinator == nil {
		return &types.ClawChatConversationResult{
			BaseResult: types.BaseResult{Success: false, Message: "聊天服务未初始化"},
		}
	}
	item, err := s.chatCoordinator.CreateConversation(agentID)
	if err != nil {
		return &types.ClawChatConversationResult{
			BaseResult: types.BaseResult{Success: false, Message: err.Error()},
		}
	}
	return &types.ClawChatConversationResult{
		BaseResult: types.BaseResult{Success: true, Message: "创建聊天会话成功"},
		Data:       item,
	}
}

// ListChatConversations 返回聊天会话列表。
func (s *ClawService) ListChatConversations() *types.ClawChatConversationsResult {
	if s.chatCoordinator == nil {
		return &types.ClawChatConversationsResult{
			BaseResult: types.BaseResult{Success: false, Message: "聊天服务未初始化"},
			Items:      []clawchat.Conversation{},
		}
	}
	items, err := s.chatCoordinator.ListConversations()
	if err != nil {
		return &types.ClawChatConversationsResult{
			BaseResult: types.BaseResult{Success: false, Message: err.Error()},
			Items:      []clawchat.Conversation{},
		}
	}
	return &types.ClawChatConversationsResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取聊天会话成功"},
		Items:      items,
	}
}

// GetChatMessages 返回指定会话消息列表。
func (s *ClawService) GetChatMessages(conversationID string) *types.ClawChatMessagesResult {
	if s.chatCoordinator == nil {
		return &types.ClawChatMessagesResult{
			BaseResult: types.BaseResult{Success: false, Message: "聊天服务未初始化"},
			Items:      []clawchat.Message{},
		}
	}
	items, err := s.chatCoordinator.ListMessages(conversationID)
	if err != nil {
		return &types.ClawChatMessagesResult{
			BaseResult: types.BaseResult{Success: false, Message: err.Error()},
			Items:      []clawchat.Message{},
		}
	}
	return &types.ClawChatMessagesResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取聊天消息成功"},
		Items:      items,
	}
}

// SendChatMessage 向指定会话发送消息。
func (s *ClawService) SendChatMessage(conversationID, text string) *types.ClawChatSendResult {
	if s.chatCoordinator == nil {
		return &types.ClawChatSendResult{
			BaseResult: types.BaseResult{Success: false, Message: "聊天服务未初始化"},
		}
	}
	runID, err := s.chatCoordinator.SendMessage(s.Context(), conversationID, text)
	if err != nil {
		return &types.ClawChatSendResult{
			BaseResult: types.BaseResult{Success: false, Message: err.Error()},
		}
	}
	return &types.ClawChatSendResult{
		BaseResult: types.BaseResult{Success: true, Message: "消息已发送"},
		RunID:      runID,
	}
}

// initRuntimeContext 初始化 Claw 相关运行目录、端口与环境变量上下文。
func (s *ClawService) initRuntimeContext() {
	s.openClawDir = strings.TrimSpace(os.Getenv("OPENCLAW_DIR"))
	if s.openClawDir == "" {
		s.openClawDir = defaultOpenClawDir()
	}
	s.openClawApp = strings.TrimSpace(os.Getenv("OPENCLAW_APP"))
	s.dataDir = strings.TrimSpace(os.Getenv("BOXIFY_DATA_DIR"))
	if s.dataDir == "" {
		home, _ := os.UserHomeDir()
		s.dataDir = filepath.Join(home, ".boxify")
	}
	s.managerPort = defaultClawManagerPort
	if raw := strings.TrimSpace(os.Getenv("CLAW_MANAGER_PORT")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			s.managerPort = v
		}
	}
	s.pluginPort = 32124
	if raw := strings.TrimSpace(os.Getenv("BOXIFY_PLUGIN_INBOX_PORT")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			s.pluginPort = v
		}
	}
	_ = os.MkdirAll(s.dataDir, 0o755)
	_ = os.MkdirAll(s.openClawDir, 0o755)
	if token := strings.TrimSpace(os.Getenv("BOXIFY_CHAT_SHARED_TOKEN")); token != "" {
		s.chatToken = token
	} else if s.chatToken == "" {
		s.chatToken, s.chatTokenGenerated = s.loadOrCreateChatSharedToken()
	}
}

// rebuildManagers 按当前上下文重建并替换所有 Claw 依赖管理器实例。
func (s *ClawService) rebuildManagers() {
	s.manager = clawprocess.NewManager(clawprocess.ManagerConfig{
		OpenClawDir: s.openClawDir,
		OpenClawApp: s.openClawApp,
		ManagerPort: s.managerPort,
	}, s.Logger())
	s.pluginCfg = &clawplugin.Config{OpenClawDir: s.openClawDir, DataDir: s.dataDir}
	s.pluginManager = clawplugin.NewManager(s.pluginCfg, s.Logger())
	if s.chatTokenGenerated {
		s.syncGeneratedChatSharedTokenToOpenClawConfig()
		s.chatTokenGenerated = false
	}
	s.skillManager = clawskill.NewManager(s.pluginCfg, s.openClawDir, s.openClawApp, s.Logger())
	s.taskManager = clawtaskman.NewManager(nil, s.Logger())
	s.updater = clawupdate.NewUpdater(resolveCurrentVersion(), s.dataDir, s.Logger())
	s.chatCoordinator = clawchat.NewChannelCoordinator(
		conversationstore.NewJSONConversationStore("", s.Logger()),
		clawchat.NewHTTPChannelClient(fmt.Sprintf("http://127.0.0.1:%d", s.pluginPort), s.chatToken),
		clawchat.NewWailsEventPublisher(s.App(), s.Logger()),
		s.manager,
		s.Logger(),
	)
	if s.napcatMonitor != nil {
		s.napcatMonitor.Stop()
	}
	s.napcatMonitor = clawmonitor.NewNapCatMonitor(&clawmonitor.MonitorConfig{
		DataDir:     s.dataDir,
		OpenClawDir: s.openClawDir,
	}, nil, s.Logger())
}

func generateChatSharedToken() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("boxify-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

// boxifyConfigPath 返回 Boxify 本地配置文件路径。
func (s *ClawService) boxifyConfigPath() string {
	return filepath.Join(s.dataDir, boxifyConfigFileName)
}

// loadBoxifyLocalConfig 读取 Boxify 本地配置。
func (s *ClawService) loadBoxifyLocalConfig() boxifyLocalConfig {
	var cfg boxifyLocalConfig
	data, err := os.ReadFile(s.boxifyConfigPath())
	if err != nil {
		return cfg
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		s.Logger().Warn("读取 boxify.json 失败，忽略并回退默认值", "path", s.boxifyConfigPath(), "error", err)
		return boxifyLocalConfig{}
	}
	return cfg
}

// saveBoxifyLocalConfig 持久化 Boxify 本地配置。
func (s *ClawService) saveBoxifyLocalConfig(cfg boxifyLocalConfig) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		s.Logger().Warn("序列化 boxify.json 失败", "path", s.boxifyConfigPath(), "error", err)
		return
	}
	if err := os.WriteFile(s.boxifyConfigPath(), data, 0o600); err != nil {
		s.Logger().Warn("写入 boxify.json 失败", "path", s.boxifyConfigPath(), "error", err)
	}
}

// loadOrCreateChatSharedToken 读取已有令牌，不存在时生成并写入 boxify.json。
func (s *ClawService) loadOrCreateChatSharedToken() (string, bool) {
	cfg := s.loadBoxifyLocalConfig()
	if token := strings.TrimSpace(cfg.ChatSharedToken); token != "" {
		return token, false
	}

	token := generateChatSharedToken()
	cfg.ChatSharedToken = token
	s.saveBoxifyLocalConfig(cfg)
	return token, true
}

// syncGeneratedChatSharedTokenToOpenClawConfig 在首次生成 token 后同步写入 openclaw.json。
func (s *ClawService) syncGeneratedChatSharedTokenToOpenClawConfig() {
	if s.pluginCfg == nil || strings.TrimSpace(s.chatToken) == "" {
		return
	}

	ocConfig, err := s.pluginCfg.ReadOpenClawJSON()
	if err != nil {
		s.Logger().Warn("读取 openclaw.json 失败，无法同步聊天共享令牌", "error", err)
		return
	}
	if ocConfig == nil {
		ocConfig = map[string]interface{}{}
	}

	channels, _ := ocConfig["channels"].(map[string]interface{})
	if channels == nil {
		channels = map[string]interface{}{}
	}
	boxifyChannel, _ := channels["boxify"].(map[string]interface{})
	if boxifyChannel == nil {
		boxifyChannel = map[string]interface{}{}
	}
	boxifyChannel["sharedToken"] = s.chatToken
	channels["boxify"] = boxifyChannel
	ocConfig["channels"] = channels

	if err := s.pluginCfg.WriteOpenClawJSON(ocConfig); err != nil {
		s.Logger().Warn("写入 openclaw.json 失败，无法同步聊天共享令牌", "error", err)
		return
	}
	s.Logger().Info("首次生成聊天共享令牌，已同步写入 openclaw.json")
}

// defaultOpenClawDir 返回当前用户下的默认 OpenClaw 配置目录。
func defaultOpenClawDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		return ".openclaw"
	}
	return filepath.Join(home, ".openclaw")
}

// resolveCurrentVersion 从环境变量解析当前应用版本号，缺省返回 unknown。
func resolveCurrentVersion() string {
	if v := strings.TrimSpace(os.Getenv("BOXIFY_VERSION")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("APP_VERSION")); v != "" {
		return v
	}
	return "unknown"
}
