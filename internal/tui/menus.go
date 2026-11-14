package tui

import (
	"brew-sentinel/internal/inventory"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (u *UI) menu(title string, items []menuItem) {
	list := tview.NewList().ShowSecondaryText(true).SetMainTextColor(ink).SetSecondaryTextColor(muted).SetSelectedBackgroundColor(line).SetSelectedTextColor(accent).SetHighlightFullLine(true)
	panel(list.Box, title+u.tr(" · ↑↓ select · Enter open · Esc back"))
	var close func()
	for _, item := range items {
		item := item
		list.AddItem(item.label, item.description, 0, func() { close(); item.run() })
	}
	close = u.showPage("menu", center(list, 64, len(items)*2+3), list)
	list.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape {
			close()
			return nil
		}
		return e
	})
}

type menuItem struct {
	label, description string
	key                rune
	run                func()
}

func (u *UI) packageActions(p inventory.Package) {
	u.menu(u.tr("Actions · ")+p.Name, []menuItem{{u.tr("Inspect"), u.tr("Versions, origin and relationships"), 'e', func() { u.detail(p) }}, {u.tr("Explore dependencies"), u.tr("Browse the dependency tree"), 't', func() { u.dependencies(p) }}, {u.tr("Upgrade package"), u.tr("Review the preview before confirming"), 'u', func() { u.mutate("upgrade", p.Kind, p.FullName) }}, {u.tr("Uninstall package"), u.tr("Review dependents and confirm removal"), 'x', func() { u.mutate("uninstall", p.Kind, p.FullName) }}})
}
func (u *UI) maintenance() {
	u.menu(u.tr("Maintenance"), []menuItem{{u.tr("Reload inventory"), u.tr("Read installed packages again"), 'r', u.refresh}, {u.tr("Measure disk usage"), u.tr("Formula kegs · no file changes"), 'd', u.measure}, {u.tr("Scan vulnerabilities"), u.tr("Check CVEs/advisories through Homebrew"), 'v', u.scan}, {u.tr("Clean old versions and cache"), u.tr("Preview what brew cleanup would remove"), 'c', func() { u.executeCleanup("cleanup") }}, {u.tr("Remove orphaned dependencies"), u.tr("Preview what brew autoremove would remove"), 'a', func() { u.executeCleanup("autoremove") }}})
}
