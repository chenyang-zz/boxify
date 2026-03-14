package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var versionTokenRe = regexp.MustCompile(`^v?(\d+)(?:\.(\d+))?(?:\.(\d+))?`)

var runtimePathsByHome sync.Map // map[string][]string

// BuildExecEnv 构建可执行命令环境，补全 HOME/PATH。
func BuildExecEnv() []string {
	home := runtimeHomeDir()
	path := BuildAugmentedPath(os.Getenv("PATH"))
	env := os.Environ()
	env = append(env, "HOME="+home, "PATH="+path)
	if runtime.GOOS == "windows" {
		env = append(env, "USERPROFILE="+home)
	}
	return env
}

// BuildAugmentedPath 补充常见 node/runtime 二进制路径。
func BuildAugmentedPath(currentPath string) string {
	sep := ":"
	if runtime.GOOS == "windows" {
		sep = ";"
	}
	var parts []string
	if currentPath != "" {
		parts = append(parts, strings.Split(currentPath, sep)...)
	}
	for _, home := range candidateHomes() {
		parts = append(parts, runtimeExtraBinPaths(home)...)
	}
	return strings.Join(dedupeNonEmpty(parts), sep)
}

// DetectOpenClawBinaryPath 检测 openclaw 可执行文件路径。
func DetectOpenClawBinaryPath() string {
	return detectRuntimeBinaryPath("openclaw", "openclaw.cmd", []string{
		"/usr/local/bin/openclaw",
		"/usr/bin/openclaw",
		"/snap/bin/openclaw",
		"/opt/homebrew/bin/openclaw",
	})
}

// DetectNodeBinaryPath 检测 node 可执行文件路径。
func DetectNodeBinaryPath() string {
	return detectRuntimeBinaryPath("node", "node.exe", []string{
		"/usr/local/bin/node",
		"/usr/bin/node",
		"/snap/bin/node",
		"/opt/homebrew/bin/node",
	})
}

// DetectNpmBinaryPath 检测 npm 可执行文件路径。
func DetectNpmBinaryPath() string {
	return detectRuntimeBinaryPath("npm", "npm.cmd", []string{
		"/usr/local/bin/npm",
		"/usr/bin/npm",
		"/snap/bin/npm",
		"/opt/homebrew/bin/npm",
	})
}

// detectRuntimeBinaryPath 结合扩展 PATH 与常见目录检测运行时命令。
func detectRuntimeBinaryPath(commandName, windowsCommandName string, unixCandidates []string) string {
	if p, err := exec.LookPath(commandName); err == nil && p != "" {
		return p
	}

	exeName := commandName
	if runtime.GOOS == "windows" {
		exeName = windowsCommandName
	}

	var candidates []string
	home := runtimeHomeDir()
	for _, h := range candidateHomes() {
		for _, p := range runtimeExtraBinPaths(h) {
			candidates = append(candidates, filepath.Join(p, exeName))
		}
	}
	if runtime.GOOS == "windows" {
		candidates = append(candidates,
			filepath.Join(home, "AppData", "Roaming", "npm", exeName),
			filepath.Join(home, ".local", "bin", exeName),
			filepath.Join(`C:\Program Files\nodejs`, exeName),
		)
	} else {
		candidates = append(candidates, unixCandidates...)
	}

	for _, c := range dedupeNonEmpty(candidates) {
		if fileExists(c) {
			return c
		}
	}
	return ""
}

// runtimeHomeDir 获取运行时 HOME，兼容服务态下的空 HOME 场景。
func runtimeHomeDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = os.Getenv("HOME")
	}
	if home == "" {
		switch runtime.GOOS {
		case "darwin":
			home = "/var/root"
		case "windows":
			home = os.Getenv("USERPROFILE")
			if home == "" {
				home = `C:\Users\Administrator`
			}
		default:
			home = "/root"
		}
	}
	return home
}

// candidateHomes 返回用于扫描 runtime 二进制目录的 home 候选集合。
func candidateHomes() []string {
	homes := []string{runtimeHomeDir()}
	if runtime.GOOS == "windows" {
		return dedupeNonEmpty(homes)
	}
	if runtimeHomeDir() != "/root" {
		homes = append(homes, "/root")
	}
	if entries, err := os.ReadDir("/home"); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				homes = append(homes, filepath.Join("/home", e.Name()))
			}
		}
	}
	return dedupeNonEmpty(homes)
}

// runtimeExtraBinPaths 按 home 读取缓存后的扩展 PATH 目录。
func runtimeExtraBinPaths(home string) []string {
	if cached, ok := runtimePathsByHome.Load(home); ok {
		if paths, ok := cached.([]string); ok {
			return paths
		}
	}
	paths := computeRuntimeExtraBinPaths(home)
	runtimePathsByHome.Store(home, paths)
	return paths
}

// computeRuntimeExtraBinPaths 计算 node/nvm/fnm 等常见运行时目录。
func computeRuntimeExtraBinPaths(home string) []string {
	if runtime.GOOS == "windows" {
		return dedupeNonEmpty([]string{
			filepath.Join(home, "AppData", "Roaming", "npm"),
			filepath.Join(home, ".local", "bin"),
			`C:\Program Files\nodejs`,
		})
	}

	paths := []string{
		"/usr/local/bin", "/usr/local/sbin", "/usr/bin", "/bin", "/usr/sbin", "/sbin",
		"/snap/bin", "/opt/homebrew/bin", "/opt/homebrew/sbin",
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, ".npm-global", "bin"),
		filepath.Join(home, ".volta", "bin"),
		filepath.Join(home, ".asdf", "shims"),
		filepath.Join(home, ".bun", "bin"),
		filepath.Join(home, ".local", "share", "fnm", "current", "bin"),
		filepath.Join(home, ".fnm", "current", "bin"),
	}

	paths = append(paths, globVersionedDirs(filepath.Join(home, ".nvm", "versions", "node", "*", "bin"))...)
	paths = append(paths, globVersionedDirs(filepath.Join(home, ".local", "share", "fnm", "node-versions", "*", "installation", "bin"))...)
	paths = append(paths, globVersionedDirs(filepath.Join(home, ".fnm", "node-versions", "*", "installation", "bin"))...)

	if uid := os.Geteuid(); uid >= 0 {
		paths = append(paths, globVersionedDirs(filepath.Join("/run", "user", fmt.Sprintf("%d", uid), "fnm_multishells", "*", "bin"))...)
	}
	return dedupeNonEmpty(paths)
}

// globVersionedDirs 按语义版本降序返回匹配目录列表。
func globVersionedDirs(pattern string) []string {
	items, err := filepath.Glob(pattern)
	if err != nil || len(items) == 0 {
		return nil
	}
	sort.SliceStable(items, func(i, j int) bool {
		ai, aj := extractVersionTuple(items[i]), extractVersionTuple(items[j])
		if ai[0] != aj[0] {
			return ai[0] > aj[0]
		}
		if ai[1] != aj[1] {
			return ai[1] > aj[1]
		}
		if ai[2] != aj[2] {
			return ai[2] > aj[2]
		}
		return items[i] > items[j]
	})
	return items
}

// extractVersionTuple 从路径片段中提取主次补丁版本号。
func extractVersionTuple(path string) [3]int {
	parts := strings.Split(path, string(os.PathSeparator))
	for _, p := range parts {
		m := versionTokenRe.FindStringSubmatch(p)
		if len(m) == 0 {
			continue
		}
		var t [3]int
		for i := 1; i <= 3; i++ {
			if i < len(m) && m[i] != "" {
				if v, err := strconv.Atoi(m[i]); err == nil {
					t[i-1] = v
				}
			}
		}
		return t
	}
	return [3]int{0, 0, 0}
}

// dedupeNonEmpty 去除空字符串并保持首次出现顺序去重。
func dedupeNonEmpty(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

// fileExists 判断文件路径是否存在且不是目录。
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
