package parser

import (
	"testing"

	"github.com/jlowell000/feed-tracker/internal/domain"
)

func TestParseActivityPub(t *testing.T) {
	ap := `{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type": "OrderedCollection",
		"totalItems": 2,
		"orderedItems": [
			{
				"type": "Create",
				"published": "2024-01-01T00:00:00Z",
				"actor": {
					"type": "Person",
					"name": "Test Author"
				},
				"object": {
					"type": "Note",
					"id": "https://example.com/note/1",
					"url": "https://example.com/note/1",
					"content": "<p>Hello world</p>",
					"summary": "A test note",
					"published": "2024-01-01T00:00:00Z"
				}
			},
			{
				"type": "Create",
				"published": "2024-01-02T00:00:00Z",
				"actor": "https://example.com/user/1",
				"object": {
					"type": "Article",
					"id": "https://example.com/article/1",
					"url": "https://example.com/article/1",
					"content": "Full article content",
					"published": "2024-01-02T00:00:00Z"
				}
			},
			{
				"type": "Like",
				"object": "https://example.com/some-post"
			}
		]
	}`

	parsed, err := parseActivityPub([]byte(ap), "https://example.com/outbox")
	if err != nil {
		t.Fatalf("parseActivityPub: %v", err)
	}

	if parsed.Feed.FeedType != domain.FeedTypeActivityPub {
		t.Errorf("FeedType = %q, want %q", parsed.Feed.FeedType, domain.FeedTypeActivityPub)
	}
	if parsed.Feed.FeedURL != "https://example.com/outbox" {
		t.Errorf("FeedURL = %q, want %q", parsed.Feed.FeedURL, "https://example.com/outbox")
	}

	if len(parsed.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(parsed.Entries))
	}

	// First entry: Note
	e1 := parsed.Entries[0]
	if e1.ExternalID != "https://example.com/note/1" {
		t.Errorf("Entry[0].ExternalID = %q", e1.ExternalID)
	}
	if e1.Content != "<p>Hello world</p>" {
		t.Errorf("Entry[0].Content = %q", e1.Content)
	}
	if e1.Author != "Test Author" {
		t.Errorf("Entry[0].Author = %q, want %q", e1.Author, "Test Author")
	}
	if e1.PublishedAt.Year() != 2024 {
		t.Errorf("Entry[0].PublishedAt.Year = %d", e1.PublishedAt.Year())
	}

	// Second entry: Article
	e2 := parsed.Entries[1]
	if e2.ExternalID != "https://example.com/article/1" {
		t.Errorf("Entry[1].ExternalID = %q", e2.ExternalID)
	}
	if e2.Content != "Full article content" {
		t.Errorf("Entry[1].Content = %q", e2.Content)
	}
	if e2.Summary != "" {
		t.Errorf("Entry[1].Summary = %q, want empty", e2.Summary)
	}
	if e2.Author != "https://example.com/user/1" {
		t.Errorf("Entry[1].Author = %q, want actor URL", e2.Author)
	}
}

func TestParseActivityPub_EmptyOutbox(t *testing.T) {
	ap := `{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type": "OrderedCollection",
		"totalItems": 0,
		"orderedItems": []
	}`

	parsed, err := parseActivityPub([]byte(ap), "https://example.com/outbox")
	if err != nil {
		t.Fatalf("parseActivityPub: %v", err)
	}
	if len(parsed.Entries) != 0 {
		t.Errorf("len(Entries) = %d, want 0", len(parsed.Entries))
	}
}

func TestParseActivityPub_NonNoteObject(t *testing.T) {
	ap := `{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type": "OrderedCollection",
		"totalItems": 1,
		"orderedItems": [
			{
				"type": "Create",
				"object": {
					"type": "Video",
					"id": "https://example.com/video/1"
				}
			}
		]
	}`

	parsed, err := parseActivityPub([]byte(ap), "https://example.com/outbox")
	if err != nil {
		t.Fatalf("parseActivityPub: %v", err)
	}
	if len(parsed.Entries) != 0 {
		t.Errorf("len(Entries) = %d, want 0 (Video should be skipped)", len(parsed.Entries))
	}
}

func TestParseActivityPub_InvalidJSON(t *testing.T) {
	_, err := parseActivityPub([]byte("not json"), "")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
