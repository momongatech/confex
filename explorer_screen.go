package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type ExplorerScreen struct {
	Panes         []Pane
	ActivePaneIdx int
	Parent        *EntryApp
}

func NewExplorerScreen() *ExplorerScreen {
	wd, _ := os.Getwd()
	hostPane := NewPane("host", Host, wd)
	hostPane.Items = []string{
		"usr", "lib",
	}

	containerPane := NewPane("ubuntu", Container, "/")
	containerPane.Items = []string{
		"usr", "root", "var",
	}

	return &ExplorerScreen{
		Panes: []Pane{
			hostPane,
			containerPane,
		},
		ActivePaneIdx: 0,
	}
}

//// Bubbletea standard methods

func (s *ExplorerScreen) Init() tea.Cmd {
	return nil
}

func (s *ExplorerScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			s.Panes[s.ActivePaneIdx].CursorInc(-1)
			return s, nil
		case "down", "j":
			s.Panes[s.ActivePaneIdx].CursorInc(+1)
			return s, nil
		case "left", "h":
			s.CursorInc(-1)
			return s, nil
		case "right", "l":
			s.CursorInc(+1)
			return s, nil
		case "ctrl+c", "q":
			return s, tea.Quit
		case "o":
			s.Parent.currentScreen = s.Parent.containerSelectionScreen
			return s.Parent.containerSelectionScreen, nil
		}
	}
	return s, nil
}

func (s *ExplorerScreen) View() string {
	rows := ""
	for i, p := range s.Panes {
		rows += p.Name
		if i < len(s.Panes)-1 {
			rows += " | "
		}
	}

	rows += "\n"

	activePane := s.Panes[s.ActivePaneIdx]
	rows += activePane.RenderPane()
	return rows
}

//// App-specific methods

// Traverse over opened panes, jumping with specified amount.
func (s *ExplorerScreen) CursorInc(amount int) {
	s.ActivePaneIdx += amount
	if s.ActivePaneIdx < 0 {
		s.ActivePaneIdx = 0
	}
	if s.ActivePaneIdx > len(s.Panes)-1 {
		s.ActivePaneIdx = len(s.Panes) - 1
	}
}
