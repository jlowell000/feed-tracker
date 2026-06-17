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
	}
	return ""
}

func (m model) feedsListView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(" Feed Tracker"))
	b.WriteString(helpStyle.Render("  [?] Help  [g] Folder  [q] Quit"))
	b.WriteString("\n\n")

	if len(m.displayItems) == 0 {
		b.WriteString(emptyStyle.Render("  No feeds tracked yet. Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		for i, item := range m.displayItems {
			indent := strings.Repeat("  ", item.depth)

			switch item.kind {
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
	}
	filter := "unread"
	if m.showRead {
		filter = "all"
	}
	b.WriteString(headerStyle.Render(fmt.Sprintf(" < %s", title)))
	b.WriteString(helpStyle.Render(fmt.Sprintf("  [u] %s  [Esc] Back  [q] Quit", filter)))
	b.WriteString("\n\n")

	if len(m.entries) == 0 {
		b.WriteString(emptyStyle.Render("  No entries found."))
		b.WriteString("\n")
	} else {
		for i, entry := range m.entries {
			pub := "(no date)"
			if !entry.PublishedAt.IsZero() {
				pub = entry.PublishedAt.Format("2006-01-02 15:04")
			}

			eTitle := entry.Title
			if eTitle == "" {
				eTitle = "(no title)"
			}

			line := fmt.Sprintf("  %s  %s",
				pub,
				truncate(eTitle, widthForCol(m.width, 60)),
			)

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
	b.WriteString(helpStyle.Render("  [Esc] Back  [M] Unread  [o] Open  [q] Quit"))
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

func (m model) addFeedView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(" < Add Feed"))
	b.WriteString(helpStyle.Render("  [Enter] Add  [Esc] Back  [q] Quit"))
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

func (m model) helpView() string {
	help := strings.Join([]string{
		"  Navigation",
		"    ↑/k         Move up",
		"    ↓/j         Move down",
		"    Enter       Select / Confirm",
		"    Esc         Back",
		"",
		"  Actions",
		"    a           Add a new feed",
		"    g           Create a folder",
		"    m           Move feed to folder",
		"    d           Delete folder",
		"    R           Rename folder",
		"    Enter/Space Toggle folder collapse",
		"    f           Fetch all feeds",
		"    r           Refresh current view",
		"    u           Toggle show read entries",
		"    o           Open entry URL in browser",
		"    M           Mark entry unread",
		"",
		"  Global",
		"    ?           Toggle this help",
		"    q/Ctrl+C    Quit",
	}, "\n") + "\n"

	box := helpBoxStyle.Render(help)

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
			status = m.spinner.View() + " Fetching all feeds..."
		} else if m.loading {
			status = m.spinner.View() + " Loading..."
		} else {
			status = "Ready"
		}
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
	b.WriteString(helpStyle.Render("  [Enter] Create  [Esc] Back  [q] Quit"))
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
	b.WriteString(helpStyle.Render("  [Enter] Rename  [Esc] Back  [q] Quit"))
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
	b.WriteString(helpStyle.Render("  [0-9] Select  [Esc] Cancel"))
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
