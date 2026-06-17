package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runFeeds(ctx context.Context, cfgPath string, args []string) {
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tTYPE\tURL\tLAST FETCHED")
	fmt.Fprintln(w, "--\t-----\t----\t---\t------------")
	for _, f := range feeds {
		lastFetched := "never"
		if !f.LastFetched.IsZero() {
			lastFetched = f.LastFetched.Format("2006-01-02 15:04")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", f.ID, f.Title, f.FeedType, f.FeedURL, lastFetched)
	}
	w.Flush()
}
