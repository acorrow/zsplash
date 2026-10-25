package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
	out := stripANSI(buf.String())
	for _, expected := range []string{
		"Release Target",
		"zsh: banners & step headers",
		"zsh: badges",
		"zsh: structured layout & dividers",
		"zsh: async & progress",
		"zsh: themes, defaults & docs",
		`zsp banner "ZSPLASH"`,
		`zsp step 1 5`,
		`zsp rule "SUMMARY" --color secondary`,
		`zsp ok   "Database schema verified"   --meta "42ms"`,
		`zsp kv --title "Release Target"`,
		`zsp bar 78 100`,
		`zsp theme nes`,
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected demo output to contain %q, got:\n%s", expected, out)
		}
	}

	// Verify every framed block in demo has a flush right border
	inCodeBlock := false
	expectedWidth := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "  ╭─ zsh:") {
			inCodeBlock = true
			expectedWidth = visibleLen(line)
			continue
		}
		if inCodeBlock {
			if visibleLen(line) != expectedWidth {
				t.Errorf("code block border misalignment: expected width %d, got %d on line %q", expectedWidth, visibleLen(line), line)
			}
			if strings.HasPrefix(line, "  ╰") {
				inCodeBlock = false
			}
		}
	}
}

func TestSolidDefaultAndNESTheme(t *testing.T) {
	t.Setenv("ZSPLASH_CONFIG", filepath.Join(t.TempDir(), "config.json"))

	// Default banner in modern theme uses solid #00F5D4 (0;245;212), not a multi-stop gradient
	var buf bytes.Buffer
	run([]string{"banner", "NES"}, &buf)
	rawModern := buf.String()
	if !strings.Contains(rawModern, "38;2;0;245;212m") {
		t.Fatalf("expected solid modern primary color (0;245;212) by default, got: %q", rawModern)
	}

	// Passing --theme nes switches to the NES console palette (#E60012 -> 230;0;18)
	buf.Reset()
	run([]string{"--theme", "nes", "banner", "NES"}, &buf)
	rawNES := buf.String()
	if !strings.Contains(rawNES, "38;2;230;0;18m") {
		t.Fatalf("expected solid NES primary color (230;0;18) with --theme nes, got: %q", rawNES)
	}

	// Passing --theme xbox switches to the Xbox jewel green palette (#18E618 -> 24;230;24)
	buf.Reset()
	run([]string{"--theme", "xbox", "banner", "XBOX"}, &buf)
	if !strings.Contains(buf.String(), "38;2;24;230;24m") {
		t.Fatalf("expected solid Xbox primary color (24;230;24) with --theme xbox, got: %q", buf.String())
	}

	// Listing themes includes built-in palettes
	buf.Reset()
	run([]string{"themes"}, &buf)
	themesOut := stripANSI(buf.String())
	for _, th := range []string{"modern", "nes", "snes", "xbox", "halo", "gameboy", "pipboy", "dracula", "catppuccin", "nord", "gruvbox", "tokyonight", "monokai", "solarized", "rosepine", "everforest", "kanagawa", "synthwave", "industrial", "amber"} {
		if !strings.Contains(themesOut, th) {
			t.Errorf("expected themes list to include %q, got:\n%s", th, themesOut)
		}
	}
}

func TestConfigPersistenceAndOverrides(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("ZSPLASH_CONFIG", cfgPath)

	var buf bytes.Buffer
	// Set default theme to nes
	run([]string{"theme", "nes"}, &buf)
	if !strings.Contains(stripANSI(buf.String()), `Default theme set to "nes"`) {
		t.Fatalf("expected confirmation when setting theme to nes, got: %q", stripANSI(buf.String()))
	}

	// Verify banner now defaults to NES red without passing --theme
	buf.Reset()
	run([]string{"banner", "P1"}, &buf)
	if !strings.Contains(buf.String(), "38;2;230;0;18m") {
		t.Fatalf("expected persisted nes theme color (230;0;18), got: %q", buf.String())
	}

	// Override box.border to double and color.primary to #112233 (17;34;51)
	buf.Reset()
	run([]string{"config", "set", "box.border", "double"}, &buf)
	buf.Reset()
	run([]string{"config", "set", "color.primary", "#112233"}, &buf)

	buf.Reset()
	run([]string{"box", "Custom Box"}, &buf)
	boxRaw := buf.String()
	boxPlain := stripANSI(boxRaw)
	if !strings.Contains(boxPlain, "╔") || !strings.Contains(boxRaw, "38;2;17;34;51m") {
		t.Fatalf("expected double border and custom primary color (17;34;51), got: %q", boxRaw)
	}

	// Reset config and verify it returns to modern defaults
	buf.Reset()
	run([]string{"config", "reset"}, &buf)
	buf.Reset()
	run([]string{"banner", "P1"}, &buf)
	if !strings.Contains(buf.String(), "38;2;0;245;212m") {
		t.Fatalf("expected modern primary color (0;245;212) after config reset, got: %q", buf.String())
	}
}

func TestDocsAndSchemaAndSubcommandHelp(t *testing.T) {
	var buf bytes.Buffer
	run([]string{"docs"}, &buf)
	md := buf.String()
	for _, want := range []string{
		"# zsplash (`zsp`) v",
		"## 1. Core Execution & Stream Contracts",
		"## 3. Color System, Built-in Palettes & Gradients",
		"## 5. Subcommand Reference",
		"### `zsp spin`",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("expected `zsp docs` markdown to contain %q", want)
		}
	}

	buf.Reset()
	run([]string{"docs", "--json"}, &buf)
	var spec CLISpec
	if err := json.Unmarshal(buf.Bytes(), &spec); err != nil {
		t.Fatalf("expected valid JSON from `zsp docs --json`, got error: %v", err)
	}
	if spec.Name != "zsplash" || spec.Alias != "zsp" || len(spec.Commands) < 15 || len(spec.Themes) != 20 {
		t.Fatalf("unexpected CLI JSON spec structure: commands=%d, themes=%d", len(spec.Commands), len(spec.Themes))
	}

	// Verify both `zsp help spin` and `zsp spin --help` render subcommand specification
	for _, args := range [][]string{{"help", "spin"}, {"spin", "--help"}, {"banner", "-h"}} {
		buf.Reset()
		run(args, &buf)
		helpOut := stripANSI(buf.String())
		if !strings.Contains(helpOut, "SYNOPSIS") || !strings.Contains(helpOut, "STREAM & EXIT CONTRACT") {
			t.Errorf("args %v: expected SYNOPSIS and STREAM & EXIT CONTRACT in subcommand help, got:\n%s", args, helpOut)
		}
	}
}

func TestSecurityHardening(t *testing.T) {
	// 1. Escape sequence and control character sanitization (OSC 52, CSI, CR, NUL, DEL, C1)
	hostile := "safe\x1b]52;c;ZWNobyBoYWNrZWQ=\x07\x1b[2J\r\x00\x7f\u009btext"
	var buf bytes.Buffer
	run([]string{"ok", hostile, "--meta", "\x1b]0;pwn\x0742ms"}, &buf)
	out := buf.String()
	if strings.Contains(out, "\x1b]52;") || strings.Contains(out, "\x1b[2J") || strings.Contains(out, "\x1b]0;") || strings.Contains(out, "\x00") || strings.Contains(out, "\x7f") {
		t.Fatalf("expected hostile OSC/CSI/C0 sequences to be stripped from output, got: %q", out)
	}
	if !strings.Contains(stripANSI(out), "safetext") {
		t.Fatalf("expected sanitized text 'safetext' in output, got: %q", stripANSI(out))
	}

	// Verify truncateVisible strips escapes even when string is shorter than maxRunes
	shortHostile := "hi\x1b]52;c;abc\x07"
	if got := truncateVisible(shortHostile, 50); got != "hi" {
		t.Fatalf("expected truncateVisible to strip escapes on short string, got: %q", got)
	}

	// 2. ZSPLASH_CONFIG refuses to delete or overwrite a non-config file (e.g. .zshrc)
	fakeZshrc := filepath.Join(t.TempDir(), ".zshrc")
	if err := os.WriteFile(fakeZshrc, []byte("export PATH=/usr/bin:$PATH\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ZSPLASH_CONFIG", fakeZshrc)
	buf.Reset()
	code := runWithExit([]string{"config", "reset"}, &buf, &buf)
	if code == 0 {
		t.Fatalf("expected non-zero exit when resetting a non-config file")
	}
	if _, err := os.Stat(fakeZshrc); err != nil {
		t.Fatalf("expected non-config file %q to NOT be deleted: %v", fakeZshrc, err)
	}
	buf.Reset()
	code = runWithExit([]string{"config", "set", "theme", "nes"}, &buf, &buf)
	if code == 0 {
		t.Fatalf("expected non-zero exit when overwriting a non-config file")
	}

	// 3. Extreme widths are safely clamped
	t.Setenv("ZSPLASH_WIDTH", "2000000000")
	buf.Reset()
	run([]string{"bar", "5", "10", "Test", "--width", "2000000000"}, &buf)
	if len(buf.String()) > 2000 {
		t.Fatalf("expected width to be clamped, got output length %d", len(buf.String()))
	}
}
