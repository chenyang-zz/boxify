package service

import (
	"context"
	"os"
	"strings"

	clawprocess "github.com/chenyang-zz/boxify/internal/claw/process"
	"github.com/chenyang-zz/boxify/internal/types"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const defaultClawManagerPort = 19527

// ClawService 提供 OpenClaw 进程管理能力。
type ClawService struct {
	BaseService
	manager *clawprocess.Manager
}

// NewClawService 创建 Claw 服务。
func NewClawService(deps *ServiceDeps) *ClawService {
	return &ClawService{BaseService: NewBaseService(deps)}
}

// ServiceStartup 服务启动。
func (s *ClawService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.SetContext(ctx)
	s.manager = clawprocess.NewManager(s.defaultManagerConfig(), s.Logger())
	s.Logger().Info("服务启动", "service", "ClawService")
	return nil
}

// ServiceShutdown 服务关闭。
func (s *ClawService) ServiceShutdown() error {
	if s.manager != nil {
		s.manager.StopAll()
	}
	s.Logger().Info("服务关闭", "service", "ClawService")
	return nil
}

// Configure 更新 Claw 管理配置并重建管理器。
func (s *ClawService) Configure(cfg types.ClawManagerConfig) *types.BaseResult {
	if s.manager != nil {
		s.manager.StopAll()
	}
	next := clawprocess.ManagerConfig{
		OpenClawDir: strings.TrimSpace(cfg.OpenClawDir),
		OpenClawApp: strings.TrimSpace(cfg.OpenClawApp),
		ManagerPort: cfg.ManagerPort,
	}
	if next.ManagerPort <= 0 {
		next.ManagerPort = defaultClawManagerPort
	}
	s.manager = clawprocess.NewManager(next, s.Logger())
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

// GetLogs 获取最近日志。
func (s *ClawService) GetLogs(n int) *types.ClawLogsResult {
	lines := s.manager.GetLogs(n)
	return &types.ClawLogsResult{
		BaseResult: types.BaseResult{Success: true, Message: "获取日志成功"},
		Data:       &types.ClawLogsData{Lines: lines},
	}
}

func (s *ClawService) defaultManagerConfig() clawprocess.ManagerConfig {
	cfg := clawprocess.ManagerConfig{
		OpenClawDir: strings.TrimSpace(os.Getenv("OPENCLAW_DIR")),
		OpenClawApp: strings.TrimSpace(os.Getenv("OPENCLAW_APP")),
		ManagerPort: defaultClawManagerPort,
	}
	return cfg
}
