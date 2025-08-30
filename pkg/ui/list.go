package ui

import (
	"github.com/gdamore/tcell/v2"
)

func DrawList(s tcell.Screen, items []string, selected int) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	sel := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	for i, it := range items {
		st := style
		if i == selected {
			st = sel
		}
		for j, r := range it {
			s.SetContent(4+j, 2+i, r, nil, st)
		}
	}
	s.Show()
}
