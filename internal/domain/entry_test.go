package domain

import (
	"testing"
	"time"
)

func TestEntryDefaults(t *testing.T) {
	e := Entry{}
	if e.ID != "" {
		t.Errorf("expected empty ID, got %q", e.ID)
	}
	if e.Read {
		t.Error("expected Read to be false")
	}
}

func TestEntryFields(t *testing.T) {
	now := time.Now()
	e := Entry{
		ID:          "e1",
		FeedID:      "f1",
		FeedTitle:   "Test Feed",
		ExternalID:  "ext-1",
		Title:       "Test Entry",
		URL:         "https://example.com/entry",
		Summary:     "A summary",
		Content:     "<p>Content</p>",
		Author:      "Author",
		PublishedAt: now,
		UpdatedAt:   now,
		FetchedAt:   now,
		Read:        true,
	}
	if e.ID != "e1" || e.FeedID != "f1" || e.FeedTitle != "Test Feed" {
		t.Errorf("basic fields mismatch")
	}
	if e.Title != "Test Entry" || e.URL != "https://example.com/entry" {
		t.Errorf("title/url mismatch")
	}
	if e.Summary != "A summary" || e.Content != "<p>Content</p>" {
		t.Errorf("content fields mismatch")
	}
	if e.Author != "Author" {
		t.Errorf("author mismatch")
	}
	if !e.PublishedAt.Equal(now) {
		t.Errorf("published_at mismatch")
	}
	if !e.Read {
		t.Error("expected Read to be true")
	}
}
