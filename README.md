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

# Layout, prose, cards, tables, trees, and code blocks
zsp step 1 3 "Migrating database"
zsp box "All migrations applied" --title "DB" --border double
zsp prose --eyebrow "RELEASE" --title "What Changed" "Run `zsp demo` with --theme nes or set $ZSPLASH_WIDTH." --tip "Pipe notes into zsp read."
zsp kv --title "Target" "Env" "prod" "Branch" "main"
zsp tsv --header "SERVICE\tSTATUS\tLATENCY" "api\thealthy\t12ms" "worker\tready\t4ms"
printf "api\thealthy\t12ms\nworker\tready\t4ms\n" | zsp tsv --headers "SERVICE,STATUS,LATENCY"
zsp tree --title "Checks" "Auth valid" "Cluster reachable"
zsp cmd "docker build -t app:latest ."

# Spinners (/dev/tty animation keeps stdout clean for variable capture)
SHA=$(zsp spin "Resolving HEAD..." --persist -- git rev-parse --short HEAD)

# Progress bars
for i in {1..50}; do zsp bar $i 50 "Syncing $i" --in-place; sleep 0.02; done
find . -type f | zsp progress --total 100 --title "Indexing"
```

## Cheat Sheet

### Subcommands

| Command (Aliases) | Syntax | Key Flags | Description |
| :--- | :--- | :--- | :--- |
| `banner` (`big`) | `zsp banner <text> [flags]` | `-f, --font block\|compact\|line`<br>`-c, --color <color>`<br>`-g, --gradient <grad>`<br>`--from <#hex> --to <#hex>`<br>`--sub <text>`<br>`--frame` | 3-row block, 2-row compact, or 3-row box-drawing headline banner |
| `box` | `zsp box <text> [flags]`<br>`zsp -- <text>` | `-t, --title <t>`<br>`--footer <f>`<br>`-b, --border rounded\|double\|heavy\|ascii`<br>`-c, --color <color>`<br>`-g, --gradient <grad>`<br>`-w, --width <cols>` | Word-wrapped callout box sized to terminal width (default subcommand when omitted) |
| `prose` (`read`) | `zsp prose [flags] [paragraphs...]`<br>`<cmd> \| zsp read [flags]` | `-e, --eyebrow <label>`<br>`-t, --title <title>`<br>`--tip <text>` `--note <text>` `--warn <text>`<br>`-f, --file <path>`<br>`-w, --width <cols>`<br>`-c, --color <c>` `-g, --gradient <g>` | High-readability terminal prose capped at a 66-column measure with inline code/flag/env highlights, hanging bullets, and callout rails |
| `ok` (`pass`, `success`)<br>`info`<br>`warn` (`warning`)<br>`err` (`fail`, `error`) | `zsp ok <msg> [--meta <detail>]` | `-m, --meta <detail>` | Inverted status pill badge (`PASS`, `INFO`, `WARN`, `FAIL`) with optional right-aligned dot-leader metadata |
| `step` | `zsp step <curr> <total> <title>` | `-c, --color <color>`<br>`-g, --gradient <grad>` | Full-width phase divider (`━━━ [curr/total] ━━━`) and `▸ <title>` header |
| `rule` (`hr`) | `zsp rule [label] [flags]` | `-c, --color <color>`<br>`-g, --gradient <grad>`<br>`-w, --width <cols>` | Horizontal divider line (`───`) with optional centered label |
| `kv` (`card`) | `zsp kv [flags] <k1> <v1> [<k2> <v2> ...]` | `-t, --title <t>`<br>`-c, --color <color>`<br>`-g, --gradient <grad>` | Framed key-value card with aligned dot-leaders |
| `tsv` (`table`) | `zsp tsv [flags] [rows...]`<br>`<cmd> \| zsp tsv [flags]` | `-H, --header`<br>`--headers <cols>`<br>`-f, --file <path>`<br>`-d, --delim <sep>`<br>`-a, --align <spec>`<br>`--compact` `--no-truncate`<br>`-t, --title <t>` `-b, --border <b>`<br>`-c, --color <c>` `-g, --gradient <g>`<br>`-w, --width <cols>` | Aligned TSV table from `stdin`, file, or args with auto right-aligned numeric columns and proportional narrow-terminal truncation |
| `tree` (`list`) | `zsp tree [flags] <item1> [<item2> ...]` | `-t, --title <t>`<br>`-c, --color <color>` | Unicode checklist (`├─ ●`, `╰─ ●`) with optional root header |
| `cmd` | `zsp cmd <shell command>` | (none) | Shell command preview with colored gutter (`┃ ▸`) and syntax highlighting |
| `code` | `zsp code [flags] <line1> [<line2> ...]` | `-t, --title <t>` | Framed multi-line Zsh code block (`╭─ title ───╮`) |
| `style` | `zsp style <text> [flags]` | `--fg <color>` `--bg <color>`<br>`-g, --gradient <grad>`<br>`--from <#hex> --to <#hex>`<br>`-b, --bold` `--dim`<br>`-i, --italic` `-u, --underline` | Inline styled text |
| `spin` | `zsp spin [title] [flags] [-- <cmd> [args...]]` | `-t, --title <t>`<br>`-s, --style dots\|pulse\|meter\|line`<br>`-c, --color <color>`<br>`-p, --persist`<br>`--static`<br>`--stdin`<br>`-d, --duration <ms>` | Live spinner on `/dev/tty` (`stdout` stays clean for `$(...)`; propagates `<cmd>` exit code) |
| `bar` | `zsp bar <curr> <total> [label] [flags]` | `-r, --in-place`<br>`-w, --width <cols>`<br>`-c, --color <color>`<br>`-g, --gradient <grad>` | Half-cell (`█` / `▌` / `░`) progress bar with `%` and `[curr/total]` |
| `progress` | `<cmd> \| zsp progress [flags]` | `--total <n>`<br>`-t, --title <label>`<br>`-c, --color <color>`<br>`-g, --gradient <grad>` | Stream line counter from `stdin` that renders a live bar on `/dev/tty` |
| `themes` | `zsp themes` | (none) | List all 20 built-in color palettes with color swatches |
| `theme` | `zsp theme [<name> \| list \| preview <name>]` | (none) | Set persistent default theme or preview a theme |
| `config` | `zsp config [list \| set <k> <v> \| reset \| path]` | (none) | Inspect, set, or reset persistent defaults in `~/.config/zsplash/config.json` |
| `demo` | `zsp demo [--theme <name>]` | `-T, --theme <name>` | Render full visual showcase with copyable Zsh snippets |
| `help` / `docs` (`schema`, `manual`) | `zsp help [<cmd>]`<br>`zsp docs [--markdown \| --json \| <cmd>]` | `-j, --json`<br>`--markdown` | Per-command help, full Markdown manual, or JSON CLI schema |

### Global Flags & Environment Variables

| Flag / Variable | Description |
| :--- | :--- |
| `-T, --theme <name>` | Override the active theme for a single command (`zsp --theme nes banner "HI"`) |
| `--raw` | Must be the first argument (`zsp --raw <cmd> ...`): allow raw ANSI escapes in input text |
| `--` | End of flags: treat all remaining arguments as positional text (or child command in `spin`) |
| `-h, --help` / `-v, --version` | Show command help (`zsp <cmd> --help`) or version (`zsp --version`) |
| `ZSPLASH_THEME` | Override active theme for the current shell session |
| `ZSPLASH_CONFIG` | Custom path for `config.json` (default: `~/.config/zsplash/config.json`) |
| `ZSPLASH_WIDTH` | Force terminal column width (`20..500`; default: `/dev/tty` -> `$COLUMNS` -> `80`) |
| `NO_COLOR` / `TERM=dumb` | Disable all ANSI color codes and use plain bracket badges (`[PASS]`, `[WARN]`, `[FAIL]`, `[INFO]`) |

### Colors, Gradients & Config Keys

- **`<color>` values**:
  - **Theme roles**: `primary`, `secondary`, `pass` (`ok`), `warn`, `fail` (`err`), `info`, `muted`, `text`
  - **Named colors**: `cyan`, `green`, `yellow`, `red`, `magenta`, `purple`, `pink`, `blue`, `orange`, `gold`, `white`, `gray`, `nes-red`, `nes-zapper`, `nes-grey`, `nes-darkgrey`, `nes-black`, `nes-lava`, `nes-blue`, `nes-cobalt`, `nes-green`, `nes-gold`, `nes-orange`, `xbox-green`, `xbox-lime`, `xbox-red`
  - **Hex**: `#RRGGBB` (e.g. `#00F5D4`)
- **`<grad>` presets**: `cyber`, `ember`, `ocean`, `sunset`, `matrix`, `gold`, `mono` (or custom `--from "#HEX" --to "#HEX"`)
- **`zsp config set <key> <value>` keys**: `theme`, `banner.font`, `banner.color`, `box.border`, `box.color`, `step.color`, `rule.color`, `kv.color`, `bar.color`, `spin.style`, `spin.color`, `color.primary`, `color.secondary`, `color.pass`, `color.warn`, `color.fail`, `color.info`, `color.muted`, `color.text`

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
