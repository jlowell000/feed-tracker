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
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: ft read <entry-id>")
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
	if err := store.MarkEntryRead(ctx, entryID); err != nil {
		fmt.Fprintf(os.Stderr, "error: mark read: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("marked entry %s as read\n", entryID)
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
