package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jlowell000/feed-tracker/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	cfgPath := flag.String("config", "./config.yaml", "path to config file")
	flagidx := 0
	for i, a := range args {
		if a == "--config" || a == "-config" {
			if i+1 < len(args) {
				cfgPath = &args[i+1]
				flagidx = i + 2
				break
			}
		}
		if len(a) > 9 && a[:9] == "--config=" {
			v := a[9:]
			cfgPath = &v
			flagidx = i + 1
			break
		}
	}
	remaining := args
	if flagidx > 0 {
		remaining = args[flagidx:]
	}

	ctx := context.Background()

	switch subcommand {
	case "completion":
		if len(remaining) < 1 {
			fmt.Fprintln(os.Stderr, "usage: ft completion bash|zsh")
			os.Exit(1)
		}
		runCompletion(remaining[0])
		return
	case "help", "-h", "--help":
		printUsage()
		return
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		os.Exit(1)
	}

	timeout := cfg.HTTP.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	switch subcommand {
	case "migrate":
		runMigrate(ctx, *cfgPath)
	case "add":
		runAdd(ctx, *cfgPath, remaining)
	case "fetch":
		runFetch(ctx, *cfgPath, remaining)
	case "feeds":
		runFeeds(ctx, *cfgPath, remaining)
	case "feed":
		runFeed(ctx, *cfgPath, remaining)
	case "folder":
		runFolder(ctx, *cfgPath, remaining)
	case "import":
		runImport(ctx, *cfgPath, remaining)
	case "export":
		runExport(ctx, *cfgPath, remaining)
	case "delete":
		runDelete(ctx, *cfgPath, remaining)
	case "list":
		runList(ctx, *cfgPath, remaining)
	case "search":
		runSearch(ctx, *cfgPath, remaining)
	case "read":
		runRead(ctx, *cfgPath, remaining)
	case "unread":
		runUnread(ctx, *cfgPath, remaining)
	case "open":
		runOpen(ctx, *cfgPath, remaining)
	case "prune":
		runPrune(ctx, *cfgPath)
	case "vacuum":
		runVacuum(ctx, *cfgPath)
	case "db":
		runDB(ctx, *cfgPath, remaining)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, `Usage: ft [--config <path>] <command> [args]

Commands:
  migrate           Create or update the database schema
  add    <url>      Add a new feed by URL
  fetch  [<name> | --feed-id <id>]  Fetch new entries from feed(s)
  feeds             List all tracked feeds
  feed   update <name> [--title <title>] [--url <url>]  Update feed title or URL
  folder [create|delete|rename]  Manage folders
  import [--dry-run] <file.opml>  Import feeds from OPML file
  export [--output <file>]        Export feeds to OPML file
  delete <name | --feed-id <id>>  Delete a feed and all its entries
  list   [<name> | --feed-id <id>] [--limit <n>] [--unread]  List entries
  search [--limit <n>] <query>  Search entries by keyword
  read   [--all | --feed <name> | --feed-id <id> | <entry-id>]  Mark entries as read
  unread <entry-id>  Mark entry as unread
  open   <entry-id>  Open entry URL in system browser
  prune             Delete entries older than the configured prune.max_age
  vacuum            Reclaim database storage space (VACUUM)
  db     optimize   Optimize database performance (PRAGMA optimize)
  completion bash|zsh  Generate shell completion script

Flags:
  --config <path>  Path to config file (default: ./config.yaml)
`)
}
