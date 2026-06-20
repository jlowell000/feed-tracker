package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/jlowell000/feed-tracker/internal/domain"
	"github.com/jlowell000/feed-tracker/internal/storage"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 7
		if !m.ready {
			m.ready = true
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		return m.handleKeyMsg(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case feedsLoadedMsg:
		m.feeds = msg.feeds
		if m.feed != nil {
			for _, f := range m.feeds {
				if f.ID == m.feed.ID {
					m.feed = f
					break
				}
			}
		}
		m.displayItems = buildDisplayItems(m.feeds, m.folders, m.unreadCounts, m.collapsed)
		if m.displayCursor >= len(m.displayItems) {
			m.displayCursor = max(0, len(m.displayItems)-1)
		}
		return m, tea.Batch(
			loadUnreadCountsCmd(m.store, m.cfg.HTTP.Timeout),
			loadFoldersCmd(m.store, m.cfg.HTTP.Timeout),
		)

	case foldersLoadedMsg:
		m.folders = msg.folders
		m.displayItems = buildDisplayItems(m.feeds, m.folders, m.unreadCounts, m.collapsed)
		if m.displayCursor >= len(m.displayItems) {
			m.displayCursor = max(0, len(m.displayItems)-1)
		}
		return m, nil

	case unreadCountsLoadedMsg:
		m.unreadCounts = msg.counts
		m.displayItems = buildDisplayItems(m.feeds, m.folders, m.unreadCounts, m.collapsed)
		return m, nil

	case folderCreatedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error creating folder: %v", msg.err)
		} else {
			m.status = "Folder created"
		}
		return m, loadFeedsCmd(m.store, m.cfg.HTTP.Timeout)

	case folderDeletedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error deleting folder: %v", msg.err)
		} else {
			m.status = "Folder deleted"
		}
		return m, loadFeedsCmd(m.store, m.cfg.HTTP.Timeout)

	case folderRenamedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error renaming folder: %v", msg.err)
		} else {
			m.status = "Folder renamed"
		}
		return m, loadFeedsCmd(m.store, m.cfg.HTTP.Timeout)

	case feedFolderSetMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error moving feed: %v", msg.err)
		} else {
			m.status = "Feed moved"
		}
		return m, loadFeedsCmd(m.store, m.cfg.HTTP.Timeout)

	case feedDeletedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error deleting feed: %v", msg.err)
		} else {
			m.status = "Feed deleted"
		}
		return m, loadFeedsCmd(m.store, m.cfg.HTTP.Timeout)

	case exportCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Export error: %v", msg.err)
		} else {
			m.status = fmt.Sprintf("Exported %d feeds to %s", len(m.feeds), msg.path)
		}
		return m, nil

	case importPreviewMsg:
		m.loading = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
			m.screen = feedsListScreen
		} else {
			m.importSpecs = msg.specs
			m.screen = importDryRunScreen
		}
		return m, nil

	case importCompleteMsg:
		m.loading = false
		m.importSpecs = nil
		m.screen = feedsListScreen
		if msg.err != nil {
			m.status = fmt.Sprintf("Import error: %v", msg.err)
		} else {
			s := fmt.Sprintf("Imported %d feeds", msg.n)
			if msg.errs > 0 {
				s += fmt.Sprintf(" (%d folder errors)", msg.errs)
			}
			m.status = s
		}
		return m, loadFeedsCmd(m.store, m.cfg.HTTP.Timeout)

	case entriesLoadedMsg:
		m.entries = msg.entries
		m.searchQuery = ""
		m.entryCursor = 0
		m.entryOffset = len(msg.entries)
		return m, nil

	case moreEntriesLoadedMsg:
		m.entries = append(m.entries, msg.entries...)
		m.entryOffset = len(m.entries)
		return m, nil

	case searchResultsMsg:
		m.entries = msg.entries
		m.entryCursor = 0
		m.entryOffset = len(msg.entries)
		return m, nil

	case entriesMarkedReadMsg:
		m.status = fmt.Sprintf("Marked %d entries as read", msg.n)
		return m, loadEntriesCmd(m.store, effectiveFeedID(m), m.showRead, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)

	case feedMarkedReadMsg:
		m.status = "Marked all entries as read"
		return m, tea.Batch(
			loadEntriesCmd(m.store, effectiveFeedID(m), m.showRead, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout),
			loadUnreadCountsCmd(m.store, m.cfg.HTTP.Timeout),
		)

	case feedUpdatedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error updating feed: %v", msg.err)
		} else {
			m.status = "Feed updated"
		}
		cmds := []tea.Cmd{loadFeedsCmd(m.store, m.cfg.HTTP.Timeout)}
		if m.screen == entriesListScreen {
			cmds = append(cmds, loadEntriesCmd(m.store, effectiveFeedID(m), m.showRead, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout))
		}
		return m, tea.Batch(cmds...)

	case feedAddedMsg:
		m.loading = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error adding feed: %v", msg.err)
		} else {
			m.status = fmt.Sprintf("Added feed: %s", msg.feed.Title)
		}
		return m, loadFeedsCmd(m.store, m.cfg.HTTP.Timeout)

	case fetchCompleteMsg:
		m.fetching = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Fetch error: %v", msg.err)
		} else {
			m.status = fmt.Sprintf("Fetched all — %d new entries", msg.totalNew)
		}
		cmds := []tea.Cmd{loadFeedsCmd(m.store, m.cfg.HTTP.Timeout)}
		if m.autoRefreshInterval > 0 {
			m.autoRefreshRemaining = m.autoRefreshInterval
			cmds = append(cmds, autoRefreshTick(m.autoRefreshInterval), countdownTick())
		}
		return m, tea.Batch(cmds...)

	case autoRefreshCountdownMsg:
		if m.autoRefreshInterval > 0 {
			m.autoRefreshRemaining -= time.Second
			if m.autoRefreshRemaining < 0 {
				m.autoRefreshRemaining = 0
			}
			return m, countdownTick()
		}
		return m, nil

	case errMsg:
		m.loading = false
		m.status = fmt.Sprintf("Error: %v", msg.err)
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case feedsListScreen:
		return m.handleFeedsListKey(msg)
	case entriesListScreen:
		return m.handleEntriesListKey(msg)
	case entryDetailScreen:
		return m.handleEntryDetailKey(msg)
	case addFeedScreen:
		return m.handleAddFeedKey(msg)
	case helpScreen:
		if msg.String() == "?" || msg.Type == tea.KeyEscape || msg.Type == tea.KeyEnter {
			m.screen = m.prevScreen
		}
		return m, nil
	case folderCreateScreen:
		return m.handleFolderCreateKey(msg)
	case folderRenameScreen:
		return m.handleFolderRenameKey(msg)
	case folderPickScreen:
		return m.handleFolderPickKey(msg)
	case importScreen:
		return m.handleImportKey(msg)
	case importDryRunScreen:
		return m.handleImportDryRunKey(msg)
	case exportPickScreen:
		return m.handleExportPickKey(msg)
	case feedPickScreen:
		return m.handleFeedPickKey(msg)
	case searchScreen:
		return m.handleSearchKey(msg)
	case editFeedScreen:
		return m.handleEditFeedKey(msg)
	}
	return m, nil
}

func (m model) handleFeedsListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.displayCursor > 0 {
			m.displayCursor--
		}
	case "down", "j":
		if m.displayCursor < len(m.displayItems)-1 {
			m.displayCursor++
		}
	case "pgup":
		page := max(1, m.height-5)
		m.displayCursor = max(0, m.displayCursor-page)
	case "pgdown":
		page := max(1, m.height-5)
		m.displayCursor = min(len(m.displayItems)-1, m.displayCursor+page)
	case "home":
		m.displayCursor = 0
	case "end":
		m.displayCursor = len(m.displayItems) - 1
	case "enter":
		if len(m.displayItems) == 0 {
			return m, nil
		}
		item := m.displayItems[m.displayCursor]
		switch item.kind {
		case allEntriesItem:
			m.feed = nil
			m.filterFeedID = ""
			m.prevScreen = m.screen
			m.screen = entriesListScreen
			return m, loadEntriesCmd(m.store, effectiveFeedID(m), m.showRead, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)
		case folderHeaderItem:
			if item.folder != nil {
				id := item.folder.ID
				if m.collapsed[id] {
					delete(m.collapsed, id)
				} else {
					m.collapsed[id] = true
				}
				m.displayItems = buildDisplayItems(m.feeds, m.folders, m.unreadCounts, m.collapsed)
			}
		case feedItem:
			if item.feed != nil {
				m.feed = item.feed
				m.filterFeedID = ""
				m.prevScreen = m.screen
				m.screen = entriesListScreen
				return m, loadEntriesCmd(m.store, effectiveFeedID(m), m.showRead, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)
			}
		}
	case "E":
		if len(m.displayItems) > 0 && m.displayCursor < len(m.displayItems) {
			item := m.displayItems[m.displayCursor]
			if item.kind == feedItem && item.feed != nil {
				m.editFeed = item.feed
				m.editTitleInput.SetValue(item.feed.Title)
				m.editURLInput.SetValue(item.feed.FeedURL)
				m.prevScreen = m.screen
				m.screen = editFeedScreen
				m.editTitleInput.Focus()
			}
		}
	case "e":
		if !m.loading && !m.fetching && len(m.feeds) > 0 {
			m.prevScreen = m.screen
			m.exportFilter = ""
			m.screen = exportPickScreen
		}
	case "i":
		m.prevScreen = m.screen
		m.screen = importScreen
		m.textInput.SetValue("")
		m.textInput.Placeholder = "/path/to/feeds.opml"
		m.textInput.Focus()
	case "a":
		m.prevScreen = m.screen
		m.screen = addFeedScreen
		m.textInput.SetValue("")
		m.textInput.Placeholder = "https://example.com/feed.xml"
		m.textInput.Focus()
	case "f":
		if !m.fetching && !m.loading && len(m.feeds) > 0 {
			m.fetching = true
			m.status = "Fetching all feeds..."
			return m, fetchAllFeedsCmd(m.tracker, m.store, m.cfg.HTTP.Timeout)
		}
	case "g":
		m.prevScreen = m.screen
		m.screen = folderCreateScreen
		m.textInput.SetValue("")
		m.textInput.Placeholder = "Folder name"
		m.textInput.Focus()
	case "m":
		if len(m.displayItems) > 0 && m.displayCursor < len(m.displayItems) {
			item := m.displayItems[m.displayCursor]
			if item.kind == feedItem && item.feed != nil {
				m.moveFeedID = item.feed.ID
				m.prevScreen = m.screen
				m.screen = folderPickScreen
			}
		}
	case "d":
		if len(m.displayItems) > 0 && m.displayCursor < len(m.displayItems) {
			item := m.displayItems[m.displayCursor]
			switch {
			case item.kind == folderHeaderItem && item.folder != nil:
				hasFeeds := false
				for _, f := range m.feeds {
					if f.FolderID == item.folder.ID {
						hasFeeds = true
						break
					}
				}
				if hasFeeds {
					m.status = fmt.Sprintf("Cannot delete %q — move feeds out first", item.folder.Name)
				} else {
					return m, deleteFolderCmd(m.store, item.folder.ID, m.cfg.HTTP.Timeout)
				}
			case item.kind == feedItem && item.feed != nil:
				return m, deleteFeedCmd(m.store, item.feed.ID, m.cfg.HTTP.Timeout)
			}
		}
	case "R":
		if len(m.displayItems) > 0 && m.displayCursor < len(m.displayItems) {
			item := m.displayItems[m.displayCursor]
			if item.kind == folderHeaderItem && item.folder != nil {
				m.renameFolderID = item.folder.ID
				m.prevScreen = m.screen
				m.screen = folderRenameScreen
				m.textInput.SetValue(item.folder.Name)
				m.textInput.Placeholder = "New folder name"
				m.textInput.Focus()
			}
		}
	case "r":
		m.status = ""
		return m, loadFeedsCmd(m.store, m.cfg.HTTP.Timeout)
	case "?":
		m.prevScreen = m.screen
		m.screen = helpScreen
	}
	return m, nil
}

func (m model) handleEntriesListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.entryCursor > 0 {
			m.entryCursor--
		}
	case "down", "j":
		if m.entryCursor < len(m.entries)-1 {
			m.entryCursor++
		}
	case "pgup":
		page := max(1, m.height-5)
		m.entryCursor = max(0, m.entryCursor-page)
	case "pgdown":
		page := max(1, m.height-5)
		m.entryCursor = min(len(m.entries)-1, m.entryCursor+page)
	case "home":
		m.entryCursor = 0
	case "end":
		m.entryCursor = len(m.entries) - 1
	case "enter":
		if len(m.entries) > 0 {
			m.entry = m.entries[m.entryCursor]
			m.prevScreen = m.screen
			m.screen = entryDetailScreen
			m.viewport.SetContent(entryDetailContent(m))
			m.viewport.GotoTop()
			return m, markEntryReadCmd(m.store, m.entry.ID, m.cfg.HTTP.Timeout)
		}
	case "s":
		if !m.loading && len(m.entries) > 0 {
			m.prevScreen = m.screen
			m.screen = searchScreen
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Search entries..."
			m.textInput.Focus()
		}
	case "f":
		if m.feed == nil && !m.loading && len(m.feeds) > 0 {
			m.prevScreen = m.screen
			m.screen = feedPickScreen
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Feed number (0 for none)"
			m.textInput.Focus()
		}
	case "u":
		m.showRead = !m.showRead
		return m, loadEntriesCmd(m.store, effectiveFeedID(m), m.showRead, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)
	case "a":
		if len(m.entries) > 0 {
			return m, markDisplayedReadCmd(m.store, m.entries, m.cfg.HTTP.Timeout)
		}
	case "A":
		if !m.loading {
			return m, markFeedReadAllCmd(m.store, effectiveFeedID(m), m.cfg.HTTP.Timeout)
		}
	case "esc":
		if m.filterFeedID != "" {
			m.screen = feedPickScreen
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Feed number (0 for none)"
			m.textInput.Focus()
			return m, nil
		}
		m.screen = feedsListScreen
		m.feed = nil
		m.filterFeedID = ""
		m.entry = nil
		m.entryCursor = 0
		m.searchQuery = ""
		return m, loadFeedsCmd(m.store, m.cfg.HTTP.Timeout)
	case "r":
		m.entryOffset = 0
		m.searchQuery = ""
		return m, loadEntriesCmd(m.store, effectiveFeedID(m), m.showRead, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)
	case "E":
		currentID := effectiveFeedID(m)
		if currentID != "" {
			for _, f := range m.feeds {
				if f.ID == currentID {
					m.editFeed = f
					m.editTitleInput.SetValue(f.Title)
					m.editURLInput.SetValue(f.FeedURL)
					m.prevScreen = m.screen
					m.screen = editFeedScreen
					m.editTitleInput.Focus()
					return m, nil
				}
			}
		}
	case "L":
		if m.entryOffset > 0 && len(m.entries) >= m.entryPageSize {
			return m, loadMoreEntriesCmd(m.store, effectiveFeedID(m), m.showRead, m.entryPageSize, m.entryOffset, m.cfg.HTTP.Timeout)
		}
	case "[":
		if len(m.displayItems) == 0 {
			return m, nil
		}
		currentID := effectiveFeedID(m)
		if currentID == "" {
			return m, nil
		}
		idx := -1
		for i, item := range m.displayItems {
			if item.kind == feedItem && item.feed != nil && item.feed.ID == currentID {
				idx = i
				break
			}
		}
		if idx < 0 {
			return m, nil
		}
		for i := idx - 1; i >= 0; i-- {
			if m.displayItems[i].kind == feedItem && m.displayItems[i].feed != nil {
				m.feed = m.displayItems[i].feed
				m.filterFeedID = ""
				m.showRead = false
				return m, loadEntriesCmd(m.store, m.feed.ID, m.showRead, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)
			}
		}
	case "]":
		if len(m.displayItems) == 0 {
			return m, nil
		}
		currentID := effectiveFeedID(m)
		if currentID == "" {
			return m, nil
		}
		idx := -1
		for i, item := range m.displayItems {
			if item.kind == feedItem && item.feed != nil && item.feed.ID == currentID {
				idx = i
				break
			}
		}
		if idx < 0 {
			return m, nil
		}
		for i := idx + 1; i < len(m.displayItems); i++ {
			if m.displayItems[i].kind == feedItem && m.displayItems[i].feed != nil {
				m.feed = m.displayItems[i].feed
				m.filterFeedID = ""
				m.showRead = false
				return m, loadEntriesCmd(m.store, m.feed.ID, m.showRead, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)
			}
		}
	case "?":
		m.prevScreen = m.screen
		m.screen = helpScreen
	}
	return m, nil
}

func (m model) handleEntryDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg.String() {
	case "esc", "backspace":
		m.screen = entriesListScreen
		m.entry = nil
		if m.searchQuery != "" {
			return m, searchEntriesCmd(m.store, m.searchQuery, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)
		}
		return m, loadEntriesCmd(m.store, effectiveFeedID(m), m.showRead, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)
	case "o":
		if m.entry != nil && m.entry.URL != "" {
			openURL(m.entry.URL)
		}
	case "M":
		if m.entry != nil {
			m.screen = entriesListScreen
			entryID := m.entry.ID
			m.entry = nil
			if m.searchQuery != "" {
				ctx, cancel := context.WithTimeout(context.Background(), m.cfg.HTTP.Timeout)
				defer cancel()
				if err := m.store.MarkEntryUnread(ctx, entryID); err != nil {
					return m, func() tea.Msg { return errMsg{err} }
				}
				return m, searchEntriesCmd(m.store, m.searchQuery, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)
			}
			return m, markUnreadAndReloadCmd(m.store, effectiveFeedID(m), m.showRead, entryID, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)
		}
	case "?":
		m.prevScreen = m.screen
		m.screen = helpScreen
		return m, nil
	}
	return m, vpCmd
}

func (m model) handleAddFeedKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var tiCmd tea.Cmd
	m.textInput, tiCmd = m.textInput.Update(msg)

	switch msg.Type {
	case tea.KeyEnter:
		url := m.textInput.Value()
		if url != "" {
			m.loading = true
			m.status = "Adding feed..."
			m.textInput.Blur()
			m.screen = feedsListScreen
			return m, tea.Batch(tiCmd, addFeedCmd(m.tracker, url, m.cfg.HTTP.Timeout))
		}
	case tea.KeyEscape:
		m.textInput.Blur()
		m.screen = feedsListScreen
		return m, tiCmd
	}

	return m, tiCmd
}

func (m model) handleFolderCreateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var tiCmd tea.Cmd
	m.textInput, tiCmd = m.textInput.Update(msg)

	switch msg.Type {
	case tea.KeyEnter:
		name := m.textInput.Value()
		if name != "" {
			m.textInput.Blur()
			m.screen = feedsListScreen
			return m, tea.Batch(tiCmd, createFolderCmd(m.store, name, m.cfg.HTTP.Timeout))
		}
	case tea.KeyEscape:
		m.textInput.Blur()
		m.screen = feedsListScreen
		return m, tiCmd
	}
	return m, tiCmd
}

func (m model) handleFolderRenameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var tiCmd tea.Cmd
	m.textInput, tiCmd = m.textInput.Update(msg)

	switch msg.Type {
	case tea.KeyEnter:
		name := m.textInput.Value()
		if name != "" && m.renameFolderID != "" {
			m.textInput.Blur()
			rid := m.renameFolderID
			m.renameFolderID = ""
			m.screen = feedsListScreen
			return m, tea.Batch(tiCmd, renameFolderCmd(m.store, rid, name, m.cfg.HTTP.Timeout))
		}
	case tea.KeyEscape:
		m.textInput.Blur()
		m.renameFolderID = ""
		m.screen = feedsListScreen
		return m, tiCmd
	}
	return m, tiCmd
}

func (m model) handleImportDryRunKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if len(m.importSpecs) > 0 {
			m.loading = true
			m.status = "Importing feeds..."
			path := m.textInput.Value()
			return m, importFeedsCmd(m.tracker, m.store, path, m.cfg.HTTP.Timeout)
		}
		return m, nil
	case "esc", "q":
		m.importSpecs = nil
		m.screen = feedsListScreen
		return m, nil
	}
	return m, nil
}

func (m model) handleFeedPickKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var tiCmd tea.Cmd
	m.textInput, tiCmd = m.textInput.Update(msg)

	switch msg.Type {
	case tea.KeyEscape:
		m.textInput.Blur()
		m.screen = feedsListScreen
		return m, tea.Batch(tiCmd, loadFeedsCmd(m.store, m.cfg.HTTP.Timeout))
	case tea.KeyEnter:
		input := strings.TrimSpace(m.textInput.Value())
		m.textInput.Blur()
		if input == "" {
			m.screen = entriesListScreen
			return m, nil
		}
		n, err := strconv.Atoi(input)
		if err != nil || n < 0 || n > len(m.feeds) {
			m.status = fmt.Sprintf("Invalid feed number: %s", input)
			m.screen = entriesListScreen
			return m, nil
		}
		if n == 0 {
			m.filterFeedID = ""
		} else {
			m.filterFeedID = m.feeds[n-1].ID
		}
		m.screen = entriesListScreen
		return m, loadEntriesCmd(m.store, effectiveFeedID(m), m.showRead, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout)
	}

	return m, tiCmd
}

func (m model) handleExportPickKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "a":
		m.exportFilter = ""
		m.loading = true
		m.status = "Exporting feeds..."
		m.screen = feedsListScreen
		return m, exportFilteredCmd(m.store, "", m.cfg.HTTP.Timeout)
	case "f":
		m.exportFilter = "folders"
		m.loading = true
		m.status = "Exporting feeds..."
		m.screen = feedsListScreen
		return m, exportFilteredCmd(m.store, "folders", m.cfg.HTTP.Timeout)
	case "u":
		m.exportFilter = "feeds"
		m.loading = true
		m.status = "Exporting feeds..."
		m.screen = feedsListScreen
		return m, exportFilteredCmd(m.store, "feeds", m.cfg.HTTP.Timeout)
	case "esc":
		m.screen = feedsListScreen
		return m, nil
	}
	return m, nil
}

func (m model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var tiCmd tea.Cmd
	m.textInput, tiCmd = m.textInput.Update(msg)

	switch msg.Type {
	case tea.KeyEnter:
		query := m.textInput.Value()
		if query != "" {
			m.searchQuery = query
			m.textInput.Blur()
			m.screen = entriesListScreen
			return m, tea.Batch(tiCmd, searchEntriesCmd(m.store, query, m.cfg.TUI.EntryLimit, m.cfg.HTTP.Timeout))
		}
	case tea.KeyEscape:
		m.textInput.Blur()
		m.screen = entriesListScreen
		return m, tiCmd
	}

	return m, tiCmd
}

func updateFeedCmd(store storage.Storage, feed *domain.Feed, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := store.UpdateFeed(ctx, feed); err != nil {
			return feedUpdatedMsg{err: err}
		}
		return feedUpdatedMsg{}
	}
}

func (m model) handleEditFeedKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		ti := m.editTitleInput.Value()
		ui := m.editURLInput.Value()
		if ti != "" || ui != "" {
			feed := *m.editFeed
			if ti != "" {
				feed.Title = ti
			}
			if ui != "" {
				feed.FeedURL = ui
			}
			m.editFeed = nil
			m.screen = m.prevScreen
			return m, updateFeedCmd(m.store, &feed, m.cfg.HTTP.Timeout)
		}
	case tea.KeyEscape:
		m.editFeed = nil
		m.screen = m.prevScreen
		return m, nil
	}

	var etiCmd, euiCmd tea.Cmd
	m.editTitleInput, etiCmd = m.editTitleInput.Update(msg)
	m.editURLInput, euiCmd = m.editURLInput.Update(msg)
	return m, tea.Batch(etiCmd, euiCmd)
}

func (m model) handleImportKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var tiCmd tea.Cmd
	m.textInput, tiCmd = m.textInput.Update(msg)

	switch msg.Type {
	case tea.KeyEnter:
		path := m.textInput.Value()
		if path != "" {
			m.loading = true
			m.status = "Parsing OPML..."
			m.textInput.Blur()
			return m, tea.Batch(tiCmd, importPreviewCmd(m.tracker, path))
		}
	case tea.KeyEscape:
		m.textInput.Blur()
		m.screen = feedsListScreen
		return m, tiCmd
	}

	return m, tiCmd
}

func (m model) handleFolderPickKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.screen = feedsListScreen
		return m, nil
	case tea.KeyEnter:
		if m.moveFeedID != "" {
			fid := m.moveFeedID
			m.moveFeedID = ""
			m.screen = feedsListScreen
			return m, setFeedFolderCmd(m.store, fid, "", m.cfg.HTTP.Timeout)
		}
		return m, nil
	}

	if msg.String() == "0" || msg.String() == "1" || msg.String() == "2" || msg.String() == "3" || msg.String() == "4" || msg.String() == "5" || msg.String() == "6" || msg.String() == "7" || msg.String() == "8" || msg.String() == "9" {
		idx := int(msg.String()[0] - '0')
		if idx > 0 && idx <= len(m.folders) {
			folder := m.folders[idx-1]
			fid := m.moveFeedID
			m.moveFeedID = ""
			m.screen = feedsListScreen
			return m, setFeedFolderCmd(m.store, fid, folder.ID, m.cfg.HTTP.Timeout)
		}
	}
	return m, nil
}
