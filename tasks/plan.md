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

### Remaining gaps (for future phases)

- Bulk operations (mark-all-read, delete-all-read)
- Feed title editing / URL updating
- Entry search/filter by keyword
- Database maintenance (vacuum, stats)
- Browser-open in CLI

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
- Bulk delete entries (`ft delete --read`)
- Entry filter by feed while in All Entries view
- Keyboard-driven feed switching

---

## Phase 5: Hardening & Polish ✅ (partial)

### Completed items

| Item | Approach |
|---|---|
| **Context timeouts (storage layer)** | `Fetch` now accepts `context.Context`; `FetchWithTimeout` removed. CLI root context wraps config timeout via `context.WithTimeout`. TUI all 22 command functions pass timeout from config. Bare `context.Background()` eliminated from TUI commands. |
| **Database maintenance** | `Vacuum(ctx)` and `Optimize(ctx)` added to storage interface + SQLite. `ft vacuum` and `ft db optimize` CLI subcommands added. |

### Remaining for Phase 5

| Item | Approach |
|---|---|
| **Virtualized feed list** | Refactor `buildDisplayItems` to only render visible rows. Worthwhile above ~200 feeds. |
| **Browser-open in CLI** | Add `feed-tracker open <id>` to open feed URL in browser. |
| **Feed title/URL editing** | Allow editing feed metadata from TUI and CLI. |
| **Lazy content/summary** | Skip loading `content`/`summary` columns in list queries; load on detail view. Currently low ROI since JOIN removal already covers main query cost. |

---

## Deferred Work (collected)

Items deferred across all completed phases, ordered by estimated value:

| Priority | Item | Phase | Notes |
|---|---|---|---|---|
| High | Context deadlines + timeouts | P2 | Still `context.Background()` everywhere. Non-trivial — touches all CLI/TUI entry points. |
| High | FTS5 full-text search | P4 | LIKE-based search implemented in Phase 4; upgrade to FTS5 for performance |
| High | Bulk delete read entries | P4 | Deferred from Phase 4; needs confirmation UX |
| Medium | Virtualized feed list | P3 | Major refactor of `buildDisplayItems`. Only pays off at 200+ feeds. |
| Medium | Database maintenance commands | P1 | Vacuum, stats, integrity check. |
| Low | Lazy content/summary loading | P3 | JOIN elimination already covers main query cost. |
| Low | Feed title/URL editing | P1 | Nice-to-have, low complexity. |
| Low | Browser-open in CLI | P1 | Trivial to add. |
| Low | Domain package tests | P2 | Trivial getter structs; low value. |

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
9. **Phase 5** — Hardening & polish
