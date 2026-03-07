//go:build !windows

package updater

import "syscall"

// sysProcAttr 返回 Unix 平台的子进程属性，确保更新器子进程独立进程组运行。
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
