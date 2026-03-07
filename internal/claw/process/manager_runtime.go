package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// findOpenClawBin 搜索 openclaw 可执行文件路径。
func (m *Manager) findOpenClawBin() string {
	if p := DetectOpenClawBinaryPath(); p != "" {
		return p
	}

	candidates := []string{"openclaw"}
	home, _ := os.UserHomeDir()
	if home != "" {
		candidates = append(candidates,
			filepath.Join(home, ".local", "bin", "openclaw"),
			filepath.Join(home, "openclaw", "app", "openclaw"),
		)
	}

	switch runtime.GOOS {
	case "linux":
		candidates = append(candidates, "/usr/local/bin/openclaw", "/usr/bin/openclaw", "/snap/bin/openclaw")
	case "darwin":
		candidates = append(candidates, "/usr/local/bin/openclaw", "/opt/homebrew/bin/openclaw")
	case "windows":
		candidates = append(candidates,
			`C:\Program Files\openclaw\openclaw.exe`,
			filepath.Join(home, "AppData", "Roaming", "npm", "openclaw.cmd"),
		)
	}

	for _, c := range candidates {
		if p, err := exec.LookPath(c); err == nil {
			return p
		}
	}
	return ""
}

// killWindowsPortListeners 通过 netstat/taskkill 清理指定端口监听进程。
func (m *Manager) killWindowsPortListeners(port string) int {
	cmd := exec.Command("cmd", "/C", "netstat -ano -p tcp")
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	pidSet := map[int]struct{}{}
	needle := ":" + port
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, needle) || !strings.Contains(strings.ToUpper(line), "LISTENING") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		pid, atoiErr := strconv.Atoi(fields[len(fields)-1])
		if atoiErr != nil || pid <= 0 {
			continue
		}
		pidSet[pid] = struct{}{}
	}

	killed := 0
	for pid := range pidSet {
		k := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F")
		if err := k.Run(); err == nil {
			killed++
		}
	}
	return killed
}

// resolveOpenClawDir 返回最终使用的 OpenClaw 配置目录。
func (m *Manager) resolveOpenClawDir() string {
	if strings.TrimSpace(m.cfg.OpenClawDir) != "" {
		return m.cfg.OpenClawDir
	}
	home, _ := os.UserHomeDir()
	if strings.TrimSpace(home) == "" {
		return ".openclaw"
	}
	return filepath.Join(home, ".openclaw")
}

// runDockerOutput 执行 docker 命令并兼容 macOS arch 变体。
func runDockerOutput(args ...string) ([]byte, error) {
	bins := []string{"docker", "/usr/local/bin/docker", "/opt/homebrew/bin/docker"}
	for _, bin := range bins {
		cmd := exec.Command(bin, args...)
		cmd.Env = BuildExecEnv()
		if out, err := cmd.Output(); err == nil {
			return out, nil
		}
		if runtime.GOOS == "darwin" {
			for _, archFlag := range []string{"-arm64", "-x86_64"} {
				altArgs := append([]string{archFlag, bin}, args...)
				alt := exec.Command("arch", altArgs...)
				alt.Env = BuildExecEnv()
				if out, err := alt.Output(); err == nil {
					return out, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("docker command unavailable")
}
