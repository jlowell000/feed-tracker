package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/feedtracker"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runAdd(ctx context.Context, cfgPath string, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: feedtracker add <feed-url>")
		os.Exit(1)
	}
	feedURL := args[0]

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		os.Exit(1)
	}

	store, err := storage.New(cfg.Database.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	tracker := feedtracker.New(cfg, store)
	feed, err := tracker.AddFeed(ctx, feedURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: add feed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("added feed: %s (%s)\n  url: %s\n", feed.Title, feed.FeedType, feed.FeedURL)
}
