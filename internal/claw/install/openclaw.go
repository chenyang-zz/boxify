package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	clawprocess "github.com/chenyang-zz/boxify/internal/claw/process"
)

const (
	minNodeMajor      = 22
	minNodeMinor      = 16
	minNodePatch      = 0
	preferredNodeLine = "24"
)

// CommandStep 描述一条可执行安装命令。
type CommandStep struct {
	Name    string   // 步骤名称。
	Command string   // 命令路径。
	Args    []string // 命令参数。
	Script  string   // 需要通过 shell 执行的脚本。
}

// OpenClawSetupInspection 描述自动安装前的环境检查结果。
type OpenClawSetupInspection struct {
	NodeInstalled        bool   // 是否已检测到 node。
	NodeVersionSatisfied bool   // 当前 node 版本是否满足 OpenClaw 最低要求。
	NpmInstalled         bool   // 是否已检测到 npm。
	OpenClawInstalled    bool   // 是否已检测到 openclaw。
	Configured           bool   // 是否已检测到 openclaw.json。
	ConfigDir            string // 实际复用的 OpenClaw 配置目录。
	ConfigPath           string // 实际复用的 openclaw.json 路径。
	NodePath             string // node 可执行文件路径。
	NodeVersion          string // node 版本号。
	NpmPath              string // npm 可执行文件路径。
	OpenClawPath         string // openclaw 可执行文件路径。
	AutoInstallSupported bool   // 当前环境是否支持自动安装。
	AutoInstallHint      string // 当前环境不支持自动安装时的提示。
}

// InspectOpenClawSetup 汇总 OpenClaw 自动安装所需环境状态。
func InspectOpenClawSetup(openClawDir string) OpenClawSetupInspection {
	nodePath := strings.TrimSpace(clawprocess.DetectNodeBinaryPath())
	npmPath := strings.TrimSpace(clawprocess.DetectNpmBinaryPath())
	openClawPath := strings.TrimSpace(clawprocess.DetectOpenClawBinaryPath())
	configDir, configPath := resolveReusableOpenClawConfig(openClawDir)
	nodeVersion, nodeVersionSatisfied := detectNodeVersion(nodePath)

	inspection := OpenClawSetupInspection{
		NodeInstalled:        nodePath != "",
		NodeVersionSatisfied: nodeVersionSatisfied,
		NpmInstalled:         npmPath != "",
		OpenClawInstalled:    openClawPath != "",
		Configured:           readableFileExists(configPath),
		ConfigDir:            configDir,
		ConfigPath:           configPath,
		NodePath:             nodePath,
		NodeVersion:          nodeVersion,
		NpmPath:              npmPath,
		OpenClawPath:         openClawPath,
	}
	if inspection.NodeInstalled && inspection.NodeVersionSatisfied && inspection.NpmInstalled {
		inspection.AutoInstallSupported = true
		return inspection
	}

	steps, hint := ResolveNodeInstallSteps()
	inspection.AutoInstallSupported = len(steps) > 0
	if strings.TrimSpace(hint) == "" && inspection.NodeInstalled && !inspection.NodeVersionSatisfied {
		hint = fmt.Sprintf("当前 Node.js 版本 %s 低于 OpenClaw 要求的 >=22.16.0，推荐切换到 Node.js 24", inspection.NodeVersion)
	}
	inspection.AutoInstallHint = strings.TrimSpace(hint)
	return inspection
}

// ResolveNodeInstallSteps 返回当前系统自动安装 Node.js 所需命令。
func ResolveNodeInstallSteps() ([]CommandStep, string) {
	switch runtime.GOOS {
	case "darwin":
		if fnm := lookupCommand("fnm", "/opt/homebrew/bin/fnm", "/usr/local/bin/fnm"); fnm != "" {
			return []CommandStep{
				{Name: "通过 fnm 安装 Node.js 24", Command: fnm, Args: []string{"install", preferredNodeLine}},
				{Name: "将 Node.js 24 设为 fnm 默认版本", Command: fnm, Args: []string{"default", preferredNodeLine}},
			}, ""
		}
		if script := resolveUnixNvmScript(); script != "" {
			return []CommandStep{{
				Name:   "通过 nvm 安装并启用 Node.js 24",
				Script: script + " && nvm install 24 && nvm alias default 24 && nvm use 24",
			}}, ""
		}
		if brew := lookupCommand("brew", "/opt/homebrew/bin/brew", "/usr/local/bin/brew"); brew != "" {
			return []CommandStep{{
				Name:    "安装 Node.js",
				Command: brew,
				Args:    []string{"install", "node"},
			}}, ""
		}
		return nil, "未检测到 fnm、nvm 或 Homebrew，无法自动安装 Node.js；OpenClaw 要求 Node.js >=22.16.0，推荐使用 Node.js 24"
	case "windows":
		if fnm := lookupCommand("fnm", "fnm.exe"); fnm != "" {
			return []CommandStep{
				{Name: "通过 fnm 安装 Node.js 24", Command: fnm, Args: []string{"install", preferredNodeLine}},
				{Name: "将 Node.js 24 设为 fnm 默认版本", Command: fnm, Args: []string{"default", preferredNodeLine}},
			}, ""
		}
		if nvm := lookupCommand("nvm", "nvm.exe"); nvm != "" {
			return []CommandStep{
				{Name: "通过 nvm 安装 Node.js 24", Command: nvm, Args: []string{"install", "24.0.0"}},
				{Name: "切换到 Node.js 24", Command: nvm, Args: []string{"use", "24.0.0"}},
			}, ""
		}
		if winget := lookupCommand("winget"); winget != "" {
			return []CommandStep{{
				Name:    "安装 Node.js",
				Command: winget,
				Args: []string{
					"install", "--id", "OpenJS.NodeJS", "-e",
					"--accept-source-agreements", "--accept-package-agreements",
				},
			}}, ""
		}
		if choco := lookupCommand("choco"); choco != "" {
			return []CommandStep{{
				Name:    "安装 Node.js",
				Command: choco,
				Args:    []string{"install", "nodejs-lts", "-y"},
			}}, ""
		}
		return nil, "未检测到 fnm、nvm、winget 或 choco，无法自动安装 Node.js；OpenClaw 要求 Node.js >=22.16.0，推荐使用 Node.js 24"
	default:
		return resolveLinuxNodeInstallSteps()
	}
}

// ResolveNpmInstallCommand 返回 npm install -g openclaw@latest 所需命令路径。
func ResolveNpmInstallCommand() string {
	return strings.TrimSpace(clawprocess.DetectNpmBinaryPath())
}

// ResolveOpenClawOnboardCommand 返回 openclaw onboard 所需命令路径。
func ResolveOpenClawOnboardCommand() string {
	return strings.TrimSpace(clawprocess.DetectOpenClawBinaryPath())
}

// resolveLinuxNodeInstallSteps 返回 Linux 平台的 Node.js 安装步骤。
func resolveLinuxNodeInstallSteps() ([]CommandStep, string) {
	if fnm := lookupCommand("fnm", "/usr/local/bin/fnm", "/usr/bin/fnm"); fnm != "" {
		return []CommandStep{
			{Name: "通过 fnm 安装 Node.js 24", Command: fnm, Args: []string{"install", preferredNodeLine}},
			{Name: "将 Node.js 24 设为 fnm 默认版本", Command: fnm, Args: []string{"default", preferredNodeLine}},
		}, ""
	}
	if script := resolveUnixNvmScript(); script != "" {
		return []CommandStep{{
			Name:   "通过 nvm 安装并启用 Node.js 24",
			Script: script + " && nvm install 24 && nvm alias default 24 && nvm use 24",
		}}, ""
	}

	sudo := lookupCommand("sudo", "/usr/bin/sudo", "/bin/sudo")
	if sudo == "" {
		return nil, "未检测到 fnm 或 nvm，且当前环境缺少 sudo，无法继续自动安装 Node.js；OpenClaw 要求 Node.js >=22.16.0，推荐使用 Node.js 24"
	}
	if apt := lookupCommand("apt-get", "/usr/bin/apt-get", "/bin/apt-get"); apt != "" {
		return []CommandStep{
			{Name: "刷新软件源", Command: sudo, Args: []string{"-n", apt, "update"}},
			{Name: "安装 Node.js", Command: sudo, Args: []string{"-n", apt, "install", "-y", "nodejs", "npm"}},
		}, ""
	}
	if dnf := lookupCommand("dnf", "/usr/bin/dnf", "/bin/dnf"); dnf != "" {
		return []CommandStep{
			{Name: "安装 Node.js", Command: sudo, Args: []string{"-n", dnf, "install", "-y", "nodejs", "npm"}},
		}, ""
	}
	if yum := lookupCommand("yum", "/usr/bin/yum", "/bin/yum"); yum != "" {
		return []CommandStep{
			{Name: "安装 Node.js", Command: sudo, Args: []string{"-n", yum, "install", "-y", "nodejs", "npm"}},
		}, ""
	}
	if pacman := lookupCommand("pacman", "/usr/bin/pacman", "/bin/pacman"); pacman != "" {
		return []CommandStep{
			{Name: "安装 Node.js", Command: sudo, Args: []string{"-n", pacman, "-Sy", "--noconfirm", "nodejs", "npm"}},
		}, ""
	}
	if zypper := lookupCommand("zypper", "/usr/bin/zypper", "/bin/zypper"); zypper != "" {
		return []CommandStep{
			{Name: "安装 Node.js", Command: sudo, Args: []string{"-n", zypper, "--non-interactive", "install", "nodejs", "npm"}},
		}, ""
	}
	return nil, "未检测到 fnm 或 nvm，且当前 Linux 环境缺少受支持的包管理器（apt/dnf/yum/pacman/zypper）"
}

// detectNodeVersion 返回 node 版本与是否满足最低要求。
func detectNodeVersion(nodePath string) (string, bool) {
	nodePath = strings.TrimSpace(nodePath)
	if nodePath == "" {
		return "", false
	}

	cmd := exec.Command(nodePath, "--version")
	cmd.Env = clawprocess.BuildExecEnv()
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}

	version := strings.TrimSpace(string(out))
	if version == "" {
		return "", false
	}
	return version, isSupportedNodeVersion(version)
}

// isSupportedNodeVersion 判断版本是否满足 OpenClaw 最低要求。
func isSupportedNodeVersion(version string) bool {
	major, minor, patch, ok := parseVersion(version)
	if !ok {
		return false
	}
	if major != minNodeMajor {
		return major > minNodeMajor
	}
	if minor != minNodeMinor {
		return minor > minNodeMinor
	}
	return patch >= minNodePatch
}

// parseVersion 解析 vX.Y.Z 形式版本号。
func parseVersion(version string) (int, int, int, bool) {
	version = strings.TrimSpace(strings.TrimPrefix(version, "v"))
	if version == "" {
		return 0, 0, 0, false
	}

	parts := strings.Split(version, ".")
	values := [3]int{}
	for i := 0; i < len(values) && i < len(parts); i++ {
		num, err := strconv.Atoi(strings.TrimSpace(parts[i]))
		if err != nil {
			return 0, 0, 0, false
		}
		values[i] = num
	}
	return values[0], values[1], values[2], true
}

// resolveUnixNvmScript 返回可执行 nvm 的初始化脚本。
func resolveUnixNvmScript() string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".nvm", "nvm.sh"),
		"/usr/local/opt/nvm/nvm.sh",
		"/opt/homebrew/opt/nvm/nvm.sh",
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || !readableFileExists(candidate) {
			continue
		}
		return fmt.Sprintf(". %q", candidate)
	}
	return ""
}

// lookupCommand 按候选路径查找可执行文件。
func lookupCommand(name string, candidates ...string) string {
	if p, err := exec.LookPath(name); err == nil && p != "" {
		return p
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if stat, err := os.Stat(candidate); err == nil && !stat.IsDir() {
			return candidate
		}
	}
	return ""
}

// readableFileExists 判断文件是否存在且可读。
func readableFileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

// resolveReusableOpenClawConfig 返回当前应复用的 OpenClaw 配置目录与配置文件路径。
func resolveReusableOpenClawConfig(openClawDir string) (string, string) {
	candidates := candidateOpenClawDirs(openClawDir)
	for _, dir := range candidates {
		configPath := filepath.Join(dir, "openclaw.json")
		if readableFileExists(configPath) {
			return dir, configPath
		}
	}
	if len(candidates) == 0 {
		return "", ""
	}
	return candidates[0], filepath.Join(candidates[0], "openclaw.json")
}

// candidateOpenClawDirs 返回配置探测候选目录，优先当前目录，其次回退用户主目录下的 .openclaw。
func candidateOpenClawDirs(openClawDir string) []string {
	seen := make(map[string]struct{}, 2)
	candidates := make([]string, 0, 2)
	appendDir := func(dir string) {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			return
		}
		cleaned := filepath.Clean(dir)
		if _, exists := seen[cleaned]; exists {
			return
		}
		seen[cleaned] = struct{}{}
		candidates = append(candidates, cleaned)
	}

	appendDir(openClawDir)
	home, _ := os.UserHomeDir()
	if strings.TrimSpace(home) != "" {
		appendDir(filepath.Join(home, ".openclaw"))
	}
	return candidates
}
