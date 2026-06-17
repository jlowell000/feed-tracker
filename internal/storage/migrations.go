package storage

const schema = `
CREATE TABLE IF NOT EXISTS folders (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS feeds (
    id            TEXT PRIMARY KEY,
    title         TEXT NOT NULL DEFAULT '',
    description   TEXT NOT NULL DEFAULT '',
    site_url      TEXT NOT NULL DEFAULT '',
    feed_url      TEXT NOT NULL UNIQUE,
    feed_type     TEXT NOT NULL,
    etag          TEXT NOT NULL DEFAULT '',
    last_modified TEXT NOT NULL DEFAULT '',
    folder_id     TEXT DEFAULT '' REFERENCES folders(id) ON DELETE SET NULL,
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL,
    last_fetched  TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS entries (
    id           TEXT PRIMARY KEY,
    feed_id      TEXT NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    external_id  TEXT NOT NULL,
    title        TEXT NOT NULL DEFAULT '',
    url          TEXT NOT NULL DEFAULT '',
    summary      TEXT NOT NULL DEFAULT '',
    content      TEXT NOT NULL DEFAULT '',
    author       TEXT NOT NULL DEFAULT '',
    published_at TEXT NOT NULL DEFAULT '',
    updated_at   TEXT NOT NULL DEFAULT '',
    fetched_at   TEXT NOT NULL,
    read         INTEGER NOT NULL DEFAULT 0,
    UNIQUE(feed_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_entries_feed_id ON entries(feed_id);
CREATE INDEX IF NOT EXISTS idx_entries_published ON entries(published_at);
`
