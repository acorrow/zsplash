package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
	italic    = "\033[3m"
	underline = "\033[4m"
)

// RGB represents a 24-bit TrueColor value.
type RGB struct {
	R, G, B int
}

// fgANSI returns the 24-bit foreground escape sequence.
func (c RGB) fgANSI() string {
	if noColorEnabled() {
		return ""
	}
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", c.R, c.G, c.B)
}

// bgANSI returns the 24-bit background escape sequence.
func (c RGB) bgANSI() string {
	if noColorEnabled() {
		return ""
	}
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", c.R, c.G, c.B)
}

var namedColors = map[string]RGB{
	"cyan":    {0, 245, 212},
	"green":   {46, 213, 115},
	"yellow":  {255, 165, 2},
	"red":     {255, 71, 87},
	"magenta": {196, 113, 237},
	"purple":  {155, 93, 229},
	"pink":    {241, 91, 181},
	"blue":    {0, 187, 249},
	"orange":  {255, 123, 0},
	"gold":    {254, 228, 64},
	"white":   {245, 246, 250},
	"gray":    {130, 140, 160},
}

// parseColor resolves a named color or #RRGGBB hex string into an RGB struct.
func parseColor(spec string, fallback RGB) RGB {
	s := strings.ToLower(strings.TrimSpace(spec))
	if s == "" {
		return fallback
	}
	if c, ok := namedColors[s]; ok {
		return c
	}
	s = strings.TrimPrefix(s, "#")
	if len(s) == 6 {
		r, err1 := strconv.ParseInt(s[0:2], 16, 32)
		g, err2 := strconv.ParseInt(s[2:4], 16, 32)
		b, err3 := strconv.ParseInt(s[4:6], 16, 32)
		if err1 == nil && err2 == nil && err3 == nil {
			return RGB{int(r), int(g), int(b)}
		}
	}
	return fallback
}

// Gradient defines a two-stop or three-stop TrueColor gradient.
type Gradient struct {
	Stops []RGB
}

var namedGradients = map[string]Gradient{
	"cyber": {Stops: []RGB{
		{0, 245, 212},
		{0, 187, 249},
		{155, 93, 229},
	}},
	"ocean": {Stops: []RGB{
		{72, 202, 228},
		{0, 119, 182},
		{3, 4, 94},
	}},
	"ember": {Stops: []RGB{
		{255, 71, 87},
		{255, 123, 0},
		{254, 228, 64},
	}},
	"sunset": {Stops: []RGB{
		{241, 91, 181},
		{255, 123, 0},
		{254, 228, 64},
	}},
	"matrix": {Stops: []RGB{
		{0, 255, 135},
		{46, 213, 115},
		{16, 124, 65},
	}},
	"gold": {Stops: []RGB{
		{254, 228, 64},
		{255, 165, 2},
		{217, 119, 6},
	}},
	"mono": {Stops: []RGB{
		{255, 255, 255},
		{180, 190, 210},
		{110, 120, 140},
	}},
}

// resolveGradient returns a Gradient by preset name or custom from/to hex stops.
func resolveGradient(preset, fromSpec, toSpec string) (Gradient, bool) {
	if fromSpec != "" || toSpec != "" {
		c1 := parseColor(fromSpec, namedColors["cyan"])
		c2 := parseColor(toSpec, namedColors["magenta"])
		return Gradient{Stops: []RGB{c1, c2}}, true
	}
	p := strings.ToLower(strings.TrimSpace(preset))
	if g, ok := namedGradients[p]; ok {
		return g, true
	}
	return Gradient{}, false
}

// At interpolates the gradient at position t in [0.0, 1.0].
func (g Gradient) At(t float64) RGB {
	if len(g.Stops) == 0 {
		return namedColors["cyan"]
	}
	if len(g.Stops) == 1 || t <= 0 {
		return g.Stops[0]
	}
	if t >= 1 {
		return g.Stops[len(g.Stops)-1]
	}

	segments := float64(len(g.Stops) - 1)
	scaled := t * segments
	idx := int(scaled)
	if idx >= len(g.Stops)-1 {
		idx = len(g.Stops) - 2
	}
	localT := scaled - float64(idx)
	a := g.Stops[idx]
	b := g.Stops[idx+1]

	return RGB{
		R: int(float64(a.R) + localT*float64(b.R-a.R)),
		G: int(float64(a.G) + localT*float64(b.G-a.G)),
		B: int(float64(a.B) + localT*float64(b.B-a.B)),
	}
}

// applyGradientLine colors each rune of s along the gradient horizontally.
func applyGradientLine(s string, g Gradient, isBold bool) string {
	if noColorEnabled() {
		return s
	}
	runes := []rune(s)
	n := len(runes)
	if n == 0 {
		return ""
	}
	var b strings.Builder
	if isBold {
		b.WriteString(bold)
	}
	for i, r := range runes {
		if r == ' ' {
			b.WriteRune(' ')
			continue
		}
		t := 0.0
		if n > 1 {
			t = float64(i) / float64(n-1)
		}
		b.WriteString(g.At(t).fgANSI())
		b.WriteRune(r)
	}
	b.WriteString(reset)
	return b.String()
}

func ansiReset() string {
	if noColorEnabled() {
		return ""
	}
	return reset
}

func ansiBold() string {
	if noColorEnabled() {
		return ""
	}
	return bold
}

func ansiDim() string {
	if noColorEnabled() {
		return ""
	}
	return dim
}
