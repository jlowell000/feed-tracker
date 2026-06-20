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
	allEntriesItem itemKind = iota
	folderHeaderItem
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
	importDryRunScreen
	exportPickScreen
	searchScreen
	editFeedScreen
	feedPickScreen
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
	entryOffset   int
	entryPageSize int
	collapsed      map[string]bool
	moveFeedID     string
	renameFolderID string
	exportFilter   string
	importSpecs    []opml.FeedSpec
	editFeed       *domain.Feed
	editTitleInput textinput.Model
	editURLInput   textinput.Model

	searchQuery  string
	filterFeedID string

	err    error
	status string
	width  int
	height int

	loading   bool
	fetching  bool
	showRead  bool

	autoRefreshInterval time.Duration
autoRefreshRemaining time.Duration

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

type moreEntriesLoadedMsg struct {
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
	n    int
	errs int
	err  error
}

type importPreviewMsg struct {
	specs []opml.FeedSpec
	err   error
}

type errMsg struct {
	err error
}

type searchResultsMsg struct {
	entries []*domain.Entry
}

type entriesMarkedReadMsg struct {
	n int
}

type feedMarkedReadMsg struct{}

type feedUpdatedMsg struct {
	err error
}

type autoRefreshCountdownMsg struct{}

func initialModel(cfg *config.Config, store storage.Storage, tracker *feedtracker.Tracker) model {
	ti := textinput.New()
	ti.Placeholder = "https://example.com/feed.xml"
	ti.Width = 60
	ti.CharLimit = 2048

	eti := textinput.New()
	eti.Placeholder = "Feed title"
	eti.Width = 60
	eti.CharLimit = 1024

	eui := textinput.New()
	eui.Placeholder = "Feed URL"
	eui.Width = 60
	eui.CharLimit = 2048

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	vp := viewport.New(80, 20)

	pageSize := cfg.TUI.EntryLimit
	if pageSize <= 0 {
		pageSize = 50
	}
	return model{
		screen:              feedsListScreen,
		cfg:                 cfg,
		store:               store,
		tracker:             tracker,
		collapsed:           make(map[string]bool),
		textInput:           ti,
		editTitleInput:      eti,
		editURLInput:        eui,
		spinner:             s,
		viewport:            vp,
		autoRefreshInterval:  cfg.TUI.AutoRefresh,
		autoRefreshRemaining: cfg.TUI.AutoRefresh,
		entryPageSize:        pageSize,
	}
}

func ctxWithTimeout(cfg *config.Config) (context.Context, context.CancelFunc) {
	timeout := cfg.HTTP.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		loadFeedsCmd(m.store, m.cfg.HTTP.Timeout),
		m.spinner.Tick,
	}
	if m.autoRefreshInterval > 0 {
		cmds = append(cmds, autoRefreshTick(m.autoRefreshInterval), countdownTick())
	}
	return tea.Batch(cmds...)
}

func loadFeedsCmd(store storage.Storage, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		feeds, err := store.ListFeeds(ctx)
		if err != nil {
			return errMsg{err}
		}
		return feedsLoadedMsg{feeds: feeds}
	}
}

func effectiveFeedID(m model) string {
	if m.feed != nil {
		return m.feed.ID
	}
	return m.filterFeedID
}

func loadEntriesCmd(store storage.Storage, feedID string, showRead bool, limit int, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		var entries []*domain.Entry
		var err error
		if showRead {
			entries, err = store.ListEntries(ctx, feedID, limit, 0)
		} else {
			entries, err = store.ListEntriesUnread(ctx, feedID, limit, 0)
		}
		if err != nil {
			return errMsg{err}
		}
		return entriesLoadedMsg{entries: entries}
	}
}

func loadMoreEntriesCmd(store storage.Storage, feedID string, showRead bool, limit, offset int, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		var entries []*domain.Entry
		var err error
		if showRead {
			entries, err = store.ListEntries(ctx, feedID, limit, offset)
		} else {
			entries, err = store.ListEntriesUnread(ctx, feedID, limit, offset)
		}
		if err != nil {
			return errMsg{err}
		}
		return moreEntriesLoadedMsg{entries: entries}
	}
}

func loadUnreadCountsCmd(store storage.Storage, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		counts, err := store.UnreadCountByFeed(ctx)
		if err != nil {
			return errMsg{err}
		}
		return unreadCountsLoadedMsg{counts: counts}
	}
}

func loadFoldersCmd(store storage.Storage, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		folders, err := store.ListFolders(ctx)
		if err != nil {
			return errMsg{err}
		}
		return foldersLoadedMsg{folders: folders}
	}
}

func createFolderCmd(store storage.Storage, name string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
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

func deleteFolderCmd(store storage.Storage, folderID string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := store.DeleteFolder(ctx, folderID); err != nil {
			return folderDeletedMsg{err: err}
		}
		return folderDeletedMsg{}
	}
}

func renameFolderCmd(store storage.Storage, folderID, name string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
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

func setFeedFolderCmd(store storage.Storage, feedID, folderID string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := store.SetFeedFolder(ctx, feedID, folderID); err != nil {
			return feedFolderSetMsg{err: err}
		}
		return feedFolderSetMsg{}
	}
}

func markEntryReadCmd(store storage.Storage, entryID string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := store.MarkEntryRead(ctx, entryID); err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func markUnreadAndReloadCmd(store storage.Storage, feedID string, showRead bool, entryID string, limit int, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := store.MarkEntryUnread(ctx, entryID); err != nil {
			return errMsg{err}
		}
		var entries []*domain.Entry
		var err error
		if showRead {
			entries, err = store.ListEntries(ctx, feedID, limit, 0)
		} else {
			entries, err = store.ListEntriesUnread(ctx, feedID, limit, 0)
		}
		if err != nil {
			return errMsg{err}
		}
		return entriesLoadedMsg{entries: entries}
	}
}

func addFeedCmd(tracker *feedtracker.Tracker, url string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		feed, err := tracker.AddFeed(ctx, url)
		return feedAddedMsg{feed: feed, err: err}
	}
}

func buildDisplayItems(feeds []*domain.Feed, folders []*domain.Folder, counts map[string]int, collapsed map[string]bool) []displayItem {
	totalUnread := 0
	for _, n := range counts {
		totalUnread += n
	}
	items := []displayItem{
		{kind: allEntriesItem, unread: totalUnread},
	}

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

func exportFeedsCmd(store storage.Storage, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
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

func importFeedsCmd(tracker *feedtracker.Tracker, store storage.Storage, path string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		specs, err := opml.ParseFile(path)
		if err != nil {
			return importCompleteMsg{err: fmt.Errorf("parse opml: %w", err)}
		}
		n := 0
		errs := 0
		for _, s := range specs {
			feed, addErr := tracker.AddFeed(ctx, s.URL)
			if addErr != nil {
				return importCompleteMsg{err: fmt.Errorf("import feed %q: %w", s.URL, addErr)}
			}
			n++

			if s.Folder != "" {
				folder, fErr := store.GetFolderByName(ctx, s.Folder)
				if fErr != nil {
					folder = &domain.Folder{
						ID:        uuid.New().String(),
						Name:      s.Folder,
						CreatedAt: time.Now(),
					}
					if aErr := store.AddFolder(ctx, folder); aErr != nil {
						errs++
						continue
					}
				}
				if sErr := store.SetFeedFolder(ctx, feed.ID, folder.ID); sErr != nil {
					errs++
				}
			}
		}
		return importCompleteMsg{n: n, errs: errs}
	}
}

func deleteFeedCmd(store storage.Storage, feedID string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := store.DeleteFeed(ctx, feedID); err != nil {
			return feedDeletedMsg{err: err}
		}
		return feedDeletedMsg{}
	}
}

func fetchAllFeedsCmd(tracker *feedtracker.Tracker, store storage.Storage, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		totalNew, err := tracker.FetchAllFeeds(ctx)
		return fetchCompleteMsg{totalNew: totalNew, err: err}
	}
}

func autoRefreshTick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return fetchCompleteMsg{}
	})
}

func countdownTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return autoRefreshCountdownMsg{}
	})
}

func importPreviewCmd(tracker *feedtracker.Tracker, path string) tea.Cmd {
	return func() tea.Msg {
		specs, err := opml.ParseFile(path)
		return importPreviewMsg{specs: specs, err: err}
	}
}

func searchEntriesCmd(store storage.Storage, query string, limit int, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		entries, err := store.SearchEntries(ctx, query, limit, 0)
		if err != nil {
			return errMsg{err}
		}
		return searchResultsMsg{entries: entries}
	}
}

func markDisplayedReadCmd(store storage.Storage, entries []*domain.Entry, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		n := 0
		for _, e := range entries {
			if !e.Read {
				if err := store.MarkEntryRead(ctx, e.ID); err != nil {
					return errMsg{err}
				}
				n++
			}
		}
		return entriesMarkedReadMsg{n: n}
	}
}

func markFeedReadAllCmd(store storage.Storage, feedID string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		var err error
		if feedID == "" {
			err = store.MarkAllRead(ctx)
		} else {
			err = store.MarkFeedRead(ctx, feedID)
		}
		if err != nil {
			return errMsg{err}
		}
		return feedMarkedReadMsg{}
	}
}

func exportFilteredCmd(store storage.Storage, filter string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
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
			if filter == "folders" && feed.FolderID == "" {
				continue
			}
			if filter == "feeds" && feed.FolderID != "" {
				continue
			}
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
