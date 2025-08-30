package ui

import "github.com/gdamore/tcell/v2"

func DrawCenteredText(s tcell.Screen, x, y int, text string, style tcell.Style) {
	for i, r := range text {
		s.SetContent(x+i, y, r, nil, style)
	}
}
