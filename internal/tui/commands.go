package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (u *UI) commands() {
	items := []menuItem{
		{u.tr("Language"), u.tr("System default, English or Portuguese"), 0, u.languageMenu},
		{u.tr("Search installed packages"), u.tr("Name, description or combined filters"), '/', func() { u.app.SetFocus(u.search) }},
		{u.tr("Quick filters"), u.tr("Requested packages, dependencies, apps and updates"), 'f', u.filterMenu},
		{u.tr("Browse catalog"), u.tr("Find and install new packages"), 'i', u.install},
		{u.tr("Security report"), u.tr("View findings, coverage and errors from the latest scan"), 's', u.securityReport},
		{u.tr("Rescan vulnerabilities"), u.tr("Refresh known vulnerability findings"), 'v', u.scan},
		{u.tr("Measure disk usage"), u.tr("Refresh formula disk usage"), 'd', u.measure},
		{u.tr("Maintenance"), u.tr("Preview cleanup and orphan removal"), 'm', u.maintenance},
		{u.tr("Refresh inventory"), u.tr("Reload installed packages"), 'r', u.refresh},
	}
	u.menu(u.tr("COMMANDS"), items)
}
func (u *UI) filterMenu() {
	apply := func(query string) func() {
		return func() { u.section = 0; u.search.SetText(query); u.render(); u.app.SetFocus(u.table) }
	}
	u.menu(u.tr("QUICK FILTERS"), []menuItem{
		{u.tr("All packages"), u.tr("Clear filters and search"), '1', apply("")},
		{u.tr("Requested by you"), u.tr("The receipt records a requested installation"), '2', apply("reason:explicit")},
		{u.tr("Dependencies"), u.tr("Packages installed for other tools"), '3', apply("reason:dependency")},
		{u.tr("Available updates"), u.tr("Based on local Homebrew metadata"), '4', apply("outdated:true")},
		{u.tr("Applications"), u.tr("Casks only"), '5', apply("type:cask")},
		{u.tr("Unknown origin"), u.tr("Packages with no known installation reason"), '6', apply("reason:unknown")},
	})
}
func (u *UI) securityReport() {
	view := textView().SetScrollable(true).SetWrap(true).SetWordWrap(true)
	panel(view.Box, u.tr("SECURITY REPORT")+u.tr(" · Esc back"))
	u.securityView = view
	u.updateSecurityView()
	close := u.showPage("detail", view, view)
	view.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape {
			if u.securityView == view {
				u.securityView = nil
			}
			close()
			return nil
		}
		return e
	})
}

func (u *UI) updateSecurityView() {
	if u.securityView == nil {
		return
	}
	var text string
	switch {
	case u.scanning:
		text = "[#86baff::b]" + u.tr("Scanning vulnerabilities…") + "[-::-]\n\n" + u.tr("This report will update automatically when the scan finishes.") + "\n\n" + u.tr("You can press Esc to continue browsing.")
	case u.scanError != "":
		text = "[#ff919b::b]" + u.tr("Scan failed") + "[-::-]\n\n" + tview.Escape(u.scanError) + u.tr("\n\nPress v in the list to try again.")
	case u.report == nil:
		text = u.tr("There is no report for this inventory yet. Press v in the list to start a scan.")
	default:
		text = u.securityReportText()
	}
	if u.securityView.GetText(false) != text {
		u.securityView.SetText(text)
	}
}
