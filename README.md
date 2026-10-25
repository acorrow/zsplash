# zsplash (`zsp`)

Fast, reactive CLI utility for Zsh and shell scripts to render colorful banners, boxes, status badges, phase headers, key-value cards, command previews, spinners, and progress bars.

Both `zsplash` and its short alias **`zsp`** are installed out of the box. Every component automatically adapts to the live terminal width (`$COLUMNS` / `/dev/tty`) so resizing your terminal never scrambles borders or animations.

---

## Installation

### Homebrew (macOS & Linux)

```zsh
brew install acorrow/tap/zsplash
```

### Upgrade to Latest Release

```zsh
brew update && brew upgrade zsplash
```

---

## Built-in Reference & Schema (`zsp docs`, `zsp help`)

`zsplash` ships with a built-in technical specification and JSON schema directly inside the binary:

```zsh
# View overview of all commands
zsp --help

# View the complete specification, stream contract, flags, and examples for any command
zsp help spin
zsp banner --help
zsp kv --help

# Emit the complete single-page Markdown technical reference manual to stdout
zsp docs

# Emit the complete machine-readable JSON CLI schema (all commands, flags, themes, colors, config keys)
zsp docs --json
```

---

## 1. Core Execution & Stream Contracts

1. **Binary Names**: `zsplash` and `zsp` are identical binaries (`zsp` is installed as a symlink next to `zsplash`).
2. **Solid Colors by Default**: Every component renders solid colors from the active palette by default. Gradients are strictly opt-in via `--gradient <name>` (or `--from #RRGGBB --to #RRGGBB`).
3. **Standard Decoration Commands (`stdout`)**: `banner`, `box`, `ok`, `info`, `warn`, `err`, `step`, `rule`, `kv`, `cmd`, `code`, `tree`, `style`, and `bar` write formatted output to `stdout` and exit `0`.
4. **Async Commands (`/dev/tty` Isolation)**: `spin` and `progress` render live animations directly to `/dev/tty` (falling back to `stderr` if no controlling tty exists). In `zsp spin <title> -- <cmd> [args...]`, the child command's `stdout` and `stderr` pass through untouched so you can capture command output cleanly in Zsh (`OUT=$(zsp spin "..." -- cmd)`) and inspect the child's exact exit status (`$?`).
5. **Signal Safety (`SIGINT` / `SIGTERM`)**: Pressing `Ctrl+C` during `zsp spin` terminates the child process, restores the terminal cursor (`\033[?25h`), clears the spinner line, and exits with code `130`.
6. **Reactive Terminal Width**: Every component detects the live terminal width on each render (`$ZSPLASH_WIDTH` -> `/dev/tty` ioctl -> `stdout`/`stderr` ioctl -> `$COLUMNS` -> `80`).

---

## 2. Global Flags & Environment Variables

### Global Flags

| Flag | Values | Default | Description |
| :--- | :--- | :--- | :--- |
| `-T, --theme <theme>` | `modern`, `nes`, `snes`, `gameboy`, `dracula`, `catppuccin`, `nord`, `gruvbox`, `tokyonight`, `amber` | `modern` (or `config.json`) | Temporarily apply a built-in color palette for the current command without modifying `config.json`. Can be placed anywhere before `--`. |
| `-h, --help` | - | - | Show overview help, or show detailed subcommand documentation when passed after a subcommand (`zsp <cmd> --help`). |
| `-v, --version` | - | - | Print version string (`zsplash (zsp) vX.Y.Z`) and exit. |

### Environment Variables

| Variable | Default | Description |
| :--- | :--- | :--- |
| `ZSPLASH_THEME` | (unset) | Override the active color palette for the current shell session (lower precedence than `--theme`, higher precedence than `config.json`). |
| `ZSPLASH_CONFIG` | `~/.config/zsplash/config.json` | Custom file path for reading and writing persistent user configuration. |
| `ZSPLASH_WIDTH` | Auto-detected | Force a specific terminal column width (integer >= 20) for wrapping and layout calculations. |
| `NO_COLOR` | (unset) | When set to any non-empty value (or when `TERM=dumb`), disables all ANSI color codes and uses plain ASCII bracket badges (`[PASS]`, `[WARN]`, `[FAIL]`, `[INFO]`). |

---

## 3. Color System, Built-in Palettes & Gradients

Any `--color`, `--fg`, or `--bg` flag accepts:
1. **Semantic Theme Roles**: `primary`, `secondary`, `pass`, `warn`, `fail`, `info`, `muted`, `text`
2. **Named Colors**: `cyan`, `magenta`, `purple`, `violet`, `blue`, `sky`, `green`, `mint`, `lime`, `amber`, `gold`, `yellow`, `orange`, `coral`, `red`, `crimson`, `rose`, `pink`, `white`, `slate`, `gray`, `black`, plus NES hardware names (`nes-red`, `nes-blue`, `nes-green`, `nes-gold`, `nes-orange`, `nes-lava`, `nes-cobalt`, `nes-magenta`, `nes-teal`, `nes-gray`).
3. **24-bit Hex Codes**: `#RRGGBB` (e.g. `#E40058`).

### Built-in Theme Palettes (`--theme <name>` or `zsp theme <name>`)

| Theme | Primary | Secondary | Pass | Warn | Fail | Info | Description |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `modern` | `#00F5D4` | `#9B5DE5` | `#00F5A0` | `#FFB703` | `#FF2A85` | `#00BBF9` | Crisp solid cyan, violet, and emerald for modern dark terminals |
| `nes` | `#E40058` | `#0088F8` | `#00A800` | `#F8B800` | `#F83800` | `#3CBCFC` | Nintendo Entertainment System (Ricoh 2C02 PPU hardware palette) |
| `snes` | `#7868C8` | `#B8B0E8` | `#30B848` | `#F8D830` | `#E03030` | `#2878E8` | Super Nintendo 16-bit lavender, periwinkle, and primary button colors |
| `gameboy` | `#8BAC0F` | `#9BBC0F` | `#9BBC0F` | `#8BAC0F` | `#306230` | `#8BAC0F` | Nintendo Game Boy DMG-01 4-shade olive LCD phosphor green |
| `dracula` | `#BD93F9` | `#FF79C6` | `#50FA7B` | `#F1FA8C` | `#FF5555` | `#8BE9FD` | Dracula dark theme with purple, pink, and neon green accents |
| `catppuccin` | `#CBA6F7` | `#89B4FA` | `#A6E3A1` | `#F9E2AF` | `#F38BA8` | `#89DCEB` | Catppuccin Mocha pastel mauve, sapphire, and green |
| `nord` | `#88C0D0` | `#81A1C1` | `#A3BE8C` | `#EBCB8B` | `#BF616A` | `#5E81AC` | Arctic Nord frost blues and aurora borealis status colors |
| `gruvbox` | `#FE8019` | `#FABD2F` | `#B8BB26` | `#FABD2F` | `#FB4934` | `#83A598` | Gruvbox warm retro earth tones, orange, and olive green |
| `tokyonight` | `#7AA2F7` | `#BB9AF7` | `#9ECE6A` | `#E0AF68` | `#F7768E` | `#7DCFFF` | Tokyo Night neon indigo, magenta, and teal |
| `amber` | `#FFB000` | `#FFCC00` | `#FFD866` | `#FF9500` | `#FF5500` | `#FFB000` | Vintage 1980s monochrome amber phosphor CRT terminal |

### Opt-in Gradients (`--gradient <name>`)

Available gradient presets: `cyber`, `ember`, `ocean`, `sunset`, `matrix`, `gold`, `mono` (or custom `--from #RRGGBB --to #RRGGBB`).

---

## 4. Persistent Configuration (`~/.config/zsplash/config.json`)

Use `zsp config set <key> <value>` to override any built-in default:

| Key | Allowed Values | Built-in Default | Description |
| :--- | :--- | :--- | :--- |
| `theme` | `modern`, `nes`, `snes`, `gameboy`, `dracula`, `catppuccin`, `nord`, `gruvbox`, `tokyonight`, `amber` | `modern` | Default color palette used across all commands. |
| `banner.font` | `block`, `compact`, `line` | `block` | Default font for `zsp banner`. |
| `banner.color` | `<role>`, `<named-color>`, `#RRGGBB` | `primary` | Default solid color for `zsp banner`. |
| `box.border` | `rounded`, `double`, `heavy`, `ascii` | `rounded` | Default border style for `zsp box` and `zsp kv`. |
| `box.color` | `<role>`, `<named-color>`, `#RRGGBB` | `primary` | Default border color for `zsp box`. |
| `step.color` | `<role>`, `<named-color>`, `#RRGGBB` | `primary` | Default rule color for `zsp step`. |
| `rule.color` | `<role>`, `<named-color>`, `#RRGGBB` | `primary` | Default line color for `zsp rule`. |
| `kv.color` | `<role>`, `<named-color>`, `#RRGGBB` | `secondary` | Default key and border color for `zsp kv`. |
| `bar.color` | `<role>`, `<named-color>`, `#RRGGBB` | `primary` | Default filled bar color for `zsp bar` and `zsp progress`. |
| `spin.style` | `dots`, `pulse`, `meter`, `line` | `dots` | Default animation style for `zsp spin`. |
| `spin.color` | `<role>`, `<named-color>`, `#RRGGBB` | `primary` | Default spinner frame color for `zsp spin`. |
| `color.<role>` | `<named-color>`, `#RRGGBB` | (from active theme) | Override `color.primary`, `color.secondary`, `color.pass`, `color.warn`, `color.fail`, `color.info`, `color.muted`, or `color.text`. |

---

## 5. Subcommand Reference

### `zsp banner`
Render large, high-contrast Unicode block typography (`A-Z`, `0-9`, punctuation) with optional framing and subtitle.
- **Synopsis**: `zsp banner <text> [--font block|compact|line] [--color <color>] [--gradient <grad>] [--from <#hex> --to <#hex>] [--sub <subtitle>] [--frame] [--align left|center|right]`
- **Stream**: `stdout` | **Exit Code**: `0`

```zsh
zsp banner "DEPLOY" --font block --frame --sub "Production Pipeline"
zsp banner "STAGE 1" --font compact --color secondary
zsp banner "RETRO" --theme nes --color nes-red --frame
```

### `zsp box`
Render a reactive Unicode callout box that word-wraps long text to fit the live terminal width.
- **Synopsis**: `zsp box <text> [--title <title>] [--footer <footer>] [--border rounded|double|heavy|ascii] [--color <color>] [--gradient <grad>] [--align left|center|right] [--width <cols>] [--padding <n>]`
- **Stream**: `stdout` | **Exit Code**: `0`

```zsh
zsp box "All 14 migrations applied cleanly." --title "Database" --border rounded
zsp box "Line 1\nLine 2" --title "Release v1.4" --footer "sha: 8f3a1b" --border double --color pass
```

### `zsp ok` / `zsp info` / `zsp warn` / `zsp err`
Render inverted status pill badges (`PASS`, `INFO`, `WARN`, `FAIL`) with a message and optional right-aligned dot-leader metadata.
- **Synopsis**: `zsp ok|info|warn|err <message> [--meta <detail>]`
- **Stream**: `stdout` | **Exit Code**: `0`

```zsh
zsp ok   "Database schema verified"   --meta "42ms"
zsp info "Worker pool initialized"    --meta "8 threads"
zsp warn "Rate limit at 85% capacity" --meta "throttling"
zsp err  "Upstream webhook timeout"   --meta "exit 1"
```

### `zsp step` & `zsp rule`
Render full-width phase headers (`━━━ [curr/total] ━━━━━━`) or horizontal divider lines (`───`).
- **Synopsis**:
  - `zsp step <current> <total> <title> [--color <color>] [--gradient <grad>]`
  - `zsp rule [label] [--color <color>] [--gradient <grad>] [--char <rune>]`
- **Stream**: `stdout` | **Exit Code**: `0`

```zsh
zsp step 1 4 "Validating environment and credentials"
zsp rule "ARTIFACTS" --color secondary
```

### `zsp kv`, `zsp tree`, `zsp cmd`, `zsp code`, `zsp style`
Render structured key-value cards, checklists, command previews, framed multi-line Zsh code blocks, or inline styled text.
- **Synopsis**:
  - `zsp kv [--title <title>] [--color <color>] <key1> <val1> [<key2> <val2> ...]`
  - `zsp tree [--title <title>] [--color <color>] <item1> [<item2> ...]`
  - `zsp cmd <command string>`
  - `zsp code [--title <title>] <line1> [<line2> ...]`
  - `zsp style <text> [--fg <color>] [--bg <color>] [--bold] [--dim] [--italic] [--underline] [--no-newline|-n]`
- **Stream**: `stdout` | **Exit Code**: `0`

```zsh
zsp kv --title "Deploy Target" "Environment" "prod" "Branch" "main" "Host" "10.0.4.12"
zsp tree --title "Preflight Checks" "Auth token valid" "Cluster reachable" "Migrations clean"
zsp cmd "docker build --platform linux/amd64 -t api:latest ."
zsp code --title "zsh: deploy" 'zsp step 1 2 "Build"' 'zsp ok "Done" --meta "12ms"'
```

### `zsp spin`
Animate a live spinner on `/dev/tty` while executing a child command or timed wait, keeping `stdout` completely clean for Zsh variable capture.
- **Synopsis**: `zsp spin <title> [--style dots|pulse|meter|line] [--color <color>] [--persist] [--duration <ms>] [-- <command> [args...]]`
- **Stream**: Animation renders to `/dev/tty` (fallback: `stderr`). Child `stdout` and `stderr` pass through untouched.
- **Exit Code**: Propagates the exact exit code of `<command>` (`130` on `SIGINT`).

```zsh
# Capture stdout cleanly while displaying a live spinner on /dev/tty:
USER_COUNT=$(zsp spin "Querying users table..." --persist -- psql -tA -c "SELECT count(*) FROM users")

# Wrap any long command (HTTP call, build, sleep):
zsp spin "Fetching release manifest..." --style dots --persist -- curl -fsSL https://example.com -o /tmp/manifest
```

### `zsp bar` & `zsp progress`
Render half-cell (`█` / `▌` / `░`) progress bars in loops or pipelines.
- **Synopsis**:
  - `zsp bar <current> <total> [label] [--in-place|-r] [--width <cols>] [--color <color>] [--gradient <grad>]`
  - `<producer-cmd> | zsp progress [--total <n>] [--title <label>] [--color <color>]`

```zsh
for i in {1..50}; do
  zsp bar $i 50 "Syncing shard $i" --in-place
  sleep 0.02
done

find . -maxdepth 2 -type f | zsp progress --total 100 --title "Indexing files"
```

---

## 6. Complete Zsh Script Template

```zsh
#!/usr/bin/env zsh
set -uo pipefail

export ZSPLASH_THEME="nes"

zsp banner "DEPLOY" --font compact --frame --sub "Production Release Pipeline"

zsp step 1 3 "Verifying cluster & database connectivity"
COMMIT_SHA=$(zsp spin "Resolving latest git commit..." --persist -- git rev-parse --short HEAD)
if [[ $? -ne 0 ]]; then
  zsp err "Failed to resolve git commit" --meta "abort"
  exit 1
fi

zsp step 2 3 "Syncing release artifacts"
for i in {1..20}; do
  zsp bar $i 20 "Uploading chunk $i" --in-place
  sleep 0.02
done
zsp ok "All 20 chunks uploaded" --meta "100%"

zsp step 3 3 "Release Summary"
zsp kv --title "Deployment" \
  "Environment" "production" \
  "Commit"      "$COMMIT_SHA" \
  "Status"      "Healthy"
```

## License

MIT (c) Adam Corrow
