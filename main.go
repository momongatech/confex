package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// The program has two main screens. Define them as follows.
	explorerScreen := NewExplorerScreen()
	containerSelectorScreen := NewContainerSelectorScreen()

	// All screens implement bubbletea.Model interface, so we can
	// build a tree-like structure to describe hierarchy.
	//
	// Note that panes (host pane and container pane) are just another struct.
	// They don't implement bubbletea.Model interface.
	//
	//	┏━━━━━━━━━━━━━━━━┓
	//	┃  Root Program  ┃
	//	┗━━━━━━━━━━━━━━━━┛                                        ┌────────────────┐
	//			 │                                   ┌───▶│   Host pane    │
	//			 │   ┌───────────────────────────┐   │    └────────────────┘
	//			 ├──▶│   File Explorer Screen    │───┤
	//			 │   └───────────────────────────┘   │    ┌────────────────┐
	//			 │                                   └───▶│ Container pane │
	//			 │   ┌───────────────────────────┐        └────────────────┘
	//			 └──▶│Container Selection Screen │
	//			     └───────────────────────────┘

	app := NewEntryApp(explorerScreen, containerSelectorScreen)
	program := tea.NewProgram(app, tea.WithAltScreen())
	program.Run()
}
