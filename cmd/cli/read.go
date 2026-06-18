package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runRead(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("read", flag.ExitOnError)
	all := fs.Bool("all", false, "mark all entries as read")
	feedName := fs.String("feed", "", "mark all entries in a feed as read (by name)")
	feedID := fs.String("feed-id", "", "mark all entries in a feed as read (by ID)")
	fs.Parse(args)

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

	switch {
	case *all:
		if err := store.MarkAllRead(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "error: mark all read: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("marked all entries as read")

	case *feedName != "":
		feed, err := store.GetFeedByTitle(ctx, *feedName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: feed %q not found\n", *feedName)
			os.Exit(1)
		}
		if err := store.MarkFeedRead(ctx, feed.ID); err != nil {
			fmt.Fprintf(os.Stderr, "error: mark feed read: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("marked all entries in %q as read\n", feed.Title)

	case *feedID != "":
		if err := store.MarkFeedRead(ctx, *feedID); err != nil {
			fmt.Fprintf(os.Stderr, "error: mark feed read: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("marked all entries in feed %s as read\n", *feedID)

	default:
		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "usage: ft read [--all | --feed <name> | --feed-id <id> | <entry-id>]")
			os.Exit(1)
		}
		entryID := fs.Arg(0)
		if err := store.MarkEntryRead(ctx, entryID); err != nil {
			fmt.Fprintf(os.Stderr, "error: mark read: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("marked entry %s as read\n", entryID)
	}
}

func runUnread(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("unread", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: ft unread <entry-id>")
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

	entryID := fs.Arg(0)
	if err := store.MarkEntryUnread(ctx, entryID); err != nil {
		fmt.Fprintf(os.Stderr, "error: mark unread: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("marked entry %s as unread\n", entryID)
}
