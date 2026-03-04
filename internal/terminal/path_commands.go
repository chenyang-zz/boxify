// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terminal

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// ExecutableCommand PATH 中的可执行命令
type ExecutableCommand struct {
	Name string
	Path string
}

// PathCommandScanner 负责扫描 PATH 可执行命令并处理终端类型解析。
type PathCommandScanner struct {
	logger        *slog.Logger
	shellDetector *ShellDetector
}

// NewPathCommandScanner 创建 PATH 命令扫描器。
func NewPathCommandScanner(logger *slog.Logger, shellDetector *ShellDetector) *PathCommandScanner {
	if logger == nil {
		logger = slog.Default()
	}
	if shellDetector == nil {
		shellDetector = NewShellDetector()
	}

	return &PathCommandScanner{
		logger:        logger,
		shellDetector: shellDetector,
	}
}

// ListExecutableCommandsFromPATH 列出当前进程 PATH 中的可执行命令。
func (s *PathCommandScanner) ListExecutableCommandsFromPATH() []ExecutableCommand {
	pathValue := os.Getenv("PATH")
	pathExt := os.Getenv("PATHEXT")
	s.logger.Debug("开始扫描 PATH 可执行命令", "os", runtime.GOOS)

	commands := listExecutableCommands(pathValue, pathExt, runtime.GOOS, s.logger)
	s.logger.Info("PATH 可执行命令扫描完成", "count", len(commands))
	return commands
}

// ResolveShellType 解析 shell 类型，auto 或空值会解析为系统默认 shell 类型。
func (s *PathCommandScanner) ResolveShellType(shellType ShellType) (ShellType, error) {
	if shellType == "" {
		shellType = ShellTypeAuto
	}

	if shellType == ShellTypeAuto {
		path := s.shellDetector.DetectShell(ShellTypeAuto)
		resolved := s.shellDetector.DetectShellTypeFromPath(path)
		s.logger.Debug("自动解析终端类型", "shellPath", path, "resolvedShell", resolved)
		return resolved, nil
	}

	if !s.IsKnownShellType(shellType) {
		return "", fmt.Errorf("不支持的终端类型: %s", shellType)
	}

	return shellType, nil
}

// IsKnownShellType 判断是否为支持的 shell 类型。
func (s *PathCommandScanner) IsKnownShellType(shellType ShellType) bool {
	switch shellType {
	case ShellTypeCmd, ShellTypePowershell, ShellTypePwsh, ShellTypeBash, ShellTypeZsh, ShellTypeSh, ShellTypeAuto:
		return true
	default:
		return false
	}
}

// GetDefaultCommands 获取 shell 的默认命令（主要为 shell 内建命令）。
func (s *PathCommandScanner) GetDefaultCommands(shellType ShellType) []string {
	switch shellType {
	case ShellTypeCmd:
		return []string{"cd", "dir", "echo", "set", "cls", "type", "copy", "del"}
	case ShellTypePowershell, ShellTypePwsh:
		return []string{"Set-Location", "Get-ChildItem", "Get-Command", "Set-Item", "Get-Content", "Clear-Host", "Remove-Item", "Copy-Item"}
	case ShellTypeBash, ShellTypeZsh, ShellTypeSh:
		return []string{"cd", "pwd", "echo", "export", "alias", "history", "source", "type"}
	default:
		return []string{}
	}
}

// listExecutableCommands 按平台规则扫描 PATH 目录，返回去重后的命令列表。
func listExecutableCommands(pathValue, pathExt, goos string, logger *slog.Logger) []ExecutableCommand {
	dirs := filepath.SplitList(pathValue)
	if len(dirs) == 0 {
		return []ExecutableCommand{}
	}

	results := make([]ExecutableCommand, 0)
	seen := make(map[string]struct{})

	for _, dir := range dirs {
		if strings.TrimSpace(dir) == "" {
			continue
		}

		if goos == "windows" {
			results = collectWindowsCommands(dir, pathExt, results, seen, logger)
			continue
		}

		results = collectUnixCommands(dir, results, seen, logger)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}

// collectUnixCommands 扫描 Unix 目录中的可执行常规文件（带执行位）。
func collectUnixCommands(dir string, results []ExecutableCommand, seen map[string]struct{}, logger *slog.Logger) []ExecutableCommand {
	entries, err := os.ReadDir(dir)
	if err != nil {
		logger.Warn("读取 PATH 目录失败，已跳过", "dir", dir, "error", err)
		return results
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
			continue
		}

		name := entry.Name()
		if _, ok := seen[name]; ok {
			continue
		}

		seen[name] = struct{}{}
		results = append(results, ExecutableCommand{
			Name: name,
			Path: filepath.Join(dir, name),
		})
	}
	return results
}

// collectWindowsCommands 扫描 Windows 目录中的可执行文件（由 PATHEXT 决定）。
func collectWindowsCommands(dir, pathExt string, results []ExecutableCommand, seen map[string]struct{}, logger *slog.Logger) []ExecutableCommand {
	entries, err := os.ReadDir(dir)
	if err != nil {
		logger.Warn("读取 PATH 目录失败，已跳过", "dir", dir, "error", err)
		return results
	}

	allowedExt, extPriority := parseWindowsPathExt(pathExt)
	perDir := make(map[string]ExecutableCommand)
	perDirPriority := make(map[string]int)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		priority, ok := extPriority[ext]
		if !ok || !allowedExt[ext] {
			continue
		}

		base := strings.TrimSuffix(name, filepath.Ext(name))
		if base == "" {
			continue
		}

		existingPriority, exists := perDirPriority[base]
		if !exists || priority < existingPriority {
			perDir[base] = ExecutableCommand{
				Name: base,
				Path: filepath.Join(dir, name),
			}
			perDirPriority[base] = priority
		}
	}

	for name, cmd := range perDir {
		normalizedName := strings.ToLower(name)
		if _, ok := seen[normalizedName]; ok {
			continue
		}
		seen[normalizedName] = struct{}{}
		results = append(results, cmd)
	}

	return results
}

// parseWindowsPathExt 解析 PATHEXT，返回允许扩展名集合及扩展优先级。
func parseWindowsPathExt(pathExt string) (map[string]bool, map[string]int) {
	if strings.TrimSpace(pathExt) == "" {
		pathExt = ".COM;.EXE;.BAT;.CMD"
	}

	parts := strings.Split(pathExt, ";")
	allowed := make(map[string]bool)
	priority := make(map[string]int)

	for idx, raw := range parts {
		ext := strings.TrimSpace(strings.ToLower(raw))
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if _, exists := priority[ext]; !exists {
			priority[ext] = idx
		}
		allowed[ext] = true
	}

	if len(allowed) == 0 {
		allowed[".com"] = true
		allowed[".exe"] = true
		allowed[".bat"] = true
		allowed[".cmd"] = true
		priority[".com"] = 0
		priority[".exe"] = 1
		priority[".bat"] = 2
		priority[".cmd"] = 3
	}

	return allowed, priority
}
