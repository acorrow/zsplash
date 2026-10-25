//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

func prepareChild(cmd *exec.Cmd, interactive bool) {}

func signalChild(cmd *exec.Cmd, sig syscall.Signal) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	if sig == syscall.SIGKILL {
		_ = cmd.Process.Kill()
		return
	}
	_ = cmd.Process.Signal(sig)
}
