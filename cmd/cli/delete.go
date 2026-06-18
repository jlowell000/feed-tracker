package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runDelete(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	feedID := fs.String("feed-id", "", "delete a feed by ID")
	fs.Parse(args)

	feedName := ""
	if fs.Arg(0) != "" {
		feedName = fs.Arg(0)
	}

	if *feedID == "" && feedName == "" {
		fmt.Fprintln(os.Stderr, "usage: ft delete <name | --feed-id <id>>")
		os.Exit(1)
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

	resolvedID := *feedID
	if feedName != "" {
		feed, err := store.GetFeedByTitle(ctx, feedName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: feed %q not found\n", feedName)
			os.Exit(1)
		}
		resolvedID = feed.ID
	}

	feed, err := store.GetFeed(ctx, resolvedID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: get feed: %v\n", err)
		os.Exit(1)
	}

	if err := store.DeleteFeed(ctx, resolvedID); err != nil {
		fmt.Fprintf(os.Stderr, "error: delete feed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("deleted feed: %s (%s)\n", feed.Title, feed.FeedURL)
}
