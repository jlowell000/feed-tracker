package opml

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExport_FlatFeeds(t *testing.T) {
	specs := []FeedSpec{
		{URL: "https://example.com/one", Title: "Feed One"},
		{URL: "https://example.com/two", Title: "Feed Two"},
	}

	var buf bytes.Buffer
	if err := Export(specs, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `version="2.0"`) {
		t.Error("missing version attribute")
	}
	if !strings.Contains(output, "Feed One") {
		t.Error("missing feed title")
	}
	if !strings.Contains(output, "https://example.com/one") {
		t.Error("missing feed URL")
	}
}

func TestExport_FolderGrouped(t *testing.T) {
	specs := []FeedSpec{
		{URL: "https://a.com/feed", Title: "Ars", Folder: "Tech"},
		{URL: "https://b.com/feed", Title: "HN", Folder: "Tech"},
		{URL: "https://c.com/feed", Title: "Reuters", Folder: "News"},
	}

	var buf bytes.Buffer
	if err := Export(specs, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `<outline text="News">`) {
		t.Error("missing News folder outline")
	}
	if !strings.Contains(output, `<outline text="Tech">`) {
		t.Error("missing Tech folder outline")
	}
	if !strings.Contains(output, `xmlUrl="https://a.com/feed"`) {
		t.Error("missing feed URL")
	}
}

func TestExport_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := Export(nil, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "opml") {
		t.Error("expected valid opml")
	}
	if !strings.Contains(output, "<body></body>") {
		t.Error("expected empty body")
	}
}

func TestExport_ValidXML(t *testing.T) {
	specs := []FeedSpec{
		{URL: "https://example.com/feed", Title: "Test"},
	}

	var buf bytes.Buffer
	Export(specs, &buf)

	specsBack, err := ParseFile(writeTempOPML(t, buf.String()))
	if err != nil {
		t.Fatalf("ParseFile exported output: %v", err)
	}
	if len(specsBack) != 1 {
		t.Fatalf("got %d specs back, want 1", len(specsBack))
	}
	if specsBack[0].URL != "https://example.com/feed" {
		t.Errorf("URL = %q, want %q", specsBack[0].URL, "https://example.com/feed")
	}
	if specsBack[0].Title != "Test" {
		t.Errorf("Title = %q, want %q", specsBack[0].Title, "Test")
	}
}

func TestExport_RoundTripFolder(t *testing.T) {
	specs := []FeedSpec{
		{URL: "https://a.com/rss", Title: "A Feed", Folder: "Tech"},
		{URL: "https://b.com/atom", Title: "B Feed"},
	}

	var buf bytes.Buffer
	Export(specs, &buf)

	specsBack, err := ParseFile(writeTempOPML(t, buf.String()))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(specsBack) != 2 {
		t.Fatalf("got %d specs, want 2", len(specsBack))
	}
	if specsBack[0].Folder != "Tech" {
		t.Errorf("specsBack[0].Folder = %q, want %q", specsBack[0].Folder, "Tech")
	}
	if specsBack[1].Folder != "" {
		t.Errorf("specsBack[1].Folder = %q, want empty", specsBack[1].Folder)
	}
}

func TestExport_NoTitle(t *testing.T) {
	specs := []FeedSpec{
		{URL: "https://example.com/feed"},
	}

	var buf bytes.Buffer
	if err := Export(specs, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `xmlUrl="https://example.com/feed"`) {
		t.Error("missing feed URL")
	}
}

func TestExport_SpecialChars(t *testing.T) {
	specs := []FeedSpec{
		{URL: "https://example.com/feed", Title: `AT&T "News" & <Sports>`, Folder: `Tech & Co`},
	}

	var buf bytes.Buffer
	if err := Export(specs, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "AT&amp;T") {
		t.Error("expected escaped ampersand")
	}
	if !strings.Contains(output, "&lt;") {
		t.Error("expected escaped less-than")
	}

	specsBack, err := ParseFile(writeTempOPML(t, output))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(specsBack) != 1 {
		t.Fatalf("got %d specs, want 1", len(specsBack))
	}
	if specsBack[0].Title != `AT&T "News" & <Sports>` {
		t.Errorf("Title = %q, want %q", specsBack[0].Title, `AT&T "News" & <Sports>`)
	}
	if specsBack[0].Folder != "Tech & Co" {
		t.Errorf("Folder = %q, want %q", specsBack[0].Folder, "Tech & Co")
	}
}

func TestExport_WriteFile(t *testing.T) {
	specs := []FeedSpec{
		{URL: "https://example.com/feed", Title: "Test"},
	}

	var buf bytes.Buffer
	if err := Export(specs, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	path := filepath.Join(t.TempDir(), "export.opml")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d specs, want 1", len(got))
	}
}

func TestExport_FolderOrder(t *testing.T) {
	specs := []FeedSpec{
		{URL: "https://z.com/feed", Title: "Z", Folder: "ZZZ"},
		{URL: "https://a.com/feed", Title: "A", Folder: "AAA"},
	}

	var buf bytes.Buffer
	Export(specs, &buf)

	output := buf.String()
	aaaIdx := strings.Index(output, `text="AAA"`)
	zzzIdx := strings.Index(output, `text="ZZZ"`)
	if aaaIdx < 0 || zzzIdx < 0 {
		t.Fatal("missing folder outlines")
	}
	if aaaIdx > zzzIdx {
		t.Error("AAA should appear before ZZZ (alphabetical)")
	}
}
