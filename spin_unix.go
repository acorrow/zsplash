//go:build !windows && !linux

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func spinnerSignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT}
}

func prepareChild(cmd *exec.Cmd, hasTTY bool) {
	if shouldIsolateProcessGroup(hasTTY) {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
}

func signalChild(cmd *exec.Cmd, sig syscall.Signal) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	if cmd.SysProcAttr != nil && cmd.SysProcAttr.Setpgid {
		_ = syscall.Kill(-cmd.Process.Pid, sig)
		return
	}
	_ = cmd.Process.Signal(sig)
}
