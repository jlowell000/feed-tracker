package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runStar(ctx context.Context, cfgPath string, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ft star <entry-id>")
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

	if err := store.StarEntry(ctx, args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Entry starred")
}
