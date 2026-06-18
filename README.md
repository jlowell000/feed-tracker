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

```bash
# Build
go build -o bin/ft ./cmd/cli

# Create config
cp config.example.yaml config.yaml

# Initialize database
./bin/ft migrate

# Add a feed
./bin/ft add https://example.com/feed.xml

# Fetch new entries
./bin/ft fetch

# List feeds
./bin/ft feeds

# List entries from a feed (by ID)
./bin/ft list --feed-id <feed-uuid> --limit 20

# List entries from a feed (by name)
./bin/ft list "My Feed Name" --limit 10

# List all entries across all feeds
./bin/ft list --limit 30
```

## Quick Start (TUI)

```bash
# Build
go build -o bin/ftui ./cmd/tui

# Run (uses config.yaml by default)
./bin/ftui

# Or specify a config path
./bin/ftui --config /path/to/config.yaml
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
| `export [--output <file>]` | Export feeds to OPML file (preserves folders) |
| `import [--dry-run] <file.opml>` | Import feeds from OPML file (preserves folders) |
| `delete <name> \| --feed-id <id>` | Delete a feed and all its entries |
| `list [<name> \| --feed-id <id>] [--limit <n>] [--unread]` | List entries, newest first. `--unread` filters to unread only. |
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
| `e` | Export feeds to OPML file |
| `i` | Import feeds from OPML file |
| `r` | Refresh current view |
| `u` | Toggle show read entries |
| `M` | Mark entry unread (in entry detail) |
| `o` | Open entry URL in browser |
| `?` | Toggle help overlay |
| `q/Ctrl+C` | Quit |

## Configuration

See `config.example.yaml`:

```yaml
database:
  path: ./data/feeds.db

http:
  timeout: 30s
  user_agent: "feed-tracker/0.1"
```

## Development

```bash
go test ./...
go vet ./...
go build ./...
```
