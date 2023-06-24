package main

import tea "github.com/charmbracelet/bubbletea"

type EntryApp struct {
	explorerScreen           *ExplorerScreen
	containerSelectionScreen *ContainerSelectorScreen
	currentScreen            tea.Model
	ScreenWidth              int
	ScreenHeight             int
}

func NewEntryApp(explorerScreen *ExplorerScreen, containerSelectorScreen *ContainerSelectorScreen) *EntryApp {
	return &EntryApp{
		explorerScreen:           explorerScreen,
		containerSelectionScreen: containerSelectorScreen,
		currentScreen:            explorerScreen,
	}
}

func (a *EntryApp) Init() tea.Cmd {
	a.explorerScreen.Parent = a
	a.containerSelectionScreen.Parent = a
	return a.currentScreen.Init()
}

func (a *EntryApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.ScreenWidth = msg.Width
		a.ScreenHeight = msg.Height
	}
	return a.currentScreen.Update(msg)
}

func (a *EntryApp) View() string {
	return a.currentScreen.View()
}
