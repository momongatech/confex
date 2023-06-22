package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"momonga.tech/confex/explorer"
)

type EntryApp struct {
	Screens []tea.Model
}

func main() {
	hostPane := explorer.NewPane("host", explorer.Host)
	hostPane.Items = []string{
		"usr", "lib",
	}

	containerPane := explorer.NewPane("ubuntu", explorer.Container)
	containerPane.Items = []string{
		"usr", "root", "var",
	}

	app := &explorer.ExplorerScreen{
		Panes: []explorer.Pane{
			hostPane,
			containerPane,
		},
		ActivePaneIdx: 0,
	}

	program := tea.NewProgram(app, tea.WithAltScreen())
	program.Run()
}
