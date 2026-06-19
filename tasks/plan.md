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

## Phase 6: Automatic Entry Pruning

Global automatic deletion of entries based on age, replacing manual bulk delete.

| Item | Approach |
|---|---|
| **Config: `prune.max_age`** | Add to `config.example.yaml` (e.g. `30d`, `90d`, `0` to disable) |
| **Storage: `DeleteEntriesOlderThan`** | New method on storage interface + SQLite impl (`DELETE FROM entries WHERE published_at < ?`) |
| **CLI: `ft prune`** | New command to trigger manual prune |
| **Auto-prune after fetch** | Run prune automatically after `ft fetch` based on config |
| **Tests** | Storage prune query, CLI command |

> Future: per-feed-type and per-feed pruning controls deferred to Phase 9.

---

## Phase 7: Feed & Entry Management

| Item | Approach |
|---|---|
| **Feed title/URL editing** | CLI: `ft feed update <name> --title ... --url ...`. TUI: edit screen on feed select. |
| **Browser-open in CLI** | `ft open <entry-id>` — opens entry URL in system browser. |
| **Entry filter by feed** | Filter entries by feed while in All Entries view (TUI). |
| **Keyboard-driven feed switching** | Navigate between feeds via keyboard shortcuts in TUI. |

---

## Phase 8: Performance & Polish

| Item | Approach |
|---|---|
| **FTS5 full-text search** | Upgrade from LIKE-based search to SQLite FTS5 for performance. |
| **Virtualized feed list** | Refactor `buildDisplayItems` to only render visible rows. Pays off at 200+ feeds. |
| **Lazy content/summary** | Skip loading content/summary columns in list queries; load on detail view. |
| **TUI auto-refresh countdown** | Show time remaining until next auto-refresh in status bar. |
| **Domain package tests** | Low-value but completes test coverage for trivial getter structs. |

---

## Phase 9: Granular Pruning Controls

Per-feed-type and per-feed overrides for entry age-based deletion.

| Item | Approach |
|---|---|
| **Per-feed-type config** | Config overrides per feed type (`rss`, `atom`, `jsonfeed`, `activitypub`), e.g. `prune.overrides.type.activitypub.max_age: 7d` |
| **Per-feed config** | Add `max_age` column to feeds table or use feed metadata. |
| **CLI editing** | `ft feed update <name> --prune-age 14d` |
| **TUI editing** | Per-feed prune age setting in TUI feed view. |
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
10. **Phase 6** — Automatic entry pruning (global)
11. **Phase 7** — Feed & entry management
12. **Phase 8** — Performance & polish
13. **Phase 9** — Granular pruning controls (per-feed-type, per-feed)
