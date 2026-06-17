package feedtracker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/domain"
	"github.com/jlowell000/feed-tracker/internal/fetcher"
	"github.com/jlowell000/feed-tracker/internal/parser"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

type Tracker struct {
	cfg     *config.Config
	store   storage.Storage
	fetcher *fetcher.Fetcher
}

func New(cfg *config.Config, store storage.Storage) *Tracker {
	return &Tracker{
		cfg:     cfg,
		store:   store,
		fetcher: fetcher.New(cfg.HTTP),
	}
}

func (t *Tracker) AddFeed(ctx context.Context, feedURL string) (*domain.Feed, error) {
	result, err := t.fetcher.Fetch(feedURL, "", "")
	if err != nil {
		return nil, fmt.Errorf("fetch feed url: %w", err)
	}

	feedType := parser.DetectType(result.Body)
	parsed, err := parser.Parse(result.Body, feedURL, feedType)
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}

	now := time.Now()
	feed := parsed.Feed
	feed.ID = uuid.New().String()
	feed.ETag = result.ETag
	feed.LastModified = result.LastModified
	feed.CreatedAt = now
	feed.UpdatedAt = now
	feed.LastFetched = now

	if err := t.store.AddFeed(ctx, feed); err != nil {
		return nil, fmt.Errorf("store feed: %w", err)
	}

	for _, entry := range parsed.Entries {
		entry.ID = uuid.New().String()
		entry.FeedID = feed.ID
		entry.FetchedAt = now
		if _, err := t.store.UpsertEntry(ctx, entry); err != nil {
			log.Printf("warning: upsert entry: %v", err)
		}
	}

	return feed, nil
}

func (t *Tracker) FetchFeed(ctx context.Context, feed *domain.Feed) (int, error) {
	result, err := t.fetcher.Fetch(feed.FeedURL, feed.ETag, feed.LastModified)
	if err != nil {
		return 0, fmt.Errorf("fetch %s: %w", feed.FeedURL, err)
	}

	if result.Status == 304 {
		feed.LastFetched = time.Now()
		feed.UpdatedAt = time.Now()
		if err := t.store.UpdateFeed(ctx, feed); err != nil {
			log.Printf("warning: update feed after 304: %v", err)
		}
		return 0, nil
	}

	parsed, err := parser.Parse(result.Body, feed.FeedURL, feed.FeedType)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", feed.FeedURL, err)
	}

	now := time.Now()
	newCount := 0
	for _, entry := range parsed.Entries {
		entry.ID = uuid.New().String()
		entry.FeedID = feed.ID
		entry.FetchedAt = now
		isNew, err := t.store.UpsertEntry(ctx, entry)
		if err != nil {
			log.Printf("warning: upsert entry: %v", err)
			continue
		}
		if isNew {
			newCount++
		}
	}

	feed.Title = parsed.Feed.Title
	feed.Description = parsed.Feed.Description
	feed.SiteURL = parsed.Feed.SiteURL
	feed.ETag = result.ETag
	feed.LastModified = result.LastModified
	feed.LastFetched = now
	feed.UpdatedAt = now

	if err := t.store.UpdateFeed(ctx, feed); err != nil {
		log.Printf("warning: update feed after fetch: %v", err)
	}

	return newCount, nil
}

func (t *Tracker) FetchAllFeeds(ctx context.Context) (int, error) {
	feeds, err := t.store.ListFeeds(ctx)
	if err != nil {
		return 0, fmt.Errorf("list feeds: %w", err)
	}

	total := 0
	for _, feed := range feeds {
		n, err := t.FetchFeed(ctx, feed)
		if err != nil {
			log.Printf("error fetching %s (%s): %v", feed.Title, feed.FeedURL, err)
			continue
		}
		total += n
	}
	return total, nil
}
