package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/jlowell000/feed-tracker/internal/domain"
	"github.com/jlowell000/feed-tracker/internal/opml"
)

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	switch m.screen {
	case feedsListScreen:
		return m.feedsListView()
	case entriesListScreen:
		return m.entriesListView()
	case entryDetailScreen:
		return m.entryDetailView()
	case addFeedScreen:
		return m.addFeedView()
	case helpScreen:
		return m.helpView()
	case folderCreateScreen:
		return m.folderCreateView()
	case folderRenameScreen:
		return m.folderRenameView()
	case folderPickScreen:
		return m.folderPickView()
	case importScreen:
		return m.importView()
	case importDryRunScreen:
		return m.importDryRunView()
	case exportPickScreen:
		return m.exportPickView()
	case feedPickScreen:
		return m.feedPickView()
	case searchScreen:
		return m.searchView()
	case editFeedScreen:
		return m.editFeedView()
	}
	return ""
}

func (m model) feedsListView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(" Feed Tracker"))
	b.WriteString("\n")
	hints := append([]*helpBinding{}, feedListHints...)
	hints = append(hints, exitHints...)
	b.WriteString(helpStyle.Width(m.width - 4).Render(renderHintLine(hints, nil)))
	b.WriteString("\n\n")

	if len(m.displayItems) == 0 {
		b.WriteString(emptyStyle.Render("  No feeds tracked yet. Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		avail := m.height - 4 // header + blank + blank + status
		if avail < 1 {
			avail = 1
		}

		start, end := windowItems(len(m.displayItems), m.displayCursor, avail)
		needsUp := start > 0
		needsDown := end < len(m.displayItems)
		indicatorLines := 0
		if needsUp {
			indicatorLines++
		}
		if needsDown {
			indicatorLines++
		}
		avail -= indicatorLines
		if avail < 1 {
			avail = 1
		}
		start, end = windowItems(len(m.displayItems), m.displayCursor, avail)
		needsUp = start > 0
		needsDown = end < len(m.displayItems)

		for i := start; i < end; i++ {
			item := m.displayItems[i]
			indent := strings.Repeat("  ", item.depth)

			switch item.kind {
			case allEntriesItem:
				line := fmt.Sprintf("  All Entries  %s", unreadCountStr(item.unread)+" unread")
				var rendered string
				if i == m.displayCursor {
					rendered = selectedItemStyle.Render("> " + line)
				} else {
					rendered = titleStyle.Render("  " + line)
				}
				b.WriteString(rendered)
				b.WriteString("\n")

			case folderHeaderItem:
				marker := "▸"
				if !m.collapsed[item.folder.ID] {
					marker = "▾"
				}
				line := fmt.Sprintf("  %s %s  %s",
					marker,
					item.folder.Name,
					unreadCountStr(item.unread)+" unread",
				)
				var rendered string
				if i == m.displayCursor {
					rendered = selectedItemStyle.Render("> " + line)
				} else {
					rendered = folderHeaderStyle.Render("  " + line)
				}
				b.WriteString(rendered)
				b.WriteString("\n")

			case feedItem:
				feed := item.feed
				title := feed.Title
				if title == "" {
					title = "(no title)"
				}
				feedType := string(feed.FeedType)
				if feedType == "" {
					feedType = "?"
				}
				lastFetched := "never"
				if !feed.LastFetched.IsZero() {
					lastFetched = formatDuration(time.Since(feed.LastFetched))
				}
				line := fmt.Sprintf("%s%s  %-4s  %s",
					indent,
					truncate(title, widthForCol(m.width, 48)),
					feedType,
					lastFetched,
				)
				var rendered string
				if i == m.displayCursor {
					rendered = selectedItemStyle.Render("> " + unreadCountStr(item.unread) + "  " + line)
				} else if item.unread > 0 {
					rendered = normalItemStyle.Render("  " + unreadCountStr(item.unread) + "  " + line)
				} else {
					rendered = dimmedStyle.Render("  " + unreadCountStr(item.unread) + "  " + line)
				}
				b.WriteString(rendered)
				b.WriteString("\n")
			}
		}

		if needsUp {
			b.WriteString(scrollStyle.Render(fmt.Sprintf("  ↑ %d more", start)))
			b.WriteString("\n")
		}
		if needsDown {
			b.WriteString(scrollStyle.Render(fmt.Sprintf("  ↓ %d more", len(m.displayItems)-end)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m model) entriesListView() string {
	var b strings.Builder

	title := "All Entries"
	if m.feed != nil && m.feed.Title != "" {
		title = m.feed.Title
	} else if m.feed == nil && m.filterFeedID != "" {
		for _, f := range m.feeds {
			if f.ID == m.filterFeedID {
				title = fmt.Sprintf("Filter: %s", f.Title)
				break
			}
		}
	} else if m.feed == nil {
		title = "All Entries"
	}
	filter := "unread"
	if m.showRead {
		filter = "all"
	}
	searchLabel := ""
	if m.searchQuery != "" {
		searchLabel = fmt.Sprintf(" — search: %q", m.searchQuery)
	}
	b.WriteString(headerStyle.Render(fmt.Sprintf(" < %s%s", title, searchLabel)))
	b.WriteString("\n")
	var allHints []*helpBinding
	allHints = append(allHints, entriesListHints...)
	feedSwitching := m.feed != nil || m.filterFeedID != ""
	if feedSwitching {
		allHints = append(allHints, &bindingPrevNext)
	}
	allHints = append(allHints, entriesListSuffixHints...)
	allHints = append(allHints, exitHints...)
	b.WriteString(helpStyle.Width(m.width - 4).Render(renderHintLine(allHints, map[string]string{"u": filter})))
	b.WriteString("\n\n")

	if len(m.entries) == 0 {
		b.WriteString(emptyStyle.Render("  No entries found."))
		b.WriteString("\n")
	} else {
		showLoadMore := len(m.entries) >= m.entryPageSize

		overhead := 4 // header + blank + blank + status
		if showLoadMore {
			overhead++
		}
		avail := m.height - overhead
		if avail < 1 {
			avail = 1
		}

		// Two-pass: compute window, check indicator lines needed, then adjust
		start, end := windowItems(len(m.entries), m.entryCursor, avail)
		needsUp := start > 0
		needsDown := end < len(m.entries)
		indicatorLines := 0
		if needsUp {
			indicatorLines++
		}
		if needsDown {
			indicatorLines++
		}
		avail -= indicatorLines
		if avail < 1 {
			avail = 1
		}
		start, end = windowItems(len(m.entries), m.entryCursor, avail)
		needsUp = start > 0
		needsDown = end < len(m.entries)

		for i := start; i < end; i++ {
			entry := m.entries[i]
			pub := "(no date)"
			if !entry.PublishedAt.IsZero() {
				pub = entry.PublishedAt.Format("2006-01-02 15:04")
			}

			eTitle := entry.Title
			if eTitle == "" {
				eTitle = "(no title)"
			}

			showingAllFeeds := m.feed == nil
			line := fmt.Sprintf("  %s  %s",
				pub,
				truncate(eTitle, widthForCol(m.width, 60)),
			)
			if showingAllFeeds {
				feedLabel := entry.FeedTitle
				if feedLabel == "" {
					feedLabel = "?"
				}
				line = fmt.Sprintf("  %s  [%s] %s",
					pub,
					feedLabel,
					truncate(eTitle, widthForCol(m.width, 50)),
				)
			}

			var rendered string
			if i == m.entryCursor {
				rendered = selectedItemStyle.Render("> " + line)
			} else if entry.Read {
				rendered = readItemStyle.Render("  " + line)
			} else {
				rendered = normalItemStyle.Render("  " + line)
			}
			b.WriteString(rendered)
			b.WriteString("\n")
		}

		if needsUp {
			b.WriteString(scrollStyle.Render(fmt.Sprintf("  ↑ %d more", start)))
			b.WriteString("\n")
		}
		if needsDown {
			b.WriteString(scrollStyle.Render(fmt.Sprintf("  ↓ %d more", len(m.entries)-end)))
			b.WriteString("\n")
		}

		if showLoadMore {
			b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingLoadMore}, map[string]string{"L": fmt.Sprintf("Load more (%d loaded)", len(m.entries))})))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m model) entryDetailView() string {
	var b strings.Builder

	title := "(no title)"
	if m.entry != nil && m.entry.Title != "" {
		title = m.entry.Title
	}
	b.WriteString(headerStyle.Render(fmt.Sprintf(" < %s", title)))
	b.WriteString("\n")
	hints := append([]*helpBinding{}, detailActionHints...)
	hints = append(hints, exitHints...)
	b.WriteString(helpStyle.Width(m.width - 4).Render(renderHintLine(hints, nil)))
	b.WriteString("\n\n")

	content := m.viewport.View()
	b.WriteString(content)
	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func entryDetailContent(m model) string {
	e := m.entry
	if e == nil {
		return ""
	}

	var b strings.Builder

	if !e.PublishedAt.IsZero() {
		b.WriteString(detailLabelStyle.Render("Published: "))
		b.WriteString(detailValueStyle.Render(e.PublishedAt.Format("2006-01-02 15:04")))
		b.WriteString("\n")
	}
	if e.Author != "" {
		b.WriteString(detailLabelStyle.Render("Author: "))
		b.WriteString(detailValueStyle.Render(e.Author))
		b.WriteString("\n")
	}
	if e.URL != "" {
		b.WriteString(detailLabelStyle.Render("Link: "))
		b.WriteString(detailValueStyle.Render(e.URL))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	body := e.Content
	if body == "" {
		body = e.Summary
	}
	if body != "" {
		b.WriteString(detailValueStyle.Render(body))
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) editFeedView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(" < Edit Feed"))
	b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingEnterSave, &bindingEscCancel, &bindingHintQuit}, nil)))
	b.WriteString("\n\n")

	b.WriteString(detailLabelStyle.Render("  Title:"))
	b.WriteString("\n\n")
	b.WriteString("  ")
	b.WriteString(m.editTitleInput.View())
	b.WriteString("\n\n")

	b.WriteString(detailLabelStyle.Render("  URL:"))
	b.WriteString("\n\n")
	b.WriteString("  ")
	b.WriteString(m.editURLInput.View())
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m model) addFeedView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(" < Add Feed"))
	b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingEnterAdd, &bindingHintBack, &bindingHintQuit}, nil)))
	b.WriteString("\n\n\n")

	b.WriteString(detailLabelStyle.Render("  Enter feed URL:"))
	b.WriteString("\n\n")
	b.WriteString("  ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

type helpBinding struct {
	key  string
	desc string
}

// Unique bindings — each {key, desc} pair defined once.
var (
	bindingUp             = helpBinding{"↑/k", "Move up"}
	bindingDown           = helpBinding{"↓/j", "Move down"}
	bindingPgUp           = helpBinding{"PgUp", "Page up"}
	bindingPgDn           = helpBinding{"PgDn", "Page down"}
	bindingHome           = helpBinding{"Home", "Go to first"}
	bindingEnd            = helpBinding{"End", "Go to last"}
	bindingEnterSel       = helpBinding{"Enter", "Select / Confirm"}
	bindingBack           = helpBinding{"Esc", "Back"}
	bindingHelp           = helpBinding{"?", "Help"}
	bindingQuit           = helpBinding{"q/Ctrl+C", "Quit"}
	bindingAllEntries     = helpBinding{"All Entries", "Shows entries from all feeds"}
	bindingAdd            = helpBinding{"a", "Add"}
	bindingExport         = helpBinding{"e", "Export"}
	bindingImport         = helpBinding{"i", "Import"}
	bindingFolder         = helpBinding{"g", "Folder"}
	bindingRename         = helpBinding{"R", "Rename"}
	bindingMove           = helpBinding{"m", "Move"}
	bindingDelete         = helpBinding{"d", "Delete"}
	bindingFeedFetch      = helpBinding{"f", "Fetch"}
	bindingToggleCollapse = helpBinding{"Enter/Space", "Toggle folder collapse"}
	bindingEnterDetail    = helpBinding{"Enter", "Open entry detail"}
	bindingFilter         = helpBinding{"f", "Filter"}
	bindingMarkRead       = helpBinding{"a", "Mark Read"}
	bindingAllRead        = helpBinding{"A", "All Read"}
	bindingEdit           = helpBinding{"E", "Edit"}
	bindingRefresh        = helpBinding{"r", "Refresh"}
	bindingSearch         = helpBinding{"s", "Search"}
	bindingLoadMore       = helpBinding{"L", "Load more"}
	bindingToggleRead     = helpBinding{"u", "Toggle read"}
	bindingUnread         = helpBinding{"M", "Unread"}
	bindingOpen           = helpBinding{"o", "Open"}
	bindingPrevNext       = helpBinding{"[/]", "Previous/next feed"}
	bindingScroll         = helpBinding{"↑/↓", "Scroll line by line"}
	bindingPage           = helpBinding{"PgUp/PgDn", "Scroll page by page"}
	bindingHintHelp       = helpBinding{"?", "Help"}
	bindingHintBack       = helpBinding{"Esc", "Back"}
	bindingHintQuit       = helpBinding{"q", "Quit"}
	bindingEnterSave      = helpBinding{"Enter", "Save"}
	bindingEnterAdd       = helpBinding{"Enter", "Add"}
	bindingEnterCreate    = helpBinding{"Enter", "Create"}
	bindingEnterRename    = helpBinding{"Enter", "Rename"}
	bindingEnterImport    = helpBinding{"Enter", "Import"}
	bindingEnterConfirm   = helpBinding{"Enter", "Confirm"}
	bindingEnterSearch    = helpBinding{"Enter", "Search"}
	bindingNumberSel      = helpBinding{"0-9", "Select"}
	bindingEscCancel      = helpBinding{"Esc", "Cancel"}
	bindingExportAll      = helpBinding{"a", "All"}
	bindingExportFolders  = helpBinding{"f", "Folders only"}
	bindingExportUngrouped= helpBinding{"u", "Ungrouped only"}
)

// Help view sections (referenced by pointer — no duplication).
var (
	navBindings = []*helpBinding{
		&bindingUp, &bindingDown, &bindingPgUp, &bindingPgDn,
		&bindingHome, &bindingEnd, &bindingEnterSel, &bindingBack,
	}
	globalBindings = []*helpBinding{&bindingHelp, &bindingBack, &bindingQuit}
	feedListBindings = []*helpBinding{
		&bindingAdd, &bindingExport, &bindingImport,
		&bindingFolder, &bindingRename, &bindingMove, &bindingDelete,
		&bindingEdit, &bindingFeedFetch, &bindingRefresh,
		&bindingToggleCollapse,
	}
	entriesListBindings = []*helpBinding{
		&bindingEnterDetail, &bindingToggleRead, &bindingFilter, &bindingSearch,
		&bindingMarkRead, &bindingAllRead, &bindingLoadMore, &bindingEdit,
		&bindingPrevNext, &bindingUnread, &bindingOpen, &bindingRefresh,
	}
	detailBindings = []*helpBinding{
		&bindingScroll, &bindingPage, &bindingUnread, &bindingOpen,
		&bindingRefresh,
	}
)

// On-screen hint slices (same bindings, no string duplication).
var (
	feedListHints = []*helpBinding{
		&bindingAdd, &bindingFolder, &bindingFeedFetch, &bindingRefresh,
		&bindingExport, &bindingImport,
		&bindingEdit, &bindingDelete, &bindingMove, &bindingRename,
	}
	entriesListHints = []*helpBinding{
		&bindingToggleRead, &bindingFilter, &bindingSearch,
		&bindingMarkRead, &bindingAllRead, &bindingLoadMore, &bindingEdit,
	}
	entriesListSuffixHints = []*helpBinding{
		&bindingUnread, &bindingOpen, &bindingRefresh,
	}
	detailActionHints = []*helpBinding{
		&bindingUnread, &bindingOpen, &bindingRefresh,
		&bindingScroll, &bindingPage,
	}
	exitHints = []*helpBinding{
		&bindingHintHelp, &bindingHintBack, &bindingHintQuit,
	}
)

func renderHintLine(bindings []*helpBinding, overrides map[string]string) string {
	var parts []string
	for _, b := range bindings {
		desc := b.desc
		if override, ok := overrides[b.key]; ok {
			desc = override
		}
		if desc == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("[%s] %s", b.key, desc))
	}
	return strings.Join(parts, "  ")
}

func renderHelpSection(title string, bindings []*helpBinding) []string {
	lines := []string{"  " + title}
	for _, b := range bindings {
		lines = append(lines, fmt.Sprintf("    %-12s %s", b.key, b.desc))
	}
	return lines
}

func (m model) helpView() string {
	nav := renderHelpSection("Navigation", navBindings)
	global := renderHelpSection("Global", globalBindings)

	var section []string
	switch m.prevScreen {
	case feedsListScreen:
		section = renderHelpSection("Feed List", feedListBindings)
	case entriesListScreen:
		section = renderHelpSection("Entries List", entriesListBindings)
	case entryDetailScreen:
		section = renderHelpSection("Entry Detail", detailBindings)
	default:
		section = renderHelpSection("Actions", feedListBindings)
	}

	help := strings.Join([]string{
		strings.Join(nav, "\n"),
		"",
		strings.Join(section, "\n"),
		"",
		strings.Join(global, "\n"),
	}, "\n") + "\n"

	boxWidth := m.width - 4
	if boxWidth > 60 {
		boxWidth = 60
	}
	if boxWidth < 20 {
		boxWidth = 20
	}
	box := helpBoxStyle.Width(boxWidth).Render(help)

	var b strings.Builder
	b.WriteString(headerStyle.Render(" < Help"))
	b.WriteString("\n\n")
	b.WriteString(box)
	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m model) statusBar() string {
	status := m.status
	if status == "" {
		if m.fetching {
			status = "Fetching all feeds..."
		} else if m.loading {
			status = "Loading..."
		} else {
			status = "Ready"
		}
	}
	if m.fetching || m.loading {
		status = m.spinner.View() + " " + status
	}

	if m.autoRefreshInterval > 0 && !m.fetching && status != "Ready" {
		status += fmt.Sprintf(" | %s", formatDurationRemaining(m.autoRefreshRemaining))
	} else if m.autoRefreshInterval > 0 && status == "Ready" {
		status = fmt.Sprintf("Ready — auto-refresh in %s", formatDurationRemaining(m.autoRefreshRemaining))
	}

	left := statusStyle.Render(status)

	var right string
	switch m.screen {
	case feedsListScreen:
		totalUnread := 0
		for _, n := range m.unreadCounts {
			totalUnread += n
		}
		r := fmt.Sprintf("%d unread · %d feeds", totalUnread, len(m.feeds))
		if len(m.folders) > 0 {
			r = fmt.Sprintf("%s · %d folders", r, len(m.folders))
		}
		right = statusStyle.Render(r)
	case entriesListScreen:
		unread := countUnread(m.entries)
		if m.showRead {
			right = statusStyle.Render(fmt.Sprintf("%d entries (%d unread)", len(m.entries), unread))
		} else {
			right = statusStyle.Render(fmt.Sprintf("%d unread", len(m.entries)))
		}
	case entryDetailScreen:
		if m.viewport.TotalLineCount() > 0 {
			percent := m.viewport.ScrollPercent()
			right = statusStyle.Render(fmt.Sprintf("%d%%", int(percent*100)))
		}
	}

	gap := m.width - lipglossWidth(left) - lipglossWidth(right) - 2
	if gap < 1 {
		gap = 1
	}

	return fmt.Sprintf("%s%s%s",
		left,
		dimmedStyle.Render(strings.Repeat(" ", gap)),
		right,
	)
}

func formatDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d min ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return fmt.Sprintf("%.0f days ago", d.Hours()/24)
	}
}

func formatDurationRemaining(d time.Duration) string {
	if d <= 0 {
		return "now"
	}
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:1]
	}
	return s[:n-1] + "…"
}

func lipglossWidth(s string) int {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 {
		return len(lines[len(lines)-1])
	}
	return len(s)
}

func widthForCol(totalWidth, max int) int {
	if totalWidth < 40 {
		return min(20, max)
	}
	return min(totalWidth-30, max)
}

func windowItems(total, cursor, capacity int) (start, end int) {
	if total <= capacity {
		return 0, total
	}
	half := capacity / 2
	start = cursor - half
	if start < 0 {
		start = 0
	}
	end = start + capacity
	if end > total {
		end = total
		start = total - capacity
	}
	return
}

func unreadCountStr(n int) string {
	if n > 99 {
		return "99+"
	}
	return fmt.Sprintf("%2d", n)
}

func countUnread(entries []*domain.Entry) int {
	n := 0
	for _, e := range entries {
		if !e.Read {
			n++
		}
	}
	return n
}

func (m model) folderCreateView() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(" < Create Folder"))
	b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingEnterCreate, &bindingHintBack, &bindingHintQuit}, nil)))
	b.WriteString("\n\n\n")
	b.WriteString(detailLabelStyle.Render("  Enter folder name:"))
	b.WriteString("\n\n")
	b.WriteString("  ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m model) folderRenameView() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(" < Rename Folder"))
	b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingEnterRename, &bindingHintBack, &bindingHintQuit}, nil)))
	b.WriteString("\n\n\n")
	b.WriteString(detailLabelStyle.Render("  New folder name:"))
	b.WriteString("\n\n")
	b.WriteString("  ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m model) folderPickView() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(" < Move Feed to Folder"))
	b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingNumberSel, &bindingEscCancel}, nil)))
	b.WriteString("\n\n")

	b.WriteString(detailLabelStyle.Render("  Select a folder (or 0 for no folder):"))
	b.WriteString("\n\n")

	b.WriteString("  0  (none)\n")
	for i, f := range m.folders {
		line := fmt.Sprintf("  %d  %s", i+1, f.Name)
		b.WriteString(normalItemStyle.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m model) importView() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(" < Import OPML"))
	b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingEnterImport, &bindingHintBack, &bindingHintQuit}, nil)))
	b.WriteString("\n\n\n")
	b.WriteString(detailLabelStyle.Render("  Enter path to OPML file:"))
	b.WriteString("\n\n")
	b.WriteString("  ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m model) importDryRunView() string {
	var b strings.Builder
	if m.loading {
		b.WriteString(headerStyle.Render(" < Importing..."))
		b.WriteString("\n\n\n\n")
		b.WriteString(centerStyle.Render(m.spinner.View() + " Importing feeds..."))
		b.WriteString("\n\n")
		b.WriteString(m.statusBar())
		return b.String()
	}
	b.WriteString(headerStyle.Render(" < Import Preview"))
	b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingEnterConfirm, &bindingEscCancel, &bindingHintQuit}, nil)))
	b.WriteString("\n\n")

	if len(m.importSpecs) == 0 {
		b.WriteString(emptyStyle.Render("  No feeds found in OPML file."))
		b.WriteString("\n")
	} else {
		byFolder := make(map[string][]opml.FeedSpec)
		var noFolder []opml.FeedSpec
		for _, s := range m.importSpecs {
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
			b.WriteString(folderHeaderStyle.Render(fmt.Sprintf("  %s (%d feeds)", name, len(feeds))))
			b.WriteString("\n")
			for _, f := range feeds {
				title := f.Title
				if title == "" {
					title = "(no title)"
				}
				b.WriteString(dimmedStyle.Render(fmt.Sprintf("    %s", title)))
				b.WriteString("\n")
				b.WriteString(helpStyle.Render(fmt.Sprintf("      %s", f.URL)))
				b.WriteString("\n")
			}
		}

		if len(noFolder) > 0 {
			b.WriteString(folderHeaderStyle.Render(fmt.Sprintf("  Uncategorized (%d feeds)", len(noFolder))))
			b.WriteString("\n")
			for _, f := range noFolder {
				title := f.Title
				if title == "" {
					title = "(no title)"
				}
				b.WriteString(dimmedStyle.Render(fmt.Sprintf("    %s", title)))
				b.WriteString("\n")
				b.WriteString(helpStyle.Render(fmt.Sprintf("      %s", f.URL)))
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m model) exportPickView() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(" < Export Feeds"))
	b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingExportAll, &bindingExportFolders, &bindingExportUngrouped, &bindingEscCancel}, nil)))
	b.WriteString("\n\n")

	b.WriteString(detailLabelStyle.Render("  Choose which feeds to export:"))
	b.WriteString("\n\n")
	b.WriteString(normalItemStyle.Render("  a  All feeds"))
	b.WriteString("\n")
	b.WriteString(normalItemStyle.Render("  f  Feeds in folders only"))
	b.WriteString("\n")
	b.WriteString(normalItemStyle.Render("  u  Ungrouped feeds only"))
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m model) feedPickView() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(" < Filter by Feed"))
	b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingEnterConfirm, &bindingEscCancel, &bindingHintQuit}, nil)))
	b.WriteString("\n\n")

	b.WriteString(detailLabelStyle.Render("  Enter feed number (0 for none):"))
	b.WriteString("\n\n")
	b.WriteString("  ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	for i, f := range m.feeds {
		title := f.Title
		if title == "" {
			title = "(no title)"
		}
		line := fmt.Sprintf("  %d  %s", i+1, title)
		b.WriteString(normalItemStyle.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m model) searchView() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(" < Search Entries"))
	b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingEnterSearch, &bindingHintBack, &bindingHintQuit}, nil)))
	b.WriteString("\n\n\n")
	b.WriteString(detailLabelStyle.Render("  Enter search query:"))
	b.WriteString("\n\n")
	b.WriteString("  ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func openURL(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	exec.Command(cmd, args...).Start()
}
