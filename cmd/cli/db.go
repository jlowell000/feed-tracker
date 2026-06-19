package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runDB(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("db", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: ft db <subcommand>\n\nSubcommands:\n  optimize   Run PRAGMA optimize to maintain database performance")
		os.Exit(1)
	}

	switch fs.Arg(0) {
	case "optimize":
		runDBOptimize(ctx, cfgPath)
	default:
		fmt.Fprintf(os.Stderr, "unknown db subcommand: %s\n", fs.Arg(0))
		os.Exit(1)
	}
}

func runDBOptimize(ctx context.Context, cfgPath string) {
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

	if err := store.Optimize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: optimize: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Optimize complete")
}
