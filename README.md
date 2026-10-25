# zsplash (`zsp`)

Terminal text formatting, banners, boxes, status badges, spinners, and progress bars for Zsh scripts.

## Install

```zsh
brew install acorrow/tap/zsplash
```

Both `zsplash` and the `zsp` shorthand are installed.

## Usage

```zsh
# Visual showcase
zsp demo
zsp demo --theme nes

# Banners (solid color by default; --gradient is opt-in)
zsp banner "DEPLOY" --font block --frame --sub "Production"
zsp banner "STAGE 1" --font compact --color secondary
zsp banner "CYBER" --font line --gradient cyber

# Status badges
zsp ok   "Schema verified"   --meta "42ms"
zsp info "Workers ready"     --meta "8 threads"
zsp warn "Rate limit 85%"    --meta "throttling"
zsp err  "Webhook timed out" --meta "exit 1"

# Layout, cards, trees, and code blocks
zsp step 1 3 "Migrating database"
zsp box "All migrations applied" --title "DB" --border double
zsp kv --title "Target" "Env" "prod" "Branch" "main"
zsp tree --title "Checks" "Auth valid" "Cluster reachable"
zsp cmd "docker build -t app:latest ."

# Spinners (/dev/tty animation keeps stdout clean for variable capture)
SHA=$(zsp spin "Resolving HEAD..." --persist -- git rev-parse --short HEAD)

# Progress bars
for i in {1..50}; do zsp bar $i 50 "Syncing $i" --in-place; sleep 0.02; done
find . -type f | zsp progress --total 100 --title "Indexing"
```

## Using `zsp` in Scripts

- **Text is cleaned by default**: Escape sequences, control characters, and Unicode bidi/format characters are stripped automatically. Only a leading `zsp --raw <cmd>` disables this.
- **Pass untrusted values after `--`**: Use `zsp ok -- "$msg"`, `zsp tree --title Files -- "${files[@]}"`, or `zsp -- "$text"`. Without `--`, values like `--meta` or `--theme` are treated as flags.
- **Always name the subcommand**: `zsp "$var"` treats words like `theme` or `config reset` as commands; use `zsp box -- "$var"` or `zsp -- "$var"`.
- **Exit codes**: Display commands always exit `0`. `spin` passes on the child's exit code (`128 + signal` when killed by a signal). `progress` and `config` exit non-zero on error.
- **`spin` in scripts**:
  - `stdout` stays clean for `$(...)` capture.
  - Pass untrusted titles with `--title`: `zsp spin --title="$var" -- cmd`.
  - Use `--static` for commands that prompt (`sudo`, `ssh`, `git`, `gpg`).
  - Use `--stdin` only for interactive input (piped `stdin` is passed through automatically).
  - Prefer `SIGTERM` over `kill -INT` when a terminal is attached. `SIGHUP` and `SIGQUIT` are also forwarded with a 5s grace period before `SIGKILL` (exiting `129` and `131`).
  - On Linux, `Pdeathsig` (`SIGKILL`) is set on the direct child if `zsp` is killed outright (`SIGKILL`), but it only covers some cases:
    1. `Pdeathsig` only applies to the direct child: if `zsp` is `SIGKILL`ed, the command `spin` started is killed, but anything it started in the background (e.g. `zsp spin -- sh -c 'a & b'`) can keep running. Wrap commands with `exec` where possible (`sh -c 'exec ...'`), or use `SIGTERM`/`SIGHUP`/`SIGINT`, which signal the whole process group when no terminal is attached.
    2. Linux clears `Pdeathsig` when the child runs a setuid/setgid program, so `zsp spin -- sudo ...` (and other setuid programs) is not protected if `zsp` is `SIGKILL`ed. Normal signals (`SIGTERM`, `SIGHUP`, `SIGINT`) still work as documented.
  - macOS has no equivalent of `Pdeathsig`, so `SIGKILL` of `zsp` on macOS can leave the child running when no terminal is attached.
- **Log-friendly output**: `NO_COLOR` and `TERM=dumb` disable color. When no terminal is attached (`CI`, `cron`, logs redirected to a file), `spin` and `progress` skip in-place redraws automatically.

## Themes & Config

Built-in palettes: `modern` (default), `nes`, `snes`, `xbox`, `halo`, `gameboy`, `pipboy`, `dracula`, `catppuccin`, `nord`, `gruvbox`, `tokyonight`, `monokai`, `solarized`, `rosepine`, `everforest`, `kanagawa`, `synthwave`, `industrial`, `amber`.

```zsh
zsp themes                            # list palettes
zsp theme nes                         # save default theme in ~/.config/zsplash/config.json
zsp config set banner.font compact    # override default settings
zsp config set color.primary "#FC7460"
zsp config reset
```

## Reference & Schema

```zsh
zsp help <cmd>     # command flags, stream rules, and examples
zsp docs           # full Markdown reference manual
zsp docs --json    # machine-readable JSON CLI schema
```

## License

MIT
