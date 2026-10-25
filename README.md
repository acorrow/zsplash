# zsplash (`zsp`)

A fast, reactive CLI utility for Zsh and shell scripts to render colorful, expressive terminal text: crisp Unicode block banners, boxes, status pill badges, phase dividers, key-value cards, command previews, spinners, and progress bars.

Both `zsplash` and its short alias **`zsp`** are installed out of the box. Every component automatically adapts to the live terminal width (`$COLUMNS` / `/dev/tty`) so resizing your terminal never scrambles borders or animations.

---

## Installation

### Homebrew (macOS & Linux)

```zsh
brew install acorrow/tap/zsplash
```

### Fast Dev Updates (`develop` branch)

```zsh
brew upgrade --fetch-HEAD zsplash
```

---

## Quick Start

### 1. Visual Showcase (`zsp demo`)

```zsh
zsp demo
```

### 2. Crisp "Anti-Figlet" Block Banners (`zsp banner`)

Render high-contrast Unicode block typography in 3 built-in fonts (`block`, `compact`, `line`) with smooth 24-bit TrueColor gradients (`cyber`, `ember`, `ocean`, `sunset`, `matrix`, `gold`, `mono`):

```zsh
zsp banner "DEPLOY" --font block --gradient cyber --frame --sub "Production Pipeline"
zsp banner "FAST" --font compact --gradient ember
zsp banner "ARCHITECT" --font line --gradient matrix
```

### 3. Inverted Pill Badges (`zsp ok`, `zsp warn`, `zsp err`, `zsp info`)

```zsh
zsp ok "Database schema verified" --meta "42ms"
zsp info "Listening on port 8080" --meta "tls"
zsp warn "Rate limit approaching" --meta "85%"
zsp err "Connection timed out" --meta "exit 1"
```

### 4. Step Dividers, Horizontal Rules, Key-Value Cards & Command Previews

```zsh
zsp step 2 5 "Migrating database"
zsp rule "ARTIFACTS" --gradient cyber
zsp kv --title "Deploy Target" "Environment" "prod" "Branch" "main" "Host" "10.0.4.12"
zsp cmd "docker build --platform linux/amd64 -t api:latest ."
zsp tree --title "Preflight Checks" "Auth token valid" "Cluster reachable" "Migrations clean"
```

### 5. Async Spinners & Progress Bars in Zsh Scripts

#### Wrap a Slow Command with `zsp spin`
`zsp spin` animates directly on `/dev/tty` (or `stderr`), leaving `stdout` untouched so you can capture command output into a Zsh variable while showing a live spinner:

```zsh
# Capture stdout cleanly while displaying a live spinner on the terminal:
USER_COUNT=$(zsp spin "Querying database..." --persist -- psql -tA -c "SELECT count(*) FROM users")

# Or wrap any long task (HTTP call, sleep, build):
zsp spin "Fetching remote manifest..." --style dots --persist -- curl -sL https://example.com -o /tmp/manifest
```

If a user presses `Ctrl+C` mid-spin, `zsp` catches `SIGINT`, kills the child process, restores the terminal cursor (`\033[?25h`), clears the line, and exits cleanly.

#### Loop Progress Bar (`zsp bar --in-place`)

```zsh
for i in {1..50}; do
  zsp bar $i 50 "Syncing shard $i" --in-place --gradient cyber
  sleep 0.03
done
```

#### Pipeline Stream Progress (`zsp progress`)

```zsh
find . -type f | zsp progress --total 100 --title "Indexing files"
```

## License

MIT (c) Adam Corrow
