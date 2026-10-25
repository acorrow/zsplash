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

## Themes & Config

Built-in palettes: `modern` (default), `nes`, `xbox`, `snes`, `n64`, `gamecube`, `gameboy`, `ps1`, `sega`, `pipboy`, `cyberpunk`, `halo`, `doom`, `dracula`, `catppuccin`, `nord`, `gruvbox`, `tokyonight`, `amber`.

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
