package ui

import (
	"strings"
	"unicode/utf8"

	"sm-cli/pkg/api"

	"github.com/gdamore/tcell/v2"
)

const logo = `
                                       __
   ________  _______  __________  ____/ /
  / ___/ _ \/ ___/ / / / ___/ _ \/ __  / 
 (__  )  __/ /__/ /_/ / /  /  __/ /_/ /  
/____/\___/\___/\__,_/_/   \___/\__,_/   
                                         `

const compactLogo = "Secrets Vault"

const tagline = "the best way to store your secrets instead of forgetting them."

type UI struct {
	s tcell.Screen
}

func New(s tcell.Screen) *UI {
	return &UI{s: s}
}

func (u *UI) DrawSplash(w, h int) {
	// show static logo and menu
	u.RenderMainMenu(0)
}

func (u *UI) drawText(x, y int, str string, st tcell.Style) {
	cx := x
	for _, r := range str {
		u.s.SetContent(cx, y, r, nil, st)
		cx++
	}
}

func (u *UI) RenderMainMenu(selected int) {
	u.s.Clear()
	w, h := u.s.Size()

	// prepare logo lines
	fullLines := strings.Split(strings.Trim(logo, "\n"), "\n")
	maxLine := 0
	for _, l := range fullLines {
		rl := utf8.RuneCountInString(l)
		if rl > maxLine {
			maxLine = rl
		}
	}

	// decide when to show full ASCII logo vs compact
	paddingX := 6
	paddingY := 2
	minMenuSpace := 4
	showFull := (w >= maxLine+minMenuSpace+paddingX*2) && maxLine > 0

	var logoLines []string
	if showFull {
		logoLines = fullLines
	} else {
		logoLines = []string{compactLogo}
	}

	// get menu items and selectable flags from MenuOptions
	menu, selFlags := u.MenuOptions()

	// compute widths
	logoWidth := 0
	for _, l := range logoLines {
		rl := utf8.RuneCountInString(l)
		if rl > logoWidth {
			logoWidth = rl
		}
	}

	// account for focus arrow + space (use rune counts)
	menuWidth := 0
	for _, it := range menu {
		// arrow + space + label (rune counts)
		lineLen := 2 + utf8.RuneCountInString(it)
		if lineLen > menuWidth {
			menuWidth = lineLen
		}
	}

	contentWidth := logoWidth
	if menuWidth > contentWidth {
		contentWidth = menuWidth
	}

	blockWidth := contentWidth + paddingX*2

	// compute height: top padding + logo + gap + tagline + gap + menu + bottom padding
	gap := 1
	itemSpacing := 0 // no blank lines between menu items (compact list)
	totalHeight := paddingY + len(logoLines) + gap + 1 + gap + len(menu)*(1+itemSpacing) + paddingY

	startX := 0
	if w > blockWidth {
		startX = (w - blockWidth) / 2
	}
	startY := 0
	if h > totalHeight {
		startY = (h - totalHeight) / 2
	}

	// draw card background (clean area)
	bgStyle := tcell.StyleDefault
	for y := 0; y < totalHeight; y++ {
		for x := 0; x < blockWidth; x++ {
			u.s.SetContent(startX+x, startY+y, ' ', nil, bgStyle)
		}
	}

	// render logo (centered inside content area)
	logoStartX := startX + paddingX
	curY := startY + paddingY
	titleStyle := tcell.StyleDefault.Bold(true)
	for i, line := range logoLines {
		lineX := logoStartX + (contentWidth-utf8.RuneCountInString(line))/2
		u.drawText(lineX, curY+i, line, titleStyle)
	}
	curY += len(logoLines)

	// tagline below logo
	curY += gap
	tagX := logoStartX + (contentWidth-utf8.RuneCountInString(tagline))/2
	tagStyle := tcell.StyleDefault
	u.drawText(tagX, curY, tagline, tagStyle)
	curY += 1

	// small spacer before menu
	curY += gap

	// render menu: each item is a centered button sized to the label with inner padding
	innerPad := 6 // increased spaces padding inside each button (left/right) for roomy white bar
	// compute maximum label length (without arrow) using rune counts
	maxLabelLen := 0
	for _, it := range menu {
		rlen := utf8.RuneCountInString(it)
		if rlen > maxLabelLen {
			maxLabelLen = rlen
		}
	}
	// base button width (label + two chars for possible arrow + inner padding)
	baseButtonWidth := maxLabelLen + 2 + innerPad*2
	// cap button width to contentWidth so it doesn't overflow
	if baseButtonWidth > contentWidth {
		baseButtonWidth = contentWidth
	}

	// render each menu entry with compact spacing and perfect centering between markers
	for i, it := range menu {
		barY := curY + i*(1+itemSpacing)
		// content area start
		contentX := logoStartX

		// draw background across content area for consistent look
		for x := 0; x < contentWidth; x++ {
			u.s.SetContent(contentX+x, barY, ' ', nil, tcell.StyleDefault)
		}

		// prepare label (no leading spaces) and clamp by rune count
		label := it
		// reserve space for markers and internal gaps
		markerGap := 2   // space between marker and text
		markerWidth := 1 // markers are single-run characters
		minInner := 1
		innerAvailable := contentWidth - (markerWidth*2 + markerGap*2) - innerPad*2
		if innerAvailable < minInner {
			innerAvailable = minInner
		}
		if utf8.RuneCountInString(label) > innerAvailable {
			r := []rune(label)
			label = string(r[:innerAvailable])
		}

		// compute marker positions
		leftMarkerX := contentX + innerPad/2
		rightMarkerX := contentX + contentWidth - 1 - innerPad/2
		// compute inner bounds where label should be centered
		innerStart := leftMarkerX + markerGap + markerWidth
		innerEnd := rightMarkerX - markerGap - markerWidth
		if innerStart > innerEnd {
			// fallback: center across the entire content area
			innerStart = contentX
			innerEnd = contentX + contentWidth - 1
		}
		innerWidth := innerEnd - innerStart + 1
		labelRuneLen := utf8.RuneCountInString(label)
		labelX := innerStart + (innerWidth-labelRuneLen)/2

		// choose styles and markers based on state
		disabled := false
		if i < len(selFlags) && !selFlags[i] {
			disabled = true
		}

		if i == selected {
			// selected item (always shown reversed) — but if disabled, still highlight but show '!' markers
			selStyle := tcell.StyleDefault.Reverse(true)
			for x := 0; x < contentWidth; x++ {
				u.s.SetContent(contentX+x, barY, ' ', nil, selStyle)
			}
			// draw left marker
			if leftMarkerX >= contentX && leftMarkerX < contentX+contentWidth {
				marker := "▶"
				if disabled {
					marker = "!"
				}
				u.drawText(leftMarkerX, barY, marker, selStyle)
			}
			// draw right marker
			if rightMarkerX >= contentX && rightMarkerX < contentX+contentWidth {
				marker := "◀"
				if disabled {
					marker = "!"
				}
				u.drawText(rightMarkerX, barY, marker, selStyle)
			}
			// draw centered label between markers
			u.drawText(labelX, barY, label, selStyle)
		} else {
			// not selected
			if disabled {
				// greyed out style with '!' markers
				disStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
				// draw markers
				if leftMarkerX >= contentX && leftMarkerX < contentX+contentWidth {
					u.drawText(leftMarkerX, barY, "!", disStyle)
				}
				if rightMarkerX >= contentX && rightMarkerX < contentX+contentWidth {
					u.drawText(rightMarkerX, barY, "!", disStyle)
				}
				// draw label in grey
				u.drawText(labelX, barY, label, disStyle)
			} else {
				st := tcell.StyleDefault
				// draw centered label without markers
				u.drawText(labelX, barY, label, st)
			}
		}
	}

	// footer hint: draw tokens inline without a full reversed background; color only bracketed keys
	tokens := []struct{ key, label string }{
		{"↑↓", "move"},
		{"↵", "select"},
		{"h", "help"},
		{"q", "quit"},
	}
	sep := " ・ "
	col := tcell.ColorYellow
	colStyle := tcell.StyleDefault.Foreground(col).Bold(true)
	normal := tcell.StyleDefault

	// draw tokens inside the card area (no full-width reverse/background)
	statusY := startY + totalHeight
	x := startX + 2
	maxX := startX + blockWidth - 2
	for i, tkn := range tokens {
		if i > 0 {
			for _, r := range sep {
				if x >= maxX {
					break
				}
				u.s.SetContent(x, statusY, r, nil, normal)
				x++
			}
		}

		// draw '[' + key + ']' with colored style
		if x < maxX {
			u.s.SetContent(x, statusY, '[', nil, colStyle)
			x++
		}
		for _, r := range tkn.key {
			if x >= maxX {
				break
			}
			u.s.SetContent(x, statusY, r, nil, colStyle)
			x++
		}
		if x < maxX {
			u.s.SetContent(x, statusY, ']', nil, colStyle)
			x++
		}

		// space then label (normal style)
		if x < maxX {
			u.s.SetContent(x, statusY, ' ', nil, normal)
			x++
		}
		for _, r := range tkn.label {
			if x >= maxX {
				break
			}
			u.s.SetContent(x, statusY, r, nil, normal)
			x++
		}
	}

	// if not logged in, show login hint below footer in bright red
	if !api.HasToken() {
		hint := "You are not logged in. Press 'L' to login with an API key."
		red := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
		hx := startX + 2
		hy := startY + totalHeight + 1
		for i, r := range hint {
			u.s.SetContent(hx+i, hy, r, nil, red)
		}
	} else {
		// show logged in user's email on footer right
		email, err := api.GetCurrentUserEmail()
		if err == nil && email != "" {
			info := "Logged in: " + email
			// draw at right side of card
			x := startX + blockWidth - 2 - utf8.RuneCountInString(info)
			if x < startX+2 {
				x = startX + 2
			}
			col := tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
			for i, r := range info {
				u.s.SetContent(x+i, statusY, r, nil, col)
			}
		}
	}

	u.s.Show()
}

func (u *UI) ShowMainMenu() {
	u.RenderMainMenu(0)
}

func (u *UI) ShowHelp() {
	u.s.Clear()
	w, h := u.s.Size()
	helpLines := []string{
		"Secrets Vault CLI - Help",
		"",
		"Navigation:",
		"  - Use Up/Down arrow keys to move through the main menu.",
		"  - Press Enter to select an item.",
		"  - Shortcuts: L=Login, S=Signup, H=Help, Q=Quit",
		"",
		"Features:",
		"  - Login / Signup with email + master password",
		"  - List, create, update and delete secrets (encrypted using master password)",
		"  - Manage API keys (create, revoke, list)",
		"",
		"Press any key to return to the main menu",
	}

	// draw boxed help in center using simple ASCII border (safer for terminals)
	boxW := 70
	if boxW > w-4 {
		boxW = w - 4
	}
	boxH := len(helpLines) + 4
	boxStartX := (w - boxW) / 2
	boxStartY := (h - boxH) / 2
	borderStyle := tcell.StyleDefault
	fillStyle := tcell.StyleDefault

	// draw box background
	for y := 0; y < boxH; y++ {
		for x := 0; x < boxW; x++ {
			u.s.SetContent(boxStartX+x, boxStartY+y, ' ', nil, fillStyle)
		}
	}
	// draw border using ASCII
	for x := 0; x < boxW; x++ {
		u.s.SetContent(boxStartX+x, boxStartY, '-', nil, borderStyle)
		u.s.SetContent(boxStartX+x, boxStartY+boxH-1, '-', nil, borderStyle)
	}
	for y := 0; y < boxH; y++ {
		u.s.SetContent(boxStartX, boxStartY+y, '|', nil, borderStyle)
		u.s.SetContent(boxStartX+boxW-1, boxStartY+y, '|', nil, borderStyle)
	}
	u.s.SetContent(boxStartX, boxStartY, '+', nil, borderStyle)
	u.s.SetContent(boxStartX+boxW-1, boxStartY, '+', nil, borderStyle)
	u.s.SetContent(boxStartX, boxStartY+boxH-1, '+', nil, borderStyle)
	u.s.SetContent(boxStartX+boxW-1, boxStartY+boxH-1, '+', nil, borderStyle)

	// draw help text
	for i, line := range helpLines {
		for j, r := range line {
			// trim if line too long for box
			if j+2 >= boxW-2 {
				break
			}
			u.s.SetContent(boxStartX+2+j, boxStartY+2+i, r, nil, fillStyle)
		}
	}
	u.s.Show()

	// wait for any key or resize (ignore mouse)
	for {
		e := u.s.PollEvent()
		switch e.(type) {
		case *tcell.EventKey:
			u.RenderMainMenu(0)
			return
		case *tcell.EventResize:
			// re-render main menu on resize
			u.RenderMainMenu(0)
			return
		default:
			// ignore other events (mouse, etc.)
		}
	}
}

func (u *UI) MenuOptions() ([]string, []bool) {
	// build menu depending on login state: if logged in, hide Login
	if api.HasToken() {
		menu := []string{"Secrets", "Help", "Quit"}
		sel := make([]bool, len(menu))
		for i := range sel {
			sel[i] = true
		}
		return menu, sel
	}

	menu := []string{"Login", "Secrets", "Help", "Quit"}
	sel := make([]bool, len(menu))
	for i := range sel {
		sel[i] = true
	}
	// Secrets disabled when not logged in
	for i, it := range menu {
		if it == "Secrets" {
			sel[i] = false
		}
	}
	return menu, sel
}

func (u *UI) DrawCenteredText(x, y int, s string, fg, bg tcell.Color) {
	style := tcell.StyleDefault.Foreground(fg).Background(bg)
	for i, r := range s {
		u.s.SetContent(x+i, y, r, nil, style)
	}
}

// ShowDisabledWarning displays a bright red hint near the bottom and waits for a key press
func (u *UI) ShowDisabledWarning(msg string) {
	w, h := u.s.Size()
	red := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	downY := h - 3
	if downY < 0 {
		downY = 0
	}
	text := "Hold on — " + msg + " requires login with an API key."
	for i := 0; i < w; i++ {
		u.s.SetContent(i, downY, ' ', nil, tcell.StyleDefault)
	}
	for i, r := range text {
		if i >= w-2 {
			break
		}
		u.s.SetContent(2+i, downY, r, nil, red)
	}
	u.s.Show()
	// wait for any key to clear
	for {
		e := u.s.PollEvent()
		switch e.(type) {
		case *tcell.EventKey:
			// clear the line
			for i := 0; i < w; i++ {
				u.s.SetContent(i, downY, ' ', nil, tcell.StyleDefault)
			}
			u.s.Show()
			return
		case *tcell.EventResize:
			return
		}
	}
}
