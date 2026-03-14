//go:build windows

package taskman

import (
	"fmt"
	"os/exec"
	"syscall"
)

// configureTaskCommand 配置 Windows 命令进程属性。
func configureTaskCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

// pauseProcess 暂停当前任务主进程。
func pauseProcess(cmd *exec.Cmd) error {
	return exec.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf("Suspend-Process -Id %d", cmd.Process.Pid)).Run()
}

// resumeProcess 恢复当前任务主进程。
func resumeProcess(cmd *exec.Cmd) error {
	return exec.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf("Resume-Process -Id %d", cmd.Process.Pid)).Run()
}

// terminateProcess 终止当前任务进程树。
func terminateProcess(cmd *exec.Cmd) error {
	return exec.Command("taskkill", "/PID", fmt.Sprintf("%d", cmd.Process.Pid), "/T", "/F").Run()
}
