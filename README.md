# O2S CLI

> A beautiful, powerful Swiss-army-knife CLI for developers, sysadmins and humans.

`o2s` bundles four daily-driver toolkits behind one polished, themed command-line tool:

- **`dev`** - project scaffolding, static serving, script runners, multi-repo git status
- **`sys`** - system info, live process viewer, port scanner, TCP ping, DNS lookups
- **`files`** - tree, fuzzy find, hashing, JSON/YAML conversion, duplicate finder
- **`prod`** - todos, markdown notes, countdown timer, full-screen Pomodoro focus mode
- **`plugin`** - drop-in executable extensions (`o2s plugin …`)

Built in Go on top of the [Charm](https://charm.sh) ecosystem (Bubble Tea, Lipgloss, Huh, Glamour) so every screen looks and feels like the same product.

---

## Install

### Build from source (recommended for now)

You need Go 1.22+.

```bash
git clone https://github.com/O2S-Y/o2s
cd o2s
make build           # produces ./bin/o2s
./bin/o2s --help
```

Or install directly into your `GOBIN`:

```bash
go install github.com/O2S-Y/o2s/cmd/o2s@latest
```

### Pre-built binaries

Tagged releases publish cross-platform binaries to GitHub Releases (Linux, macOS, Windows on amd64 + arm64) via [GoReleaser](https://goreleaser.com). See `.goreleaser.yaml` for the build matrix and the optional Scoop / Homebrew tap blocks (commented out until a public repo exists).

---

## Quick tour

```bash
o2s                          # interactive dashboard (Bubble Tea)
o2s --theme neon sys info    # change theme on the fly
o2s sys top                  # live process viewer
o2s files tree -L 2          # styled tree, depth 2
o2s files find router        # fuzzy file search
o2s prod todo add "ship it" --priority high
o2s prod focus               # Pomodoro: pick a todo and lock in
o2s dev init my-app          # interactive project scaffolder
o2s dev serve ./public -p 8080
o2s dev git status .         # multi-repo overview
o2s dev http GET https://example.com
o2s dev env print --format powershell .env
o2s dev jwt decode eyJhbGciOiJIUzI1NiJ9....
o2s prod snippet add curl --body 'curl -fsSL https://example.com'
o2s prod clip get
o2s plugin init
o2s plugin list
```

Every command supports:

| Flag         | What it does                                  |
|--------------|-----------------------------------------------|
| `--theme`    | `default \| neon \| mono \| dracula`          |
| `--no-color` | Strip ANSI styling (good for pipes / CI)      |
| `--json`     | Emit machine-readable JSON where supported    |
| `-v`         | Verbose / debug logging                       |

---

## Commands

### `o2s` — dashboard
Greets you, shows live CPU/memory, recent todos, and useful next-step hints. Quit with `q`, refresh with `r`.

### `o2s sys`
- `info` - hostname, OS, CPU, RAM, disk, primary IPv4, uptime
- `top` - live process viewer (sort with `c`, `m`, `n`)
- `ports` - listening sockets + owning processes
- `ping <host>` - TCP-based ping (no admin/raw-socket needed)
- `dns <domain>` - A / AAAA / CNAME / MX / NS / TXT records

### `o2s files`
- `tree [path]` - pretty tree, gitignore-aware by default
- `find <query>` - fuzzy file search with match highlighting
- `hash <file>...` - md5 / sha1 / sha256 / sha512
- `convert <file>` - JSON ↔ YAML (auto-detect from extension)
- `dedup [path]` - find duplicate files (size + sha256), report savings

### `o2s prod`
- `todo add | list | done | rm` - lightweight todos (BoltDB, no CGO)
- `note new | list | open | show` - markdown notes; `show` renders via Glamour
- `timer <duration>` - countdown with progress bar
- `focus` - full-screen Pomodoro session bound to a chosen todo
- `snippet add | list | get | copy | rm` - saved text snippets (`copy` writes to clipboard)
- `clip get | set` - read/write clipboard (via `github.com/atotto/clipboard`)

### `o2s dev`
- `init [name]` - Huh wizard scaffolds Node/Go/Python/static templates
- `serve [path]` - static HTTP server with LAN URL list
- `run [name]` - script picker across `package.json`, `Makefile`, `Taskfile.yml`
- `git status [path]` - multi-repo branch / dirty / ahead-behind overview
- `git sync [path]` - fetch + rebase shorthand
- `env print|get` - load `.env` and print as table / JSON / shell export snippets
- `http <method> <url>` - tiny HTTP client (`-H`, `--data`, `--body`)
- `jwt decode <token>` - decode header + payload (**does not verify** signatures)

### `o2s plugin`
- `path` / `init` - print / create the plugins directory
- `list` - enumerate executables dropped into the folder
- `run <name> -- [...]` - execute a plugin with forwarded args

Plugins receive `O2S_PLUGIN=1`, `O2S_JSON`, and `O2S_THEME`. Override the folder with `O2S_PLUGIN_DIR`.

### `o2s config`
- `show` - print current config
- `path` - print config file path
- `set-theme [name]` - persist a theme
- `reset` - back to defaults

---

## Storage layout

| Path                                 | Purpose                                       |
|--------------------------------------|-----------------------------------------------|
| `<config>/o2s/config.yaml`           | preferences (theme, editor, name, no-color)   |
| `<config>/o2s/data/o2s.db`           | BoltDB — todos + snippets                       |
| `<config>/o2s/plugins/`             | drop-in plugin executables                      |
| `<config>/o2s/data/notes/*.md`       | Markdown notes                                |

Where `<config>` is `%APPDATA%` on Windows and `$XDG_CONFIG_HOME` (or `~/.config`) elsewhere. Override the whole tree with `O2S_CONFIG_DIR`.

> ⚙️ **Why BoltDB instead of SQLite?** It's pure-Go and CGO-free, which means a single static binary cross-compiles to every OS without a C toolchain. Perfect for a CLI distributed via Goreleaser.

---

## Development

```bash
make tidy        # go mod tidy
make build       # build ./bin/o2s
make run         # go run ./cmd/o2s
make test        # go test ./...
make fmt vet     # gofmt + go vet
make release-snapshot   # local goreleaser snapshot, no publish
```

CI runs on Linux/macOS/Windows via GitHub Actions; tagged pushes (`v*`) trigger Goreleaser.

---

## Recording demo GIFs

Use [`vhs`](https://github.com/charmbracelet/vhs) to record consistent GIFs of each command. A starter tape would look like:

```text
Output demo.gif
Set FontSize 18
Type "o2s sys info"
Enter
Sleep 3s
```

Drop completed tapes into `docs/tapes/` and link the rendered GIFs from this README.

---

## Roadmap

- Plugin system: drop a binary into `~/.config/o2s/plugins` to register new subcommands.
- More `sys` tools: speedtest, traceroute (when raw sockets are OK).
- More `files` conversions: csv ↔ json, md → html (Glamour reverse render).
- Richer dashboard widgets (network throughput, per-disk usage).
- Shell completions (`o2s completion {bash|zsh|fish|powershell}`) — Cobra ships this for free.

---

## License

MIT — see [LICENSE](LICENSE).
