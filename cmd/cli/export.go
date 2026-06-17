package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/opml"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runExport(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	output := fs.String("output", "", "output file (default: feed-tracker-<timestamp>.opml)")
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

	if err := store.Migrate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: migrate: %v\n", err)
		os.Exit(1)
	}

	feeds, err := store.ListFeeds(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: list feeds: %v\n", err)
		os.Exit(1)
	}

	folders, err := store.ListFolders(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: list folders: %v\n", err)
		os.Exit(1)
	}

	folderNames := make(map[string]string)
	for _, f := range folders {
		folderNames[f.ID] = f.Name
	}

	var specs []opml.FeedSpec
	for _, feed := range feeds {
		s := opml.FeedSpec{
			URL:   feed.FeedURL,
			Title: feed.Title,
		}
		if feed.FolderID != "" {
			s.Folder = folderNames[feed.FolderID]
		}
		specs = append(specs, s)
	}

	sort.Slice(specs, func(i, j int) bool {
		if specs[i].Folder != specs[j].Folder {
			return specs[i].Folder < specs[j].Folder
		}
		return specs[i].Title < specs[j].Title
	})

	outPath := *output
	if outPath == "" {
		outPath = fmt.Sprintf("feed-tracker-%s.opml", time.Now().Format("2006-01-02-150405"))
	}

	f, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: create file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := opml.Export(specs, f); err != nil {
		fmt.Fprintf(os.Stderr, "error: export opml: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("exported %s to %s\n", plural(len(specs), "feed"), outPath)
}
