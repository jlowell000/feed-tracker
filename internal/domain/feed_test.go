package domain

import (
	"testing"
	"time"
)

func TestFeedDefaults(t *testing.T) {
	f := Feed{}
	if f.ID != "" {
		t.Errorf("expected empty ID, got %q", f.ID)
	}
	if f.FeedType != "" {
		t.Errorf("expected empty FeedType, got %q", f.FeedType)
	}
}

func TestFeedFields(t *testing.T) {
	now := time.Now()
	f := Feed{
		ID:           "f1",
		Title:        "Test Feed",
		Description:  "A test feed",
		SiteURL:      "https://example.com",
		FeedURL:      "https://example.com/feed.xml",
		FeedType:     FeedTypeRSS,
		ETag:         `"abc123"`,
		LastModified: "Mon, 01 Jan 2024 00:00:00 GMT",
		FolderID:     "folder1",
		CreatedAt:    now,
		UpdatedAt:    now,
		LastFetched:  now,
	}
	if f.ID != "f1" || f.Title != "Test Feed" || f.Description != "A test feed" {
		t.Errorf("basic fields mismatch")
	}
	if f.SiteURL != "https://example.com" || f.FeedURL != "https://example.com/feed.xml" {
		t.Errorf("url fields mismatch")
	}
	if f.FeedType != FeedTypeRSS {
		t.Errorf("expected RSS feed type")
	}
	if f.ETag != `"abc123"` || f.LastModified != "Mon, 01 Jan 2024 00:00:00 GMT" {
		t.Errorf("http fields mismatch")
	}
	if f.FolderID != "folder1" {
		t.Errorf("folder id mismatch")
	}
}

func TestFeedTypeConstants(t *testing.T) {
	if FeedTypeRSS != "rss" {
		t.Errorf("FeedTypeRSS should be 'rss'")
	}
	if FeedTypeAtom != "atom" {
		t.Errorf("FeedTypeAtom should be 'atom'")
	}
	if FeedTypeJSONFeed != "jsonfeed" {
		t.Errorf("FeedTypeJSONFeed should be 'jsonfeed'")
	}
	if FeedTypeActivityPub != "activitypub" {
		t.Errorf("FeedTypeActivityPub should be 'activitypub'")
	}
}
