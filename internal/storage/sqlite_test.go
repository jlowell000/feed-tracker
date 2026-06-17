package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jlowell000/feed-tracker/internal/domain"
)

func newTestStore(t *testing.T) Storage {
	t.Helper()
	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return s
}

func TestAddAndGetFeed(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed := &domain.Feed{
		ID:      uuid.New().String(),
		Title:   "Test Feed",
		FeedURL: "https://example.com/feed",
		FeedType: domain.FeedTypeRSS,
	}
	if err := s.AddFeed(ctx, feed); err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	got, err := s.GetFeed(ctx, feed.ID)
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if got.Title != "Test Feed" {
		t.Errorf("Title = %q, want %q", got.Title, "Test Feed")
	}
}

func TestGetFeedByURL(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed := &domain.Feed{
		ID:      uuid.New().String(),
		Title:   "Test",
		FeedURL: "https://example.com/feed",
		FeedType: domain.FeedTypeRSS,
	}
	s.AddFeed(ctx, feed)

	got, err := s.GetFeedByURL(ctx, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GetFeedByURL: %v", err)
	}
	if got.Title != "Test" {
		t.Errorf("Title = %q, want %q", got.Title, "Test")
	}
}

func TestListFeeds(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	s.AddFeed(ctx, &domain.Feed{ID: uuid.New().String(), Title: "A", FeedURL: "https://a.com/feed", FeedType: domain.FeedTypeRSS})
	s.AddFeed(ctx, &domain.Feed{ID: uuid.New().String(), Title: "B", FeedURL: "https://b.com/feed", FeedType: domain.FeedTypeAtom})

	feeds, err := s.ListFeeds(ctx)
	if err != nil {
		t.Fatalf("ListFeeds: %v", err)
	}
	if len(feeds) != 2 {
		t.Errorf("len(feeds) = %d, want 2", len(feeds))
	}
}

func TestUpsertEntry(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	entry := &domain.Entry{
		ID:          uuid.New().String(),
		FeedID:      feedID,
		ExternalID:  "ext-1",
		Title:       "Entry 1",
		FetchedAt:   time.Now(),
	}

	isNew, err := s.UpsertEntry(ctx, entry)
	if err != nil {
		t.Fatalf("UpsertEntry: %v", err)
	}
	if !isNew {
		t.Error("expected new entry")
	}

	// Same external_id should not insert
	isNew, err = s.UpsertEntry(ctx, entry)
	if err != nil {
		t.Fatalf("second UpsertEntry: %v", err)
	}
	if isNew {
		t.Error("expected duplicate to not be new")
	}
}

func TestListEntries(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e1",
		Title: "Latest", PublishedAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e2",
		Title: "Older", PublishedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})

	entries, err := s.ListEntries(ctx, feedID, 10)
	if err != nil {
		t.Fatalf("ListEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if entries[0].Title != "Latest" {
		t.Errorf("first entry.Title = %q, want %q", entries[0].Title, "Latest")
	}
}
