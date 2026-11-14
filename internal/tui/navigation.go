package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// actionBar is shared by every dialog. Horizontal arrows select, Enter activates.
func actionBar(actions []menuItem) *tview.Table {
	bar := tview.NewTable().SetSelectable(false, true).SetSelectedStyle(tcell.StyleDefault.Foreground(bg).Background(accent).Bold(true))
	bar.SetBackgroundColor(surface)
	for i, a := range actions {
		bar.SetCell(0, i, tview.NewTableCell("  "+a.label+"  ").SetTextColor(ink).SetBackgroundColor(line).SetAlign(tview.AlignCenter))
	}
	bar.Select(0, 0)
	bar.SetSelectedFunc(func(_ int, col int) {
		if col >= 0 && col < len(actions) && actions[col].run != nil {
			actions[col].run()
		}
	})
	return bar
}
func (u *UI) activateNav(row int) {
	if row >= 1 && row <= 5 {
		u.setSection(row - 1)
		u.app.SetFocus(u.table)
		return
	}
	switch row {
	case 8:
		u.app.SetFocus(u.search)
	case 9:
		u.install()
	case 10:
		u.filterMenu()
	case 11:
		u.sortMenu()
	case 12:
		u.maintenance()
	case 13:
		u.securityReport()
	case 14:
		u.help()
	case 15:
		u.languageMenu()
	case 16:
		u.app.Stop()
	}
}
func (u *UI) sortMenu() {
	u.menu(u.tr("Sort packages"), []menuItem{{u.tr("Name"), u.tr("Alphabetical order"), 0, func() { u.sortDependents = false; u.sortSize = false; u.render() }}, {u.tr("Largest disk usage"), u.tr("Unknown sizes appear last"), 0, func() { u.sortSize = true; u.sortDependents = false; u.render() }}, {u.tr("Most dependents"), u.tr("Packages used by the most tools"), 0, func() { u.sortDependents = true; u.sortSize = false; u.render() }}})
}
func (u *UI) help() {
	u.message(u.tr("Navigation help"), u.tr("Use the same controls on every screen:\n\n  ↑ ↓    Browse items or scroll text\n  ← →    Switch panes, tabs or buttons\n  Enter  Open or activate the selected item\n  Esc    Go back or cancel\n\nPACKAGE LIST\n← opens the sidebar. ↑ on the first package opens search.\n→ opens the side preview, or details on smaller screens.\nIn the sidebar, ↑↓ select an option and Enter opens it.\n→ returns to the list, applying the selected library filter.\n\nSEARCH\nType a name, description or combined filters.\n←→ move the cursor. ↓ returns to the list; ↑ opens the sidebar.\nEnter applies the search. Esc clears the text and returns.\n\nPACKAGE DETAILS\n←→ switch tabs. ↑↓ browse the content.\nEnter opens a related package or an action.\n\nPREVIEWS\n←→ choose Cancel or Confirm. ↑ moves to the scrollable preview.\n←→ return to the buttons. Enter activates the selected button only.\n\nOptional shortcuts: / search · i catalog · m maintenance\ns security report · v rescan · r refresh · q quit\n\nUpdates use local Homebrew metadata. The automatic security\nscan covers formulae, with coverage limitations."))
}
