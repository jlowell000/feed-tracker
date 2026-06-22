package storage

import (
	"context"
	"time"

	"github.com/jlowell000/feed-tracker/internal/domain"
)

type Storage interface {
	Close() error

	Migrate(ctx context.Context) error

	AddFolder(ctx context.Context, folder *domain.Folder) error
	ListFolders(ctx context.Context) ([]*domain.Folder, error)
	GetFolderByName(ctx context.Context, name string) (*domain.Folder, error)
	DeleteFolder(ctx context.Context, id string) error
	SetFeedFolder(ctx context.Context, feedID, folderID string) error

	AddFeed(ctx context.Context, feed *domain.Feed) error
	GetFeed(ctx context.Context, id string) (*domain.Feed, error)
	GetEntry(ctx context.Context, id string) (*domain.Entry, error)
	GetFeedByURL(ctx context.Context, url string) (*domain.Feed, error)
	GetFeedByTitle(ctx context.Context, title string) (*domain.Feed, error)
	ListFeeds(ctx context.Context) ([]*domain.Feed, error)
	UpdateFeed(ctx context.Context, feed *domain.Feed) error
	DeleteFeed(ctx context.Context, id string) error

	UpsertEntry(ctx context.Context, entry *domain.Entry) (bool, error)
	ListEntries(ctx context.Context, feedID string, limit, offset int) ([]*domain.Entry, error)
	ListEntriesUnread(ctx context.Context, feedID string, limit, offset int) ([]*domain.Entry, error)
	MarkEntryRead(ctx context.Context, entryID string) error
	MarkEntryUnread(ctx context.Context, entryID string) error
	MarkFeedRead(ctx context.Context, feedID string) error
	MarkAllRead(ctx context.Context) error
	UnreadCountByFeed(ctx context.Context) (map[string]int, error)

	SearchEntries(ctx context.Context, query string, limit, offset int) ([]*domain.Entry, error)

	DeleteEntriesOlderThan(ctx context.Context, age time.Duration) (int64, error)
	DeleteEntriesOlderThanForFeed(ctx context.Context, feedID string, age time.Duration) (int64, error)

	Vacuum(ctx context.Context) error
	Optimize(ctx context.Context) error
}
