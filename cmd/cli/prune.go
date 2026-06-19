package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runPrune(ctx context.Context, cfgPath string) {
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

	maxAge := time.Duration(cfg.Prune.MaxAge)
	if maxAge <= 0 {
		fmt.Println("Pruning is disabled (prune.max_age is 0)")
		return
	}

	fmt.Printf("Deleting entries older than %s...\n", maxAge)
	n, err := store.DeleteEntriesOlderThan(ctx, maxAge)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: prune: %v\n", err)
		os.Exit(1)
	}
	if n == 0 {
		fmt.Println("No entries to prune")
	} else {
		fmt.Printf("Pruned %d entr%s\n", n, map[bool]string{true: "y", false: "ies"}[n == 1])
	}
}
