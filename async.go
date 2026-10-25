package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var spinnerStyles = map[string][]string{
	"dots":  {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	"pulse": {"◜", "◠", "◝", "◞", "◡", "◟"},
	"meter": {"▱▱▱", "▰▱▱", "▰▰▱", "▰▰▰", "▱▰▰", "▱▱▰"},
	"line":  {"-", "\\", "|", "/"},
}

type SpinOptions struct {
	Title      string
	Style      string
	Color      string
	Persist    bool
	DurationMs int
	Command    []string
}

// openTTYWriter tries to open /dev/tty so animations do not pollute captured stdout.
func openTTYWriter(fallback io.Writer) (io.Writer, func()) {
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err == nil {
		return tty, func() { _ = tty.Close() }
	}
	return fallback, func() {}
}

// runSpinner animates a terminal spinner while executing a child command or sleeping for DurationMs.
func runSpinner(stdout, stderr io.Writer, opts SpinOptions) int {
	ttyOut, closeTTY := openTTYWriter(stderr)
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
	title := strings.TrimSpace(opts.Title)
	if title == "" {
		title = "Working..."
	}

	start := time.Now()
	doneCh := make(chan error, 1)
	var cmdStderr bytes.Buffer
	var cmd *exec.Cmd

	if len(opts.Command) > 0 {
		cmd = exec.Command(opts.Command[0], opts.Command[1:]...)
		cmd.Stdout = stdout
		cmd.Stderr = &cmdStderr
		cmd.Stdin = os.Stdin
		if err := cmd.Start(); err != nil {
			renderPillBadge(ttyOut, "FAIL", pal.Fail, title, err.Error())
			return 1
		}
		go func() {
			doneCh <- cmd.Wait()
		}()
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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// Hide cursor while spinning.
	if !noColorEnabled() {
		fmt.Fprint(ttyOut, "\033[?25l")
	}
	restoreTerminal := func() {
		if !noColorEnabled() {
			fmt.Fprint(ttyOut, "\r\033[2K\033[?25h")
		} else {
			fmt.Fprint(ttyOut, "\r")
		}
	}

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	frameIdx := 0
	for {
		select {
		case <-sigCh:
			if cmd != nil && cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			restoreTerminal()
			renderPillBadge(ttyOut, "INT", pal.Warn, title, "interrupted")
			return 130

		case err := <-doneCh:
			restoreTerminal()
			elapsed := formatDuration(time.Since(start))
			if err != nil {
				renderPillBadge(ttyOut, "FAIL", pal.Fail, title, elapsed)
				if cmdStderr.Len() > 0 {
					fmt.Fprint(stderr, cmdStderr.String())
				}
				if exitErr, ok := err.(*exec.ExitError); ok {
					return exitErr.ExitCode()
				}
				return 1
			}
			if opts.Persist {
				renderPillBadge(ttyOut, "PASS", pal.Pass, title, elapsed)
			}
			return 0

		case <-ticker.C:
			termWidth := detectTermWidth()
			elapsed := formatDuration(time.Since(start))
			frame := frames[frameIdx%len(frames)]
			frameIdx++

			maxTitleLen := termWidth - visibleLen(frame) - visibleLen(elapsed) - 8
			if maxTitleLen < 8 {
				maxTitleLen = 8
			}
			clippedTitle := truncateVisible(title, maxTitleLen)
			fmt.Fprintf(ttyOut, "\r\033[2K%s%s%s%s %s%s%s %s(%s)%s",
				ansiBold(), c.fgANSI(), frame, ansiReset(),
				ansiBold(), clippedTitle, ansiReset(),
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
	Current  int
	Total    int
	Label    string
	Width    int
	Color    string
	Gradient string
	InPlace  bool
}

// fractionalBlocks provides 1/8th cell horizontal resolution for smooth progress bars.
var fractionalBlocks = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"}

// formatProgressBar builds a single-line progress bar fitted to the current terminal width.
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
		labelPrefix = truncateVisible(opts.Label, maxLabel) + " "
	}

	barCells := opts.Width
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
	filledLen := len([]rune(filledStr))
	emptyLen := barCells - filledLen
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

	var coloredFilled string
	if useGrad {
		coloredFilled = applyGradientLine(filledStr, grad, false)
	} else {
		coloredFilled = c.fgANSI() + filledStr + ansiReset()
	}

	return fmt.Sprintf("%s%s%s%s%s%s%s%s%s",
		ansiBold(), labelPrefix, ansiReset(),
		coloredFilled,
		ansiDim(), emptyStr, ansiReset(),
		ansiBold(), statsSuffix+ansiReset())
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

// runStreamProgress reads lines from stdin and updates an in-place progress bar on the terminal.
func runStreamProgress(in io.Reader, out io.Writer, total int, title, colorSpec, gradSpec string) {
	ttyOut, closeTTY := openTTYWriter(out)
	defer closeTTY()

	if total <= 0 {
		total = 100
	}
	if title == "" {
		title = "Processing"
	}

	start := time.Now()
	scanner := bufio.NewScanner(in)
	count := 0

	renderBar(ttyOut, BarOptions{
		Current:  0,
		Total:    total,
		Label:    title,
		Color:    colorSpec,
		Gradient: gradSpec,
		InPlace:  true,
	})

	for scanner.Scan() {
		count++
		curr := count
		if curr > total {
			total = curr
		}
		renderBar(ttyOut, BarOptions{
			Current:  curr,
			Total:    total,
			Label:    title,
			Color:    colorSpec,
			Gradient: gradSpec,
			InPlace:  true,
		})
	}

	fmt.Fprint(ttyOut, "\r\033[2K")
	renderPillBadge(ttyOut, "PASS", activePalette().Pass, title, fmt.Sprintf("%d items in %s", count, formatDuration(time.Since(start))))
}
