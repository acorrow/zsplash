# Changelog

## v0.6.3 (2026-10-25)

- `spin`: keep child processes in the foreground terminal process group whenever `/dev/tty` is available so commands that open `/dev/tty` directly (`sudo`, `ssh`, `git`) never hang with `SIGTTIN`. Separate process groups (`Setpgid`) are used only when no controlling terminal is attached (CI, cron, non-interactive `ssh`).
- `spin`: when a terminal is attached, the first `SIGINT` (`Ctrl-C`) no longer starts a 5-second `SIGKILL` timer, allowing wrapped commands (`terraform`, `git`, database migrations) to finish their cleanup unless a second `Ctrl-C` is pressed.
- `config reset`: when `ZSPLASH_CONFIG` or `~/.config/zsplash/config.json` is a symlink to a valid config file (such as in a dotfiles repository), `config reset` removes the symlink itself rather than deleting the target file.

## v0.6.2 (2026-10-25)

- Removed the `ZSPLASH_RAW` environment variable; `--raw` remains supported strictly before the subcommand (`zsp --raw <cmd> ...`).
- `currentThemeName()`: validate `--theme` and `ZSPLASH_THEME` against `builtinThemes` before printing or applying them, falling back to `modern`.
- Added `--` end-of-flags support across all subcommands (`banner`, `box`, `ok`, `warn`, `err`, `info`, `step`, `rule`, `kv`, `tree`, `style`, `bar`, `progress`) as well as top-level `zsp -- <text>`.
- `spin`: automatically pass `stdin` through to the child command when `stdin` is piped or redirected (`!isTerminal(os.Stdin)`), while defaulting to `/dev/null` when `stdin` is an interactive terminal unless `--stdin` is passed.
- `spin`: place non-terminal child commands in their own process group on Unix (`Setpgid`) and forward `SIGINT`/`SIGTERM` to the process group so background grandchildren are cleaned up.
- `config`: resolve symlinks (`filepath.EvalSymlinks`) when reading, validating, and atomically updating `config.json` so symlinked dotfile configs remain intact on `config set`.

## v0.6.1 (2026-10-25)

- `spin`: register signal handlers before starting the child process, return `128 + signal` exit codes (`130` on `SIGINT`, `143` on `SIGTERM`), and default interactive child `stdin` to `/dev/null` unless `--stdin` is passed.
- Restrict `--raw` to precede the subcommand so positional arguments cannot disable sanitization, and strip ANSI sequences inside `visibleLen` even in `--raw` mode.
- Strip Unicode format and bidirectional control characters (`unicode.Cf`, preserving `U+200D` ZWJ) and reject signed hex color strings.
- Harden `ZSPLASH_CONFIG` validation (require regular file, top-level `{` JSON object, `DisallowUnknownFields`, and no trailing data) and write config updates atomically via temporary file rename.

## v0.6.0 (2026-10-25)

- Sanitize CSI, OSC (including `OSC 52` clipboard writes and window title escapes), DCS/PM/APC, and `C0`/`DEL`/`C1` control characters from user-supplied text at all render boundaries, with `--raw` opt-out.
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
