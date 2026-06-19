package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return nil, fmt.Errorf("enable wal: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &sqliteStorage{db: db}, nil
}

func (s *sqliteStorage) Close() error {
	return s.db.Close()
}

func (s *sqliteStorage) Migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	for _, alter := range []string{
		`ALTER TABLE entries ADD COLUMN read INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE feeds ADD COLUMN folder_id TEXT DEFAULT '' REFERENCES folders(id) ON DELETE SET NULL`,
	} {
		if _, err := s.db.ExecContext(ctx, alter); err != nil {
			// Ignore "duplicate column" errors on re-run
			if !isDupColumnError(err) {
				return fmt.Errorf("migrate alter: %w", err)
			}
		}
	}
	return nil
}

func isDupColumnError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate column") ||
		strings.Contains(err.Error(), "already exists"))
}

func (s *sqliteStorage) AddFeed(ctx context.Context, feed *domain.Feed) error {
	const q = `INSERT INTO feeds (id, title, description, site_url, feed_url, feed_type, etag, last_modified, folder_id, created_at, updated_at, last_fetched)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, q,
		feed.ID, feed.Title, feed.Description, feed.SiteURL, feed.FeedURL,
		string(feed.FeedType), feed.ETag, feed.LastModified,
		nullIfEmpty(feed.FolderID),
		feed.CreatedAt.Format(time.RFC3339), feed.UpdatedAt.Format(time.RFC3339),
		feed.LastFetched.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("add feed: %w", err)
	}
	return nil
}

const feedCols = `id, title, description, site_url, feed_url, feed_type, etag, last_modified, folder_id, created_at, updated_at, last_fetched`

func (s *sqliteStorage) GetFeed(ctx context.Context, id string) (*domain.Feed, error) {
	const q = `SELECT ` + feedCols + ` FROM feeds WHERE id = ?`
	row := s.db.QueryRowContext(ctx, q, id)
	return scanFeed(row)
}

func (s *sqliteStorage) GetFeedByURL(ctx context.Context, url string) (*domain.Feed, error) {
	const q = `SELECT ` + feedCols + ` FROM feeds WHERE feed_url = ?`
	row := s.db.QueryRowContext(ctx, q, url)
	return scanFeed(row)
}

func (s *sqliteStorage) GetFeedByTitle(ctx context.Context, title string) (*domain.Feed, error) {
	const q = `SELECT ` + feedCols + ` FROM feeds WHERE title = ?`
	row := s.db.QueryRowContext(ctx, q, title)
	return scanFeed(row)
}

func (s *sqliteStorage) ListFeeds(ctx context.Context) ([]*domain.Feed, error) {
	const q = `SELECT ` + feedCols + ` FROM feeds ORDER BY title`
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
	const q = `UPDATE feeds SET title=?, description=?, site_url=?, feed_type=?, etag=?, last_modified=?, folder_id=?, updated_at=?, last_fetched=? WHERE id=?`
	_, err := s.db.ExecContext(ctx, q,
		feed.Title, feed.Description, feed.SiteURL, string(feed.FeedType),
		feed.ETag, feed.LastModified, nullIfEmpty(feed.FolderID),
		feed.UpdatedAt.Format(time.RFC3339), feed.LastFetched.Format(time.RFC3339),
		feed.ID,
	)
	if err != nil {
		return fmt.Errorf("update feed: %w", err)
	}
	return nil
}

func (s *sqliteStorage) DeleteFeed(ctx context.Context, id string) error {
	const q = `DELETE FROM feeds WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete feed: %w", err)
	}
	return nil
}

func (s *sqliteStorage) UpsertEntry(ctx context.Context, entry *domain.Entry) (bool, error) {
	const q = `INSERT INTO entries (id, feed_id, external_id, title, url, summary, content, author, published_at, updated_at, fetched_at, read)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
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

const entryCols = `id, feed_id, external_id, title, url, summary, content, author, published_at, updated_at, fetched_at, read`
const entryColsPrefixed = `e.id, e.feed_id, e.external_id, e.title, e.url, e.summary, e.content, e.author, e.published_at, e.updated_at, e.fetched_at, e.read`

func (s *sqliteStorage) ListEntries(ctx context.Context, feedID string, limit, offset int) ([]*domain.Entry, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var rows *sql.Rows
	var err error
	if feedID == "" {
		q := `SELECT ` + entryColsPrefixed + `, COALESCE(f.title, '')
			FROM entries e LEFT JOIN feeds f ON e.feed_id = f.id
			ORDER BY e.published_at DESC LIMIT ? OFFSET ?`
		rows, err = s.db.QueryContext(ctx, q, limit, offset)
	} else {
		q := `SELECT ` + entryCols + `, '' as feed_title
			FROM entries
			WHERE feed_id = ?
			ORDER BY published_at DESC LIMIT ? OFFSET ?`
		rows, err = s.db.QueryContext(ctx, q, feedID, limit, offset)
	}
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

func (s *sqliteStorage) ListEntriesUnread(ctx context.Context, feedID string, limit, offset int) ([]*domain.Entry, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var rows *sql.Rows
	var err error
	if feedID == "" {
		q := `SELECT ` + entryColsPrefixed + `, COALESCE(f.title, '')
			FROM entries e LEFT JOIN feeds f ON e.feed_id = f.id
			WHERE e.read = 0
			ORDER BY e.published_at DESC LIMIT ? OFFSET ?`
		rows, err = s.db.QueryContext(ctx, q, limit, offset)
	} else {
		q := `SELECT ` + entryCols + `, '' as feed_title
			FROM entries
			WHERE feed_id = ? AND read = 0
			ORDER BY published_at DESC LIMIT ? OFFSET ?`
		rows, err = s.db.QueryContext(ctx, q, feedID, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("list entries unread: %w", err)
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

func (s *sqliteStorage) UnreadCountByFeed(ctx context.Context) (map[string]int, error) {
	const q = `SELECT feed_id, COUNT(*) FROM entries WHERE read = 0 GROUP BY feed_id`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("unread count by feed: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var feedID string
		var n int
		if err := rows.Scan(&feedID, &n); err != nil {
			return nil, fmt.Errorf("scan unread count: %w", err)
		}
		counts[feedID] = n
	}
	return counts, rows.Err()
}

func (s *sqliteStorage) MarkEntryRead(ctx context.Context, entryID string) error {
	const q = `UPDATE entries SET read = 1 WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, entryID)
	if err != nil {
		return fmt.Errorf("mark entry read: %w", err)
	}
	return nil
}

func (s *sqliteStorage) MarkEntryUnread(ctx context.Context, entryID string) error {
	const q = `UPDATE entries SET read = 0 WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, entryID)
	if err != nil {
		return fmt.Errorf("mark entry unread: %w", err)
	}
	return nil
}

func (s *sqliteStorage) MarkFeedRead(ctx context.Context, feedID string) error {
	const q = `UPDATE entries SET read = 1 WHERE feed_id = ? AND read = 0`
	_, err := s.db.ExecContext(ctx, q, feedID)
	if err != nil {
		return fmt.Errorf("mark feed read: %w", err)
	}
	return nil
}

func (s *sqliteStorage) MarkAllRead(ctx context.Context) error {
	const q = `UPDATE entries SET read = 1 WHERE read = 0`
	_, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}

func (s *sqliteStorage) SearchEntries(ctx context.Context, query string, limit, offset int) ([]*domain.Entry, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	pattern := "%" + query + "%"
	q := `SELECT ` + entryColsPrefixed + `, COALESCE(f.title, '')
		FROM entries e LEFT JOIN feeds f ON e.feed_id = f.id
		WHERE e.title LIKE ? OR e.summary LIKE ?
		ORDER BY e.published_at DESC LIMIT ? OFFSET ?`
	rows, err := s.db.QueryContext(ctx, q, pattern, pattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search entries: %w", err)
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

func (s *sqliteStorage) DeleteEntriesOlderThan(ctx context.Context, age time.Duration) (int64, error) {
	if age <= 0 {
		return 0, nil
	}
	cutoff := time.Now().Add(-age)
	const q = `DELETE FROM entries WHERE published_at != '' AND published_at < ?`
	res, err := s.db.ExecContext(ctx, q, cutoff.Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("delete entries older than: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (s *sqliteStorage) Vacuum(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `VACUUM`)
	if err != nil {
		return fmt.Errorf("vacuum: %w", err)
	}
	return nil
}

func (s *sqliteStorage) Optimize(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `PRAGMA optimize`)
	if err != nil {
		return fmt.Errorf("optimize: %w", err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanFeed(row scanner) (*domain.Feed, error) {
	var f domain.Feed
	var feedType, createdAt, updatedAt, lastFetched string
	var folderID sql.NullString
	err := row.Scan(&f.ID, &f.Title, &f.Description, &f.SiteURL, &f.FeedURL,
		&feedType, &f.ETag, &f.LastModified, &folderID,
		&createdAt, &updatedAt, &lastFetched)
	if err != nil {
		return nil, fmt.Errorf("scan feed: %w", err)
	}
	f.FeedType = domain.FeedType(feedType)
	f.FolderID = folderID.String
	f.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse feed created_at: %w", err)
	}
	f.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse feed updated_at: %w", err)
	}
	f.LastFetched, err = time.Parse(time.RFC3339, lastFetched)
	if err != nil {
		return nil, fmt.Errorf("parse feed last_fetched: %w", err)
	}
	return &f, nil
}

func (s *sqliteStorage) AddFolder(ctx context.Context, folder *domain.Folder) error {
	const q = `INSERT INTO folders (id, name, created_at) VALUES (?, ?, ?)`
	_, err := s.db.ExecContext(ctx, q, folder.ID, folder.Name, folder.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("add folder: %w", err)
	}
	return nil
}

func (s *sqliteStorage) ListFolders(ctx context.Context) ([]*domain.Folder, error) {
	const q = `SELECT id, name, created_at FROM folders ORDER BY name`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}
	defer rows.Close()

	var folders []*domain.Folder
	for rows.Next() {
		var f domain.Folder
		var createdAt string
		if err := rows.Scan(&f.ID, &f.Name, &createdAt); err != nil {
			return nil, fmt.Errorf("scan folder: %w", err)
		}
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		folders = append(folders, &f)
	}
	return folders, rows.Err()
}

func (s *sqliteStorage) GetFolderByName(ctx context.Context, name string) (*domain.Folder, error) {
	const q = `SELECT id, name, created_at FROM folders WHERE name = ?`
	var f domain.Folder
	var createdAt string
	err := s.db.QueryRowContext(ctx, q, name).Scan(&f.ID, &f.Name, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get folder by name: %w", err)
	}
	f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &f, nil
}

func (s *sqliteStorage) DeleteFolder(ctx context.Context, id string) error {
	const q = `DELETE FROM folders WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}
	return nil
}

func (s *sqliteStorage) SetFeedFolder(ctx context.Context, feedID, folderID string) error {
	const q = `UPDATE feeds SET folder_id = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, nullIfEmpty(folderID), feedID)
	if err != nil {
		return fmt.Errorf("set feed folder: %w", err)
	}
	return nil
}

func scanEntry(row scanner) (*domain.Entry, error) {
	var e domain.Entry
	var publishedAt, updatedAt, fetchedAt string
	var read int
	err := row.Scan(&e.ID, &e.FeedID, &e.ExternalID, &e.Title, &e.URL,
		&e.Summary, &e.Content, &e.Author,
		&publishedAt, &updatedAt, &fetchedAt, &read, &e.FeedTitle)
	if err != nil {
		return nil, fmt.Errorf("scan entry: %w", err)
	}
	e.PublishedAt, err = parseTime(publishedAt)
	if err != nil {
		return nil, fmt.Errorf("parse entry published_at: %w", err)
	}
	e.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse entry updated_at: %w", err)
	}
	e.FetchedAt, err = time.Parse(time.RFC3339, fetchedAt)
	if err != nil {
		return nil, fmt.Errorf("parse entry fetched_at: %w", err)
	}
	e.Read = read != 0
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

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
