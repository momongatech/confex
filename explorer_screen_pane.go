package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type PaneType int
type PaneItemType int

const (
	Host PaneType = iota
	Container
)

const (
	PaneItemTypeDirectory PaneItemType = iota
	PaneItemTypeFile
)

type PaneItem struct {
	Path     string
	ItemType PaneItemType
	Selected bool
}

type Pane struct {
	Cwd        string
	PaneOffset int
	PaneRows   int
	CurIdx     int
	Items      []PaneItem
	Name       string
	PType      PaneType
	Parent     *ExplorerScreen
}

func runCommand(cmd string) (string, int) {
	command := exec.Command("/bin/sh", "-c", cmd)

	// Set up output capture
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()

	// Extract exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
	}

	// Combine stdout and stderr
	output := stdout.String() + stderr.String()

	// Return combined output and exit code
	return output, exitCode
}

func NewPane(paneName string, pType PaneType, cwd string) *Pane {
	return &Pane{
		Cwd:   cwd,
		Name:  paneName,
		PType: pType,
	}
}

func (p *Pane) CursorInc(amount int) {
	p.CurIdx += amount

	if p.CurIdx < 0 {
		p.CurIdx = 0
	}

	if p.CurIdx < p.PaneOffset {
		p.PaneOffset = p.CurIdx
	}

	// if the cursor index is greater than the last visible row in the pane
	if p.CurIdx >= p.PaneOffset+p.PaneRows {
		p.PaneOffset = p.CurIdx - p.PaneRows + 1
	}

	// if the pane offset is greater than the difference between the length of items and the number of pane rows
	if p.PaneOffset > len(p.Items)-p.PaneRows {
		p.PaneOffset = max(0, len(p.Items)-p.PaneRows)
	}

	// if the cursor index is greater than the last item index
	if p.CurIdx >= len(p.Items) {
		p.CurIdx = max(0, len(p.Items)-1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (p *Pane) RenderPane() string {
	if p.Name == "" {
		return lipgloss.NewStyle().
			Width(p.Parent.Parent.ScreenWidth/2).
			Height(p.Parent.Parent.ScreenHeight/2).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Open a container")
	}

	rows := ""

	isFocused := p.Parent.Panes[p.Parent.ActivePaneIdx] == p

	color := FocusColor

	DirTextColor := lipgloss.AdaptiveColor{Light: "#6554AF", Dark: "#6554AF"}
	FileTextColor := lipgloss.AdaptiveColor{Light: "#E966A0", Dark: "#E966A0"}

	if !isFocused {
		color = NoFocusColor
		DirTextColor = NoFocusColor
		FileTextColor = NoFocusColor
	}

	// Render panel title containing host name
	// and Cwd
	paneTitleStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(color).
		Width(p.Parent.Parent.ScreenWidth/2 - 4).
		Faint(!isFocused)

	cwdStyle := lipgloss.NewStyle().Foreground(color)

	machineIcon := " 🏠 "
	if p.PType == Container {
		machineIcon = " 🐳 "
	}
	rows += paneTitleStyle.Render(cwdStyle.Render(fmt.Sprintf("%s %s: %s", machineIcon, strings.TrimPrefix(p.Name, "/"), p.Cwd))) + "\n"

	// Render listed items within pane viewport
	for i := p.PaneOffset; i < p.PaneOffset+p.PaneRows; i += 1 {
		if i < len(p.Items) {
			item := p.Items[i]
			ptrChar := "   "
			if i == p.CurIdx {
				ptrChar = " ► "
			}

			var pathStyle lipgloss.Style
			if item.ItemType == PaneItemTypeDirectory {
				pathStyle = lipgloss.NewStyle().Foreground(DirTextColor).Bold(true).Italic(true)
			} else {
				pathStyle = lipgloss.NewStyle().Foreground(FileTextColor)
			}

			checkBox := lipgloss.NewStyle().Faint(true).Render(" ☐ ")
			if item.Selected {
				checkBox = " ☑ "
			}
			if item.Path == ".." {
				checkBox = "   "
			}
			rows += lipgloss.NewStyle().Foreground(color).Render(ptrChar+checkBox) + pathStyle.Render(item.Path) + "\n"
		}
	}
	return rows
}

func (p *Pane) DoCopy(otherPane *Pane) int {
	fromPrefix := ""
	toPrefix := fmt.Sprintf("%s:", otherPane.Name)
	if p.PType == Container {
		fromPrefix = fmt.Sprintf("%s:", p.Name)
		toPrefix = ""
	}
	toDir := fmt.Sprintf("%s%s", toPrefix, otherPane.Cwd)
	nCopied := 0
	for _, i := range p.Items {
		if i.Selected {
			fromPath := fmt.Sprintf("%s%s", fromPrefix, path.Join(p.Cwd, i.Path))
			cpCmd := fmt.Sprintf("docker cp %s %s", fromPath, toDir)
			nCopied += 1
			runCommand(cpCmd)
		}
	}
	return nCopied
}

func (p *Pane) PopulateItems(items []string) {
	p.Items = []PaneItem{}
	p.Items = append(p.Items, PaneItem{Path: "..", ItemType: PaneItemTypeDirectory})

	for _, i := range items {
		item := PaneItem{}
		item.Path = i
		if strings.HasSuffix(i, "/") {
			item.ItemType = PaneItemType(PaneItemTypeDirectory)
		} else {
			item.ItemType = PaneItemType(PaneItemTypeFile)
		}
		p.Items = append(p.Items, item)
	}
}

func (p *Pane) ListDir() error {
	items := []string{}
	if p.PType == Host {
		out, code := runCommand(fmt.Sprintf("ls -p %s", p.Cwd))
		if code != 0 {
			return errors.New("`ls` command on host failed")
		}
		out = strings.TrimSpace(out)
		items = strings.Split(out, "\n")
	} else if p.PType == Container {
		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			return err
		}

		ctx := context.Background()
		cli.NegotiateAPIVersion(ctx)
		optionsCreate := types.ExecConfig{
			AttachStdout: true,
			AttachStderr: true,
			Cmd:          []string{"ls", "-p", p.Cwd},
		}

		res, err := cli.ContainerExecCreate(ctx, p.Name, optionsCreate)
		if err != nil {
			return err
		}

		response, err := cli.ContainerExecAttach(context.Background(), res.ID, types.ExecStartCheck{Tty: true})
		if err != nil {
			return err
		}
		defer response.Close()

		out, err := ioutil.ReadAll(response.Reader)
		if err != nil {
			return err
		}
		items = strings.Split(strings.TrimSpace(string(out)), "\n")
	}

	p.PopulateItems(items)
	return nil
}
