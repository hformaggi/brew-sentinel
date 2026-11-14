package tui

import (
	"brew-sentinel/internal/inventory"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

func (u *UI) showPage(name string, root, focus tview.Primitive) func() {
	base := name
	for n := 2; u.pages.HasPage(name); n++ {
		name = fmt.Sprintf("%s-%d", base, n)
	}
	previous := u.app.GetFocus()
	closed := false
	close := func() {
		if closed {
			return
		}
		closed = true
		u.pages.RemovePage(name)
		if previous == u.inspector && u.width < 140 {
			previous = u.table
		}
		if previous != nil {
			u.app.SetFocus(previous)
		} else {
			u.app.SetFocus(u.table)
		}
	}
	u.pages.AddPage(name, root, true, true)
	u.app.SetFocus(focus)
	return close
}
func (u *UI) message(title, text string) {
	u.styledMessage(title, tview.Escape(text))
}
func (u *UI) styledMessage(title, text string) {
	view := textView().SetText(text).SetScrollable(true).SetWrap(true).SetWordWrap(true)
	panel(view.Box, title+u.tr(" · Esc back"))
	close := u.showPage("detail", view, view)
	view.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape {
			close()
			return nil
		}
		return e
	})
}
func (u *UI) sourceLabel(source string) string {
	switch source {
	case "installation receipt":
		return u.tr("installation receipt")
	case "current recipe":
		return u.tr("current recipe · may differ from the installation")
	default:
		return u.tr("partial receipt · incomplete relationships")
	}
}
func (u *UI) detail(p inventory.Package) {
	overview := textView().SetScrollable(true)
	panel(overview.Box, u.tr("Overview"))
	text := fmt.Sprintf(u.tr("[#67dfb5::b]%s[-::-]\n%s\n\n[#8396ad]Installed[-]       %s\n[#8396ad]Available[-]      %s\n[#8396ad]Disk usage[-]  %s\n\n[#86baff::b]Why is it installed?[-::-]\n"), tview.Escape(p.Name), tview.Escape(p.Description), tview.Escape(Versions(p)), tview.Escape(p.Available), u.sizeText(p.ID))
	switch p.Reason {
	case "explicit":
		text += u.tr("You requested this package's installation.\n")
	case "dependency":
		text += u.tr("This package was installed as another tool's dependency.\n")
	default:
		text += u.tr("The receipt does not record why this package was installed.\n")
	}
	paths := u.graph.RootPaths(p.ID)
	if !p.Explicit && len(paths) > 0 {
		text += u.tr("\nKnown paths to this package:\n")
		for i, path := range paths {
			if i >= 3 {
				text += u.tr("  … see also the dependency tree\n")
				break
			}
			names := []string{}
			for _, id := range path {
				names = append(names, strings.TrimPrefix(strings.TrimPrefix(id, "formula:"), "cask:"))
			}
			text += "  " + tview.Escape(strings.Join(names, " → ")) + "\n"
		}
	}
	text += u.tr("\n[#8396ad]Relationship evidence[-]\n") + u.sourceLabel(p.DependencySource) + u.tr("\n\n[#8396ad]Project website[-]\n") + tview.Escape(p.Homepage)
	if p.Pinned {
		text += u.tr("\n\n[#f1c079]This version is pinned.[-]")
	}
	if p.Deprecated || p.Disabled {
		text += u.tr("\n\n[#f1c079]Package deprecated or disabled in Homebrew.[-]")
	}
	overview.SetText(text)
	index := map[string]inventory.Package{}
	for _, item := range u.snapshot.Packages {
		index[item.ID] = item
	}
	relationList := func(title string, ids []string) *tview.List {
		list := readingList()
		panel(list.Box, title)
		for _, id := range ids {
			item, ok := index[id]
			if !ok {
				list.AddItem(id, u.tr("Not found in inventory"), 0, nil)
				continue
			}
			itemCopy := item
			list.AddItem(item.Name, Versions(item)+" · "+u.reasonLabel(item.Reason), 0, func() { u.detail(itemCopy) })
		}
		if len(ids) == 0 {
			list.AddItem(u.tr("No relationships recorded"), u.tr("Coverage depends on the package's receipts and recipe."), 0, nil)
		}
		return list
	}
	deps := relationList(u.tr("This package depends on"), p.Dependencies)
	dependents := relationList(u.tr("Packages that depend on this one"), u.graph.Dependents(p.ID))
	security := textView().SetScrollable(true)
	panel(security.Box, u.tr("Security"))
	securityText := u.tr("[#86baff::b]Known vulnerabilities[-::-]\n\n") + tview.Escape(u.securityText(p)) + "\n\n"
	if p.Kind == "cask" {
		securityText += u.tr("Homebrew's scanner does not check apps/casks.")
	} else {
		securityText += u.tr("The scan covers the scope reported by Homebrew. Retained older versions may not have been scanned.\n")
	}
	if u.report != nil {
		securityText += u.tr("\nScanned on ") + u.report.CollectedAt.Format(u.tr("Jan 2, 15:04")) + "\n"
		for _, f := range u.report.Findings {
			if f.Formula != p.FullName {
				continue
			}
			securityText += "\n[#8396ad]" + u.tr("Scanned version") + ": " + tview.Escape(f.Version) + "[-]\n"
			for _, v := range f.Vulnerabilities {
				securityText += u.advisoryText(v, f.Version, false)
			}
			for _, v := range f.Patched {
				securityText += u.advisoryText(v, f.Version, true)
			}
		}
	}
	security.SetText(securityText)
	actions := readingList()
	panel(actions.Box, u.tr("Available actions"))
	actions.AddItem(u.tr("Upgrade package"), u.tr("Review changes before confirming"), 0, func() { u.mutate("upgrade", p.Kind, p.FullName) }).AddItem(u.tr("Uninstall package"), u.tr("Check dependents and review removal"), 0, func() { u.mutate("uninstall", p.Kind, p.FullName) }).AddItem(u.tr("Explore dependency tree"), u.tr("Browse direct dependencies"), 0, func() { u.dependencies(p) })
	views := []tview.Primitive{overview, deps, dependents, security, actions}
	labels := []string{u.tr("Overview"), fmt.Sprintf(u.tr("Dependencies %d"), len(p.Dependencies)), fmt.Sprintf(u.tr("Dependents %d"), len(u.graph.Dependents(p.ID))), u.tr("Security"), u.tr("Actions")}
	tabs := tview.NewTable().SetSelectable(false, true).SetSelectedStyle(tcell.StyleDefault.Foreground(bg).Background(accent))
	tabs.SetBackgroundColor(bg)
	for i, label := range labels {
		tabs.SetCell(0, i, tview.NewTableCell(" "+label+" ").SetTextColor(muted).SetBackgroundColor(bg))
	}
	content := tview.NewPages()
	for i, v := range views {
		content.AddPage(fmt.Sprint(i), v, true, i == 0)
	}
	header := textView().SetText(" [#67dfb5::b]" + tview.Escape(p.Name) + u.tr("[-::-]  [#8396ad]Packages / details[-]"))
	header.SetBackgroundColor(bg)
	footer := textView().SetText(u.tr(" [#67dfb5]←→[-] switch tab   [#67dfb5]↑↓[-] browse content   [#67dfb5]Enter[-] open   [#67dfb5]Esc[-] back"))
	footer.SetBackgroundColor(bg)
	root := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(header, 2, 0, false).AddItem(tabs, 2, 0, false).AddItem(content, 0, 1, true).AddItem(footer, 2, 0, false)
	root.SetBackgroundColor(bg)
	position := 0
	selectTab := func(i int) {
		position = i
		tabs.Select(0, i)
		content.SwitchToPage(fmt.Sprint(i))
		u.app.SetFocus(views[i])
	}
	close := u.showPage("detail", root, overview)
	tabs.SetSelectedFunc(func(row, col int) { selectTab(col) })
	root.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		switch e.Key() {
		case tcell.KeyEscape:
			close()
			return nil
		case tcell.KeyLeft:
			selectTab((position + len(views) - 1) % len(views))
			return nil
		case tcell.KeyRight:
			selectTab((position + 1) % len(views))
			return nil
		}
		return e
	})
}
func readingList() *tview.List {
	return tview.NewList().ShowSecondaryText(true).SetMainTextColor(ink).SetSecondaryTextColor(muted).SetSelectedTextColor(ink).SetSelectedBackgroundColor(line).SetHighlightFullLine(true)
}
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
