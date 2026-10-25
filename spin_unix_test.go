//go:build !windows

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

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
	inCI := os.Getenv("CI") != ""
	if _, err := exec.LookPath("script"); err != nil {
		if inCI {
			t.Fatalf("required 'script' utility not available in CI: %v", err)
		}
		t.Skip("skipping PTY test: 'script' utility not available")
	}

	binPath := filepath.Join(t.TempDir(), "zsplash")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build zsplash binary: %v\n%s", err, string(out))
	}

	var scriptCmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmdStr := fmt.Sprintf("%q spin --static Prompting -- sh -c 'echo READY; read x </dev/tty; echo got:$x'", binPath)
		scriptCmd = exec.Command("script", "-q", "-e", "-c", cmdStr, "/dev/null")
	} else {
		scriptCmd = exec.Command("script", "-q", "/dev/null", binPath, "spin", "--static", "Prompting", "--", "sh", "-c", "echo READY; read x </dev/tty; echo got:$x")
	}

	stdinR, stdinW := io.Pipe()
	defer func() { _ = stdinW.Close() }()
	scriptCmd.Stdin = stdinR
	var outBuf syncBuffer
	scriptCmd.Stdout = &outBuf
	scriptCmd.Stderr = &outBuf

	if err := scriptCmd.Start(); err != nil {
		if inCI {
			t.Fatalf("script failed to start in CI: %v", err)
		}
		t.Skipf("skipping PTY test: script failed to start: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- scriptCmd.Wait()
	}()

	go func() {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if strings.Contains(outBuf.String(), "READY") {
				break
			}
			time.Sleep(25 * time.Millisecond)
		}
		_, _ = io.WriteString(stdinW, "pty-secret\n")
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
	case <-time.After(10 * time.Second):
		_ = stdinW.Close()
		_ = scriptCmd.Process.Kill()
		t.Fatalf("spin child reading /dev/tty hung inside pty (SIGTTIN regression); output so far:\n%s", outBuf.String())
	}
}

func waitForProcessGroupGone(pgid int) error {
	for i := 0; i < 30; i++ {
		err := syscall.Kill(-pgid, 0)
		if err == syscall.ESRCH {
			return syscall.ESRCH
		}
		time.Sleep(50 * time.Millisecond)
	}
	return syscall.Kill(-pgid, 0)
}

func TestSpinUnixProcessGroupSignals(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "zsplash")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build zsplash binary: %v\n%s", err, string(out))
	}

	t.Run("no-tty spin and progress skip line-clear redraws", func(t *testing.T) {
		spinCmd := exec.Command(binPath, "spin", "Detached spin", "--", "true")
		spinCmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		var spinErr bytes.Buffer
		spinCmd.Stderr = &spinErr
		if err := spinCmd.Run(); err != nil {
			t.Fatalf("expected detached spin to succeed, got: %v", err)
		}
		if strings.Contains(spinErr.String(), "\r\033[2K") {
			t.Fatalf("expected detached spin not to emit \\r\\033[2K redraws, got: %q", spinErr.String())
		}

		progCmd := exec.Command(binPath, "progress", "--total", "2", "--title", "Detached progress")
		progCmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		progCmd.Stdin = strings.NewReader("line1\nline2\n")
		var progErr bytes.Buffer
		progCmd.Stderr = &progErr
		if err := progCmd.Run(); err != nil {
			t.Fatalf("expected detached progress to succeed, got: %v", err)
		}
		if strings.Contains(progErr.String(), "\r\033[2K") {
			t.Fatalf("expected detached progress not to emit \\r\\033[2K redraws, got: %q", progErr.String())
		}
	})

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
		time.Sleep(150 * time.Millisecond)

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
		case <-time.After(12 * time.Second):
			_ = cmd.Process.Kill()
			t.Fatal("timed out waiting for zsp spin to exit after SIGTERM")
		}

		if err := waitForProcessGroupGone(pgid); err != syscall.ESRCH {
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
		time.Sleep(150 * time.Millisecond)

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
		case <-time.After(12 * time.Second):
			_ = cmd.Process.Kill()
			t.Fatal("timed out waiting for zsp spin to exit after SIGINT")
		}

		if err := waitForProcessGroupGone(pgid); err != syscall.ESRCH {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
			t.Fatalf("expected process group %d to be gone (ESRCH), got err=%v", pgid, err)
		}
	})
}
