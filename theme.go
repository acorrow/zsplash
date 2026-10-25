package main

import (
	"bytes"
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
	"time"
	"unicode/utf8"
)

const (
	reset       = "\033[0m"
	bold        = "\033[1m"
	dim         = "\033[2m"
	italic      = "\033[3m"
	underline   = "\033[4m"
	maxTermCols = 500
)

// ansiRegex matches CSI, OSC (BEL or ST terminated), DCS/SOS/PM/APC, and 2-byte ESC sequences.
var ansiRegex = regexp.MustCompile(`\x1b(?:\[[0-9:;<=>?]*[ -/]*[@-~]|\][^\x07\x1b]*(?:\x07|\x1b\\)|[PX^_][^\x1b]*\x1b\\|[@-Z\\-_])`)

var rawOutputEnabled bool
var cachedWidth int
var cachedWidthAt time.Time

func stripANSI(s string) string { return ansiRegex.ReplaceAllString(s, "") }

// sanitizeText removes terminal escape sequences and C0/DEL/C1 control characters from user input
// unless --raw is explicitly enabled.
func sanitizeText(s string, allowNewline bool) string {
	if rawOutputEnabled {
		return s
	}
	cleaned := stripANSI(s)
	var b strings.Builder
	b.Grow(len(cleaned))
	for _, r := range cleaned {
		switch {
		case r == '\n':
			if allowNewline {
				b.WriteRune('\n')
			} else {
				b.WriteRune(' ')
			}
		case r == '\t':
			b.WriteString("  ")
		case r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f):
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func visibleLen(s string) int { return utf8.RuneCountInString(sanitizeText(s, true)) }

func clampWidth(w, minW, maxW int) int {
	if w < minW {
		return minW
	}
	if w > maxW {
		return maxW
	}
	return w
}

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
				return clampWidth(n, 20, maxTermCols)
			}
		}
	}
	if cachedWidth >= 20 && time.Since(cachedWidthAt) < 250*time.Millisecond {
		return cachedWidth
	}
	cols := 80
	if runtime.GOOS != "windows" {
		sttyPath := ""
		for _, candidate := range []string{"/bin/stty", "/usr/bin/stty"} {
			if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
				sttyPath = candidate
				break
			}
		}
		if sttyPath != "" {
			if tty, err := os.Open("/dev/tty"); err == nil {
				defer func() { _ = tty.Close() }()
				cmd := exec.Command(sttyPath, "size") // #nosec G204
				cmd.Stdin = tty
				if out, err := cmd.Output(); err == nil {
					if parts := strings.Fields(strings.TrimSpace(string(out))); len(parts) == 2 {
						if n, err := strconv.Atoi(parts[1]); err == nil && n >= 20 {
							cols = clampWidth(n, 20, maxTermCols)
						}
					}
				}
			}
		}
	}
	cachedWidth = cols
	cachedWidthAt = time.Now()
	return cols
}

func wrapText(text string, maxWidth int) []string {
	maxWidth = clampWidth(maxWidth, 10, maxTermCols)
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
	targetWidth = clampWidth(targetWidth, 0, maxTermCols)
	if vLen := visibleLen(s); vLen < targetWidth {
		return s + strings.Repeat(" ", targetWidth-vLen)
	}
	return s
}

func truncateVisible(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(sanitizeText(s, false))
	if len(runes) <= maxRunes {
		return string(runes)
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
	"modern", "nes", "snes", "xbox", "halo", "gameboy", "pipboy",
	"dracula", "catppuccin", "nord", "gruvbox", "tokyonight",
	"monokai", "solarized", "rosepine", "everforest", "kanagawa",
	"synthwave", "industrial", "amber",
}

var builtinThemes = map[string]ThemePalette{
	"modern":     {"modern", "Crisp teal, sky blue, and emerald (Default)", RGB{0, 245, 212}, RGB{0, 187, 249}, RGB{46, 213, 115}, RGB{255, 165, 2}, RGB{255, 71, 87}, RGB{56, 189, 248}, RGB{120, 132, 158}, RGB{245, 246, 250}},
	"nes":        {"nes", "Two-tone deck grey, A/B red, charcoal, and Zapper orange", RGB{230, 0, 18}, RGB{184, 184, 184}, RGB{88, 216, 84}, RGB{255, 88, 0}, RGB{216, 40, 0}, RGB{255, 88, 0}, RGB{104, 104, 104}, RGB{236, 236, 236}},
	"snes":       {"snes", "Deck grey, concave/convex purple & lavender, and 4-color accents", RGB{123, 104, 238}, RGB{178, 164, 255}, RGB{0, 200, 83}, RGB{255, 214, 0}, RGB{229, 57, 53}, RGB{30, 136, 229}, RGB{141, 141, 146}, RGB{244, 244, 246}},
	"xbox":       {"xbox", "Neon jewel green, energy lime, matte charcoal, and crimson", RGB{24, 230, 24}, RGB{155, 240, 11}, RGB{24, 230, 24}, RGB{255, 185, 0}, RGB{232, 17, 35}, RGB{166, 176, 166}, RGB{92, 102, 92}, RGB{242, 247, 242}},
	"halo":       {"halo", "Mjolnir olive drab, visor gold, shield cyan, and holo-blue", RGB{120, 158, 78}, RGB{255, 168, 0}, RGB{72, 216, 255}, RGB{255, 168, 0}, RGB{255, 56, 40}, RGB{110, 158, 255}, RGB{90, 102, 80}, RGB{232, 240, 228}},
	"gameboy":    {"gameboy", "DMG-01 4-shade olive LCD green, grey bezel, and maroon accents", RGB{155, 188, 15}, RGB{139, 172, 15}, RGB{155, 188, 15}, RGB{139, 172, 15}, RGB{154, 32, 82}, RGB{139, 172, 15}, RGB{123, 125, 123}, RGB{202, 220, 159}},
	"pipboy":     {"pipboy", "Phosphor green CRT, warm amber warning, and crimson alert", RGB{24, 255, 98}, RGB{16, 196, 76}, RGB{24, 255, 98}, RGB{255, 182, 66}, RGB{255, 59, 48}, RGB{20, 224, 86}, RGB{30, 107, 56}, RGB{200, 255, 216}},
	"dracula":    {"dracula", "Dark purple, neon pink, cyan, and spring green", RGB{189, 147, 249}, RGB{255, 121, 198}, RGB{80, 250, 123}, RGB{241, 250, 140}, RGB{255, 85, 85}, RGB{139, 233, 253}, RGB{98, 114, 164}, RGB{248, 248, 242}},
	"catppuccin": {"catppuccin", "Catppuccin Mocha pastel mauve, sapphire, peach, and green", RGB{203, 166, 247}, RGB{245, 194, 231}, RGB{166, 227, 161}, RGB{249, 226, 175}, RGB{243, 139, 168}, RGB{137, 220, 235}, RGB{108, 112, 134}, RGB{205, 214, 244}},
	"nord":       {"nord", "Arctic frost blues and aurora borealis status colors", RGB{136, 192, 208}, RGB{129, 161, 193}, RGB{163, 190, 140}, RGB{235, 203, 139}, RGB{191, 97, 106}, RGB{94, 129, 172}, RGB{106, 118, 142}, RGB{236, 239, 244}},
	"gruvbox":    {"gruvbox", "Warm retro earth tones, burnt orange, and olive green", RGB{254, 128, 25}, RGB{250, 189, 47}, RGB{184, 187, 38}, RGB{250, 189, 47}, RGB{251, 73, 52}, RGB{131, 165, 152}, RGB{146, 131, 116}, RGB{235, 219, 178}},
	"tokyonight": {"tokyonight", "Tokyo Night neon indigo, violet, and dusk cyan", RGB{122, 162, 247}, RGB{187, 154, 247}, RGB{158, 206, 106}, RGB{224, 175, 104}, RGB{247, 118, 142}, RGB{125, 207, 255}, RGB{110, 120, 165}, RGB{192, 202, 245}},
	"monokai":    {"monokai", "Warm charcoal with neon magenta, lime green, orange, and cyan", RGB{249, 38, 114}, RGB{102, 217, 239}, RGB{166, 226, 46}, RGB{253, 151, 31}, RGB{249, 38, 114}, RGB{174, 129, 255}, RGB{117, 113, 94}, RGB{248, 248, 242}},
	"solarized":  {"solarized", "Precision cyan, cobalt blue, warm yellow, and magenta", RGB{42, 161, 152}, RGB{38, 139, 210}, RGB{133, 153, 0}, RGB{181, 137, 0}, RGB{220, 50, 47}, RGB{108, 113, 196}, RGB{101, 123, 131}, RGB{238, 232, 213}},
	"rosepine":   {"rosepine", "Soho pine teal, muted rose, iris lavender, and warm gold", RGB{235, 188, 186}, RGB{196, 167, 231}, RGB{49, 116, 143}, RGB{246, 193, 119}, RGB{235, 111, 146}, RGB{156, 207, 216}, RGB{110, 106, 134}, RGB{224, 222, 244}},
	"everforest": {"everforest", "Deep woodland sage green, warm aqua, and soft terracotta", RGB{167, 192, 128}, RGB{127, 187, 179}, RGB{167, 192, 128}, RGB{219, 188, 127}, RGB{230, 126, 128}, RGB{131, 192, 146}, RGB{133, 146, 137}, RGB{211, 198, 170}},
	"kanagawa":   {"kanagawa", "Wave ink blue, samurai red, Onigiri purple, and autumn gold", RGB{126, 156, 216}, RGB{149, 127, 184}, RGB{152, 187, 108}, RGB{230, 195, 132}, RGB{228, 104, 118}, RGB{127, 180, 202}, RGB{114, 113, 105}, RGB{220, 215, 186}},
	"synthwave":  {"synthwave", "Outrun sunset magenta, electric cyan, laser violet, and gold", RGB{255, 42, 133}, RGB{0, 240, 255}, RGB{54, 249, 246}, RGB{254, 222, 93}, RGB{254, 68, 80}, RGB{184, 147, 255}, RGB{132, 115, 156}, RGB{255, 245, 252}},
	"industrial": {"industrial", "Safety orange, caution yellow, brushed steel grey, and oxide red", RGB{255, 103, 0}, RGB{168, 178, 188}, RGB{82, 196, 26}, RGB{250, 173, 20}, RGB{217, 48, 37}, RGB{140, 156, 172}, RGB{102, 110, 118}, RGB{236, 240, 244}},
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
	return ""
}

// validateExistingConfigFile ensures that if path already exists on disk, it is a valid zsplash JSON config
// (preventing accidental overwrite or deletion of arbitrary files via ZSPLASH_CONFIG).
func validateExistingConfigFile(path string) error {
	if path == "" {
		return fmt.Errorf("could not resolve config path ($HOME and $ZSPLASH_CONFIG are unset)")
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.IsDir() || info.Size() > 64*1024 {
		return fmt.Errorf("refusing to modify non-config path %q", path)
	}
	data, err := os.ReadFile(filepath.Clean(path)) // #nosec G304
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	var probe UserConfig
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&probe); err != nil {
		return fmt.Errorf("refusing to modify %q: existing file is not a valid zsplash config", path)
	}
	return nil
}

func loadUserConfig() UserConfig {
	cfg := UserConfig{Colors: make(map[string]string)}
	if rawPath := configFilePath(); rawPath != "" {
		path := filepath.Clean(rawPath)
		if info, err := os.Stat(path); err == nil && !info.IsDir() && info.Size() <= 64*1024 {
			if data, err := os.ReadFile(path); err == nil { // #nosec G304
				_ = json.Unmarshal(data, &cfg)
			}
		}
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
	rawPath := configFilePath()
	if err := validateExistingConfigFile(rawPath); err != nil {
		return err
	}
	path := filepath.Clean(rawPath)
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
		rawPath := configFilePath()
		if rawPath != "" {
			if err := validateExistingConfigFile(rawPath); err != nil {
				renderPillBadge(out, "FAIL", activePalette().Fail, err.Error(), "")
				return 1
			}
			_ = os.Remove(filepath.Clean(rawPath))
		}
		activeConfig = UserConfig{Colors: make(map[string]string)}
		renderPillBadge(out, "PASS", activePalette().Pass, "Configuration reset to defaults", rawPath)
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
