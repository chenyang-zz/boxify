//go:build !windows

package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// launchNapCatInUserSession 在非 Windows 平台不支持交互会话启动。
func launchNapCatInUserSession(_, _ string) error {
	return fmt.Errorf("launchNapCatInUserSession is only supported on Windows")
}

// findNapCatInnerDir 扫描目录并定位包含 napcat.mjs 的资源目录。
func findNapCatInnerDir(shellDir string) string {
	var found string
	_ = filepath.WalkDir(shellDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || found != "" {
			return nil
		}
		if !d.IsDir() && strings.EqualFold(d.Name(), "napcat.mjs") {
			found = filepath.Dir(path)
			return filepath.SkipAll
		}
		return nil
	})
	return found
}
