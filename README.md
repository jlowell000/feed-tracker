# feed-tracker

A general-purpose feed tracker that consumes RSS, Atom, JSON Feed, and ActivityPub feeds. Collects feed entries into a local SQLite database. Designed for cron-based periodic fetching — no daemon or UI required.

## Formats

| Format | Parsing |
|---|---|
| RSS 0.90–2.0 | gofeed |
| Atom 0.3, 1.0 | gofeed |
| JSON Feed 1.0, 1.1 | gofeed |
| ActivityPub Outbox | Custom parser |

## Quick Start

```bash
# Build
go build -o bin/feedtracker ./cmd/feedtracker

# Create config
cp config.example.yaml config.yaml

# Initialize database
./bin/feedtracker migrate

# Add a feed
./bin/feedtracker add https://example.com/feed.xml

# Fetch new entries
./bin/feedtracker fetch

# List feeds
./bin/feedtracker feeds

# List entries from a feed
./bin/feedtracker list --feed-id <feed-uuid> --limit 20
```

## Cron Setup

```cron
# Every hour, fetch all feeds
0 * * * * /usr/bin/feedtracker --config /etc/feedtracker.yaml fetch
```

## Commands

| Command | Description |
|---|---|
| `migrate` | Create or update the database schema |
| `add <url>` | Add a feed by URL — detects format automatically |
| `fetch [--feed-id <id>]` | Fetch new entries from all feeds, or a specific one |
| `feeds` | List all tracked feeds with metadata |
| `list --feed-id <id> [--limit <n>]` | List entries for a feed, newest first |

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
