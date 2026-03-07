package process

import (
	"io"
	"log/slog"
	"os/exec"
	"sync"
	"time"
)

// Status 描述 OpenClaw 进程状态。
type Status struct {
	Running           bool      `json:"running"`                     // 是否运行中
	PID               int       `json:"pid"`                         // 进程 ID（外部托管时可能为 0）
	StartedAt         time.Time `json:"startedAt,omitempty"`         // 启动时间
	Uptime            int64     `json:"uptime"`                      // 运行时长（秒）
	ExitCode          int       `json:"exitCode,omitempty"`          // 最近一次退出码
	Daemonized        bool      `json:"daemonized,omitempty"`        // 是否为 daemon fork 模式
	ManagedExternally bool      `json:"managedExternally,omitempty"` // 是否由外部进程管理
}

// ManagerConfig 管理器运行配置。
type ManagerConfig struct {
	OpenClawDir string // OpenClaw 状态目录（openclaw.json 所在目录）
	OpenClawApp string // OpenClaw 应用目录（用于插件补丁扫描）
	ManagerPort int    // Boxify 管理端口（用于注入插件日志回传地址）
}

// Manager 负责 OpenClaw 网关进程生命周期管理。
type Manager struct {
	cfg                ManagerConfig                // 运行配置
	logger             *slog.Logger                 // 日志器
	cmd                *exec.Cmd                    // 当前托管的 gateway 主进程
	daemonized         bool                         // 是否识别为 daemon fork 模式
	bindHostCheck      func(host string) bool       // 可注入的 host 绑定探测（测试/特化）
	gatewayProbe       func(host, port string) bool // 可注入的网关识别探测（测试/特化）
	lastGatewayProbeAt time.Time                    // 最近一次探测时间
	lastGatewayProbeOK bool                         // 最近一次探测结果
	status             Status                       // 当前状态快照
	mu                 sync.RWMutex                 // 进程状态并发保护
	logLines           []string                     // 内存日志缓冲
	logMu              sync.RWMutex                 // 日志缓冲并发保护
	maxLog             int                          // 日志缓冲上限
	stopCh             chan struct{}                // 日志流终止信号
	logReader          io.ReadCloser                // stdout/stderr 合并读取器
}

// NewManager 创建进程管理器。
func NewManager(cfg ManagerConfig, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	return &Manager{
		cfg:    cfg,
		logger: logger,
		maxLog: 5000,
		stopCh: make(chan struct{}),
	}
}
