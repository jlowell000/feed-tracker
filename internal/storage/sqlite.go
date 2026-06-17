package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jlowell000/feed-tracker/internal/domain"

	_ "modernc.org/sqlite"
)

type sqliteStorage struct {
	db *sql.DB
}

func New(path string) (Storage, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	return &sqliteStorage{db: db}, nil
}

func (s *sqliteStorage) Close() error {
	return s.db.Close()
}

func (s *sqliteStorage) Migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

func (s *sqliteStorage) AddFeed(ctx context.Context, feed *domain.Feed) error {
	const q = `INSERT INTO feeds (id, title, description, site_url, feed_url, feed_type, etag, last_modified, created_at, updated_at, last_fetched)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, q,
		feed.ID, feed.Title, feed.Description, feed.SiteURL, feed.FeedURL,
		string(feed.FeedType), feed.ETag, feed.LastModified,
		feed.CreatedAt.Format(time.RFC3339), feed.UpdatedAt.Format(time.RFC3339),
		feed.LastFetched.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("add feed: %w", err)
	}
	return nil
}

func (s *sqliteStorage) GetFeed(ctx context.Context, id string) (*domain.Feed, error) {
	const q = `SELECT id, title, description, site_url, feed_url, feed_type, etag, last_modified, created_at, updated_at, last_fetched FROM feeds WHERE id = ?`
	row := s.db.QueryRowContext(ctx, q, id)
	return scanFeed(row)
}

func (s *sqliteStorage) GetFeedByURL(ctx context.Context, url string) (*domain.Feed, error) {
	const q = `SELECT id, title, description, site_url, feed_url, feed_type, etag, last_modified, created_at, updated_at, last_fetched FROM feeds WHERE feed_url = ?`
	row := s.db.QueryRowContext(ctx, q, url)
	return scanFeed(row)
}

func (s *sqliteStorage) ListFeeds(ctx context.Context) ([]*domain.Feed, error) {
	const q = `SELECT id, title, description, site_url, feed_url, feed_type, etag, last_modified, created_at, updated_at, last_fetched FROM feeds ORDER BY title`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list feeds: %w", err)
	}
	defer rows.Close()

	var feeds []*domain.Feed
	for rows.Next() {
		f, err := scanFeed(rows)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

func (s *sqliteStorage) UpdateFeed(ctx context.Context, feed *domain.Feed) error {
	const q = `UPDATE feeds SET title=?, description=?, site_url=?, feed_type=?, etag=?, last_modified=?, updated_at=?, last_fetched=? WHERE id=?`
	_, err := s.db.ExecContext(ctx, q,
		feed.Title, feed.Description, feed.SiteURL, string(feed.FeedType),
		feed.ETag, feed.LastModified,
		feed.UpdatedAt.Format(time.RFC3339), feed.LastFetched.Format(time.RFC3339),
		feed.ID,
	)
	if err != nil {
		return fmt.Errorf("update feed: %w", err)
	}
	return nil
}

func (s *sqliteStorage) UpsertEntry(ctx context.Context, entry *domain.Entry) (bool, error) {
	const q = `INSERT INTO entries (id, feed_id, external_id, title, url, summary, content, author, published_at, updated_at, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(feed_id, external_id) DO NOTHING`
	res, err := s.db.ExecContext(ctx, q,
		entry.ID, entry.FeedID, entry.ExternalID,
		entry.Title, entry.URL, entry.Summary, entry.Content,
		entry.Author,
		formatTime(entry.PublishedAt), formatTime(entry.UpdatedAt),
		entry.FetchedAt.Format(time.RFC3339),
	)
	if err != nil {
		return false, fmt.Errorf("upsert entry: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *sqliteStorage) ListEntries(ctx context.Context, feedID string, limit int) ([]*domain.Entry, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `SELECT id, feed_id, external_id, title, url, summary, content, author, published_at, updated_at, fetched_at
		FROM entries WHERE feed_id = ? ORDER BY published_at DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, q, feedID, limit)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.Entry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanFeed(row scanner) (*domain.Feed, error) {
	var f domain.Feed
	var feedType, createdAt, updatedAt, lastFetched string
	err := row.Scan(&f.ID, &f.Title, &f.Description, &f.SiteURL, &f.FeedURL,
		&feedType, &f.ETag, &f.LastModified,
		&createdAt, &updatedAt, &lastFetched)
	if err != nil {
		return nil, fmt.Errorf("scan feed: %w", err)
	}
	f.FeedType = domain.FeedType(feedType)
	f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	f.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	f.LastFetched, _ = time.Parse(time.RFC3339, lastFetched)
	return &f, nil
}

func scanEntry(row scanner) (*domain.Entry, error) {
	var e domain.Entry
	var publishedAt, updatedAt, fetchedAt string
	err := row.Scan(&e.ID, &e.FeedID, &e.ExternalID, &e.Title, &e.URL,
		&e.Summary, &e.Content, &e.Author,
		&publishedAt, &updatedAt, &fetchedAt)
	if err != nil {
		return nil, fmt.Errorf("scan entry: %w", err)
	}
	e.PublishedAt, _ = parseTime(publishedAt)
	e.UpdatedAt, _ = parseTime(updatedAt)
	e.FetchedAt, _ = time.Parse(time.RFC3339, fetchedAt)
	return &e, nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, s)
}
