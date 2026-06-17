package parser

import (
	"encoding/json"
	"fmt"

	"github.com/jlowell000/feed-tracker/internal/domain"
)

type activityPubOutbox struct {
	Context      any              `json:"@context"`
	Type         string           `json:"type"`
	TotalItems   int              `json:"totalItems"`
	OrderedItems []activityPubObj `json:"orderedItems"`
}

type activityPubObj struct {
	Type      string `json:"type"`
	Object    any    `json:"object"`
	Published string `json:"published"`
	Actor     any    `json:"actor"`
}

type activityPubNote struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	URL         string `json:"url"`
	AttributedTo any   `json:"attributedTo"`
	Summary     string `json:"summary"`
	Content     string `json:"content"`
	Published   string `json:"published"`
	Updated     string `json:"updated"`
}

func parseActivityPub(body []byte, feedURL string) (*ParsedFeed, error) {
	var outbox activityPubOutbox
	if err := json.Unmarshal(body, &outbox); err != nil {
		return nil, fmt.Errorf("ap json decode: %w", err)
	}

	feed := &domain.Feed{
		FeedURL:  feedURL,
		FeedType: domain.FeedTypeActivityPub,
		Title:    feedURL,
	}

	var entries []*domain.Entry
	for _, item := range outbox.OrderedItems {
		if item.Type != "Create" {
			continue
		}

		note, err := extractNote(item.Object)
		if err != nil {
			continue
		}

		entry := &domain.Entry{
			ExternalID:  note.ID,
			URL:         note.URL,
			Summary:     note.Summary,
			Content:     note.Content,
			PublishedAt: mustParseTime(note.Published),
			UpdatedAt:   mustParseTime(note.Updated),
		}

		if note.AttributedTo != nil {
			if s, ok := note.AttributedTo.(string); ok {
				entry.Author = s
			} else if m, ok := note.AttributedTo.(map[string]any); ok {
				if name, ok := m["name"].(string); ok {
					entry.Author = name
				}
			}
		}
		if entry.Author == "" && item.Actor != nil {
			if s, ok := item.Actor.(string); ok {
				entry.Author = s
			} else if m, ok := item.Actor.(map[string]any); ok {
				if name, ok := m["name"].(string); ok {
					entry.Author = name
				}
			}
		}

		if entry.ExternalID == "" {
			entry.ExternalID = entry.URL
		}
		if entry.Title == "" {
			entry.Title = truncate(entry.Summary, 80)
		}

		entries = append(entries, entry)
	}

	return &ParsedFeed{Feed: feed, Entries: entries}, nil
}

func extractNote(obj any) (*activityPubNote, error) {
	switch v := obj.(type) {
	case map[string]any:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var note activityPubNote
		if err := json.Unmarshal(b, &note); err != nil {
			return nil, err
		}
		if note.Type != "Note" && note.Type != "Article" {
			return nil, fmt.Errorf("unexpected object type: %s", note.Type)
		}
		return &note, nil
	default:
		return nil, fmt.Errorf("unexpected object type")
	}
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "…"
}
