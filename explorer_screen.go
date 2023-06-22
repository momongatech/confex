package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ExplorerScreen struct {
	Panes         []Pane
	ActivePaneIdx int
	Parent        *EntryApp
	ScreenWidth   int
}

func NewExplorerScreen() *ExplorerScreen {
	wd, _ := os.Getwd()
	hostPane := NewPane("host", Host, wd)
	hostPane.ListDir()

	containerPane := NewPane("crunner", Container, "/")
	containerPane.ListDir()

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
			s.Parent.containerSelectionScreen.RefreshContainerList()
			return s.Parent.containerSelectionScreen, nil
		}
	case tea.WindowSizeMsg:
		s.ScreenWidth = msg.Width
	}
	return s, nil
}

func (s *ExplorerScreen) View() string {
	rows := ""

	tabItems := []string{}
	for i, p := range s.Panes {
		if i == s.ActivePaneIdx {
			tabItems = append(tabItems, lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				Align(lipgloss.Center).
				Padding(0, 2).
				UnsetBorderBottom().
				Render(p.Name))
		} else {
			tabItems = append(tabItems, lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder(), true).
				Align(lipgloss.Center).
				Padding(0, 2).
				Render(p.Name))
		}
	}

	tab := lipgloss.JoinHorizontal(lipgloss.Top, tabItems...)
	rows += tab
	rows += "\n"

	paneStyle := lipgloss.NewStyle().Width((s.ScreenWidth-4)/2).Border(lipgloss.RoundedBorder(), true).Height(20)
	rows += lipgloss.JoinHorizontal(lipgloss.Top,
		paneStyle.Render(s.Panes[0].RenderPane()),
		paneStyle.Render(s.Panes[1].RenderPane()),
	)
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
