package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runList(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	feedID := fs.String("feed-id", "", "feed ID to list entries for")
	limit := fs.Int("limit", 20, "max entries to show")
	fs.Parse(args)

	if *feedID == "" {
		fmt.Fprintln(os.Stderr, "error: --feed-id is required")
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

	entries, err := store.ListEntries(ctx, *feedID, *limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: list entries: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("no entries found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PUBLISHED\tTITLE\tURL")
	fmt.Fprintln(w, "---------\t-----\t---")
	for _, e := range entries {
		pub := e.PublishedAt.Format("2006-01-02 15:04")
		title := e.Title
		if title == "" {
			title = "(no title)"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", pub, title, e.URL)
	}
	w.Flush()
}
