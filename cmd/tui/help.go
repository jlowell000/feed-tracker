package main

import (
	"strings"
)

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
