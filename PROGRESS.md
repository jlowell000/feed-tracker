# Progress

| # | Step | Status | Notes |
|---|---|---|---|---|
| 1 | Scaffold — go mod init, deps, main.go routing | ✓ | |
| 2 | Domain types — Feed, Entry, FeedType | ✓ | `internal/domain/` |
| 3 | Config — YAML loading | ✓ | `internal/config/` |
| 4 | Storage — SQLite migrations + CRUD | ✓ | `internal/storage/` |
| 5 | Fetcher — HTTP client with conditional GET | ✓ | `internal/fetcher/` |
| 6 | Parser — gofeed + ActivityPub | ✓ | `internal/parser/` |
| 7 | Tracker — fetch → parse → store | ✓ | `internal/feedtracker/` |
| 8 | CLI: fetch command | ✓ | |
| 9 | CLI: add command | ✓ | |
| 10 | CLI: feeds/list commands | ✓ | |
| 11 | CLI: migrate command | ✓ | |
| 12 | README + PROGRESS docs | ✓ | |
| 13 | Polish — error handling, tests, CI | ✓ | Tests and CI in place |
| 14 | CLI: feed name + all-entries for list | ✓ | Positional feed name, FEED column, all-entries mode |
| 15 | CLI: feed name for fetch | ✓ | Also accepts positional feed name |
| 16 | CLI: completion subcommand | ✓ | bash/zsh shell completion script generation |
| 17 | CLI: updated CLI help/usage | ✓ | main.go usage text updated |
| 18 | Update docs | ✓ | README + PROGRESS |
| 19 | TUI: interactive terminal UI | ✓ | `cmd/tui/` using Bubble Tea |
| 20 | TUI: read state tracking + toggle | ✓ | `read` field on entries, `u` to toggle unread/all, auto-mark read on view |
| 21 | TUI: unread counts per feed | ✓ | `UnreadCountByFeed` query, shown in feeds list + status bar |
| 22 | Folders: group feeds into folders | ✓ | `Folder` domain type, folders table, CLI subcommand, TUI grouped display |
| 23 | CLI: OPML import with folders + dry-run | ✓ | `internal/opml/`, `ft import [--dry-run] <file.opml>`, folder creation |
