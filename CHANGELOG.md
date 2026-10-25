# Changelog

## v0.6.8 (2026-10-25)

- `flags`, `spin` (bug fix): value-taking flags across all commands (including `-T`/`--theme` and `spin`'s `-s`, `-c`, `-d`, `-t`) never consume `--` as their value. Previously, passing a title such as `"-s"` before `--` (`zsp spin "$step" -- ./deploy.sh`) caused `-s` to consume `--` as its style argument, skipping the child command while exiting `0`.
- `spin`: added `--title <t>` / `--title=<t>` (`-t`) so scripts can pass untrusted spinner titles safely (`zsp spin --title="$step" -- ./deploy.sh`).
- `spin`: forward `SIGHUP` (exit `129`) and `SIGQUIT` (exit `131`) to the child (or child process group when no terminal is attached) with a 5s grace period before `SIGKILL`.
- `spin`: on Linux, set `SysProcAttr.Pdeathsig = syscall.SIGKILL` with `runtime.LockOSThread()` held across `cmd.Start()` through `cmd.Wait()` so the child does not outlive `zsp` if `zsp` is `SIGKILL`ed (documented the macOS limitation where no `Pdeathsig` equivalent exists).

## v0.6.7 (2026-10-25)

- `banner`: do not emit `\033[1m` (`ansiBold`) on `--font line` (or `--font thin`), and keep gradient `box` top borders and `rule` lines unbolded. Prevents VS Code's integrated terminal on macOS from falling back to bold font glyphs for corners (`╭`, `╮`, `╰`, `╯`) and tees (`├`, `┬`, `┴`) while rendering straight segments (`─`, `│`, `╶`, `╴`, `╷`, `╵`) as thin 1px strokes.

## v0.6.6 (2026-10-25)

- `spin`: separated `canAnimate` (`/dev/tty` available or `stderr` is a terminal) from `hasControllingTTY` (`/dev/tty` available) so process-group isolation (`Setpgid`) and `SIGINT` forwarding still apply when `stderr` is a terminal but no controlling terminal is attached.
- `spin`: note on `v0.6.5` non-terminal behavior: in CI and cron (when neither `/dev/tty` nor `stderr` is a terminal), `spin` always prints the title line and a final `[PASS]` badge on `stderr` even without `--persist`; `stdout` is unaffected.
- `tests`: made `/dev/tty` opening replaceable in unit tests so `spin` and `progress` buffer assertions pass when `go test ./...` is run from an interactive terminal as well as in CI, and closed the PTY input pipe in `TestSpinTerminalPromptInPTY` once output arrives so `script` exits immediately.
- `ci`: added top-level `permissions: contents: read` and pinned GitHub Actions to full commit SHAs in both `zsplash` and `homebrew-tap`.

## v0.6.5 (2026-10-25)

- `spin`, `progress`: when `/dev/tty` is unavailable and `stderr` is not a terminal (CI, cron, logs redirected to a file), `spin` behaves as if `--static` were set (printing the title once and the final badge with no 80ms redraws or cursor escapes) and `progress` skips per-line redraws and prints only the final `PASS`/`FAIL` badge.
- `docs`, `README.md`: added a "Using `zsp` in Scripts" guide covering default sanitization, passing untrusted values after `--`, naming the subcommand, exit codes, `spin` flags (`--static`, `--stdin`, `SIGTERM`), and `NO_COLOR`/`TERM=dumb`.
- `help`: use `len(themeOrder)` (`20`) instead of a hard-coded `10` in `zsp help` and `zsp docs`.
- `ci`: added GitHub Actions workflows for `ubuntu-latest` and `macos-latest` builds/tests plus a `windows` cross-build, and Homebrew tap formula audit and build verification.

## v0.6.4 (2026-10-25)

- `spin`: added `--static` for commands that prompt (`sudo`, `ssh`, `git`, `gpg`). Prints the title once without 80ms line redraws so terminal prompts are never overwritten, then prints the final status badge on completion.
- `docs`: updated `zsp help spin`, `zsp docs`, and `zsp docs --json` to document `--static` and the signal/process-group behavior with and without a controlling terminal.
- `tests`: added unit tests for the terminal process-group decision and a PTY integration test (`script`) verifying `spin` child reads from `/dev/tty` without hanging on `SIGTTIN`.

## v0.6.3 (2026-10-25)

- `spin`: keep child processes in the foreground terminal process group whenever `/dev/tty` is available so commands that open `/dev/tty` directly (`sudo`, `ssh`, `git`) never hang with `SIGTTIN`. Separate process groups (`Setpgid`) are used only when no controlling terminal is attached (CI, cron, non-interactive `ssh`).
- `spin`: when a terminal is attached, the first `SIGINT` (`Ctrl-C`) no longer starts a 5-second `SIGKILL` timer, allowing wrapped commands (`terraform`, `git`, database migrations) to finish their cleanup unless a second `Ctrl-C` is pressed. Note: when a terminal is attached, a `SIGINT` sent directly to `zsp`'s PID (via `kill -INT`) is not passed on to the child; send `SIGTERM` instead (which reaches the direct child, followed by `SIGKILL` after 5s).
- `config reset`: when `ZSPLASH_CONFIG` or `~/.config/zsplash/config.json` is a symlink to a valid config file (such as in a dotfiles repository), `config reset` removes the symlink itself rather than deleting the target file.

## v0.6.2 (2026-10-25)

- **Breaking, Security:** Removed the `ZSPLASH_RAW` environment variable; `--raw` remains supported strictly as the first argument before the subcommand (`zsp --raw <cmd> ...`).
- **Security:** `currentThemeName()`: validate `--theme` and `ZSPLASH_THEME` against `builtinThemes` before printing or applying them, falling back to `modern`.
- Added `--` end-of-flags support across all subcommands (`banner`, `box`, `ok`, `warn`, `err`, `info`, `step`, `rule`, `kv`, `tree`, `style`, `bar`, `progress`) as well as top-level `zsp -- <text>`.
- `spin`: automatically pass `stdin` through to the child command when `stdin` is piped or redirected (`!isTerminal(os.Stdin)`), while defaulting to `/dev/null` when `stdin` is an interactive terminal unless `--stdin` is passed.
- `spin`: place non-terminal child commands in their own process group on Unix (`Setpgid`) and forward `SIGINT`/`SIGTERM` to the process group so background grandchildren are cleaned up.
- `config`: resolve symlinks (`filepath.EvalSymlinks`) when reading, validating, and atomically updating `config.json` so symlinked dotfile configs remain intact on `config set`.

## v0.6.1 (2026-10-25)

- **Breaking:** `spin` defaults child `stdin` to `/dev/null` unless `--stdin` is passed so unexpected interactive prompts do not hang under spinner redraws (relaxed in `v0.6.2` to pass piped or redirected `stdin` through automatically).
- **Security:** Restrict `--raw` to precede the subcommand so positional arguments cannot disable sanitization, strip ANSI sequences inside `visibleLen` even in `--raw` mode, strip Unicode format/bidirectional control characters (`unicode.Cf`, preserving `U+200D` ZWJ), and reject signed hex color strings.
- **Security:** Harden `ZSPLASH_CONFIG` validation (require regular file, top-level `{` JSON object, `DisallowUnknownFields`, and no trailing data) and write config updates atomically via temporary file rename.
- `spin`: register signal handlers before starting the child process and return `128 + signal` exit codes (`130` on `SIGINT`, `143` on `SIGTERM`).

## v0.6.0 (2026-10-25)

- **Security:** Sanitize CSI, OSC (including `OSC 52` clipboard writes and window title escapes), DCS/PM/APC, and `C0`/`DEL`/`C1` control characters from user-supplied text at all render boundaries, with `--raw` opt-out.
- `spin`: stream child `stderr` directly instead of buffering in memory.
- `progress`: expand `bufio.Scanner` buffer to 10 MB and return exit code `1` with a `FAIL` badge on scanner errors.
- Resolve `stty` only via `/bin/stty` or `/usr/bin/stty`, cache detected terminal width for 250ms, and clamp widths to `20..500` columns.

## v0.5.2 (2026-10-25)

- Curated 20 built-in color palettes: `modern`, `nes`, `snes`, `xbox`, `halo`, `gameboy`, `pipboy`, `dracula`, `catppuccin`, `nord`, `gruvbox`, `tokyonight`, `monokai`, `solarized`, `rosepine`, `everforest`, `kanagawa`, `synthwave`, `industrial`, and `amber`.
- Added `zsp step`, `zsp rule`, and `zsp cmd` examples to `zsp demo`.

## v0.4.0 (2026-10-25)

- Added built-in technical reference manual (`zsp docs`), machine-readable JSON CLI schema (`zsp docs --json`), and subcommand help (`zsp help <cmd>` / `zsp <cmd> --help`).

## v0.3.0 (2026-10-25)

- Changed all components to use solid colors from the active palette by default (`--gradient` is opt-in).
- Added persistent user configuration (`zsp theme <name>`, `zsp config set <key> <value>`, `zsp config reset`).
