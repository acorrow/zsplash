package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// ArgSpec documents a positional argument for a command.
type ArgSpec struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Repeatable  bool   `json:"repeatable,omitempty"`
	Description string `json:"description"`
}

// FlagSpec documents a command-line flag, its allowed values, default, and config key.
type FlagSpec struct {
	Short       string   `json:"short,omitempty"`
	Long        string   `json:"long"`
	ValueName   string   `json:"value_name,omitempty"`
	Allowed     []string `json:"allowed,omitempty"`
	Default     string   `json:"default,omitempty"`
	ConfigKey   string   `json:"config_key,omitempty"`
	Description string   `json:"description"`
}

// CommandSpec documents a subcommand's full behavioral contract.
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

// EnvVarSpec documents an environment variable recognized by zsplash.
type EnvVarSpec struct {
	Name        string `json:"name"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
}

// ConfigKeySpec documents a persistent key in ~/.config/zsplash/config.json.
type ConfigKeySpec struct {
	Key         string   `json:"key"`
	Allowed     []string `json:"allowed,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description"`
}

// ThemeHexSpec exports hex values for each role in a built-in theme palette.
type ThemeHexSpec struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Roles       map[string]string `json:"roles"`
}

// RecipeSpec documents a complete Zsh scripting pattern.
type RecipeSpec struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Code        string `json:"code"`
}

// CLISpec is the top-level machine-readable specification emitted by `zsp docs --json`.
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

func buildCLISpec() CLISpec {
	themeNames := []string{"modern", "nes", "snes", "gameboy", "dracula", "catppuccin", "nord", "gruvbox", "tokyonight", "amber"}
	var themes []ThemeHexSpec
	for _, name := range themeNames {
		t := builtinThemes[name]
		themes = append(themes, ThemeHexSpec{
			Name:        t.Name,
			Description: t.Description,
			Roles: map[string]string{
				"primary":   t.Primary.Hex(),
				"secondary": t.Secondary.Hex(),
				"pass":      t.Pass.Hex(),
				"warn":      t.Warn.Hex(),
				"fail":      t.Fail.Hex(),
				"info":      t.Info.Hex(),
				"muted":     t.Muted.Hex(),
				"text":      t.Text.Hex(),
			},
		})
	}

	namedHex := make(map[string]string, len(namedColors))
	for k, v := range namedColors {
		namedHex[k] = v.Hex()
	}

	return CLISpec{
		Name:    "zsplash",
		Alias:   "zsp",
		Version: version,
		Summary: "Fast, reactive CLI utility for Zsh and shell scripts to render colorful banners, boxes, status badges, phase headers, key-value cards, command previews, spinners, and progress bars.",
		GlobalFlags: []FlagSpec{
			{
				Short:       "-T",
				Long:        "--theme",
				ValueName:   "<theme>",
				Allowed:     themeNames,
				Default:     "modern (or value in config.json)",
				Description: "Temporarily apply a built-in color palette for the current command without modifying config.json. Can be placed anywhere before `--`.",
			},
			{
				Short:       "-h",
				Long:        "--help",
				Description: "Show overview help, or show detailed subcommand documentation when passed after a subcommand (`zsp <cmd> --help`).",
			},
			{
				Short:       "-v",
				Long:        "--version",
				Description: "Print version string (`zsplash (zsp) vX.Y.Z`) and exit.",
			},
		},
		EnvVars: []EnvVarSpec{
			{
				Name:        "ZSPLASH_THEME",
				Default:     "",
				Description: "Override the active color palette for the current shell session (lower precedence than `--theme`, higher precedence than `config.json`).",
			},
			{
				Name:        "ZSPLASH_CONFIG",
				Default:     "~/.config/zsplash/config.json",
				Description: "Custom file path for reading and writing persistent user configuration.",
			},
			{
				Name:        "ZSPLASH_WIDTH",
				Default:     "auto-detected via /dev/tty or stdout ioctl, then $COLUMNS, then 80",
				Description: "Force a specific terminal column width (integer >= 20) for wrapping and layout calculations.",
			},
			{
				Name:        "NO_COLOR",
				Default:     "",
				Description: "When set to any non-empty value (or when `TERM=dumb`), disables all ANSI color codes and uses plain ASCII bracket badges (`[PASS]`, `[WARN]`, `[FAIL]`, `[INFO]`).",
			},
		},
		SemanticRoles: []string{"primary", "secondary", "pass", "warn", "fail", "info", "muted", "text"},
		Themes:        themes,
		NamedColors:   namedHex,
		Gradients:     []string{"cyber", "ember", "ocean", "sunset", "matrix", "gold", "mono"},
		ConfigKeys:    buildConfigKeySpecs(),
		Commands:      buildCommandSpecs(),
		Recipes:       buildRecipeSpecs(),
	}
}

func buildConfigKeySpecs() []ConfigKeySpec {
	return []ConfigKeySpec{
		{Key: "theme", Allowed: []string{"modern", "nes", "snes", "gameboy", "dracula", "catppuccin", "nord", "gruvbox", "tokyonight", "amber"}, Default: "modern", Description: "Default color palette used across all commands."},
		{Key: "banner.font", Allowed: []string{"block", "compact", "line"}, Default: "block", Description: "Default font for `zsp banner`."},
		{Key: "banner.color", Allowed: []string{"<role>", "<named-color>", "#RRGGBB"}, Default: "primary", Description: "Default solid color for `zsp banner`."},
		{Key: "box.border", Allowed: []string{"rounded", "double", "heavy", "ascii"}, Default: "rounded", Description: "Default border style for `zsp box` and `zsp kv`."},
		{Key: "box.color", Allowed: []string{"<role>", "<named-color>", "#RRGGBB"}, Default: "primary", Description: "Default border color for `zsp box`."},
		{Key: "step.color", Allowed: []string{"<role>", "<named-color>", "#RRGGBB"}, Default: "primary", Description: "Default rule color for `zsp step`."},
		{Key: "rule.color", Allowed: []string{"<role>", "<named-color>", "#RRGGBB"}, Default: "primary", Description: "Default line color for `zsp rule`."},
		{Key: "kv.color", Allowed: []string{"<role>", "<named-color>", "#RRGGBB"}, Default: "secondary", Description: "Default key and border color for `zsp kv`."},
		{Key: "bar.color", Allowed: []string{"<role>", "<named-color>", "#RRGGBB"}, Default: "primary", Description: "Default filled bar color for `zsp bar` and `zsp progress`."},
		{Key: "spin.style", Allowed: []string{"dots", "pulse", "meter", "line"}, Default: "dots", Description: "Default animation style for `zsp spin`."},
		{Key: "spin.color", Allowed: []string{"<role>", "<named-color>", "#RRGGBB"}, Default: "primary", Description: "Default spinner frame color for `zsp spin`."},
		{Key: "color.primary", Allowed: []string{"<named-color>", "#RRGGBB"}, Default: "(from active theme)", Description: "Override the active theme's `primary` color slot."},
		{Key: "color.secondary", Allowed: []string{"<named-color>", "#RRGGBB"}, Default: "(from active theme)", Description: "Override the active theme's `secondary` color slot."},
		{Key: "color.pass", Allowed: []string{"<named-color>", "#RRGGBB"}, Default: "(from active theme)", Description: "Override the active theme's `pass` (success/ok) color slot."},
		{Key: "color.warn", Allowed: []string{"<named-color>", "#RRGGBB"}, Default: "(from active theme)", Description: "Override the active theme's `warn` color slot."},
		{Key: "color.fail", Allowed: []string{"<named-color>", "#RRGGBB"}, Default: "(from active theme)", Description: "Override the active theme's `fail` (error) color slot."},
		{Key: "color.info", Allowed: []string{"<named-color>", "#RRGGBB"}, Default: "(from active theme)", Description: "Override the active theme's `info` color slot."},
		{Key: "color.muted", Allowed: []string{"<named-color>", "#RRGGBB"}, Default: "(from active theme)", Description: "Override the active theme's `muted` color slot."},
		{Key: "color.text", Allowed: []string{"<named-color>", "#RRGGBB"}, Default: "(from active theme)", Description: "Override the active theme's `text` color slot."},
	}
}

func buildCommandSpecs() []CommandSpec {
	return []CommandSpec{
		{
			Name:     "banner",
			Category: "Typography & Framing",
			Summary:  "Render large, high-contrast Unicode block typography (A-Z, 0-9, punctuation) with optional framing and subtitle.",
			Synopsis: "zsp banner <text> [--font block|compact|line] [--color <color>] [--gradient <grad>] [--from <#hex> --to <#hex>] [--sub <subtitle>] [--frame] [--align left|center|right]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "text", Required: true, Description: "Headline text to render in large Unicode glyphs (automatically uppercased). Multiple positional words are joined with spaces."},
			},
			Flags: []FlagSpec{
				{Short: "-f", Long: "--font", ValueName: "<font>", Allowed: []string{"block", "compact", "line"}, Default: "block", ConfigKey: "banner.font", Description: "Glyph font: `block` (3-row heavy half-block), `compact` (2-row tight script header), or `line` (3-row thin box-drawing)."},
				{Short: "-c", Long: "--color", ValueName: "<color>", Allowed: []string{"primary", "secondary", "pass", "warn", "fail", "info", "muted", "text", "<named-color>", "#RRGGBB"}, Default: "primary", ConfigKey: "banner.color", Description: "Solid foreground color (default behavior uses solid color, not gradient)."},
				{Short: "-g", Long: "--gradient", ValueName: "<grad>", Allowed: []string{"cyber", "ember", "ocean", "sunset", "matrix", "gold", "mono"}, Default: "", Description: "Opt into a 24-bit TrueColor horizontal gradient across the banner."},
				{Long: "--from", ValueName: "<#RRGGBB>", Description: "Start color for a custom 2-stop gradient (requires `--to`)."},
				{Long: "--to", ValueName: "<#RRGGBB>", Description: "End color for a custom 2-stop gradient (requires `--from`)."},
				{Short: "-s", Long: "--sub", ValueName: "<subtitle>", Description: "Subtitle line rendered below the glyphs (or inside the bottom section when `--frame` is set)."},
				{Long: "--frame", Description: "Wrap the banner (and optional `--sub` subtitle) inside a rounded Unicode box."},
				{Short: "-a", Long: "--align", ValueName: "<align>", Allowed: []string{"left", "center", "right"}, Default: "left", Description: "Align the rendered banner within the current terminal width."},
			},
			Rules: []string{
				"Solid color is used by default (`primary` from the active theme or `banner.color` in config.json). Gradients are only used when `--gradient` or `--from`/`--to` are passed.",
				"If the rendered banner is wider than the current terminal width, each glyph row is safely truncated to fit without wrapping mid-glyph.",
			},
			Examples: []string{
				`zsp banner "DEPLOY" --font block --frame --sub "Production Pipeline"`,
				`zsp banner "STAGE 1" --font compact --color secondary`,
				`zsp banner "RETRO" --theme nes --color nes-red --frame`,
				`zsp banner "CYBER" --font line --gradient cyber`,
			},
		},
		{
			Name:     "box",
			Category: "Typography & Framing",
			Summary:  "Render a reactive Unicode callout box that word-wraps long text to fit the live terminal width.",
			Synopsis: "zsp box <text> [--title <title>] [--footer <footer>] [--border rounded|double|heavy|ascii] [--color <color>] [--gradient <grad>] [--align left|center|right] [--width <cols>] [--padding <n>]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "text", Required: true, Description: "Body text inside the box. Literal `\\n` sequences or actual newlines create multiple paragraphs; long lines automatically word-wrap to fit the terminal."},
			},
			Flags: []FlagSpec{
				{Short: "-t", Long: "--title", ValueName: "<title>", Description: "Header label embedded in the top border (`╭─ Title ───╮`)."},
				{Long: "--footer", ValueName: "<footer>", Description: "Footer label embedded in the bottom border (`╰─── Footer ─╯`)."},
				{Short: "-b", Long: "--border", ValueName: "<style>", Allowed: []string{"rounded", "double", "heavy", "ascii"}, Default: "rounded", ConfigKey: "box.border", Description: "Border character set."},
				{Short: "-c", Long: "--color", ValueName: "<color>", Default: "primary", ConfigKey: "box.color", Description: "Solid color for the box border and title."},
				{Short: "-g", Long: "--gradient", ValueName: "<grad>", Allowed: []string{"cyber", "ember", "ocean", "sunset", "matrix", "gold", "mono"}, Description: "Opt into a horizontal gradient across the top and bottom borders."},
				{Short: "-a", Long: "--align", ValueName: "<align>", Allowed: []string{"left", "center", "right"}, Default: "left", Description: "Horizontal alignment of the box within the terminal."},
				{Short: "-w", Long: "--width", ValueName: "<cols>", Default: "0 (auto-fit content up to terminal width)", Description: "Target outer box width in columns (automatically clamped if wider than the terminal)."},
				{Short: "-p", Long: "--padding", ValueName: "<spaces>", Default: "2", Description: "Horizontal inner padding (in spaces) on the left and right of the body text."},
			},
			Rules: []string{
				"Never overflows the detected terminal width; long lines wrap on word boundaries.",
				"Supports embedded `\\n` escape sequences in `<text>` for multi-line callouts.",
			},
			Examples: []string{
				`zsp box "All 14 migrations applied cleanly." --title "Database" --border rounded`,
				`zsp box "Line 1\nLine 2" --title "Release v1.4" --footer "sha: 8f3a1b" --border double --color pass`,
			},
		},
		{
			Name:     "ok",
			Aliases:  []string{"pass", "success"},
			Category: "Status & Phase Indicators",
			Summary:  "Render an inverted green `PASS` pill badge with a message and optional right-aligned dot-leader metadata.",
			Synopsis: "zsp ok <message> [--meta <detail>]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "message", Required: true, Description: "Status message displayed next to the `PASS` pill."},
			},
			Flags: []FlagSpec{
				{Short: "-m", Long: "--meta", ValueName: "<detail>", Description: "Right-aligned metadata connected to the message via dimmed dot-leaders (`····`), e.g. duration `42ms` or host name."},
			},
			Rules: []string{
				"Uses the active theme's `pass` color for the inverted background pill (`bgANSI` + bold dark foreground text).",
				"When `--meta` is provided, the dot-leader automatically expands or contracts to match the live terminal width.",
			},
			Examples: []string{
				`zsp ok "Database schema verified" --meta "42ms"`,
				`zsp ok "Build artifacts signed"`,
			},
		},
		{
			Name:     "info",
			Category: "Status & Phase Indicators",
			Summary:  "Render an inverted cyan/blue `INFO` pill badge with a message and optional right-aligned metadata.",
			Synopsis: "zsp info <message> [--meta <detail>]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "message", Required: true, Description: "Informational message displayed next to the `INFO` pill."},
			},
			Flags: []FlagSpec{
				{Short: "-m", Long: "--meta", ValueName: "<detail>", Description: "Right-aligned metadata connected via dot-leaders."},
			},
			Examples: []string{
				`zsp info "Worker pool initialized" --meta "8 threads"`,
			},
		},
		{
			Name:     "warn",
			Category: "Status & Phase Indicators",
			Summary:  "Render an inverted amber/gold `WARN` pill badge with a message and optional right-aligned metadata.",
			Synopsis: "zsp warn <message> [--meta <detail>]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "message", Required: true, Description: "Warning message displayed next to the `WARN` pill."},
			},
			Flags: []FlagSpec{
				{Short: "-m", Long: "--meta", ValueName: "<detail>", Description: "Right-aligned metadata connected via dot-leaders."},
			},
			Examples: []string{
				`zsp warn "Rate limit at 85% capacity" --meta "throttling"`,
			},
		},
		{
			Name:     "err",
			Aliases:  []string{"fail", "error"},
			Category: "Status & Phase Indicators",
			Summary:  "Render an inverted red `FAIL` pill badge with a message and optional right-aligned metadata.",
			Synopsis: "zsp err <message> [--meta <detail>]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "message", Required: true, Description: "Error message displayed next to the `FAIL` pill."},
			},
			Flags: []FlagSpec{
				{Short: "-m", Long: "--meta", ValueName: "<detail>", Description: "Right-aligned metadata connected via dot-leaders (e.g., `exit 1`)."},
			},
			Examples: []string{
				`zsp err "Upstream webhook timeout" --meta "exit 1"`,
			},
		},
		{
			Name:     "step",
			Category: "Status & Phase Indicators",
			Summary:  "Render a full-width phase divider (`━━━ [curr/total] ━━━━━━`) followed by a bold `▸ <title>` line.",
			Synopsis: "zsp step <current> <total> <title> [--color <color>] [--gradient <grad>]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "current", Required: true, Description: "Current step number or label (e.g. `1`)."},
				{Name: "total", Required: true, Description: "Total number of steps (e.g. `4`)."},
				{Name: "title", Required: true, Description: "Phase description printed on the second line after `▸`."},
			},
			Flags: []FlagSpec{
				{Short: "-c", Long: "--color", ValueName: "<color>", Default: "primary", ConfigKey: "step.color", Description: "Solid color for the heavy horizontal rule."},
				{Short: "-g", Long: "--gradient", ValueName: "<grad>", Description: "Optional gradient across the horizontal rule."},
			},
			Examples: []string{
				`zsp step 1 4 "Validating environment and credentials"`,
				`zsp step 2 4 "Building container image" --color secondary`,
			},
		},
		{
			Name:     "rule",
			Aliases:  []string{"hr"},
			Category: "Status & Phase Indicators",
			Summary:  "Render a full-width horizontal divider line (`───`) with an optional centered label.",
			Synopsis: "zsp rule [label] [--color <color>] [--gradient <grad>] [--char <rune>]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "label", Required: false, Description: "Optional centered text label surrounded by horizontal line segments."},
			},
			Flags: []FlagSpec{
				{Short: "-c", Long: "--color", ValueName: "<color>", Default: "primary", ConfigKey: "rule.color", Description: "Solid color for the rule."},
				{Short: "-g", Long: "--gradient", ValueName: "<grad>", Description: "Optional gradient across the rule."},
				{Long: "--char", ValueName: "<rune>", Default: "─", Description: "Single character repeated across the rule (e.g. `━`, `═`, `─`, `-`)."},
			},
			Examples: []string{
				`zsp rule`,
				`zsp rule "ARTIFACTS" --color secondary`,
			},
		},
		{
			Name:     "kv",
			Category: "Structured Data & Code",
			Summary:  "Render aligned key-value pairs connected by dot-leaders inside a framed summary card.",
			Synopsis: "zsp kv [--title <title>] [--color <color>] [--gradient <grad>] <key1> <val1> [<key2> <val2> ...]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "key val", Required: true, Repeatable: true, Description: "Alternating `<key>` `<value>` arguments. Pass as many pairs as needed."},
			},
			Flags: []FlagSpec{
				{Short: "-t", Long: "--title", ValueName: "<title>", Description: "Card title embedded in the top border."},
				{Short: "-c", Long: "--color", ValueName: "<color>", Default: "secondary", ConfigKey: "kv.color", Description: "Solid color for keys and card border."},
				{Short: "-g", Long: "--gradient", ValueName: "<grad>", Description: "Optional gradient for the card border."},
			},
			Rules: []string{
				"Positional arguments are consumed in pairs: `args[0]` is Key 1, `args[1]` is Value 1, `args[2]` is Key 2, `args[3]` is Value 2, and so on.",
				"Card width automatically scales to fit the longest key-value pair up to the terminal width (max 68 cols).",
			},
			Examples: []string{
				`zsp kv --title "Deploy Target" "Environment" "prod" "Branch" "main" "Host" "10.0.4.12"`,
			},
		},
		{
			Name:     "cmd",
			Category: "Structured Data & Code",
			Summary:  "Render a single shell command preview with a colored left gutter (`┃ ▸`) and syntax-highlighted tokens.",
			Synopsis: "zsp cmd <command string>",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "command", Required: true, Description: "Shell command line to preview (pass as a quoted string or multiple words)."},
			},
			Examples: []string{
				`zsp cmd "docker build --platform linux/amd64 -t api:latest ."`,
			},
		},
		{
			Name:     "code",
			Category: "Structured Data & Code",
			Summary:  "Render a framed multi-line Zsh code block (`╭─ title ───╮`) with syntax highlighting and preserved column alignment.",
			Synopsis: "zsp code [--title <title>] <line1> [<line2> ...]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "line", Required: true, Repeatable: true, Description: "Each positional argument is rendered as one line inside the framed code block (literal `\\n` inside an argument also splits into lines)."},
			},
			Flags: []FlagSpec{
				{Short: "-t", Long: "--title", ValueName: "<title>", Default: "zsh", Description: "Header label embedded in the top-left corner of the code block."},
			},
			Rules: []string{
				"Preserves exact spacing within each line so aligned arguments or flags remain vertically aligned.",
			},
			Examples: []string{
				`zsp code --title "zsh: deploy" 'zsp step 1 2 "Build"' 'zsp ok "Done" --meta "12ms"'`,
			},
		},
		{
			Name:     "tree",
			Category: "Structured Data & Code",
			Summary:  "Render a hierarchical Unicode checklist (`├─ ●`, `╰─ ●`) under an optional header.",
			Synopsis: "zsp tree [--title <title>] [--color <color>] <item1> [<item2> ...]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "item", Required: true, Repeatable: true, Description: "Checklist items rendered as tree branches."},
			},
			Flags: []FlagSpec{
				{Short: "-t", Long: "--title", ValueName: "<title>", Description: "Root header printed above the tree (`▸ <title>`)."},
				{Short: "-c", Long: "--color", ValueName: "<color>", Default: "primary", Description: "Color for the root header."},
			},
			Examples: []string{
				`zsp tree --title "Preflight Checks" "Auth token valid" "Cluster reachable" "Migrations clean"`,
			},
		},
		{
			Name:     "style",
			Category: "Typography & Framing",
			Summary:  "Render inline styled text with custom foreground, background, gradient, bold, dim, italic, or underline attributes.",
			Synopsis: "zsp style <text> [--fg <color>] [--bg <color>] [--gradient <grad>] [--bold] [--dim] [--italic] [--underline] [--no-newline|-n]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "text", Required: true, Description: "Text to style."},
			},
			Flags: []FlagSpec{
				{Long: "--fg", ValueName: "<color>", Description: "Foreground color (`<role>`, `<named-color>`, or `#RRGGBB`)."},
				{Long: "--bg", ValueName: "<color>", Description: "Background color (`<role>`, `<named-color>`, or `#RRGGBB`)."},
				{Short: "-g", Long: "--gradient", ValueName: "<grad>", Description: "Apply a horizontal TrueColor gradient across `<text>`."},
				{Long: "--bold", Description: "Apply bold text weight."},
				{Long: "--dim", Description: "Apply dim/faint intensity."},
				{Long: "--italic", Description: "Apply italic style."},
				{Long: "--underline", Description: "Apply underline style."},
				{Short: "-n", Long: "--no-newline", Description: "Omit the trailing newline (useful for inline shell prompts)."},
			},
			Examples: []string{
				`zsp style "PRODUCTION" --fg text --bg fail --bold`,
				`zsp style "v2.4.0" --fg pass --bold --no-newline`,
			},
		},
		{
			Name:     "spin",
			Category: "Async & Progress",
			Summary:  "Animate a live spinner on `/dev/tty` while executing a child command or timed wait, keeping `stdout` completely clean for Zsh variable capture.",
			Synopsis: "zsp spin <title> [--style dots|pulse|meter|line] [--color <color>] [--persist] [--duration <ms>] [-- <command> [args...]]",
			Stream:   "Animation renders to `/dev/tty` (fallback: `stderr`). Child process `stdout` and `stderr` pass through untouched to the caller's `stdout` and `stderr`.",
			ExitCode: "Returns the exact exit code of `<command>` (or `0` for `--duration` mode; `130` if interrupted by `SIGINT` / `Ctrl+C`).",
			Positionals: []ArgSpec{
				{Name: "title", Required: true, Description: "Status label shown next to the animated spinner (and in the final `PASS`/`FAIL` badge when `--persist` is enabled)."},
				{Name: "-- <command> [args...]", Required: false, Description: "Everything after `--` is executed directly as a child process while the spinner animates."},
			},
			Flags: []FlagSpec{
				{Short: "-s", Long: "--style", ValueName: "<style>", Allowed: []string{"dots", "pulse", "meter", "line"}, Default: "dots", ConfigKey: "spin.style", Description: "Spinner animation glyph set (`dots` Braille, `pulse` geometric, `meter` blocks, `line` ASCII)."},
				{Short: "-c", Long: "--color", ValueName: "<color>", Default: "primary", ConfigKey: "spin.color", Description: "Color of the animated spinner glyph."},
				{Short: "-p", Long: "--persist", Description: "When the task finishes, print a permanent `PASS` (or `FAIL` on non-zero exit) pill badge with elapsed time (`ms` or `s`) to `/dev/tty`."},
				{Short: "-d", Long: "--duration", ValueName: "<ms>", Default: "1000 (when no `-- <command>` is given)", Description: "Run a timed spinner for `<ms>` milliseconds when no child command is provided."},
			},
			Rules: []string{
				"CRITICAL FOR ZSH SCRIPTS: Because the spinner renders to `/dev/tty`, you can safely capture child `stdout` in a variable: `OUT=$(zsp spin \"Querying...\" --persist -- psql -tA -c \"SELECT 1\")`.",
				"Always place `--` before the child command and its arguments so child flags (such as `-s`, `-c`, `-p`) are passed to the child command rather than parsed by `zsp spin`.",
				"If the user presses `Ctrl+C` (`SIGINT` / `SIGTERM`), `zsp spin` terminates the child process, restores the terminal cursor (`\\033[?25h`), clears the active spinner line, and exits with code `130`.",
			},
			Examples: []string{
				`USER_COUNT=$(zsp spin "Querying users table..." --persist -- psql -tA -c "SELECT count(*) FROM users")`,
				`zsp spin "Fetching release manifest..." --style dots --persist -- curl -fsSL https://example.com/manifest.json -o /tmp/manifest.json`,
				`zsp spin "Warming cache..." --duration 800 --persist`,
			},
		},
		{
			Name:     "bar",
			Category: "Async & Progress",
			Summary:  "Render a reactive half-cell (`█` / `▌` / `░`) progress bar with percentage and `[current/total]` counter.",
			Synopsis: "zsp bar <current> <total> [label] [--in-place|-r] [--width <cols>] [--color <color>] [--gradient <grad>]",
			Stream:   "stdout",
			ExitCode: "0",
			Positionals: []ArgSpec{
				{Name: "current", Required: true, Description: "Completed units (integer >= 0)."},
				{Name: "total", Required: true, Description: "Total units (integer >= 1)."},
				{Name: "label", Required: false, Description: "Prefix label displayed to the left of the bar."},
			},
			Flags: []FlagSpec{
				{Short: "-r", Long: "--in-place", Description: "Use carriage return (`\\r\\033[2K`) to overwrite the current terminal line in a Zsh `for`/`while` loop; automatically prints a final newline (`\\n`) when `current >= total`."},
				{Short: "-w", Long: "--width", ValueName: "<cols>", Default: "0 (auto-scales to terminal width)", Description: "Explicit bar cell width."},
				{Short: "-c", Long: "--color", ValueName: "<color>", Default: "primary", ConfigKey: "bar.color", Description: "Solid color for the filled bar segment."},
				{Short: "-g", Long: "--gradient", ValueName: "<grad>", Description: "Optional horizontal gradient across the filled bar segment."},
			},
			Rules: []string{
				"In a Zsh loop, pass `--in-place` (or `-r`) on each iteration. When `<current>` reaches `<total>`, `zsp bar` automatically emits the trailing newline so subsequent script output starts on a fresh line.",
			},
			Examples: []string{
				`zsp bar 78 100 "Deploying shards"`,
				`for i in {1..50}; do zsp bar $i 50 "Syncing shard $i" --in-place; sleep 0.02; done`,
			},
		},
		{
			Name:     "progress",
			Category: "Async & Progress",
			Summary:  "Read lines from `stdin` in a pipeline and render a live updating progress bar on `/dev/tty`.",
			Synopsis: "<producer-cmd> | zsp progress [--total <n>] [--title <label>] [--color <color>] [--gradient <grad>]",
			Stream:   "Reads lines from `stdin`; renders the live bar on `/dev/tty` (fallback: `stderr`).",
			ExitCode: "0",
			Flags: []FlagSpec{
				{Long: "--total", ValueName: "<n>", Default: "100", Description: "Expected total number of input lines."},
				{Short: "-t", Long: "--title", ValueName: "<label>", Default: "Processing", Description: "Prefix label shown while processing; if omitted, uses the most recently read stdin line."},
				{Short: "-c", Long: "--color", ValueName: "<color>", Default: "primary", ConfigKey: "bar.color", Description: "Solid color for the progress bar."},
				{Short: "-g", Long: "--gradient", ValueName: "<grad>", Description: "Optional gradient for the progress bar."},
			},
			Examples: []string{
				`find . -maxdepth 2 -type f | zsp progress --total 100 --title "Indexing files"`,
			},
		},
		{
			Name:     "themes",
			Category: "Themes & Configuration",
			Summary:  "List all 10 built-in color palettes with live color swatches, descriptions, and active/default status.",
			Synopsis: "zsp themes",
			Stream:   "stdout",
			ExitCode: "0",
			Examples: []string{
				`zsp themes`,
			},
		},
		{
			Name:     "theme",
			Category: "Themes & Configuration",
			Summary:  "Set your persistent default color palette in `config.json`, list palettes, or preview a palette.",
			Synopsis: "zsp theme [<theme-name> | list | preview <theme-name>]",
			Stream:   "stdout",
			ExitCode: "0 on success; 1 if `<theme-name>` is unknown.",
			Positionals: []ArgSpec{
				{Name: "theme-name", Required: false, Description: "One of: `modern`, `nes`, `snes`, `gameboy`, `dracula`, `catppuccin`, `nord`, `gruvbox`, `tokyonight`, `amber`. If omitted or `list`, displays all palettes."},
			},
			Examples: []string{
				`zsp theme nes`,
				`zsp theme preview nes`,
				`zsp theme list`,
			},
		},
		{
			Name:     "config",
			Category: "Themes & Configuration",
			Summary:  "View, set, or reset persistent user defaults in `~/.config/zsplash/config.json` (or `$ZSPLASH_CONFIG`).",
			Synopsis: "zsp config [list | set <key> <value> | reset | path]",
			Stream:   "stdout",
			ExitCode: "0 on success; 1 on invalid key/value.",
			Positionals: []ArgSpec{
				{Name: "subcommand", Required: false, Description: "`list` (default), `set <key> <value>`, `reset`, or `path`."},
			},
			Examples: []string{
				`zsp config list`,
				`zsp config set theme nes`,
				`zsp config set banner.font compact`,
				`zsp config set box.border double`,
				`zsp config set color.primary "#FC7460"`,
				`zsp config reset`,
			},
		},
		{
			Name:     "demo",
			Category: "Reference & Showcase",
			Summary:  "Render the visual showcase of all `zsplash` components along with the grouped Zsh commands below each section.",
			Synopsis: "zsp demo [--theme <theme>]",
			Stream:   "stdout",
			ExitCode: "0",
			Examples: []string{
				`zsp demo`,
				`zsp demo --theme nes`,
			},
		},
		{
			Name:     "docs",
			Aliases:  []string{"schema", "manual"},
			Category: "Reference & Showcase",
			Summary:  "Output the complete technical specification of `zsplash` in Markdown (default) or machine-readable JSON (`--json`).",
			Synopsis: "zsp docs [--markdown | --json | <command>]",
			Stream:   "stdout",
			ExitCode: "0",
			Flags: []FlagSpec{
				{Long: "--markdown", Description: "Emit the complete single-page Markdown reference manual (default)."},
				{Long: "--json", Description: "Emit the complete machine-readable JSON specification of all commands, flags, themes, colors, config keys, and Zsh recipes."},
			},
			Examples: []string{
				`zsp docs`,
				`zsp docs --json`,
				`zsp docs spin`,
				`zsp help banner`,
			},
		},
	}
}

func buildRecipeSpecs() []RecipeSpec {
	return []RecipeSpec{
		{
			Title:       "Complete Production Deployment Script",
			Description: "Shows phase headers (`step`), `/dev/tty` spinners (`spin`) that capture stdout cleanly into Zsh variables while checking exit codes, a summary key-value card (`kv`), and status badges (`ok`/`err`).",
			Code: `#!/usr/bin/env zsh
set -uo pipefail

# Optional: pin a theme for this script without altering the user's global config
export ZSPLASH_THEME="nes"

zsp banner "DEPLOY" --font compact --frame --sub "Production Release Pipeline"

zsp step 1 3 "Verifying cluster & database connectivity"
DB_VERSION=$(zsp spin "Checking PostgreSQL version..." --persist -- psql -tA -c "SHOW server_version;")
if [[ $? -ne 0 ]]; then
  zsp err "Database connection failed" --meta "abort"
  exit 1
fi
zsp info "Connected to PostgreSQL" --meta "$DB_VERSION"

zsp step 2 3 "Syncing release artifacts"
for i in {1..20}; do
  zsp bar $i 20 "Uploading chunk $i" --in-place
  sleep 0.02
done
zsp ok "All 20 chunks uploaded" --meta "100%"

zsp step 3 3 "Release Summary"
zsp kv --title "Deployment" \
  "Environment" "production" \
  "Database"    "$DB_VERSION" \
  "Status"      "Healthy"`,
		},
		{
			Title:       "Capturing Command Output While Displaying a Spinner",
			Description: "Because `zsp spin` renders its animation exclusively to `/dev/tty` (falling back to `stderr`), the child command's `stdout` remains completely uncorrupted inside `$(...)`.",
			Code: `# Capture JSON or text output from a slow HTTP or CLI call:
COMMIT_SHA=$(zsp spin "Resolving latest commit..." --persist -- git rev-parse --short HEAD)
zsp ok "Resolved commit" --meta "$COMMIT_SHA"`,
		},
	}
}

func findCommandSpec(name string) (CommandSpec, bool) {
	target := strings.ToLower(strings.TrimSpace(name))
	for _, cmd := range buildCommandSpecs() {
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

// handleDocsCommand handles `zsp docs [--json|--markdown|<cmd>]` and `zsp help [<cmd>|all]`.
func handleDocsCommand(out io.Writer, mode string, args []string) int {
	for _, a := range args {
		if a == "--json" || a == "-j" {
			return renderDocsJSON(out)
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
		pal := activePalette()
		renderPillBadge(out, "FAIL", pal.Fail, fmt.Sprintf("Unknown command topic %q (try: zsp help or zsp docs)", topic), "")
		return 1
	}

	if mode == "docs" {
		renderDocsMarkdown(out)
		return 0
	}

	printHelp(out)
	return 0
}

func renderDocsJSON(out io.Writer) int {
	spec := buildCLISpec()
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(spec); err != nil {
		return 1
	}
	return 0
}

// renderCommandHelp prints a detailed, structured manual page for a single subcommand.
func renderCommandHelp(out io.Writer, c CommandSpec) {
	pal := activePalette()
	title := fmt.Sprintf("zsp %s", c.Name)
	if len(c.Aliases) > 0 {
		title += fmt.Sprintf(" (aliases: %s)", strings.Join(c.Aliases, ", "))
	}

	fmt.Fprintf(out, "%s%s%s%s  %s[%s]%s\n",
		ansiBold(), pal.Primary.fgANSI(), title, ansiReset(),
		ansiDim(), c.Category, ansiReset())
	fmt.Fprintf(out, "  %s\n\n", c.Summary)

	fmt.Fprintf(out, "%s%sSYNOPSIS%s\n", ansiBold(), pal.Secondary.fgANSI(), ansiReset())
	fmt.Fprintf(out, "  %s\n\n", styleCommandTokens(c.Synopsis))

	fmt.Fprintf(out, "%s%sSTREAM & EXIT CONTRACT%s\n", ansiBold(), pal.Secondary.fgANSI(), ansiReset())
	fmt.Fprintf(out, "  Stream:    %s\n", c.Stream)
	fmt.Fprintf(out, "  Exit code: %s\n\n", c.ExitCode)

	if len(c.Positionals) > 0 {
		fmt.Fprintf(out, "%s%sPOSITIONAL ARGUMENTS%s\n", ansiBold(), pal.Secondary.fgANSI(), ansiReset())
		for _, p := range c.Positionals {
			req := "optional"
			if p.Required {
				req = "required"
			}
			if p.Repeatable {
				req += ", repeatable"
			}
			fmt.Fprintf(out, "  %s%-22s%s (%s) %s\n",
				ansiBold()+pal.Text.fgANSI(), p.Name, ansiReset(), req, p.Description)
		}
		fmt.Fprintln(out)
	}

	if len(c.Flags) > 0 {
		fmt.Fprintf(out, "%s%sFLAGS%s\n", ansiBold(), pal.Secondary.fgANSI(), ansiReset())
		for _, f := range c.Flags {
			flagLabel := f.Long
			if f.Short != "" {
				flagLabel = f.Short + ", " + f.Long
			}
			if f.ValueName != "" {
				flagLabel += " " + f.ValueName
			}
			fmt.Fprintf(out, "  %s%-28s%s %s\n",
				pal.Warn.fgANSI(), flagLabel, ansiReset(), f.Description)
			var meta []string
			if len(f.Allowed) > 0 {
				meta = append(meta, "allowed: "+strings.Join(f.Allowed, " | "))
			}
			if f.Default != "" {
				meta = append(meta, "default: "+f.Default)
			}
			if f.ConfigKey != "" {
				meta = append(meta, "config key: "+f.ConfigKey)
			}
			if len(meta) > 0 {
				fmt.Fprintf(out, "  %-28s %s(%s)%s\n", "", ansiDim(), strings.Join(meta, "; "), ansiReset())
			}
		}
		fmt.Fprintln(out)
	}

	if len(c.Rules) > 0 {
		fmt.Fprintf(out, "%s%sBEHAVIORAL RULES%s\n", ansiBold(), pal.Secondary.fgANSI(), ansiReset())
		for _, r := range c.Rules {
			fmt.Fprintf(out, "  * %s\n", r)
		}
		fmt.Fprintln(out)
	}

	if len(c.Examples) > 0 {
		renderCodeBlock(out, fmt.Sprintf("zsh: %s examples", c.Name), c.Examples)
	}
}

// renderDocsMarkdown emits the complete, self-contained Markdown reference manual to out.
func renderDocsMarkdown(out io.Writer) {
	spec := buildCLISpec()

	fmt.Fprintf(out, "# %s (`%s`) v%s Technical Reference Manual\n\n", spec.Name, spec.Alias, spec.Version)
	fmt.Fprintf(out, "%s\n\n", spec.Summary)

	fmt.Fprintln(out, "## 1. Core Execution & Stream Contracts")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "- **Binary Names**: `zsplash` and `zsp` are identical binaries (`zsp` is installed as a symlink next to `zsplash`).")
	fmt.Fprintln(out, "- **Solid Colors by Default**: Every component renders solid colors from the active palette by default. Gradients are strictly opt-in via `--gradient <name>` (or `--from #RRGGBB --to #RRGGBB`).")
	fmt.Fprintln(out, "- **Standard Decoration Commands (`stdout`)**: `banner`, `box`, `ok`, `info`, `warn`, `err`, `step`, `rule`, `kv`, `cmd`, `code`, `tree`, `style`, and `bar` write formatted output to `stdout` and exit `0`.")
	fmt.Fprintln(out, "- **Async Commands (`/dev/tty` Isolation)**: `spin` and `progress` render live animations directly to `/dev/tty` (falling back to `stderr` if no controlling tty exists). In `zsp spin <title> -- <cmd> [args...]`, the child command's `stdout` and `stderr` pass through untouched so you can capture command output cleanly in Zsh (`OUT=$(zsp spin \"...\" -- cmd)`) and inspect the child's exact exit status (`$?`).")
	fmt.Fprintln(out, "- **Signal Safety (`SIGINT` / `SIGTERM`)**: Pressing `Ctrl+C` during `zsp spin` terminates the child process, restores the terminal cursor (`\\033[?25h`), clears the spinner line, and exits with code `130`.")
	fmt.Fprintln(out, "- **Reactive Terminal Width**: Every component detects the live terminal width on each render (`$ZSPLASH_WIDTH` -> `/dev/tty` ioctl -> `stdout`/`stderr` ioctl -> `$COLUMNS` -> `80`) so resizing a terminal window never scrambles borders or progress bars.")
	fmt.Fprintln(out)

	fmt.Fprintln(out, "## 2. Global Flags & Environment Variables")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "### Global Flags")
	fmt.Fprintln(out, "| Flag | Values | Default | Description |")
	fmt.Fprintln(out, "| :--- | :--- | :--- | :--- |")
	for _, gf := range spec.GlobalFlags {
		fName := gf.Long
		if gf.Short != "" {
			fName = gf.Short + ", " + gf.Long
		}
		if gf.ValueName != "" {
			fName += " " + gf.ValueName
		}
		allowed := "-"
		if len(gf.Allowed) > 0 {
			allowed = strings.Join(gf.Allowed, ", ")
		}
		def := gf.Default
		if def == "" {
			def = "-"
		}
		fmt.Fprintf(out, "| `%s` | `%s` | `%s` | %s |\n", fName, allowed, def, gf.Description)
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "### Environment Variables")
	fmt.Fprintln(out, "| Variable | Default | Description |")
	fmt.Fprintln(out, "| :--- | :--- | :--- |")
	for _, ev := range spec.EnvVars {
		def := ev.Default
		if def == "" {
			def = "(unset)"
		}
		fmt.Fprintf(out, "| `%s` | `%s` | %s |\n", ev.Name, def, ev.Description)
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "## 3. Color System, Built-in Palettes & Gradients")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Any `--color`, `--fg`, or `--bg` flag accepts:")
	fmt.Fprintf(out, "1. **Semantic Theme Roles**: `%s`\n", strings.Join(spec.SemanticRoles, "`, `"))
	fmt.Fprintln(out, "2. **Named Colors**: `cyan`, `magenta`, `purple`, `violet`, `blue`, `sky`, `green`, `mint`, `lime`, `amber`, `gold`, `yellow`, `orange`, `coral`, `red`, `crimson`, `rose`, `pink`, `white`, `slate`, `gray`, `black`, plus NES hardware names (`nes-red`, `nes-blue`, `nes-green`, `nes-gold`, `nes-orange`, `nes-lava`, `nes-cobalt`, `nes-magenta`, `nes-teal`, `nes-gray`).")
	fmt.Fprintln(out, "3. **24-bit Hex Codes**: `#RRGGBB` (e.g. `#E40058`).")
	fmt.Fprintln(out)

	fmt.Fprintln(out, "### Built-in Theme Palettes (`--theme <name>` or `zsp theme <name>`)")
	fmt.Fprintln(out, "| Theme | Primary | Secondary | Pass | Warn | Fail | Info | Description |")
	fmt.Fprintln(out, "| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |")
	for _, th := range spec.Themes {
		fmt.Fprintf(out, "| `%s` | `%s` | `%s` | `%s` | `%s` | `%s` | `%s` | %s |\n",
			th.Name,
			th.Roles["primary"], th.Roles["secondary"], th.Roles["pass"],
			th.Roles["warn"], th.Roles["fail"], th.Roles["info"],
			th.Description)
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "### Opt-in Gradients (`--gradient <name>`)\nAvailable gradient presets: `%s` (or custom `--from #RRGGBB --to #RRGGBB`).\n\n", strings.Join(spec.Gradients, "`, `"))

	fmt.Fprintln(out, "## 4. Persistent Configuration (`~/.config/zsplash/config.json`)")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Use `zsp config set <key> <value>` to override any built-in default:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "| Key | Allowed Values | Built-in Default | Description |")
	fmt.Fprintln(out, "| :--- | :--- | :--- | :--- |")
	for _, ck := range spec.ConfigKeys {
		fmt.Fprintf(out, "| `%s` | `%s` | `%s` | %s |\n",
			ck.Key, strings.Join(ck.Allowed, " \\| "), ck.Default, ck.Description)
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "## 5. Subcommand Reference")
	fmt.Fprintln(out)
	for _, cmd := range spec.Commands {
		heading := fmt.Sprintf("### `zsp %s`", cmd.Name)
		if len(cmd.Aliases) > 0 {
			heading += fmt.Sprintf(" (aliases: `%s`)", strings.Join(cmd.Aliases, "`, `"))
		}
		fmt.Fprintln(out, heading)
		fmt.Fprintf(out, "%s\n\n", cmd.Summary)
		fmt.Fprintf(out, "- **Synopsis**: `%s`\n", cmd.Synopsis)
		fmt.Fprintf(out, "- **Stream**: %s\n", cmd.Stream)
		fmt.Fprintf(out, "- **Exit Code**: %s\n", cmd.ExitCode)
		if len(cmd.Positionals) > 0 {
			fmt.Fprintln(out, "- **Positional Arguments**:")
			for _, p := range cmd.Positionals {
				req := "optional"
				if p.Required {
					req = "required"
				}
				if p.Repeatable {
					req += ", repeatable"
				}
				fmt.Fprintf(out, "  - `<%s>` (%s): %s\n", p.Name, req, p.Description)
			}
		}
		if len(cmd.Flags) > 0 {
			fmt.Fprintln(out, "- **Flags**:")
			for _, f := range cmd.Flags {
				label := f.Long
				if f.Short != "" {
					label = f.Short + ", " + f.Long
				}
				if f.ValueName != "" {
					label += " " + f.ValueName
				}
				var details []string
				if len(f.Allowed) > 0 {
					details = append(details, "allowed: `"+strings.Join(f.Allowed, "|")+"`")
				}
				if f.Default != "" {
					details = append(details, "default: `"+f.Default+"`")
				}
				if f.ConfigKey != "" {
					details = append(details, "config: `"+f.ConfigKey+"`")
				}
				suffix := ""
				if len(details) > 0 {
					suffix = " (" + strings.Join(details, ", ") + ")"
				}
				fmt.Fprintf(out, "  - `%s`: %s%s\n", label, f.Description, suffix)
			}
		}
		if len(cmd.Rules) > 0 {
			fmt.Fprintln(out, "- **Rules**:")
			for _, r := range cmd.Rules {
				fmt.Fprintf(out, "  - %s\n", r)
			}
		}
		if len(cmd.Examples) > 0 {
			fmt.Fprintln(out, "\n```zsh")
			for _, ex := range cmd.Examples {
				fmt.Fprintln(out, ex)
			}
			fmt.Fprintln(out, "```")
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintln(out, "## 6. Complete Zsh Scripting Recipes")
	fmt.Fprintln(out)
	for _, rec := range spec.Recipes {
		fmt.Fprintf(out, "### %s\n%s\n\n```zsh\n%s\n```\n\n", rec.Title, rec.Description, rec.Code)
	}

	// Ensure deterministic named color listing at bottom
	var colorKeys []string
	for k := range spec.NamedColors {
		colorKeys = append(colorKeys, k)
	}
	sort.Strings(colorKeys)
}
