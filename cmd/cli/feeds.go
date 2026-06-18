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

func runFeeds(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("feeds", flag.ExitOnError)
	namesOnly := fs.Bool("names", false, "list feed names only (one per line)")
	folders := fs.Bool("folders", false, "group feeds by folder")
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

	feeds, err := store.ListFeeds(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: list feeds: %v\n", err)
		os.Exit(1)
	}

	if len(feeds) == 0 {
		fmt.Println("no feeds tracked yet")
		return
	}

	unreadCounts, err := store.UnreadCountByFeed(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: get unread counts: %v\n", err)
		os.Exit(1)
	}

	if *namesOnly {
		for _, f := range feeds {
			fmt.Println(f.Title)
		}
		return
	}

	if *folders {
		printFeedsByFolder(ctx, store, feeds, unreadCounts)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tTYPE\tURL\tUNREAD\tLAST FETCHED")
	fmt.Fprintln(w, "--\t-----\t----\t---\t------\t------------")
	for _, f := range feeds {
		lastFetched := "never"
		if !f.LastFetched.IsZero() {
			lastFetched = f.LastFetched.Format("2006-01-02 15:04")
		}
		n := unreadCounts[f.ID]
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n", f.ID, f.Title, f.FeedType, f.FeedURL, n, lastFetched)
	}
	w.Flush()
}

func printFeedsByFolder(ctx context.Context, store storage.Storage, feeds []*domain.Feed, unreadCounts map[string]int) {
	folders, err := store.ListFolders(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: list folders: %v\n", err)
		os.Exit(1)
	}

	byFolder := make(map[string][]*domain.Feed)
	var ungrouped []*domain.Feed
	for _, f := range feeds {
		if f.FolderID == "" {
			ungrouped = append(ungrouped, f)
		} else {
			byFolder[f.FolderID] = append(byFolder[f.FolderID], f)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	for _, folder := range folders {
		totalUnread := 0
		for _, f := range byFolder[folder.ID] {
			totalUnread += unreadCounts[f.ID]
		}
		fmt.Fprintf(w, "%s\t\t(%d unread, %d feeds)\n", folder.Name, totalUnread, len(byFolder[folder.ID]))
		for _, f := range byFolder[folder.ID] {
			n := unreadCounts[f.ID]
			fmt.Fprintf(w, "  %s\t%s\t%d\t%s\n", f.Title, f.FeedType, n, f.FeedURL)
		}
	}

	if len(ungrouped) > 0 {
		totalUnread := 0
		for _, f := range ungrouped {
			totalUnread += unreadCounts[f.ID]
		}
		fmt.Fprintf(w, "(ungrouped)\t\t(%d unread, %d feeds)\n", totalUnread, len(ungrouped))
		for _, f := range ungrouped {
			n := unreadCounts[f.ID]
			fmt.Fprintf(w, "  %s\t%s\t%d\t%s\n", f.Title, f.FeedType, n, f.FeedURL)
		}
	}

	w.Flush()
}
