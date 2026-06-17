package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/feedtracker"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func main() {
	cfgPath := flag.String("config", "./config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
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

	if err := store.Migrate(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: migrate database: %v\n", err)
		os.Exit(1)
	}

	tracker := feedtracker.New(cfg, store)

	p := tea.NewProgram(initialModel(cfg, store, tracker), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
