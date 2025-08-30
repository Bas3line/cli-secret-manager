package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

func DrawStatus(s tcell.Screen, msg string) {
	w, h := s.Size()
	style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	for i := 0; i < w; i++ {
		s.SetContent(i, h-1, ' ', nil, style)
	}
	for i, r := range msg {
		s.SetContent(2+i, h-1, r, nil, style)
	}
	s.Show()
	// auto clear after 5s
	go func() {
		time.Sleep(5 * time.Second)
		for i := 0; i < w; i++ {
			s.SetContent(i, h-1, ' ', nil, style)
		}
		s.Show()
	}()
}
