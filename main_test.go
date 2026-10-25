package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
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
	code := runWithExit([]string{"spin", "Testing spinner", "--duration", "40"}, &spinOut, &spinErr)
	if code != 0 {
		t.Errorf("expected exit code 0 from timed spinner, got %d", code)
	}
	if strings.Contains(spinErr.String(), "\033[2K") || strings.Contains(spinErr.String(), "\033[?25l") {
		t.Errorf("expected non-TTY spinner not to emit line-clear or cursor-hide escapes, got %q", spinErr.String())
	}
	if !strings.Contains(stripANSI(spinErr.String()), "▸ Testing spinner") || !strings.Contains(stripANSI(spinErr.String()), "[PASS]") {
		t.Errorf("expected non-TTY spinner to print static title and final PASS badge, got %q", stripANSI(spinErr.String()))
	}

	// Verify progress with non-TTY output skips per-line redraws and only prints the final badge
	var progErr bytes.Buffer
	progCode := runStreamProgress(strings.NewReader("a\nb\nc\n"), &progErr, 3, "Batch sync", "", "")
	if progCode != 0 {
		t.Errorf("expected exit code 0 from progress, got %d", progCode)
	}
	if strings.Contains(progErr.String(), "\033[2K") {
		t.Errorf("expected non-TTY progress not to emit line-clear redraws, got %q", progErr.String())
	}
	if !strings.Contains(stripANSI(progErr.String()), "[PASS]") || !strings.Contains(stripANSI(progErr.String()), "3 items") {
		t.Errorf("expected non-TTY progress to print final PASS badge with item count, got %q", stripANSI(progErr.String()))
	}

	// Verify --static prints title once without redraw escapes and emits final PASS badge
	spinOut.Reset()
	spinErr.Reset()
	code = runWithExit([]string{"spin", "Static prompt step", "--static", "--duration", "20"}, &spinOut, &spinErr)
	if code != 0 {
		t.Errorf("expected exit code 0 from static spinner, got %d", code)
	}
	staticRendered := spinErr.String()
	if strings.Contains(staticRendered, "\033[2K") {
		t.Errorf("expected --static spinner not to emit line-clear redraws, got %q", staticRendered)
	}
	if !strings.Contains(stripANSI(staticRendered), "▸ Static prompt step") || !strings.Contains(stripANSI(staticRendered), "[PASS]") {
		t.Errorf("expected --static spinner to print title and final PASS badge, got %q", stripANSI(staticRendered))
	}

	// Verify shouldIsolateProcessGroup decision helper
	if shouldIsolateProcessGroup(true) {
		t.Error("expected shouldIsolateProcessGroup(true) to be false when a terminal is attached")
	}
	if !shouldIsolateProcessGroup(false) {
		t.Error("expected shouldIsolateProcessGroup(false) to be true when no terminal is attached")
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
		"## 6. Using `zsp` in Scripts",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("expected `zsp docs` markdown to contain %q", want)
		}
	}

	buf.Reset()
	run([]string{"help"}, &buf)
	if !strings.Contains(stripANSI(buf.String()), "List all 20 built-in color palettes") {
		t.Errorf("expected `zsp help` to list 20 built-in color palettes, got:\n%s", stripANSI(buf.String()))
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
	buf.Reset()
	run([]string{"help", "spin"}, &buf)
	spinHelp := stripANSI(buf.String())
	if !strings.Contains(spinHelp, "--static") || !strings.Contains(spinHelp, "process group") {
		t.Errorf("expected 'zsp help spin' to document --static and process group behavior, got:\n%s", spinHelp)
	}
}

func TestSecurityHardening(t *testing.T) {
	// 1. Escape sequence, C0/DEL/C1, and Unicode bidi/format character sanitization (keeping U+200D ZWJ)
	hostile := "safe\x1b]52;c;ZWNobyBoYWNrZWQ=\x07\x1b[2J\r\x00\x7f\u009b\u202e\u2066\u200b\ufefftext"
	var buf bytes.Buffer
	run([]string{"ok", hostile, "--meta", "\x1b]0;pwn\x0742ms"}, &buf)
	out := buf.String()
	for _, forbidden := range []string{"\x1b]52;", "\x1b[2J", "\x1b]0;", "\x00", "\x7f", "\u202e", "\u2066", "\u200b", "\ufeff"} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("expected hostile sequence %q to be stripped from output, got: %q", forbidden, out)
		}
	}
	if !strings.Contains(stripANSI(out), "safetext") {
		t.Fatalf("expected sanitized text 'safetext' in output, got: %q", stripANSI(out))
	}
	if got := sanitizeText("a\u200db", false); got != "a\u200db" {
		t.Fatalf("expected ZWJ U+200D to be preserved, got: %q", got)
	}

	// Verify truncateVisible strips escapes even when string is shorter than maxRunes
	shortHostile := "hi\x1b]52;c;abc\x07\u202e"
	if got := truncateVisible(shortHostile, 50); got != "hi" {
		t.Fatalf("expected truncateVisible to strip escapes on short string, got: %q", got)
	}

	// 2. --raw cannot be smuggled after the subcommand, and visibleLen ignores ANSI in --raw mode
	buf.Reset()
	run([]string{"kv", "--raw", "val\x1b]52;c;pwn\x07"}, &buf)
	if strings.Contains(buf.String(), "\x1b]52;") {
		t.Fatalf("expected --raw after subcommand NOT to disable sanitization, got: %q", buf.String())
	}
	buf.Reset()
	run([]string{"--raw", "ok", "\x1b[32mcustom\x1b[0m"}, &buf)
	if !strings.Contains(buf.String(), "\x1b[32mcustom\x1b[0m") {
		t.Fatalf("expected leading --raw before subcommand to preserve ANSI, got: %q", buf.String())
	}
	if got := visibleLen("\x1b[32mhello\x1b[0m"); got != 5 {
		t.Fatalf("expected visibleLen to return 5 for ANSI-styled 'hello', got %d", got)
	}

	// 3. ZSPLASH_CONFIG refuses non-config files, empty files, 'null', and trailing garbage
	for name, content := range map[string]string{
		".zshrc":   "export PATH=/usr/bin:$PATH\n",
		"empty":    "   \n",
		"nulljson": "null\n",
		"trailing": "{}\nnot-json\n",
	} {
		target := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(target, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("ZSPLASH_CONFIG", target)
		buf.Reset()
		if code := runWithExit([]string{"config", "reset"}, &buf, &buf); code == 0 {
			t.Fatalf("case %q: expected non-zero exit when resetting invalid config file", name)
		}
		if _, err := os.Stat(target); err != nil {
			t.Fatalf("case %q: expected file %q to NOT be deleted: %v", name, target, err)
		}
		buf.Reset()
		if code := runWithExit([]string{"config", "set", "theme", "nes"}, &buf, &buf); code == 0 {
			t.Fatalf("case %q: expected non-zero exit when overwriting invalid config file", name)
		}
	}

	// 4. Negative hex colors are rejected cleanly
	fallback := RGB{10, 20, 30}
	if got := parseLiteralColor("#-1-1-1", fallback); got != fallback {
		t.Fatalf("expected negative hex color #-1-1-1 to be rejected, got: %+v", got)
	}

	// 5. Signal exit codes follow 128 + signal convention
	if got := signalExitCode(os.Interrupt); got != 130 {
		t.Fatalf("expected SIGINT exit code 130, got %d", got)
	}
	if got := signalExitCode(syscall.SIGTERM); got != 143 {
		t.Fatalf("expected SIGTERM exit code 143, got %d", got)
	}

	// 6. Extreme widths are safely clamped
	t.Setenv("ZSPLASH_WIDTH", "2000000000")
	buf.Reset()
	run([]string{"bar", "5", "10", "Test", "--width", "2000000000"}, &buf)
	if len(buf.String()) > 2000 {
		t.Fatalf("expected width to be clamped, got output length %d", len(buf.String()))
	}

	// 7. Unknown theme names (--theme or ZSPLASH_THEME) fall back to "modern" and never emit escapes in help
	t.Setenv("ZSPLASH_WIDTH", "80")
	buf.Reset()
	run([]string{"--theme", "\x1b]0;x\x07", "help"}, &buf)
	if strings.Contains(buf.String(), "\x1b]0;") || !strings.Contains(stripANSI(buf.String()), "Active theme: modern") {
		t.Fatalf("expected unknown --theme with escape sequence to fall back to 'modern', got: %q", buf.String())
	}

	// 8. "--" end-of-flags is honored across commands and as top-level box command
	buf.Reset()
	run([]string{"ok", "--", "--meta", "x"}, &buf)
	okPlain := stripANSI(buf.String())
	if !strings.Contains(okPlain, "--meta x") || strings.Contains(okPlain, "····") {
		t.Fatalf("expected 'zsp ok -- --meta x' to print '--meta x' as message with no meta, got: %q", okPlain)
	}
	buf.Reset()
	run([]string{"ok", "--", "--raw", "hi\x1b]52;c;pwn\x07"}, &buf)
	if strings.Contains(buf.String(), "\x1b]52;") {
		t.Fatalf("expected 'zsp ok -- --raw <escape>' to still strip escape, got: %q", buf.String())
	}
	buf.Reset()
	run([]string{"tree", "--", "--title", "a"}, &buf)
	treePlain := stripANSI(buf.String())
	if !strings.Contains(treePlain, "├─ ● --title") || !strings.Contains(treePlain, "╰─ ● a") {
		t.Fatalf("expected 'zsp tree -- --title a' to treat '--title' as an item, got: %q", treePlain)
	}
	buf.Reset()
	run([]string{"--", "--title", "literal box"}, &buf)
	if !strings.Contains(stripANSI(buf.String()), "--title literal box") {
		t.Fatalf("expected 'zsp -- --title literal box' to render box text, got: %q", stripANSI(buf.String()))
	}
}

func TestSpinPipedInputAndConfigSymlinks(t *testing.T) {
	// 1. Piped input automatically reaches spin child command
	binName := "zsplash"
	if runtime.GOOS == "windows" {
		binName = "zsplash.exe"
	}
	binPath := filepath.Join(t.TempDir(), binName)
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build zsplash binary: %v\n%s", err, string(out))
	}

	catCmd := []string{"cat"}
	if _, err := exec.LookPath("cat"); err != nil {
		catCmd = []string{"findstr", ".*"}
	}
	spinArgs := append([]string{"spin", "Reading stdin", "--"}, catCmd...)
	cmd := exec.Command(binPath, spinArgs...)
	cmd.Stdin = strings.NewReader("piped-line-value\n")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected spin with piped stdin to succeed, got err: %v", err)
	}
	if strings.TrimSpace(string(out)) != "piped-line-value" {
		t.Fatalf("expected piped input 'piped-line-value' on stdout, got: %q", string(out))
	}

	// 2. Symlink to a valid config can be read and updated while keeping the symlink intact;
	// symlink to a non-config file is refused by both config set and config reset.
	t.Run("config symlinks", func(t *testing.T) {
		tmpDir := t.TempDir()
		realCfg := filepath.Join(tmpDir, "real-config.json")
		if err := os.WriteFile(realCfg, []byte(`{"theme":"dracula"}`), 0600); err != nil {
			t.Fatal(err)
		}
		linkCfg := filepath.Join(tmpDir, "link-config.json")
		if err := os.Symlink(realCfg, linkCfg); err != nil {
			if runtime.GOOS == "windows" {
				t.Skipf("skipping symlink test on Windows without symlink privilege: %v", err)
			}
			t.Fatal(err)
		}

		t.Setenv("ZSPLASH_CONFIG", linkCfg)
		var buf bytes.Buffer
		run([]string{"config", "list"}, &buf)
		if !strings.Contains(stripANSI(buf.String()), "dracula") {
			t.Fatalf("expected symlinked config theme 'dracula' to be read, got: %q", stripANSI(buf.String()))
		}

		buf.Reset()
		if code := runWithExit([]string{"config", "set", "theme", "nes"}, &buf, &buf); code != 0 {
			t.Fatalf("expected config set through symlink to succeed, got code %d: %s", code, buf.String())
		}
		linkInfo, err := os.Lstat(linkCfg)
		if err != nil || linkInfo.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("expected %q to remain a symlink after update, got mode=%v err=%v", linkCfg, linkInfo.Mode(), err)
		}
		updatedBytes, err := os.ReadFile(realCfg)
		if err != nil || !strings.Contains(string(updatedBytes), `"nes"`) {
			t.Fatalf("expected target config %q to be updated to 'nes', got: %q (err=%v)", realCfg, string(updatedBytes), err)
		}

		// Resetting through a symlink removes the symlink itself, leaving the target file in place
		buf.Reset()
		if code := runWithExit([]string{"config", "reset"}, &buf, &buf); code != 0 {
			t.Fatalf("expected config reset through valid symlink to succeed, got code %d", code)
		}
		if _, err := os.Lstat(linkCfg); !os.IsNotExist(err) {
			t.Fatalf("expected symlink %q to be removed on config reset, got err=%v", linkCfg, err)
		}
		if _, err := os.Stat(realCfg); err != nil {
			t.Fatalf("expected real target file %q to remain intact after symlink reset: %v", realCfg, err)
		}

		// Symlink to a non-config file (fake .zshrc) is refused
		fakeZshrc := filepath.Join(tmpDir, ".zshrc")
		origZshrc := "export EDITOR=vim\n"
		if err := os.WriteFile(fakeZshrc, []byte(origZshrc), 0600); err != nil {
			t.Fatal(err)
		}
		badLink := filepath.Join(tmpDir, "bad-link.json")
		if err := os.Symlink(fakeZshrc, badLink); err != nil {
			t.Fatal(err)
		}
		t.Setenv("ZSPLASH_CONFIG", badLink)
		buf.Reset()
		if code := runWithExit([]string{"config", "set", "theme", "nes"}, &buf, &buf); code == 0 {
			t.Fatalf("expected config set through symlink to non-config file to fail")
		}
		buf.Reset()
		if code := runWithExit([]string{"config", "reset"}, &buf, &buf); code == 0 {
			t.Fatalf("expected config reset through symlink to non-config file to fail")
		}
		afterZshrc, err := os.ReadFile(fakeZshrc)
		if err != nil || string(afterZshrc) != origZshrc {
			t.Fatalf("expected target non-config file %q to be unchanged, got %q (err=%v)", fakeZshrc, string(afterZshrc), err)
		}
	})
}
