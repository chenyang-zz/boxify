package types

import "time"

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

// ClawLogsData Claw 日志数据。
type ClawLogsData struct {
	Lines []string `json:"lines"` // 日志行列表
}

// ClawLogsResult 获取日志结果。
type ClawLogsResult struct {
	BaseResult
	Data *ClawLogsData `json:"data,omitempty"` // 日志数据
}
