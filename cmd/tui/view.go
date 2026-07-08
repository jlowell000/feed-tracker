package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/jlowell000/feed-tracker/internal/domain"
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
		avail := m.height - 3 - m.statusBarHeight()
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

		overhead := 3 + m.statusBarHeight()
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

			star := "  "
			if entry.Starred {
				star = "★ "
			}

			showingAllFeeds := m.feed == nil
			line := fmt.Sprintf("  %s  %s%s",
				pub,
				star,
				truncate(eTitle, widthForCol(m.width, 60)),
			)
			if showingAllFeeds {
				feedLabel := entry.FeedTitle
				if feedLabel == "" {
					feedLabel = "?"
				}
				line = fmt.Sprintf("  %s  [%s] %s%s",
					pub,
					feedLabel,
					star,
					truncate(eTitle, widthForCol(m.width, 50)),
				)
			}

			var rendered string
			if i == m.entryCursor {
				rendered = selectedItemStyle.Render("> " + line)
			} else if entry.Starred {
				rendered = starredItemStyle.Render("  " + line)
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
	starLabel := ""
	if e.Starred {
		starLabel = "★"
	}
	b.WriteString(detailLabelStyle.Render("Starred: "))
	b.WriteString(detailValueStyle.Render(starLabel))
	b.WriteString("\n")
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

func (m model) statusBarLeftText() string {
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

	if !m.lastFetchTime.IsZero() && !m.fetching && !m.loading {
		status += fmt.Sprintf(" — last fetch %s |", m.lastFetchTime.Format("15:04:05"))
	}

	return status
}

func (m model) statusBarRightText() string {
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
		right = r
	case entriesListScreen:
		unread := countUnread(m.entries)
		filterLabel := ""
		if m.starredFilter {
			filterLabel = " ★"
		}
		if m.showRead {
			right = fmt.Sprintf("%d entries (%d unread)%s", len(m.entries), unread, filterLabel)
		} else {
			right = fmt.Sprintf("%d unread%s", len(m.entries), filterLabel)
		}
	case entryDetailScreen:
		if m.viewport.TotalLineCount() > 0 {
			percent := m.viewport.ScrollPercent()
			right = fmt.Sprintf("%d%%", int(percent*100))
		}
	}
	return right
}

func (m model) statusBarHeight() int {
	lw := lipglossWidth(statusStyle.Render(m.statusBarLeftText()))
	rw := lipglossWidth(statusStyle.Render(m.statusBarRightText()))
	if lw+rw+2 > m.width {
		return 2
	}
	return 1
}

func (m model) statusBar() string {
	left := statusStyle.Render(m.statusBarLeftText())
	right := m.statusBarRightText()
	if right == "" {
		return left
	}
	rw := lipglossWidth(statusStyle.Render(right))
	lw := lipglossWidth(left)
	if lw+rw+2 > m.width {
		return left + "\n" + statusStyle.Render(right)
	}
	gap := m.width - lw - rw - 2
	if gap < 1 {
		gap = 1
	}
	return fmt.Sprintf("%s%s%s",
		left,
		dimmedStyle.Render(strings.Repeat(" ", gap)),
		statusStyle.Render(right),
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
