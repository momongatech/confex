package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var FocusColor = lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#000000"}
var NoFocusColor = lipgloss.AdaptiveColor{Dark: "#555555", Light: "#eeeeee"}
var HintColor = lipgloss.AdaptiveColor{Light: "#64CCC5", Dark: "#64CCC5"}

type ExplorerScreen struct {
	Panes         []*Pane
	ActivePaneIdx int
	Parent        *EntryApp

	StatusMsg     string
	StatusMsgTick int
}

func NewExplorerScreen() *ExplorerScreen {
	wd, _ := os.Getwd()
	hostPane := NewPane("host", Host, wd)
	hostPane.ListDir()

	containerPane := NewPane("", Container, "/")
	containerPane.ListDir()

	s := &ExplorerScreen{
		Panes: []*Pane{
			hostPane,
			containerPane,
		},
		ActivePaneIdx: 0,
		StatusMsgTick: 0,
	}

	containerPane.Parent = s
	hostPane.Parent = s

	return s
}

// Messages are events that we respond to in our Update function. This
// particular one indicates that the timer has ticked.
type tickMsg time.Time

func (s *ExplorerScreen) tick() tea.Msg {
	s.StatusMsgTick += 1
	time.Sleep(time.Second)
	return tickMsg{}
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
		case "c":
			activePane := s.Panes[s.ActivePaneIdx]
			otherPane := s.Panes[len(s.Panes)-1-s.ActivePaneIdx]
			nCopied := 0
			if otherPane.Name != "" {
				nCopied = activePane.DoCopy(otherPane)
				hintStyle := lipgloss.NewStyle().Foreground(HintColor).Bold(true)
				s.StatusMsg = fmt.Sprintf("Copied %s files and/or folders from %s to %s",
					hintStyle.Render(fmt.Sprintf("%d", nCopied)),
					hintStyle.Render(activePane.Name),
					hintStyle.Render(otherPane.Name),
				)
				// Refresh the pane to reflect newly copied files
				otherPane.ListDir()
				return s, s.tick
			}
			return s, nil
		case "enter":
			activePane := s.Panes[s.ActivePaneIdx]
			if activePane.Name == "" {
				return s, nil
			}

			activePaneItem := activePane.Items[activePane.CurIdx]

			if activePane.PType == Container {
				if activePaneItem.ItemType == PaneItemTypeDirectory {
					s.ListDirContainer(activePane.Name, path.Join(activePane.Cwd, activePaneItem.Path))
				}
			} else {
				if activePaneItem.ItemType == PaneItemTypeDirectory {
					s.ListDirHost(path.Join(activePane.Cwd, activePaneItem.Path))
				}
			}
			return s, nil
		case "ctrl+c", "q":
			return s, tea.Quit
		case " ":
			activePane := s.Panes[s.ActivePaneIdx]
			if activePane.Items[activePane.CurIdx].Path != ".." {
				activePane.Items[activePane.CurIdx].Selected = !activePane.Items[activePane.CurIdx].Selected
			}
			return s, nil
		case "o":
			s.Parent.currentScreen = s.Parent.containerSelectionScreen
			s.Parent.containerSelectionScreen.RefreshContainerList()
			return s.Parent.containerSelectionScreen, nil
		}
	case tickMsg:
		if s.StatusMsgTick >= 2 {
			s.StatusMsgTick = 0
			s.StatusMsg = " "
			return s, nil
		}
		return s, s.tick
	case tea.WindowSizeMsg:
		s.Parent.ScreenWidth = msg.Width
		s.Parent.ScreenHeight = msg.Height
	default:
		return s, nil
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
	for _, p := range s.Panes {
		p.PaneRows = s.Parent.ScreenHeight - 9
	}
	rows := ""

	// Pane geometry styling
	paneStyle := lipgloss.NewStyle().
		Width((s.Parent.ScreenWidth-4)/2).
		Height(s.Parent.ScreenHeight-5).
		Border(lipgloss.RoundedBorder(), true)

	// Render two panes side by side
	rows += lipgloss.JoinHorizontal(lipgloss.Top,
		paneStyle.BorderForeground(getColor(s.ActivePaneIdx, 0)).Render(s.Panes[0].RenderPane()),
		paneStyle.BorderForeground(getColor(s.ActivePaneIdx, 1)).Render(s.Panes[1].RenderPane()),
	)
	rows += "\n"

	// Hint character styling
	cStyle := lipgloss.NewStyle().Foreground(HintColor).Bold(true)

	copyLabel := "Copy to container"
	if s.ActivePaneIdx == 1 {
		copyLabel = "Copy to host"
	}

	rows += fmt.Sprintf(
		"%s: %s | %s: Open container list | %s: Quit | %s: Select/deselect\n",
		cStyle.Render("\"c\""),
		copyLabel,
		cStyle.Render("\"o\""),
		cStyle.Render("\"q\""),
		cStyle.Render("space"),
	)

	rows += "\n"
	rows += s.StatusMsg

	return rows
}

func (s *ExplorerScreen) ListDirContainer(containerName string, pwd string) {
	s.Panes[1] = NewPane(strings.TrimPrefix(containerName, "/"), Container, pwd)
	s.Panes[1].Parent = s
	s.Panes[1].PaneRows = s.Panes[1].Parent.Parent.ScreenHeight - 9
	s.Panes[1].ListDir()
}

func (s *ExplorerScreen) ListDirHost(pwd string) {
	s.Panes[0] = NewPane("host", Host, pwd)
	s.Panes[0].Parent = s
	s.Panes[0].PaneRows = s.Panes[1].Parent.Parent.ScreenHeight - 9
	s.Panes[0].ListDir()
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
