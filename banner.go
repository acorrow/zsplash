package main

import (
	"fmt"
	"io"
	"strings"
)

// compactGlyphs is a 2-row crisp half-block font (▀ ▄ █).
var compactGlyphs = map[rune][2]string{
	'A': {"▄▀▄", "█▀█"},
	'B': {"██▄", "██▀"},
	'C': {"▄▀▀", "▀▄▄"},
	'D': {"█▀▄", "█▄▀"},
	'E': {"█▀▀", "██▄"},
	'F': {"█▀▀", "█▀ "},
	'G': {"▄▀▀", "▀▄█"},
	'H': {"█ █", "█▀█"},
	'I': {"▀█▀", "▄█▄"},
	'J': {"  █", "▀▄▀"},
	'K': {"█▄▀", "█ █"},
	'L': {"█  ", "█▄▄"},
	'M': {"█▄█", "█ █"},
	'N': {"█▀█", "█ █"},
	'O': {"▄▀▄", "▀▄▀"},
	'P': {"█▀█", "█▀▀"},
	'Q': {"▄▀▄", "▀▄█"},
	'R': {"█▀█", "█▀▄"},
	'S': {"▄▀▀", "▄▄▀"},
	'T': {"▀█▀", " █ "},
	'U': {"█ █", "▀▄▀"},
	'V': {"█ █", "▀▄▀"},
	'W': {"█ █", "█▀█"},
	'X': {"▀▄▀", "█ █"},
	'Y': {"█ █", " █ "},
	'Z': {"▀▀█", "█▄▄"},
	'0': {"▄▀▄", "▀▄▀"},
	'1': {"▄█ ", " █ "},
	'2': {"▀▀█", "█▄▄"},
	'3': {"▀▀█", "▄▄█"},
	'4': {"█ █", "▀▀█"},
	'5': {"█▀▀", "▄▄▀"},
	'6': {"▄▀▀", "█▄█"},
	'7': {"▀▀█", "  █"},
	'8': {"▄▀▄", "█▄█"},
	'9': {"█▀█", "▀▀█"},
	'-': {"   ", "▀▀▀"},
	'_': {"   ", "▄▄▄"},
	'.': {"   ", " ▄ "},
	':': {" ▄ ", " ▄ "},
	'!': {" █ ", " ▄ "},
	'/': {" ▄▀", "▀  "},
	' ': {"  ", "  "},
}

// blockGlyphs is a 3-row high-contrast shaded Unicode block font.
var blockGlyphs = map[rune][3]string{
	'A': {"▄▀▀▄", "█▀▀█", "▀  ▀"},
	'B': {"██▀▄", "██▀▄", "▀▀▀ "},
	'C': {"▄▀▀▀", "█   ", "▀▀▀▀"},
	'D': {"█▀▀▄", "█  █", "▀▀▀ "},
	'E': {"█▀▀▀", "█▀▀ ", "▀▀▀▀"},
	'F': {"█▀▀▀", "█▀▀ ", "▀   "},
	'G': {"▄▀▀▀", "█ ▀█", "▀▀▀▀"},
	'H': {"█  █", "█▀▀█", "▀  ▀"},
	'I': {"▀█▀", " █ ", "▀▀▀"},
	'J': {"   █", "   █", "▀▀▀ "},
	'K': {"█ ▄▀", "█▀▄ ", "▀  ▀"},
	'L': {"█   ", "█   ", "▀▀▀▀"},
	'M': {"█▀▄▀█", "█ ▀ █", "▀   ▀"},
	'N': {"█▀▄ █", "█ ▀▄█", "▀   ▀"},
	'O': {"▄▀▀▄", "█  █", "▀▀▀▀"},
	'P': {"█▀▀▄", "█▀▀ ", "▀   "},
	'Q': {"▄▀▀▄", "█ ▄█", "▀▀▀▄"},
	'R': {"█▀▀▄", "█▀▀▄", "▀  ▀"},
	'S': {"▄▀▀▀", "▀▀▀▄", "▀▀▀ "},
	'T': {"▀▀█▀▀", "  █  ", "  ▀  "},
	'U': {"█  █", "█  █", "▀▀▀▀"},
	'V': {"█  █", "▀▄▄▀", " ▀▀ "},
	'W': {"█   █", "█ ▄ █", "▀▄▀▄▀"},
	'X': {"▀▄ ▄▀", "  █  ", "▄▀ ▀▄"},
	'Y': {"█  █", " ▀▀█", "▀▀▀ "},
	'Z': {"▀▀▀█", " ▄▀ ", "▀▀▀▀"},
	'0': {"▄▀▀▄", "█  █", "▀▀▀▀"},
	'1': {"▄█ ", " █ ", "▀▀▀"},
	'2': {"▀▀▀▄", "▄▀▀ ", "▀▀▀▀"},
	'3': {"▀▀▀▄", " ▀▀▄", "▀▀▀ "},
	'4': {"█  █", "▀▀▀█", "   ▀"},
	'5': {"█▀▀▀", "▀▀▀▄", "▀▀▀ "},
	'6': {"▄▀▀▀", "█▀▀▄", "▀▀▀▀"},
	'7': {"▀▀▀█", "  ▄▀", " ▀  "},
	'8': {"▄▀▀▄", "▄▀▀▄", "▀▀▀▀"},
	'9': {"▄▀▀▄", "▀▀▀█", "▀▀▀ "},
	'-': {"    ", "▀▀▀▀", "    "},
	'_': {"    ", "    ", "▀▀▀▀"},
	'.': {"  ", "  ", "▀ "},
	':': {"▄ ", "  ", "▀ "},
	'!': {"█ ", "█ ", "▀ "},
	'/': {"  ▄▀", " ▄▀ ", "▀   "},
	' ': {"   ", "   ", "   "},
}

// lineGlyphs is a 3-row architectural Unicode box-drawing font.
var lineGlyphs = map[rune][3]string{
	'A': {"╭─╮", "├─┤", "╵ ╵"},
	'B': {"├─╮", "├─┤", "╰─╯"},
	'C': {"╭─╴", "│  ", "╰─╴"},
	'D': {"├─╮", "│ │", "╰─╯"},
	'E': {"╭─╴", "├─╴", "╰─╴"},
	'F': {"╭─╴", "├─╴", "╵  "},
	'G': {"╭─╴", "│ ╮", "╰─╯"},
	'H': {"╷ ╷", "├─┤", "╵ ╵"},
	'I': {"╶┬╴", " │ ", "╶┴╴"},
	'J': {"  ╷", "  │", "╰─╯"},
	'K': {"╷ ╭", "├─┤", "╵ ╰"},
	'L': {"╷  ", "│  ", "╰─╴"},
	'M': {"╭┬╮", "│││", "╵╵╵"},
	'N': {"╭─╮", "│ │", "╵ ╵"},
	'O': {"╭─╮", "│ │", "╰─╯"},
	'P': {"╭─╮", "├─╯", "╵  "},
	'Q': {"╭─╮", "│ ┼", "╰─╯"},
	'R': {"╭─╮", "├─┬", "╵ ╰"},
	'S': {"╭─╴", "╰─╮", "╶─╯"},
	'T': {"╶┬╴", " │ ", " ╵ "},
	'U': {"╷ ╷", "│ │", "╰─╯"},
	'V': {"╷ ╷", "│ │", "╰─╯"},
	'W': {"╷ ╷", "│││", "╰┴╯"},
	'X': {"╲ ╱", " ╳ ", "╱ ╲"},
	'Y': {"╷ ╷", "╰┬╯", " ╵ "},
	'Z': {"╶─╮", " ╭╯", "╰─╴"},
	'0': {"╭─╮", "│ │", "╰─╯"},
	'1': {"╶┐ ", " │ ", "╶┴╴"},
	'2': {"╶─╮", "╭─╯", "╰─╴"},
	'3': {"╶─╮", " ─┤", "╶─╯"},
	'4': {"╷ ╷", "╰─┤", "  ╵"},
	'5': {"╭─╴", "╰─╮", "╶─╯"},
	'6': {"╭─╴", "├─╮", "╰─╯"},
	'7': {"╶─╮", "  │", "  ╵"},
	'8': {"╭─╮", "├─┤", "╰─╯"},
	'9': {"╭─╮", "╰─┤", "╶─╯"},
	'-': {"   ", "───", "   "},
	'_': {"   ", "   ", "───"},
	'.': {" ", " ", "∙"},
	':': {"∙", " ", "∙"},
	'!': {"│", "│", "∙"},
	'/': {"  ╱", " ╱ ", "╱  "},
	' ': {"  ", "  ", "  "},
}

type BannerOptions struct {
	Text     string
	Font     string // "block" (default), "compact", "line"
	Gradient string // "cyber" (default), "ocean", "ember", "sunset", "matrix", "gold", "mono"
	FromHex  string
	ToHex    string
	Color    string
	Subtitle string
	Frame    bool
}

func renderBanner(out io.Writer, opts BannerOptions) {
	text := strings.ToUpper(strings.TrimSpace(opts.Text))
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

	// Split words onto multiple banner chunks if the rendered width exceeds terminal width.
	words := strings.Fields(text)
	if len(words) == 0 {
		words = []string{text}
	}

	var chunks []string
	currentChunk := words[0]
	for _, w := range words[1:] {
		candidate := currentChunk + " " + w
		if estimateBannerWidth(candidate, font) <= termWidth-4 {
			currentChunk = candidate
		} else {
			chunks = append(chunks, currentChunk)
			currentChunk = w
		}
	}
	chunks = append(chunks, currentChunk)

	var allRenderedRows []string
	for idx, chunk := range chunks {
		rows := buildFontRows(chunk, font)
		allRenderedRows = append(allRenderedRows, rows...)
		if idx < len(chunks)-1 {
			allRenderedRows = append(allRenderedRows, "")
		}
	}

	if opts.Frame {
		maxW := 0
		for _, r := range allRenderedRows {
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
		for _, r := range allRenderedRows {
			padded := padRight(r, maxW)
			var coloredInner string
			if useGrad {
				coloredInner = applyGradientLine(padded, grad, true)
			} else {
				coloredInner = solidColor.fgANSI() + ansiBold() + padded + ansiReset()
			}
			bColor := solidColor
			if useGrad {
				bColor = grad.At(0.0)
			}
			rColor := solidColor
			if useGrad {
				rColor = grad.At(1.0)
			}
			fmt.Fprintf(out, "%s│%s  %s  %s│%s\n", bColor.fgANSI(), ansiReset(), coloredInner, rColor.fgANSI(), ansiReset())
		}
		if opts.Subtitle != "" {
			sep := "├" + strings.Repeat("─", maxW+4) + "┤"
			if useGrad {
				fmt.Fprintln(out, applyGradientLine(sep, grad, false))
			} else {
				fmt.Fprintf(out, "%s%s%s\n", solidColor.fgANSI(), sep, ansiReset())
			}
			sub := padRight(truncateVisible(opts.Subtitle, maxW), maxW)
			bColor := solidColor
			rColor := solidColor
			if useGrad {
				bColor = grad.At(0.0)
				rColor = grad.At(1.0)
			}
			fmt.Fprintf(out, "%s│%s  %s%s%s  %s│%s\n", bColor.fgANSI(), ansiReset(), ansiDim(), sub, ansiReset(), rColor.fgANSI(), ansiReset())
		}
		if useGrad {
			fmt.Fprintln(out, applyGradientLine(botBorder, grad, false))
		} else {
			fmt.Fprintf(out, "%s%s%s\n", solidColor.fgANSI(), botBorder, ansiReset())
		}
		return
	}

	for _, r := range allRenderedRows {
		if useGrad {
			fmt.Fprintln(out, applyGradientLine(r, grad, true))
		} else {
			fmt.Fprintf(out, "%s%s%s%s\n", solidColor.fgANSI(), ansiBold(), r, ansiReset())
		}
	}
	if opts.Subtitle != "" {
		fmt.Fprintf(out, "%s%s%s\n", ansiDim(), opts.Subtitle, ansiReset())
	}
}

func estimateBannerWidth(text, font string) int {
	rows := buildFontRows(text, font)
	if len(rows) == 0 {
		return 0
	}
	return visibleLen(rows[0])
}

func buildFontRows(text, font string) []string {
	switch font {
	case "compact", "mini", "small":
		var r0, r1 strings.Builder
		runes := []rune(text)
		for i, ch := range runes {
			g, ok := compactGlyphs[ch]
			if !ok {
				g = compactGlyphs['?']
				if g[0] == "" {
					g = [2]string{"▄▀▄", " ▄ "}
				}
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
		runes := []rune(text)
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

	default: // "block"
		var r0, r1, r2 strings.Builder
		runes := []rune(text)
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
