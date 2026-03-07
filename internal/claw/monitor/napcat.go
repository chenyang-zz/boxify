package monitor

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// MonitorConfig 定义 NapCat 监控所需的最小配置。
type MonitorConfig struct {
	DataDir     string // NapCat 数据目录
	OpenClawDir string // OpenClaw 配置目录
}

// StatusBroadcaster 负责向外广播 NapCat 状态变化。
type StatusBroadcaster interface {
	Broadcast(payload []byte)
}

// NapCatStatus 描述当前 NapCat 连接与登录状态快照。
type NapCatStatus struct {
	ContainerRunning bool      `json:"containerRunning"`     // NapCat 进程/容器是否运行
	WSConnected      bool      `json:"wsConnected"`          // OneBot WS 端口是否可达
	HTTPAvailable    bool      `json:"httpAvailable"`        // OneBot HTTP 端口是否可达
	QQLoggedIn       bool      `json:"qqLoggedIn"`           // QQ 是否已登录
	QQNickname       string    `json:"qqNickname,omitempty"` // 当前登录 QQ 昵称
	QQID             string    `json:"qqId,omitempty"`       // 当前登录 QQ 号
	LastCheck        time.Time `json:"lastCheck"`            // 最近一次探测时间
	LastOnline       time.Time `json:"lastOnline,omitempty"` // 最近一次在线时间
	ReconnectCount   int       `json:"reconnectCount"`       // 当前累计重连次数
	MaxReconnect     int       `json:"maxReconnect"`         // 自动重连上限
	AutoReconnect    bool      `json:"autoReconnect"`        // 是否启用自动重连
	Status           string    `json:"status"`               // 总体状态：online/offline/reconnecting/login_expired/stopped
}

// ReconnectLog 记录一次重连尝试的结果。
type ReconnectLog struct {
	Time    time.Time `json:"time"`             // 重连触发时间
	Reason  string    `json:"reason"`           // 重连原因
	Success bool      `json:"success"`          // 是否重连成功
	Detail  string    `json:"detail,omitempty"` // 补充说明
}

// NapCatMonitor 负责周期检测 NapCat 状态并执行自动重连。
type NapCatMonitor struct {
	cfg            *MonitorConfig
	broadcaster    StatusBroadcaster
	logger         *slog.Logger
	mu             sync.RWMutex
	status         NapCatStatus
	logs           []ReconnectLog
	maxLogs        int // 最多保留的重连日志数量
	stopCh         chan struct{}
	running        bool
	paused         bool // QQ 通道关闭时暂停检测
	checkInterval  time.Duration
	reconnecting   bool // 正在执行重连流程时为 true
	offlineCount   int  // 连续离线计数，用于触发自动重连
	loginFailCount int  // 连续登录检查失败次数，用于判定 login_expired
}

// NewNapCatMonitor 创建 NapCatMonitor，并初始化默认重连策略。
func NewNapCatMonitor(cfg *MonitorConfig, broadcaster StatusBroadcaster, logger *slog.Logger) *NapCatMonitor {
	if cfg == nil {
		cfg = &MonitorConfig{}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &NapCatMonitor{
		cfg:           cfg,
		broadcaster:   broadcaster,
		logger:        logger,
		maxLogs:       100,
		checkInterval: 30 * time.Second,
		status: NapCatStatus{
			MaxReconnect:  10,
			AutoReconnect: true,
			Status:        "offline",
		},
		stopCh: make(chan struct{}),
	}
}

// Start 启动监控循环（幂等，重复调用不会重复启动）。
func (m *NapCatMonitor) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	go m.monitorLoop()
	m.logger.Info("NapCat 监控已启动", "interval", m.checkInterval.String())
}

// Stop 停止监控循环（幂等，重复调用不会报错）。
func (m *NapCatMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return
	}
	m.running = false
	close(m.stopCh)
	m.logger.Info("NapCat 监控已停止")
}

// GetStatus 返回当前 NapCat 状态快照。
func (m *NapCatMonitor) GetStatus() NapCatStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// GetLogs 返回重连日志副本，避免外部修改内部切片。
func (m *NapCatMonitor) GetLogs() []ReconnectLog {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]ReconnectLog, len(m.logs))
	copy(result, m.logs)
	return result
}

// Reconnect 手动触发重连流程。
func (m *NapCatMonitor) Reconnect() error {
	m.mu.Lock()
	m.status.Status = "reconnecting"
	m.mu.Unlock()

	m.broadcastStatus()
	m.logger.Info("收到 NapCat 手动重连请求")
	return m.doReconnect("手动触发重连")
}

// SetAutoReconnect 设置是否启用自动重连。
func (m *NapCatMonitor) SetAutoReconnect(enabled bool) {
	m.mu.Lock()
	m.status.AutoReconnect = enabled
	m.mu.Unlock()
	m.logger.Info("NapCat 自动重连开关已更新", "enabled", enabled)
}

// SetMaxReconnect 设置自动重连次数上限。
func (m *NapCatMonitor) SetMaxReconnect(max int) {
	m.mu.Lock()
	m.status.MaxReconnect = max
	m.mu.Unlock()
	m.logger.Info("NapCat 自动重连上限已更新", "maxReconnect", max)
}

// Pause 暂停监控检测（用于 QQ 通道关闭场景）。
func (m *NapCatMonitor) Pause() {
	m.mu.Lock()
	m.paused = true
	m.mu.Unlock()
	m.logger.Info("NapCat 监控已暂停", "reason", "QQ 通道已关闭")
}

// Resume 恢复监控检测并重置离线计数。
func (m *NapCatMonitor) Resume() {
	m.mu.Lock()
	m.paused = false
	m.offlineCount = 0
	m.loginFailCount = 0
	m.status.ReconnectCount = 0
	m.mu.Unlock()
	m.logger.Info("NapCat 监控已恢复")
}

// IsPaused 返回当前是否处于暂停检测状态。
func (m *NapCatMonitor) IsPaused() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.paused
}

// monitorLoop 执行定时检测主循环。
func (m *NapCatMonitor) monitorLoop() {
	// 首次检测前短暂等待，避免启动阶段误判。
	select {
	case <-m.stopCh:
		return
	case <-time.After(5 * time.Second):
	}

	m.checkAndUpdate()

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkAndUpdate()
		}
	}
}

// checkAndUpdate 执行一次状态探测、状态转换与自动重连判断。
func (m *NapCatMonitor) checkAndUpdate() {
	// 暂停状态或重连中时跳过检测。
	m.mu.RLock()
	if m.paused || m.reconnecting {
		m.mu.RUnlock()
		return
	}
	m.mu.RUnlock()

	containerRunning := isNapCatProcessRunning()
	wsConnected := isPortReachable(3001)
	httpAvailable := isPortReachable(3000)
	qqLoggedIn := false
	qqNickname := ""
	qqID := ""

	// 仅当 WebUI 端口可达时检查登录态。
	if isPortReachable(6099) {
		qqLoggedIn, qqNickname, qqID = checkQQLoginStatus(m.cfg)
	}

	m.mu.Lock()
	prevStatus := m.status.Status

	m.status.ContainerRunning = containerRunning
	m.status.WSConnected = wsConnected
	m.status.HTTPAvailable = httpAvailable
	m.status.QQLoggedIn = qqLoggedIn
	m.status.QQNickname = qqNickname
	m.status.QQID = qqID
	m.status.LastCheck = time.Now()

	// 计算综合状态。
	if !containerRunning {
		m.offlineCount++
		m.status.Status = "stopped"
		m.loginFailCount = 0
	} else if qqLoggedIn {
		m.status.Status = "online"
		m.status.LastOnline = time.Now()
		m.status.ReconnectCount = 0
		m.offlineCount = 0
		m.loginFailCount = 0
	} else if wsConnected || httpAvailable {
		// 进程在运行、服务可达，但 QQ 未登录。
		// 连续失败达到阈值后再判定 login_expired，降低瞬时超时误判。
		m.loginFailCount++
		if prevStatus == "online" && m.loginFailCount < 3 {
			// 在确认离线前保持 online（3 次探测约 90 秒）。
			m.status.Status = "online"
		} else if prevStatus == "online" || prevStatus == "login_expired" {
			m.status.Status = "login_expired"
		} else {
			m.status.Status = "offline"
		}
		m.offlineCount = 0
	} else if containerRunning {
		// 进程在运行但 OneBot 服务（3001/3000）不可用。
		// 若 6099 可达说明 WebUI 仍存活，仅等待扫码登录。
		// 此时判定为 login_expired（不自动重启，需用户扫码）。
		webuiUp := isPortReachable(6099)
		m.offlineCount++
		if webuiUp {
			// NapCat 运行正常，仅需重新登录，不是崩溃。
			m.status.Status = "login_expired"
			m.loginFailCount++
			m.offlineCount = 0 // 重置计数，避免触发自动重连。
		} else if m.offlineCount <= 3 {
			// 启动期允许短暂不稳定，先保持旧状态。
			m.status.Status = prevStatus // 保持上一次状态
		} else {
			m.status.Status = "offline"
		}
	} else {
		m.offlineCount++
		m.status.Status = "stopped"
	}

	currentStatus := m.status.Status
	autoReconnect := m.status.AutoReconnect
	reconnectCount := m.status.ReconnectCount
	maxReconnect := m.status.MaxReconnect
	offlineCount := m.offlineCount
	m.mu.Unlock()

	// NapCat 运行但 WS/HTTP 未监听时，补齐网络配置。
	// 处理 QQ 刚登录且配置为空的场景。
	if containerRunning && (!wsConnected || !httpAvailable) && isPortReachable(6099) {
		napcatDir := findNapCatShellDir(m.cfg)
		if napcatDir != "" {
			go ensureNapCatNetworkConfig(napcatDir)
		}
	}

	// 状态变化时广播。
	if currentStatus != prevStatus {
		m.logger.Info("NapCat 状态变更",
			"from", prevStatus,
			"to", currentStatus,
			"containerRunning", containerRunning,
			"wsConnected", wsConnected,
			"httpAvailable", httpAvailable,
			"qqLoggedIn", qqLoggedIn,
			"offlineCount", offlineCount,
		)
		m.broadcastStatus()

		// 仅记录关键状态转换。
		switch currentStatus {
		case "online":
			m.logSystemEvent("napcat.online", fmt.Sprintf("NapCat QQ 已上线 (%s: %s)", qqID, qqNickname))
		case "login_expired":
			if prevStatus == "online" {
				m.logSystemEvent("napcat.login_expired", "NapCat QQ 登录已失效，需要重新扫码")
			}
		case "stopped":
			if prevStatus == "online" || prevStatus == "login_expired" {
				m.logSystemEvent("napcat.stopped", "NapCat 容器已停止")
			}
		case "offline":
			if prevStatus == "online" {
				m.logSystemEvent("napcat.offline", "NapCat QQ 已离线")
			}
		}
	}

	// 自动重连：仅对 offline（服务异常）和 stopped（进程停止）触发。
	// login_expired 表示 WebUI 存活但会话过期，需要用户扫码，不应自动重启。
	if autoReconnect && (currentStatus == "offline" || currentStatus == "stopped") && offlineCount >= 3 {
		if reconnectCount < maxReconnect {
			m.logger.Warn("触发 NapCat 自动重连",
				"status", currentStatus,
				"offlineCount", offlineCount,
				"reconnect", reconnectCount+1,
				"maxReconnect", maxReconnect,
			)
			m.mu.Lock()
			m.status.Status = "reconnecting"
			m.reconnecting = true
			m.mu.Unlock()
			m.broadcastStatus()

			go func() {
				m.doReconnect("检测到连接断开，自动重连")
				m.mu.Lock()
				m.reconnecting = false
				m.offlineCount = 0
				m.mu.Unlock()
			}()
		} else if reconnectCount == maxReconnect {
			m.logger.Warn("NapCat 自动重连达到上限", "maxReconnect", maxReconnect)
			m.logSystemEvent("napcat.reconnect_limit", fmt.Sprintf("NapCat 自动重连已达上限 (%d 次)，停止重连", maxReconnect))
			m.mu.Lock()
			m.status.ReconnectCount = maxReconnect + 1 // 避免重复记录上限日志
			m.mu.Unlock()
		}
	}
}

// doReconnect 执行单次重连，并按结果回写状态与日志。
func (m *NapCatMonitor) doReconnect(reason string) error {
	m.mu.Lock()
	m.status.ReconnectCount++
	count := m.status.ReconnectCount
	m.mu.Unlock()

	m.logSystemEvent("napcat.reconnecting", fmt.Sprintf("NapCat 正在重连 (第 %d 次): %s", count, reason))
	m.logger.Info("NapCat 开始重连", "count", count, "reason", reason)

	rlog := ReconnectLog{
		Time:   time.Now(),
		Reason: reason,
	}

	// 重启后旧会话凭据失效，清理缓存凭据。
	cachedMonitorCred = ""

	// 按平台重启 NapCat（Linux/macOS 用 Docker，Windows 用进程）。
	out, err := restartNapCatPlatform(m.cfg, m.logger)

	if err != nil {
		rlog.Success = false
		rlog.Detail = fmt.Sprintf("docker restart 失败: %s %s", err.Error(), string(out))
		m.addLog(rlog)

		m.mu.Lock()
		m.status.Status = "offline"
		m.mu.Unlock()
		m.broadcastStatus()

		m.logSystemEvent("napcat.reconnect_failed", fmt.Sprintf("NapCat 重连失败: %v", err))
		m.logger.Error("NapCat 重连失败", "error", err, "output", strings.TrimSpace(string(out)))
		return err
	}

	// 等待 NapCat 启动完成。Windows 下 schtasks 存在调度延迟，
	// 且 3001/3000 依赖 QQ 登录，因此以 6099(WebUI) 作为恢复判断。
	time.Sleep(8 * time.Second)

	webuiOK := false
	for i := 0; i < 12; i++ {
		if isPortReachable(6099) {
			webuiOK = true
			break
		}
		time.Sleep(5 * time.Second)
	}

	if webuiOK {
		rlog.Success = true
		rlog.Detail = "NapCat 已重启，WebUI 已可用，等待 QQ 扫码登录"
		m.addLog(rlog)

		m.mu.Lock()
		m.status.ContainerRunning = true
		m.status.Status = "login_expired"
		m.mu.Unlock()
		m.broadcastStatus()

		m.logSystemEvent("napcat.reconnected_no_login", "NapCat 已重启，请扫码登录 QQ")
		m.logger.Info("NapCat 重连完成，WebUI 已恢复，等待扫码登录")
		return nil
	}

	rlog.Success = false
	rlog.Detail = "NapCat 重启后 WebUI (port 6099) 未恢复"
	m.addLog(rlog)

	m.mu.Lock()
	m.status.Status = "offline"
	m.mu.Unlock()
	m.broadcastStatus()

	m.logSystemEvent("napcat.reconnect_failed", "NapCat 重连后 WebUI 未恢复")
	m.logger.Error("NapCat 重连失败，WebUI 未恢复")
	return fmt.Errorf("NapCat WebUI 未恢复")
}

// addLog 追加重连日志并裁剪到最大保留数量。
func (m *NapCatMonitor) addLog(rlog ReconnectLog) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, rlog)
	if len(m.logs) > m.maxLogs {
		m.logs = m.logs[len(m.logs)-m.maxLogs:]
	}
}

// broadcastStatus 广播当前状态给前端订阅方。
func (m *NapCatMonitor) broadcastStatus() {
	m.mu.RLock()
	status := m.status
	m.mu.RUnlock()

	msg, _ := json.Marshal(map[string]interface{}{
		"type": "napcat-status",
		"data": status,
	})
	if m.broadcaster != nil {
		m.broadcaster.Broadcast(msg)
	}
}

// logSystemEvent 记录 NapCat 系统事件。
func (m *NapCatMonitor) logSystemEvent(action, message string) {
	m.logger.Info("NapCat 系统事件", "category", "system", "action", action, "message", message)
}

// --- 辅助函数 ---

// 缓存 NapCat WebUI 凭据，避免每次探测都重新登录。
// 当凭据过期或重启后失效时再重新认证。
var (
	cachedMonitorCred     string
	cachedMonitorCredTime time.Time
)

// isContainerRunning 检查指定 Docker 容器是否处于运行状态。
func isContainerRunning(name string) bool {
	out, err := dockerOutput("inspect", "--format", "{{.State.Running}}", name)
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// isNapCatProcessRunning 检查 NapCat 是否在交互会话（session 1）中有效运行。
// Windows 下除进程存在外还要求 6099 端口可达，避免把僵尸进程误判为运行中。
func isNapCatProcessRunning() bool {
	if runtime.GOOS == "windows" {
		out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq NapCatWinBootMain.exe", "/NH").Output()
		if err != nil || !strings.Contains(string(out), "NapCatWinBootMain") {
			out2, err2 := exec.Command("tasklist", "/FI", "IMAGENAME eq napcat.exe", "/NH").Output()
			if err2 != nil || !strings.Contains(string(out2), "napcat.exe") {
				return false
			}
		}
		// 仅当 6099 端口可达时视为运行中。
		// 失联僵尸进程应判定为 stopped，便于监控触发重启。
		return isPortReachable(6099)
	}
	return isContainerRunning("openclaw-qq")
}

// StopNapCatPlatform 按平台停止 NapCat 进程（Windows）或容器（Linux/macOS）。
func StopNapCatPlatform() {
	slog.Default().Info("请求停止 NapCat", "platform", runtime.GOOS)
	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/IM", "NapCatWinBootMain.exe").Run()
		exec.Command("taskkill", "/F", "/IM", "napcat.exe").Run()
		exec.Command("taskkill", "/F", "/IM", "QQ.exe").Run()
	} else {
		dockerRun("stop", "openclaw-qq")
	}
	slog.Default().Info("NapCat 停止请求已执行")
}

// restartNapCatPlatform 按平台执行 NapCat 重启。
// Windows 下仅在 6099 不可达时执行 kill，再通过交互会话拉起。
func restartNapCatPlatform(cfg *MonitorConfig, logger *slog.Logger) ([]byte, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if runtime.GOOS == "windows" {
		napcatDir := findNapCatShellDir(cfg)
		if napcatDir == "" {
			logger.Error("NapCat 重启失败：未找到 NapCat Shell 目录")
			return []byte("NapCat Shell directory not found"), fmt.Errorf("NapCat Shell not installed")
		}
		logger.Info("NapCat 重启准备：已定位 Shell 目录", "dir", napcatDir)
		exePath := filepath.Join(napcatDir, "NapCatWinBootMain.exe")
		if _, err := os.Stat(exePath); err != nil {
			logger.Error("NapCat 重启失败：可执行文件不存在", "path", exePath, "error", err)
			return []byte("No NapCatWinBootMain.exe found"), fmt.Errorf("NapCat executable not found")
		}
		// 仅在 WebUI 端口不可达时 kill，避免误杀正常会话。
		if !isPortReachable(6099) {
			exec.Command("taskkill", "/F", "/IM", "NapCatWinBootMain.exe").Run()
			exec.Command("taskkill", "/F", "/IM", "napcat.exe").Run()
			exec.Command("taskkill", "/F", "/IM", "QQ.exe").Run()
			time.Sleep(3 * time.Second)
		} else {
			logger.Info("NapCat WebUI 端口可用，跳过 kill", "port", 6099)
			return []byte("NapCat already running"), nil
		}
		// 启动前先写入网络配置，确保新进程读取到 WS/HTTP 配置。
		logger.Info("准备写入 NapCat 网络配置", "dir", napcatDir)
		ensureNapCatNetworkConfig(napcatDir)
		logger.Info("准备在用户会话中启动 NapCat", "exePath", exePath)
		if err := launchNapCatInUserSession(exePath, napcatDir); err != nil {
			logger.Error("在用户会话中启动 NapCat 失败", "error", err)
			return []byte(err.Error()), err
		}
		logger.Info("NapCat 用户会话启动请求已发送")
		return []byte("NapCat Shell restarted"), nil
	}
	// Linux/macOS 走 Docker 重启流程。
	logger.Info("准备重启 NapCat Docker 容器", "container", "openclaw-qq")
	return dockerOutput("restart", "openclaw-qq")
}

// ensureNapCatNetworkConfig 为 NapCat 配置目录内已知 QQ UIN 写入 OneBot 网络配置（WS+HTTP）。
// NapCat 启动时读取 onebot11_<uin>.json；若 network 为空则 3001/3000 不会监听。
func ensureNapCatNetworkConfig(napcatShellDir string) {
	// 定位内层 napcat 目录（包含 config/）。
	innerDir := findNapCatInnerDir(napcatShellDir)
	if innerDir == "" {
		innerDir = napcatShellDir
	}
	cfgDir := filepath.Join(innerDir, "config")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		log.Printf("[NapCat] ensureNapCatNetworkConfig: mkdir %s: %v", cfgDir, err)
		return
	}

	networkCfg := map[string]interface{}{
		"network": map[string]interface{}{
			"httpServers": []interface{}{
				map[string]interface{}{
					"name":              "ClawPanel-HTTP",
					"enable":            true,
					"port":              3000,
					"host":              "0.0.0.0",
					"enableCors":        true,
					"enableWebsocket":   false,
					"messagePostFormat": "array",
					"token":             "",
					"debug":             false,
				},
			},
			"httpSseServers": []interface{}{},
			"httpClients":    []interface{}{},
			"websocketServers": []interface{}{
				map[string]interface{}{
					"name":                 "ClawPanel-WS",
					"enable":               true,
					"port":                 3001,
					"host":                 "0.0.0.0",
					"messagePostFormat":    "array",
					"token":                "",
					"reportSelfMessage":    false,
					"enableForcePushEvent": true,
					"debug":                false,
					"heartInterval":        30000,
				},
			},
			"websocketClients": []interface{}{},
			"plugins":          []interface{}{},
		},
		"musicSignUrl":        "",
		"enableLocalFile2Url": false,
		"parseMultMsg":        false,
		"imageDownloadProxy":  "",
	}
	data, _ := json.MarshalIndent(networkCfg, "", "  ")

	// 为已有 onebot11_<uin>.json 写配置，同时准备默认配置。
	entries, _ := os.ReadDir(cfgDir)
	uins := []string{}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "onebot11_") && strings.HasSuffix(name, ".json") {
			uin := strings.TrimSuffix(strings.TrimPrefix(name, "onebot11_"), ".json")
			if uin != "" {
				uins = append(uins, uin)
			}
		}
	}
	// 同时检查 napcat_<uin>.json，补齐尚未生成 onebot11 文件的 UIN。
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "napcat_") && strings.HasSuffix(name, ".json") &&
			!strings.HasPrefix(name, "napcat_protocol_") {
			uin := strings.TrimSuffix(strings.TrimPrefix(name, "napcat_"), ".json")
			if uin != "" {
				found := false
				for _, u := range uins {
					if u == uin {
						found = true
						break
					}
				}
				if !found {
					uins = append(uins, uin)
				}
			}
		}
	}

	// 至少写入一份默认配置。
	if len(uins) == 0 {
		p := filepath.Join(cfgDir, "onebot11.json")
		if err := os.WriteFile(p, data, 0644); err != nil {
			log.Printf("[NapCat] write %s: %v", p, err)
		} else {
			log.Printf("[NapCat] wrote default network config: %s", p)
		}
		return
	}

	for _, uin := range uins {
		p := filepath.Join(cfgDir, "onebot11_"+uin+".json")
		// 仅在 network 为空时覆盖，避免破坏用户自定义配置。
		existing, err := os.ReadFile(p)
		if err == nil {
			var cur map[string]interface{}
			if json.Unmarshal(existing, &cur) == nil {
				if net, ok := cur["network"].(map[string]interface{}); ok {
					wsServers, _ := net["websocketServers"].([]interface{})
					httpServers, _ := net["httpServers"].([]interface{})
					if len(wsServers) > 0 || len(httpServers) > 0 {
						log.Printf("[NapCat] network config already set for UIN %s, skipping", uin)
						continue
					}
				}
			}
		}
		if err := os.WriteFile(p, data, 0644); err != nil {
			log.Printf("[NapCat] write %s: %v", p, err)
		} else {
			log.Printf("[NapCat] wrote network config for UIN %s: %s", uin, p)
		}
	}
}

// findQQInstallDir 通过运行进程与常见路径推断 QQ.exe 所在目录。
func findQQInstallDir() string {
	// 先通过 WMIC 从运行进程解析 QQ.exe 路径。
	out, err := exec.Command("wmic", "process", "where", "name='QQ.exe'", "get", "ExecutablePath", "/VALUE").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(strings.ToLower(line), "executablepath=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 && parts[1] != "" {
					return filepath.Dir(strings.TrimSpace(parts[1]))
				}
			}
		}
	}
	// 回退到常见安装目录。
	for _, p := range []string{`C:\Program Files\Tencent\QQ`, `D:\QQ`, `C:\QQ`, `D:\Program Files\Tencent\QQ`} {
		if _, err := os.Stat(filepath.Join(p, "QQ.exe")); err == nil {
			return p
		}
	}
	return ""
}

// findNapCatShellDir 在 Windows 上查找 NapCat Shell 安装目录。
func findNapCatShellDir(cfg *MonitorConfig) string {
	dataDir := ""
	if cfg != nil {
		dataDir = cfg.DataDir
	}
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(dataDir, "napcat"),
		filepath.Join(home, "NapCat"),
		filepath.Join(home, "Desktop", "NapCat"),
		`C:\NapCat`,
		filepath.Join(home, "AppData", "Local", "NapCat"),
	}
	// 不加入 QQ 安装目录候选：NapCat Shell 与 QQ 目录独立。
	// QQ 子目录中可能包含同名文件，会导致误命中。
	markers := []string{"napcat.bat", "NapCatWinBootMain.exe"}
	for _, dir := range candidates {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		found := ""
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || found != "" {
				return filepath.SkipDir
			}
			if !d.IsDir() {
				for _, m := range markers {
					if strings.EqualFold(d.Name(), m) {
						found = filepath.Dir(path)
						return filepath.SkipAll
					}
				}
			}
			return nil
		})
		if found != "" {
			return found
		}
	}
	return ""
}

// getInteractiveUsername 返回当前交互会话登录用户名。
func getInteractiveUsername() string {
	// WMIC 输出是 UTF-16LE，需要先正确解码。
	out, err := exec.Command("wmic", "computersystem", "get", "UserName", "/VALUE").Output()
	if err == nil {
		text := decodeUTF16LE(out)
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(strings.ToLower(line), "username=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}
	// 回退方案：通过 qwinsta.exe 解析活跃控制台会话。
	out2, err := exec.Command(`C:\Windows\System32\qwinsta.exe`).Output()
	if err == nil {
		for _, line := range strings.Split(string(out2), "\n") {
			if strings.Contains(line, "Active") || strings.Contains(line, "活动") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					u := strings.TrimPrefix(fields[0], ">")
					if u != "" && !strings.EqualFold(u, "services") && !strings.EqualFold(u, "console") && !strings.HasPrefix(strings.ToLower(u), "rdp-") {
						// qwinsta 首字段可能是会话名而非用户名，必要时取下一列。
						if strings.HasPrefix(strings.ToLower(u), "console") || strings.HasPrefix(strings.ToLower(u), "rdp") {
							u = fields[1]
						}
						return u
					}
				}
			}
		}
	}
	return ""
}

// decodeUTF16LE 将 UTF-16LE 字节流转换为 UTF-8 字符串。
// 用于处理 WMIC 默认 UTF-16LE 输出。
func decodeUTF16LE(b []byte) string {
	// 去除 BOM。
	if len(b) >= 2 && b[0] == 0xFF && b[1] == 0xFE {
		b = b[2:]
	}
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	u16 := make([]uint16, len(b)/2)
	for i := range u16 {
		u16[i] = uint16(b[2*i]) | uint16(b[2*i+1])<<8
	}
	var sb strings.Builder
	for i := 0; i < len(u16); {
		c := rune(u16[i])
		i++
		if c >= 0xD800 && c <= 0xDBFF && i < len(u16) {
			low := rune(u16[i])
			if low >= 0xDC00 && low <= 0xDFFF {
				c = 0x10000 + (c-0xD800)*0x400 + (low - 0xDC00)
				i++
			}
		}
		sb.WriteRune(c)
	}
	return sb.String()
}

// isPortReachable 检测本地指定 TCP 端口是否可连接。
func isPortReachable(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// checkQQLoginStatus 对 QQ 登录状态进行带重试的一次检查。
func checkQQLoginStatus(cfg *MonitorConfig) (loggedIn bool, nickname string, qqID string) {
	// 失败后重试一次，降低瞬时超时导致的误判概率。
	for attempt := 0; attempt < 2; attempt++ {
		loggedIn, nickname, qqID = doCheckQQLoginStatus(cfg)
		if loggedIn {
			return
		}
		if attempt == 0 {
			time.Sleep(2 * time.Second)
		}
	}
	return
}

// readTokenFromNapCatLogs 通过扫描日志目录提取 WebUI Token。
// 日志中会记录形如 "[WebUi] WebUi Token: <token>" 的条目。
func readTokenFromNapCatLogs(bootmainDir string) string {
	// 从 bootmain 目录向下查找包含 .log 文件的 logs 目录。
	logsDir := ""
	filepath.WalkDir(bootmainDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || logsDir != "" {
			return nil
		}
		if d.IsDir() && strings.EqualFold(d.Name(), "logs") {
			entries, _ := os.ReadDir(path)
			for _, e := range entries {
				if strings.HasSuffix(strings.ToLower(e.Name()), ".log") {
					logsDir = path
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	// 向上级目录回溯查找 logs 目录。
	if logsDir == "" {
		dir := bootmainDir
		for i := 0; i < 8; i++ {
			candidate := filepath.Join(dir, "logs")
			if entries, err := os.ReadDir(candidate); err == nil {
				for _, e := range entries {
					if strings.HasSuffix(strings.ToLower(e.Name()), ".log") {
						logsDir = candidate
						break
					}
				}
			}
			if logsDir != "" {
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if logsDir == "" {
		return ""
	}
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return ""
	}
	var newest os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".log") {
			if newest == nil || e.Name() > newest.Name() {
				newest = e
			}
		}
	}
	if newest == nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(logsDir, newest.Name()))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if idx := strings.Index(line, "[WebUi] WebUi Token: "); idx >= 0 {
			tok := strings.TrimSpace(line[idx+len("[WebUi] WebUi Token: "):])
			if tok != "" {
				return tok
			}
		}
	}
	return ""
}

// doCheckQQLoginStatus 执行一次实际登录态检查并尝试读取账号信息。
func doCheckQQLoginStatus(cfg *MonitorConfig) (loggedIn bool, nickname string, qqID string) {
	client := &http.Client{Timeout: 5 * time.Second}

	// 凭据缓存未过期（5 分钟内）则复用。
	// NapCat 4.x 每次启动会刷新 token，失效后需重新认证。
	cred := ""
	if cachedMonitorCred != "" && time.Since(cachedMonitorCredTime) < 5*time.Minute {
		cred = cachedMonitorCred
	} else {
		// 先清空缓存，确保使用当前启动后的新 token 重新认证。
		cachedMonitorCred = ""
		// 获取 WebUI token：Windows 优先从日志读取，其他平台从容器配置读取。
		token := ""
		if runtime.GOOS == "windows" {
			napcatDir := findNapCatShellDir(cfg)
			if napcatDir != "" {
				// NapCat 4.x：优先读取最新日志中的 token。
				if tok := readTokenFromNapCatLogs(napcatDir); tok != "" {
					token = tok
				}
				// 回退到 bootmain config/webui.json。
				if token == "" {
					webuiPath := filepath.Join(napcatDir, "config", "webui.json")
					if data, err := os.ReadFile(webuiPath); err == nil {
						var webui map[string]interface{}
						if json.Unmarshal(data, &webui) == nil {
							if t, ok := webui["token"].(string); ok && t != "" {
								token = t
							}
						}
					}
				}
			}
		} else {
			out, err := dockerOutput("exec", "openclaw-qq", "cat", "/app/napcat/config/webui.json")
			if err == nil {
				var webui map[string]interface{}
				if json.Unmarshal(out, &webui) == nil {
					if t, ok := webui["token"].(string); ok && t != "" {
						token = t
					}
				}
			}
		}
		if token == "" {
			return false, "", ""
		}

		hash := sha256.Sum256([]byte(token + ".napcat"))
		hashStr := fmt.Sprintf("%x", hash)
		loginBody := fmt.Sprintf(`{"hash":"%s"}`, hashStr)
		resp, err := client.Post("http://127.0.0.1:6099/api/auth/login", "application/json", strings.NewReader(loginBody))
		if err != nil {
			return false, "", ""
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var loginResp map[string]interface{}
		if json.Unmarshal(body, &loginResp) != nil {
			return false, "", ""
		}
		if code, ok := loginResp["code"].(float64); ok && code == 0 {
			if data, ok := loginResp["data"].(map[string]interface{}); ok {
				cred, _ = data["Credential"].(string)
			}
		}
		if cred == "" {
			return false, "", ""
		}
		cachedMonitorCred = cred
		cachedMonitorCredTime = time.Now()
	}

	// 检查 QQ 登录状态。
	req, _ := http.NewRequest("POST", "http://127.0.0.1:6099/api/QQLogin/CheckLoginStatus", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cred)
	resp2, err := client.Do(req)
	if err != nil {
		// HTTP 调用失败时清理缓存，下一轮重新认证。
		cachedMonitorCred = ""
		return false, "", ""
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	var statusResp map[string]interface{}
	if json.Unmarshal(body2, &statusResp) != nil {
		return false, "", ""
	}

	// 遇到未授权时清空凭据，并立即用新 token 重试一次。
	if code, ok := statusResp["code"].(float64); ok && code == -1 {
		cachedMonitorCred = ""
		cachedMonitorCredTime = time.Time{}

		freshToken := ""
		out, err := dockerOutput("exec", "openclaw-qq", "cat", "/app/napcat/config/webui.json")
		if err == nil {
			var webui map[string]interface{}
			if json.Unmarshal(out, &webui) == nil {
				if t, ok := webui["token"].(string); ok && t != "" {
					freshToken = t
				}
			}
		}
		if freshToken == "" {
			return false, "", ""
		}
		hash := sha256.Sum256([]byte(freshToken + ".napcat"))
		loginBody := fmt.Sprintf(`{"hash":"%x"}`, hash)
		respRetry, err := client.Post("http://127.0.0.1:6099/api/auth/login", "application/json", strings.NewReader(loginBody))
		if err != nil {
			return false, "", ""
		}
		defer respRetry.Body.Close()
		bodyRetry, _ := io.ReadAll(respRetry.Body)
		var loginRespRetry map[string]interface{}
		if json.Unmarshal(bodyRetry, &loginRespRetry) != nil {
			return false, "", ""
		}
		if codeRetry, ok := loginRespRetry["code"].(float64); !ok || codeRetry != 0 {
			return false, "", ""
		}
		dataRetry, _ := loginRespRetry["data"].(map[string]interface{})
		freshCred, _ := dataRetry["Credential"].(string)
		if freshCred == "" {
			return false, "", ""
		}
		cachedMonitorCred = freshCred
		cachedMonitorCredTime = time.Now()

		reqRetry, _ := http.NewRequest("POST", "http://127.0.0.1:6099/api/QQLogin/CheckLoginStatus", nil)
		reqRetry.Header.Set("Content-Type", "application/json")
		reqRetry.Header.Set("Authorization", "Bearer "+freshCred)
		respStatusRetry, err := client.Do(reqRetry)
		if err != nil {
			return false, "", ""
		}
		defer respStatusRetry.Body.Close()
		bodyStatusRetry, _ := io.ReadAll(respStatusRetry.Body)
		if json.Unmarshal(bodyStatusRetry, &statusResp) != nil {
			return false, "", ""
		}
		if codeStatusRetry, ok := statusResp["code"].(float64); !ok || codeStatusRetry != 0 {
			return false, "", ""
		}
	}

	if code, ok := statusResp["code"].(float64); !ok || code != 0 {
		return false, "", ""
	}
	statusData, _ := statusResp["data"].(map[string]interface{})
	if statusData == nil {
		return false, "", ""
	}
	isLogin, _ := statusData["isLogin"].(bool)
	if !isLogin {
		return false, "", ""
	}

	// 获取登录账号信息。
	req3, _ := http.NewRequest("POST", "http://127.0.0.1:6099/api/QQLogin/GetQQLoginInfo", nil)
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Authorization", "Bearer "+cred)
	resp3, err := client.Do(req3)
	if err != nil {
		return true, "", ""
	}
	defer resp3.Body.Close()
	body3, _ := io.ReadAll(resp3.Body)
	var infoResp map[string]interface{}
	if json.Unmarshal(body3, &infoResp) != nil {
		return true, "", ""
	}
	if infoCode, ok := infoResp["code"].(float64); ok && infoCode == 0 {
		infoData, _ := infoResp["data"].(map[string]interface{})
		if infoData != nil {
			nickname, _ = infoData["nick"].(string)
			if uid, ok := infoData["uin"].(float64); ok {
				qqID = fmt.Sprintf("%.0f", uid)
			}
			if uid, ok := infoData["uin"].(string); ok {
				qqID = uid
			}
		}
	}

	return true, nickname, qqID
}

// dockerOutput 依次尝试多个 docker 可执行路径并返回命令输出。
func dockerOutput(args ...string) ([]byte, error) {
	bins := []string{"docker", "/usr/local/bin/docker", "/opt/homebrew/bin/docker"}
	for _, bin := range bins {
		cmd := exec.Command(bin, args...)
		cmd.Env = dockerEnv()
		if out, err := cmd.CombinedOutput(); err == nil {
			return out, nil
		}
		if runtime.GOOS == "darwin" {
			for _, archFlag := range []string{"-arm64", "-x86_64"} {
				altArgs := append([]string{archFlag, bin}, args...)
				alt := exec.Command("arch", altArgs...)
				alt.Env = dockerEnv()
				if out, err := alt.CombinedOutput(); err == nil {
					return out, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("docker not available")
}

// dockerRun 执行 docker 命令并忽略标准输出，仅返回错误。
func dockerRun(args ...string) error {
	_, err := dockerOutput(args...)
	return err
}

// dockerEnv 构建执行 docker 命令所需的环境变量（PATH/HOME）。
func dockerEnv() []string {
	home := os.Getenv("HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	if home == "" {
		if runtime.GOOS == "darwin" {
			home = "/var/root"
		} else {
			home = "/root"
		}
	}
	path := os.Getenv("PATH")
	extra := "/usr/local/bin:/usr/local/sbin:/usr/bin:/bin:/usr/sbin:/sbin:/opt/homebrew/bin:/opt/homebrew/sbin"
	if path == "" {
		path = extra
	} else {
		path = path + ":" + extra
	}
	return append(os.Environ(), "PATH="+path, "HOME="+home)
}
