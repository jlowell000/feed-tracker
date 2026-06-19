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
| `s` | Search entries by keyword (in entry list) |
| `a` | Mark all displayed entries as read (in entry list) |
| `A` | Mark all entries in current feed as read (in entry list) |
| `L` | Load more entries (paginated, in entry list) |
| `M` | Mark entry unread (in entry detail) |
| `o` | Open entry URL in browser |
| `?` | Toggle help overlay |
| `q/Ctrl+C` | Quit |
