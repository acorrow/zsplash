package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var version = "0.2.0"

func main() {
	code := runWithExit(os.Args[1:], os.Stdout, os.Stderr)
	if code != 0 {
		os.Exit(code)
	}
}

func run(args []string, out io.Writer) {
	_ = runWithExit(args, out, out)
}

func runWithExit(args []string, out io.Writer, errOut io.Writer) int {
	if len(args) < 1 {
		printHelp(out)
		return 0
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "-v", "--version", "version":
		fmt.Fprintf(out, "zsplash (zsp) v%s\n", version)
		return 0

	case "-h", "--help", "help":
		printHelp(out)
		return 0

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
		renderPillBadge(out, "PASS", namedColors["green"], msg, meta)
		return 0

	case "warn", "warning":
		msg, meta := parseBadgeFlags(rest)
		renderPillBadge(out, "WARN", namedColors["yellow"], msg, meta)
		return 0

	case "err", "error", "fail":
		msg, meta := parseBadgeFlags(rest)
		renderPillBadge(out, "FAIL", namedColors["red"], msg, meta)
		return 0

	case "info":
		msg, meta := parseBadgeFlags(rest)
		renderPillBadge(out, "INFO", namedColors["blue"], msg, meta)
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
			Text:     strings.Join(args, " "),
			Border:   "rounded",
			Gradient: "cyber",
		})
		return 0
	}
}

func parseBannerFlags(args []string) BannerOptions {
	opts := BannerOptions{Font: "block"}
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
	opts := BoxOptions{Border: "rounded"}
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
	opts := SpinOptions{Style: "dots"}
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
	renderBanner(out, BannerOptions{
		Text:     "ZSPLASH",
		Font:     "block",
		Gradient: "cyber",
		Subtitle: "Reactive Zsh & CLI Text Decoration Utility (v" + version + ")",
		Frame:    true,
	})
	fmt.Fprintln(out)

	renderStep(out, "1", "4", "Crisp Unicode Block Fonts (compact, block, line)", "", "cyber")
	renderBanner(out, BannerOptions{Text: "COMPACT 2-ROW", Font: "compact", Gradient: "ember"})
	renderBanner(out, BannerOptions{Text: "LINE FONT", Font: "line", Gradient: "matrix"})

	renderStep(out, "2", "4", "Inverted Pill Badges & Metadata", "", "cyber")
	renderPillBadge(out, "PASS", namedColors["green"], "Database schema verified", "42ms")
	renderPillBadge(out, "INFO", namedColors["blue"], "Worker pool initialized", "8 threads")
	renderPillBadge(out, "WARN", namedColors["yellow"], "Rate limit at 85% capacity", "throttling")
	renderPillBadge(out, "FAIL", namedColors["red"], "Upstream webhook timeout", "exit 1")

	renderStep(out, "3", "4", "Key-Value Summary Cards & Command Previews", "", "cyber")
	renderKV(out, "Release Target", "", "cyber", []string{
		"Binary", "zsplash (zsp)",
		"Branch", "develop",
		"Terminal Width", fmt.Sprintf("%d cols", detectTermWidth()),
	})
	renderCmd(out, "brew upgrade --fetch-HEAD zsplash")

	renderStep(out, "4", "4", "Half-Cell Unicode Progress Bars", "", "cyber")
	renderBar(out, BarOptions{Current: 78, Total: 100, Label: "Deploying shards", Gradient: "cyber"})
}

func printHelp(out io.Writer) {
	renderBanner(out, BannerOptions{
		Text:     "ZSPLASH",
		Font:     "compact",
		Gradient: "cyber",
	})
	fmt.Fprintf(out, "%szsplash (zsp) v%s%s: Reactive CLI & Zsh Text Utility\n\n", ansiBold(), version, ansiReset())
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  zsp banner <text> [--font block|compact|line] [--gradient cyber|ember|ocean|matrix|gold] [--frame]")
	fmt.Fprintln(out, "  zsp box <text> [--title <t>] [--footer <f>] [--border rounded|double|heavy|ascii] [--gradient <g>]")
	fmt.Fprintln(out, "  zsp ok|warn|err|info <text> [--meta <detail>]")
	fmt.Fprintln(out, "  zsp step <curr> <total> <title> [--gradient <g>]")
	fmt.Fprintln(out, "  zsp rule [label] [--gradient <g>] [--color <c>]")
	fmt.Fprintln(out, "  zsp kv [--title <t>] <key1> <val1> [<key2> <val2> ...]")
	fmt.Fprintln(out, "  zsp cmd <shell command>")
	fmt.Fprintln(out, "  zsp tree [--title <t>] <item1> <item2> ...")
	fmt.Fprintln(out, "  zsp style <text> [--fg <c>] [--bg <c>] [--gradient <g>] [--bold] [--italic] [--underline]")
	fmt.Fprintln(out, "  zsp spin <title> [--style dots|pulse|meter|line] [--persist] -- <cmd> [args...]")
	fmt.Fprintln(out, "  zsp bar <curr> <total> [label] [--in-place|-r] [--gradient <g>]")
	fmt.Fprintln(out, "  zsp progress [--total <n>] [--title <t>] (reads lines from stdin)")
	fmt.Fprintln(out, "  zsp demo           Run full visual showcase")
	fmt.Fprintln(out, "  zsp --version      Show version info")
}
