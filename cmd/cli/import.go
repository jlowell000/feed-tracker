package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/domain"
	"github.com/jlowell000/feed-tracker/internal/feedtracker"
	"github.com/jlowell000/feed-tracker/internal/opml"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func runImport(ctx context.Context, cfgPath string, args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "preview feeds without importing")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: ft import [--dry-run] <file.opml>")
		os.Exit(1)
	}

	path := fs.Arg(0)
	specs, err := opml.ParseFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: parse opml: %v\n", err)
		os.Exit(1)
	}

	if len(specs) == 0 {
		fmt.Println("no feeds found in file")
		return
	}

	if *dryRun {
		printImportDryRun(specs)
		return
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

	if err := store.Migrate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: migrate: %v\n", err)
		os.Exit(1)
	}

	tracker := feedtracker.New(cfg, store)

	imported := 0
	skipped := 0
	errs := 0

	for _, spec := range specs {
		feed, addErr := tracker.AddFeed(ctx, spec.URL)
		if addErr != nil {
			errs++
			fmt.Fprintf(os.Stderr, "  error: %s — %v\n", spec.URL, addErr)
			continue
		}
		imported++

		if spec.Folder != "" {
			folder, fErr := store.GetFolderByName(ctx, spec.Folder)
			if fErr != nil {
				folder = &domain.Folder{
					ID:        uuid.New().String(),
					Name:      spec.Folder,
					CreatedAt: time.Now(),
				}
				if aErr := store.AddFolder(ctx, folder); aErr != nil {
					fmt.Fprintf(os.Stderr, "  warning: create folder %q: %v\n", spec.Folder, aErr)
					continue
				}
			}
			if sErr := store.SetFeedFolder(ctx, feed.ID, folder.ID); sErr != nil {
				fmt.Fprintf(os.Stderr, "  warning: assign folder %q: %v\n", spec.Folder, sErr)
			}
		}
	}

	parts := []string{plural(imported, "feed")}
	if skipped > 0 {
		parts = append(parts, plural(skipped, "skipped"))
	}
	if errs > 0 {
		parts = append(parts, plural(errs, "error"))
	}
	fmt.Printf("imported %s\n", joinParts(parts))
}

func printImportDryRun(specs []opml.FeedSpec) {
	fmt.Printf("Dry run: %d feeds found\n\n", len(specs))

	byFolder := make(map[string][]opml.FeedSpec)
	var noFolder []opml.FeedSpec
	for _, s := range specs {
		if s.Folder == "" {
			noFolder = append(noFolder, s)
		} else {
			byFolder[s.Folder] = append(byFolder[s.Folder], s)
		}
	}

	folderNames := make([]string, 0, len(byFolder))
	for name := range byFolder {
		folderNames = append(folderNames, name)
	}
	sort.Strings(folderNames)

	for _, name := range folderNames {
		feeds := byFolder[name]
		fmt.Printf("  %s (%s)\n", name, plural(len(feeds), "feed"))
		for _, f := range feeds {
			title := f.Title
			if title == "" {
				title = "(no title)"
			}
			fmt.Printf("    %-40s %s\n", title, f.URL)
		}
		fmt.Println()
	}

	if len(noFolder) > 0 {
		fmt.Printf("  Uncategorized (%s)\n", plural(len(noFolder), "feed"))
		for _, f := range noFolder {
			title := f.Title
			if title == "" {
				title = "(no title)"
			}
			fmt.Printf("    %-40s %s\n", title, f.URL)
		}
		fmt.Println()
	}
}

func plural(n int, s string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, s)
	}
	return fmt.Sprintf("%d %ss", n, s)
}

func joinParts(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	case 2:
		return parts[0] + ", " + parts[1]
	default:
		result := ""
		for i, p := range parts {
			if i > 0 {
				result += ", "
			}
			result += p
		}
		return result
	}
}
