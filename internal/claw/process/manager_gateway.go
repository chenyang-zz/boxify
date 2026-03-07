package process

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const gatewayProbeCacheTTL = 3 * time.Second

var (
	tailnetIPv4Net = mustCIDR("100.64.0.0/10")
	tailnetIPv6Net = mustCIDR("fd7a:115c:a1e0::/48")
)

// gatewayListening 返回网关探测结果，并使用短期缓存降低轮询开销。
func (m *Manager) gatewayListening(force bool) bool {
	if !force {
		m.mu.RLock()
		cachedAt := m.lastGatewayProbeAt
		cachedOK := m.lastGatewayProbeOK
		m.mu.RUnlock()
		if !cachedAt.IsZero() && time.Since(cachedAt) < gatewayProbeCacheTTL {
			return cachedOK
		}
	}

	ok := m.detectGatewayListening()
	m.mu.Lock()
	m.lastGatewayProbeAt = time.Now()
	m.lastGatewayProbeOK = ok
	m.mu.Unlock()
	return ok
}

// detectGatewayListening 在候选 host 列表上执行网关探测。
func (m *Manager) detectGatewayListening() bool {
	port, hosts := m.getGatewayProbeTargets()
	if port == "" {
		return false
	}
	probe := m.gatewayProbe
	if probe == nil {
		probe = m.isOpenClawGateway
	}
	for _, host := range hosts {
		if probe(host, port) {
			return true
		}
	}
	return false
}

// getGatewayPort 从 openclaw.json 读取网关端口，缺省回退 18789。
func (m *Manager) getGatewayPort() string {
	if gw := m.readGatewayConfig(); gw != nil {
		if port, ok := gw["port"].(float64); ok && port > 0 {
			return fmt.Sprintf("%d", int(port))
		}
	}
	return "18789"
}

// readGatewayConfig 读取并返回 gateway 节点配置。
func (m *Manager) readGatewayConfig() map[string]interface{} {
	ocDir := m.resolveOpenClawDir()
	if strings.TrimSpace(ocDir) == "" {
		return nil
	}
	cfgPath := filepath.Join(ocDir, "openclaw.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil
	}
	var cfg map[string]interface{}
	if json.Unmarshal(data, &cfg) != nil {
		return nil
	}
	if gw, ok := cfg["gateway"].(map[string]interface{}); ok && gw != nil {
		return gw
	}
	return nil
}

// defaultGatewayLoopbackTargets 返回本地回环默认探测目标。
func defaultGatewayLoopbackTargets() []string {
	return []string{"127.0.0.1", "localhost", "::1"}
}

// getGatewayProbeTargets 解析网关 port 与实际探测 host 列表。
func (m *Manager) getGatewayProbeTargets() (string, []string) {
	port := m.getGatewayPort()
	bind, custom := m.getGatewayBindSettings()
	return port, gatewayConfiguredTargets(bind, custom, collectGatewayCandidateTargets(), m.canBindGatewayHost)
}

// getGatewayPortCheckTargets 解析端口占用检查使用的 host 列表。
func (m *Manager) getGatewayPortCheckTargets() []string {
	bind, custom := m.getGatewayBindSettings()
	return gatewayConfiguredTargets(bind, custom, collectGatewayCandidateTargets(), m.canBindGatewayHost)
}

// getGatewayBindSettings 返回 bind 模式和 custom host。
func (m *Manager) getGatewayBindSettings() (string, string) {
	gw := m.readGatewayConfig()
	if gw == nil {
		return "", ""
	}
	bind, _ := gw["bind"].(string)
	custom, _ := gw["customBindHost"].(string)
	if strings.TrimSpace(custom) == "" {
		if legacy, ok := gw["bindAddress"].(string); ok {
			custom = legacy
		}
	}
	return strings.ToLower(strings.TrimSpace(bind)), custom
}

// gatewayPortCheckTargets 仅用于单元测试验证 bind 规则映射。
func gatewayPortCheckTargets(bind, custom string, allTargets []string) []string {
	return gatewayConfiguredTargets(bind, custom, allTargets, func(string) bool { return true })
}

// gatewayConfiguredTargets 根据 bind 策略和可绑定能力生成目标列表。
func gatewayConfiguredTargets(bind, custom string, allTargets []string, canBindHost func(host string) bool) []string {
	loopbacks := defaultGatewayLoopbackTargets()
	switch strings.ToLower(strings.TrimSpace(bind)) {
	case "", "auto", "loopback":
		if canBindAnyLoopback(canBindHost) {
			return loopbacks
		}
		return allTargets
	case "tailnet":
		if targets := tailnetGatewayTargets(allTargets); len(targets) > 0 {
			return targets
		}
		if canBindAnyLoopback(canBindHost) {
			return loopbacks
		}
		return allTargets
	case "lan":
		return allTargets
	case "custom":
		custom = normalizeGatewayProbeHost(custom)
		if custom == "localhost" {
			if canBindAnyLoopback(canBindHost) {
				return loopbacks
			}
			return allTargets
		}
		if ip := net.ParseIP(custom); ip != nil && ip.IsLoopback() {
			if canBindHost(custom) {
				return []string{custom}
			}
			return allTargets
		}
		if custom != "" {
			if canBindHost(custom) {
				return []string{custom}
			}
			return allTargets
		}
		return allTargets
	default:
		return loopbacks
	}
}

// collectGatewayCandidateTargets 汇总本机回环 + 网卡 IP 作为探测候选。
func collectGatewayCandidateTargets() []string {
	targets := defaultGatewayLoopbackTargets()
	ifaces, err := net.Interfaces()
	if err != nil {
		return targets
	}
	for _, iface := range ifaces {
		addrs, addrErr := iface.Addrs()
		if addrErr != nil {
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				targets = appendGatewayProbeTarget(targets, v.IP.String())
			case *net.IPAddr:
				targets = appendGatewayProbeTarget(targets, v.IP.String())
			}
		}
	}
	return targets
}

// tailnetGatewayTargets 从候选中筛出 tailscale tailnet 地址段。
func tailnetGatewayTargets(allTargets []string) []string {
	targets := make([]string, 0, len(allTargets))
	for _, host := range allTargets {
		ip := net.ParseIP(host)
		if ip == nil {
			continue
		}
		if tailnetIPv4Net.Contains(ip) || tailnetIPv6Net.Contains(ip) {
			targets = appendGatewayProbeTarget(targets, host)
		}
	}
	return targets
}

// mustCIDR 在初始化阶段解析 CIDR，失败直接 panic 以暴露配置错误。
func mustCIDR(cidr string) *net.IPNet {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	return network
}

// canBindAnyLoopback 判断当前环境是否至少可绑定一个 loopback 地址。
func canBindAnyLoopback(canBindHost func(host string) bool) bool {
	if canBindHost == nil {
		return true
	}
	for _, host := range defaultGatewayLoopbackTargets() {
		if canBindHost(host) {
			return true
		}
	}
	return false
}

// canBindGatewayHost 在 runtime 验证 host:0 是否可监听。
func (m *Manager) canBindGatewayHost(host string) bool {
	if m.bindHostCheck != nil {
		return m.bindHostCheck(host)
	}
	ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// appendGatewayProbeTarget 追加候选目标并去重。
func appendGatewayProbeTarget(targets []string, host string) []string {
	host = normalizeGatewayProbeHost(host)
	if host == "" {
		return targets
	}
	for _, existing := range targets {
		if existing == host {
			return targets
		}
	}
	return append(targets, host)
}

// normalizeGatewayProbeHost 规范化 host 表达，过滤非法/不可探测值。
func normalizeGatewayProbeHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if strings.Contains(host, "://") {
		if parsed, err := url.Parse(host); err == nil {
			host = parsed.Hostname()
		}
	}
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		if parsedHost, _, err := net.SplitHostPort(host); err == nil {
			host = parsedHost
		}
	}
	host = strings.Trim(host, "[]")
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsUnspecified() || ip.IsMulticast() {
			return ""
		}
		return ip.String()
	}
	return host
}

// isOpenClawGateway 通过首页关键字识别是否为 OpenClaw gateway。
func (m *Manager) isOpenClawGateway(host, port string) bool {
	client := &http.Client{Timeout: 1500 * time.Millisecond, Transport: &http.Transport{}}
	u := (&url.URL{Scheme: "http", Host: net.JoinHostPort(host, port), Path: "/"}).String()
	resp, err := client.Get(u)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return false
	}
	text := strings.ToLower(string(body))
	return strings.Contains(text, "openclaw control") || strings.Contains(text, "<openclaw-app")
}
