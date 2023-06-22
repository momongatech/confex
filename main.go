package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	explorerScreen := NewExplorerScreen()
	containerSelectorScreen := NewContainerSelectorScreen()

	app := NewEntryApp(explorerScreen, containerSelectorScreen)
	program := tea.NewProgram(app, tea.WithAltScreen())
	program.Run()
}
