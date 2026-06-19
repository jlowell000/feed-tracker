package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runFeed(ctx context.Context, cfgPath string, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ft feed update <name> [--title <title>] [--url <url>]")
		os.Exit(1)
	}

	switch args[0] {
	case "update":
		runFeedUpdate(ctx, cfgPath, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown feed subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func runFeedUpdate(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("feed update", flag.ExitOnError)
	title := fs.String("title", "", "new feed title")
	url := fs.String("url", "", "new feed URL")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: ft feed update <name> [--title <title>] [--url <url>]")
		os.Exit(1)
	}

	name := fs.Arg(0)

	if *title == "" && *url == "" {
		fmt.Fprintln(os.Stderr, "error: at least one of --title or --url is required")
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

	feed, err := store.GetFeedByTitle(ctx, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: feed %q not found\n", name)
		os.Exit(1)
	}

	changed := false
	if *title != "" {
		feed.Title = *title
		changed = true
	}
	if *url != "" {
		feed.FeedURL = *url
		changed = true
	}

	if !changed {
		fmt.Println("No changes made")
		return
	}

	if err := store.UpdateFeed(ctx, feed); err != nil {
		fmt.Fprintf(os.Stderr, "error: update feed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("updated %q\n", name)
}
