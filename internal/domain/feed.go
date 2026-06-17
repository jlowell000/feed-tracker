package domain

import "time"

type FeedType string

const (
	FeedTypeRSS         FeedType = "rss"
	FeedTypeAtom        FeedType = "atom"
	FeedTypeJSONFeed    FeedType = "jsonfeed"
	FeedTypeActivityPub FeedType = "activitypub"
)

type Feed struct {
	ID           string
	Title        string
	Description  string
	SiteURL      string
	FeedURL      string
	FeedType     FeedType
	ETag         string
	LastModified string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastFetched  time.Time
}
