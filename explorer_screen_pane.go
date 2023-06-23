package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type PaneType int

const (
	Host PaneType = iota
	Container
)

type Pane struct {
	Cwd    string
	CurIdx int
	Items  []string
	Name   string
	PType  PaneType
	Parent *ExplorerScreen
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
	if p.CurIdx > len(p.Items)-1 {
		p.CurIdx = len(p.Items) - 1
	}
}

func (p *Pane) RenderPane() string {
	rows := ""

	isFocused := p.Parent.Panes[p.Parent.ActivePaneIdx] == p
	color := FocusColor
	if !isFocused {
		color = NoFocusColor
	}

	// Render panel title containing host name
	// and Cwd
	paneTitleStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(color).
		Width(p.Parent.ScreenWidth/2 - 4).
		Faint(!isFocused)

	hostnameStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1).Faint(!isFocused)
	rows += paneTitleStyle.Render(fmt.Sprintf("%s: %s", hostnameStyle.Render(p.Name), p.Cwd)) + "\n"

	for i, item := range p.Items {
		ptrChar := "   "
		if i == p.CurIdx {
			ptrChar = " > "
		}
		rows += lipgloss.NewStyle().Foreground(color).Render(ptrChar+" ☐ "+item) + "\n"
	}
	return rows
}

func (p *Pane) ListDir() error {
	items := []string{}
	if p.PType == Host {
		out, code := runCommand(fmt.Sprintf("ls -F %s", p.Cwd))
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
			Cmd:          []string{"ls", "-F", p.Cwd},
		}

		res, err := cli.ContainerExecCreate(ctx, p.Name, optionsCreate)
		if err != nil {
			return err
		}

		response, err := cli.ContainerExecAttach(context.Background(), res.ID, types.ExecStartCheck{})
		if err != nil {
			return err
		}
		defer response.Close()

		out, err := ioutil.ReadAll(response.Reader)
		if err != nil {
			return err
		}
		items = strings.Split(strings.TrimSpace(string(out)), "\n")
		p.Items = items
	}

	p.Items = items
	return nil
}
