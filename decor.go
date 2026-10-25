package main

import (
	"fmt"
	"io"
	"strings"
)

type BorderGlyphs struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
	LeftTee     string
	RightTee    string
}

var borderStyles = map[string]BorderGlyphs{
	"rounded": {
		TopLeft: "╭", TopRight: "╮", BottomLeft: "╰", BottomRight: "╯",
		Horizontal: "─", Vertical: "│", LeftTee: "├", RightTee: "┤",
	},
	"double": {
		TopLeft: "╔", TopRight: "╗", BottomLeft: "╚", BottomRight: "╝",
		Horizontal: "═", Vertical: "║", LeftTee: "╠", RightTee: "╣",
	},
	"heavy": {
		TopLeft: "┏", TopRight: "┓", BottomLeft: "┗", BottomRight: "┛",
		Horizontal: "━", Vertical: "┃", LeftTee: "┣", RightTee: "┫",
	},
	"ascii": {
		TopLeft: "+", TopRight: "+", BottomLeft: "+", BottomRight: "+",
		Horizontal: "-", Vertical: "|", LeftTee: "+", RightTee: "+",
	},
}

type BoxOptions struct {
	Text     string
	Title    string
	Footer   string
	Border   string // "rounded" (default), "double", "heavy", "ascii"
	Color    string
	Gradient string
	FromHex  string
	ToHex    string
	Width    int
}

func renderBox(out io.Writer, opts BoxOptions) {
	termWidth := detectTermWidth()
	maxBoxWidth := termWidth - 2
	if opts.Width > 0 && opts.Width < maxBoxWidth {
		maxBoxWidth = opts.Width
	}
	if maxBoxWidth < 16 {
		maxBoxWidth = 16
	}

	borderName := strings.ToLower(strings.TrimSpace(opts.Border))
	if borderName == "" {
		borderName = activeConfig.BoxBorder
	}
	bStyle, ok := borderStyles[borderName]
	if !ok {
		bStyle = borderStyles["rounded"]
	}

	maxInnerText := maxBoxWidth - 6
	wrappedLines := wrapText(opts.Text, maxInnerText)

	innerWidth := 0
	for _, l := range wrappedLines {
		if visibleLen(l) > innerWidth {
			innerWidth = visibleLen(l)
		}
	}
	if opts.Title != "" && visibleLen(opts.Title)+4 > innerWidth {
		innerWidth = visibleLen(opts.Title) + 4
	}
	if opts.Footer != "" && visibleLen(opts.Footer)+4 > innerWidth {
		innerWidth = visibleLen(opts.Footer) + 4
	}
	if opts.Width > 0 && opts.Width-6 > innerWidth && opts.Width-6 <= maxInnerText {
		innerWidth = opts.Width - 6
	}
	if innerWidth > maxInnerText {
		innerWidth = maxInnerText
	}

	pal := activePalette()
	colorSpec := opts.Color
	if colorSpec == "" {
		colorSpec = activeConfig.BoxColor
	}
	grad, useGrad := resolveGradient(opts.Gradient, opts.FromHex, opts.ToHex)
	c := parseColor(colorSpec, pal.Primary)

	totalH := innerWidth + 4
	var topLine string
	if opts.Title != "" {
		titleRunes := truncateVisible(opts.Title, innerWidth)
		rem := totalH - (visibleLen(titleRunes) + 3)
		if rem < 1 {
			rem = 1
		}
		topLine = bStyle.TopLeft + bStyle.Horizontal + " " + titleRunes + " " + strings.Repeat(bStyle.Horizontal, rem) + bStyle.TopRight
	} else {
		topLine = bStyle.TopLeft + strings.Repeat(bStyle.Horizontal, totalH) + bStyle.TopRight
	}

	var botLine string
	if opts.Footer != "" {
		footerRunes := truncateVisible(opts.Footer, innerWidth)
		rem := totalH - (visibleLen(footerRunes) + 3)
		if rem < 1 {
			rem = 1
		}
		botLine = bStyle.BottomLeft + strings.Repeat(bStyle.Horizontal, rem) + " " + footerRunes + " " + bStyle.Horizontal + bStyle.BottomRight
	} else {
		botLine = bStyle.BottomLeft + strings.Repeat(bStyle.Horizontal, totalH) + bStyle.BottomRight
	}

	if useGrad {
		fmt.Fprintln(out, applyGradientLine(topLine, grad, true))
	} else {
		fmt.Fprintf(out, "%s%s%s\n", c.fgANSI(), topLine, ansiReset())
	}

	leftColor := c
	rightColor := c
	if useGrad {
		leftColor = grad.At(0.0)
		rightColor = grad.At(1.0)
	}

	for _, l := range wrappedLines {
		padded := padRight(l, innerWidth)
		fmt.Fprintf(out, "%s%s%s  %s%s%s  %s%s%s\n",
			leftColor.fgANSI(), bStyle.Vertical, ansiReset(),
			ansiBold(), padded, ansiReset(),
			rightColor.fgANSI(), bStyle.Vertical, ansiReset())
	}

	if useGrad {
		fmt.Fprintln(out, applyGradientLine(botLine, grad, false))
	} else {
		fmt.Fprintf(out, "%s%s%s\n", c.fgANSI(), botLine, ansiReset())
	}
}

// renderPillBadge renders an inverted TrueColor pill badge with message and optional right-aligned metadata.
func renderPillBadge(out io.Writer, label string, pillColor RGB, text string, meta string) {
	termWidth := detectTermWidth()
	cleanLabel := strings.TrimSpace(label)

	var pill string
	if noColorEnabled() {
		pill = fmt.Sprintf("[%s]", cleanLabel)
	} else {
		darkText := "\033[38;2;15;20;28m"
		pill = fmt.Sprintf("%s%s%s %s %s", ansiBold(), pillColor.bgANSI(), darkText, cleanLabel, ansiReset())
	}

	if meta == "" {
		fmt.Fprintf(out, "%s %s\n", pill, text)
		return
	}

	leftVisible := len([]rune(cleanLabel)) + 2 + 1 + visibleLen(text)
	metaVisible := visibleLen(meta)
	targetWidth := termWidth - 2
	if targetWidth > 76 {
		targetWidth = 76
	}

	if leftVisible+2+metaVisible <= targetWidth {
		dotsCount := targetWidth - leftVisible - metaVisible - 2
		if dotsCount < 1 {
			dotsCount = 1
		}
		dots := strings.Repeat("·", dotsCount)
		fmt.Fprintf(out, "%s %s %s%s %s%s\n", pill, text, ansiDim(), dots, meta, ansiReset())
	} else {
		fmt.Fprintf(out, "%s %s %s(%s)%s\n", pill, text, ansiDim(), meta, ansiReset())
	}
}

// renderStep renders a terminal-width phase divider for multi-step Zsh scripts.
func renderStep(out io.Writer, current, total, title, colorSpec, gradSpec string) {
	termWidth := detectTermWidth()
	barWidth := termWidth - 2
	if barWidth > 76 {
		barWidth = 76
	}
	if barWidth < 24 {
		barWidth = 24
	}

	tag := fmt.Sprintf("[%s/%s]", current, total)
	rem := barWidth - (visibleLen(tag) + 5)
	if rem < 4 {
		rem = 4
	}
	ruleLine := "━━━ " + tag + " " + strings.Repeat("━", rem)

	pal := activePalette()
	if colorSpec == "" {
		colorSpec = activeConfig.StepColor
	}
	grad, useGrad := resolveGradient(gradSpec, "", "")
	c := parseColor(colorSpec, pal.Primary)

	if useGrad {
		fmt.Fprintln(out, applyGradientLine(ruleLine, grad, true))
	} else {
		fmt.Fprintf(out, "%s%s%s%s\n", c.fgANSI(), ansiBold(), ruleLine, ansiReset())
	}

	if title != "" {
		arrowColor := pal.Secondary
		if colorSpec != "" {
			arrowColor = c
		}
		if useGrad {
			arrowColor = grad.At(0.0)
		}
		fmt.Fprintf(out, "  %s▸%s %s%s%s\n", arrowColor.fgANSI(), ansiReset(), ansiBold(), title, ansiReset())
	}
}

// renderRule renders a horizontal divider across the terminal width.
func renderRule(out io.Writer, label, colorSpec, gradSpec string, width int) {
	termWidth := detectTermWidth()
	target := termWidth - 2
	if width > 0 && width < target {
		target = width
	}
	if target < 16 {
		target = 16
	}

	var line string
	clean := strings.TrimSpace(label)
	if clean == "" {
		line = strings.Repeat("─", target)
	} else {
		clean = truncateVisible(clean, target-6)
		padTotal := target - (visibleLen(clean) + 4)
		if padTotal < 2 {
			padTotal = 2
		}
		left := padTotal / 2
		right := padTotal - left
		line = strings.Repeat("─", left) + "  " + clean + "  " + strings.Repeat("─", right)
	}

	pal := activePalette()
	if colorSpec == "" {
		colorSpec = activeConfig.RuleColor
	}
	grad, useGrad := resolveGradient(gradSpec, "", "")
	c := parseColor(colorSpec, pal.Primary)
	if useGrad {
		fmt.Fprintln(out, applyGradientLine(line, grad, true))
	} else {
		fmt.Fprintf(out, "%s%s%s\n", c.fgANSI(), line, ansiReset())
	}
}

// renderKV renders aligned key-value pairs with dot-leaders inside a rounded card.
func renderKV(out io.Writer, title, colorSpec, gradSpec string, pairs []string) {
	if len(pairs) == 0 {
		return
	}
	termWidth := detectTermWidth()
	maxCardWidth := termWidth - 2
	if maxCardWidth > 68 {
		maxCardWidth = 68
	}
	if maxCardWidth < 24 {
		maxCardWidth = 24
	}

	type kvPair struct {
		k, v string
	}
	var items []kvPair
	maxKey := 0
	maxVal := 0
	for i := 0; i < len(pairs); i += 2 {
		k := pairs[i]
		v := ""
		if i+1 < len(pairs) {
			v = pairs[i+1]
		}
		if visibleLen(k) > maxKey {
			maxKey = visibleLen(k)
		}
		if visibleLen(v) > maxVal {
			maxVal = visibleLen(v)
		}
		items = append(items, kvPair{k, v})
	}

	innerWidth := maxKey + maxVal + 8
	if title != "" && visibleLen(title)+6 > innerWidth {
		innerWidth = visibleLen(title) + 6
	}
	if innerWidth < 28 {
		innerWidth = 28
	}
	if innerWidth > maxCardWidth-6 {
		innerWidth = maxCardWidth - 6
	}

	pal := activePalette()
	if colorSpec == "" {
		colorSpec = activeConfig.KVColor
	}
	c := parseColor(colorSpec, pal.Secondary)
	var formattedLines []string
	for _, it := range items {
		dots := innerWidth - visibleLen(it.k) - visibleLen(it.v) - 2
		if dots < 2 {
			dots = 2
		}
		leader := strings.Repeat(".", dots)
		line := fmt.Sprintf("%s%s%s %s%s%s %s%s%s",
			c.fgANSI(), it.k, ansiReset(),
			ansiDim(), leader, ansiReset(),
			ansiBold()+pal.Text.fgANSI(), it.v, ansiReset())
		formattedLines = append(formattedLines, line)
	}

	renderBox(out, BoxOptions{
		Text:     strings.Join(formattedLines, "\n"),
		Title:    title,
		Border:   activeConfig.BoxBorder,
		Color:    colorSpec,
		Gradient: gradSpec,
	})
}

// styleCommandTokens returns a syntax-highlighted command line string while preserving exact spacing.
func styleCommandTokens(cmdText string) string {
	pal := activePalette()
	var out strings.Builder
	var tok strings.Builder
	var quote rune
	tokIdx := 0

	flushToken := func() {
		if tok.Len() == 0 {
			return
		}
		t := tok.String()
		tok.Reset()
		switch {
		case tokIdx == 0:
			out.WriteString(ansiBold() + pal.Pass.fgANSI() + t + ansiReset())
		case tokIdx == 1 && !strings.HasPrefix(t, "-") && !strings.HasPrefix(t, "\""):
			out.WriteString(ansiBold() + pal.Secondary.fgANSI() + t + ansiReset())
		case strings.HasPrefix(t, "-"):
			out.WriteString(pal.Warn.fgANSI() + t + ansiReset())
		case strings.HasPrefix(t, "\"") || strings.HasPrefix(t, "'"):
			out.WriteString(pal.Text.fgANSI() + t + ansiReset())
		default:
			out.WriteString(t)
		}
		tokIdx++
	}

	for _, r := range cmdText {
		switch {
		case quote != 0:
			tok.WriteRune(r)
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
			tok.WriteRune(r)
		case r == ' ' || r == '\t':
			flushToken()
			out.WriteRune(r)
		default:
			tok.WriteRune(r)
		}
	}
	flushToken()
	return out.String()
}

// renderCmd formats a single shell command preview with a left gutter and syntax-highlighted tokens.
func renderCmd(out io.Writer, cmdText string) {
	pal := activePalette()
	fmt.Fprintf(out, "  %s┃%s %s▸%s %s\n",
		pal.Primary.fgANSI(), ansiReset(),
		pal.Secondary.fgANSI(), ansiReset(),
		styleCommandTokens(cmdText))
}

// renderCodeBlock renders a framed multi-line Zsh code block below a group of visual effects.
func renderCodeBlock(out io.Writer, title string, lines []string) {
	if len(lines) == 0 {
		return
	}
	termWidth := detectTermWidth()
	maxBoxWidth := termWidth - 4
	if maxBoxWidth < 20 {
		maxBoxWidth = 20
	}
	maxInner := maxBoxWidth - 6

	if title == "" {
		title = "zsh"
	}

	innerWidth := visibleLen(title) + 4
	var styledLines []string
	for _, l := range lines {
		clipped := truncateVisible(l, maxInner)
		styled := styleCommandTokens(clipped)
		styledLines = append(styledLines, styled)
		if visibleLen(styled) > innerWidth {
			innerWidth = visibleLen(styled)
		}
	}
	if innerWidth > maxInner {
		innerWidth = maxInner
	}

	borderColor := activePalette().Muted
	totalH := innerWidth + 4
	rem := totalH - (visibleLen(title) + 3)
	if rem < 1 {
		rem = 1
	}
	topLine := "  ╭─ " + title + " " + strings.Repeat("─", rem) + "╮"
	botLine := "  ╰" + strings.Repeat("─", totalH) + "╯"

	fmt.Fprintf(out, "%s%s%s\n", borderColor.fgANSI(), topLine, ansiReset())
	for _, styled := range styledLines {
		padCount := innerWidth - visibleLen(styled)
		if padCount < 0 {
			padCount = 0
		}
		pad := strings.Repeat(" ", padCount)
		fmt.Fprintf(out, "%s  │%s  %s%s  %s│%s\n",
			borderColor.fgANSI(), ansiReset(),
			styled, pad,
			borderColor.fgANSI(), ansiReset())
	}
	fmt.Fprintf(out, "%s%s%s\n", borderColor.fgANSI(), botLine, ansiReset())
}

// renderTree renders a clean Unicode tree/checklist from a list of items.
func renderTree(out io.Writer, title string, items []string, colorSpec string) {
	pal := activePalette()
	c := parseColor(colorSpec, pal.Primary)
	okColor := pal.Pass
	if title != "" {
		fmt.Fprintf(out, "%s%s▸ %s%s\n", ansiBold(), c.fgANSI(), title, ansiReset())
	}
	for i, item := range items {
		branch := "├─"
		if i == len(items)-1 {
			branch = "╰─"
		}
		fmt.Fprintf(out, "  %s%s%s %s●%s %s\n",
			ansiDim(), branch, ansiReset(),
			okColor.fgANSI(), ansiReset(),
			item)
	}
}

type StyleOptions struct {
	Text      string
	Fg        string
	Bg        string
	Gradient  string
	FromHex   string
	ToHex     string
	Bold      bool
	Dim       bool
	Italic    bool
	Underline bool
}

func renderStyle(out io.Writer, opts StyleOptions) {
	if grad, ok := resolveGradient(opts.Gradient, opts.FromHex, opts.ToHex); ok {
		fmt.Fprintln(out, applyGradientLine(opts.Text, grad, opts.Bold))
		return
	}
	if noColorEnabled() {
		fmt.Fprintln(out, opts.Text)
		return
	}
	var prefix strings.Builder
	if opts.Bold {
		prefix.WriteString(bold)
	}
	if opts.Dim {
		prefix.WriteString(dim)
	}
	if opts.Italic {
		prefix.WriteString(italic)
	}
	if opts.Underline {
		prefix.WriteString(underline)
	}
	if opts.Fg != "" {
		prefix.WriteString(parseColor(opts.Fg, namedColors["cyan"]).fgANSI())
	}
	if opts.Bg != "" {
		prefix.WriteString(parseColor(opts.Bg, namedColors["purple"]).bgANSI())
	}
	fmt.Fprintf(out, "%s%s%s\n", prefix.String(), opts.Text, ansiReset())
}
