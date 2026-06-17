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
| 13 | Polish — error handling, tests, CI | ☐ | |
