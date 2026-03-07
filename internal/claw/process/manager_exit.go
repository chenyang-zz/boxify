package process

import (
	"bufio"
	"net"
	"os/exec"
	"time"
)

// waitForExit 监听主进程退出，并处理 daemon fork 与异常重启路径。
func (m *Manager) waitForExit() {
	if m.logReader != nil {
		scanner := bufio.NewScanner(m.logReader)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)
		for scanner.Scan() {
			m.addLogLine(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			m.logger.Warn("读取 OpenClaw 日志流失败", "error", err)
		}
	}

	if m.cmd == nil {
		return
	}

	err := m.cmd.Wait()
	m.mu.Lock()
	wasRunning := m.status.Running
	daemonized := m.daemonized
	startedAt := m.status.StartedAt
	m.status.Running = false
	m.status.Daemonized = false
	m.daemonized = false
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			m.status.ExitCode = exitCode
		}
	}
	m.mu.Unlock()

	m.logger.Info("OpenClaw 进程已退出", "code", exitCode)

	if wasRunning && !daemonized && !startedAt.IsZero() && time.Since(startedAt) < 15*time.Second {
		if m.waitForGatewayReady(8 * time.Second) {
			m.logger.Info("检测到 daemon fork 模式，父进程退出但网关仍存活")
			m.mu.Lock()
			m.status.Running = true
			m.status.ExitCode = 0
			m.status.PID = 0
			m.cmd = nil
			m.daemonized = true
			m.status.Daemonized = true
			m.mu.Unlock()
			go m.monitorDaemon()
			return
		}
	}

	if wasRunning && exitCode != 0 {
		m.logger.Warn("检测到 OpenClaw 异常退出，准备自动重启", "delay", "2s")
		time.Sleep(2 * time.Second)
		if restartErr := m.Start(); restartErr != nil {
			m.logger.Warn("自动重启失败", "error", restartErr)
		} else {
			m.logger.Info("OpenClaw 已自动重启")
		}
	}
}

// waitForGatewayReady 在超时窗口内轮询网关是否可探测。
func (m *Manager) waitForGatewayReady(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if m.gatewayListening(true) {
			return true
		}
		if time.Now().After(deadline) {
			m.logger.Warn("等待网关就绪超时", "timeout", timeout.String())
			return false
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// isPortListening 在候选 bind 目标上检查给定端口是否存在监听者。
func (m *Manager) isPortListening(port string) bool {
	hosts := m.getGatewayPortCheckTargets()
	for _, host := range hosts {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 300*time.Millisecond)
		if err != nil {
			continue
		}
		_ = conn.Close()
		return true
	}
	return false
}

// monitorDaemon 监控 daemon fork 模式网关可用性，连续失败后触发重启。
func (m *Manager) monitorDaemon() {
	failCount := 0
	for {
		time.Sleep(5 * time.Second)
		m.mu.RLock()
		running := m.status.Running
		daemonized := m.daemonized
		m.mu.RUnlock()
		if !running || !daemonized {
			return
		}
		if m.gatewayListening(true) {
			failCount = 0
			continue
		}
		failCount++
		if failCount < 3 {
			continue
		}
		m.logger.Warn("OpenClaw 守护进程不可达，准备自动重启")
		m.mu.Lock()
		m.status.Running = false
		m.status.Daemonized = false
		m.daemonized = false
		m.lastGatewayProbeAt = time.Time{}
		m.lastGatewayProbeOK = false
		m.mu.Unlock()
		time.Sleep(2 * time.Second)
		if err := m.Start(); err != nil {
			m.logger.Warn("自动重启失败", "error", err)
		} else {
			m.logger.Info("OpenClaw 已自动重启")
		}
		return
	}
}
