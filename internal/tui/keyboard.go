package tui

import (
	"brew-sentinel/internal/inventory"

	"github.com/gdamore/tcell/v2"
)

func (u *UI) handleKey(e *tcell.EventKey) *tcell.EventKey {
	if e.Key() == tcell.KeyTab || e.Key() == tcell.KeyBacktab {
		return nil
	}
	if e.Key() == tcell.KeyCtrlC && (u.busy || u.scanning || u.measuring) {
		for _, cancel := range u.insightCancels {
			cancel()
		}
		if u.cancel != nil {
			u.cancel()
			u.status = u.tr("Stopping query…")
			u.updateChrome()
		}
		return nil
	}

	if name, _ := u.pages.GetFrontPage(); name != "main" {
		return e
	}
	if e.Key() == tcell.KeyCtrlP {
		u.commands()
		return nil
	}
	switch u.app.GetFocus() {
	case u.search:
		switch e.Key() {
		case tcell.KeyUp:
			u.app.SetFocus(u.nav)
			return nil
		case tcell.KeyDown:
			u.app.SetFocus(u.table)
			return nil
		}
		return e
	case u.nav:
		switch e.Key() {
		case tcell.KeyRight:
			row, _ := u.nav.GetSelection()
			if row >= 1 && row <= 5 {
				u.setSection(row - 1)
			}
			u.app.SetFocus(u.table)
			return nil
		case tcell.KeyEnter:
			row, _ := u.nav.GetSelection()
			u.activateNav(row)
			return nil
		case tcell.KeyLeft:
			return nil
		}
	case u.table:
		switch e.Key() {
		case tcell.KeyLeft:
			u.app.SetFocus(u.nav)
			return nil
		case tcell.KeyRight:
			if u.width >= 140 {
				u.app.SetFocus(u.inspector)
			} else if p, ok := u.selected(); ok {
				u.detail(p)
			}
			return nil
		case tcell.KeyUp:
			row, _ := u.table.GetSelection()
			if row <= 1 {
				u.app.SetFocus(u.search)
				return nil
			}
		}
	case u.inspector:
		switch e.Key() {
		case tcell.KeyLeft:
			u.app.SetFocus(u.table)
			return nil
		case tcell.KeyRight:
			return nil
		case tcell.KeyEnter:
			if p, ok := u.selected(); ok {
				u.detail(p)
			}
			return nil
		}
	}
	if e.Key() == tcell.KeyEscape {
		u.search.SetText("")
		return nil
	}
	if e.Key() == tcell.KeyCtrlR {
		u.refresh()
		return nil
	}
	switch e.Rune() {
	case '1', '2', '3', '4', '5':
		u.setSection(int(e.Rune() - '1'))
	case 'q':
		u.app.Stop()
	case '/':
		u.app.SetFocus(u.search)
	case 'r':
		u.refresh()
	case 'o':
		if u.sortSize {
			u.sortSize = false
			u.sortDependents = false
		} else if u.sortDependents {
			u.sortDependents = false
			u.sortSize = true
		} else {
			u.sortDependents = true
		}
		u.render()
	case 'f':
		u.filterMenu()
	case 's':
		u.securityReport()
	case ':':
		u.commands()
	case 'v':
		u.scan()
	case 'd':
		u.measure()
	case 'i':
		u.install()
	case 'm', 'c', 'a':
		u.maintenance()
	case 'C':
		u.executeCleanup("cleanup")
	case 'A':
		u.executeCleanup("autoremove")
	case 'u', 'x':
		if p, ok := u.selected(); ok {
			a := "upgrade"
			if e.Rune() == 'x' {
				a = "uninstall"
			}
			u.mutate(a, p.Kind, p.FullName)
		}
	case 't':
		if p, ok := u.selected(); ok {
			u.dependencies(p)
		}
	case ' ':
		if p, ok := u.selected(); ok {
			u.packageActions(p)
		}
	case '?':
		u.help()
	default:
		return e
	}
	return nil
}
func (u *UI) selected() (inventory.Package, bool) {
	r, _ := u.table.GetSelection()
	if r < 1 || r > len(u.visible) {
		return inventory.Package{}, false
	}
	return u.visible[r-1], true
}
func (u *UI) setSection(section int) { u.section = section; u.render() }
