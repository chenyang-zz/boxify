package types

import (
	clawmonitor "github.com/chenyang-zz/boxify/internal/claw/monitor"
	clawplugin "github.com/chenyang-zz/boxify/internal/claw/plugin"
	clawprocess "github.com/chenyang-zz/boxify/internal/claw/process"
	clawtaskman "github.com/chenyang-zz/boxify/internal/claw/taskman"
	"time"
)

// ClawManagerConfig Claw 管理配置。
type ClawManagerConfig struct {
	OpenClawDir string `json:"openClawDir"` // OpenClaw 状态目录
	OpenClawApp string `json:"openClawApp"` // OpenClaw 应用目录
	ManagerPort int    `json:"managerPort"` // Boxify 管理端口
}

// ClawStatus Claw 进程状态。
type ClawStatus struct {
	Running           bool      `json:"running"`                     // 是否运行中
	PID               int       `json:"pid"`                         // 进程 ID（外部托管时可能为 0）
	StartedAt         time.Time `json:"startedAt,omitempty"`         // 启动时间
	Uptime            int64     `json:"uptime"`                      // 运行时长（秒）
	ExitCode          int       `json:"exitCode,omitempty"`          // 最近退出码
	Daemonized        bool      `json:"daemonized,omitempty"`        // 是否 daemon fork 模式
	ManagedExternally bool      `json:"managedExternally,omitempty"` // 是否外部托管
}

// ClawStatusResult 获取状态结果。
type ClawStatusResult struct {
	BaseResult
	Data *ClawStatus `json:"data,omitempty"` // 状态数据
}

// ClawOpenClawCheckResult OpenClaw 安装与配置检查结果。
type ClawOpenClawCheckResult struct {
	BaseResult
	Installed  bool   `json:"installed"`            // 是否检测到 openclaw 可执行文件
	Configured bool   `json:"configured"`           // 是否检测到可读 openclaw.json
	BinaryPath string `json:"binaryPath,omitempty"` // openclaw 可执行文件路径
	ConfigPath string `json:"configPath,omitempty"` // openclaw.json 路径
}

// ClawOverviewChannel 概览页通道卡片数据。
type ClawOverviewChannel struct {
	ID        string `json:"id"`                  // 通道唯一标识
	Name      string `json:"name"`                // 通道展示名称
	Type      string `json:"type"`                // 通道类型：built-in/plugin
	Status    string `json:"status"`              // 通道状态：enabled/disabled
	ManagedBy string `json:"managedBy,omitempty"` // 管理描述
}

// ClawOverviewData 概览页数据。
type ClawOverviewData struct {
	SystemStatus   string                `json:"systemStatus"`   // 系统状态：normal/warning/error
	ActiveChannels int                   `json:"activeChannels"` // 活跃通道数量
	AIModel        string                `json:"aiModel"`        // 当前默认模型
	Uptime         string                `json:"uptime"`         // 运行时长文案
	MemoryUsage    string                `json:"memoryUsage"`    // 内存占用文案
	TodayMessages  int                   `json:"todayMessages"`  // 今日消息量（当前以今日任务数近似）
	Channels       []ClawOverviewChannel `json:"channels"`       // 通道卡片列表
}

// ClawOverviewResult 概览数据结果。
type ClawOverviewResult struct {
	BaseResult
	Data *ClawOverviewData `json:"data,omitempty"` // 概览数据
}

// ClawLogsData Claw 日志数据。
type ClawLogsData struct {
	Lines []string `json:"lines"` // 日志行列表
}

// ClawLogsResult 获取日志结果。
type ClawLogsResult struct {
	BaseResult
	Data *ClawLogsData `json:"data,omitempty"` // 日志数据
}

// ClawProcessStatusResult 进程状态结果。
type ClawProcessStatusResult struct {
	BaseResult
	Status clawprocess.Status `json:"status"` // OpenClaw 进程状态快照
}

// ClawOpenConfigResult OpenClaw 原始配置结果。
type ClawOpenConfigResult struct {
	BaseResult
	Config map[string]interface{} `json:"config"` // openclaw.json 原始配置
}

// ClawModelsResult 模型配置结果。
type ClawModelsResult struct {
	BaseResult
	Providers map[string]interface{} `json:"providers"` // 模型 providers 配置
	Defaults  map[string]interface{} `json:"defaults"`  // agents.defaults 配置
}

// ClawChannelsResult 通道与插件配置结果。
type ClawChannelsResult struct {
	BaseResult
	Channels map[string]interface{} `json:"channels"` // channels 配置
	Plugins  map[string]interface{} `json:"plugins"`  // plugins 配置
}

// ClawSkill 技能列表项。
type ClawSkill struct {
	ID          string                 `json:"id"`                 // 技能唯一标识
	Name        string                 `json:"name"`               // 技能展示名称
	Description string                 `json:"description"`        // 技能说明
	Version     string                 `json:"version,omitempty"`  // 技能版本
	Enabled     bool                   `json:"enabled"`            // 是否启用
	Source      string                 `json:"source"`             // 技能来源
	Path        string                 `json:"path,omitempty"`     // 技能目录路径
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // 技能元数据
	Requires    map[string]interface{} `json:"requires,omitempty"` // 技能依赖描述
}

// ClawSkillPlugin 技能中心中的插件列表项。
type ClawSkillPlugin struct {
	ID          string `json:"id"`                    // 插件唯一标识
	Name        string `json:"name"`                  // 插件展示名称
	Description string `json:"description"`           // 插件说明
	Version     string `json:"version,omitempty"`     // 插件版本
	Enabled     bool   `json:"enabled"`               // 是否启用
	Source      string `json:"source"`                // 插件来源
	InstalledAt string `json:"installedAt,omitempty"` // 安装时间
	Path        string `json:"path,omitempty"`        // 安装目录
}

// ClawSkillsResult 技能中心列表结果。
type ClawSkillsResult struct {
	BaseResult
	Skills []ClawSkill `json:"skills"` // 技能列表
}

// ClawSkillPluginsResult 技能中心插件列表结果。
type ClawSkillPluginsResult struct {
	BaseResult
	Plugins []ClawSkillPlugin `json:"plugins"` // 插件列表
}

// ClawPluginListResult 插件列表结果。
type ClawPluginListResult struct {
	BaseResult
	Installed []*clawplugin.InstalledPlugin `json:"installed"` // 已安装插件列表
	Registry  []clawplugin.RegistryPlugin   `json:"registry"`  // 仓库插件列表
}

// ClawInstalledPluginsResult 已安装插件列表结果。
type ClawInstalledPluginsResult struct {
	BaseResult
	Plugins []*clawplugin.InstalledPlugin `json:"plugins"` // 已安装插件列表
}

// ClawPluginDetailResult 插件详情结果。
type ClawPluginDetailResult struct {
	BaseResult
	Plugin *clawplugin.InstalledPlugin `json:"plugin,omitempty"` // 插件详情
}

// ClawRegistryResult 插件仓库结果。
type ClawRegistryResult struct {
	BaseResult
	Registry *clawplugin.Registry `json:"registry,omitempty"` // 插件仓库内容
}

// ClawPluginConfigResult 插件配置结果。
type ClawPluginConfigResult struct {
	BaseResult
	Config map[string]interface{} `json:"config,omitempty"` // 插件配置
	Schema []byte                 `json:"schema,omitempty"` // 插件配置 schema 原始 JSON
}

// ClawPluginLogsResult 插件日志结果。
type ClawPluginLogsResult struct {
	BaseResult
	Logs []string `json:"logs,omitempty"` // 插件日志行
}

// ClawTasksResult 任务列表结果。
type ClawTasksResult struct {
	BaseResult
	Tasks []*clawtaskman.Task `json:"tasks"` // 最近任务列表
}

// ClawTaskDetailResult 任务详情结果。
type ClawTaskDetailResult struct {
	BaseResult
	Task *clawtaskman.Task `json:"task,omitempty"` // 任务详情
}

// ClawCheckUpdateResult 更新检查结果。
type ClawCheckUpdateResult struct {
	BaseResult
	HasUpdate     bool   `json:"hasUpdate"`     // 是否存在可用更新
	LatestVersion string `json:"latestVersion"` // 最新版本号
	ReleaseTime   string `json:"releaseTime"`   // 发布时间
	ReleaseNote   string `json:"releaseNote"`   // 更新说明
}

// ClawPanelUpdateProgressResult 更新进度结果。
type ClawPanelUpdateProgressResult struct {
	BaseResult
	Status   string   `json:"status"`          // 更新状态
	Progress int      `json:"progress"`        // 更新进度（0-100）
	Message  string   `json:"message"`         // 进度描述
	Log      []string `json:"log"`             // 日志快照
	Error    string   `json:"error,omitempty"` // 错误信息
}

// ClawUpdatePopupResult 更新弹窗结果。
type ClawUpdatePopupResult struct {
	BaseResult
	Show        bool   `json:"show"`                  // 是否展示更新弹窗
	Version     string `json:"version,omitempty"`     // 更新版本号
	ReleaseNote string `json:"releaseNote,omitempty"` // 更新说明
}

// ClawNapCatStatusResult NapCat 状态结果。
type ClawNapCatStatusResult struct {
	BaseResult
	Status clawmonitor.NapCatStatus `json:"status"` // NapCat 状态快照
}

// ClawNapCatReconnectLogsResult NapCat 重连日志结果。
type ClawNapCatReconnectLogsResult struct {
	BaseResult
	Logs []clawmonitor.ReconnectLog `json:"logs"` // NapCat 重连日志
}
