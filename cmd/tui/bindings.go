package main

import (
	"fmt"
	"strings"
)

type helpBinding struct {
	key  string
	desc string
}

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
	bindingTab            = helpBinding{"Tab", "Next field"}
	bindingShiftTab       = helpBinding{"Shift+Tab", "Prev field"}
	bindingEscCancel      = helpBinding{"Esc", "Cancel"}
	bindingExportAll      = helpBinding{"a", "All"}
	bindingExportFolders  = helpBinding{"f", "Folders only"}
	bindingExportUngrouped = helpBinding{"u", "Ungrouped only"}
)

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
