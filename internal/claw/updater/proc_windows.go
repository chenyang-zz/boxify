//go:build windows

package updater

import "syscall"

// sysProcAttr 返回 Windows 平台子进程属性。
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
