package ui

import "github.com/gdamore/tcell/v2"

func DrawMenu(s tcell.Screen, items []string, selected int) {
	style := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
	selectedStyle := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorYellow)
	for i, item := range items {
		st := style
		if i == selected {
			st = selectedStyle
		}
		for j, r := range item {
			s.SetContent(2+j, 2+i, r, nil, st)
		}
	}
	s.Show()
}
