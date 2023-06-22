package main

import "fmt"

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
}

func NewPane(paneName string, pType PaneType, cwd string) Pane {
	return Pane{
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
	rows := fmt.Sprintf("%s:%s\n", p.Name, p.Cwd)
	for i, item := range p.Items {
		ptrChar := "  "
		if i == p.CurIdx {
			ptrChar = "> "
		}
		rows += ptrChar + item + "\n"
	}
	return rows
}
