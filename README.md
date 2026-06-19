# feed-tracker

A general-purpose feed tracker that consumes RSS, Atom, JSON Feed, and ActivityPub feeds. Collects feed entries into a local SQLite database. Includes both a CLI for cron-based periodic fetching and an interactive terminal UI.

## Formats

| Format | Parsing |
|---|---|
| RSS 0.90–2.0 | gofeed |
| Atom 0.3, 1.0 | gofeed |
| JSON Feed 1.0, 1.1 | gofeed |
| ActivityPub Outbox | Custom parser |

## Quick Start

```bash
make build
cp config.example.yaml config.yaml
./bin/cli migrate
./bin/cli add https://example.com/feed.xml
./bin/cli fetch
```

## Documentation

| Topic | File |
|---|---|
| CLI commands & usage | [docs/cli.md](docs/cli.md) |
| TUI keybindings & usage | [docs/tui.md](docs/tui.md) |
| Configuration reference | [docs/config.md](docs/config.md) |
| Development guide | [docs/development.md](docs/development.md) |
