package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Back       key.Binding
	Add        key.Binding
	Export     key.Binding
	Import     key.Binding
	Fetch      key.Binding
	Refresh    key.Binding
	Open       key.Binding
	ToggleRead key.Binding
	MarkUnread   key.Binding
	CreateFolder key.Binding
	MoveFeed     key.Binding
	DeleteFolder key.Binding
	DeleteFeed   key.Binding
	RenameFolder key.Binding
	Help         key.Binding
	Quit         key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add feed"),
	),
	Export: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export OPML"),
	),
	Import: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "import OPML"),
	),
	Fetch: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "fetch all"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Open: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open url"),
	),
	ToggleRead: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "toggle read"),
	),
	MarkUnread: key.NewBinding(
		key.WithKeys("M"),
		key.WithHelp("M", "mark unread"),
	),
	CreateFolder: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "create folder"),
	),
	MoveFeed: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "move to folder"),
	),
	DeleteFolder: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete folder"),
	),
	DeleteFeed: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete feed"),
	),
	RenameFolder: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "rename folder"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
