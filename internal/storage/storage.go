package storage

import (
	"context"

	"github.com/jlowell000/feed-tracker/internal/domain"
)

type Storage interface {
	Close() error

	Migrate(ctx context.Context) error

	AddFeed(ctx context.Context, feed *domain.Feed) error
	GetFeed(ctx context.Context, id string) (*domain.Feed, error)
	GetFeedByURL(ctx context.Context, url string) (*domain.Feed, error)
	ListFeeds(ctx context.Context) ([]*domain.Feed, error)
	UpdateFeed(ctx context.Context, feed *domain.Feed) error

	UpsertEntry(ctx context.Context, entry *domain.Entry) (bool, error)
	ListEntries(ctx context.Context, feedID string, limit int) ([]*domain.Entry, error)
}
