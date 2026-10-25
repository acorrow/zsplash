//go:build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func spinnerSignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}

func prepareChild(cmd *exec.Cmd, hasTTY bool) {}

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
