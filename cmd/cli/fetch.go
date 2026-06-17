package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/feedtracker"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runFetch(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	feedID := fs.String("feed-id", "", "fetch a specific feed by ID")
	fs.Parse(args)

	feedName := ""
	if fs.Arg(0) != "" {
		feedName = fs.Arg(0)
	}

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

	resolvedID := *feedID
	if feedName != "" {
		feed, err := store.GetFeedByTitle(ctx, feedName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: feed %q not found\n", feedName)
			os.Exit(1)
		}
		resolvedID = feed.ID
	}

	if resolvedID != "" {
		feed, err := store.GetFeed(ctx, resolvedID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: get feed: %v\n", err)
			os.Exit(1)
		}
		n, err := tracker.FetchFeed(ctx, feed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: fetch feed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("fetched %q — %d new entries\n", feed.Title, n)
	} else {
		n, err := tracker.FetchAllFeeds(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: fetch feeds: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("fetched all feeds — %d new entries\n", n)
	}
}
