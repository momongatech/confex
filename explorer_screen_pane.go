package main

type PaneType int

const (
	Host PaneType = iota
	Container
)

type Pane struct {
	Name   string
	PType  PaneType
	CurIdx int
	Items  []string
}

func NewPane(paneName string, pType PaneType) Pane {
	return Pane{
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
	rows := p.Name + "\n"
	for i, item := range p.Items {
		ptrChar := "  "
		if i == p.CurIdx {
			ptrChar = "> "
		}
		rows += ptrChar + item + "\n"
	}
	return rows
}
