package main

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf8"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

// stripANSI removes all ANSI escape sequences from a string.
func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// visibleLen returns the number of printable terminal runes in s, ignoring ANSI codes.
func visibleLen(s string) int {
	return utf8.RuneCountInString(stripANSI(s))
}

// noColorEnabled returns true when NO_COLOR is set or TERM=dumb.
func noColorEnabled() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return true
	}
	return os.Getenv("TERM") == "dumb"
}

// detectTermWidth returns the current terminal column width, clamped to a safe range.
// It checks ZSPLASH_WIDTH, COLUMNS, and live /dev/tty stty size on Unix.
func detectTermWidth() int {
	for _, envName := range []string{"ZSPLASH_WIDTH", "COLUMNS"} {
		if v := strings.TrimSpace(os.Getenv(envName)); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 20 {
				return n
			}
		}
	}

	if runtime.GOOS != "windows" {
		tty, err := os.Open("/dev/tty")
		if err == nil {
			defer tty.Close()
			cmd := exec.Command("stty", "size")
			cmd.Stdin = tty
			if out, err := cmd.Output(); err == nil {
				parts := strings.Fields(strings.TrimSpace(string(out)))
				if len(parts) == 2 {
					if cols, err := strconv.Atoi(parts[1]); err == nil && cols >= 20 {
						return cols
					}
				}
			}
		}
	}

	return 80
}

// wrapText splits text into lines where each line's visible length is at most maxWidth.
func wrapText(text string, maxWidth int) []string {
	if maxWidth < 10 {
		maxWidth = 10
	}
	rawLines := strings.Split(text, "\n")
	var result []string

	for _, raw := range rawLines {
		if visibleLen(raw) <= maxWidth {
			result = append(result, raw)
			continue
		}
		words := strings.Fields(raw)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		current := words[0]
		for _, w := range words[1:] {
			if visibleLen(current)+1+visibleLen(w) <= maxWidth {
				current += " " + w
			} else {
				result = append(result, current)
				current = w
			}
		}
		result = append(result, current)
	}
	return result
}

// padRight pads s with spaces so its visible rune length equals targetWidth.
func padRight(s string, targetWidth int) string {
	vLen := visibleLen(s)
	if vLen >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-vLen)
}

// truncateVisible truncates a plain string to maxRunes, adding an ellipsis if needed.
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
