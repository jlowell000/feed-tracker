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

## Phase 6: Automatic Entry Pruning

| # | Step | Status |
|---|---|---|
| 72 | Config: prune.max_age field + example yaml | ✓ |
| 73 | Storage: DeleteEntriesOlderThan method + SQLite impl | ✓ |
| 74 | CLI: ft prune command | ✓ |
| 75 | Auto-prune after ft fetch | ✓ |
| 76 | Tests: storage DeleteEntriesOlderThan (basic, zero-age, all-deleted) | ✓ |

## Phase 7: Feed & Entry Management

| # | Step | Status |
|---|---|---|
| 77 | Storage: GetEntry method + SQLite impl + test | ✓ |
| 78 | CLI: ft feed update command (title/URL editing) | ✓ |
| 79 | CLI: ft open command (browser open entry) | ✓ |
| 80 | TUI: feed edit screen (E key) | ✓ |
| 81 | TUI: entry filter by feed in All Entries view (f key) | ✓ |
| 82 | TUI: keyboard-driven feed switching ([ / ]) | ✓ |
| 83 | Docs: completion scripts, help view, cli.md, plan/progress | ✓ |

## Phase 8: Performance & Polish

| # | Step | Status |
|---|---|---|
| 84 | Perf: lazy content/summary — list queries skip summary+content columns | ✓ |
| 85 | UX: on-screen key hints — expanded per-screen header hints | ✓ |
| 86 | UX: auto-refresh countdown — shows remaining time in status bar | ✓ |
| 87 | UX: context-aware help screen — filtered by current screen | ✓ |
| 88 | Tests: domain package — entry, feed, folder struct tests | ✓ |
| 89 | Docs: plan.md updated for Phase 8 | ✓ |
| 90 | UX: width-adaptive hint rendering — hints on own line, wrap if too long | ✓ |
| 91 | Refactor: centralized helpBinding system — all bindings defined once, no string duplication | ✓ |
| 92 | UX: help screen width matches terminal width (capped at 60) | ✓ |
| 93 | Fix: help opens on entry detail view | ✓ |
| 94 | Fix: feed list help no longer shows prev/next binding | ✓ |
| 95 | Cleanup: remove dead keys.go, unused markEntryUnreadCmd, dead Width(40) | ✓ |
| 96 | Polish: secondary views (add, import, search, etc.) use binding system | ✓ |

## Phase 9: view.go Refactor

| # | Step | Status |
|---|---|---|
| 97 | Refactor: split view.go into bindings.go, help.go, views.go | ✓ |
