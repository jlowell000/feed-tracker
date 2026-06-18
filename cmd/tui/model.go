package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/jlowell000/feed-tracker/internal/config"
	"github.com/jlowell000/feed-tracker/internal/domain"
	"github.com/jlowell000/feed-tracker/internal/feedtracker"
	"github.com/jlowell000/feed-tracker/internal/opml"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

type itemKind int

const (
	folderHeaderItem itemKind = iota
	feedItem
)

type displayItem struct {
	kind   itemKind
	folder *domain.Folder
	feed   *domain.Feed
	unread int
	depth  int
}

type screen int

const (
	feedsListScreen screen = iota
	entriesListScreen
	entryDetailScreen
	addFeedScreen
	helpScreen
	folderCreateScreen
	folderRenameScreen
	folderPickScreen
	importScreen
)

type model struct {
	screen     screen
	prevScreen screen

	cfg     *config.Config
	store   storage.Storage
	tracker *feedtracker.Tracker

	feeds        []*domain.Feed
	folders      []*domain.Folder
	entries      []*domain.Entry
	unreadCounts map[string]int
	displayItems []displayItem
	feed         *domain.Feed
	entry        *domain.Entry

	displayCursor int
	entryCursor   int
	collapsed      map[string]bool
	moveFeedID     string
	renameFolderID string

	err    error
	status string
	width  int
	height int

	loading   bool
	fetching  bool
	showRead  bool

	textInput textinput.Model
	spinner   spinner.Model
	viewport  viewport.Model

	ready bool
}

type feedsLoadedMsg struct {
	feeds []*domain.Feed
}

type entriesLoadedMsg struct {
	entries []*domain.Entry
}

type feedAddedMsg struct {
	feed *domain.Feed
	err  error
}

type fetchCompleteMsg struct {
	totalNew int
	err      error
}

type unreadCountsLoadedMsg struct {
	counts map[string]int
}

type foldersLoadedMsg struct {
	folders []*domain.Folder
}

type folderCreatedMsg struct {
	err error
}

type folderDeletedMsg struct {
	err error
}

type folderRenamedMsg struct {
	err error
}

type feedFolderSetMsg struct {
	err error
}

type feedDeletedMsg struct {
	err error
}

type exportCompleteMsg struct {
	path string
	err  error
}

type importCompleteMsg struct {
	n   int
	err error
}

type errMsg struct {
	err error
}

func initialModel(cfg *config.Config, store storage.Storage, tracker *feedtracker.Tracker) model {
	ti := textinput.New()
	ti.Placeholder = "https://example.com/feed.xml"
	ti.Width = 60
	ti.CharLimit = 2048

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	vp := viewport.New(80, 20)

	return model{
		screen:      feedsListScreen,
		cfg:         cfg,
		store:       store,
		tracker:     tracker,
		collapsed:   make(map[string]bool),
		textInput:   ti,
		spinner:     s,
		viewport:    vp,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		loadFeedsCmd(m.store),
		m.spinner.Tick,
	)
}

func loadFeedsCmd(store storage.Storage) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		feeds, err := store.ListFeeds(ctx)
		if err != nil {
			return errMsg{err}
		}
		return feedsLoadedMsg{feeds: feeds}
	}
}

func loadEntriesCmd(store storage.Storage, feedID string, showRead bool, limit int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var entries []*domain.Entry
		var err error
		if showRead {
			entries, err = store.ListEntries(ctx, feedID, limit)
		} else {
			entries, err = store.ListEntriesUnread(ctx, feedID, limit)
		}
		if err != nil {
			return errMsg{err}
		}
		return entriesLoadedMsg{entries: entries}
	}
}

func loadUnreadCountsCmd(store storage.Storage) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		counts, err := store.UnreadCountByFeed(ctx)
		if err != nil {
			return errMsg{err}
		}
		return unreadCountsLoadedMsg{counts: counts}
	}
}

func loadFoldersCmd(store storage.Storage) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		folders, err := store.ListFolders(ctx)
		if err != nil {
			return errMsg{err}
		}
		return foldersLoadedMsg{folders: folders}
	}
}

func createFolderCmd(store storage.Storage, name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		f := &domain.Folder{
			ID:        uuid.New().String(),
			Name:      name,
			CreatedAt: time.Now(),
		}
		if err := store.AddFolder(ctx, f); err != nil {
			return folderCreatedMsg{err: err}
		}
		return folderCreatedMsg{}
	}
}

func deleteFolderCmd(store storage.Storage, folderID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := store.DeleteFolder(ctx, folderID); err != nil {
			return folderDeletedMsg{err: err}
		}
		return folderDeletedMsg{}
	}
}

func renameFolderCmd(store storage.Storage, folderID, name string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		// Delete & recreate with same ID to "rename" (simple approach)
		if err := store.DeleteFolder(ctx, folderID); err != nil {
			return folderRenamedMsg{err: err}
		}
		f := &domain.Folder{
			ID:        folderID,
			Name:      name,
			CreatedAt: time.Now(),
		}
		if err := store.AddFolder(ctx, f); err != nil {
			return folderRenamedMsg{err: err}
		}
		return folderRenamedMsg{}
	}
}

func setFeedFolderCmd(store storage.Storage, feedID, folderID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := store.SetFeedFolder(ctx, feedID, folderID); err != nil {
			return feedFolderSetMsg{err: err}
		}
		return feedFolderSetMsg{}
	}
}

func markEntryReadCmd(store storage.Storage, entryID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := store.MarkEntryRead(ctx, entryID); err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func markEntryUnreadCmd(store storage.Storage, entryID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := store.MarkEntryUnread(ctx, entryID); err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func markUnreadAndReloadCmd(store storage.Storage, feedID string, showRead bool, entryID string, limit int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := store.MarkEntryUnread(ctx, entryID); err != nil {
			return errMsg{err}
		}
		var entries []*domain.Entry
		var err error
		if showRead {
			entries, err = store.ListEntries(ctx, feedID, limit)
		} else {
			entries, err = store.ListEntriesUnread(ctx, feedID, limit)
		}
		if err != nil {
			return errMsg{err}
		}
		return entriesLoadedMsg{entries: entries}
	}
}

func addFeedCmd(tracker *feedtracker.Tracker, url string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		feed, err := tracker.AddFeed(ctx, url)
		return feedAddedMsg{feed: feed, err: err}
	}
}

func buildDisplayItems(feeds []*domain.Feed, folders []*domain.Folder, counts map[string]int, collapsed map[string]bool) []displayItem {
	var items []displayItem

	byFolder := make(map[string][]*domain.Feed)
	var ungrouped []*domain.Feed
	for _, f := range feeds {
		if f.FolderID == "" {
			ungrouped = append(ungrouped, f)
		} else {
			byFolder[f.FolderID] = append(byFolder[f.FolderID], f)
		}
	}

	for _, folder := range folders {
		folderFeeds := byFolder[folder.ID]
		total := 0
		for _, ff := range folderFeeds {
			if counts != nil {
				total += counts[ff.ID]
			}
		}
		items = append(items, displayItem{kind: folderHeaderItem, folder: folder, unread: total})
		if !collapsed[folder.ID] {
			for _, ff := range folderFeeds {
				n := 0
				if counts != nil {
					n = counts[ff.ID]
				}
				items = append(items, displayItem{kind: feedItem, feed: ff, unread: n, depth: 1})
			}
		}
	}

	for _, f := range ungrouped {
		n := 0
		if counts != nil {
			n = counts[f.ID]
		}
		items = append(items, displayItem{kind: feedItem, feed: f, unread: n, depth: 0})
	}

	return items
}

func exportFeedsCmd(store storage.Storage) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		feeds, err := store.ListFeeds(ctx)
		if err != nil {
			return exportCompleteMsg{err: fmt.Errorf("list feeds: %w", err)}
		}
		folders, err := store.ListFolders(ctx)
		if err != nil {
			return exportCompleteMsg{err: fmt.Errorf("list folders: %w", err)}
		}

		folderNames := make(map[string]string)
		for _, f := range folders {
			folderNames[f.ID] = f.Name
		}

		var specs []opml.FeedSpec
		for _, feed := range feeds {
			s := opml.FeedSpec{
				URL:   feed.FeedURL,
				Title: feed.Title,
			}
			if feed.FolderID != "" {
				s.Folder = folderNames[feed.FolderID]
			}
			specs = append(specs, s)
		}

		path := fmt.Sprintf("feed-tracker-%s.opml", time.Now().Format("2006-01-02-150405"))
		f, err := os.Create(path)
		if err != nil {
			return exportCompleteMsg{err: fmt.Errorf("create file: %w", err)}
		}
		defer f.Close()

		if err := opml.Export(specs, f); err != nil {
			return exportCompleteMsg{err: fmt.Errorf("export opml: %w", err)}
		}

		return exportCompleteMsg{path: path}
	}
}

func importFeedsCmd(tracker *feedtracker.Tracker, path string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		specs, err := opml.ParseFile(path)
		if err != nil {
			return importCompleteMsg{err: fmt.Errorf("parse opml: %w", err)}
		}
		n := 0
		for _, s := range specs {
			if _, err := tracker.AddFeed(ctx, s.URL); err != nil {
				return importCompleteMsg{err: fmt.Errorf("import feed %q: %w", s.URL, err)}
			}
			n++
		}
		return importCompleteMsg{n: n}
	}
}

func deleteFeedCmd(store storage.Storage, feedID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := store.DeleteFeed(ctx, feedID); err != nil {
			return feedDeletedMsg{err: err}
		}
		return feedDeletedMsg{}
	}
}

func fetchAllFeedsCmd(tracker *feedtracker.Tracker, store storage.Storage) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		totalNew, err := tracker.FetchAllFeeds(ctx)
		return fetchCompleteMsg{totalNew: totalNew, err: err}
	}
}
