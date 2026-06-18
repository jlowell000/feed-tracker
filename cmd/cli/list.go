package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/domain"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runList(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	feedID := fs.String("feed-id", "", "feed ID to list entries for")
	limit := fs.Int("limit", 20, "max entries to show")
	unreadOnly := fs.Bool("unread", false, "show only unread entries")
	detail := fs.Bool("detail", false, "show detailed entry view")
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

	resolvedID := *feedID
	if feedName != "" {
		feed, err := store.GetFeedByTitle(ctx, feedName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: feed %q not found\n", feedName)
			os.Exit(1)
		}
		resolvedID = feed.ID
	}

	var entries []*domain.Entry
	if *unreadOnly {
		entries, err = store.ListEntriesUnread(ctx, resolvedID, *limit)
	} else {
		entries, err = store.ListEntries(ctx, resolvedID, *limit)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: list entries: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("no entries found")
		return
	}

	if *detail {
		printEntriesDetail(entries, resolvedID)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	showFeed := resolvedID == ""
	if showFeed {
		fmt.Fprintln(w, "PUBLISHED\tFEED\tTITLE\tURL\tREAD")
		fmt.Fprintln(w, "---------\t----\t-----\t---\t----")
	} else {
		fmt.Fprintln(w, "PUBLISHED\tTITLE\tURL\tREAD")
		fmt.Fprintln(w, "---------\t-----\t---\t----")
	}
	for _, e := range entries {
		pub := e.PublishedAt.Format("2006-01-02 15:04")
		title := e.Title
		if title == "" {
			title = "(no title)"
		}
		read := "no"
		if e.Read {
			read = "yes"
		}
		if showFeed {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", pub, e.FeedTitle, title, e.URL, read)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", pub, title, e.URL, read)
		}
	}
	w.Flush()
}

func printEntriesDetail(entries []*domain.Entry, resolvedID string) {
	for i, e := range entries {
		if i > 0 {
			fmt.Println("---")
		}
		pub := "(no date)"
		if !e.PublishedAt.IsZero() {
			pub = e.PublishedAt.Format("2006-01-02 15:04")
		}
		title := e.Title
		if title == "" {
			title = "(no title)"
		}
		read := "no"
		if e.Read {
			read = "yes"
		}
		fmt.Printf("Title:   %s\n", title)
		fmt.Printf("Published: %s\n", pub)
		if e.Author != "" {
			fmt.Printf("Author:  %s\n", e.Author)
		}
		if resolvedID == "" {
			fmt.Printf("Feed:    %s\n", e.FeedTitle)
		}
		fmt.Printf("URL:     %s\n", e.URL)
		fmt.Printf("Read:    %s\n", read)
		body := e.Content
		if body == "" {
			body = e.Summary
		}
		if body != "" {
			if len(body) > 200 {
				body = body[:200] + "..."
			}
			fmt.Printf("Content: %s\n", body)
		}
	}
	fmt.Println("---")
}
