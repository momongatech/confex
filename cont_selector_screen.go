package main

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var ContainerIdTextColor = lipgloss.AdaptiveColor{Light: "#6554AF", Dark: "#6554AF"}
var ContainerNameTextColor = lipgloss.AdaptiveColor{Light: "#E966A0", Dark: "#E966A0"}

type ContainerListItem struct {
	ContainerName string
	ContainerId   string
}

type ContainerSelectorScreen struct {
	Parent     *EntryApp
	Containers []ContainerListItem
	CurIdx     int
}

func NewContainerSelectorScreen() *ContainerSelectorScreen {
	return &ContainerSelectorScreen{}
}

func (a *ContainerSelectorScreen) Init() tea.Cmd {
	return nil
}

func (a *ContainerSelectorScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			a.Parent.explorerScreen.refreshDirContainerWithCwd(a.Containers[a.CurIdx].ContainerName, "/")
			a.Parent.explorerScreen.ActivePaneIdx = 1
			a.Parent.currentScreen = a.Parent.explorerScreen
			return a.Parent.explorerScreen, nil
		case "q":
			return a.Parent.explorerScreen, nil
		case "up", "k":
			a.cursorInc(-1)
			return a, nil
		case "down", "j":
			a.cursorInc(+1)
			return a, nil
		}
	case tea.WindowSizeMsg:
		a.Parent.ScreenHeight = msg.Height
		a.Parent.ScreenWidth = msg.Width
	}
	return a, nil
}

func (a *ContainerSelectorScreen) View() string {
	a.Init()

	rows := ""

	rows += lipgloss.NewStyle().
		Width(a.Parent.ScreenWidth).
		Align(lipgloss.Center).
		Bold(true).
		Render("Choose a container")
	rows += "\n\n"

	for i, c := range a.Containers {
		ptrChar := "   "
		if i == a.CurIdx {
			ptrChar = " ► "
		}

		containerIdStyle := lipgloss.NewStyle().Foreground(ContainerIdTextColor)
		containerNameStyle := lipgloss.NewStyle().Foreground(ContainerNameTextColor)
		rows += fmt.Sprintf("%s %s  %s\n", ptrChar, containerIdStyle.Render(c.ContainerId), containerNameStyle.Render(c.ContainerName))
	}

	boxLayout := lipgloss.NewStyle().
		Width(a.Parent.ScreenWidth-2).
		Height(a.Parent.ScreenHeight-4).
		Padding(1).
		Border(lipgloss.RoundedBorder(), true)

	rows = boxLayout.Render(rows)
	rows += "\n"

	cStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#64CCC5", Dark: "#64CCC5"})
	rows += fmt.Sprintf(
		"%s: Back\n",
		cStyle.Render("\"q\""))

	return rows
}

func (p *ContainerSelectorScreen) cursorInc(amount int) {
	p.CurIdx += amount
	if p.CurIdx < 0 {
		p.CurIdx = 0
	}
	if p.CurIdx > len(p.Containers)-1 {
		p.CurIdx = len(p.Containers) - 1
	}
}

func (a *ContainerSelectorScreen) refreshContainerList() {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	ctx := context.Background()
	if err != nil {
		panic(err)
	}
	cli.NegotiateAPIVersion(ctx)
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	a.Containers = []ContainerListItem{}
	for _, c := range containers {
		a.Containers = append(a.Containers, ContainerListItem{
			ContainerName: c.Names[0],
			ContainerId:   c.ID,
		})
	}
}
