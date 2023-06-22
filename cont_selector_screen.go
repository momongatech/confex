package main

import tea "github.com/charmbracelet/bubbletea"

type ContainerSelectorScreen struct {
	Screens []tea.Model
	Parent  *EntryApp
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
			a.Parent.currentScreen = a.Parent.explorerScreen
			return a.Parent.explorerScreen, nil
		}
	}
	return a, nil
}

func (a *ContainerSelectorScreen) View() string {
	return "Conselect"
}
