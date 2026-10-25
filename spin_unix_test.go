//go:build !windows

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestPrepareChildProcessGroupDecision(t *testing.T) {
	cmdWithTTY := exec.Command("true")
	prepareChild(cmdWithTTY, true)
	if cmdWithTTY.SysProcAttr != nil && cmdWithTTY.SysProcAttr.Setpgid {
		t.Fatal("expected prepareChild(cmd, true) to keep child in terminal process group (Setpgid=false)")
	}

	cmdWithoutTTY := exec.Command("true")
	prepareChild(cmdWithoutTTY, false)
	if cmdWithoutTTY.SysProcAttr == nil || !cmdWithoutTTY.SysProcAttr.Setpgid {
		t.Fatal("expected prepareChild(cmd, false) to isolate child process group (Setpgid=true)")
	}
}

func TestSpinTerminalPromptInPTY(t *testing.T) {
	if _, err := exec.LookPath("script"); err != nil {
		t.Skip("skipping PTY test: 'script' utility not available")
	}

	binPath := filepath.Join(t.TempDir(), "zsplash")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build zsplash binary: %v\n%s", err, string(out))
	}

	var scriptCmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmdStr := fmt.Sprintf("%q spin --static Prompting -- sh -c 'read x </dev/tty; echo got:$x'", binPath)
		scriptCmd = exec.Command("script", "-q", "-e", "-c", cmdStr, "/dev/null")
	} else {
		scriptCmd = exec.Command("script", "-q", "/dev/null", binPath, "spin", "--static", "Prompting", "--", "sh", "-c", "read x </dev/tty; echo got:$x")
	}

	stdinR, stdinW := io.Pipe()
	scriptCmd.Stdin = stdinR
	var outBuf bytes.Buffer
	scriptCmd.Stdout = &outBuf
	scriptCmd.Stderr = &outBuf

	if err := scriptCmd.Start(); err != nil {
		t.Skipf("skipping PTY test: script failed to start: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- scriptCmd.Wait()
	}()

	go func() {
		time.Sleep(250 * time.Millisecond)
		_, _ = io.WriteString(stdinW, "pty-secret\n")
		time.Sleep(250 * time.Millisecond)
		_ = stdinW.Close()
	}()

	select {
	case err := <-done:
		out := outBuf.String()
		if err != nil && !strings.Contains(out, "got:pty-secret") {
			t.Fatalf("script pty command failed (%v), output:\n%s", err, out)
		}
		if !strings.Contains(out, "got:pty-secret") {
			t.Fatalf("expected pty output to contain 'got:pty-secret', got:\n%s", out)
		}
	case <-time.After(5 * time.Second):
		_ = stdinW.Close()
		_ = scriptCmd.Process.Kill()
		t.Fatalf("spin child reading /dev/tty hung inside pty (SIGTTIN regression); output so far:\n%s", outBuf.String())
	}
}

func TestSpinUnixProcessGroupSignals(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "zsplash")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build zsplash binary: %v\n%s", err, string(out))
	}

	t.Run("SIGTERM stops process group and exits 143", func(t *testing.T) {
		cmd := exec.Command(binPath, "spin", "Testing SIGTERM", "--", "sh", "-c", "echo $$; sleep 30 & sleep 30")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}

		reader := bufio.NewReader(stdout)
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("failed to read child pgid: %v", err)
		}
		pgid, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || pgid <= 0 {
			t.Fatalf("invalid child pgid %q: %v", line, err)
		}
		// Give background sleep a brief moment to fork into the process group
		time.Sleep(100 * time.Millisecond)

		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			t.Fatalf("failed to send SIGTERM: %v", err)
		}

		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()

		select {
		case waitErr := <-done:
			exitErr, ok := waitErr.(*exec.ExitError)
			if !ok {
				t.Fatalf("expected ExitError, got: %v", waitErr)
			}
			if exitErr.ExitCode() != 143 {
				t.Fatalf("expected exit code 143 on SIGTERM, got %d", exitErr.ExitCode())
			}
		case <-time.After(6 * time.Second):
			_ = cmd.Process.Kill()
			t.Fatal("timed out waiting for zsp spin to exit after SIGTERM")
		}

		// Wait briefly for kernel process table reap, then verify no process in pgid remains alive
		time.Sleep(100 * time.Millisecond)
		if err := syscall.Kill(-pgid, 0); err != syscall.ESRCH {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
			t.Fatalf("expected process group %d to be gone (ESRCH), got err=%v", pgid, err)
		}
	})

	t.Run("SIGINT via kill stops child and exits 130", func(t *testing.T) {
		cmd := exec.Command(binPath, "spin", "Testing SIGINT", "--", "sh", "-c", "echo $$; sleep 30")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}

		reader := bufio.NewReader(stdout)
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("failed to read child pgid: %v", err)
		}
		pgid, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || pgid <= 0 {
			t.Fatalf("invalid child pgid %q: %v", line, err)
		}
		time.Sleep(100 * time.Millisecond)

		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			t.Fatalf("failed to send SIGINT: %v", err)
		}

		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()

		select {
		case waitErr := <-done:
			exitErr, ok := waitErr.(*exec.ExitError)
			if !ok {
				t.Fatalf("expected ExitError, got: %v", waitErr)
			}
			if exitErr.ExitCode() != 130 {
				t.Fatalf("expected exit code 130 on SIGINT, got %d", exitErr.ExitCode())
			}
		case <-time.After(6 * time.Second):
			_ = cmd.Process.Kill()
			t.Fatal("timed out waiting for zsp spin to exit after SIGINT")
		}

		time.Sleep(100 * time.Millisecond)
		if err := syscall.Kill(-pgid, 0); err != syscall.ESRCH {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
			t.Fatalf("expected process group %d to be gone (ESRCH), got err=%v", pgid, err)
		}
	})
}
