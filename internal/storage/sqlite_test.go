package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jlowell000/feed-tracker/internal/domain"
)

func newTestStore(t *testing.T) Storage {
	t.Helper()
	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return s
}

func TestAddAndGetFeed(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed := &domain.Feed{
		ID:      uuid.New().String(),
		Title:   "Test Feed",
		FeedURL: "https://example.com/feed",
		FeedType: domain.FeedTypeRSS,
	}
	if err := s.AddFeed(ctx, feed); err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	got, err := s.GetFeed(ctx, feed.ID)
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if got.Title != "Test Feed" {
		t.Errorf("Title = %q, want %q", got.Title, "Test Feed")
	}
}

func TestGetFeedByURL(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed := &domain.Feed{
		ID:      uuid.New().String(),
		Title:   "Test",
		FeedURL: "https://example.com/feed",
		FeedType: domain.FeedTypeRSS,
	}
	s.AddFeed(ctx, feed)

	got, err := s.GetFeedByURL(ctx, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GetFeedByURL: %v", err)
	}
	if got.Title != "Test" {
		t.Errorf("Title = %q, want %q", got.Title, "Test")
	}
}

func TestListFeeds(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	s.AddFeed(ctx, &domain.Feed{ID: uuid.New().String(), Title: "A", FeedURL: "https://a.com/feed", FeedType: domain.FeedTypeRSS})
	s.AddFeed(ctx, &domain.Feed{ID: uuid.New().String(), Title: "B", FeedURL: "https://b.com/feed", FeedType: domain.FeedTypeAtom})

	feeds, err := s.ListFeeds(ctx)
	if err != nil {
		t.Fatalf("ListFeeds: %v", err)
	}
	if len(feeds) != 2 {
		t.Errorf("len(feeds) = %d, want 2", len(feeds))
	}
}

func TestUpsertEntry(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	entry := &domain.Entry{
		ID:          uuid.New().String(),
		FeedID:      feedID,
		ExternalID:  "ext-1",
		Title:       "Entry 1",
		FetchedAt:   time.Now(),
	}

	isNew, err := s.UpsertEntry(ctx, entry)
	if err != nil {
		t.Fatalf("UpsertEntry: %v", err)
	}
	if !isNew {
		t.Error("expected new entry")
	}

	// Same external_id should not insert
	isNew, err = s.UpsertEntry(ctx, entry)
	if err != nil {
		t.Fatalf("second UpsertEntry: %v", err)
	}
	if isNew {
		t.Error("expected duplicate to not be new")
	}
}

func TestListEntries(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e1",
		Title: "Latest", PublishedAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e2",
		Title: "Older", PublishedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})

	entries, err := s.ListEntries(ctx, feedID, 10, 0)
	if err != nil {
		t.Fatalf("ListEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if entries[0].Title != "Latest" {
		t.Errorf("first entry.Title = %q, want %q", entries[0].Title, "Latest")
	}
}

func TestAddAndListFolders(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	f1 := &domain.Folder{ID: uuid.New().String(), Name: "Tech"}
	f2 := &domain.Folder{ID: uuid.New().String(), Name: "News"}
	if err := s.AddFolder(ctx, f1); err != nil {
		t.Fatalf("AddFolder: %v", err)
	}
	if err := s.AddFolder(ctx, f2); err != nil {
		t.Fatalf("AddFolder: %v", err)
	}

	folders, err := s.ListFolders(ctx)
	if err != nil {
		t.Fatalf("ListFolders: %v", err)
	}
	if len(folders) != 2 {
		t.Fatalf("len(folders) = %d, want 2", len(folders))
	}
}

func TestGetFolderByName(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	f := &domain.Folder{ID: uuid.New().String(), Name: "Tech"}
	s.AddFolder(ctx, f)

	got, err := s.GetFolderByName(ctx, "Tech")
	if err != nil {
		t.Fatalf("GetFolderByName: %v", err)
	}
	if got.Name != "Tech" {
		t.Errorf("Name = %q, want %q", got.Name, "Tech")
	}
}

func TestGetFolderByName_Missing(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	_, err := s.GetFolderByName(ctx, "Nonexistent")
	if err == nil {
		t.Fatal("expected error for missing folder")
	}
}

func TestDeleteFolder(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	f := &domain.Folder{ID: uuid.New().String(), Name: "Tech"}
	s.AddFolder(ctx, f)

	if err := s.DeleteFolder(ctx, f.ID); err != nil {
		t.Fatalf("DeleteFolder: %v", err)
	}

	folders, err := s.ListFolders(ctx)
	if err != nil {
		t.Fatalf("ListFolders: %v", err)
	}
	if len(folders) != 0 {
		t.Errorf("len(folders) = %d, want 0", len(folders))
	}
}

func TestSetFeedFolder(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	folder := &domain.Folder{ID: uuid.New().String(), Name: "Tech"}
	s.AddFolder(ctx, folder)

	feed := &domain.Feed{
		ID:       uuid.New().String(),
		Title:    "Test",
		FeedURL:  "https://example.com/feed",
		FeedType: domain.FeedTypeRSS,
	}
	s.AddFeed(ctx, feed)

	if err := s.SetFeedFolder(ctx, feed.ID, folder.ID); err != nil {
		t.Fatalf("SetFeedFolder: %v", err)
	}

	got, err := s.GetFeed(ctx, feed.ID)
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if got.FolderID != folder.ID {
		t.Errorf("FolderID = %q, want %q", got.FolderID, folder.ID)
	}
}

func TestSetFeedFolder_ClearsFolder(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	folder := &domain.Folder{ID: uuid.New().String(), Name: "Tech"}
	s.AddFolder(ctx, folder)

	feed := &domain.Feed{
		ID:       uuid.New().String(),
		Title:    "Test",
		FeedURL:  "https://example.com/feed",
		FeedType: domain.FeedTypeRSS,
	}
	s.AddFeed(ctx, feed)
	s.SetFeedFolder(ctx, feed.ID, folder.ID)

	if err := s.SetFeedFolder(ctx, feed.ID, ""); err != nil {
		t.Fatalf("SetFeedFolder(empty): %v", err)
	}

	got, err := s.GetFeed(ctx, feed.ID)
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if got.FolderID != "" {
		t.Errorf("FolderID = %q, want empty", got.FolderID)
	}
}

func TestListFeedsWithFolderID(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	folder := &domain.Folder{ID: uuid.New().String(), Name: "Tech"}
	s.AddFolder(ctx, folder)

	feed := &domain.Feed{
		ID:       uuid.New().String(),
		Title:    "Test",
		FeedURL:  "https://example.com/feed",
		FeedType: domain.FeedTypeRSS,
	}
	s.AddFeed(ctx, feed)
	s.SetFeedFolder(ctx, feed.ID, folder.ID)

	feeds, err := s.ListFeeds(ctx)
	if err != nil {
		t.Fatalf("ListFeeds: %v", err)
	}
	if len(feeds) != 1 {
		t.Fatalf("len(feeds) = %d, want 1", len(feeds))
	}
	if feeds[0].FolderID != folder.ID {
		t.Errorf("FolderID = %q, want %q", feeds[0].FolderID, folder.ID)
	}
}

func TestAddFolder_DuplicateName(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	f1 := &domain.Folder{ID: uuid.New().String(), Name: "Tech"}
	s.AddFolder(ctx, f1)

	f2 := &domain.Folder{ID: uuid.New().String(), Name: "Tech"}
	if err := s.AddFolder(ctx, f2); err == nil {
		t.Fatal("expected error for duplicate folder name")
	}
}

func TestDeleteFolder_FeedFolderBecomesNull(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	folder := &domain.Folder{ID: uuid.New().String(), Name: "Tech"}
	s.AddFolder(ctx, folder)

	feed := &domain.Feed{
		ID:       uuid.New().String(),
		Title:    "Test",
		FeedURL:  "https://example.com/feed",
		FeedType: domain.FeedTypeRSS,
	}
	s.AddFeed(ctx, feed)
	s.SetFeedFolder(ctx, feed.ID, folder.ID)

	if err := s.DeleteFolder(ctx, folder.ID); err != nil {
		t.Fatalf("DeleteFolder: %v", err)
	}

	got, err := s.GetFeed(ctx, feed.ID)
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	// Foreign keys are enforced with ON DELETE SET NULL
	if got.FolderID != "" {
		t.Error("expected folder_id to be cleared after folder delete")
	}
}

func TestUnreadCountByFeed(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	entry := &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e1",
		Title: "Entry", FetchedAt: time.Now(),
	}
	s.UpsertEntry(ctx, entry)

	counts, err := s.UnreadCountByFeed(ctx)
	if err != nil {
		t.Fatalf("UnreadCountByFeed: %v", err)
	}
	if counts[feedID] != 1 {
		t.Errorf("counts[%s] = %d, want 1", feedID, counts[feedID])
	}

	s.MarkEntryRead(ctx, entry.ID)
	counts, _ = s.UnreadCountByFeed(ctx)
	if counts[feedID] != 0 {
		t.Errorf("after mark read, counts[%s] = %d, want 0", feedID, counts[feedID])
	}
}

func TestMarkEntryReadUnread(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	entryID := uuid.New().String()
	s.UpsertEntry(ctx, &domain.Entry{
		ID: entryID, FeedID: feedID, ExternalID: "e1",
		Title: "Entry", FetchedAt: time.Now(),
	})

	if err := s.MarkEntryRead(ctx, entryID); err != nil {
		t.Fatalf("MarkEntryRead: %v", err)
	}

	entries, _ := s.ListEntriesUnread(ctx, feedID, 10, 0)
	if len(entries) != 0 {
		t.Error("expected no unread entries after mark read")
	}

	if err := s.MarkEntryUnread(ctx, entryID); err != nil {
		t.Fatalf("MarkEntryUnread: %v", err)
	}

	entries, _ = s.ListEntriesUnread(ctx, feedID, 10, 0)
	if len(entries) != 1 {
		t.Error("expected 1 unread entry after mark unread")
	}
}

func TestListEntriesUnread(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	e1ID := uuid.New().String()
	e2ID := uuid.New().String()

	s.UpsertEntry(ctx, &domain.Entry{
		ID: e1ID, FeedID: feedID, ExternalID: "e1",
		Title: "Unread", FetchedAt: time.Now(),
	})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: e2ID, FeedID: feedID, ExternalID: "e2",
		Title: "Read", FetchedAt: time.Now(),
	})

	s.MarkEntryRead(ctx, e2ID)

	unread, err := s.ListEntriesUnread(ctx, feedID, 10, 0)
	if err != nil {
		t.Fatalf("ListEntriesUnread: %v", err)
	}
	if len(unread) != 1 {
		t.Errorf("len(unread) = %d, want 1", len(unread))
	}
	if unread[0].Title != "Unread" {
		t.Errorf("unread[0].Title = %q, want %q", unread[0].Title, "Unread")
	}
}

func TestListEntriesAllFeeds(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed1 := &domain.Feed{ID: uuid.New().String(), Title: "Feed A", FeedURL: "https://a.com/feed", FeedType: domain.FeedTypeRSS}
	feed2 := &domain.Feed{ID: uuid.New().String(), Title: "Feed B", FeedURL: "https://b.com/feed", FeedType: domain.FeedTypeAtom}
	s.AddFeed(ctx, feed1)
	s.AddFeed(ctx, feed2)

	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feed1.ID, ExternalID: "a1",
		Title: "Entry A1", PublishedAt: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feed2.ID, ExternalID: "b1",
		Title: "Entry B1", PublishedAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feed1.ID, ExternalID: "a2",
		Title: "Entry A2", PublishedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})

	entries, err := s.ListEntries(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("ListEntries all feeds: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("len(entries) = %d, want 3", len(entries))
	}
	// Ordered by published_at DESC across all feeds
	if entries[0].Title != "Entry A1" || entries[1].Title != "Entry B1" || entries[2].Title != "Entry A2" {
		t.Error("entries not in expected order (A1, B1, A2)")
	}
	if entries[0].FeedTitle != "Feed A" {
		t.Errorf("entries[0].FeedTitle = %q, want %q", entries[0].FeedTitle, "Feed A")
	}
}

func TestListEntriesZeroLimit(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed := &domain.Feed{ID: uuid.New().String(), Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS}
	s.AddFeed(ctx, feed)

	for i := range 60 {
		s.UpsertEntry(ctx, &domain.Entry{
			ID: uuid.New().String(), FeedID: feed.ID, ExternalID: fmt.Sprintf("e%d", i),
			Title: fmt.Sprintf("Entry %d", i), FetchedAt: time.Now(),
		})
	}

	entries, err := s.ListEntries(ctx, feed.ID, 0, 0)
	if err != nil {
		t.Fatalf("ListEntries zero limit: %v", err)
	}
	if len(entries) != 50 {
		t.Errorf("len(entries) = %d, want 50 (default limit)", len(entries))
	}
}

func TestGetEntry(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	entryID := uuid.New().String()
	entry := &domain.Entry{
		ID: entryID, FeedID: feedID, ExternalID: "ext-1",
		Title: "Test Entry", URL: "https://example.com/entry",
		Summary: "Summary text", Content: "Full content",
		Author: "Author", PublishedAt: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		FetchedAt: time.Now(), Read: false,
	}
	if _, err := s.UpsertEntry(ctx, entry); err != nil {
		t.Fatalf("UpsertEntry: %v", err)
	}

	got, err := s.GetEntry(ctx, entryID)
	if err != nil {
		t.Fatalf("GetEntry: %v", err)
	}
	if got.Title != entry.Title {
		t.Errorf("Title = %q, want %q", got.Title, entry.Title)
	}
	if got.URL != entry.URL {
		t.Errorf("URL = %q, want %q", got.URL, entry.URL)
	}
	if got.Summary != entry.Summary {
		t.Errorf("Summary = %q, want %q", got.Summary, entry.Summary)
	}
	if got.Author != entry.Author {
		t.Errorf("Author = %q, want %q", got.Author, entry.Author)
	}
	if got.Read {
		t.Error("expected Read=false")
	}
}

func TestGetEntry_Missing(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	_, err := s.GetEntry(ctx, "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestDeleteFeedCascadeDeletesEntries(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed := &domain.Feed{ID: uuid.New().String(), Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS}
	s.AddFeed(ctx, feed)

	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feed.ID, ExternalID: "e1",
		Title: "Entry", FetchedAt: time.Now(),
	})

	if err := s.DeleteFeed(ctx, feed.ID); err != nil {
		t.Fatalf("DeleteFeed: %v", err)
	}

	entries, err := s.ListEntries(ctx, feed.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListEntries after delete: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) after delete = %d, want 0", len(entries))
	}

	_, err = s.GetFeed(ctx, feed.ID)
	if err == nil {
		t.Error("expected error getting deleted feed")
	}
}

func TestUpdateFeed(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed := &domain.Feed{
		ID:      uuid.New().String(),
		Title:   "Original",
		FeedURL: "https://example.com/feed",
		FeedType: domain.FeedTypeRSS,
	}
	if err := s.AddFeed(ctx, feed); err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	feed.Title = "Updated"
	feed.Description = "New description"
	if err := s.UpdateFeed(ctx, feed); err != nil {
		t.Fatalf("UpdateFeed: %v", err)
	}

	got, err := s.GetFeed(ctx, feed.ID)
	if err != nil {
		t.Fatalf("GetFeed after update: %v", err)
	}
	if got.Title != "Updated" {
		t.Errorf("Title = %q, want %q", got.Title, "Updated")
	}
	if got.Description != "New description" {
		t.Errorf("Description = %q, want %q", got.Description, "New description")
	}
}

func TestListEntriesUnreadAllFeeds(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed := &domain.Feed{ID: uuid.New().String(), Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS}
	s.AddFeed(ctx, feed)

	e1 := &domain.Entry{
		ID: uuid.New().String(), FeedID: feed.ID, ExternalID: "e1",
		Title: "Unread 1", FetchedAt: time.Now(),
	}
	e2 := &domain.Entry{
		ID: uuid.New().String(), FeedID: feed.ID, ExternalID: "e2",
		Title: "Read 1", FetchedAt: time.Now(),
	}
	s.UpsertEntry(ctx, e1)
	s.UpsertEntry(ctx, e2)
	s.MarkEntryRead(ctx, e2.ID)

	entries, err := s.ListEntriesUnread(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("ListEntriesUnread all feeds: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].Title != "Unread 1" {
		t.Errorf("entries[0].Title = %q, want %q", entries[0].Title, "Unread 1")
	}
}

func TestSearchEntries_MatchTitle(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test Feed", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e1",
		Title: "Hello World", PublishedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e2",
		Title: "Goodbye World", PublishedAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})

	entries, err := s.SearchEntries(ctx, "Hello", 10, 0)
	if err != nil {
		t.Fatalf("SearchEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].Title != "Hello World" {
		t.Errorf("entries[0].Title = %q, want %q", entries[0].Title, "Hello World")
	}
}

func TestSearchEntries_MatchSummary(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test Feed", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e1",
		Title: "Entry", Summary: "This contains a secret keyword",
		FetchedAt: time.Now(),
	})

	entries, err := s.SearchEntries(ctx, "secret", 10, 0)
	if err != nil {
		t.Fatalf("SearchEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
}

func TestSearchEntries_NoMatch(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test Feed", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e1",
		Title: "Entry", FetchedAt: time.Now(),
	})

	entries, err := s.SearchEntries(ctx, "nonexistent", 10, 0)
	if err != nil {
		t.Fatalf("SearchEntries: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0", len(entries))
	}
}

func TestSearchEntries_DefaultLimit(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test Feed", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	for i := range 60 {
		s.UpsertEntry(ctx, &domain.Entry{
			ID: uuid.New().String(), FeedID: feedID, ExternalID: fmt.Sprintf("e%d", i),
			Title: fmt.Sprintf("Match Entry %d", i), FetchedAt: time.Now(),
		})
	}

	entries, err := s.SearchEntries(ctx, "Match", 0, 0)
	if err != nil {
		t.Fatalf("SearchEntries zero limit: %v", err)
	}
	if len(entries) != 50 {
		t.Errorf("len(entries) = %d, want 50 (default limit)", len(entries))
	}
}

func TestMarkFeedRead(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed1 := &domain.Feed{ID: uuid.New().String(), Title: "Feed 1", FeedURL: "https://a.com/feed", FeedType: domain.FeedTypeRSS}
	feed2 := &domain.Feed{ID: uuid.New().String(), Title: "Feed 2", FeedURL: "https://b.com/feed", FeedType: domain.FeedTypeRSS}
	s.AddFeed(ctx, feed1)
	s.AddFeed(ctx, feed2)

	for _, fid := range []string{feed1.ID, feed2.ID} {
		for i := range 3 {
			s.UpsertEntry(ctx, &domain.Entry{
				ID: uuid.New().String(), FeedID: fid, ExternalID: fmt.Sprintf("e%d", i),
				Title: "Entry", FetchedAt: time.Now(),
			})
		}
	}

	if err := s.MarkFeedRead(ctx, feed1.ID); err != nil {
		t.Fatalf("MarkFeedRead: %v", err)
	}

	unread1, _ := s.ListEntriesUnread(ctx, feed1.ID, 10, 0)
	if len(unread1) != 0 {
		t.Errorf("feed1 unread = %d, want 0", len(unread1))
	}

	unread2, _ := s.ListEntriesUnread(ctx, feed2.ID, 10, 0)
	if len(unread2) != 3 {
		t.Errorf("feed2 unread = %d, want 3 (unchanged)", len(unread2))
	}
}

func TestMarkAllRead(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed1 := &domain.Feed{ID: uuid.New().String(), Title: "Feed 1", FeedURL: "https://a.com/feed", FeedType: domain.FeedTypeRSS}
	feed2 := &domain.Feed{ID: uuid.New().String(), Title: "Feed 2", FeedURL: "https://b.com/feed", FeedType: domain.FeedTypeRSS}
	s.AddFeed(ctx, feed1)
	s.AddFeed(ctx, feed2)

	for _, fid := range []string{feed1.ID, feed2.ID} {
		for i := range 2 {
			s.UpsertEntry(ctx, &domain.Entry{
				ID: uuid.New().String(), FeedID: fid, ExternalID: fmt.Sprintf("e%d", i),
				Title: "Entry", FetchedAt: time.Now(),
			})
		}
	}

	if err := s.MarkAllRead(ctx); err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}

	unread1, _ := s.ListEntriesUnread(ctx, feed1.ID, 10, 0)
	if len(unread1) != 0 {
		t.Errorf("feed1 unread = %d, want 0", len(unread1))
	}

	unread2, _ := s.ListEntriesUnread(ctx, feed2.ID, 10, 0)
	if len(unread2) != 0 {
		t.Errorf("feed2 unread = %d, want 0", len(unread2))
	}

	entries, _ := s.ListEntries(ctx, feed1.ID, 10, 0)
	if len(entries) != 2 {
		t.Errorf("entries still exist = %d, want 2", len(entries))
	}
}

func TestDeleteEntriesOlderThan(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	now := time.Now()
	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "old",
		Title: "Old Entry", PublishedAt: now.Add(-90 * 24 * time.Hour), FetchedAt: time.Now(),
	})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "recent",
		Title: "Recent Entry", PublishedAt: now.Add(-5 * 24 * time.Hour), FetchedAt: time.Now(),
	})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "today",
		Title: "Today Entry", PublishedAt: now, FetchedAt: time.Now(),
	})

	n, err := s.DeleteEntriesOlderThan(ctx, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("DeleteEntriesOlderThan: %v", err)
	}
	if n != 1 {
		t.Errorf("deleted %d, want 1", n)
	}

	remaining, err := s.ListEntries(ctx, feedID, 10, 0)
	if err != nil {
		t.Fatalf("ListEntries: %v", err)
	}
	if len(remaining) != 2 {
		t.Fatalf("len(remaining) = %d, want 2", len(remaining))
	}
	if remaining[0].Title != "Today Entry" {
		t.Errorf("remaining[0].Title = %q, want %q", remaining[0].Title, "Today Entry")
	}
	if remaining[1].Title != "Recent Entry" {
		t.Errorf("remaining[1].Title = %q, want %q", remaining[1].Title, "Recent Entry")
	}
}

func TestDeleteEntriesOlderThan_ZeroAge(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e1",
		Title: "Entry", PublishedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})

	n, err := s.DeleteEntriesOlderThan(ctx, 0)
	if err != nil {
		t.Fatalf("DeleteEntriesOlderThan(0): %v", err)
	}
	if n != 0 {
		t.Errorf("deleted %d, want 0", n)
	}
}

func TestDeleteEntriesOlderThan_AllDeleted(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "v1",
		Title: "Very Old", PublishedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "v2",
		Title: "Also Old", PublishedAt: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})

	n, err := s.DeleteEntriesOlderThan(ctx, 1*time.Hour)
	if err != nil {
		t.Fatalf("DeleteEntriesOlderThan: %v", err)
	}
	if n != 2 {
		t.Errorf("deleted %d, want 2", n)
	}

	remaining, err := s.ListEntries(ctx, feedID, 10, 0)
	if err != nil {
		t.Fatalf("ListEntries: %v", err)
	}
	if len(remaining) != 0 {
		t.Errorf("len(remaining) = %d, want 0", len(remaining))
	}
}

func TestDeleteEntriesOlderThanForFeed(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID1 := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID1, Title: "Feed A", FeedURL: "https://a.com/feed", FeedType: domain.FeedTypeRSS})
	feedID2 := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID2, Title: "Feed B", FeedURL: "https://b.com/feed", FeedType: domain.FeedTypeAtom})

	now := time.Now()
	for _, id := range []string{feedID1, feedID2} {
		s.UpsertEntry(ctx, &domain.Entry{
			ID: uuid.New().String(), FeedID: id, ExternalID: "old-" + id,
			Title: "Old Entry", PublishedAt: now.Add(-90 * 24 * time.Hour), FetchedAt: time.Now(),
		})
		s.UpsertEntry(ctx, &domain.Entry{
			ID: uuid.New().String(), FeedID: id, ExternalID: "new-" + id,
			Title: "New Entry", PublishedAt: now, FetchedAt: time.Now(),
		})
	}

	n, err := s.DeleteEntriesOlderThanForFeed(ctx, feedID1, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("DeleteEntriesOlderThanForFeed: %v", err)
	}
	if n != 1 {
		t.Errorf("deleted %d, want 1", n)
	}

	remaining1, err := s.ListEntries(ctx, feedID1, 10, 0)
	if err != nil {
		t.Fatalf("ListEntries feed1: %v", err)
	}
	if len(remaining1) != 1 {
		t.Errorf("len(remaining1) = %d, want 1", len(remaining1))
	}
	if remaining1[0].Title != "New Entry" {
		t.Errorf("remaining[0].Title = %q, want %q", remaining1[0].Title, "New Entry")
	}

	remaining2, err := s.ListEntries(ctx, feedID2, 10, 0)
	if err != nil {
		t.Fatalf("ListEntries feed2: %v", err)
	}
	if len(remaining2) != 2 {
		t.Errorf("len(remaining2) = %d, want 2 (other feed unaffected)", len(remaining2))
	}
}

func TestDeleteEntriesOlderThanForFeed_ZeroAge(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e1",
		Title: "Entry", PublishedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})

	n, err := s.DeleteEntriesOlderThanForFeed(ctx, feedID, 0)
	if err != nil {
		t.Fatalf("DeleteEntriesOlderThanForFeed(0): %v", err)
	}
	if n != 0 {
		t.Errorf("deleted %d, want 0", n)
	}
}

func TestStarEntry(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	entryID := uuid.New().String()
	s.UpsertEntry(ctx, &domain.Entry{
		ID: entryID, FeedID: feedID, ExternalID: "e1",
		Title: "Entry", FetchedAt: time.Now(),
	})

	if err := s.StarEntry(ctx, entryID); err != nil {
		t.Fatalf("StarEntry: %v", err)
	}

	entries, err := s.ListStarredEntries(ctx, feedID, 10, 0)
	if err != nil {
		t.Fatalf("ListStarredEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if !entries[0].Starred {
		t.Error("expected entry to be starred")
	}
}

func TestUnstarEntry(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	entryID := uuid.New().String()
	s.UpsertEntry(ctx, &domain.Entry{
		ID: entryID, FeedID: feedID, ExternalID: "e1",
		Title: "Entry", FetchedAt: time.Now(),
	})

	s.StarEntry(ctx, entryID)
	if err := s.UnstarEntry(ctx, entryID); err != nil {
		t.Fatalf("UnstarEntry: %v", err)
	}

	entries, err := s.ListStarredEntries(ctx, feedID, 10, 0)
	if err != nil {
		t.Fatalf("ListStarredEntries: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) after unstar = %d, want 0", len(entries))
	}
}

func TestListStarredEntries(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	e1ID := uuid.New().String()
	e2ID := uuid.New().String()

	s.UpsertEntry(ctx, &domain.Entry{
		ID: e1ID, FeedID: feedID, ExternalID: "e1",
		Title: "Starred Entry", PublishedAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: e2ID, FeedID: feedID, ExternalID: "e2",
		Title: "Not Starred", PublishedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})

	s.StarEntry(ctx, e1ID)

	entries, err := s.ListStarredEntries(ctx, feedID, 10, 0)
	if err != nil {
		t.Fatalf("ListStarredEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].Title != "Starred Entry" {
		t.Errorf("entries[0].Title = %q, want %q", entries[0].Title, "Starred Entry")
	}
}

func TestListStarredEntries_Empty(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feedID := uuid.New().String()
	s.AddFeed(ctx, &domain.Feed{ID: feedID, Title: "Test", FeedURL: "https://example.com/feed", FeedType: domain.FeedTypeRSS})

	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feedID, ExternalID: "e1",
		Title: "Entry", FetchedAt: time.Now(),
	})

	entries, err := s.ListStarredEntries(ctx, feedID, 10, 0)
	if err != nil {
		t.Fatalf("ListStarredEntries: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0", len(entries))
	}
}

func TestListStarredEntries_AllFeeds(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	feed1 := &domain.Feed{ID: uuid.New().String(), Title: "Feed A", FeedURL: "https://a.com/feed", FeedType: domain.FeedTypeRSS}
	feed2 := &domain.Feed{ID: uuid.New().String(), Title: "Feed B", FeedURL: "https://b.com/feed", FeedType: domain.FeedTypeAtom}
	s.AddFeed(ctx, feed1)
	s.AddFeed(ctx, feed2)

	e1ID := uuid.New().String()
	s.UpsertEntry(ctx, &domain.Entry{
		ID: e1ID, FeedID: feed1.ID, ExternalID: "a1",
		Title: "Feed A Entry", PublishedAt: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})
	s.UpsertEntry(ctx, &domain.Entry{
		ID: uuid.New().String(), FeedID: feed2.ID, ExternalID: "b1",
		Title: "Feed B Entry", PublishedAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), FetchedAt: time.Now(),
	})

	s.StarEntry(ctx, e1ID)

	entries, err := s.ListStarredEntries(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("ListStarredEntries all feeds: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].FeedTitle != "Feed A" {
		t.Errorf("entries[0].FeedTitle = %q, want %q", entries[0].FeedTitle, "Feed A")
	}
}
