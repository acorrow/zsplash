package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type ArgSpec struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Repeatable  bool   `json:"repeatable,omitempty"`
	Description string `json:"description"`
}

type FlagSpec struct {
	Short       string   `json:"short,omitempty"`
	Long        string   `json:"long"`
	ValueName   string   `json:"value_name,omitempty"`
	Allowed     []string `json:"allowed,omitempty"`
	Default     string   `json:"default,omitempty"`
	ConfigKey   string   `json:"config_key,omitempty"`
	Description string   `json:"description"`
}

type CommandSpec struct {
	Name        string     `json:"name"`
	Aliases     []string   `json:"aliases,omitempty"`
	Category    string     `json:"category"`
	Summary     string     `json:"summary"`
	Synopsis    string     `json:"synopsis"`
	Stream      string     `json:"stream"`
	ExitCode    string     `json:"exit_code"`
	Positionals []ArgSpec  `json:"positionals,omitempty"`
	Flags       []FlagSpec `json:"flags,omitempty"`
	Rules       []string   `json:"rules,omitempty"`
	Examples    []string   `json:"examples"`
}

type EnvVarSpec struct {
	Name        string `json:"name"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
}

type ConfigKeySpec struct {
	Key         string   `json:"key"`
	Allowed     []string `json:"allowed,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description"`
}

type ThemeHexSpec struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Roles       map[string]string `json:"roles"`
}

type RecipeSpec struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Code        string `json:"code"`
}

type CLISpec struct {
	Name          string            `json:"name"`
	Alias         string            `json:"alias"`
	Version       string            `json:"version"`
	Summary       string            `json:"summary"`
	GlobalFlags   []FlagSpec        `json:"global_flags"`
	EnvVars       []EnvVarSpec      `json:"environment_variables"`
	SemanticRoles []string          `json:"semantic_color_roles"`
	Themes        []ThemeHexSpec    `json:"themes"`
	NamedColors   map[string]string `json:"named_colors"`
	Gradients     []string          `json:"gradients"`
	ConfigKeys    []ConfigKeySpec   `json:"config_keys"`
	Commands      []CommandSpec     `json:"commands"`
	Recipes       []RecipeSpec      `json:"zsh_recipes"`
}

func hasHelpFlag(args []string) bool {
	for _, a := range args {
		if a == "--" {
			return false
		}
		if a == "--help" || a == "-h" {
			return true
		}
	}
	return false
}

func flg(short, long, val string, allowed []string, def, cfgKey, desc string) FlagSpec {
	return FlagSpec{Short: short, Long: long, ValueName: val, Allowed: allowed, Default: def, ConfigKey: cfgKey, Description: desc}
}

func buildCLISpec() CLISpec {
	var themes []ThemeHexSpec
	for _, name := range themeOrder {
		t := builtinThemes[name]
		themes = append(themes, ThemeHexSpec{
			Name: t.Name, Description: t.Description,
			Roles: map[string]string{
				"primary": t.Primary.Hex(), "secondary": t.Secondary.Hex(), "pass": t.Pass.Hex(), "warn": t.Warn.Hex(),
				"fail": t.Fail.Hex(), "info": t.Info.Hex(), "muted": t.Muted.Hex(), "text": t.Text.Hex(),
			},
		})
	}
	namedHex := make(map[string]string, len(namedColors))
	for k, v := range namedColors {
		namedHex[k] = v.Hex()
	}
	grads := []string{"cyber", "ember", "ocean", "sunset", "matrix", "gold", "mono"}
	colorSpecVals := []string{"<role>", "<named-color>", "#RRGGBB"}

	return CLISpec{
		Name: "zsplash", Alias: "zsp", Version: version,
		Summary: "Fast, reactive CLI utility for Zsh and shell scripts to render colorful banners, boxes, status badges, phase headers, key-value cards, command previews, spinners, and progress bars.",
		GlobalFlags: []FlagSpec{
			flg("-T", "--theme", "<theme>", themeOrder, "modern", "theme", "Temporarily apply a built-in color palette for this command."),
			flg("", "--raw", "", nil, "false", "", "Must precede the subcommand (`zsp --raw <cmd> ...`): allow raw escape sequences in user text."),
			flg("-h", "--help", "", nil, "", "", "Show overview help or subcommand spec (`zsp <cmd> --help`)."),
			flg("-v", "--version", "", nil, "", "", "Print version string and exit."),
		},
		EnvVars: []EnvVarSpec{
			{"ZSPLASH_THEME", "", "Override active color palette for the current shell session."},
			{"ZSPLASH_CONFIG", "~/.config/zsplash/config.json", "Custom path for reading and writing persistent user configuration."},
			{"ZSPLASH_WIDTH", "auto (/dev/tty -> $COLUMNS -> 80)", "Force a specific terminal column width (integer 20..500)."},
			{"NO_COLOR", "", "Disable all ANSI color escapes and use bracket badges (`[PASS]`, `[WARN]`, `[FAIL]`, `[INFO]`)."},
		},
		SemanticRoles: []string{"primary", "secondary", "pass", "warn", "fail", "info", "muted", "text"},
		Themes:        themes, NamedColors: namedHex, Gradients: grads,
		ConfigKeys: []ConfigKeySpec{
			{"theme", themeOrder, "modern", "Default color palette across all commands."},
			{"banner.font", []string{"block", "compact", "line"}, "block", "Default font for `zsp banner`."},
			{"banner.color", colorSpecVals, "primary", "Default solid color for `zsp banner`."},
			{"box.border", []string{"rounded", "double", "heavy", "ascii"}, "rounded", "Default border style for `zsp box` and `zsp kv`."},
			{"box.color", colorSpecVals, "primary", "Default border color for `zsp box`."},
			{"step.color", colorSpecVals, "primary", "Default rule color for `zsp step`."},
			{"rule.color", colorSpecVals, "primary", "Default line color for `zsp rule`."},
			{"kv.color", colorSpecVals, "secondary", "Default key and border color for `zsp kv`."},
			{"bar.color", colorSpecVals, "primary", "Default filled bar color for `zsp bar` and `zsp progress`."},
			{"spin.style", []string{"dots", "pulse", "meter", "line"}, "dots", "Default animation style for `zsp spin`."},
			{"spin.color", colorSpecVals, "primary", "Default spinner color for `zsp spin`."},
			{"color.<role>", []string{"<named-color>", "#RRGGBB"}, "(theme)", "Override `color.primary`, `color.secondary`, `color.pass`, `color.warn`, `color.fail`, `color.info`, `color.muted`, or `color.text`."},
		},
		Commands: buildCommandSpecs(grads),
		Recipes: []RecipeSpec{
			{
				Title:       "Production Deployment Script with Spinner Output Capture",
				Description: "`zsp spin` animates on `/dev/tty`, keeping `stdout` clean for `$(...)` capture while propagating the child exit code (`$?`).",
				Code:        "#!/usr/bin/env zsh\nset -uo pipefail\nexport ZSPLASH_THEME=\"nes\"\n\nzsp banner \"DEPLOY\" --font compact --frame --sub \"Release Pipeline\"\nzsp step 1 2 \"Verifying git revision\"\nSHA=$(zsp spin \"Resolving HEAD...\" --persist -- git rev-parse --short HEAD)\nif [[ $? -ne 0 ]]; then zsp err \"Git check failed\" --meta \"exit 1\"; exit 1; fi\n\nzsp step 2 2 \"Summary\"\nzsp kv --title \"Release\" \"Commit\" \"$SHA\" \"Status\" \"Healthy\"",
			},
		},
	}
}

func buildCommandSpecs(grads []string) []CommandSpec {
	return []CommandSpec{
		{
			Name: "banner", Aliases: []string{"big"}, Category: "Typography & Framing",
			Summary:  "Render high-contrast Unicode block typography (A-Z, 0-9, punctuation) with optional frame and subtitle.",
			Synopsis: "zsp banner <text> [--font block|compact|line] [--color <c>] [--gradient <g>] [--from <#hex> --to <#hex>] [--sub <text>] [--frame]",
			Stream:   "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "text", Required: true, Description: "Headline text (automatically uppercased)."}},
			Flags: []FlagSpec{
				flg("-f", "--font", "<font>", []string{"block", "compact", "line"}, "block", "banner.font", "Glyph font (`block` 3-row, `compact` 2-row, `line` 3-row box-drawing)."),
				flg("-c", "--color", "<color>", nil, "primary", "banner.color", "Solid color (`<role>`, `<named-color>`, or `#RRGGBB`)."),
				flg("-g", "--gradient", "<grad>", grads, "", "", "Opt-in 24-bit horizontal gradient."),
				flg("", "--from", "<#hex>", nil, "", "", "Custom gradient start hex (with `--to`)."),
				flg("", "--to", "<#hex>", nil, "", "", "Custom gradient end hex (with `--from`)."),
				flg("", "--sub", "<text>", nil, "", "", "Subtitle rendered below the banner (or inside `--frame`)."),
				flg("", "--frame", "", nil, "false", "", "Enclose banner and subtitle in a rounded box."),
			},
			Rules:    []string{"Solid color is used by default; gradients are only applied when `--gradient` or `--from`/`--to` are passed."},
			Examples: []string{`zsp banner "DEPLOY" --font block --frame --sub "Production Pipeline"`, `zsp banner "8BIT" --theme nes --color nes-red`},
		},
		{
			Name: "box", Category: "Typography & Framing",
			Summary:  "Render a reactive Unicode callout box that word-wraps to fit the live terminal width.",
			Synopsis: "zsp box <text> [--title <t>] [--footer <f>] [--border rounded|double|heavy|ascii] [--color <c>] [--gradient <g>] [--width <cols>]",
			Stream:   "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "text", Required: true, Description: "Body text (wraps automatically to terminal width)."}},
			Flags: []FlagSpec{
				flg("-t", "--title", "<title>", nil, "", "", "Header embedded in the top border."),
				flg("", "--footer", "<footer>", nil, "", "", "Footer embedded in the bottom border."),
				flg("-b", "--border", "<style>", []string{"rounded", "double", "heavy", "ascii"}, "rounded", "box.border", "Border glyph set."),
				flg("-c", "--color", "<color>", nil, "primary", "box.color", "Solid border color."),
				flg("-g", "--gradient", "<grad>", grads, "", "", "Opt-in border gradient."),
				flg("-w", "--width", "<cols>", nil, "auto", "", "Target outer width in columns."),
			},
			Examples: []string{`zsp box "All 14 migrations applied." --title "Database" --border double`},
		},
		{
			Name: "ok", Aliases: []string{"pass", "success"}, Category: "Status & Phase Indicators",
			Summary:  "Render an inverted `PASS` pill badge with message and optional right-aligned dot-leader `--meta`.",
			Synopsis: "zsp ok <message> [--meta <detail>]", Stream: "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "message", Required: true, Description: "Status message."}},
			Flags:       []FlagSpec{flg("-m", "--meta", "<detail>", nil, "", "", "Right-aligned metadata connected via dot-leaders (`····`).")},
			Examples:    []string{`zsp ok "Database schema verified" --meta "42ms"`},
		},
		{
			Name: "info", Category: "Status & Phase Indicators",
			Summary:  "Render an inverted `INFO` pill badge with message and optional `--meta`.",
			Synopsis: "zsp info <message> [--meta <detail>]", Stream: "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "message", Required: true, Description: "Informational message."}},
			Flags:       []FlagSpec{flg("-m", "--meta", "<detail>", nil, "", "", "Right-aligned metadata.")},
			Examples:    []string{`zsp info "Worker pool initialized" --meta "8 threads"`},
		},
		{
			Name: "warn", Aliases: []string{"warning"}, Category: "Status & Phase Indicators",
			Summary:  "Render an inverted `WARN` pill badge with message and optional `--meta`.",
			Synopsis: "zsp warn <message> [--meta <detail>]", Stream: "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "message", Required: true, Description: "Warning message."}},
			Flags:       []FlagSpec{flg("-m", "--meta", "<detail>", nil, "", "", "Right-aligned metadata.")},
			Examples:    []string{`zsp warn "Rate limit at 85% capacity" --meta "throttling"`},
		},
		{
			Name: "err", Aliases: []string{"fail", "error"}, Category: "Status & Phase Indicators",
			Summary:  "Render an inverted `FAIL` pill badge with message and optional `--meta`.",
			Synopsis: "zsp err <message> [--meta <detail>]", Stream: "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "message", Required: true, Description: "Error message."}},
			Flags:       []FlagSpec{flg("-m", "--meta", "<detail>", nil, "", "", "Right-aligned metadata.")},
			Examples:    []string{`zsp err "Upstream webhook timeout" --meta "exit 1"`},
		},
		{
			Name: "step", Category: "Status & Phase Indicators",
			Summary:  "Render a full-width phase divider (`━━━ [curr/total] ━━━━━━`) and bold `▸ <title>` line.",
			Synopsis: "zsp step <current> <total> <title> [--color <c>] [--gradient <g>]", Stream: "stdout", ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "current", Required: true, Description: "Current step number."},
				{Name: "total", Required: true, Description: "Total steps."},
				{Name: "title", Required: true, Description: "Phase description."},
			},
			Flags:    []FlagSpec{flg("-c", "--color", "<color>", nil, "primary", "step.color", "Solid rule color."), flg("-g", "--gradient", "<grad>", grads, "", "", "Opt-in rule gradient.")},
			Examples: []string{`zsp step 1 4 "Validating credentials"`},
		},
		{
			Name: "rule", Aliases: []string{"hr"}, Category: "Status & Phase Indicators",
			Summary:  "Render a horizontal divider (`───`) with an optional centered label.",
			Synopsis: "zsp rule [label] [--color <c>] [--gradient <g>] [--width <cols>]", Stream: "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "label", Required: false, Description: "Optional centered label."}},
			Flags:       []FlagSpec{flg("-c", "--color", "<color>", nil, "primary", "rule.color", "Line color."), flg("-w", "--width", "<cols>", nil, "auto", "", "Custom width.")},
			Examples:    []string{`zsp rule "ARTIFACTS" --color secondary`},
		},
		{
			Name: "kv", Aliases: []string{"card"}, Category: "Structured Data & Code",
			Summary:  "Render aligned key-value pairs with dot-leaders inside a framed card.",
			Synopsis: "zsp kv [--title <t>] [--color <c>] <key1> <val1> [<key2> <val2> ...]", Stream: "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "key val", Required: true, Repeatable: true, Description: "Alternating `<key>` `<value>` pairs."}},
			Flags:       []FlagSpec{flg("-t", "--title", "<title>", nil, "", "", "Card header."), flg("-c", "--color", "<color>", nil, "secondary", "kv.color", "Key and border color.")},
			Examples:    []string{`zsp kv --title "Deploy Target" "Environment" "prod" "Branch" "main"`},
		},
		{
			Name: "cmd", Category: "Structured Data & Code",
			Summary:  "Render a single shell command preview with a colored gutter (`┃ ▸`) and syntax highlighting.",
			Synopsis: "zsp cmd <shell command>", Stream: "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "command", Required: true, Description: "Command string to preview."}},
			Examples:    []string{`zsp cmd "docker build --platform linux/amd64 -t api:latest ."`},
		},
		{
			Name: "code", Category: "Structured Data & Code",
			Summary:  "Render a framed multi-line Zsh code block (`╭─ title ───╮`) with syntax highlighting.",
			Synopsis: "zsp code [--title <t>] <line1> [<line2> ...]", Stream: "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "line", Required: true, Repeatable: true, Description: "Code lines."}},
			Flags:       []FlagSpec{flg("-t", "--title", "<title>", nil, "zsh", "", "Top-left block title.")},
			Examples:    []string{`zsp code --title "zsh: build" 'zsp step 1 2 "Compile"' 'zsp ok "Done" --meta "12ms"'`},
		},
		{
			Name: "tree", Aliases: []string{"list"}, Category: "Structured Data & Code",
			Summary:  "Render a hierarchical Unicode checklist (`├─ ●`, `╰─ ●`).",
			Synopsis: "zsp tree [--title <t>] [--color <c>] <item1> [<item2> ...]", Stream: "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "item", Required: true, Repeatable: true, Description: "Checklist items."}},
			Flags:       []FlagSpec{flg("-t", "--title", "<title>", nil, "", "", "Root header."), flg("-c", "--color", "<color>", nil, "primary", "", "Header color.")},
			Examples:    []string{`zsp tree --title "Preflight" "Config valid" "Migrations clean"`},
		},
		{
			Name: "style", Category: "Typography & Framing",
			Summary:  "Render inline styled text with custom foreground, background, gradient, or text attributes.",
			Synopsis: "zsp style <text> [--fg <c>] [--bg <c>] [--gradient <g>] [--bold] [--dim] [--italic] [--underline]", Stream: "stdout", ExitCode: "0",
			Positionals: []ArgSpec{{Name: "text", Required: true, Description: "Text to style."}},
			Flags: []FlagSpec{
				flg("", "--fg", "<color>", nil, "", "", "Foreground color."),
				flg("", "--bg", "<color>", nil, "", "", "Background color."),
				flg("-b", "--bold", "", nil, "false", "", "Bold text."),
				flg("", "--dim", "", nil, "false", "", "Dim text."),
				flg("-i", "--italic", "", nil, "false", "", "Italic text."),
				flg("-u", "--underline", "", nil, "false", "", "Underline text."),
			},
			Examples: []string{`zsp style "PRODUCTION" --fg text --bg fail --bold`},
		},
		{
			Name: "spin", Category: "Async & Progress",
			Summary:  "Animate a live spinner on `/dev/tty` while running a child command, keeping `stdout` untouched for `$(...)` capture.",
			Synopsis: "zsp spin [title] [--title <t>] [--style dots|pulse|meter|line] [--color <c>] [--persist] [--static] [--stdin] [--duration <ms>] [-- <command> [args...]]",
			Stream:   "Animation writes to `/dev/tty` (fallback: `stderr`). Child `stdout` and `stderr` pass through untouched; child `stdin` is passed through when piped or redirected; `/dev/null` when `stdin` is an interactive terminal unless `--stdin` is set.",
			ExitCode: "Propagates exact exit code of `<command>` (`128 + signal` when terminated by a signal: `129` on `SIGHUP`, `130` on `SIGINT`, `131` on `SIGQUIT`, `143` on `SIGTERM`).",
			Positionals: []ArgSpec{
				{Name: "title", Required: false, Description: "Spinner label (or pass `--title=<t>` for untrusted values)."},
				{Name: "-- <command> [args...]", Required: false, Description: "Child command executed while spinning."},
			},
			Flags: []FlagSpec{
				flg("-t", "--title", "<title>", nil, "Working...", "", "Spinner label (use `--title=\"$var\"` for untrusted titles)."),
				flg("-s", "--style", "<style>", []string{"dots", "pulse", "meter", "line"}, "dots", "spin.style", "Spinner animation frames."),
				flg("-c", "--color", "<color>", nil, "primary", "spin.color", "Spinner glyph color."),
				flg("-p", "--persist", "", nil, "false", "", "Print final `PASS`/`FAIL` badge with elapsed time on `/dev/tty`."),
				flg("", "--static", "", nil, "false", "", "Print the title once without redrawing, then print the final badge (recommended for commands that prompt, such as `sudo`, `ssh`, `git`, or `gpg`)."),
				flg("", "--stdin", "", nil, "false", "", "Pass parent `stdin` through to the child command instead of `/dev/null`."),
				flg("-d", "--duration", "<ms>", nil, "600", "", "Timed spin duration in ms when no `-- <command>` is given."),
			},
			Rules: []string{
				"Safe inside `VAR=$(zsp spin \"...\" --persist -- cmd args)`: animation renders on `/dev/tty`, never polluting `VAR`.",
				"Always separate the child command with `--` so child flags are not parsed by `zsp`. For untrusted titles, use `zsp spin --title=\"$var\" -- cmd`.",
				"Use `--static` for commands that may prompt for input (`sudo`, `ssh`, `git`, `gpg`) so the 80ms spinner redraw does not overwrite the prompt.",
				"With a terminal attached (`/dev/tty` available): the child stays in the terminal's process group so it can prompt (`sudo`, `ssh`, `git`, `gpg`). The first `Ctrl-C` waits for cleanup; a second `Ctrl-C` kills the child. A `SIGINT` sent directly to `zsp`'s process ID is not passed on (use `SIGTERM` instead). `SIGTERM`, `SIGHUP`, and `SIGQUIT` reach the child only (not grandchildren), followed by `SIGKILL` after 5s.",
				"Without a terminal (CI, cron): the child gets its own process group, and `SIGTERM`, `SIGHUP`, `SIGQUIT`, and `SIGINT` signal the whole process group, followed by `SIGKILL` after 5s. On Linux, `Pdeathsig` (`SIGKILL`) is set so the child is killed if `zsp` receives `SIGKILL`; macOS has no `Pdeathsig` equivalent, so `SIGKILL` of `zsp` on macOS can leave the child running when no terminal is attached.",
			},
			Examples: []string{
				`COUNT=$(zsp spin "Querying DB..." --persist -- psql -tA -c "SELECT count(*) FROM users")`,
				`zsp spin --title="$step" -- ./deploy.sh`,
				`zsp spin "Deploying with sudo..." --static -- sudo ./deploy.sh`,
			},
		},
		{
			Name: "bar", Category: "Async & Progress",
			Summary:  "Render a half-cell (`█` / `▌` / `░`) progress bar with percentage and `[current/total]` counter.",
			Synopsis: "zsp bar <current> <total> [label] [--in-place|-r] [--width <cols>] [--color <c>] [--gradient <g>]",
			Stream:   "stdout", ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "current", Required: true, Description: "Completed units."},
				{Name: "total", Required: true, Description: "Total units."},
				{Name: "label", Required: false, Description: "Prefix label."},
			},
			Flags: []FlagSpec{
				flg("-r", "--in-place", "", nil, "false", "", "Overwrite current terminal line (`\\r\\033[2K`); emits newline when `current >= total`."),
				flg("-w", "--width", "<cols>", nil, "auto", "", "Bar cell width."),
				flg("-c", "--color", "<color>", nil, "primary", "bar.color", "Filled bar color."),
			},
			Examples: []string{`for i in {1..50}; do zsp bar $i 50 "Syncing $i" --in-place; sleep 0.02; done`},
		},
		{
			Name: "progress", Category: "Async & Progress",
			Summary:  "Read lines from `stdin` in a pipeline and render a live progress bar on `/dev/tty`.",
			Synopsis: "<cmd> | zsp progress [--total <n>] [--title <label>] [--color <c>]",
			Stream:   "Reads `stdin`; renders bar on `/dev/tty` (fallback: `stderr`).", ExitCode: "0",
			Flags:    []FlagSpec{flg("", "--total", "<n>", nil, "100", "", "Expected line count."), flg("-t", "--title", "<label>", nil, "Processing", "", "Prefix label.")},
			Examples: []string{`find . -type f | zsp progress --total 100 --title "Indexing"`},
		},
		{
			Name: "themes", Category: "Themes & Configuration",
			Summary:  fmt.Sprintf("List all %d built-in color palettes with live swatches.", len(themeOrder)),
			Synopsis: "zsp themes", Stream: "stdout", ExitCode: "0",
			Examples: []string{`zsp themes`},
		},
		{
			Name: "theme", Category: "Themes & Configuration",
			Summary:  "Set persistent default color palette (`nes`, `snes`, `dracula`, etc.) or preview a palette.",
			Synopsis: "zsp theme [<name> | list | preview <name>]", Stream: "stdout", ExitCode: "0 (1 on unknown theme)",
			Examples: []string{`zsp theme nes`, `zsp theme preview nes`},
		},
		{
			Name: "config", Category: "Themes & Configuration",
			Summary:  "View, set, or reset persistent user defaults in `~/.config/zsplash/config.json`.",
			Synopsis: "zsp config [list | set <key> <value> | reset | path]", Stream: "stdout", ExitCode: "0 (1 on invalid key)",
			Examples: []string{`zsp config set banner.font compact`, `zsp config set color.primary "#FC7460"`, `zsp config reset`},
		},
		{
			Name: "demo", Category: "Reference & Showcase",
			Summary:  "Render the full visual showcase along with grouped Zsh commands.",
			Synopsis: "zsp demo [--theme <name>]", Stream: "stdout", ExitCode: "0",
			Examples: []string{`zsp demo`, `zsp demo --theme nes`},
		},
		{
			Name: "docs", Aliases: []string{"schema", "manual"}, Category: "Reference & Showcase",
			Summary:  "Output complete Markdown reference manual or machine-readable JSON CLI schema (`--json`).",
			Synopsis: "zsp docs [--markdown | --json | <command>]", Stream: "stdout", ExitCode: "0",
			Flags:    []FlagSpec{flg("", "--markdown", "", nil, "true", "", "Emit full Markdown reference manual."), flg("-j", "--json", "", nil, "false", "", "Emit machine-readable JSON CLI schema.")},
			Examples: []string{`zsp docs`, `zsp docs --json`, `zsp help spin`},
		},
	}
}

func findCommandSpec(name string) (CommandSpec, bool) {
	target := strings.ToLower(strings.TrimSpace(name))
	for _, cmd := range buildCommandSpecs(nil) {
		if cmd.Name == target {
			return cmd, true
		}
		for _, alias := range cmd.Aliases {
			if alias == target {
				return cmd, true
			}
		}
	}
	return CommandSpec{}, false
}

func handleDocsCommand(out io.Writer, mode string, args []string) int {
	for _, a := range args {
		if a == "--json" || a == "-j" {
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			_ = enc.Encode(buildCLISpec())
			return 0
		}
		if a == "--markdown" || a == "--md" || a == "all" {
			renderDocsMarkdown(out)
			return 0
		}
	}
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		topic := strings.ToLower(strings.TrimSpace(args[0]))
		if topic == "all" || topic == "full" || topic == "docs" {
			renderDocsMarkdown(out)
			return 0
		}
		if spec, ok := findCommandSpec(topic); ok {
			renderCommandHelp(out, spec)
			return 0
		}
		renderPillBadge(out, "FAIL", activePalette().Fail, fmt.Sprintf("Unknown command topic %q", topic), "try: zsp docs")
		return 1
	}
	if mode == "docs" {
		renderDocsMarkdown(out)
		return 0
	}
	printHelp(out)
	return 0
}

func renderCommandHelp(out io.Writer, c CommandSpec) {
	pal := activePalette()
	fmt.Fprintf(out, "%s%szsp %s%s  %s[%s]%s\n  %s\n\n", ansiBold(), pal.Primary.fgANSI(), c.Name, ansiReset(), ansiDim(), c.Category, ansiReset(), c.Summary)
	fmt.Fprintf(out, "%s%sSYNOPSIS%s\n  %s\n\n", ansiBold(), pal.Secondary.fgANSI(), ansiReset(), styleCommandTokens(c.Synopsis))
	fmt.Fprintf(out, "%s%sSTREAM & EXIT CONTRACT%s\n  Stream:    %s\n  Exit code: %s\n\n", ansiBold(), pal.Secondary.fgANSI(), ansiReset(), c.Stream, c.ExitCode)
	if len(c.Flags) > 0 {
		fmt.Fprintf(out, "%s%sFLAGS%s\n", ansiBold(), pal.Secondary.fgANSI(), ansiReset())
		for _, f := range c.Flags {
			lbl := f.Long
			if f.Short != "" {
				lbl = f.Short + ", " + f.Long
			}
			if f.ValueName != "" {
				lbl += " " + f.ValueName
			}
			fmt.Fprintf(out, "  %s%-26s%s %s\n", pal.Warn.fgANSI(), lbl, ansiReset(), f.Description)
		}
		fmt.Fprintln(out)
	}
	if len(c.Rules) > 0 {
		fmt.Fprintf(out, "%s%sNOTES & RULES%s\n", ansiBold(), pal.Secondary.fgANSI(), ansiReset())
		for _, r := range c.Rules {
			fmt.Fprintf(out, "  • %s\n", r)
		}
		fmt.Fprintln(out)
	}
	if len(c.Examples) > 0 {
		renderCodeBlock(out, "zsh: "+c.Name, c.Examples)
	}
}

func renderDocsMarkdown(out io.Writer) {
	spec := buildCLISpec()
	fmt.Fprintf(out, "# %s (`%s`) v%s Technical Reference Manual\n\n%s\n\n", spec.Name, spec.Alias, spec.Version, spec.Summary)
	fmt.Fprintln(out, "## 1. Core Execution & Stream Contracts")
	fmt.Fprintln(out, "- **Binaries**: `zsplash` and `zsp` are identical.")
	fmt.Fprintln(out, "- **Solid Colors by Default**: Solid theme colors are used by default; pass `--gradient <name>` to opt into gradients.")
	fmt.Fprintln(out, "- **Stream Isolation**: Decoration commands write to `stdout`. `zsp spin` and `zsp progress` animate on `/dev/tty` (fallback: `stderr`), leaving child `stdout` clean for `VAR=$(zsp spin ... -- cmd)` and preserving child exit codes (`$?`).")
	fmt.Fprintln(out, "\n## 2. Global Flags & Environment Variables")
	for _, gf := range spec.GlobalFlags {
		fmt.Fprintf(out, "- `%s`: %s\n", gf.Long, gf.Description)
	}
	for _, ev := range spec.EnvVars {
		fmt.Fprintf(out, "- `%s`: %s\n", ev.Name, ev.Description)
	}
	fmt.Fprintln(out, "\n## 3. Color System, Built-in Palettes & Gradients")
	for _, th := range spec.Themes {
		fmt.Fprintf(out, "- **`%s`** (`primary: %s`, `secondary: %s`, `pass: %s`, `warn: %s`, `fail: %s`, `info: %s`): %s\n",
			th.Name, th.Roles["primary"], th.Roles["secondary"], th.Roles["pass"], th.Roles["warn"], th.Roles["fail"], th.Roles["info"], th.Description)
	}
	fmt.Fprintln(out, "\n## 4. Persistent Configuration (`~/.config/zsplash/config.json`)")
	for _, ck := range spec.ConfigKeys {
		fmt.Fprintf(out, "- `%s` (default: `%s`): %s\n", ck.Key, ck.Default, ck.Description)
	}
	fmt.Fprintln(out, "\n## 5. Subcommand Reference")
	for _, cmd := range spec.Commands {
		fmt.Fprintf(out, "\n### `zsp %s`\n%s\n- **Synopsis**: `%s`\n- **Stream**: %s | **Exit Code**: %s\n", cmd.Name, cmd.Summary, cmd.Synopsis, cmd.Stream, cmd.ExitCode)
		for _, f := range cmd.Flags {
			fmt.Fprintf(out, "  - `%s %s`: %s\n", f.Long, f.ValueName, f.Description)
		}
		for _, r := range cmd.Rules {
			fmt.Fprintf(out, "  - **Note**: %s\n", r)
		}
	}
	fmt.Fprintln(out, "\n## 6. Using `zsp` in Scripts")
	fmt.Fprintln(out, "- **Text is cleaned by default**: Escape sequences, control characters, and Unicode bidi/format characters are stripped automatically. Only a leading `zsp --raw <cmd>` disables this.")
	fmt.Fprintln(out, "- **Pass untrusted values after `--`**: Use `zsp ok -- \"$msg\"`, `zsp tree --title Files -- \"${files[@]}\"`, or `zsp -- \"$text\"`. Without `--`, values like `--meta` or `--theme` are treated as flags.")
	fmt.Fprintln(out, "- **Always name the subcommand**: `zsp \"$var\"` treats words like `theme` or `config reset` as commands; use `zsp box -- \"$var\"` or `zsp -- \"$var\"`.")
	fmt.Fprintln(out, "- **Exit codes**: Display commands always exit `0`. `spin` passes on the child's exit code (`128 + signal` when killed by a signal: `129` for `SIGHUP`, `130` for `SIGINT`, `131` for `SIGQUIT`, `143` for `SIGTERM`). `progress` and `config` exit non-zero on error.")
	fmt.Fprintln(out, "- **`spin` in scripts**: `stdout` stays clean for `$(...)` capture; pass untrusted titles via `zsp spin --title=\"$var\" -- cmd`; use `--static` for commands that prompt (`sudo`, `ssh`, `git`, `gpg`); use `--stdin` only for interactive terminal input (piped `stdin` passes through automatically); prefer `SIGTERM` over `kill -INT` when a terminal is attached. `SIGHUP` and `SIGQUIT` are forwarded like `SIGTERM` (5s grace period then `SIGKILL`). On Linux, `Pdeathsig` stops the child if `zsp` is `SIGKILL`ed; on macOS, `SIGKILL` of `zsp` can leave the child running when no terminal is attached.")
	fmt.Fprintln(out, "- **Log-friendly output**: `NO_COLOR` and `TERM=dumb` disable color. When no terminal is attached (`CI`, `cron`, redirected logs), `spin` and `progress` automatically skip in-place redraws and print only static output.")
	fmt.Fprintln(out, "\n## 7. Complete Zsh Recipe")
	for _, r := range spec.Recipes {
		fmt.Fprintf(out, "```zsh\n%s\n```\n", r.Code)
	}
}
