package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jlowell000/feed-tracker/internal/opml"
)

func (m model) editFeedView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(" < Edit Feed"))
	b.WriteString(helpStyle.Render("  " + renderHintLine([]*helpBinding{&bindingTab, &bindingEnterSave, &bindingEscCancel, &bindingHintQuit}, nil)))
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
	b.WriteString("\n\n")

	b.WriteString(detailLabelStyle.Render("  Max Age:"))
	b.WriteString("\n\n")
	b.WriteString("  ")
	b.WriteString(m.editMaxAgeInput.View())
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
