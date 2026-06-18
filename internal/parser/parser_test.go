package parser

import (
	"strings"
	"testing"

	"github.com/jlowell000/feed-tracker/internal/domain"
)

func TestDetectType_RSS(t *testing.T) {
	body := []byte(`<rss version="2.0"><channel><title>Test</title></channel></rss>`)
	if got := DetectType(body); got != domain.FeedTypeRSS {
		t.Errorf("DetectType = %q, want %q", got, domain.FeedTypeRSS)
	}
}

func TestDetectType_Atom(t *testing.T) {
	body := []byte(`<feed xmlns="http://www.w3.org/2005/Atom"><title>Test</title></feed>`)
	if got := DetectType(body); got != domain.FeedTypeAtom {
		t.Errorf("DetectType = %q, want %q", got, domain.FeedTypeAtom)
	}
}

func TestDetectType_JSONFeed(t *testing.T) {
	body := []byte(`{"version": "https://jsonfeed.org/version/1.1", "title": "Test"}`)
	if got := DetectType(body); got != domain.FeedTypeJSONFeed {
		t.Errorf("DetectType = %q, want %q", got, domain.FeedTypeJSONFeed)
	}
}

func TestDetectType_ActivityPub(t *testing.T) {
	body := []byte(`{"@context": "https://www.w3.org/ns/activitystreams", "type": "OrderedCollection", "totalItems": 0, "orderedItems": []}`)
	if got := DetectType(body); got != domain.FeedTypeActivityPub {
		t.Errorf("DetectType = %q, want %q", got, domain.FeedTypeActivityPub)
	}
}

func TestParseRSS(t *testing.T) {
	rss := `<?xml version="1.0" encoding="UTF-8"?>
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
<author>author@example.com</author>
</item>
</channel>
</rss>`
	parsed, err := Parse([]byte(rss), "https://example.com/feed", domain.FeedTypeRSS)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed.Feed.Title != "Test Feed" {
		t.Errorf("Feed.Title = %q, want %q", parsed.Feed.Title, "Test Feed")
	}
	if len(parsed.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(parsed.Entries))
	}
	if parsed.Entries[0].Title != "Item 1" {
		t.Errorf("Entry.Title = %q, want %q", parsed.Entries[0].Title, "Item 1")
	}
	if parsed.Entries[0].ExternalID != "uuid-1" {
		t.Errorf("Entry.ExternalID = %q, want %q", parsed.Entries[0].ExternalID, "uuid-1")
	}
}

func TestParseAtom(t *testing.T) {
	atom := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
<title>Atom Feed</title>
<subtitle>Atom subtitle</subtitle>
<link href="https://example.com"/>
<entry>
<title>Atom Entry</title>
<link href="https://example.com/atom-1"/>
<summary>Atom summary</summary>
<id>atom-id-1</id>
<published>2024-01-01T00:00:00Z</published>
<author><name>Author Name</name></author>
</entry>
</feed>`
	parsed, err := Parse([]byte(atom), "https://example.com/atom", domain.FeedTypeAtom)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed.Feed.Title != "Atom Feed" {
		t.Errorf("Feed.Title = %q, want %q", parsed.Feed.Title, "Atom Feed")
	}
	if len(parsed.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(parsed.Entries))
	}
	if parsed.Entries[0].Author != "Author Name" {
		t.Errorf("Entry.Author = %q, want %q", parsed.Entries[0].Author, "Author Name")
	}
}

func TestParseJSONFeed(t *testing.T) {
	jf := `{
		"version": "https://jsonfeed.org/version/1.1",
		"title": "JSON Feed",
		"home_page_url": "https://example.com",
		"items": [
			{
				"id": "json-id-1",
				"url": "https://example.com/json-1",
				"title": "JSON Entry",
				"summary": "JSON summary",
				"content_text": "JSON content",
				"date_published": "2024-01-01T00:00:00Z"
			}
		]
	}`
	parsed, err := Parse([]byte(jf), "https://example.com/json", domain.FeedTypeJSONFeed)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed.Feed.Title != "JSON Feed" {
		t.Errorf("Feed.Title = %q, want %q", parsed.Feed.Title, "JSON Feed")
	}
	if len(parsed.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(parsed.Entries))
	}
}

func TestTruncate(t *testing.T) {
	s := "hello world"
	if got := truncate(s, 5); got != "hello…" {
		t.Errorf("truncate = %q, want %q", got, "hello…")
	}
	if got := truncate(s, 20); got != s {
		t.Errorf("truncate = %q, want %q", got, s)
	}
}

func TestParse_EmptyBody(t *testing.T) {
	_, err := Parse([]byte(""), "https://example.com/feed", domain.FeedTypeRSS)
	if err == nil {
		t.Error("expected error for empty body")
	}
}

func TestParse_NoItems(t *testing.T) {
	rss := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>Empty Feed</title>
<description>No items</description>
<link>https://example.com</link>
</channel>
</rss>`
	parsed, err := Parse([]byte(rss), "https://example.com/feed", domain.FeedTypeRSS)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(parsed.Entries) != 0 {
		t.Errorf("len(Entries) = %d, want 0", len(parsed.Entries))
	}
	if parsed.Feed.Title != "Empty Feed" {
		t.Errorf("Feed.Title = %q, want %q", parsed.Feed.Title, "Empty Feed")
	}
}

func TestParse_MalformedDates(t *testing.T) {
	rss := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>Test</title>
<item>
<title>Item 1</title>
<link>https://example.com/1</link>
<guid>uuid-1</guid>
<pubDate>not a date</pubDate>
</item>
</channel>
</rss>`
	parsed, err := Parse([]byte(rss), "https://example.com/feed", domain.FeedTypeRSS)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(parsed.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(parsed.Entries))
	}
	if !parsed.Entries[0].PublishedAt.IsZero() {
		t.Error("expected zero published_at for malformed date")
	}
}

func TestParseWithGofeed_InvalidXML(t *testing.T) {
	_, err := parseWithGofeed([]byte("not xml"), "")
	if err == nil {
		t.Error("expected error for invalid XML")
	}
	if err != nil && !strings.Contains(err.Error(), "gofeed parse") {
		t.Errorf("unexpected error: %v", err)
	}
}
