//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/jlowell000/feed-tracker/internal/domain"
)

func main() {
	entries := []domain.Entry{
		{Title: "Starred Entry", Starred: true, Read: false},
		{Title: "Not Starred", Starred: false, Read: false},
		{Title: "Read Starred", Starred: true, Read: true},
	}

	for _, e := range entries {
		star := "  "
		if e.Starred {
			star = "★ "
		}
		line := fmt.Sprintf("  %s%s", star, e.Title)
		fmt.Printf("Raw line: %q\n", line)
	}
}
ENDOFFEST
go run cmd/ tui/test_star_display.go 2>&1 || true
rm -f cmd/tui/test_star_display.go
