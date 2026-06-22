# Feed Tracker — Development Phases

## Phase 1: TUI/CLI Feature Gaps + Auto-Refresh ✅

> Completed — all items implemented.

### What was done

| Feature | Files changed |
|---|---|
| Config `auto_refresh` field | `internal/config/config.go`, `config.example.yaml` |
| Auto-refresh ticker in TUI | `cmd/tui/model.go` (Init, autoRefreshTick, fetchCompleteMsg resets ticker), `cmd/tui/update.go` |
| Global "All Entries" view | `cmd/tui/model.go` (allEntriesItem kind, buildDisplayItems), `cmd/tui/view.go` (rendering + feed name in entries list), `cmd/tui/update.go` (enter handler) |
| Import dry-run preview | `cmd/tui/model.go` (importPreviewMsg, importPreviewCmd), `cmd/tui/update.go` (importDryRunScreen, handleImportDryRunKey), `cmd/tui/view.go` (importDryRunView) |
| Export filter screen | `cmd/tui/model.go` (exportFilteredCmd, exportFilter field), `cmd/tui/update.go` (exportPickScreen, handleExportPickKey), `cmd/tui/view.go` (exportPickView) |
---

## Phase 2: Testing & Hardening ✅

> Completed — all items implemented.

### What was done

| Category | Item |
|---|---|
| **Storage hardening** | WAL mode (`PRAGMA journal_mode=WAL`) |
| | Foreign key enforcement (`PRAGMA foreign_keys=ON`) |
| | Composite index `(feed_id, published_at DESC)` |
| | DB `Ping()` on open to verify connection |
| | Migration ALTER TABLE errors now caught (`isDupColumnError` ignores re-run) |
| | Time parse errors propagated instead of silently swallowed |
| | `nullIfEmpty()` helper for NULL folder_id storage |
| | `scanFeed`/`scanEntry` return errors on bad timestamps |
| | Folder ID uses `sql.NullString` for NULL-safe scanning |
| **Completion scripts** | Bash: all commands, `folder` subcommands, `import` file completion, `export --output` |
| | Zsh: all commands, `folder` subcommands, `import`/`export` flags |
| **New tests** | Storage: `TestListEntriesAllFeeds`, `TestListEntriesZeroLimit`, `TestDeleteFeedCascadeDeletesEntries`, `TestUpdateFeed`, `TestListEntriesUnreadAllFeeds`, `TestDeleteFolder_FeedFolderBecomesNull` |
| | Config: `TestLoadWithTUIConfig`, `TestSetDefaults` extends TUI checks |
| | Feedtracker: `TestAddFeed_NetworkError`, `TestAddFeed_MalformedFeed`, `TestFetchFeed_NetworkError` |
| | Parser: `TestParse_EmptyBody`, `TestParse_NoItems`, `TestParse_MalformedDates` |

### Not done (deferred)

- **Context.Background() everywhere** — still `context.Background()` at handler level. Adding timeouts/deadlines is non-trivial and touches CLI/TUI entry points. Worth doing in a dedicated phase.
- **Domain tests** — trivial getter structs, low value.

---

## Phase 3: Large List Performance ✅

> Completed — see below for implementation summary.

### What was done

| Item | Implementation |
|---|---|
| **Entry list without JOIN** | `sqlite.go` — `entryCols`/`entryColsPrefixed` constants; no LEFT JOIN when feedID is known (avoids `COALESCE(f.title)` scan per row) |
| **Bounded concurrency** | `tracker.go` — `FetchAllFeeds` uses a worker pool (`chan struct{}` semaphore + `sync.WaitGroup`). Configurable via `fetch_concurrency` default 3 |
| **Feed staleness check** | `tracker.go` — `shouldFetch()` skips feeds fetched within `fetch_cooldown` (default 0 = always fetch) |
| **Cursor pagination in TUI** | `model.go` — `entryOffset`, `entryPageSize`, `loadMoreEntriesCmd`; `update.go` — `moreEntriesLoadedMsg` appends results; `view.go` — `[L]` button in entry list; `L` key loads next page |
| **Offset support in storage** | `storage.go` + `sqlite.go` — `ListEntries`/`ListEntriesUnread` accept `offset int` with `LIMIT ? OFFSET ?` |

### Config additions

```yaml
http:
  fetch_concurrency: 3   # concurrent fetches during "fetch all"
  fetch_cooldown: 5m     # skip feeds fetched within this duration

tui:
  entry_page_size: 50    # entries loaded per page (from entry_limit)
```

### Not done

- **Lazy content/summary loading** — too much complexity for marginal gain; JOIN elimination already covers the main query cost
- **Virtualized feed list** — major refactor of `buildDisplayItems`; worth doing if feed count exceeds ~200
- **FTS5 full-text search** — deferred to a search/filter phase

---

## Phase 3b: Build Tooling (Makefile) ✅

> Completed — standard Makefile at project root.

### What was done

| Target | Command | Purpose |
|---|---|---|
| `build` | `go build -o ./bin/ ./cmd/...` | Build both `cli` and `tui` into `./bin/` |
| `test` | `go test -race ./... -count=1` | All tests with race detector |
| `vet` | `go vet ./...` | Static analysis |
| `lint` | `golangci-lint run` (if installed) | Linting |
| `tidy` | `go mod tidy` | Tidy module dependencies |
| `clean` | `rm -rf ./bin` | Remove build artifacts |
| `run-cli` | `go run ./cmd/cli/...` | Quick CLI run |
| `run-tui` | `go run ./cmd/tui/...` | Quick TUI run |
| `install` | `go install ./cmd/...` | Install to `$GOPATH/bin` |
| `all` | build + vet + test | CI-style workflow |
| `pre-push` hook | `make all` before every push | Enforced via `core.hooksPath` |

### Files changed

| File | Change |
|---|---|
| `Makefile` | New file at project root |
| `.githooks/pre-push` | New hook — runs `make all` before push |
| `README.md` | Development section now references Makefile targets |
| `PROGRESS.md` | Steps 54–56 added |
| `tasks/plan.md` | This section added, phases renumbered |

---

## Phase 4: Search & Bulk Mark-Read ✅

> Completed — see below for implementation summary.

### What was done

| Feature | Files changed |
|---|---|
| `SearchEntries` (LIKE-based) | `internal/storage/storage.go`, `internal/storage/sqlite.go` |
| `MarkFeedRead` / `MarkAllRead` | `internal/storage/storage.go`, `internal/storage/sqlite.go` |
| `ft search <query>` CLI command | `cmd/cli/search.go` (new), `cmd/cli/main.go` |
| `ft read --all / --feed / --feed-id` | `cmd/cli/read.go` |
| `ft list --search <q>` | `cmd/cli/list.go` |
| TUI search screen (`s` key) | `cmd/tui/model.go`, `cmd/tui/update.go`, `cmd/tui/view.go` |
| TUI bulk mark-read (`a` / `A` keys) | `cmd/tui/update.go`, `cmd/tui/model.go` |
| Storage tests | `internal/storage/sqlite_test.go` (6 new tests) |
| Docs | `README.md`, `tasks/plan.md`, `tasks/PROGRESS.md` |

### Remaining gaps (for future phases)

- FTS5 full-text search (LIKE covers basic case)
- Entry filter by feed while in All Entries view
- Keyboard-driven feed switching

---

## Phase 5: Hardening & Polish ✅

> Completed — all items implemented.

| Item | Approach |
|---|---|
| **Context timeouts (storage layer)** | `Fetch` now accepts `context.Context`; `FetchWithTimeout` removed. CLI root context wraps config timeout via `context.WithTimeout`. TUI all 22 command functions pass timeout from config. Bare `context.Background()` eliminated from TUI commands. |
| **Database maintenance** | `Vacuum(ctx)` and `Optimize(ctx)` added to storage interface + SQLite. `ft vacuum` and `ft db optimize` CLI subcommands added. |

---

## Phase 6: Automatic Entry Pruning ✅

> Completed — all items implemented.

### What was done

| Item | Files changed |
|---|---|
| **Config: `prune.max_age`** | `internal/config/config.go`, `config.example.yaml` |
| **Storage: `DeleteEntriesOlderThan`** | `internal/storage/storage.go`, `internal/storage/sqlite.go` |
| **CLI: `ft prune`** | `cmd/cli/prune.go` (new), `cmd/cli/main.go` |
| **Auto-prune after fetch** | `internal/feedtracker/tracker.go` (Prune method), `cmd/cli/fetch.go` |
| **Tests** | `internal/storage/sqlite_test.go` (3 new tests: basic, zero-age, all-deleted) |
| **Docs** | `tasks/plan.md`, `tasks/PROGRESS.md`, `docs/cli.md`, `docs/config.md` |

> Future: per-feed-type and per-feed pruning controls deferred to Phase 9.

---

## Phase 7: Feed & Entry Management ✅

> Completed — all items implemented.

### What was done

| Item | Files changed |
|---|---|
| **CLI: `ft feed update`** | `cmd/cli/feed.go` (new), `cmd/cli/main.go`, `cmd/cli/completion.go` |
| **Storage: `GetEntry`** | `internal/storage/storage.go`, `internal/storage/sqlite.go`, `internal/storage/sqlite_test.go` |
| **CLI: `ft open`** | `cmd/cli/open.go` (new), `cmd/cli/main.go`, `cmd/cli/completion.go` |
| **TUI: feed edit screen (`E` key)** | `cmd/tui/model.go`, `cmd/tui/update.go`, `cmd/tui/view.go` |
| **TUI: entry filter by feed (`f` key)** | `cmd/tui/model.go`, `cmd/tui/update.go`, `cmd/tui/view.go` |
| **TUI: feed switching (`[` / `]`)** | `cmd/tui/update.go`, `cmd/tui/view.go` |
| **Tests** | `internal/storage/sqlite_test.go` (TestGetEntry, TestGetEntry_Missing) |
| **Docs** | `tasks/plan.md`, `tasks/PROGRESS.md`, `docs/cli.md` |

---

## Phase 8: Performance & Polish ✅

> Completed — all items implemented.

### What was done

| Item | Files changed |
|---|---|
| **Lazy content/summary loading** | `internal/storage/sqlite.go` — split `entryCols` into list/detail variants; list queries skip `summary, content` columns |
| **On-screen key hints** | `cmd/tui/view.go` — expanded header hints on all three main screens to show all available keybindings |
| **TUI auto-refresh countdown** | `cmd/tui/model.go` (countdown field + ticker), `cmd/tui/update.go` (decrement + reset), `cmd/tui/view.go` (status bar display) |
| **Context-aware help screen** | `cmd/tui/view.go` — `helpView()` now filters keybindings by `m.prevScreen` (feed list, entries list, or entry detail) |
| **Domain package tests** | `internal/domain/entry_test.go`, `feed_test.go`, `folder_test.go` — basic construction and field access tests |
| **Edit feed from entries list** | Already implemented in Phase 7 (`cmd/tui/update.go:478`) — marked done |
| **Centralized helpBinding refactor** | `cmd/tui/view.go` — all unique `{key, desc}` pairs defined once as package-level `var`s, referenced by pointer in all slices. Removed `label` field from `helpBinding`; single `desc` field used by both hints and help view. |
| **Width-adaptive hint lines** | `cmd/tui/view.go` — hints moved to their own line, wrapped at terminal width via `helpStyle.Width()`. Removed `maxWidth` truncation from `renderHintLine`. |
| **Dynamic help box width** | `cmd/tui/view.go` — help screen width computed from terminal width (capped at 60). Removed dead `Width(40)` from `styles.go`. |
| **Help on entry detail** | `cmd/tui/update.go` — added `"?"` handler to `handleEntryDetailKey`. |
| **Feed list help fixes** | `cmd/tui/view.go` — removed `bindingPrevNext`, `bindingSearch`, `bindingToggleRead`, `bindingLoadMore`, `bindingAllEntries` from `feedListBindings` (not applicable to feed list). |
| **Dead code removal** | Deleted `cmd/tui/keys.go` (entire file unused). Removed unused `markEntryUnreadCmd` from `cmd/tui/model.go`. Removed unused `lipgloss` import from `cmd/tui/view.go`. |
| **Secondary views use bindings** | `cmd/tui/view.go` — all 11 modal views (add feed, edit feed, folder create/rename/pick, import, export pick, feed pick, search) use `renderHintLine()` with `helpBinding` vars instead of hardcoded strings. |
| **globalBindings fixed** | `cmd/tui/view.go` — added `Esc`/Back to help screen global section. |

### Remaining (deferred)

| Item | Notes |
|---|---|
| **FTS5 full-text search** | Upgrade from LIKE-based search to SQLite FTS5. Requires migration, trigger management, and config. |
| **Virtualized feed list** | Refactor `buildDisplayItems` to only render visible rows. Pays off at 200+ feeds. |

---

## Phase 9: view.go Refactor ✅

> Completed — see below for implementation summary.

### What was done

| Item | Files changed |
|---|---|
| **Create `bindings.go`** | `cmd/tui/bindings.go` — helpBinding struct, all 46 binding vars, help sections, hint slices, renderHintLine, renderHelpSection (~130 lines) |
| **Create `help.go`** | `cmd/tui/help.go` — helpView() method (~40 lines) |
| **Create `views.go`** | `cmd/tui/views.go` — all 10 secondary modal views (editFeed, addFeed, folder CRUD, import, export, feedPick, search) (~210 lines) |
| **Trim `view.go`** | `cmd/tui/view.go` — 954 → 541 lines. Keeps View(), feedsListView, entriesListView, entryDetailView, entryDetailContent, statusBar, and utility helpers. Removed `sort` and `internal/opml` imports. |

---

## Phase 10: Granular Pruning Controls ✅

> Completed — all items implemented.

### What was done

| Item | Files changed |
|---|---|
| **Config: `prune.overrides.type`** | `internal/config/config.go`, `config.example.yaml`, `config.yaml` |
| **Domain: `Feed.MaxAge` field** | `internal/domain/feed.go` |
| **Storage: `max_age` column + migration** | `internal/storage/migrations.go`, `internal/storage/sqlite.go` |
| **Storage: `DeleteEntriesOlderThanForFeed`** | `internal/storage/storage.go`, `internal/storage/sqlite.go` |
| **Tracker: granular `Prune()`** | `internal/feedtracker/tracker.go` — iterates feeds, resolves feed-level > type-level > global max_age |
| **CLI: `ft feed update --prune-age`** | `cmd/cli/feed.go` |
| **TUI: MaxAge field in feed edit** | `cmd/tui/model.go`, `cmd/tui/update.go`, `cmd/tui/views.go` |
| **Tests** | `internal/storage/sqlite_test.go` (2 new tests), `internal/domain/feed_test.go` (MaxAge field check) |
| **Docs** | `docs/cli.md`, `docs/config.md`, `tasks/plan.md`, `tasks/PROGRESS.md` |
---

## Phase 11: Star/Bookmark Entries

Add a `Starred` field so users can save important entries for later, accessible from both CLI and TUI.

| Item | Files |
|---|---|
| **Domain**: `Starred bool` on `Entry` | `internal/domain/entry.go` |
| **Storage**: migration, `StarEntry`/`UnstarEntry`/`ListStarredEntries`, update `scanEntry` | `internal/storage/storage.go`, `sqlite.go`, `sqlite_test.go` |
| **CLI**: `ft star <id>`, `ft unstar <id>`, `ft list --starred` | `cmd/cli/star.go`, `unstar.go`, `list.go`, `main.go` |
| **TUI entry detail**: `s` key toggles star/unstar | `cmd/tui/update.go`, `view.go` |
| **TUI entry list**: star indicator, starred style, `S` key for starred-only filter | `cmd/tui/view.go`, `update.go`, `model.go`, `styles.go` |
| **Help + bindings**: update for new keys | `cmd/tui/bindings.go`, `help.go` |

---

## Phase 12: Manual Entry Deletion

Allow deleting individual entries from the TUI and CLI — complements star (save what matters, delete noise).

| Item | Files |
|---|---|
| **Storage**: `DeleteEntry(ctx, id)` method + test | `internal/storage/storage.go`, `sqlite.go`, `sqlite_test.go` |
| **CLI**: `ft delete-entry <id>` | `cmd/cli/delete-entry.go`, `main.go` |
| **TUI entry list**: `d` key deletes entry (with confirmation) | `cmd/tui/update.go` |
| **TUI entry detail**: `d` key deletes entry, returns to list | `cmd/tui/update.go` |
| **Help + bindings**: update | `cmd/tui/bindings.go` |

---

## Phase 13: TUI Maintenance Operations

Add keybindings so users never need to drop to CLI for database maintenance.

| Item | Files |
|---|---|
| **TUI**: `P` key prunes old entries (shows count deleted in status bar) | `cmd/tui/update.go`, `model.go` |
| **TUI**: `V` key vacuums database | `cmd/tui/update.go`, `model.go` |
| **TUI**: `O` key optimizes database | `cmd/tui/update.go`, `model.go` |
| **Help + bindings**: update | `cmd/tui/bindings.go` |

---

## Phase 14: Desktop Notifications

Send OS-level notifications when auto-refresh (or manual fetch) finds new entries.

| Item | Files |
|---|---|
| **Dependency**: Add `beeep` for OS notifications | `go.mod` |
| **TUI**: After fetch completes, if new entries > 0, fire notification | `cmd/tui/update.go` |
| **Config**: `tui.notifications` toggle | `internal/config/config.go` |

---

## Phase 15: Themes / Color Customization

Move hardcoded color values from `styles.go` into the config file, letting users customize the look.

| Item | Files |
|---|---|
| **Config**: Add `tui.theme` section with color overrides | `internal/config/config.go`, `config.example.yaml` |
| **TUI**: Apply config theme in `styles.go`, falling back to defaults | `cmd/tui/styles.go` |

---

## Phase 16: Per-Feed Refresh Schedule

Allow overriding the global `auto_refresh` interval per feed.

| Item | Files |
|---|---|
| **Domain**: `RefreshInterval` field on Feed (0 = use global) | `internal/domain/feed.go` |
| **Storage**: migration + column + UpdateFeed field | `internal/storage/migrations.go`, `sqlite.go` |
| **TUI**: Feed edit screen adds refresh interval field | `cmd/tui/views.go`, `model.go`, `update.go` |
| **TUI**: Timer uses per-feed interval when set | `cmd/tui/model.go` |
| **CLI**: `ft feed update --refresh-interval` | `cmd/cli/feed.go` |

---

## Phase 17: FTS5 Full-Text Search

Upgrade from LIKE-based search to SQLite FTS5 for better performance and relevance.

| Item | Files |
|---|---|
| **Storage**: FTS5 virtual table migration + triggers | `internal/storage/migrations.go` |
| **Storage**: Replace SearchEntries LIKE with FTS5 MATCH | `internal/storage/sqlite.go` |
| **Storage**: Backfill existing entries into FTS index | `internal/storage/migrations.go` |
| **Tests**: Update search tests for FTS5 | `internal/storage/sqlite_test.go` |

---

## Recommended Order

1. ✅ ~~Auto-refresh in TUI + feature gaps~~ (Phase 1 complete)
2. ✅ ~~FK enforcement + WAL mode + composite index~~ (Phase 2 complete)
3. ✅ ~~Fill test gaps~~ (Phase 2 complete)
4. ✅ ~~Context deadlines + hardened error handling~~ (Phase 2 complete)
5. ✅ ~~Bounded concurrent fetching~~ (Phase 3 complete)
6. ✅ ~~Cursor pagination~~ (Phase 3 complete)
7. ✅ ~~Makefile + build tooling~~ (Phase 3b complete)
8. ✅ ~~Search, filters & bulk mark-read~~ (Phase 4 complete)
9. ✅ ~~Phase 5~~ — Hardening & polish (completed)
10. ✅ ~~Phase 6~~ — Automatic entry pruning (completed)
11. ✅ ~~Phase 7~~ — Feed & entry management (completed)
12. ✅ ~~Phase 8~~ — Performance & polish (completed)
13. ✅ ~~Phase 9 — view.go refactor (split into bindings.go, help.go, views.go)~~
14. ✅ ~~Phase 10 — Granular pruning controls (per-feed-type, per-feed)~~
15. **Phase 11** — Star/Bookmark entries
16. **Phase 12** — Manual entry deletion
17. **Phase 13** — TUI maintenance operations (prune/vacuum/optimize)
18. **Phase 14** — Desktop notifications
19. **Phase 15** — Themes / color customization
20. **Phase 16** — Per-feed refresh schedule
21. **Phase 17** — FTS5 full-text search
