package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00BFFF")).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D0D0D0"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#44FF44"))

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00BFFF")).
			Padding(1).
			Width(40)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888"))

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D0D0D0"))

	folderHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true)

	readItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	scrollStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	centerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).
			Bold(true).
			Align(lipgloss.Center).
			Width(60)
)
