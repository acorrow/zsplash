package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	var buf bytes.Buffer
	run([]string{"--version"}, &buf)
	out := buf.String()
	if !strings.Contains(out, "zsplash (zsp) v") {
		t.Fatalf("expected version output, got: %q", out)
	}
}

func TestBoxOutputAndReactiveWrapping(t *testing.T) {
	t.Setenv("ZSPLASH_WIDTH", "36")
	var buf bytes.Buffer
	run([]string{"box", "This is a long sentence that should automatically wrap inside a narrow terminal window", "--title", "Status", "--border", "double"}, &buf)
	out := stripANSI(buf.String())
	if !strings.Contains(out, "╔") || !strings.Contains(out, "╚") || !strings.Contains(out, "Status") {
		t.Fatalf("expected double-bordered box with title, got: %q", out)
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if visibleLen(line) > 36 {
			t.Errorf("expected box line to respect ZSPLASH_WIDTH=36, got len %d: %q", visibleLen(line), line)
		}
	}
}

func TestBadgesAndMeta(t *testing.T) {
	cases := []struct {
		cmd   string
		label string
	}{
		{"ok", "PASS"},
		{"warn", "WARN"},
		{"err", "FAIL"},
		{"info", "INFO"},
	}

	for _, tc := range cases {
		var buf bytes.Buffer
		run([]string{tc.cmd, "sample message", "--meta", "42ms"}, &buf)
		out := stripANSI(buf.String())
		if !strings.Contains(out, tc.label) || !strings.Contains(out, "sample message") || !strings.Contains(out, "42ms") {
			t.Errorf("command %q: expected %q, message, and meta in output, got %q", tc.cmd, tc.label, out)
		}
	}
}

func TestBannerFonts(t *testing.T) {
	for _, font := range []string{"block", "compact", "line"} {
		var buf bytes.Buffer
		run([]string{"banner", "ZSP", "--font", font, "--gradient", "cyber", "--frame"}, &buf)
		out := stripANSI(buf.String())
		if len(strings.Split(strings.TrimSpace(out), "\n")) < 2 {
			t.Errorf("font %q: expected multi-line banner output, got %q", font, out)
		}
	}
}

func TestStepAndRule(t *testing.T) {
	var buf bytes.Buffer
	run([]string{"step", "2", "5", "Deploying containers"}, &buf)
	out := stripANSI(buf.String())
	if !strings.Contains(out, "[2/5]") || !strings.Contains(out, "Deploying containers") {
		t.Errorf("expected step header, got %q", out)
	}

	buf.Reset()
	run([]string{"rule", "SECTION"}, &buf)
	if !strings.Contains(stripANSI(buf.String()), "SECTION") {
		t.Errorf("expected rule with label, got %q", buf.String())
	}
}

func TestKVAndCmdAndTree(t *testing.T) {
	var buf bytes.Buffer
	run([]string{"kv", "--title", "Summary", "Env", "prod", "Host", "10.0.0.1"}, &buf)
	out := stripANSI(buf.String())
	if !strings.Contains(out, "Summary") || !strings.Contains(out, "Env") || !strings.Contains(out, "prod") {
		t.Errorf("expected kv card output, got %q", out)
	}

	buf.Reset()
	run([]string{"cmd", "docker build --tag app:latest ."}, &buf)
	if !strings.Contains(stripANSI(buf.String()), "docker build") {
		t.Errorf("expected command preview output, got %q", buf.String())
	}

	buf.Reset()
	run([]string{"tree", "--title", "Checks", "Lint", "Unit Tests"}, &buf)
	treeOut := stripANSI(buf.String())
	if !strings.Contains(treeOut, "Checks") || !strings.Contains(treeOut, "╰─") {
		t.Errorf("expected tree output, got %q", treeOut)
	}
}

func TestBarAndSpin(t *testing.T) {
	var buf bytes.Buffer
	run([]string{"bar", "50", "100", "Uploading"}, &buf)
	out := stripANSI(buf.String())
	if !strings.Contains(out, "50%") || !strings.Contains(out, "[50/100]") || !strings.Contains(out, "Uploading") {
		t.Errorf("expected progress bar output, got %q", out)
	}

	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")
	var spinOut, spinErr bytes.Buffer
	code := runWithExit([]string{"spin", "Testing spinner", "--duration", "40", "--persist"}, &spinOut, &spinErr)
	if code != 0 {
		t.Errorf("expected exit code 0 from timed spinner, got %d", code)
	}
}

func TestDemo(t *testing.T) {
	var buf bytes.Buffer
	run([]string{"demo"}, &buf)
	if !strings.Contains(stripANSI(buf.String()), "Release Target") {
		t.Errorf("expected demo output to render sections, got %q", buf.String())
	}
}
