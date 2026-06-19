# Progress

| # | Step | Status | Notes |
|---|---|---|---|
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
| 24 | CLI: OPML export with folders | ✓ | `ft export [--output <file>]`, `internal/opml/export.go` |
| 25 | CLI: delete feed | ✓ | `ft delete <name>`, `storage.DeleteFeed` |
| 26 | TUI: OPML import/export | ✓ | `e`/`i` keys for export/import screens |
| 27 | TUI: delete feed | ✓ | `d` key deletes feed or folder |
| 28 | CLI: read/unread commands | ✓ | `ft read`/`ft unread <entry-id>` |
| 29 | CLI: list --unread + read column | ✓ | `ft list --unread`, READ column in output |
| 30 | CLI: feeds unread counts + --folders | ✓ | Unread column in `ft feeds`, `ft feeds --folders` grouped view |
| 31 | CLI: folder move | ✓ | `ft folder move <feed> <folder>` |
| 32 | CLI: list --detail mode | ✓ | `ft list --detail` shows author, content snippet, read status |
| 33 | TUI: configurable entry limit | ✓ | `tui.entry_limit` in config replaces hardcoded 100 |
| 34 | CLI: export --folders-only / --feeds-only | ✓ | Selective export flags |
| 35 | TUI: auto-refresh with configurable interval | ✓ | `tui.auto_refresh` config, periodic fetch |
| 36 | TUI: global "All Entries" view | ✓ | Shows entries across all feeds in feed list |
| 37 | TUI: import dry-run preview | ✓ | Parses OPML, shows preview before importing |
| 38 | TUI: export folder-only / feeds-only filter | ✓ | Pick screen: all, folders only, ungrouped only |
| 39 | Storage: WAL mode + foreign key enforcement | ✓ | `PRAGMA journal_mode=WAL`, `PRAGMA foreign_keys=ON` in `New()` |
| 40 | Storage: composite index (feed_id, published_at) | ✓ | `idx_entries_feed_published` |
| 41 | Storage: time parse errors no longer silent | ✓ | `scanFeed`/`scanEntry` propagate parse errors |
| 42 | Storage: migration ALTER TABLE errors handled | ✓ | `isDupColumnError` ignores re-run errors, catches real failures |
| 43 | Storage: DB Ping on open | ✓ | `db.Ping()` in `New()` |
| 44 | Completion scripts: all missing commands added | ✓ | `folder`, `import`, `export`, `delete`, `read`, `unread` |
| 45 | Tests: storage all-feeds + cascade + UpdateFeed | ✓ | 7 new tests added |
| 46 | Tests: config TUI defaults + auto_refresh | ✓ | `TestLoadWithTUIConfig` |
| 47 | Tests: feedtracker network/parse error handling | ✓ | `TestAddFeed_NetworkError`, `TestAddFeed_MalformedFeed`, `TestFetchFeed_NetworkError` |
| 48 | Tests: parser edge cases | ✓ | empty body, no items, malformed dates |
| 49 | Perf: entry list query without JOIN (known feedID) | ✓ | `entryCols` constant, no LEFT JOIN per-feed |
| 50 | Perf: cursor-based pagination in TUI entry list | ✓ | `L` key loads next page, `entryOffset` tracking |
| 51 | Perf: bounded concurrency for FetchAllFeeds | ✓ | Worker pool via `fetch_concurrency`, `sync.WaitGroup` |
| 52 | Perf: feed staleness check | ✓ | `fetch_cooldown` config skips recently-fetched feeds |
| 53 | Perf: offset support in ListEntries/ListEntriesUnread | ✓ | `LIMIT ? OFFSET ?` in storage layer |
| 54 | Build: Makefile | ✓ | `build`, `test`, `vet`, `lint`, `tidy`, `clean`, `run-cli`, `run-tui`, `install`, `all` targets |
| 55 | Docs: README updated | ✓ | Development section with Makefile table |
| 56 | Docs: plan.md updated | ✓ | Phase 3b added, phases renumbered |
| 57 | Build: pre-push hook | ✓ | `.githooks/pre-push` runs `make all`, `core.hooksPath` configured |
| 58 | Storage: SearchEntries (LIKE-based) | ✓ | `WHERE title LIKE ? OR summary LIKE ?` |
| 59 | Storage: MarkFeedRead / MarkAllRead | ✓ | Bulk mark-read by feed or all |
| 60 | CLI: ft search command | ✓ | `ft search [--limit <n>] <query>` |
| 61 | CLI: ft read --all / --feed / --feed-id | ✓ | Bulk mark-read via CLI |
| 62 | CLI: ft list --search | ✓ | Filter list output by keyword |
| 63 | TUI: search screen | ✓ | `s` key opens search input, results replace entry list |
| 64 | TUI: bulk mark-read | ✓ | `a` marks displayed, `A` marks all in feed |
| 65 | Tests: storage search + mark-read | ✓ | 6 new tests |
| 66 | Docs: README, plan.md, PROGRESS.md | ✓ | Phase 4 documented |
| 67 | Storage: Vacuum / Optimize methods | ✓ | Added to interface + SQLite |
| 68 | CLI: ft vacuum / ft db optimize | ✓ | New subcommands |
| 69 | Fetcher: Fetch accepts context.Context | ✓ | Removed FetchWithTimeout |
| 70 | TUI: replace context.Background() with timeouts | ✓ | 22 command functions + update.go inline call updated |
| 71 | Tests: fetcher context param | ✓ | f.Fetch calls updated to pass context.Background() |
