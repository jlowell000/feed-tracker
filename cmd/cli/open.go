package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runOpen(ctx context.Context, cfgPath string, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ft open <entry-id>")
		os.Exit(1)
	}

	entryID := args[0]

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

	entry, err := store.GetEntry(ctx, entryID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: entry %q not found\n", entryID)
		os.Exit(1)
	}

	if entry.URL == "" {
		fmt.Fprintf(os.Stderr, "error: entry %q has no URL\n", entryID)
		os.Exit(1)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", entry.URL)
	case "linux":
		cmd = exec.Command("xdg-open", entry.URL)
	default:
		fmt.Fprintf(os.Stderr, "error: unsupported platform: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error: open browser: %v\n", err)
		os.Exit(1)
	}
}
