package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
	italic    = "\033[3m"
	underline = "\033[4m"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

func stripANSI(s string) string { return ansiRegex.ReplaceAllString(s, "") }
func visibleLen(s string) int   { return utf8.RuneCountInString(stripANSI(s)) }
func noColorEnabled() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return ok || os.Getenv("TERM") == "dumb"
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

func detectTermWidth() int {
	for _, envName := range []string{"ZSPLASH_WIDTH", "COLUMNS"} {
		if v := strings.TrimSpace(os.Getenv(envName)); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 20 {
				return n
			}
		}
	}
	if runtime.GOOS != "windows" {
		if tty, err := os.Open("/dev/tty"); err == nil {
			defer func() { _ = tty.Close() }()
			cmd := exec.Command("stty", "size")
			cmd.Stdin = tty
			if out, err := cmd.Output(); err == nil {
				if parts := strings.Fields(strings.TrimSpace(string(out))); len(parts) == 2 {
					if cols, err := strconv.Atoi(parts[1]); err == nil && cols >= 20 {
						return cols
					}
				}
			}
		}
	}
	return 80
}

func wrapText(text string, maxWidth int) []string {
	if maxWidth < 10 {
		maxWidth = 10
	}
	var result []string
	for _, raw := range strings.Split(text, "\n") {
		if visibleLen(raw) <= maxWidth {
			result = append(result, raw)
			continue
		}
		words := strings.Fields(raw)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		curr := words[0]
		for _, w := range words[1:] {
			if visibleLen(curr)+1+visibleLen(w) <= maxWidth {
				curr += " " + w
			} else {
				result = append(result, curr)
				curr = w
			}
		}
		result = append(result, curr)
	}
	return result
}

func padRight(s string, targetWidth int) string {
	if vLen := visibleLen(s); vLen < targetWidth {
		return s + strings.Repeat(" ", targetWidth-vLen)
	}
	return s
}

func truncateVisible(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(stripANSI(s))
	if len(runes) <= maxRunes {
		return s
	}
	if maxRunes <= 1 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-1]) + "…"
}

type RGB struct{ R, G, B int }

func (c RGB) Hex() string { return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B) }
func (c RGB) fgANSI() string {
	if noColorEnabled() {
		return ""
	}
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", c.R, c.G, c.B)
}
func (c RGB) bgANSI() string {
	if noColorEnabled() {
		return ""
	}
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", c.R, c.G, c.B)
}

type ThemePalette struct {
	Name, Description                                       string
	Primary, Secondary, Pass, Warn, Fail, Info, Muted, Text RGB
}

var themeOrder = []string{
	"modern", "nes", "xbox", "snes", "n64", "gamecube", "gameboy",
	"ps1", "sega", "pipboy", "cyberpunk", "halo", "doom",
	"dracula", "catppuccin", "nord", "gruvbox", "tokyonight", "amber",
}

var builtinThemes = map[string]ThemePalette{
	"modern":     {"modern", "High-contrast crisp teal, sky blue, and emerald (Default)", RGB{0, 245, 212}, RGB{0, 187, 249}, RGB{46, 213, 115}, RGB{255, 165, 2}, RGB{255, 71, 87}, RGB{56, 189, 248}, RGB{120, 132, 158}, RGB{245, 246, 250}},
	"nes":        {"nes", "NES Control Deck two-tone grey, A/B button red, charcoal, and Zapper orange", RGB{230, 0, 18}, RGB{184, 184, 184}, RGB{88, 216, 84}, RGB{255, 88, 0}, RGB{216, 40, 0}, RGB{255, 88, 0}, RGB{104, 104, 104}, RGB{236, 236, 236}},
	"xbox":       {"xbox", "Xbox neon jewel green, energy lime, matte chassis charcoal, and Red Ring crimson", RGB{24, 230, 24}, RGB{155, 240, 11}, RGB{24, 230, 24}, RGB{255, 185, 0}, RGB{232, 17, 35}, RGB{166, 176, 166}, RGB{92, 102, 92}, RGB{242, 247, 242}},
	"snes":       {"snes", "Super Nintendo deck grey, concave/convex purple & lavender, and 4-button accents", RGB{123, 104, 238}, RGB{178, 164, 255}, RGB{0, 200, 83}, RGB{255, 214, 0}, RGB{229, 57, 53}, RGB{30, 136, 229}, RGB{141, 141, 146}, RGB{244, 244, 246}},
	"n64":        {"n64", "Nintendo 64 charcoal deck with A-blue, B-green, C-yellow, and Start-red buttons", RGB{43, 123, 255}, RGB{255, 208, 0}, RGB{0, 184, 72}, RGB{255, 208, 0}, RGB{230, 32, 32}, RGB{56, 136, 255}, RGB{110, 112, 117}, RGB{236, 238, 242}},
	"gamecube":   {"gamecube", "GameCube Indigo purple, Spice Orange controller, teal A-button, and C-stick yellow", RGB{124, 111, 230}, RGB{255, 112, 0}, RGB{0, 200, 160}, RGB{255, 216, 0}, RGB{230, 30, 42}, RGB{140, 130, 240}, RGB{142, 146, 153}, RGB{242, 240, 255}},
	"gameboy":    {"gameboy", "Game Boy DMG-01 4-shade olive LCD green, grey bezel, and maroon A/B buttons", RGB{155, 188, 15}, RGB{139, 172, 15}, RGB{155, 188, 15}, RGB{139, 172, 15}, RGB{154, 32, 82}, RGB{139, 172, 15}, RGB{123, 125, 123}, RGB{202, 220, 159}},
	"ps1":        {"ps1", "PlayStation 1 battleship grey with Triangle green, Circle red, Cross blue, Square pink", RGB{88, 130, 250}, RGB{224, 130, 180}, RGB{0, 196, 154}, RGB{245, 176, 0}, RGB{240, 58, 62}, RGB{168, 174, 184}, RGB{112, 118, 128}, RGB{236, 239, 244}},
	"sega":       {"sega", "Sega Genesis / Mega Drive matte black, Sonic cobalt blue, gold ring, and 16-bit red", RGB{30, 100, 255}, RGB{255, 215, 0}, RGB{0, 230, 118}, RGB{255, 215, 0}, RGB{229, 37, 33}, RGB{64, 144, 255}, RGB{104, 108, 115}, RGB{245, 247, 250}},
	"pipboy":     {"pipboy", "Fallout Pip-Boy 3000 CRT phosphor green, Vault-Tec amber, and Rad Alert red", RGB{24, 255, 98}, RGB{16, 196, 76}, RGB{24, 255, 98}, RGB{255, 182, 66}, RGB{255, 59, 48}, RGB{20, 224, 86}, RGB{30, 107, 56}, RGB{200, 255, 216}},
	"cyberpunk":  {"cyberpunk", "Night City cyber yellow, Kiroshi optics cyan, and Relic glitch crimson", RGB{252, 238, 10}, RGB{0, 240, 255}, RGB{0, 255, 159}, RGB{252, 238, 10}, RGB{255, 0, 60}, RGB{0, 240, 255}, RGB{122, 111, 155}, RGB{248, 249, 250}},
	"halo":       {"halo", "UNSC Mjolnir Mark V olive drab, Sierra-117 gold visor, and Cortana holo-blue", RGB{120, 158, 78}, RGB{255, 168, 0}, RGB{72, 216, 255}, RGB{255, 168, 0}, RGB{255, 56, 40}, RGB{110, 158, 255}, RGB{90, 102, 80}, RGB{232, 240, 228}},
	"doom":       {"doom", "UAC Hellfire blood red, molten Argent orange, Praetor green, and shotgun steel", RGB{230, 46, 0}, RGB{255, 136, 0}, RGB{92, 230, 92}, RGB{255, 196, 0}, RGB{255, 26, 26}, RGB{77, 166, 255}, RGB{110, 115, 123}, RGB{244, 237, 232}},
	"dracula":    {"dracula", "Dracula dark theme with purple, pink, and neon green accents", RGB{189, 147, 249}, RGB{255, 121, 198}, RGB{80, 250, 123}, RGB{241, 250, 140}, RGB{255, 85, 85}, RGB{139, 233, 253}, RGB{98, 114, 164}, RGB{248, 248, 242}},
	"catppuccin": {"catppuccin", "Catppuccin Mocha soothing pastel mauve, sapphire, and green", RGB{203, 166, 247}, RGB{245, 194, 231}, RGB{166, 227, 161}, RGB{249, 226, 175}, RGB{243, 139, 168}, RGB{137, 220, 235}, RGB{108, 112, 134}, RGB{205, 214, 244}},
	"nord":       {"nord", "Arctic Nord frost blues and aurora borealis status colors", RGB{136, 192, 208}, RGB{129, 161, 193}, RGB{163, 190, 140}, RGB{235, 203, 139}, RGB{191, 97, 106}, RGB{94, 129, 172}, RGB{106, 118, 142}, RGB{236, 239, 244}},
	"gruvbox":    {"gruvbox", "Retro groove warm amber, earthy orange, and olive green", RGB{254, 128, 25}, RGB{250, 189, 47}, RGB{184, 187, 38}, RGB{250, 189, 47}, RGB{251, 73, 52}, RGB{131, 165, 152}, RGB{146, 131, 116}, RGB{235, 219, 178}},
	"tokyonight": {"tokyonight", "Tokyo Night neon blue, cyber purple, and teal", RGB{122, 162, 247}, RGB{187, 154, 247}, RGB{158, 206, 106}, RGB{224, 175, 104}, RGB{247, 118, 142}, RGB{125, 207, 255}, RGB{110, 120, 165}, RGB{192, 202, 245}},
	"amber":      {"amber", "Vintage 1980s monochrome amber phosphor CRT terminal", RGB{255, 176, 0}, RGB{255, 204, 0}, RGB{255, 208, 64}, RGB{255, 144, 0}, RGB{255, 85, 0}, RGB{255, 176, 0}, RGB{165, 110, 0}, RGB{255, 224, 130}},
}

var namedColors = map[string]RGB{
	"cyan": {0, 245, 212}, "green": {46, 213, 115}, "yellow": {255, 165, 2}, "red": {255, 71, 87},
	"magenta": {196, 113, 237}, "purple": {155, 93, 229}, "pink": {241, 91, 181}, "blue": {0, 187, 249},
	"orange": {255, 123, 0}, "gold": {254, 228, 64}, "white": {245, 246, 250}, "gray": {130, 140, 160},
	"nes-red": {230, 0, 18}, "nes-zapper": {255, 88, 0}, "nes-grey": {184, 184, 184}, "nes-gray": {184, 184, 184},
	"nes-darkgrey": {104, 104, 104}, "nes-black": {32, 32, 32}, "nes-lava": {216, 40, 0}, "nes-blue": {60, 188, 252},
	"nes-cobalt": {0, 120, 248}, "nes-green": {88, 216, 84}, "nes-gold": {248, 184, 0}, "nes-orange": {255, 88, 0},
	"xbox-green": {24, 230, 24}, "xbox-lime": {155, 240, 11}, "xbox-red": {232, 17, 35},
}

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
		r, e1 := strconv.ParseInt(s[0:2], 16, 32)
		g, e2 := strconv.ParseInt(s[2:4], 16, 32)
		b, e3 := strconv.ParseInt(s[4:6], 16, 32)
		if e1 == nil && e2 == nil && e3 == nil {
			return RGB{int(r), int(g), int(b)}
		}
	}
	return fallback
}

func parseColor(spec string, fallback RGB) RGB {
	s := strings.ToLower(strings.TrimSpace(spec))
	switch s {
	case "":
		return fallback
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

type Gradient struct{ Stops []RGB }

var namedGradients = map[string]Gradient{
	"cyber":  {[]RGB{{0, 245, 212}, {0, 187, 249}, {155, 93, 229}}},
	"nes":    {[]RGB{{228, 0, 88}, {248, 184, 0}, {60, 188, 252}}},
	"ocean":  {[]RGB{{72, 202, 228}, {0, 119, 182}, {3, 4, 94}}},
	"ember":  {[]RGB{{255, 71, 87}, {255, 123, 0}, {254, 228, 64}}},
	"sunset": {[]RGB{{241, 91, 181}, {255, 123, 0}, {254, 228, 64}}},
	"matrix": {[]RGB{{0, 255, 135}, {46, 213, 115}, {16, 124, 65}}},
	"gold":   {[]RGB{{254, 228, 64}, {255, 165, 2}, {217, 119, 6}}},
	"mono":   {[]RGB{{255, 255, 255}, {180, 190, 210}, {110, 120, 140}}},
}

func resolveGradient(preset, fromSpec, toSpec string) (Gradient, bool) {
	if fromSpec != "" || toSpec != "" {
		pal := activePalette()
		return Gradient{Stops: []RGB{parseColor(fromSpec, pal.Primary), parseColor(toSpec, pal.Secondary)}}, true
	}
	if g, ok := namedGradients[strings.ToLower(strings.TrimSpace(preset))]; ok {
		return g, true
	}
	return Gradient{}, false
}

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
	scaled := t * float64(len(g.Stops)-1)
	idx := int(scaled)
	if idx >= len(g.Stops)-1 {
		idx = len(g.Stops) - 2
	}
	lt := scaled - float64(idx)
	a, b := g.Stops[idx], g.Stops[idx+1]
	return RGB{int(float64(a.R) + lt*float64(b.R-a.R)), int(float64(a.G) + lt*float64(b.G-a.G)), int(float64(a.B) + lt*float64(b.B-a.B))}
}

func applyGradientLine(s string, g Gradient, isBold bool) string {
	if noColorEnabled() {
		return s
	}
	runes := []rune(s)
	if len(runes) == 0 {
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
		if len(runes) > 1 {
			t = float64(i) / float64(len(runes)-1)
		}
		b.WriteString(g.At(t).fgANSI())
		b.WriteRune(r)
	}
	b.WriteString(reset)
	return b.String()
}

type UserConfig struct {
	Theme       string            `json:"theme,omitempty"`
	BannerFont  string            `json:"banner_font,omitempty"`
	BannerColor string            `json:"banner_color,omitempty"`
	BoxBorder   string            `json:"box_border,omitempty"`
	BoxColor    string            `json:"box_color,omitempty"`
	RuleColor   string            `json:"rule_color,omitempty"`
	StepColor   string            `json:"step_color,omitempty"`
	KVColor     string            `json:"kv_color,omitempty"`
	BarColor    string            `json:"bar_color,omitempty"`
	SpinStyle   string            `json:"spin_style,omitempty"`
	SpinColor   string            `json:"spin_color,omitempty"`
	Colors      map[string]string `json:"colors,omitempty"`
}

var activeConfig UserConfig
var cliThemeOverride string

func init() { activeConfig = loadUserConfig() }

func configFilePath() string {
	if custom := strings.TrimSpace(os.Getenv("ZSPLASH_CONFIG")); custom != "" {
		return filepath.Clean(custom)
	}
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		return filepath.Join(xdg, "zsplash", "config.json")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".config", "zsplash", "config.json")
	}
	return "zsplash-config.json"
}

func loadUserConfig() UserConfig {
	cfg := UserConfig{Colors: make(map[string]string)}
	path := filepath.Clean(configFilePath())
	if data, err := os.ReadFile(path); err == nil { // #nosec G304
		_ = json.Unmarshal(data, &cfg)
	}
	if cfg.Colors == nil {
		cfg.Colors = make(map[string]string)
	}
	if envTheme := strings.TrimSpace(os.Getenv("ZSPLASH_THEME")); envTheme != "" {
		cfg.Theme = envTheme
	}
	return cfg
}

func saveUserConfig(cfg UserConfig) error {
	path := filepath.Clean(configFilePath())
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func currentThemeName() string {
	if cliThemeOverride != "" {
		return strings.ToLower(cliThemeOverride)
	}
	if activeConfig.Theme != "" {
		return strings.ToLower(activeConfig.Theme)
	}
	return "modern"
}

func activePalette() ThemePalette {
	pal, ok := builtinThemes[currentThemeName()]
	if !ok {
		pal = builtinThemes["modern"]
	}
	for role, target := range map[string]*RGB{
		"primary": &pal.Primary, "secondary": &pal.Secondary, "pass": &pal.Pass, "warn": &pal.Warn,
		"fail": &pal.Fail, "info": &pal.Info, "muted": &pal.Muted, "text": &pal.Text,
	} {
		if v := activeConfig.Colors[role]; v != "" {
			*target = parseLiteralColor(v, *target)
		}
	}
	return pal
}

func setConfigKey(key, val string) error {
	k, v := strings.ToLower(strings.TrimSpace(key)), strings.TrimSpace(val)
	if activeConfig.Colors == nil {
		activeConfig.Colors = make(map[string]string)
	}
	switch k {
	case "theme":
		if _, ok := builtinThemes[strings.ToLower(v)]; !ok {
			return fmt.Errorf("unknown theme %q (run 'zsp themes' to see palettes)", v)
		}
		activeConfig.Theme = strings.ToLower(v)
	case "banner.font", "banner_font":
		activeConfig.BannerFont = strings.ToLower(v)
	case "banner.color", "banner_color":
		activeConfig.BannerColor = v
	case "box.border", "box_border":
		activeConfig.BoxBorder = strings.ToLower(v)
	case "box.color", "box_color":
		activeConfig.BoxColor = v
	case "rule.color", "rule_color":
		activeConfig.RuleColor = v
	case "step.color", "step_color":
		activeConfig.StepColor = v
	case "kv.color", "kv_color":
		activeConfig.KVColor = v
	case "bar.color", "bar_color":
		activeConfig.BarColor = v
	case "spin.style", "spin_style":
		activeConfig.SpinStyle = strings.ToLower(v)
	case "spin.color", "spin_color":
		activeConfig.SpinColor = v
	case "color.primary", "primary":
		activeConfig.Colors["primary"] = v
	case "color.secondary", "secondary":
		activeConfig.Colors["secondary"] = v
	case "color.pass", "pass", "ok":
		activeConfig.Colors["pass"] = v
	case "color.warn", "warn":
		activeConfig.Colors["warn"] = v
	case "color.fail", "fail", "err":
		activeConfig.Colors["fail"] = v
	case "color.info", "info":
		activeConfig.Colors["info"] = v
	case "color.muted", "muted":
		activeConfig.Colors["muted"] = v
	case "color.text", "text":
		activeConfig.Colors["text"] = v
	default:
		return fmt.Errorf("unknown config key %q", key)
	}
	return saveUserConfig(activeConfig)
}

func handleConfigCommand(out io.Writer, args []string) int {
	if len(args) == 0 || args[0] == "list" || args[0] == "show" {
		renderConfigSummary(out)
		return 0
	}
	switch strings.ToLower(args[0]) {
	case "path":
		fmt.Fprintln(out, configFilePath())
		return 0
	case "reset":
		activeConfig = UserConfig{Colors: make(map[string]string)}
		_ = os.Remove(configFilePath())
		renderPillBadge(out, "PASS", activePalette().Pass, "Configuration reset to defaults", configFilePath())
		return 0
	case "set":
		if len(args) < 3 {
			renderPillBadge(out, "FAIL", activePalette().Fail, "Usage: zsp config set <key> <value>", "")
			return 1
		}
		if err := setConfigKey(args[1], args[2]); err != nil {
			renderPillBadge(out, "FAIL", activePalette().Fail, err.Error(), "")
			return 1
		}
		renderPillBadge(out, "PASS", activePalette().Pass, fmt.Sprintf("Set %s = %s", args[1], args[2]), configFilePath())
		return 0
	default:
		if len(args) == 2 {
			if err := setConfigKey(args[0], args[1]); err != nil {
				renderPillBadge(out, "FAIL", activePalette().Fail, err.Error(), "")
				return 1
			}
			renderPillBadge(out, "PASS", activePalette().Pass, fmt.Sprintf("Set %s = %s", args[0], args[1]), configFilePath())
			return 0
		}
		renderConfigSummary(out)
		return 0
	}
}

func renderConfigSummary(out io.Writer) {
	defOr := func(v, d string) string {
		if v == "" {
			return d + " (default)"
		}
		return v
	}
	pairs := []string{
		"Config File", configFilePath(), "Active Theme", currentThemeName(),
		"banner.font", defOr(activeConfig.BannerFont, "block"),
		"box.border", defOr(activeConfig.BoxBorder, "rounded"),
		"spin.style", defOr(activeConfig.SpinStyle, "dots"),
	}
	var keys []string
	for k := range activeConfig.Colors {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		pairs = append(pairs, "color."+k, activeConfig.Colors[k])
	}
	renderKV(out, "zsplash Configuration & Defaults", "", "", pairs)
	fmt.Fprintln(out)
	renderCodeBlock(out, "zsh: customize defaults", []string{
		`zsp theme nes`,
		`zsp config set banner.font compact`,
		`zsp config set box.border double`,
		`zsp config set color.primary "#3cbcfc"`,
		`zsp config reset`,
	})
}

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
			t.Primary.fgANSI(), ansiReset(), t.Secondary.fgANSI(), ansiReset(), t.Warn.fgANSI(), ansiReset(),
			t.Pass.fgANSI(), ansiReset(), t.Fail.fgANSI(), ansiReset(), t.Muted.fgANSI(), ansiReset())
		fmt.Fprintf(out, "%s%s%s%s%s  %s  %s%s%s\n", marker, ansiBold(), t.Primary.fgANSI(), padRight(name, 12), ansiReset(), swatches, ansiDim(), t.Description, ansiReset())
	}
	fmt.Fprintln(out)
	renderCodeBlock(out, "zsh: switch or customize theme", []string{
		`zsp theme nes`,
		`zsp banner "NINTENDO" --theme nes`,
		`zsp config set color.primary "#e40058"`,
	})
}
