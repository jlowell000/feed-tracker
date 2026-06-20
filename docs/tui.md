# TUI Reference

The `ftui` binary provides an interactive terminal interface with keyboard navigation.

## Quick Start

#### Build (produces ./bin/cli and ./bin/tui)
```bash
make build
```

#### Run (uses config.yaml by default)
```bash
./bin/tui
```

#### Run with custom config
```bash
./bin/tui --config /path/to/config.yaml
```

## Usage

Read state is tracked per-entry — entries are automatically marked as read when viewed. Use `u` to toggle between showing only unread entries or all entries.

The top entry in the feed list is **All Entries** — select it to see entries from all feeds at once.

## Keybindings

| Key | Action | Context |
|---|---|---|
| `↑/k` | Move up | All lists |
| `↓/j` | Move down | All lists |
| `PgUp` | Page up | All lists |
| `PgDn` | Page down | All lists |
| `Home` | Go to first | All lists |
| `End` | Go to last | All lists |
| `Enter` | Select / Confirm | All |
| `Esc` | Go back | All |
| `?` | Toggle help overlay | All |
| `q/Ctrl+C` | Quit | All |
| `a` | Add a new feed | Feed list |
| `g` | Create a folder | Feed list |
| `d` | Delete folder or feed | Feed list |
| `R` | Rename folder | Feed list |
| `Enter/Space` | Toggle folder collapse | Feed list |
| `E` | Edit feed title/URL | Feed list / Entries list |
| `m` | Move feed to folder | Feed list |
| `f` | Fetch all feeds | Feed list |
| `e` | Export feeds to OPML | Feed list |
| `i` | Import feeds from OPML | Feed list |
| `r` | Refresh feed list | Feed list |
| `u` | Toggle show read / unread | Entries list |
| `s` | Search entries | Entries list |
| `a` | Mark displayed entries read | Entries list |
| `A` | Mark all in feed read | Entries list |
| `L` | Load more entries | Entries list |
| `f` | Filter by feed | Entries list (All Entries view) |
| `[` / `]` | Switch to prev/next feed | Entries list |
| `M` | Mark entry unread | Entry detail |
| `o` | Open entry URL in browser | Entry detail |
| `↑/↓` | Scroll line by line | Entry detail |
| `PgUp/PgDn` | Scroll page by page | Entry detail |
