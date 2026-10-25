package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// UserConfig stores persistent user overrides in ~/.config/zsplash/config.json.
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

func init() {
	activeConfig = loadUserConfig()
}

// configFilePath resolves the path to config.json, respecting ZSPLASH_CONFIG and XDG_CONFIG_HOME.
func configFilePath() string {
	if custom := strings.TrimSpace(os.Getenv("ZSPLASH_CONFIG")); custom != "" {
		return custom
	}
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		return filepath.Join(xdg, "zsplash", "config.json")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "zsplash-config.json"
	}
	return filepath.Join(home, ".config", "zsplash", "config.json")
}

func loadUserConfig() UserConfig {
	cfg := UserConfig{
		Colors: make(map[string]string),
	}
	path := configFilePath()
	data, err := os.ReadFile(path)
	if err == nil {
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
	path := configFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// currentThemeName returns the active theme name from CLI flag, env, config, or "modern".
func currentThemeName() string {
	if cliThemeOverride != "" {
		return strings.ToLower(cliThemeOverride)
	}
	if activeConfig.Theme != "" {
		return strings.ToLower(activeConfig.Theme)
	}
	return "modern"
}

// activePalette returns the resolved ThemePalette including any user role overrides.
func activePalette() ThemePalette {
	name := currentThemeName()
	pal, ok := builtinThemes[name]
	if !ok {
		pal = builtinThemes["modern"]
	}
	if v, ok := activeConfig.Colors["primary"]; ok && v != "" {
		pal.Primary = parseLiteralColor(v, pal.Primary)
	}
	if v, ok := activeConfig.Colors["secondary"]; ok && v != "" {
		pal.Secondary = parseLiteralColor(v, pal.Secondary)
	}
	if v, ok := activeConfig.Colors["pass"]; ok && v != "" {
		pal.Pass = parseLiteralColor(v, pal.Pass)
	}
	if v, ok := activeConfig.Colors["warn"]; ok && v != "" {
		pal.Warn = parseLiteralColor(v, pal.Warn)
	}
	if v, ok := activeConfig.Colors["fail"]; ok && v != "" {
		pal.Fail = parseLiteralColor(v, pal.Fail)
	}
	if v, ok := activeConfig.Colors["info"]; ok && v != "" {
		pal.Info = parseLiteralColor(v, pal.Info)
	}
	if v, ok := activeConfig.Colors["muted"]; ok && v != "" {
		pal.Muted = parseLiteralColor(v, pal.Muted)
	}
	if v, ok := activeConfig.Colors["text"]; ok && v != "" {
		pal.Text = parseLiteralColor(v, pal.Text)
	}
	return pal
}

// setConfigKey updates a configuration key and writes config.json.
func setConfigKey(key, val string) error {
	k := strings.ToLower(strings.TrimSpace(key))
	v := strings.TrimSpace(val)
	if activeConfig.Colors == nil {
		activeConfig.Colors = make(map[string]string)
	}

	switch k {
	case "theme":
		if _, ok := builtinThemes[strings.ToLower(v)]; !ok {
			return fmt.Errorf("unknown theme %q (run 'zsp themes' to see available palettes)", v)
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
	sub := strings.ToLower(args[0])
	switch sub {
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
	font := activeConfig.BannerFont
	if font == "" {
		font = "block (default)"
	}
	border := activeConfig.BoxBorder
	if border == "" {
		border = "rounded (default)"
	}
	spinStyle := activeConfig.SpinStyle
	if spinStyle == "" {
		spinStyle = "dots (default)"
	}

	pairs := []string{
		"Config File", configFilePath(),
		"Active Theme", currentThemeName(),
		"banner.font", font,
		"box.border", border,
		"spin.style", spinStyle,
	}
	if activeConfig.BannerColor != "" {
		pairs = append(pairs, "banner.color", activeConfig.BannerColor)
	}
	if activeConfig.BoxColor != "" {
		pairs = append(pairs, "box.color", activeConfig.BoxColor)
	}
	if activeConfig.BarColor != "" {
		pairs = append(pairs, "bar.color", activeConfig.BarColor)
	}

	var colorKeys []string
	for k := range activeConfig.Colors {
		colorKeys = append(colorKeys, k)
	}
	sort.Strings(colorKeys)
	for _, k := range colorKeys {
		pairs = append(pairs, "color."+k, activeConfig.Colors[k])
	}

	renderKV(out, "zsplash Configuration & Defaults", "", "", pairs)
	fmt.Fprintln(out)
	renderCodeBlock(out, "zsh: customize defaults", []string{
		`zsp theme nes                          # switch default palette to NES`,
		`zsp config set banner.font compact     # set default banner font`,
		`zsp config set box.border double       # set default box border`,
		`zsp config set color.primary "#3cbcfc" # override primary theme color`,
		`zsp config reset                       # restore factory defaults`,
	})
}
