package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/domain"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runFolder(ctx context.Context, cfgPath string, args []string) {
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

	if len(args) == 0 {
		listFolders(ctx, store)
		return
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "create":
		if len(rest) < 1 {
			fmt.Fprintln(os.Stderr, "usage: ft folder create <name>")
			os.Exit(1)
		}
		createFolder(ctx, store, rest[0])
	case "delete":
		if len(rest) < 1 {
			fmt.Fprintln(os.Stderr, "usage: ft folder delete <name>")
			os.Exit(1)
		}
		deleteFolder(ctx, store, rest[0])
	case "rename":
		if len(rest) < 2 {
			fmt.Fprintln(os.Stderr, "usage: ft folder rename <old-name> <new-name>")
			os.Exit(1)
		}
		renameFolder(ctx, store, rest[0], rest[1])
	default:
		fmt.Fprintf(os.Stderr, "unknown folder subcommand: %s\n", sub)
		fmt.Fprintln(os.Stderr, "usage: ft folder [create|delete|rename]")
		os.Exit(1)
	}
}

func listFolders(ctx context.Context, store storage.Storage) {
	folders, err := store.ListFolders(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: list folders: %v\n", err)
		os.Exit(1)
	}

	feeds, err := store.ListFeeds(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: list feeds: %v\n", err)
		os.Exit(1)
	}

	countByFolder := make(map[string]int)
	var ungrouped int
	for _, f := range feeds {
		if f.FolderID == "" {
			ungrouped++
		} else {
			countByFolder[f.FolderID]++
		}
	}

	if len(folders) == 0 {
		fmt.Printf("no folders — %d feeds ungrouped\n", ungrouped)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tFEEDS")
	fmt.Fprintln(w, "----\t-----")
	for _, f := range folders {
		n := countByFolder[f.ID]
		fmt.Fprintf(w, "%s\t%d\n", f.Name, n)
	}
	if ungrouped > 0 {
		fmt.Fprintf(w, "(ungrouped)\t%d\n", ungrouped)
	}
	w.Flush()
}

func createFolder(ctx context.Context, store storage.Storage, name string) {
	now := time.Now()
	f := &domain.Folder{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedAt: now,
	}
	if err := store.AddFolder(ctx, f); err != nil {
		fmt.Fprintf(os.Stderr, "error: create folder: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("created folder: %s\n", name)
}

func deleteFolder(ctx context.Context, store storage.Storage, name string) {
	f, err := store.GetFolderByName(ctx, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: folder %q not found\n", name)
		os.Exit(1)
	}
	// Move feeds in this folder to no folder
	feeds, err := store.ListFeeds(ctx)
	if err == nil {
		for _, feed := range feeds {
			if feed.FolderID == f.ID {
				if err := store.SetFeedFolder(ctx, feed.ID, ""); err != nil {
					fmt.Fprintf(os.Stderr, "warning: could not unset feed %q folder: %v\n", feed.Title, err)
				}
			}
		}
	}
	if err := store.DeleteFolder(ctx, f.ID); err != nil {
		fmt.Fprintf(os.Stderr, "error: delete folder: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("deleted folder: %s\n", name)
}

func renameFolder(ctx context.Context, store storage.Storage, oldName, newName string) {
	f, err := store.GetFolderByName(ctx, oldName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: folder %q not found\n", oldName)
		os.Exit(1)
	}
	// Since storage doesn't have RenameFolder, we delete and recreate
	if err := store.DeleteFolder(ctx, f.ID); err != nil {
		fmt.Fprintf(os.Stderr, "error: rename folder: %v\n", err)
		os.Exit(1)
	}
	nf := &domain.Folder{
		ID:        f.ID,
		Name:      newName,
		CreatedAt: f.CreatedAt,
	}
	if err := store.AddFolder(ctx, nf); err != nil {
		fmt.Fprintf(os.Stderr, "error: rename folder: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("renamed folder %q to %q\n", oldName, newName)
}
