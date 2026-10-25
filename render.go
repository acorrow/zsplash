package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var compactGlyphs = map[rune][2]string{
	'A': {"▄▀▄", "█▀█"}, 'B': {"██▄", "██▀"}, 'C': {"▄▀▀", "▀▄▄"}, 'D': {"█▀▄", "█▄▀"},
	'E': {"█▀▀", "██▄"}, 'F': {"█▀▀", "█▀ "}, 'G': {"▄▀▀", "▀▄█"}, 'H': {"█ █", "█▀█"},
	'I': {"▀█▀", "▄█▄"}, 'J': {"  █", "▀▄▀"}, 'K': {"█▄▀", "█ █"}, 'L': {"█  ", "█▄▄"},
	'M': {"█▄█", "█ █"}, 'N': {"█▀█", "█ █"}, 'O': {"▄▀▄", "▀▄▀"}, 'P': {"█▀█", "█▀▀"},
	'Q': {"▄▀▄", "▀▄█"}, 'R': {"█▀█", "█▀▄"}, 'S': {"▄▀▀", "▄▄▀"}, 'T': {"▀█▀", " █ "},
	'U': {"█ █", "▀▄▀"}, 'V': {"█ █", "▀▄▀"}, 'W': {"█ █", "█▀█"}, 'X': {"▀▄▀", "█ █"},
	'Y': {"█ █", " █ "}, 'Z': {"▀▀█", "█▄▄"}, '0': {"▄▀▄", "▀▄▀"}, '1': {"▄█ ", " █ "},
	'2': {"▀▀█", "█▄▄"}, '3': {"▀▀█", "▄▄█"}, '4': {"█ █", "▀▀█"}, '5': {"█▀▀", "▄▄▀"},
	'6': {"▄▀▀", "█▄█"}, '7': {"▀▀█", "  █"}, '8': {"▄▀▄", "█▄█"}, '9': {"█▀█", "▀▀█"},
	'-': {"   ", "▀▀▀"}, '_': {"   ", "▄▄▄"}, '.': {"   ", " ▄ "}, ':': {" ▄ ", " ▄ "},
	'!': {" █ ", " ▄ "}, '/': {" ▄▀", "▀  "}, ' ': {"  ", "  "},
}

var blockGlyphs = map[rune][3]string{
	'A': {"▄▀▀▄", "█▀▀█", "▀  ▀"}, 'B': {"██▀▄", "██▀▄", "▀▀▀ "}, 'C': {"▄▀▀▀", "█   ", "▀▀▀▀"},
	'D': {"█▀▀▄", "█  █", "▀▀▀ "}, 'E': {"█▀▀▀", "█▀▀ ", "▀▀▀▀"}, 'F': {"█▀▀▀", "█▀▀ ", "▀   "},
	'G': {"▄▀▀▀", "█ ▀█", "▀▀▀▀"}, 'H': {"█  █", "█▀▀█", "▀  ▀"}, 'I': {"▀█▀", " █ ", "▀▀▀"},
	'J': {"   █", "   █", "▀▀▀ "}, 'K': {"█ ▄▀", "█▀▄ ", "▀  ▀"}, 'L': {"█   ", "█   ", "▀▀▀▀"},
	'M': {"█▀▄▀█", "█ ▀ █", "▀   ▀"}, 'N': {"█▀▄ █", "█ ▀▄█", "▀   ▀"}, 'O': {"▄▀▀▄", "█  █", "▀▀▀▀"},
	'P': {"█▀▀▄", "█▀▀ ", "▀   "}, 'Q': {"▄▀▀▄", "█ ▄█", "▀▀▀▄"}, 'R': {"█▀▀▄", "█▀▀▄", "▀  ▀"},
	'S': {"▄▀▀▀", "▀▀▀▄", "▀▀▀ "}, 'T': {"▀▀█▀▀", "  █  ", "  ▀  "}, 'U': {"█  █", "█  █", "▀▀▀▀"},
	'V': {"█  █", "▀▄▄▀", " ▀▀ "}, 'W': {"█   █", "█ ▄ █", "▀▄▀▄▀"}, 'X': {"▀▄ ▄▀", "  █  ", "▄▀ ▀▄"},
	'Y': {"█  █", " ▀▀█", "▀▀▀ "}, 'Z': {"▀▀▀█", " ▄▀ ", "▀▀▀▀"}, '0': {"▄▀▀▄", "█  █", "▀▀▀▀"},
	'1': {"▄█ ", " █ ", "▀▀▀"}, '2': {"▀▀▀▄", "▄▀▀ ", "▀▀▀▀"}, '3': {"▀▀▀▄", " ▀▀▄", "▀▀▀ "},
	'4': {"█  █", "▀▀▀█", "   ▀"}, '5': {"█▀▀▀", "▀▀▀▄", "▀▀▀ "}, '6': {"▄▀▀▀", "█▀▀▄", "▀▀▀▀"},
	'7': {"▀▀▀█", "  ▄▀", " ▀  "}, '8': {"▄▀▀▄", "▄▀▀▄", "▀▀▀▀"}, '9': {"▄▀▀▄", "▀▀▀█", "▀▀▀ "},
	'-': {"    ", "▀▀▀▀", "    "}, '_': {"    ", "    ", "▀▀▀▀"}, '.': {"  ", "  ", "▀ "},
	':': {"▄ ", "  ", "▀ "}, '!': {"█ ", "█ ", "▀ "}, '/': {"  ▄▀", " ▄▀ ", "▀   "},
	' ': {"   ", "   ", "   "},
}

var lineGlyphs = map[rune][3]string{
	'A': {"╭─╮", "├─┤", "╵ ╵"}, 'B': {"├─╮", "├─┤", "╰─╯"}, 'C': {"╭─╴", "│  ", "╰─╴"},
	'D': {"├─╮", "│ │", "╰─╯"}, 'E': {"╭─╴", "├─╴", "╰─╴"}, 'F': {"╭─╴", "├─╴", "╵  "},
	'G': {"╭─╴", "│ ╮", "╰─╯"}, 'H': {"╷ ╷", "├─┤", "╵ ╵"}, 'I': {"╶┬╴", " │ ", "╶┴╴"},
	'J': {"  ╷", "  │", "╰─╯"}, 'K': {"╷ ╭", "├─┤", "╵ ╰"}, 'L': {"╷  ", "│  ", "╰─╴"},
	'M': {"╭┬╮", "│││", "╵╵╵"}, 'N': {"╭─╮", "│ │", "╵ ╵"}, 'O': {"╭─╮", "│ │", "╰─╯"},
	'P': {"╭─╮", "├─╯", "╵  "}, 'Q': {"╭─╮", "│ ┼", "╰─╯"}, 'R': {"╭─╮", "├─┬", "╵ ╰"},
	'S': {"╭─╴", "╰─╮", "╶─╯"}, 'T': {"╶┬╴", " │ ", " ╵ "}, 'U': {"╷ ╷", "│ │", "╰─╯"},
	'V': {"╷ ╷", "│ │", "╰─╯"}, 'W': {"╷ ╷", "│││", "╰┴╯"}, 'X': {"╲ ╱", " ╳ ", "╱ ╲"},
	'Y': {"╷ ╷", "╰┬╯", " ╵ "}, 'Z': {"╶─╮", " ╭╯", "╰─╴"}, '0': {"╭─╮", "│ │", "╰─╯"},
	'1': {"╶┐ ", " │ ", "╶┴╴"}, '2': {"╶─╮", "╭─╯", "╰─╴"}, '3': {"╶─╮", " ─┤", "╶─╯"},
	'4': {"╷ ╷", "╰─┤", "  ╵"}, '5': {"╭─╴", "╰─╮", "╶─╯"}, '6': {"╭─╴", "├─╮", "╰─╯"},
	'7': {"╶─╮", "  │", "  ╵"}, '8': {"╭─╮", "├─┤", "╰─╯"}, '9': {"╭─╮", "╰─┤", "╶─╯"},
	'-': {"   ", "───", "   "}, '_': {"   ", "   ", "───"}, '.': {" ", " ", "∙"},
	':': {"∙", " ", "∙"}, '!': {"│", "│", "∙"}, '/': {"  ╱", " ╱ ", "╱  "},
	' ': {"  ", "  ", "  "},
}

type BannerOptions struct {
	Text, Font, Gradient, FromHex, ToHex, Color, Subtitle string
	Frame                                                 bool
}

func renderBanner(out io.Writer, opts BannerOptions) {
	opts.Subtitle = sanitizeText(opts.Subtitle, false)
	text := strings.ToUpper(strings.TrimSpace(sanitizeText(opts.Text, false)))
	if text == "" {
		text = "ZSPLASH"
	}
	termWidth := detectTermWidth()
	font := strings.ToLower(strings.TrimSpace(opts.Font))
	if font == "" {
		font = activeConfig.BannerFont
	}
	if font == "" {
		font = "block"
	}
	pal := activePalette()
	colorSpec := opts.Color
	if colorSpec == "" {
		colorSpec = activeConfig.BannerColor
	}
	grad, useGrad := resolveGradient(opts.Gradient, opts.FromHex, opts.ToHex)
	solidColor := parseColor(colorSpec, pal.Primary)

	words := strings.Fields(text)
	if len(words) == 0 {
		words = []string{text}
	}
	var chunks []string
	curr := words[0]
	for _, w := range words[1:] {
		if rows := buildFontRows(curr+" "+w, font); len(rows) > 0 && visibleLen(rows[0]) <= termWidth-4 {
			curr += " " + w
		} else {
			chunks = append(chunks, curr)
			curr = w
		}
	}
	chunks = append(chunks, curr)

	var allRows []string
	for idx, ch := range chunks {
		allRows = append(allRows, buildFontRows(ch, font)...)
		if idx < len(chunks)-1 {
			allRows = append(allRows, "")
		}
	}

	isLineFont := font == "line" || font == "thin"
	boldSeq := ansiBold()
	if isLineFont {
		boldSeq = ""
	}

	if opts.Frame {
		maxW := 0
		for _, r := range allRows {
			if visibleLen(r) > maxW {
				maxW = visibleLen(r)
			}
		}
		if opts.Subtitle != "" && visibleLen(opts.Subtitle) > maxW {
			maxW = visibleLen(opts.Subtitle)
		}
		if maxW+6 > termWidth {
			maxW = termWidth - 6
		}
		topBorder := "╭" + strings.Repeat("─", maxW+4) + "╮"
		botBorder := "╰" + strings.Repeat("─", maxW+4) + "╯"
		if useGrad {
			fmt.Fprintln(out, applyGradientLine(topBorder, grad, false))
		} else {
			fmt.Fprintf(out, "%s%s%s\n", solidColor.fgANSI(), topBorder, ansiReset())
		}
		bColor, rColor := solidColor, solidColor
		if useGrad {
			bColor, rColor = grad.At(0.0), grad.At(1.0)
		}
		for _, r := range allRows {
			padded := padRight(r, maxW)
			inner := solidColor.fgANSI() + boldSeq + padded + ansiReset()
			if useGrad {
				inner = applyGradientLine(padded, grad, !isLineFont)
			}
			fmt.Fprintf(out, "%s│%s  %s  %s│%s\n", bColor.fgANSI(), ansiReset(), inner, rColor.fgANSI(), ansiReset())
		}
		if opts.Subtitle != "" {
			sep := "├" + strings.Repeat("─", maxW+4) + "┤"
			if useGrad {
				fmt.Fprintln(out, applyGradientLine(sep, grad, false))
			} else {
				fmt.Fprintf(out, "%s%s%s\n", solidColor.fgANSI(), sep, ansiReset())
			}
			sub := padRight(truncateVisible(opts.Subtitle, maxW), maxW)
			fmt.Fprintf(out, "%s│%s  %s%s%s  %s│%s\n", bColor.fgANSI(), ansiReset(), ansiDim(), sub, ansiReset(), rColor.fgANSI(), ansiReset())
		}
		if useGrad {
			fmt.Fprintln(out, applyGradientLine(botBorder, grad, false))
		} else {
			fmt.Fprintf(out, "%s%s%s\n", solidColor.fgANSI(), botBorder, ansiReset())
		}
		return
	}

	for _, r := range allRows {
		if useGrad {
			fmt.Fprintln(out, applyGradientLine(r, grad, !isLineFont))
		} else {
			fmt.Fprintf(out, "%s%s%s%s\n", solidColor.fgANSI(), boldSeq, r, ansiReset())
		}
	}
	if opts.Subtitle != "" {
		fmt.Fprintf(out, "%s%s%s\n", ansiDim(), opts.Subtitle, ansiReset())
	}
}

func buildFontRows(text, font string) []string {
	runes := []rune(text)
	switch font {
	case "compact", "mini", "small":
		var r0, r1 strings.Builder
		for i, ch := range runes {
			g, ok := compactGlyphs[ch]
			if !ok {
				g = [2]string{"▄▀▄", " ▄ "}
			}
			r0.WriteString(g[0])
			r1.WriteString(g[1])
			if i < len(runes)-1 {
				r0.WriteString(" ")
				r1.WriteString(" ")
			}
		}
		return []string{r0.String(), r1.String()}
	case "line", "thin":
		var r0, r1, r2 strings.Builder
		for i, ch := range runes {
			g, ok := lineGlyphs[ch]
			if !ok {
				g = [3]string{"╭─╮", " ╭╯", " ∙ "}
			}
			r0.WriteString(g[0])
			r1.WriteString(g[1])
			r2.WriteString(g[2])
			if i < len(runes)-1 {
				r0.WriteString(" ")
				r1.WriteString(" ")
				r2.WriteString(" ")
			}
		}
		return []string{r0.String(), r1.String(), r2.String()}
	default:
		var r0, r1, r2 strings.Builder
		for i, ch := range runes {
			g, ok := blockGlyphs[ch]
			if !ok {
				g = [3]string{"▄▀▀▄", "  ▄▀", " ▀  "}
			}
			r0.WriteString(g[0])
			r1.WriteString(g[1])
			r2.WriteString(g[2])
			if i < len(runes)-1 {
				r0.WriteString(" ")
				r1.WriteString(" ")
				r2.WriteString(" ")
			}
		}
		return []string{r0.String(), r1.String(), r2.String()}
	}
}

type BorderGlyphs struct {
	TopLeft, TopRight, BottomLeft, BottomRight  string
	Horizontal, Vertical                        string
	LeftTee, RightTee, TopTee, BottomTee, Cross string
}

var borderStyles = map[string]BorderGlyphs{
	"rounded": {"╭", "╮", "╰", "╯", "─", "│", "├", "┤", "┬", "┴", "┼"},
	"double":  {"╔", "╗", "╚", "╝", "═", "║", "╠", "╣", "╦", "╩", "╬"},
	"heavy":   {"┏", "┓", "┗", "┛", "━", "┃", "┣", "┫", "┳", "┻", "╋"},
	"ascii":   {"+", "+", "+", "+", "-", "|", "+", "+", "+", "+", "+"},
}

type BoxOptions struct {
	Text, Title, Footer, Border, Color, Gradient, FromHex, ToHex string
	Width                                                        int
	preformatted                                                 bool
}

func renderBox(out io.Writer, opts BoxOptions) {
	opts.Title = sanitizeText(opts.Title, false)
	opts.Footer = sanitizeText(opts.Footer, false)
	if !opts.preformatted {
		opts.Text = sanitizeText(opts.Text, true)
	}
	if opts.Width > maxTermCols {
		opts.Width = maxTermCols
	}
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
	topLine := bStyle.TopLeft + strings.Repeat(bStyle.Horizontal, totalH) + bStyle.TopRight
	if opts.Title != "" {
		tRunes := truncateVisible(opts.Title, innerWidth)
		rem := totalH - (visibleLen(tRunes) + 3)
		if rem < 1 {
			rem = 1
		}
		topLine = bStyle.TopLeft + bStyle.Horizontal + " " + tRunes + " " + strings.Repeat(bStyle.Horizontal, rem) + bStyle.TopRight
	}
	botLine := bStyle.BottomLeft + strings.Repeat(bStyle.Horizontal, totalH) + bStyle.BottomRight
	if opts.Footer != "" {
		fRunes := truncateVisible(opts.Footer, innerWidth)
		rem := totalH - (visibleLen(fRunes) + 3)
		if rem < 1 {
			rem = 1
		}
		botLine = bStyle.BottomLeft + strings.Repeat(bStyle.Horizontal, rem) + " " + fRunes + " " + bStyle.Horizontal + bStyle.BottomRight
	}

	if useGrad {
		fmt.Fprintln(out, applyGradientLine(topLine, grad, false))
	} else {
		fmt.Fprintf(out, "%s%s%s\n", c.fgANSI(), topLine, ansiReset())
	}
	lColor, rColor := c, c
	if useGrad {
		lColor, rColor = grad.At(0.0), grad.At(1.0)
	}
	for _, l := range wrappedLines {
		fmt.Fprintf(out, "%s%s%s  %s%s%s  %s%s%s\n",
			lColor.fgANSI(), bStyle.Vertical, ansiReset(),
			ansiBold(), padRight(l, innerWidth), ansiReset(),
			rColor.fgANSI(), bStyle.Vertical, ansiReset())
	}
	if useGrad {
		fmt.Fprintln(out, applyGradientLine(botLine, grad, false))
	} else {
		fmt.Fprintf(out, "%s%s%s\n", c.fgANSI(), botLine, ansiReset())
	}
}

func renderPillBadge(out io.Writer, label string, pillColor RGB, text string, meta string) {
	termWidth := detectTermWidth()
	cleanLabel := strings.TrimSpace(sanitizeText(label, false))
	text = sanitizeText(text, false)
	meta = sanitizeText(meta, false)
	pill := fmt.Sprintf("[%s]", cleanLabel)
	if !noColorEnabled() {
		pill = fmt.Sprintf("%s%s\033[38;2;15;20;28m %s %s", ansiBold(), pillColor.bgANSI(), cleanLabel, ansiReset())
	}
	if meta == "" {
		fmt.Fprintf(out, "%s %s\n", pill, text)
		return
	}
	leftVisible := len([]rune(cleanLabel)) + 3 + visibleLen(text)
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
		fmt.Fprintf(out, "%s %s %s%s %s%s\n", pill, text, ansiDim(), strings.Repeat("·", dotsCount), meta, ansiReset())
	} else {
		fmt.Fprintf(out, "%s %s %s(%s)%s\n", pill, text, ansiDim(), meta, ansiReset())
	}
}

func renderStep(out io.Writer, current, total, title, colorSpec, gradSpec string) {
	current = sanitizeText(current, false)
	total = sanitizeText(total, false)
	title = sanitizeText(title, false)
	barWidth := detectTermWidth() - 2
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

func renderRule(out io.Writer, label, colorSpec, gradSpec string, width int) {
	label = sanitizeText(label, false)
	if width > maxTermCols {
		width = maxTermCols
	}
	target := detectTermWidth() - 2
	if width > 0 && width < target {
		target = width
	}
	if target < 16 {
		target = 16
	}
	line := strings.Repeat("─", target)
	if clean := strings.TrimSpace(label); clean != "" {
		clean = truncateVisible(clean, target-6)
		padTotal := target - (visibleLen(clean) + 4)
		if padTotal < 2 {
			padTotal = 2
		}
		left := padTotal / 2
		line = strings.Repeat("─", left) + "  " + clean + "  " + strings.Repeat("─", padTotal-left)
	}
	pal := activePalette()
	if colorSpec == "" {
		colorSpec = activeConfig.RuleColor
	}
	grad, useGrad := resolveGradient(gradSpec, "", "")
	c := parseColor(colorSpec, pal.Primary)
	if useGrad {
		fmt.Fprintln(out, applyGradientLine(line, grad, false))
	} else {
		fmt.Fprintf(out, "%s%s%s\n", c.fgANSI(), line, ansiReset())
	}
}

func renderKV(out io.Writer, title, colorSpec, gradSpec string, pairs []string) {
	if len(pairs) == 0 {
		return
	}
	title = sanitizeText(title, false)
	maxCardWidth := detectTermWidth() - 2
	if maxCardWidth > 68 {
		maxCardWidth = 68
	}
	if maxCardWidth < 24 {
		maxCardWidth = 24
	}
	type kvPair struct{ k, v string }
	var items []kvPair
	maxKey, maxVal := 0, 0
	for i := 0; i < len(pairs); i += 2 {
		k, v := sanitizeText(pairs[i], false), ""
		if i+1 < len(pairs) {
			v = sanitizeText(pairs[i+1], false)
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
		formattedLines = append(formattedLines, fmt.Sprintf("%s%s%s %s%s%s %s%s%s",
			c.fgANSI(), it.k, ansiReset(), ansiDim(), strings.Repeat(".", dots), ansiReset(), ansiBold()+pal.Text.fgANSI(), it.v, ansiReset()))
	}
	renderBox(out, BoxOptions{Text: strings.Join(formattedLines, "\n"), Title: title, Border: activeConfig.BoxBorder, Color: colorSpec, Gradient: gradSpec, preformatted: true})
}

type TableOptions struct {
	Title, Headers, File, Delim, Align, Border, Color, Gradient string
	Width                                                       int
	Header, Compact, NoTruncate                                 bool
	Rows                                                        []string
}

var numericUnitSuffixes = []string{
	"kib", "mib", "gib", "tib",
	"kb", "mb", "gb", "tb",
	"ns", "us", "µs", "ms",
	"%", "s", "b", "k", "m", "g",
}

func isNumericCell(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if s[0] == '+' || s[0] == '-' {
		s = strings.TrimSpace(s[1:])
	}
	if strings.HasPrefix(s, "$") {
		s = strings.TrimSpace(s[1:])
	}
	lower := strings.ToLower(s)
	for _, suf := range numericUnitSuffixes {
		if strings.HasSuffix(lower, suf) && len(s) > len(suf) {
			s = strings.TrimSpace(s[:len(s)-len(suf)])
			break
		}
	}
	if s == "" {
		return false
	}
	digits, dots := 0, 0
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == ',':
			continue
		case r == '.':
			dots++
			if dots > 1 {
				return false
			}
		default:
			return false
		}
	}
	return digits > 0
}

func isNumericColumn(rows [][]string, col int) bool {
	if len(rows) == 0 {
		return false
	}
	checkRange := func(start int) bool {
		nonEmpty := 0
		for i := start; i < len(rows); i++ {
			if col >= len(rows[i]) {
				continue
			}
			val := strings.TrimSpace(rows[i][col])
			if val == "" {
				continue
			}
			if !isNumericCell(val) {
				return false
			}
			nonEmpty++
		}
		return nonEmpty > 0
	}
	if checkRange(0) {
		return true
	}
	// If the first row is an unflagged header label and all remaining rows are numeric,
	// still right-align the column so headerless TSV pipelines align cleanly.
	if len(rows) >= 2 && checkRange(1) {
		return true
	}
	return false
}

func normalizeAlignToken(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "r", "right":
		return "right"
	case "c", "center", "centre":
		return "center"
	case "l", "left":
		return "left"
	default:
		return ""
	}
}

func alignCell(s string, width int, align string) string {
	vLen := visibleLen(s)
	pad := width - vLen
	if pad <= 0 {
		return s
	}
	switch align {
	case "right":
		return strings.Repeat(" ", pad) + s
	case "center":
		left := pad / 2
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", pad-left)
	default:
		return s + strings.Repeat(" ", pad)
	}
}

func shrinkColumnWidths(natural []int, avail int) []int {
	widths := append([]int(nil), natural...)
	n := len(widths)
	if n == 0 {
		return widths
	}
	if avail < n {
		avail = n
	}
	total := 0
	for _, w := range widths {
		total += w
	}
	if total <= avail {
		return widths
	}
	// Shrink widest columns first down to a floor of 4 chars before shrinking narrower columns.
	for _, floorW := range []int{4, 1} {
		for total > avail {
			maxW := 0
			for _, w := range widths {
				if w > floorW && w > maxW {
					maxW = w
				}
			}
			if maxW == 0 {
				break
			}
			for j := range widths {
				if widths[j] == maxW && total > avail {
					widths[j]--
					total--
				}
			}
		}
	}
	return widths
}

func resolveDelimiter(delim string) string {
	if delim == "" || delim == `\t` || strings.EqualFold(delim, "tab") {
		return "\t"
	}
	return delim
}

func runTable(in io.Reader, out io.Writer, errOut io.Writer, opts TableOptions) int {
	pal := activePalette()
	delim := resolveDelimiter(opts.Delim)
	var rawLines []string

	if opts.File != "" {
		data, err := os.ReadFile(filepath.Clean(opts.File)) // #nosec G304
		if err != nil {
			renderPillBadge(errOut, "FAIL", pal.Fail, err.Error(), "")
			return 1
		}
		rawLines = strings.Split(string(data), "\n")
	} else if len(opts.Rows) > 0 {
		for _, r := range opts.Rows {
			expanded := strings.ReplaceAll(r, `\n`, "\n")
			if delim == "\t" {
				expanded = strings.ReplaceAll(expanded, `\t`, "\t")
			}
			rawLines = append(rawLines, strings.Split(expanded, "\n")...)
		}
	} else if in != nil {
		if f, ok := in.(*os.File); ok && isTerminal(f) {
			return 0
		}
		scanner := bufio.NewScanner(in)
		scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
		for scanner.Scan() {
			rawLines = append(rawLines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			renderPillBadge(errOut, "FAIL", pal.Fail, err.Error(), "")
			return 1
		}
	}

	renderTableLines(out, rawLines, opts)
	return 0
}

func renderTable(out io.Writer, opts TableOptions) {
	delim := resolveDelimiter(opts.Delim)
	var rawLines []string
	for _, r := range opts.Rows {
		expanded := strings.ReplaceAll(r, `\n`, "\n")
		if delim == "\t" {
			expanded = strings.ReplaceAll(expanded, `\t`, "\t")
		}
		rawLines = append(rawLines, strings.Split(expanded, "\n")...)
	}
	renderTableLines(out, rawLines, opts)
}

func renderTableLines(out io.Writer, rawLines []string, opts TableOptions) {
	delim := resolveDelimiter(opts.Delim)
	opts.Title = sanitizeText(opts.Title, false)

	var headerRow []string
	if opts.Headers != "" {
		hdrRaw := opts.Headers
		var parts []string
		if delim == "\t" && strings.Contains(hdrRaw, `\t`) {
			hdrRaw = strings.ReplaceAll(hdrRaw, `\t`, "\t")
		}
		if strings.Contains(hdrRaw, delim) {
			parts = strings.Split(hdrRaw, delim)
		} else {
			parts = strings.Split(hdrRaw, ",")
		}
		for _, p := range parts {
			headerRow = append(headerRow, sanitizeText(strings.TrimSpace(p), false))
		}
	}

	var rows [][]string
	for _, line := range rawLines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		parts := strings.Split(line, delim)
		row := make([]string, len(parts))
		for i, cell := range parts {
			row[i] = sanitizeText(cell, false)
		}
		rows = append(rows, row)
	}

	if opts.Header && len(headerRow) == 0 && len(rows) > 0 {
		headerRow = rows[0]
		rows = rows[1:]
	}
	if len(headerRow) == 0 && len(rows) == 0 {
		return
	}

	numCols := len(headerRow)
	for _, r := range rows {
		if len(r) > numCols {
			numCols = len(r)
		}
	}
	if numCols == 0 {
		return
	}

	if len(headerRow) > 0 && len(headerRow) < numCols {
		padded := make([]string, numCols)
		copy(padded, headerRow)
		headerRow = padded
	}
	for i := range rows {
		if len(rows[i]) < numCols {
			padded := make([]string, numCols)
			copy(padded, rows[i])
			rows[i] = padded
		}
	}

	aligns := make([]string, numCols)
	if strings.TrimSpace(opts.Align) != "" {
		tokens := strings.Split(opts.Align, ",")
		if len(tokens) == 1 {
			a := normalizeAlignToken(tokens[0])
			for j := 0; j < numCols; j++ {
				aligns[j] = a
			}
		} else {
			for j := 0; j < numCols && j < len(tokens); j++ {
				aligns[j] = normalizeAlignToken(tokens[j])
			}
		}
	}
	for j := 0; j < numCols; j++ {
		if aligns[j] == "" {
			if isNumericColumn(rows, j) {
				aligns[j] = "right"
			} else {
				aligns[j] = "left"
			}
		}
	}

	naturalWidths := make([]int, numCols)
	for j := 0; j < numCols; j++ {
		naturalWidths[j] = 1
		if j < len(headerRow) {
			if w := visibleLen(headerRow[j]); w > naturalWidths[j] {
				naturalWidths[j] = w
			}
		}
		for _, r := range rows {
			if w := visibleLen(r[j]); w > naturalWidths[j] {
				naturalWidths[j] = w
			}
		}
	}

	borderName := strings.ToLower(strings.TrimSpace(opts.Border))
	if borderName == "" {
		borderName = activeConfig.BoxBorder
	}
	compact := opts.Compact || borderName == "none"

	termWidth := detectTermWidth()
	maxTableWidth := clampWidth(termWidth-2, 16, maxTermCols)
	if opts.Width > 0 {
		maxTableWidth = clampWidth(opts.Width, 16, maxTermCols)
	}

	overhead := 3*numCols + 1
	if compact {
		overhead = 2 * (numCols - 1)
	}

	colWidths := append([]int(nil), naturalWidths...)
	if !opts.NoTruncate {
		colWidths = shrinkColumnWidths(naturalWidths, maxTableWidth-overhead)
	}

	// Ensure framed table is wide enough to fit --title in the top border cleanly
	if !compact && opts.Title != "" {
		innerTotal := 3*numCols - 1
		for _, w := range colWidths {
			innerTotal += w
		}
		minTitleInner := visibleLen(opts.Title) + 4
		if !opts.NoTruncate && minTitleInner > maxTableWidth-2 {
			minTitleInner = maxTableWidth - 2
		}
		if innerTotal < minTitleInner {
			extra := minTitleInner - innerTotal
			for extra > 0 {
				for j := 0; j < numCols && extra > 0; j++ {
					colWidths[j]++
					extra--
				}
			}
		}
	}

	pal := activePalette()
	colorSpec := opts.Color
	if colorSpec == "" {
		colorSpec = activeConfig.BoxColor
	}
	grad, useGrad := resolveGradient(opts.Gradient, "", "")
	c := parseColor(colorSpec, pal.Primary)
	hdrColor := pal.Secondary
	if opts.Color != "" {
		hdrColor = c
	}

	if compact {
		if opts.Title != "" {
			fmt.Fprintf(out, "%s▸%s %s%s%s\n", c.fgANSI(), ansiReset(), ansiBold(), opts.Title, ansiReset())
		}
		if len(headerRow) > 0 {
			var hCells, uCells []string
			for j := 0; j < numCols; j++ {
				cell := truncateVisible(headerRow[j], colWidths[j])
				padded := alignCell(cell, colWidths[j], aligns[j])
				hCells = append(hCells, ansiBold()+hdrColor.fgANSI()+padded+ansiReset())
				uCells = append(uCells, c.fgANSI()+strings.Repeat("─", colWidths[j])+ansiReset())
			}
			fmt.Fprintln(out, strings.Join(hCells, "  "))
			fmt.Fprintln(out, strings.Join(uCells, "  "))
		}
		for _, r := range rows {
			var dCells []string
			for j := 0; j < numCols; j++ {
				cell := truncateVisible(r[j], colWidths[j])
				padded := alignCell(cell, colWidths[j], aligns[j])
				dCells = append(dCells, pal.Text.fgANSI()+padded+ansiReset())
			}
			fmt.Fprintln(out, strings.Join(dCells, "  "))
		}
		return
	}

	bStyle, ok := borderStyles[borderName]
	if !ok {
		bStyle = borderStyles["rounded"]
	}

	buildRule := func(left, mid, right, fill string) string {
		segs := make([]string, numCols)
		for j, w := range colWidths {
			segs[j] = strings.Repeat(fill, w+2)
		}
		return left + strings.Join(segs, mid) + right
	}

	topLine := buildRule(bStyle.TopLeft, bStyle.TopTee, bStyle.TopRight, bStyle.Horizontal)
	if opts.Title != "" {
		topRunes := []rune(topLine)
		maxTitleLen := len(topRunes) - 5
		if maxTitleLen >= 1 {
			tStr := truncateVisible(opts.Title, maxTitleLen)
			prefix := []rune(bStyle.TopLeft + bStyle.Horizontal + " " + tStr + " ")
			if len(prefix) < len(topRunes) {
				topLine = string(prefix) + string(topRunes[len(prefix):])
			}
		}
	}
	midLine := buildRule(bStyle.LeftTee, bStyle.Cross, bStyle.RightTee, bStyle.Horizontal)
	botLine := buildRule(bStyle.BottomLeft, bStyle.BottomTee, bStyle.BottomRight, bStyle.Horizontal)

	emitBorderLine := func(line string) {
		if useGrad {
			fmt.Fprintln(out, applyGradientLine(line, grad, false))
		} else {
			fmt.Fprintf(out, "%s%s%s\n", c.fgANSI(), line, ansiReset())
		}
	}

	vertAt := func(colIdx int) string {
		colColor := c
		if useGrad && numCols > 0 {
			colColor = grad.At(float64(colIdx) / float64(numCols))
		}
		return colColor.fgANSI() + bStyle.Vertical + ansiReset()
	}

	emitRow := func(cells []string, isHeader bool) {
		var b strings.Builder
		b.WriteString(vertAt(0))
		for j := 0; j < numCols; j++ {
			cell := truncateVisible(cells[j], colWidths[j])
			padded := alignCell(cell, colWidths[j], aligns[j])
			b.WriteString(" ")
			if isHeader {
				b.WriteString(ansiBold() + hdrColor.fgANSI() + padded + ansiReset())
			} else {
				b.WriteString(pal.Text.fgANSI() + padded + ansiReset())
			}
			b.WriteString(" ")
			b.WriteString(vertAt(j + 1))
		}
		fmt.Fprintln(out, b.String())
	}

	emitBorderLine(topLine)
	if len(headerRow) > 0 {
		emitRow(headerRow, true)
		if len(rows) > 0 {
			emitBorderLine(midLine)
		}
	}
	for _, r := range rows {
		emitRow(r, false)
	}
	emitBorderLine(botLine)
}

func styleCommandTokens(cmdText string) string {
	pal := activePalette()
	var out, tok strings.Builder
	var quote rune
	tokIdx := 0
	flush := func() {
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
			flush()
			out.WriteRune(r)
		default:
			tok.WriteRune(r)
		}
	}
	flush()
	return out.String()
}

func renderCmd(out io.Writer, cmdText string) {
	pal := activePalette()
	fmt.Fprintf(out, "  %s┃%s %s▸%s %s\n", pal.Primary.fgANSI(), ansiReset(), pal.Secondary.fgANSI(), ansiReset(), styleCommandTokens(sanitizeText(cmdText, false)))
}

func renderCodeBlock(out io.Writer, title string, lines []string) {
	if len(lines) == 0 {
		return
	}
	title = sanitizeText(title, false)
	maxBoxWidth := detectTermWidth() - 4
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
		styled := styleCommandTokens(truncateVisible(sanitizeText(l, false), maxInner))
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
	fmt.Fprintf(out, "%s  ╭─ %s %s╮%s\n", borderColor.fgANSI(), title, strings.Repeat("─", rem), ansiReset())
	for _, styled := range styledLines {
		pad := innerWidth - visibleLen(styled)
		if pad < 0 {
			pad = 0
		}
		fmt.Fprintf(out, "%s  │%s  %s%s  %s│%s\n", borderColor.fgANSI(), ansiReset(), styled, strings.Repeat(" ", pad), borderColor.fgANSI(), ansiReset())
	}
	fmt.Fprintf(out, "%s  ╰%s╯%s\n", borderColor.fgANSI(), strings.Repeat("─", totalH), ansiReset())
}

func renderTree(out io.Writer, title string, items []string, colorSpec string) {
	pal := activePalette()
	c := parseColor(colorSpec, pal.Primary)
	title = sanitizeText(title, false)
	if title != "" {
		fmt.Fprintf(out, "%s%s▸ %s%s\n", ansiBold(), c.fgANSI(), title, ansiReset())
	}
	for i, item := range items {
		branch := "├─"
		if i == len(items)-1 {
			branch = "╰─"
		}
		fmt.Fprintf(out, "  %s%s%s %s●%s %s\n", ansiDim(), branch, ansiReset(), pal.Pass.fgANSI(), ansiReset(), sanitizeText(item, false))
	}
}

type StyleOptions struct {
	Text, Fg, Bg, Gradient, FromHex, ToHex string
	Bold, Dim, Italic, Underline           bool
}

func renderStyle(out io.Writer, opts StyleOptions) {
	opts.Text = sanitizeText(opts.Text, true)
	if grad, ok := resolveGradient(opts.Gradient, opts.FromHex, opts.ToHex); ok {
		fmt.Fprintln(out, applyGradientLine(opts.Text, grad, opts.Bold))
		return
	}
	if noColorEnabled() {
		fmt.Fprintln(out, opts.Text)
		return
	}
	var p strings.Builder
	if opts.Bold {
		p.WriteString(bold)
	}
	if opts.Dim {
		p.WriteString(dim)
	}
	if opts.Italic {
		p.WriteString(italic)
	}
	if opts.Underline {
		p.WriteString(underline)
	}
	if opts.Fg != "" {
		p.WriteString(parseColor(opts.Fg, namedColors["cyan"]).fgANSI())
	}
	if opts.Bg != "" {
		p.WriteString(parseColor(opts.Bg, namedColors["purple"]).bgANSI())
	}
	fmt.Fprintf(out, "%s%s%s\n", p.String(), opts.Text, ansiReset())
}

var spinnerStyles = map[string][]string{
	"dots":  {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	"pulse": {"◜", "◠", "◝", "◞", "◡", "◟"},
	"meter": {"▱▱▱", "▰▱▱", "▰▰▱", "▰▰▰", "▱▰▰", "▱▱▰"},
	"line":  {"-", "\\", "|", "/"},
}

type SpinOptions struct {
	Title, Style, Color string
	Persist             bool
	Static              bool
	Stdin               bool
	DurationMs          int
	Command             []string
}

var openDevTTY = func() (*os.File, error) {
	return os.OpenFile("/dev/tty", os.O_WRONLY, 0)
}

func shouldIsolateProcessGroup(hasControllingTTY bool) bool {
	return !hasControllingTTY
}

func resolveTTYState(devTTYAvailable, stderrIsTerminal bool) (canAnimate, hasControllingTTY bool) {
	return devTTYAvailable || stderrIsTerminal, devTTYAvailable
}

func openTTYWriter(fallback io.Writer) (io.Writer, bool, bool, func()) {
	if tty, err := openDevTTY(); err == nil {
		canAnim, hasCtrl := resolveTTYState(true, false)
		return tty, canAnim, hasCtrl, func() { _ = tty.Close() }
	}
	canAnim, hasCtrl := resolveTTYState(false, isWriterTerminal(fallback))
	return fallback, canAnim, hasCtrl, func() {}
}

func isWriterTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isTerminal(f)
	}
	return false
}

func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	if dn, err := os.Stat(os.DevNull); err == nil && os.SameFile(fi, dn) {
		return false
	}
	return true
}

func signalExitCode(sig os.Signal) int {
	if s, ok := sig.(syscall.Signal); ok && int(s) > 0 {
		return 128 + int(s)
	}
	if sig == syscall.SIGTERM {
		return 143
	}
	return 130
}

func exitCodeFromError(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			return 128 + int(ws.Signal())
		}
		if code := exitErr.ExitCode(); code > 0 {
			return code
		}
	}
	return 1
}

func runSpinner(stdout, stderr io.Writer, opts SpinOptions) int {
	ttyOut, canAnimate, hasControllingTTY, closeTTY := openTTYWriter(stderr)
	defer closeTTY()
	styleName := strings.ToLower(strings.TrimSpace(opts.Style))
	if styleName == "" {
		styleName = activeConfig.SpinStyle
	}
	frames, ok := spinnerStyles[styleName]
	if !ok {
		frames = spinnerStyles["dots"]
	}
	pal := activePalette()
	colorSpec := opts.Color
	if colorSpec == "" {
		colorSpec = activeConfig.SpinColor
	}
	c := parseColor(colorSpec, pal.Primary)
	title := strings.TrimSpace(sanitizeText(opts.Title, false))
	if title == "" {
		title = "Working..."
	}

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, spinnerSignals()...)
	defer signal.Stop(sigCh)

	start := time.Now()
	doneCh := make(chan error, 1)
	var cmd *exec.Cmd

	staticMode := opts.Static || !canAnimate
	if staticMode {
		fmt.Fprintf(ttyOut, "  %s%s▸%s %s%s%s\n", ansiBold(), c.fgANSI(), ansiReset(), ansiBold(), title, ansiReset())
	} else if !noColorEnabled() {
		fmt.Fprint(ttyOut, "\033[?25l")
	}
	restoreTerminal := func() {
		if staticMode {
			return
		}
		if !noColorEnabled() {
			fmt.Fprint(ttyOut, "\r\033[2K\033[?25h")
		} else {
			fmt.Fprint(ttyOut, "\r")
		}
	}

	if len(opts.Command) > 0 {
		cmd = exec.Command(opts.Command[0], opts.Command[1:]...) // #nosec G204
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		stdinTTY := isTerminal(os.Stdin)
		if opts.Stdin || !stdinTTY {
			cmd.Stdin = os.Stdin
		}
		prepareChild(cmd, hasControllingTTY)
		startErrCh := make(chan error, 1)
		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			if err := cmd.Start(); err != nil {
				startErrCh <- err
				return
			}
			startErrCh <- nil
			doneCh <- cmd.Wait()
		}()
		if err := <-startErrCh; err != nil {
			restoreTerminal()
			renderPillBadge(ttyOut, "FAIL", pal.Fail, title, err.Error())
			return 1
		}
	} else {
		dur := opts.DurationMs
		if dur <= 0 {
			dur = 600
		}
		go func() {
			time.Sleep(time.Duration(dur) * time.Millisecond)
			doneCh <- nil
		}()
	}

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	frameIdx := 0
	sigCount := 0
	var lastSig os.Signal
	var killTimer <-chan time.Time

	for {
		select {
		case sig := <-sigCh:
			if cmd == nil || cmd.Process == nil {
				restoreTerminal()
				renderPillBadge(ttyOut, "INT", pal.Warn, title, "interrupted")
				return signalExitCode(sig)
			}
			sigCount++
			lastSig = sig
			sysSig := syscall.SIGINT
			if s, ok := sig.(syscall.Signal); ok {
				sysSig = s
			}
			if sigCount == 1 {
				if sysSig == syscall.SIGINT && hasControllingTTY {
					// A controlling terminal is attached and already delivered SIGINT to the foreground process group;
					// let the child take as long as it needs to clean up until a second Ctrl-C arrives.
					continue
				}
				signalChild(cmd, sysSig)
				killTimer = time.After(5 * time.Second)
				continue
			}
			signalChild(cmd, syscall.SIGKILL)
			select {
			case <-doneCh:
			case <-time.After(100 * time.Millisecond):
			}
			restoreTerminal()
			renderPillBadge(ttyOut, "INT", pal.Warn, title, "interrupted")
			return signalExitCode(lastSig)
		case <-killTimer:
			signalChild(cmd, syscall.SIGKILL)
			select {
			case <-doneCh:
			case <-time.After(100 * time.Millisecond):
			}
			restoreTerminal()
			renderPillBadge(ttyOut, "INT", pal.Warn, title, "terminated")
			return signalExitCode(lastSig)
		case err := <-doneCh:
			restoreTerminal()
			elapsed := formatDuration(time.Since(start))
			if sigCount > 0 {
				renderPillBadge(ttyOut, "INT", pal.Warn, title, "interrupted")
				if err != nil {
					if code := exitCodeFromError(err); code > 1 {
						return code
					}
				}
				return signalExitCode(lastSig)
			}
			if err != nil {
				renderPillBadge(ttyOut, "FAIL", pal.Fail, title, elapsed)
				return exitCodeFromError(err)
			}
			if opts.Persist || staticMode {
				renderPillBadge(ttyOut, "PASS", pal.Pass, title, elapsed)
			}
			return 0
		case <-ticker.C:
			if staticMode {
				continue
			}
			termWidth := detectTermWidth()
			elapsed := formatDuration(time.Since(start))
			frame := frames[frameIdx%len(frames)]
			frameIdx++
			maxTitleLen := termWidth - visibleLen(frame) - visibleLen(elapsed) - 8
			if maxTitleLen < 8 {
				maxTitleLen = 8
			}
			fmt.Fprintf(ttyOut, "\r\033[2K%s%s%s%s %s%s%s %s(%s)%s",
				ansiBold(), c.fgANSI(), frame, ansiReset(),
				ansiBold(), truncateVisible(title, maxTitleLen), ansiReset(),
				ansiDim(), elapsed, ansiReset())
		}
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

type BarOptions struct {
	Current, Total, Width  int
	Label, Color, Gradient string
	InPlace                bool
}

func formatProgressBar(opts BarOptions) string {
	total := opts.Total
	if total <= 0 {
		total = 100
	}
	curr := opts.Current
	if curr < 0 {
		curr = 0
	}
	if curr > total {
		curr = total
	}
	ratio := float64(curr) / float64(total)
	pct := int(math.Round(ratio * 100))
	termWidth := detectTermWidth()
	statsSuffix := fmt.Sprintf(" %3d%% [%d/%d]", pct, curr, total)
	labelPrefix := ""
	if opts.Label != "" {
		maxLabel := termWidth / 3
		if maxLabel < 10 {
			maxLabel = 10
		}
		labelPrefix = truncateVisible(sanitizeText(opts.Label, false), maxLabel) + " "
	}
	barCells := clampWidth(opts.Width, 0, maxTermCols)
	if barCells <= 0 {
		barCells = termWidth - visibleLen(labelPrefix) - visibleLen(statsSuffix) - 4
		if barCells > 42 {
			barCells = 42
		}
	}
	if barCells < 8 {
		barCells = 8
	}
	exactFilled := ratio * float64(barCells)
	fullCells := int(math.Floor(exactFilled))
	remHalf := (exactFilled - float64(fullCells)) >= 0.5
	if fullCells > barCells {
		fullCells = barCells
		remHalf = false
	}
	filledStr := strings.Repeat("█", fullCells)
	if fullCells < barCells && remHalf {
		filledStr += "▌"
	}
	emptyLen := barCells - len([]rune(filledStr))
	if emptyLen < 0 {
		emptyLen = 0
	}
	emptyStr := strings.Repeat("░", emptyLen)

	pal := activePalette()
	colorSpec := opts.Color
	if colorSpec == "" {
		colorSpec = activeConfig.BarColor
	}
	grad, useGrad := resolveGradient(opts.Gradient, "", "")
	c := parseColor(colorSpec, pal.Primary)
	coloredFilled := c.fgANSI() + filledStr + ansiReset()
	if useGrad {
		coloredFilled = applyGradientLine(filledStr, grad, false)
	}
	return fmt.Sprintf("%s%s%s%s%s%s%s%s%s", ansiBold(), labelPrefix, ansiReset(), coloredFilled, ansiDim(), emptyStr, ansiReset(), ansiBold(), statsSuffix+ansiReset())
}

func renderBar(out io.Writer, opts BarOptions) {
	line := formatProgressBar(opts)
	if opts.InPlace {
		fmt.Fprintf(out, "\r\033[2K%s", line)
		if opts.Total > 0 && opts.Current >= opts.Total {
			fmt.Fprintln(out)
		}
	} else {
		fmt.Fprintln(out, line)
	}
}

func runStreamProgress(in io.Reader, out io.Writer, total int, title, colorSpec, gradSpec string) int {
	ttyOut, canAnimate, _, closeTTY := openTTYWriter(out)
	defer closeTTY()
	if total <= 0 {
		total = 100
	}
	title = sanitizeText(title, false)
	if title == "" {
		title = "Processing"
	}
	start := time.Now()
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	count := 0
	if canAnimate {
		renderBar(ttyOut, BarOptions{Current: 0, Total: total, Label: title, Color: colorSpec, Gradient: gradSpec, InPlace: true})
	}
	for scanner.Scan() {
		count++
		if count > total {
			total = count
		}
		if canAnimate {
			renderBar(ttyOut, BarOptions{Current: count, Total: total, Label: title, Color: colorSpec, Gradient: gradSpec, InPlace: true})
		}
	}
	if canAnimate {
		fmt.Fprint(ttyOut, "\r\033[2K")
	}
	if err := scanner.Err(); err != nil {
		renderPillBadge(ttyOut, "FAIL", activePalette().Fail, title, err.Error())
		return 1
	}
	renderPillBadge(ttyOut, "PASS", activePalette().Pass, title, fmt.Sprintf("%d items in %s", count, formatDuration(time.Since(start))))
	return 0
}
