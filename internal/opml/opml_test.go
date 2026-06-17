package opml

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempOPML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "feeds.opml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp opml: %v", err)
	}
	return path
}

func TestParseFile_BasicFeeds(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>My Feeds</title></head>
  <body>
    <outline type="rss" text="Feed One" xmlUrl="https://example.com/one"/>
    <outline type="rss" text="Feed Two" xmlUrl="https://example.com/two"/>
  </body>
</opml>`

	specs, err := ParseFile(writeTempOPML(t, xml))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("got %d specs, want 2", len(specs))
	}
	if specs[0].URL != "https://example.com/one" {
		t.Errorf("specs[0].URL = %q, want %q", specs[0].URL, "https://example.com/one")
	}
	if specs[0].Title != "Feed One" {
		t.Errorf("specs[0].Title = %q, want %q", specs[0].Title, "Feed One")
	}
}

func TestParseFile_FolderHierarchy(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline text="Tech">
      <outline type="rss" text="Ars" xmlUrl="https://arstechnica.com/feed/"/>
      <outline type="rss" text="HN" xmlUrl="https://hnrss.org/frontpage"/>
    </outline>
    <outline text="News">
      <outline type="rss" text="Reuters" xmlUrl="https://reuters.com/feed/"/>
    </outline>
  </body>
</opml>`

	specs, err := ParseFile(writeTempOPML(t, xml))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(specs) != 3 {
		t.Fatalf("got %d specs, want 3", len(specs))
	}

	expected := []struct {
		title  string
		folder string
	}{
		{"Ars", "Tech"},
		{"HN", "Tech"},
		{"Reuters", "News"},
	}
	for i, e := range expected {
		if specs[i].Title != e.title {
			t.Errorf("specs[%d].Title = %q, want %q", i, specs[i].Title, e.title)
		}
		if specs[i].Folder != e.folder {
			t.Errorf("specs[%d].Folder = %q, want %q", i, specs[i].Folder, e.folder)
		}
	}
}

func TestParseFile_MixedFolderAndFlat(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline text="Tech">
      <outline type="rss" text="Ars" xmlUrl="https://arstechnica.com/feed/"/>
    </outline>
    <outline type="rss" text="Flat Blog" xmlUrl="https://example.com/blog"/>
  </body>
</opml>`

	specs, err := ParseFile(writeTempOPML(t, xml))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("got %d specs, want 2", len(specs))
	}
	if specs[0].Folder != "Tech" {
		t.Errorf("specs[0].Folder = %q, want %q", specs[0].Folder, "Tech")
	}
	if specs[1].Folder != "" {
		t.Errorf("specs[1].Folder = %q, want empty", specs[1].Folder)
	}
}

func TestParseFile_NestedFolders(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline text="Top">
      <outline text="Sub">
        <outline type="rss" text="Deep Feed" xmlUrl="https://example.com/deep"/>
      </outline>
    </outline>
  </body>
</opml>`

	specs, err := ParseFile(writeTempOPML(t, xml))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("got %d specs, want 1", len(specs))
	}
	if specs[0].Folder != "Sub" {
		t.Errorf("specs[0].Folder = %q, want %q", specs[0].Folder, "Sub")
	}
}

func TestParseFile_NoFeeds(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Empty</title></head>
  <body></body>
</opml>`

	specs, err := ParseFile(writeTempOPML(t, xml))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(specs) != 0 {
		t.Errorf("got %d specs, want 0", len(specs))
	}
}

func TestParseFile_NoOutline(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>No Outline</title></head>
  <body></body>
</opml>`

	specs, err := ParseFile(writeTempOPML(t, xml))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(specs) != 0 {
		t.Errorf("got %d specs, want 0", len(specs))
	}
}

func TestParseFile_UsesTitleAttr(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline type="rss" text="Fallback" title="Real Title" xmlUrl="https://example.com/feed"/>
  </body>
</opml>`

	specs, err := ParseFile(writeTempOPML(t, xml))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("got %d specs, want 1", len(specs))
	}
	if specs[0].Title != "Real Title" {
		t.Errorf("specs[0].Title = %q, want %q", specs[0].Title, "Real Title")
	}
}

func TestParseFile_MissingFile(t *testing.T) {
	_, err := ParseFile("/nonexistent/opml.opml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseFile_InvalidXML(t *testing.T) {
	_, err := ParseFile(writeTempOPML(t, "not xml"))
	if err == nil {
		t.Fatal("expected error for invalid xml")
	}
}

func TestParseFile_OutlineWithTitle(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline title="Folder Title">
      <outline type="rss" text="Feed" xmlUrl="https://example.com/feed"/>
    </outline>
  </body>
</opml>`

	specs, err := ParseFile(writeTempOPML(t, xml))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("got %d specs, want 1", len(specs))
	}
	if specs[0].Folder != "Folder Title" {
		t.Errorf("specs[0].Folder = %q, want %q", specs[0].Folder, "Folder Title")
	}
}
