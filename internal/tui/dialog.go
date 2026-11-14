package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// actionDialog shares one arrow-key model between confirmations and package actions.
func (u *UI) actionDialog(name, title, text string, actions []menuItem) {
	view := textView().SetDynamicColors(false).SetText(text).SetScrollable(true).SetWrap(true)
	panel(view.Box, title)
	var close func()
	wrapped := make([]menuItem, len(actions))
	for i, a := range actions {
		a := a
		wrapped[i] = menuItem{label: a.label, run: func() {
			close()
			if a.run != nil {
				a.run()
			}
		}}
	}
	buttons := actionBar(wrapped)
	hint := textView().SetText(u.tr(" ↑↓ scroll  ·  ←→ choose button  ·  Enter open  ·  Esc back"))
	hint.SetBackgroundColor(bg)
	root := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(view, 0, 1, false).AddItem(buttons, 1, 0, true).AddItem(hint, 2, 0, false)
	root.SetBackgroundColor(surface)
	w := 92
	if u.width > 0 && u.width < 100 {
		w = u.width - 4
	}
	close = u.showPage(name, center(root, w, 20), buttons)
	root.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		switch e.Key() {
		case tcell.KeyEscape:
			close()
			return nil
		case tcell.KeyLeft, tcell.KeyRight:
			if u.app.GetFocus() != buttons {
				u.app.SetFocus(buttons)
				return nil
			}
			return e
		case tcell.KeyUp:
			if u.app.GetFocus() == buttons {
				u.app.SetFocus(view)
				return nil
			}
		case tcell.KeyDown:
			if u.app.GetFocus() == buttons {
				return nil
			}
		}
		return e
	})
	view.SetFocusFunc(func() { view.SetBorderColor(accent) })
	view.SetBlurFunc(func() { view.SetBorderColor(line) })
}
