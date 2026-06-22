package feedtracker

import (
	"context"
	"fmt"
	"log"
	"sync"
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

func (t *Tracker) shouldFetch(feed *domain.Feed) bool {
	cooldown := t.cfg.HTTP.FetchCooldown
	if cooldown <= 0 {
		return true
	}
	return time.Since(feed.LastFetched) >= cooldown
}

func (t *Tracker) AddFeed(ctx context.Context, feedURL string) (*domain.Feed, error) {
	result, err := t.fetcher.Fetch(ctx, feedURL, "", "")
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

func (t *Tracker) Prune(ctx context.Context) {
	globalMaxAge := time.Duration(t.cfg.Prune.MaxAge)
	if globalMaxAge <= 0 {
		return
	}

	feeds, err := t.store.ListFeeds(ctx)
	if err != nil {
		log.Printf("warning: auto-prune list feeds: %v", err)
		return
	}

	for _, feed := range feeds {
		age := t.effectiveMaxAge(feed)
		if age <= 0 {
			continue
		}
		n, err := t.store.DeleteEntriesOlderThanForFeed(ctx, feed.ID, age)
		if err != nil {
			log.Printf("warning: auto-prune feed %s: %v", feed.Title, err)
			continue
		}
		if n > 0 {
			log.Printf("auto-prune: removed %d entr%s from %q (older than %s)", n, map[bool]string{true: "y", false: "ies"}[n == 1], feed.Title, age)
		}
	}
}

func (t *Tracker) effectiveMaxAge(feed *domain.Feed) time.Duration {
	if feed.MaxAge != "" {
		d, err := parseDuration(feed.MaxAge)
		if err == nil && d > 0 {
			return d
		}
	}
	if t.cfg.Prune.Overrides.Type != nil {
		if d, ok := t.cfg.Prune.Overrides.Type[string(feed.FeedType)]; ok && time.Duration(d) > 0 {
			return time.Duration(d)
		}
	}
	return time.Duration(t.cfg.Prune.MaxAge)
}

func (t *Tracker) FetchFeed(ctx context.Context, feed *domain.Feed) (int, error) {
	result, err := t.fetcher.Fetch(ctx, feed.FeedURL, feed.ETag, feed.LastModified)
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

	t.Prune(ctx)

	return newCount, nil
}

func (t *Tracker) FetchAllFeeds(ctx context.Context) (int, error) {
	feeds, err := t.store.ListFeeds(ctx)
	if err != nil {
		return 0, fmt.Errorf("list feeds: %w", err)
	}

	concurrency := t.cfg.HTTP.FetchConcurrency
	if concurrency <= 0 {
		concurrency = 3
	}

	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	total := 0
	var wg sync.WaitGroup

	for _, feed := range feeds {
		if !t.shouldFetch(feed) {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(f *domain.Feed) {
			defer wg.Done()
			defer func() { <-sem }()

			n, err := t.FetchFeed(ctx, f)
			if err != nil {
				log.Printf("error fetching %s (%s): %v", f.Title, f.FeedURL, err)
				return
			}
			mu.Lock()
			total += n
			mu.Unlock()
		}(feed)
	}

	wg.Wait()
	return total, nil
}

func parseDuration(s string) (time.Duration, error) {
	if len(s) > 1 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
