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

func runSearch(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	limit := fs.Int("limit", 20, "max entries to show")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: ft search [--limit <n>] <query>")
		os.Exit(1)
	}

	query := fs.Arg(0)

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

	entries, err := store.SearchEntries(ctx, query, *limit, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: search entries: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("no matching entries found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PUBLISHED\tFEED\tTITLE\tURL\tREAD")
	fmt.Fprintln(w, "---------\t----\t-----\t---\t----")
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
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", pub, e.FeedTitle, title, e.URL, read)
	}
	w.Flush()
}
