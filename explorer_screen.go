package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var FocusColor = lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#000000"}
var NoFocusColor = lipgloss.AdaptiveColor{Dark: "#555555", Light: "#eeeeee"}

type ExplorerScreen struct {
	Panes         []*Pane
	ActivePaneIdx int
	Parent        *EntryApp
	ScreenWidth   int
	ScreenHeight  int
}

func NewExplorerScreen() *ExplorerScreen {
	wd, _ := os.Getwd()
	hostPane := NewPane("host", Host, wd)
	hostPane.ListDir()

	containerPane := NewPane("crunner", Container, "/")
	containerPane.ListDir()

	s := &ExplorerScreen{
		Panes: []*Pane{
			hostPane,
			containerPane,
		},
		ActivePaneIdx: 0,
	}

	containerPane.Parent = s
	hostPane.Parent = s

	return s
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
		s.ScreenHeight = msg.Height
		for _, p := range s.Panes {
			p.PaneRows = s.ScreenHeight - 5
		}
	}
	return s, nil
}

func getColor(activeIdx int, currIdx int) lipgloss.AdaptiveColor {
	if activeIdx == currIdx {
		return FocusColor
	}
	return NoFocusColor
}

func (s *ExplorerScreen) View() string {
	rows := ""

	paneStyle := lipgloss.NewStyle().
		Width((s.ScreenWidth-4)/2).
		Height(s.ScreenHeight-5).
		Border(lipgloss.RoundedBorder(), true)

	rows += lipgloss.JoinHorizontal(lipgloss.Top,
		paneStyle.BorderForeground(getColor(s.ActivePaneIdx, 0)).Render(s.Panes[0].RenderPane()),
		paneStyle.BorderForeground(getColor(s.ActivePaneIdx, 1)).Render(s.Panes[1].RenderPane()),
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
