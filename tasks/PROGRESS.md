# Progress

## Core Development

| # | Step | Status |
|---|---|---|
| 1 | Scaffold — go mod init, deps, main.go routing | ✓ |
| 2 | Domain types — Feed, Entry, FeedType | ✓ |
| 3 | Config — YAML loading | ✓ |
| 4 | Storage — SQLite migrations + CRUD | ✓ |
| 5 | Fetcher — HTTP client with conditional GET | ✓ |
| 6 | Parser — gofeed + ActivityPub | ✓ |
| 7 | Tracker — fetch → parse → store | ✓ |
| 8 | CLI: fetch command | ✓ |
| 9 | CLI: add command | ✓ |
| 10 | CLI: feeds/list commands | ✓ |
| 11 | CLI: migrate command | ✓ |
| 12 | README + PROGRESS docs | ✓ |
| 13 | Polish — error handling, tests, CI | ✓ |
| 14 | CLI: feed name + all-entries for list | ✓ |
| 15 | CLI: feed name for fetch | ✓ |
| 16 | CLI: completion subcommand | ✓ |
| 17 | CLI: updated CLI help/usage | ✓ |
| 18 | Update docs | ✓ |
| 19 | TUI: interactive terminal UI | ✓ |
| 20 | TUI: read state tracking + toggle | ✓ |
| 21 | TUI: unread counts per feed | ✓ |
| 22 | Folders: group feeds into folders | ✓ |
| 23 | CLI: OPML import with folders + dry-run | ✓ |
| 24 | CLI: OPML export with folders | ✓ |
| 25 | CLI: delete feed | ✓ |
| 26 | TUI: OPML import/export | ✓ |
| 27 | TUI: delete feed | ✓ |
| 28 | CLI: read/unread commands | ✓ |
| 29 | CLI: list --unread + read column | ✓ |
| 30 | CLI: feeds unread counts + --folders | ✓ |
| 31 | CLI: folder move | ✓ |
| 32 | CLI: list --detail mode | ✓ |

## Phase 1: TUI/CLI Feature Gaps

| # | Step | Status |
|---|---|---|
| 33 | TUI: configurable entry limit | ✓ |
| 34 | CLI: export --folders-only / --feeds-only | ✓ |
| 35 | TUI: auto-refresh with configurable interval | ✓ |
| 36 | TUI: global "All Entries" view | ✓ |
| 37 | TUI: import dry-run preview | ✓ |
| 38 | TUI: export folder-only / feeds-only filter | ✓ |

## Phase 2: Testing & Hardening

| # | Step | Status |
|---|---|---|
| 39 | Storage: WAL mode + foreign key enforcement | ✓ |
| 40 | Storage: composite index (feed_id, published_at) | ✓ |
| 41 | Storage: time parse errors no longer silent | ✓ |
| 42 | Storage: migration ALTER TABLE errors handled | ✓ |
| 43 | Storage: DB Ping on open | ✓ |
| 44 | Completion scripts: all missing commands added | ✓ |
| 45 | Tests: storage all-feeds + cascade + UpdateFeed | ✓ |
| 46 | Tests: config TUI defaults + auto_refresh | ✓ |
| 47 | Tests: feedtracker network/parse error handling | ✓ |
| 48 | Tests: parser edge cases | ✓ |

## Phase 3: Large List Performance

| # | Step | Status |
|---|---|---|
| 49 | Perf: entry list query without JOIN (known feedID) | ✓ |
| 50 | Perf: cursor-based pagination in TUI entry list | ✓ |
| 51 | Perf: bounded concurrency for FetchAllFeeds | ✓ |
| 52 | Perf: feed staleness check | ✓ |
| 53 | Perf: offset support in ListEntries/ListEntriesUnread | ✓ |

## Phase 3b: Build Tooling

| # | Step | Status |
|---|---|---|
| 54 | Build: Makefile | ✓ |
| 55 | Docs: README updated | ✓ |
| 56 | Docs: plan.md updated | ✓ |

## Build: Pre-push Hook

| # | Step | Status |
|---|---|---|
| 57 | Build: pre-push hook | ✓ |

## Phase 4: Search & Bulk Mark-Read

| # | Step | Status |
|---|---|---|
| 58 | Storage: SearchEntries (LIKE-based) | ✓ |
| 59 | Storage: MarkFeedRead / MarkAllRead | ✓ |
| 60 | CLI: ft search command | ✓ |
| 61 | CLI: ft read --all / --feed / --feed-id | ✓ |
| 62 | CLI: ft list --search | ✓ |
| 63 | TUI: search screen | ✓ |
| 64 | TUI: bulk mark-read | ✓ |
| 65 | Tests: storage search + mark-read | ✓ |
| 66 | Docs: README, plan.md, PROGRESS.md | ✓ |

## Phase 5: Hardening & Polish

| # | Step | Status |
|---|---|---|
| 67 | Storage: Vacuum / Optimize methods | ✓ |
| 68 | CLI: ft vacuum / ft db optimize | ✓ |
| 69 | Fetcher: Fetch accepts context.Context | ✓ |
| 70 | TUI: replace context.Background() with timeouts | ✓ |
| 71 | Tests: fetcher context param | ✓ |
