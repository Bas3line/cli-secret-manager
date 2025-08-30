package app

import (
	"fmt"
	"net/http"
	"time"

	"sm-cli/pkg/api"
	"sm-cli/pkg/ui"

	"github.com/gdamore/tcell/v2"
)

var backendURL = "http://localhost:8080"

func Run() error {
	s, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("failed to create tcell screen: %w", err)
	}
	if err := s.Init(); err != nil {
		return fmt.Errorf("failed to init screen: %w", err)
	}
	defer s.Fini()

	s.Clear()
	w, h := s.Size()

	u := ui.New(s)
	u.DrawSplash(w, h)

	// Non-blocking health check
	go func() {
		client := http.Client{Timeout: 3 * time.Second}
		resp, _ := client.Get(backendURL + "/health")
		if resp != nil {
			resp.Body.Close()
		}
	}()

	selected := 0
	// ensure selected lands on a selectable item
	menu, selFlags := u.MenuOptions()
	if len(menu) > 0 {
		// clamp selected
		if selected >= len(menu) {
			selected = 0
		}
		if !selFlags[selected] {
			// move forward to first selectable
			for i := 0; i < len(menu); i++ {
				if selFlags[i] {
					selected = i
					break
				}
			}
		}
	}

	u.RenderMainMenu(selected)

	for {
		// refresh menu & flags each loop (in case login state changed)
		menu, selFlags = u.MenuOptions()
		if len(menu) == 0 {
			// nothing to do, just continue
			continue
		}
		// clamp selected to menu size and ensure it's selectable
		if selected >= len(menu) {
			selected = 0
		}
		if !selFlags[selected] {
			// find next selectable forward, if none, try backward
			found := -1
			for i := 0; i < len(menu); i++ {
				if selFlags[i] {
					found = i
					break
				}
			}
			if found >= 0 {
				selected = found
			}
		}

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			// redraw on resize
			w, _ = s.Size()
			u.RenderMainMenu(selected)
		case *tcell.EventKey:
			// navigation
			swit := ev.Key()
			switch swit {
			case tcell.KeyUp:
				// move up to previous selectable
				for {
					selected = (selected - 1 + len(menu)) % len(menu)
					if selFlags[selected] {
						break
					}
				}
				u.RenderMainMenu(selected)
				continue
			case tcell.KeyDown:
				// move down to next selectable
				for {
					selected = (selected + 1) % len(menu)
					if selFlags[selected] {
						break
					}
				}
				u.RenderMainMenu(selected)
				continue
			case tcell.KeyEnter:
				// activate
				sel := menu[selected]
				switch sel {
				case "Login":
					if !api.HasToken() {
						u.ShowLogin()
					}
				case "Secrets":
					if selFlags[selected] {
						u.ShowSecretsList(1)
					} else {
						// disabled: show warning
						u.ShowDisabledWarning(sel)
					}
				case "Help":
					u.ShowHelp()
				case "Quit":
					return nil
				}
				// after action, re-render menu
				u.RenderMainMenu(selected)
				continue
			case tcell.KeyCtrlC, tcell.KeyEscape:
				return nil
			case tcell.KeyRune:
				// handled below
			}

			// rune shortcuts
			r := ev.Rune()
			switch r {
			case 'q', 'Q':
				return nil
			case 'l', 'L':
				if !api.HasToken() {
					u.ShowLogin()
				}
			case 'h', 'H':
				u.ShowHelp()
			case 'j', 'J':
				// down
				for {
					selected = (selected + 1) % len(menu)
					if selFlags[selected] {
						break
					}
				}
				u.RenderMainMenu(selected)
			case 'k', 'K':
				// up
				for {
					selected = (selected - 1 + len(menu)) % len(menu)
					if selFlags[selected] {
						break
					}
				}
				u.RenderMainMenu(selected)
			}
		}
	}
}
