//go:build !windows

package taskman

import (
	"os/exec"
	"syscall"
)

// configureTaskCommand 配置命令为独立进程组，便于暂停/取消整个任务树。
func configureTaskCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// pauseProcess 暂停当前任务进程组。
func pauseProcess(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGSTOP)
}

// resumeProcess 恢复当前任务进程组。
func resumeProcess(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGCONT)
}

// terminateProcess 终止当前任务进程组。
func terminateProcess(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
