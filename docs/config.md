# Configuration

See `config.example.yaml`:

```yaml
database:
  path: ./data/feeds.db

http:
  timeout: 30s
  user_agent: "feed-tracker/0.1"
  fetch_concurrency: 3   # concurrent fetches during "fetch all"
  fetch_cooldown: 5m     # skip feeds fetched within this duration

tui:
  entry_limit: 100      # max entries loaded per page in TUI
  entry_page_size: 50   # entries loaded per page (aliased from entry_limit)
  auto_refresh: 5m      # auto-fetch interval (e.g. 5m, 30m, 0 to disable)

prune:
  max_age: 30d          # delete entries older than this (e.g. 30d, 90d, 0 to disable)
```

## Fields

### database

| Field | Description |
|---|---|
| `path` | Path to the SQLite database file |

### http

| Field | Default | Description |
|---|---|---|
| `timeout` | `30s` | HTTP request timeout |
| `user_agent` | `feed-tracker/0.1` | User-Agent header value |
| `fetch_concurrency` | `3` | Max concurrent fetches during "fetch all" |
| `fetch_cooldown` | `0` | Skip feeds fetched within this duration (e.g. `5m`, `30m`) |

### tui

| Field | Default | Description |
|---|---|---|
| `entry_limit` | `100` | Max entries loaded per page in TUI |
| `auto_refresh` | `0` | Auto-fetch interval (e.g. `5m`, `30m`, `0` to disable) |

### prune

| Field | Default | Description |
|-------|---------|-------------|
| `max_age` | `0` | Delete entries older than this duration (e.g. `30d`, `90d`). `0` disables pruning. |
