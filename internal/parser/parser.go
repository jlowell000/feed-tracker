package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/jlowell000/feed-tracker/internal/domain"
	"github.com/mmcdole/gofeed"
)

type ParsedFeed struct {
	Feed    *domain.Feed
	Entries []*domain.Entry
}

func Parse(body []byte, feedURL string, feedType domain.FeedType) (*ParsedFeed, error) {
	switch feedType {
	case domain.FeedTypeActivityPub:
		return parseActivityPub(body, feedURL)
	default:
		return parseWithGofeed(body, feedURL)
	}
}

func DetectType(body []byte) domain.FeedType {
	trimmed := strings.TrimSpace(string(body))
	if strings.HasPrefix(trimmed, "{") {
		if strings.Contains(trimmed, `"@context"`) &&
			strings.Contains(trimmed, `"OrderedCollection"`) {
			return domain.FeedTypeActivityPub
		}
		return domain.FeedTypeJSONFeed
	}
	if strings.Contains(trimmed, `<rss`) {
		return domain.FeedTypeRSS
	}
	if strings.Contains(trimmed, `<feed`) {
		return domain.FeedTypeAtom
	}
	return domain.FeedTypeRSS
}

func parseWithGofeed(body []byte, feedURL string) (*ParsedFeed, error) {
	fp := gofeed.NewParser()
	gf, err := fp.ParseString(string(body))
	if err != nil {
		return nil, fmt.Errorf("gofeed parse: %w", err)
	}

	ft := domain.FeedTypeRSS
	if gf.FeedType == "atom" {
		ft = domain.FeedTypeAtom
	} else if gf.FeedType == "json" {
		ft = domain.FeedTypeJSONFeed
	}

	feed := &domain.Feed{
		Title:       gf.Title,
		Description: gf.Description,
		SiteURL:     gf.Link,
		FeedURL:     feedURL,
		FeedType:    ft,
	}

	var entries []*domain.Entry
	for _, item := range gf.Items {
		entry := &domain.Entry{
			ExternalID: item.GUID,
			Title:      item.Title,
			URL:        item.Link,
			Summary:    item.Description,
			Content:    item.Content,
		}
		if item.Author != nil {
			entry.Author = item.Author.Name
		}
		if item.PublishedParsed != nil {
			entry.PublishedAt = *item.PublishedParsed
		}
		if item.UpdatedParsed != nil {
			entry.UpdatedAt = *item.UpdatedParsed
		}
		if entry.ExternalID == "" {
			entry.ExternalID = entry.URL
		}
		entries = append(entries, entry)
	}

	if feed.Title == "" && len(gf.Items) > 0 {
		feed.Title = feedURL
	}

	return &ParsedFeed{Feed: feed, Entries: entries}, nil
}

func mustParseTime(s string) time.Time {
	for _, layout := range []string{
		time.RFC3339,
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}
