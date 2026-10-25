package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var version = "0.5.2"

func main() {
	code := runWithExit(os.Args[1:], os.Stdout, os.Stderr)
	if code != 0 {
		os.Exit(code)
	}
}

func run(args []string, out io.Writer) {
	_ = runWithExit(args, out, out)
}

// extractGlobalThemeFlag strips --theme <name> from args and sets cliThemeOverride.
func extractGlobalThemeFlag(args []string) []string {
	cliThemeOverride = ""
	var filtered []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			filtered = append(filtered, args[i:]...)
			break
		}
		if (a == "--theme" || a == "-T") && i+1 < len(args) {
			cliThemeOverride = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(a, "--theme=") {
			cliThemeOverride = strings.TrimPrefix(a, "--theme=")
			continue
		}
		if strings.HasPrefix(a, "-T=") {
			cliThemeOverride = strings.TrimPrefix(a, "-T=")
			continue
		}
		filtered = append(filtered, a)
	}
	return filtered
}

func runWithExit(rawArgs []string, out io.Writer, errOut io.Writer) int {
	activeConfig = loadUserConfig()
	args := extractGlobalThemeFlag(rawArgs)

	if len(args) < 1 {
		printHelp(out)
		return 0
	}

	cmd := args[0]
	rest := args[1:]
	pal := activePalette()

	if cmd != "help" && cmd != "-h" && cmd != "--help" && cmd != "docs" && cmd != "schema" && cmd != "manual" && hasHelpFlag(rest) {
		if spec, ok := findCommandSpec(cmd); ok {
			renderCommandHelp(out, spec)
			return 0
		}
	}

	switch cmd {
	case "-v", "--version", "version":
		fmt.Fprintf(out, "zsplash (zsp) v%s\n", version)
		return 0

	case "-h", "--help", "help":
		return handleDocsCommand(out, "help", rest)

	case "docs", "schema", "manual":
		return handleDocsCommand(out, "docs", rest)

	case "themes":
		renderThemesCatalog(out)
		return 0

	case "theme":
		if len(rest) == 0 || rest[0] == "list" {
			renderThemesCatalog(out)
			return 0
		}
		target := strings.ToLower(strings.TrimSpace(rest[0]))
		if target == "preview" && len(rest) > 1 {
			cliThemeOverride = strings.ToLower(strings.TrimSpace(rest[1]))
			renderDemo(out)
			return 0
		}
		if err := setConfigKey("theme", target); err != nil {
			renderPillBadge(out, "FAIL", pal.Fail, err.Error(), "")
			return 1
		}
		p := activePalette()
		renderPillBadge(out, "PASS", p.Pass, fmt.Sprintf("Default theme set to %q", target), configFilePath())
		return 0

	case "config":
		return handleConfigCommand(out, rest)

	case "banner", "big":
		opts := parseBannerFlags(rest)
		renderBanner(out, opts)
		return 0

	case "box":
		opts := parseBoxFlags(rest)
		renderBox(out, opts)
		return 0

	case "ok", "pass", "success":
		msg, meta := parseBadgeFlags(rest)
		renderPillBadge(out, "PASS", pal.Pass, msg, meta)
		return 0

	case "warn", "warning":
		msg, meta := parseBadgeFlags(rest)
		renderPillBadge(out, "WARN", pal.Warn, msg, meta)
		return 0

	case "err", "error", "fail":
		msg, meta := parseBadgeFlags(rest)
		renderPillBadge(out, "FAIL", pal.Fail, msg, meta)
		return 0

	case "info":
		msg, meta := parseBadgeFlags(rest)
		renderPillBadge(out, "INFO", pal.Info, msg, meta)
		return 0

	case "step":
		curr, total, title, colorSpec, gradSpec := parseStepFlags(rest)
		renderStep(out, curr, total, title, colorSpec, gradSpec)
		return 0

	case "rule", "hr":
		label, colorSpec, gradSpec, width := parseRuleFlags(rest)
		renderRule(out, label, colorSpec, gradSpec, width)
		return 0

	case "kv", "card":
		title, colorSpec, gradSpec, pairs := parseKVFlags(rest)
		renderKV(out, title, colorSpec, gradSpec, pairs)
		return 0

	case "cmd":
		renderCmd(out, strings.Join(rest, " "))
		return 0

	case "code":
		title, _, items := parseTreeFlags(rest)
		renderCodeBlock(out, title, items)
		return 0

	case "tree", "list":
		title, colorSpec, items := parseTreeFlags(rest)
		renderTree(out, title, items, colorSpec)
		return 0

	case "style":
		opts := parseStyleFlags(rest)
		renderStyle(out, opts)
		return 0

	case "bar":
		opts := parseBarFlags(rest)
		renderBar(out, opts)
		return 0

	case "progress":
		total, title, colorSpec, gradSpec := parseProgressFlags(rest)
		runStreamProgress(os.Stdin, errOut, total, title, colorSpec, gradSpec)
		return 0

	case "spin":
		opts := parseSpinFlags(rest)
		return runSpinner(out, errOut, opts)

	case "demo":
		renderDemo(out)
		return 0

	default:
		renderBox(out, BoxOptions{
			Text: strings.Join(args, " "),
		})
		return 0
	}
}

func parseBannerFlags(args []string) BannerOptions {
	var opts BannerOptions
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case (a == "--font" || a == "-f") && i+1 < len(args):
			opts.Font = args[i+1]
			i++
		case strings.HasPrefix(a, "--font="):
			opts.Font = strings.TrimPrefix(a, "--font=")
		case (a == "--gradient" || a == "-g") && i+1 < len(args):
			opts.Gradient = args[i+1]
			i++
		case strings.HasPrefix(a, "--gradient="):
			opts.Gradient = strings.TrimPrefix(a, "--gradient=")
		case a == "--from" && i+1 < len(args):
			opts.FromHex = args[i+1]
			i++
		case a == "--to" && i+1 < len(args):
			opts.ToHex = args[i+1]
			i++
		case (a == "--color" || a == "-c") && i+1 < len(args):
			opts.Color = args[i+1]
			i++
		case strings.HasPrefix(a, "--color="):
			opts.Color = strings.TrimPrefix(a, "--color=")
		case (a == "--sub" || a == "--subtitle") && i+1 < len(args):
			opts.Subtitle = args[i+1]
			i++
		case a == "--frame":
			opts.Frame = true
		default:
			pos = append(pos, a)
		}
	}
	opts.Text = strings.Join(pos, " ")
	return opts
}

func parseBoxFlags(args []string) BoxOptions {
	var opts BoxOptions
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case (a == "--title" || a == "-t") && i+1 < len(args):
			opts.Title = args[i+1]
			i++
		case strings.HasPrefix(a, "--title="):
			opts.Title = strings.TrimPrefix(a, "--title=")
		case a == "--footer" && i+1 < len(args):
			opts.Footer = args[i+1]
			i++
		case strings.HasPrefix(a, "--footer="):
			opts.Footer = strings.TrimPrefix(a, "--footer=")
		case (a == "--border" || a == "-b") && i+1 < len(args):
			opts.Border = args[i+1]
			i++
		case strings.HasPrefix(a, "--border="):
			opts.Border = strings.TrimPrefix(a, "--border=")
		case (a == "--color" || a == "-c") && i+1 < len(args):
			opts.Color = args[i+1]
			i++
		case strings.HasPrefix(a, "--color="):
			opts.Color = strings.TrimPrefix(a, "--color=")
		case (a == "--gradient" || a == "-g") && i+1 < len(args):
			opts.Gradient = args[i+1]
			i++
		case strings.HasPrefix(a, "--gradient="):
			opts.Gradient = strings.TrimPrefix(a, "--gradient=")
		case a == "--from" && i+1 < len(args):
			opts.FromHex = args[i+1]
			i++
		case a == "--to" && i+1 < len(args):
			opts.ToHex = args[i+1]
			i++
		case (a == "--width" || a == "-w") && i+1 < len(args):
			opts.Width, _ = strconv.Atoi(args[i+1])
			i++
		default:
			pos = append(pos, a)
		}
	}
	opts.Text = strings.Join(pos, " ")
	if opts.Text == "" {
		opts.Text = "zsplash"
	}
	return opts
}

func parseBadgeFlags(args []string) (string, string) {
	var meta string
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case (a == "--meta" || a == "-m") && i+1 < len(args):
			meta = args[i+1]
			i++
		case strings.HasPrefix(a, "--meta="):
			meta = strings.TrimPrefix(a, "--meta=")
		default:
			pos = append(pos, a)
		}
	}
	return strings.Join(pos, " "), meta
}

func parseStepFlags(args []string) (string, string, string, string, string) {
	var colorSpec, gradSpec string
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case (a == "--color" || a == "-c") && i+1 < len(args):
			colorSpec = args[i+1]
			i++
		case (a == "--gradient" || a == "-g") && i+1 < len(args):
			gradSpec = args[i+1]
			i++
		default:
			pos = append(pos, a)
		}
	}
	if len(pos) == 0 {
		return "1", "1", "Step", colorSpec, gradSpec
	}
	if len(pos) == 1 {
		return "1", "1", pos[0], colorSpec, gradSpec
	}
	if len(pos) == 2 {
		return pos[0], pos[1], "", colorSpec, gradSpec
	}
	return pos[0], pos[1], strings.Join(pos[2:], " "), colorSpec, gradSpec
}

func parseRuleFlags(args []string) (string, string, string, int) {
	var colorSpec, gradSpec string
	var width int
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case (a == "--color" || a == "-c") && i+1 < len(args):
			colorSpec = args[i+1]
			i++
		case (a == "--gradient" || a == "-g") && i+1 < len(args):
			gradSpec = args[i+1]
			i++
		case (a == "--width" || a == "-w") && i+1 < len(args):
			width, _ = strconv.Atoi(args[i+1])
			i++
		default:
			pos = append(pos, a)
		}
	}
	return strings.Join(pos, " "), colorSpec, gradSpec, width
}

func parseKVFlags(args []string) (string, string, string, []string) {
	var title, colorSpec, gradSpec string
	var pairs []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case (a == "--title" || a == "-t") && i+1 < len(args):
			title = args[i+1]
			i++
		case strings.HasPrefix(a, "--title="):
			title = strings.TrimPrefix(a, "--title=")
		case (a == "--color" || a == "-c") && i+1 < len(args):
			colorSpec = args[i+1]
			i++
		case (a == "--gradient" || a == "-g") && i+1 < len(args):
			gradSpec = args[i+1]
			i++
		default:
			pairs = append(pairs, a)
		}
	}
	return title, colorSpec, gradSpec, pairs
}

func parseTreeFlags(args []string) (string, string, []string) {
	var title, colorSpec string
	var items []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case (a == "--title" || a == "-t") && i+1 < len(args):
			title = args[i+1]
			i++
		case (a == "--color" || a == "-c") && i+1 < len(args):
			colorSpec = args[i+1]
			i++
		default:
			items = append(items, a)
		}
	}
	return title, colorSpec, items
}

func parseStyleFlags(args []string) StyleOptions {
	var opts StyleOptions
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--fg" && i+1 < len(args):
			opts.Fg = args[i+1]
			i++
		case a == "--bg" && i+1 < len(args):
			opts.Bg = args[i+1]
			i++
		case (a == "--gradient" || a == "-g") && i+1 < len(args):
			opts.Gradient = args[i+1]
			i++
		case a == "--from" && i+1 < len(args):
			opts.FromHex = args[i+1]
			i++
		case a == "--to" && i+1 < len(args):
			opts.ToHex = args[i+1]
			i++
		case a == "--bold" || a == "-b":
			opts.Bold = true
		case a == "--dim":
			opts.Dim = true
		case a == "--italic" || a == "-i":
			opts.Italic = true
		case a == "--underline" || a == "-u":
			opts.Underline = true
		default:
			pos = append(pos, a)
		}
	}
	opts.Text = strings.Join(pos, " ")
	return opts
}

func parseBarFlags(args []string) BarOptions {
	var opts BarOptions
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--in-place" || a == "-r":
			opts.InPlace = true
		case (a == "--width" || a == "-w") && i+1 < len(args):
			opts.Width, _ = strconv.Atoi(args[i+1])
			i++
		case (a == "--color" || a == "-c") && i+1 < len(args):
			opts.Color = args[i+1]
			i++
		case (a == "--gradient" || a == "-g") && i+1 < len(args):
			opts.Gradient = args[i+1]
			i++
		default:
			pos = append(pos, a)
		}
	}
	if len(pos) >= 1 {
		opts.Current, _ = strconv.Atoi(pos[0])
	}
	if len(pos) >= 2 {
		opts.Total, _ = strconv.Atoi(pos[1])
	} else {
		opts.Total = 100
	}
	if len(pos) >= 3 {
		opts.Label = strings.Join(pos[2:], " ")
	}
	return opts
}

func parseProgressFlags(args []string) (int, string, string, string) {
	total := 100
	var title, colorSpec, gradSpec string
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--total" && i+1 < len(args):
			total, _ = strconv.Atoi(args[i+1])
			i++
		case (a == "--title" || a == "-t") && i+1 < len(args):
			title = args[i+1]
			i++
		case (a == "--color" || a == "-c") && i+1 < len(args):
			colorSpec = args[i+1]
			i++
		case (a == "--gradient" || a == "-g") && i+1 < len(args):
			gradSpec = args[i+1]
			i++
		default:
			pos = append(pos, a)
		}
	}
	if title == "" && len(pos) > 0 {
		title = strings.Join(pos, " ")
	}
	return total, title, colorSpec, gradSpec
}

func parseSpinFlags(args []string) SpinOptions {
	var opts SpinOptions
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			opts.Command = args[i+1:]
			break
		}
		switch {
		case (a == "--style" || a == "-s") && i+1 < len(args):
			opts.Style = args[i+1]
			i++
		case (a == "--color" || a == "-c") && i+1 < len(args):
			opts.Color = args[i+1]
			i++
		case a == "--persist" || a == "-p":
			opts.Persist = true
		case (a == "--duration" || a == "-d") && i+1 < len(args):
			opts.DurationMs, _ = strconv.Atoi(args[i+1])
			i++
		default:
			pos = append(pos, a)
		}
	}
	opts.Title = strings.Join(pos, " ")
	return opts
}

func renderDemo(out io.Writer) {
	pal := activePalette()
	renderBanner(out, BannerOptions{
		Text:     "ZSPLASH",
		Font:     "block",
		Subtitle: fmt.Sprintf("Reactive Zsh & CLI Text Utility (v%s | Theme: %s)", version, currentThemeName()),
		Frame:    true,
	})
	fmt.Fprintln(out)

	renderStep(out, "1", "5", "Crisp Unicode Block Fonts (Solid by Default)", "", "")
	fmt.Fprintln(out)
	renderBanner(out, BannerOptions{Text: "COMPACT 2-ROW", Font: "compact", Color: "secondary"})
	renderBanner(out, BannerOptions{Text: "LINE FONT", Font: "line", Color: "warn"})
	fmt.Fprintln(out)
	renderCodeBlock(out, "zsh: banners & step headers", []string{
		`zsp banner "ZSPLASH" --font block --frame`,
		`zsp step 1 5 "Crisp Unicode Block Fonts (Solid by Default)"`,
		`zsp banner "COMPACT 2-ROW" --font compact --color secondary`,
		`zsp banner "LINE FONT" --font line --color warn`,
	})
	fmt.Fprintln(out)

	renderStep(out, "2", "5", "Inverted Pill Badges & Metadata", "", "")
	fmt.Fprintln(out)
	renderPillBadge(out, "PASS", pal.Pass, "Database schema verified", "42ms")
	renderPillBadge(out, "INFO", pal.Info, "Worker pool initialized", "8 threads")
	renderPillBadge(out, "WARN", pal.Warn, "Rate limit at 85% capacity", "throttling")
	renderPillBadge(out, "FAIL", pal.Fail, "Upstream webhook timeout", "exit 1")
	fmt.Fprintln(out)
	renderCodeBlock(out, "zsh: badges", []string{
		`zsp ok   "Database schema verified"   --meta "42ms"`,
		`zsp info "Worker pool initialized"    --meta "8 threads"`,
		`zsp warn "Rate limit at 85% capacity" --meta "throttling"`,
		`zsp err  "Upstream webhook timeout"   --meta "exit 1"`,
	})
	fmt.Fprintln(out)

	renderStep(out, "3", "5", "Boxes, Dividers, Key-Value Cards & Trees", "", "")
	fmt.Fprintln(out)
	renderBox(out, BoxOptions{Text: "Deployment complete", Title: "Prod", Border: "double"})
	fmt.Fprintln(out)
	renderRule(out, "SUMMARY", "secondary", "", 0)
	fmt.Fprintln(out)
	renderKV(out, "Release Target", "", "", []string{
		"Binary", "zsplash (zsp)",
		"Active Theme", currentThemeName(),
		"Terminal Width", fmt.Sprintf("%d cols", detectTermWidth()),
	})
	fmt.Fprintln(out)
	renderTree(out, "Preflight Checks", []string{"Config valid", "Migrations clean"}, "")
	fmt.Fprintln(out)
	renderCmd(out, "docker build --tag app:latest .")
	fmt.Fprintln(out)
	renderCodeBlock(out, "zsh: structured layout & dividers", []string{
		`zsp step 3 5 "Boxes, Dividers, Key-Value Cards & Trees"`,
		`zsp box "Deployment complete" --title "Prod" --border double`,
		`zsp rule "SUMMARY" --color secondary`,
		`zsp kv --title "Release Target" "Binary" "zsplash (zsp)" "Active Theme" "` + currentThemeName() + `"`,
		`zsp tree --title "Preflight Checks" "Config valid" "Migrations clean"`,
		`zsp cmd "docker build --tag app:latest ."`,
	})
	fmt.Fprintln(out)

	renderStep(out, "4", "5", "Async Spinners & Half-Cell Progress Bars", "", "")
	fmt.Fprintln(out)
	renderPillBadge(out, "PASS", pal.Pass, "Querying database...", "142ms")
	renderBar(out, BarOptions{Current: 78, Total: 100, Label: "Deploying shards"})
	fmt.Fprintln(out)
	renderCodeBlock(out, "zsh: async & progress", []string{
		`zsp spin "Querying database..." --persist -- psql -c "SELECT 1"`,
		`zsp bar 78 100 "Deploying shards"`,
	})
	fmt.Fprintln(out)

	renderStep(out, "5", "5", "Built-in Color Palettes & Persistent Defaults", "", "")
	fmt.Fprintln(out)
	for _, tName := range []string{"nes", "snes", "halo", "xbox", "dracula", "monokai", "rosepine", "everforest", "synthwave", "industrial"} {
		t := builtinThemes[tName]
		swatches := fmt.Sprintf("%s██%s %s██%s %s██%s %s██%s %s██%s %s██%s",
			t.Primary.fgANSI(), ansiReset(),
			t.Secondary.fgANSI(), ansiReset(),
			t.Warn.fgANSI(), ansiReset(),
			t.Pass.fgANSI(), ansiReset(),
			t.Fail.fgANSI(), ansiReset(),
			t.Muted.fgANSI(), ansiReset(),
		)
		fmt.Fprintf(out, "  %s%s%-11s%s %s  %s%s%s\n",
			ansiBold(), t.Primary.fgANSI(), tName, ansiReset(),
			swatches, ansiDim(), t.Description, ansiReset())
	}
	fmt.Fprintln(out)
	renderCodeBlock(out, "zsh: themes, defaults & docs", []string{
		`zsp themes`,
		`zsp demo --theme nes`,
		`zsp demo --theme xbox`,
		`zsp theme nes`,
		`zsp config set banner.font compact`,
		`zsp docs`,
	})
}

func printHelp(out io.Writer) {
	renderBanner(out, BannerOptions{
		Text: "ZSPLASH",
		Font: "compact",
	})
	fmt.Fprintf(out, "%szsplash (zsp) v%s%s: Reactive CLI & Zsh Text Utility (Active theme: %s)\n\n",
		ansiBold(), version, ansiReset(), currentThemeName())
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  zsp banner <text> [--font block|compact|line] [--color <c>] [--gradient <g>] [--frame]")
	fmt.Fprintln(out, "  zsp box <text> [--title <t>] [--footer <f>] [--border rounded|double|heavy|ascii] [--color <c>]")
	fmt.Fprintln(out, "  zsp ok|warn|err|info <text> [--meta <detail>]")
	fmt.Fprintln(out, "  zsp step <curr> <total> <title> [--color <c>]")
	fmt.Fprintln(out, "  zsp rule [label] [--color <c>]")
	fmt.Fprintln(out, "  zsp kv [--title <t>] <key1> <val1> [<key2> <val2> ...]")
	fmt.Fprintln(out, "  zsp cmd <shell command>")
	fmt.Fprintln(out, "  zsp code [--title <t>] <line1> <line2> ...")
	fmt.Fprintln(out, "  zsp tree [--title <t>] <item1> <item2> ...")
	fmt.Fprintln(out, "  zsp style <text> [--fg <c>] [--bg <c>] [--bold] [--italic] [--underline]")
	fmt.Fprintln(out, "  zsp spin <title> [--style dots|pulse|meter|line] [--persist] -- <cmd> [args...]")
	fmt.Fprintln(out, "  zsp bar <curr> <total> [label] [--in-place|-r] [--color <c>]")
	fmt.Fprintln(out, "  zsp progress [--total <n>] [--title <t>] (reads lines from stdin)")
	fmt.Fprintln(out, "  zsp themes         List all 10 built-in color palettes (nes, snes, dracula, nord, etc.)")
	fmt.Fprintln(out, "  zsp theme <name>   Set your persistent default color palette")
	fmt.Fprintln(out, "  zsp config         Show or override persistent defaults (zsp config set <key> <val>)")
	fmt.Fprintln(out, "  zsp demo           Run full visual showcase (try: zsp demo --theme nes)")
	fmt.Fprintln(out, "  zsp help <cmd>     Show detailed specification, stream rules, and examples for <cmd>")
	fmt.Fprintln(out, "  zsp docs           Output complete Markdown technical reference manual")
	fmt.Fprintln(out, "  zsp docs --json    Output complete machine-readable JSON CLI schema")
	fmt.Fprintln(out, "  zsp --version      Show version info")
}
