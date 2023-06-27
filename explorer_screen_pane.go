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

// Run shell command via `sh`
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

	// Combine stdout and stderr as the returned execution result
	output := stdout.String() + stderr.String()
	return output, exitCode
}

func NewPane(paneName string, pType PaneType, cwd string) *Pane {
	return &Pane{
		Cwd:   cwd,
		Name:  paneName,
		PType: pType,
	}
}

// Move cursor in current pane, up or down
func (p *Pane) cursorInc(amount int) {
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

func (p *Pane) RenderPane() string {
	rows := ""

	// When there's no container opened, just print a simple
	// instruction for users
	if p.Name == "" {
		return lipgloss.NewStyle().
			Width(p.Parent.Parent.ScreenWidth/2).
			Height(p.Parent.Parent.ScreenHeight/2).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Open a container")
	}

	// Check if current panel is active, then Determine colors based on
	// whether/not this panel is active
	isFocused := p.Parent.Panes[p.Parent.ActivePaneIdx] == p

	foreColor := FocusColor
	DirTextColor := lipgloss.AdaptiveColor{Light: "#6554AF", Dark: "#6554AF"}
	FileTextColor := lipgloss.AdaptiveColor{Light: "#E966A0", Dark: "#E966A0"}
	if !isFocused {
		foreColor = NoFocusColor
		DirTextColor = NoFocusColor
		FileTextColor = NoFocusColor
	}

	// Render panel title containing host name and Cwd
	paneTitleStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(foreColor).
		Width(p.Parent.Parent.ScreenWidth/2 - 4).
		Faint(!isFocused)

	cwdStyle := lipgloss.NewStyle().Foreground(foreColor)

	// Machine icon will be rendered at cwd
	machineIcon := " 🏠 "
	if p.PType == Container {
		machineIcon = " 🐳 "
	}
	rows += paneTitleStyle.Render(
		cwdStyle.Render(fmt.Sprintf("%s %s: %s", machineIcon, strings.TrimPrefix(p.Name, "/"), p.Cwd)),
	) + "\n"

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
			rows += lipgloss.NewStyle().Foreground(foreColor).Render(ptrChar+checkBox) + pathStyle.Render(item.Path) + "\n"
		}
	}
	return rows
}

//// App-specific methods

// Perform shell script for copying selected items in the current pane to the
// target (in other pane's Cwd)
func (p *Pane) executeFileAndDirCopy(otherPane *Pane) int {
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
			// Construct a string of copy command, e.g., `docker cp prefix:source-path prefix:target-path`
			// For host machine, the prefix is an empty string, while for container, the prefix is `container-name:`
			fromPath := fmt.Sprintf("%s%s", fromPrefix, path.Join(p.Cwd, i.Path))
			cpCmd := fmt.Sprintf("docker cp %s %s", fromPath, toDir)

			// Execute the copy command
			runCommand(cpCmd)
			nCopied += 1
		}
	}
	return nCopied
}

// Convert a slice of string (file names as the result of executing `ls -p`) into
// a slice of PaneItems, and assign it to p.Items. The conversion aims to further
// extract information of a filename (e.g., whether it is a file, directory, etc)
func (p *Pane) populateItems(items []string) {
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

// List Cwd when current panel is a host machine
func (p *Pane) listDirHost() ([]string, error) {
	out, code := runCommand(fmt.Sprintf("ls -p %s", p.Cwd))
	if code != 0 {
		return nil, errors.New("`ls` command on host failed")
	}
	out = strings.TrimSpace(out)
	items := strings.Split(out, "\n")
	return items, nil
}

// List Cwd when current panel is a container
func (p *Pane) listDirContainer() ([]string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	// Prepare docker cli client and the command to execute, i.e., ls
	ctx := context.Background()
	cli.NegotiateAPIVersion(ctx)
	optionsCreate := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"ls", "-p", p.Cwd},
	}

	// Execute the ls command in tty mode
	res, err := cli.ContainerExecCreate(ctx, p.Name, optionsCreate)
	if err != nil {
		return nil, err
	}
	response, err := cli.ContainerExecAttach(context.Background(), res.ID, types.ExecStartCheck{Tty: true})
	if err != nil {
		return nil, err
	}
	defer response.Close()

	// Read all the resulted ls command execution
	out, err := ioutil.ReadAll(response.Reader)
	if err != nil {
		return nil, err
	}
	items := strings.Split(strings.TrimSpace(string(out)), "\n")
	return items, nil
}

// List directory contents
func (p *Pane) listDir() error {
	if p.PType == Host {
		items, err := p.listDirHost()
		if err != nil {
			return err
		}
		p.populateItems(items)
		return nil
	} else {
		items, err := p.listDirContainer()
		if err != nil {
			return err
		}
		p.populateItems(items)
		return nil
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
