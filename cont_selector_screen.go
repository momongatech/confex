package main

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

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

func (p *ContainerSelectorScreen) CursorInc(amount int) {
	p.CurIdx += amount
	if p.CurIdx < 0 {
		p.CurIdx = 0
	}
	if p.CurIdx > len(p.Containers)-1 {
		p.CurIdx = len(p.Containers) - 1
	}
}

func (a *ContainerSelectorScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// TOOD: better accessing container pane
			a.Parent.explorerScreen.Panes[1] = NewPane(a.Containers[a.CurIdx].ContainerName, Container, "/")
			a.Parent.explorerScreen.Panes[1].Parent = a.Parent.explorerScreen
			a.Parent.explorerScreen.Panes[1].PaneRows = a.Parent.explorerScreen.Panes[1].Parent.ScreenHeight - 9
			a.Parent.explorerScreen.Panes[1].ListDir()
			a.Parent.currentScreen = a.Parent.explorerScreen
			a.Parent.explorerScreen.ActivePaneIdx = 1
			return a.Parent.explorerScreen, nil
		case "up", "k":
			a.CursorInc(-1)
			return a, nil
		case "down", "j":
			a.CursorInc(+1)
			return a, nil
		}
	}
	return a, nil
}

func (a *ContainerSelectorScreen) View() string {
	a.Init()
	rows := ""
	for i, c := range a.Containers {
		ptrChar := "  "
		if i == a.CurIdx {
			ptrChar = "> "
		}
		rows += fmt.Sprintf("%s %s  %s\n", ptrChar, c.ContainerId[:8], c.ContainerName)
	}
	return rows
}

func (a *ContainerSelectorScreen) RefreshContainerList() {
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
