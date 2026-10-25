package main

import (
	"fmt"
	"io"
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

// Hex returns the #RRGGBB representation of the color.
func (c RGB) Hex() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
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

// ThemePalette defines solid semantic colors for all zsplash components.
type ThemePalette struct {
	Name        string
	Description string
	Primary     RGB
	Secondary   RGB
	Pass        RGB
	Warn        RGB
	Fail        RGB
	Info        RGB
	Muted       RGB
	Text        RGB
}

var themeOrder = []string{
	"modern",
	"nes",
	"snes",
	"gameboy",
	"dracula",
	"catppuccin",
	"nord",
	"gruvbox",
	"tokyonight",
	"amber",
}

// builtinThemes contains all built-in solid color palettes, including the NES 2C02 hardware palette.
var builtinThemes = map[string]ThemePalette{
	"modern": {
		Name:        "modern",
		Description: "High-contrast crisp teal, sky blue, and emerald (Default)",
		Primary:     RGB{0, 245, 212},
		Secondary:   RGB{0, 187, 249},
		Pass:        RGB{46, 213, 115},
		Warn:        RGB{255, 165, 2},
		Fail:        RGB{255, 71, 87},
		Info:        RGB{56, 189, 248},
		Muted:       RGB{120, 132, 158},
		Text:        RGB{245, 246, 250},
	},
	"nes": {
		Name:        "nes",
		Description: "Nintendo Entertainment System (Ricoh 2C02 PPU hardware palette)",
		Primary:     RGB{228, 0, 88},    // NES $16/$26 Mario & Logo Crimson Red (#E40058)
		Secondary:   RGB{60, 188, 252},  // NES $21 Mega Man Sky Blue (#3CBCFC)
		Pass:        RGB{88, 216, 84},   // NES $2A 1-Up & Pipe Green (#58D854)
		Warn:        RGB{248, 184, 0},   // NES $28 Coin Gold (#F8B800)
		Fail:        RGB{216, 40, 0},    // NES $16 Fire/Lava Red (#D82800)
		Info:        RGB{60, 188, 252},  // NES $21 Sky Blue (#3CBCFC)
		Muted:       RGB{124, 124, 124}, // NES $00 Controller Deck Gray (#7C7C7C)
		Text:        RGB{252, 252, 252}, // NES $30 PPU Pure White (#FCFCFC)
	},
	"snes": {
		Name:        "snes",
		Description: "Super Nintendo 16-bit lavender, periwinkle, and primary button colors",
		Primary:     RGB{123, 104, 238},
		Secondary:   RGB{178, 164, 255},
		Pass:        RGB{0, 200, 83},
		Warn:        RGB{255, 214, 0},
		Fail:        RGB{229, 57, 53},
		Info:        RGB{30, 136, 229},
		Muted:       RGB{141, 141, 146},
		Text:        RGB{244, 244, 246},
	},
	"gameboy": {
		Name:        "gameboy",
		Description: "Nintendo Game Boy DMG-01 4-shade olive LCD phosphor green",
		Primary:     RGB{155, 188, 15},
		Secondary:   RGB{139, 172, 15},
		Pass:        RGB{155, 188, 15},
		Warn:        RGB{139, 172, 15},
		Fail:        RGB{48, 98, 48},
		Info:        RGB{139, 172, 15},
		Muted:       RGB{48, 98, 48},
		Text:        RGB{202, 220, 159},
	},
	"dracula": {
		Name:        "dracula",
		Description: "Dracula dark theme with purple, pink, and neon green accents",
		Primary:     RGB{189, 147, 249},
		Secondary:   RGB{255, 121, 198},
		Pass:        RGB{80, 250, 123},
		Warn:        RGB{241, 250, 140},
		Fail:        RGB{255, 85, 85},
		Info:        RGB{139, 233, 253},
		Muted:       RGB{98, 114, 164},
		Text:        RGB{248, 248, 242},
	},
	"catppuccin": {
		Name:        "catppuccin",
		Description: "Catppuccin Mocha soothing pastel mauve, sapphire, and green",
		Primary:     RGB{203, 166, 247},
		Secondary:   RGB{245, 194, 231},
		Pass:        RGB{166, 227, 161},
		Warn:        RGB{249, 226, 175},
		Fail:        RGB{243, 139, 168},
		Info:        RGB{137, 220, 235},
		Muted:       RGB{108, 112, 134},
		Text:        RGB{205, 214, 244},
	},
	"nord": {
		Name:        "nord",
		Description: "Arctic Nord frost blues and aurora borealis status colors",
		Primary:     RGB{136, 192, 208},
		Secondary:   RGB{129, 161, 193},
		Pass:        RGB{163, 190, 140},
		Warn:        RGB{235, 203, 139},
		Fail:        RGB{191, 97, 106},
		Info:        RGB{94, 129, 172},
		Muted:       RGB{106, 118, 142},
		Text:        RGB{236, 239, 244},
	},
	"gruvbox": {
		Name:        "gruvbox",
		Description: "Retro groove warm amber, earthy orange, and olive green",
		Primary:     RGB{254, 128, 25},
		Secondary:   RGB{250, 189, 47},
		Pass:        RGB{184, 187, 38},
		Warn:        RGB{250, 189, 47},
		Fail:        RGB{251, 73, 52},
		Info:        RGB{131, 165, 152},
		Muted:       RGB{146, 131, 116},
		Text:        RGB{235, 219, 178},
	},
	"tokyonight": {
		Name:        "tokyonight",
		Description: "Tokyo Night neon blue, cyber purple, and teal",
		Primary:     RGB{122, 162, 247},
		Secondary:   RGB{187, 154, 247},
		Pass:        RGB{158, 206, 106},
		Warn:        RGB{224, 175, 104},
		Fail:        RGB{247, 118, 142},
		Info:        RGB{125, 207, 255},
		Muted:       RGB{110, 120, 165},
		Text:        RGB{192, 202, 245},
	},
	"amber": {
		Name:        "amber",
		Description: "Vintage 1980s monochrome amber phosphor CRT terminal",
		Primary:     RGB{255, 176, 0},
		Secondary:   RGB{255, 204, 0},
		Pass:        RGB{255, 208, 64},
		Warn:        RGB{255, 144, 0},
		Fail:        RGB{255, 85, 0},
		Info:        RGB{255, 176, 0},
		Muted:       RGB{165, 110, 0},
		Text:        RGB{255, 224, 130},
	},
}

var namedColors = map[string]RGB{
	"cyan":        {0, 245, 212},
	"green":       {46, 213, 115},
	"yellow":      {255, 165, 2},
	"red":         {255, 71, 87},
	"magenta":     {196, 113, 237},
	"purple":      {155, 93, 229},
	"pink":        {241, 91, 181},
	"blue":        {0, 187, 249},
	"orange":      {255, 123, 0},
	"gold":        {254, 228, 64},
	"white":       {245, 246, 250},
	"gray":        {130, 140, 160},
	"nes-red":     {228, 0, 88},
	"nes-lava":    {216, 40, 0},
	"nes-blue":    {60, 188, 252},
	"nes-cobalt":  {0, 120, 248},
	"nes-green":   {88, 216, 84},
	"nes-gold":    {248, 184, 0},
	"nes-orange":  {252, 152, 56},
	"nes-magenta": {216, 0, 204},
	"nes-teal":    {0, 232, 216},
	"nes-gray":    {124, 124, 124},
}

// parseLiteralColor resolves a named color or #RRGGBB hex string into an RGB struct.
func parseLiteralColor(spec string, fallback RGB) RGB {
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

// parseColor resolves a semantic role, named color, or #RRGGBB hex string into an RGB struct.
func parseColor(spec string, fallback RGB) RGB {
	s := strings.ToLower(strings.TrimSpace(spec))
	if s == "" {
		return fallback
	}
	switch s {
	case "primary":
		return activePalette().Primary
	case "secondary":
		return activePalette().Secondary
	case "pass", "ok":
		return activePalette().Pass
	case "warn":
		return activePalette().Warn
	case "fail", "err":
		return activePalette().Fail
	case "info":
		return activePalette().Info
	case "muted":
		return activePalette().Muted
	case "text":
		return activePalette().Text
	}
	return parseLiteralColor(s, fallback)
}

// Gradient defines an optional multi-stop TrueColor gradient (used only when --gradient is explicitly set).
type Gradient struct {
	Stops []RGB
}

var namedGradients = map[string]Gradient{
	"cyber": {Stops: []RGB{
		{0, 245, 212},
		{0, 187, 249},
		{155, 93, 229},
	}},
	"nes": {Stops: []RGB{
		{228, 0, 88},
		{248, 184, 0},
		{60, 188, 252},
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

// resolveGradient returns a Gradient only when preset, fromSpec, or toSpec is explicitly provided.
func resolveGradient(preset, fromSpec, toSpec string) (Gradient, bool) {
	if fromSpec != "" || toSpec != "" {
		pal := activePalette()
		c1 := parseColor(fromSpec, pal.Primary)
		c2 := parseColor(toSpec, pal.Secondary)
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
		return activePalette().Primary
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

// renderThemesCatalog displays all built-in color palettes with live color swatches.
func renderThemesCatalog(out io.Writer) {
	pal := activePalette()
	renderStep(out, "THEMES", fmt.Sprintf("%d", len(themeOrder)), "Built-in Solid Color Palettes (Active: "+currentThemeName()+")", "", "")
	fmt.Fprintln(out)

	for _, name := range themeOrder {
		t := builtinThemes[name]
		marker := "  "
		if name == currentThemeName() {
			marker = pal.Pass.fgANSI() + "▸ " + ansiReset()
		}
		swatches := fmt.Sprintf("%s██%s %s██%s %s██%s %s██%s %s██%s %s██%s",
			t.Primary.fgANSI(), ansiReset(),
			t.Secondary.fgANSI(), ansiReset(),
			t.Pass.fgANSI(), ansiReset(),
			t.Warn.fgANSI(), ansiReset(),
			t.Fail.fgANSI(), ansiReset(),
			t.Info.fgANSI(), ansiReset(),
		)
		paddedName := padRight(name, 12)
		fmt.Fprintf(out, "%s%s%s%s%s  %s  %s%s%s\n",
			marker, ansiBold(), t.Primary.fgANSI(), paddedName, ansiReset(),
			swatches,
			ansiDim(), t.Description, ansiReset())
	}
	fmt.Fprintln(out)
	renderCodeBlock(out, "zsh: switch or customize theme", []string{
		`zsp theme nes                          # set active default palette to NES`,
		`zsp banner "NINTENDO" --theme nes      # preview a theme for one command`,
		`zsp config set color.primary "#e40058" # override any role in your config`,
	})
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
