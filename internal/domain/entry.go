package domain

import "time"

type Entry struct {
	ID          string
	FeedID      string
	ExternalID  string
	Title       string
	URL         string
	Summary     string
	Content     string
	Author      string
	PublishedAt time.Time
	UpdatedAt   time.Time
	FetchedAt   time.Time
}
