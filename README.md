# feed-tracker

A general-purpose feed tracker that consumes RSS, Atom, JSON Feed, and ActivityPub feeds. Collects feed entries into a local SQLite database. Includes both a CLI for cron-based periodic fetching and an interactive terminal UI.

## Formats

| Format | Parsing |
|---|---|
| RSS 0.90–2.0 | gofeed |
| Atom 0.3, 1.0 | gofeed |
| JSON Feed 1.0, 1.1 | gofeed |
| ActivityPub Outbox | Custom parser |

## Quick Start (CLI)

#### Build (produces ./bin/cli and ./bin/tui)
```bash
make build
```
#### Create config
```bash
cp config.example.yaml config.yaml
```
#### Initialize database
```bash
./bin/cli migrate
```
#### Add a feed
```bash
./bin/cli add https://example.com/feed.xml
```
#### Fetch new entries
```bash
./bin/cli fetch
```
#### List feeds
```bash
./bin/cli feeds
```
#### List entries from a feed (by ID)
```bash
./bin/cli list --feed-id <feed-uuid> --limit 20
```
#### List entries from a feed (by name)
```bash
./bin/cli list "My Feed Name" --limit 10
```
#### List all entries across all feeds
```bash
./bin/cli list --limit 30
```

## Quick Start (TUI)
#### Build (produces ./bin/cli and ./bin/tui)
```bash
make build
```
#### Run (uses config.yaml by default)
```bash
./bin/tui
```
#### Or specify a config path
```bash
./bin/tui --config /path/to/config.yaml
```

## Cron Setup

```cron
# Every hour, fetch all feeds
0 * * * * /usr/bin/ft --config /etc/feedtracker.yaml fetch
```

## Commands (CLI)

| Command | Description |
|---|---|
| `migrate` | Create or update the database schema |
| `add <url>` | Add a feed by URL — detects format automatically |
| `fetch [<name> \| --feed-id <id>]` | Fetch new entries from all feeds, or a specific one |
| `feeds [--names \| --folders]` | List feeds with metadata, unread counts. `--names` one per line, `--folders` grouped by folder |
| `folder` | List folders with feed counts |
| `folder create <name>` | Create a folder |
| `folder rename <old> <new>` | Rename a folder |
| `folder delete <name>` | Delete a folder (feeds become ungrouped) |
| `folder move <feed> <folder>` | Move a feed to a folder |
| `export [--output <file>] [--folders-only \| --feeds-only]` | Export feeds to OPML file. `--folders-only`/`--feeds-only` filter by folder status |
| `import [--dry-run] <file.opml>` | Import feeds from OPML file (preserves folders) |
| `delete <name> \| --feed-id <id>` | Delete a feed and all its entries |
| `list [<name> \| --feed-id <id>] [--limit <n>] [--unread] [--detail]` | List entries. `--unread` filters, `--detail` shows full entry info |
| `read <entry-id>` | Mark an entry as read |
| `unread <entry-id>` | Mark an entry as unread |
| `completion bash\|zsh` | Generate shell completion script |

## TUI

The TUI binary (`ftui`) provides an interactive terminal interface with keyboard navigation.

Read state is tracked per-entry — entries are automatically marked as read when viewed.
Use `u` to toggle between showing only unread entries or all entries.

### Keybindings

| Key | Action |
|---|---|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `Enter` | Select / Confirm |
| `Esc` | Go back |
| `a` | Add a new feed |
| `g` | Create a folder |
| `m` | Move feed to folder |
| `d` | Delete folder or feed |
| `R` | Rename folder |
| `Enter/Space` | Toggle folder collapse |
| `f` | Fetch all feeds |
| `e` | Export feeds to OPML (filter by all/folders/ungrouped) |
| `i` | Import feeds from OPML (with preview before importing) |
| `r` | Refresh current view |
| `u` | Toggle show read entries |
| `L` | Load more entries (paginated, in entry list) |
| `M` | Mark entry unread (in entry detail) |
| `o` | Open entry URL in browser |
| `?` | Toggle help overlay |
| `q/Ctrl+C` | Quit |

The top entry in the feed list is **All Entries** — select it to see entries from all feeds at once.

## Configuration

See `config.example.yaml`:

```yaml
database:
  path: ./data/feeds.db

http:
  timeout: 30s
  user_agent: "feed-tracker/0.1"
  fetch_concurrency: 3   # concurrent fetches during "fetch all"
  fetch_cooldown: 5m     # skip feeds fetched within this duration

tui:
  entry_limit: 100      # max entries loaded per page in TUI
  auto_refresh: 5m      # auto-fetch interval (e.g. 5m, 30m, 0 to disable)
```

## Development

A Makefile is provided for common tasks:

| Command | What it does |
|---|---|
| `make build` | Build both `cli` and `tui` binaries into `./bin/` |
| `make test` | Run all tests with race detector |
| `make vet` | Run `go vet` static analysis |
| `make lint` | Run golangci-lint (if installed) |
| `make tidy` | Tidy module dependencies |
| `make clean` | Remove build artifacts |
| `make run-cli` | Quick CLI run via `go run` |
| `make run-tui` | Quick TUI run via `go run` |
| `make install` | Install binaries to `$GOPATH/bin` |
| `make all` | Build, vet, test (CI-style) |

Set `RACE=0` to disable the race detector for faster test runs:

```bash
make test RACE=0
```

### Pre-push hook

A pre-push hook is configured at `.githooks/pre-push` that runs `make all` (build + vet + test) before every `git push`. The hook directory is set via `git config core.hooksPath .githooks` — already configured in this repo. New clones should run:

```bash
git config core.hooksPath .githooks
```
