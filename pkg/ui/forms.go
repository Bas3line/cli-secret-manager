package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

type Field struct {
	Label  string
	Value  string
	Masked bool
	Width  int
}

// PromptForm renders a simple form and returns values map when submitted or cancelled.
func PromptForm(s tcell.Screen, title string, fields []Field) (map[string]string, bool) {
	styleTitle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue)
	styleLabel := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	styleInput := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	active := 0

	render := func() {
		s.Clear()
		// title
		for i, r := range title {
			s.SetContent(2+i, 1, r, nil, styleTitle)
		}
		// fields
		for i := range fields {
			label := fields[i].Label + ":"
			for j, r := range label {
				s.SetContent(2+j, 3+i*2, r, nil, styleLabel)
			}
			// input box
			val := fields[i].Value
			if fields[i].Masked {
				val = strings.Repeat("*", len(val))
			}
			for j := 0; j < fields[i].Width; j++ {
				ch := ' '
				if j < len(val) {
					ch = rune(val[j])
				}
				st := styleInput
				if i == active {
					st = st.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
				}
				s.SetContent(15+j, 3+i*2, ch, nil, st)
			}
		}
		// footer
		hint := "Enter=Submit  Esc=Cancel  Tab=Next"
		for i, r := range hint {
			s.SetContent(2+i, 3+len(fields)*2, r, nil, styleLabel)
		}
		s.Show()
	}

	render()

	for {
		e := s.PollEvent()
		switch ev := e.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEsc {
				return nil, true
			}
			if ev.Key() == tcell.KeyTAB || ev.Key() == tcell.KeyRight {
				active = (active + 1) % len(fields)
				render()
				continue
			}
			if ev.Key() == tcell.KeyBacktab || ev.Key() == tcell.KeyLeft {
				active = (active - 1 + len(fields)) % len(fields)
				render()
				continue
			}
			if ev.Key() == tcell.KeyEnter {
				// collect values
				out := make(map[string]string)
				for i := range fields {
					out[fields[i].Label] = fields[i].Value
				}
				return out, false
			}
			if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				if len(fields[active].Value) > 0 {
					fields[active].Value = fields[active].Value[:len(fields[active].Value)-1]
					render()
				}
				continue
			}
			if ev.Rune() != 0 {
				ch := ev.Rune()
				fields[active].Value = fields[active].Value + string(ch)
				render()
			}
		}
	}
}
