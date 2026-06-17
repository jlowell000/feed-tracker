package feedtracker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/domain"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

const testRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>Test Feed</title>
<description>A test feed</description>
<link>https://example.com</link>
<item>
<title>Item 1</title>
<link>https://example.com/1</link>
<description>Desc 1</description>
<guid>uuid-1</guid>
<pubDate>Mon, 01 Jan 2024 00:00:00 GMT</pubDate>
</item>
<item>
<title>Item 2</title>
<link>https://example.com/2</link>
<description>Desc 2</description>
<guid>uuid-2</guid>
<pubDate>Tue, 02 Jan 2024 00:00:00 GMT</pubDate>
</item>
</channel>
</rss>`

const updatedRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>Test Feed</title>
<description>A test feed</description>
<link>https://example.com</link>
<item>
<title>Item 1</title>
<link>https://example.com/1</link>
<description>Desc 1</description>
<guid>uuid-1</guid>
<pubDate>Mon, 01 Jan 2024 00:00:00 GMT</pubDate>
</item>
<item>
<title>Item 2</title>
<link>https://example.com/2</link>
<description>Desc 2</description>
<guid>uuid-2</guid>
<pubDate>Tue, 02 Jan 2024 00:00:00 GMT</pubDate>
</item>
<item>
<title>Item 3</title>
<link>https://example.com/3</link>
<description>Desc 3</description>
<guid>uuid-3</guid>
<pubDate>Wed, 03 Jan 2024 00:00:00 GMT</pubDate>
</item>
</channel>
</rss>`

func newTestTracker(t *testing.T) (*Tracker, *httptest.Server) {
	t.Helper()

	store, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	if err := store.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"etag-v2"`)
		fmt.Fprint(w, testRSS)
	}))
	t.Cleanup(ts.Close)

	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Timeout:   5 * time.Second,
			UserAgent: "test/1.0",
		},
	}

	tracker := New(cfg, store)
	return tracker, ts
}

func TestAddFeed(t *testing.T) {
	ctx := context.Background()
	tracker, ts := newTestTracker(t)

	feed, err := tracker.AddFeed(ctx, ts.URL)
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	if feed.Title != "Test Feed" {
		t.Errorf("Title = %q, want %q", feed.Title, "Test Feed")
	}
	if feed.FeedType != domain.FeedTypeRSS {
		t.Errorf("FeedType = %q, want %q", feed.FeedType, domain.FeedTypeRSS)
	}
	if feed.FeedURL != ts.URL {
		t.Errorf("FeedURL = %q, want %q", feed.FeedURL, ts.URL)
	}
	if feed.ETag != `"etag-v2"` {
		t.Errorf("ETag = %q, want %q", feed.ETag, `"etag-v2"`)
	}
	if feed.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestAddFeed_DuplicateURL(t *testing.T) {
	ctx := context.Background()
	tracker, ts := newTestTracker(t)

	_, err := tracker.AddFeed(ctx, ts.URL)
	if err != nil {
		t.Fatalf("first AddFeed: %v", err)
	}

	_, err = tracker.AddFeed(ctx, ts.URL)
	if err == nil {
		t.Fatal("expected error for duplicate feed URL")
	}
}

func TestFetchFeed_NewEntries(t *testing.T) {
	ctx := context.Background()
	tracker, ts := newTestTracker(t)

	feed, err := tracker.AddFeed(ctx, ts.URL)
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	// First fetch adds entries on add, second fetch should see 0 new
	n, err := tracker.FetchFeed(ctx, feed)
	if err != nil {
		t.Fatalf("FetchFeed: %v", err)
	}
	if n != 0 {
		t.Errorf("new entries on refetch = %d, want 0", n)
	}
}

func TestFetchFeed_WithNewContent(t *testing.T) {
	ctx := context.Background()

	store, err := storage.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	store.Migrate(ctx)

	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.Header.Get("If-None-Match") == `"etag-v2"` {
			// Second request: serve updated content
			w.Header().Set("ETag", `"etag-v3"`)
			fmt.Fprint(w, updatedRSS)
			return
		}
		w.Header().Set("ETag", `"etag-v2"`)
		fmt.Fprint(w, testRSS)
	}))
	defer ts.Close()

	cfg := &config.Config{
		HTTP: config.HTTPConfig{Timeout: 5 * time.Second, UserAgent: "test/1.0"},
	}
	tracker := New(cfg, store)

	feed, err := tracker.AddFeed(ctx, ts.URL)
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	// Refetch should get new item (Item 3)
	n, err := tracker.FetchFeed(ctx, feed)
	if err != nil {
		t.Fatalf("FetchFeed: %v", err)
	}
	if n != 1 {
		t.Errorf("new entries = %d, want 1", n)
	}

	// Third fetch should have 0 new
	n, err = tracker.FetchFeed(ctx, feed)
	if err != nil {
		t.Fatalf("third FetchFeed: %v", err)
	}
	if n != 0 {
		t.Errorf("new entries on third fetch = %d, want 0", n)
	}
}

func TestFetchFeed_NotModified(t *testing.T) {
	ctx := context.Background()

	store, err := storage.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	store.Migrate(ctx)

	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount > 1 {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"etag-fixed"`)
		fmt.Fprint(w, testRSS)
	}))
	defer ts.Close()

	cfg := &config.Config{
		HTTP: config.HTTPConfig{Timeout: 5 * time.Second, UserAgent: "test/1.0"},
	}
	tracker := New(cfg, store)

	feed, err := tracker.AddFeed(ctx, ts.URL)
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	n, err := tracker.FetchFeed(ctx, feed)
	if err != nil {
		t.Fatalf("FetchFeed: %v", err)
	}
	if n != 0 {
		t.Errorf("new entries on 304 = %d, want 0", n)
	}
}

func TestFetchAllFeeds(t *testing.T) {
	ctx := context.Background()
	tracker, ts := newTestTracker(t)

	// Add two feeds (same server, but tracker stores them separately)
	feed1, err := tracker.AddFeed(ctx, ts.URL+"/feed1")
	if err != nil {
		t.Fatalf("AddFeed feed1: %v", err)
	}
	feed2, err := tracker.AddFeed(ctx, ts.URL+"/feed2")
	if err != nil {
		t.Fatalf("AddFeed feed2: %v", err)
	}

	// Verify they were stored
	feeds, _ := tracker.store.ListFeeds(ctx)
	if len(feeds) != 2 {
		t.Errorf("len(feeds) = %d, want 2", len(feeds))
	}

	n, err := tracker.FetchAllFeeds(ctx)
	if err != nil {
		t.Fatalf("FetchAllFeeds: %v", err)
	}
	if n != 0 {
		t.Errorf("new entries total = %d, want 0", n)
	}

	// Verify titles were properly persisted
	if feed1.Title != "Test Feed" || feed2.Title != "Test Feed" {
		t.Errorf("feed titles = %q, %q", feed1.Title, feed2.Title)
	}
}
