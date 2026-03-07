package process

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Start 启动 OpenClaw gateway 进程。
func (m *Manager) Start() error {
	if status := m.GetStatus(); status.Running {
		if status.ManagedExternally {
			m.logger.Warn("启动被拒绝：网关由外部进程管理", "managedExternally", true)
			return fmt.Errorf("OpenClaw 网关已由外部进程管理并在运行中")
		}
		m.logger.Warn("启动被拒绝：进程已在运行", "pid", status.PID)
		return fmt.Errorf("OpenClaw 已在运行中 (PID: %d)", status.PID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status.Running {
		m.logger.Warn("启动被拒绝：进程状态已标记为运行中", "pid", m.status.PID)
		return fmt.Errorf("OpenClaw 已在运行中 (PID: %d)", m.status.PID)
	}
	if gatewayPort := m.getGatewayPort(); gatewayPort != "" && m.isPortListening(gatewayPort) {
		if m.detectGatewayListening() {
			m.logger.Warn("启动被拒绝：检测到外部网关监听", "port", gatewayPort)
			return fmt.Errorf("OpenClaw 网关已由外部进程管理并在运行中")
		}
		m.logger.Warn("启动被拒绝：网关端口已被其他服务占用", "port", gatewayPort)
		return fmt.Errorf("OpenClaw 网关端口 %s 已被其他本地服务占用", gatewayPort)
	}

	m.ensureOpenClawConfig()

	openclawBin := m.findOpenClawBin()
	if openclawBin == "" {
		m.logger.Error("启动失败：未找到 openclaw 可执行文件")
		return fmt.Errorf("未找到 openclaw 可执行文件，请确保已安装 OpenClaw")
	}
	m.logger.Info("准备启动 OpenClaw", "bin", openclawBin, "dir", m.cfg.OpenClawDir)

	m.cmd = exec.Command(openclawBin, "gateway")
	m.cmd.Dir = m.cfg.OpenClawDir
	m.cmd.Env = append(BuildExecEnv(),
		fmt.Sprintf("OPENCLAW_DIR=%s", m.cfg.OpenClawDir),
		fmt.Sprintf("OPENCLAW_STATE_DIR=%s", m.cfg.OpenClawDir),
		fmt.Sprintf("OPENCLAW_CONFIG_PATH=%s/openclaw.json", m.cfg.OpenClawDir),
	)

	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建 stdout 管道失败: %w", err)
	}
	stderr, err := m.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建 stderr 管道失败: %w", err)
	}

	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("启动 OpenClaw 失败: %w", err)
	}

	m.status = Status{Running: true, PID: m.cmd.Process.Pid, StartedAt: time.Now()}
	m.daemonized = false
	m.lastGatewayProbeAt = time.Time{}
	m.lastGatewayProbeOK = false
	m.logReader = io.NopCloser(io.MultiReader(stdout, stderr))
	go m.waitForExit()

	m.logger.Info("OpenClaw 已启动", "pid", m.status.PID)
	return nil
}

// Stop 停止 OpenClaw 进程。
func (m *Manager) Stop() error {
	if status := m.GetStatus(); status.ManagedExternally {
		m.logger.Warn("停止被拒绝：网关由外部进程管理")
		return fmt.Errorf("OpenClaw 网关当前由外部进程管理，无法在 Boxify 内停止")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.daemonized {
		m.logger.Warn("停止被拒绝：当前为 daemon fork 模式")
		return fmt.Errorf("OpenClaw 当前以 daemon fork 模式运行，Boxify 暂不支持直接停止")
	}
	if !m.status.Running {
		m.logger.Warn("停止被拒绝：进程未运行")
		return fmt.Errorf("OpenClaw 未在运行")
	}

	m.logger.Info("正在停止 OpenClaw", "pid", m.status.PID)
	gatewayPort := m.getGatewayPort()

	if bin := m.findOpenClawBin(); bin != "" {
		cmd := exec.Command(bin, "gateway", "stop")
		cmd.Dir = m.cfg.OpenClawDir
		cmd.Env = append(BuildExecEnv(),
			fmt.Sprintf("OPENCLAW_DIR=%s", m.cfg.OpenClawDir),
			fmt.Sprintf("OPENCLAW_STATE_DIR=%s", m.cfg.OpenClawDir),
			fmt.Sprintf("OPENCLAW_CONFIG_PATH=%s/openclaw.json", m.cfg.OpenClawDir),
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			m.logger.Warn("gateway stop 命令失败", "error", err, "output", strings.TrimSpace(string(out)))
		}
	} else {
		m.logger.Warn("未找到 openclaw 可执行文件，跳过 gateway stop 命令并继续尝试直接终止")
	}

	for i := 0; i < 10; i++ {
		if !m.isPortListening(gatewayPort) {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}

	if runtime.GOOS == "windows" && m.isPortListening(gatewayPort) {
		if killed := m.killWindowsPortListeners(gatewayPort); killed > 0 {
			m.logger.Info("已强制终止端口占用进程", "count", killed, "port", gatewayPort)
		}
		for i := 0; i < 10; i++ {
			if !m.isPortListening(gatewayPort) {
				break
			}
			time.Sleep(300 * time.Millisecond)
		}
	}

	if m.cmd != nil && m.cmd.Process != nil {
		if runtime.GOOS == "windows" {
			_ = m.cmd.Process.Kill()
		} else {
			_ = m.cmd.Process.Signal(os.Interrupt)
			done := make(chan struct{})
			go func() {
				_ = m.cmd.Wait()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				_ = m.cmd.Process.Kill()
			}
		}
	}

	if m.isPortListening(gatewayPort) {
		m.logger.Error("停止失败：网关端口仍被占用", "port", gatewayPort)
		return fmt.Errorf("OpenClaw 网关端口 %s 仍被占用，停止失败", gatewayPort)
	}

	m.status.Running = false
	m.status.PID = 0
	m.status.Daemonized = false
	m.lastGatewayProbeAt = time.Time{}
	m.lastGatewayProbeOK = false
	m.logger.Info("OpenClaw 已停止")
	return nil
}

// Restart 重启 OpenClaw 进程。
func (m *Manager) Restart() error {
	status := m.GetStatus()
	if status.ManagedExternally {
		m.logger.Warn("重启被拒绝：网关由外部进程管理")
		return fmt.Errorf("OpenClaw 网关当前由外部进程管理，请在外部环境中重启")
	}

	m.mu.RLock()
	daemonized := m.daemonized
	m.mu.RUnlock()
	if daemonized {
		m.logger.Warn("重启被拒绝：当前为 daemon fork 模式")
		return fmt.Errorf("OpenClaw 当前以 daemon fork 模式运行，请在外部环境中重启")
	}

	if status.Running {
		if err := m.Stop(); err != nil {
			m.logger.Warn("重启前停止失败", "error", err)
		}
		time.Sleep(time.Second)
	}
	return m.Start()
}

// GatewayListening 检查是否存在可探测的 OpenClaw gateway。
func (m *Manager) GatewayListening() bool {
	return m.gatewayListening(false)
}

// StopAll 停止所有托管进程。
func (m *Manager) StopAll() {
	status := m.GetStatus()
	if status.ManagedExternally {
		return
	}
	if status.Running {
		_ = m.Stop()
	}
}

// GetStatus 返回当前进程状态。
func (m *Manager) GetStatus() Status {
	m.mu.RLock()
	s := m.status
	m.mu.RUnlock()

	if s.Running {
		s.Uptime = int64(time.Since(s.StartedAt).Seconds())
		return s
	}
	if m.GatewayListening() {
		s.Running = true
		s.PID = 0
		s.StartedAt = time.Time{}
		s.Uptime = 0
		s.ExitCode = 0
		s.Daemonized = false
		s.ManagedExternally = true
	}
	return s
}
