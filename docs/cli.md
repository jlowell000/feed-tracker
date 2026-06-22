# CLI Reference

The `ft` binary is the command-line interface for feed-tracker.

## Quick Start


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

## Commands

| Command | Description |
|---|---|
| `migrate` | Create or update the database schema |
| `add <url>` | Add a feed by URL — detects format automatically |
| `fetch [<name> \| --feed-id <id>]` | Fetch new entries from all feeds, or a specific one |
| `feeds [--names \| --folders]` | List feeds with metadata, unread counts. `--names` one per line, `--folders` grouped by folder |
| `feed update <name> [--title <title>] [--url <url>] [--prune-age <duration>]` | Update feed title, URL, and/or per-feed prune age |
| `open <entry-id>` | Open entry URL in system browser |
| `folder` | List folders with feed counts |
| `folder create <name>` | Create a folder |
| `folder rename <old> <new>` | Rename a folder |
| `folder delete <name>` | Delete a folder (feeds become ungrouped) |
| `folder move <feed> <folder>` | Move a feed to a folder |
| `export [--output <file>] [--folders-only \| --feeds-only]` | Export feeds to OPML file. `--folders-only`/`--feeds-only` filter by folder status |
| `import [--dry-run] <file.opml>` | Import feeds from OPML file (preserves folders) |
| `delete <name> \| --feed-id <id>` | Delete a feed and all its entries |
| `list [<name> \| --feed-id <id>] [--limit <n>] [--unread] [--detail] [--search <q>]` | List entries. `--unread` filters, `--detail` shows full entry info, `--search` filters by keyword |
| `search [--limit <n>] <query>` | Search entries by keyword across title and summary |
| `read [--all \| --feed <name> \| --feed-id <id> \| <entry-id>]` | Mark entries as read: all, by feed, or single entry |
| `unread <entry-id>` | Mark an entry as unread |
| `prune` | Delete entries older than the configured `prune.max_age` |
| `completion bash\|zsh` | Generate shell completion script |

## Cron Setup

```cron
# Every hour, fetch all feeds
0 * * * * /usr/bin/ft --config /etc/feedtracker.yaml fetch
```
