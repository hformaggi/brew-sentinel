package tui

import (
	"context"
	"fmt"

	"brew-sentinel/internal/brew"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (u *UI) install() {
	if !u.ready() {
		return
	}
	query := tview.NewInputField().SetLabel(u.tr("Name  ")).SetFieldBackgroundColor(line).SetFieldTextColor(ink).SetLabelColor(muted).SetPlaceholder(u.tr("e.g. git, postgres, firefox"))
	query.SetBackgroundColor(surface)
	kinds := actionBar([]menuItem{{label: u.tr("Formulae")}, {label: u.tr("Applications")}})
	var close func()
	search := func() {
		index := 0
		_, index = kinds.GetSelection()
		kind := "formula"
		if index == 1 {
			kind = "cask"
		}
		q := query.GetText()
		if q == "" {
			return
		}
		u.startTask("Searching the catalog…", func(ctx context.Context) (func(), error) {
			items, e := u.client.Search(ctx, kind, q)
			return func() {
				if !u.pages.HasPage("catalog-search") {
					return
				}
				if e != nil {
					u.message(u.tr("Search failed"), u.errorText(e))
					return
				}
				close()
				u.catalogResults(q, items)
			}, nil
		})
	}
	buttons := actionBar([]menuItem{{label: u.tr("Search"), run: search}, {label: u.tr("Back"), run: func() { close() }}})
	intro := textView().SetText(u.tr("Find a tool or application in Homebrew.\nType a name and press Enter to search."))
	kindLabel := textView().SetText(u.tr("Package type"))
	hint := textView().SetText(u.tr("↑↓ switch field · ←→ choose type or button · Esc back"))
	root := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(intro, 3, 0, false).AddItem(query, 2, 0, true).AddItem(kindLabel, 1, 0, false).AddItem(kinds, 2, 0, false).AddItem(buttons, 2, 0, false).AddItem(hint, 1, 0, false)
	root.Box = tview.NewBox().SetBackgroundColor(surface)
	panel(root.Box, u.tr("Install package"))
	root.SetBorderPadding(1, 1, 2, 2)
	close = u.showPage("catalog-search", center(root, 74, 16), query)
	controls := []tview.Primitive{query, kinds, buttons}
	root.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape {
			close()
			return nil
		}
		index := 0
		for i, p := range controls {
			if u.app.GetFocus() == p {
				index = i
			}
		}
		switch e.Key() {
		case tcell.KeyUp:
			if index > 0 {
				u.app.SetFocus(controls[index-1])
			}
			return nil
		case tcell.KeyDown:
			if index < len(controls)-1 {
				u.app.SetFocus(controls[index+1])
			}
			return nil
		case tcell.KeyEnter:
			if index == 0 {
				search()
				return nil
			}
			if index == 1 {
				u.app.SetFocus(buttons)
				return nil
			}
		}
		return e
	})
}
func (u *UI) catalogResults(query string, items []brew.Candidate) {
	list := tview.NewList().ShowSecondaryText(true).SetMainTextColor(ink).SetSecondaryTextColor(muted).SetSelectedTextColor(accent).SetSelectedBackgroundColor(line).SetHighlightFullLine(true)
	panel(list.Box, fmt.Sprintf(u.tr("CATALOG · %s · %d results · Esc back"), query, len(items)))
	for _, item := range items {
		item := item
		list.AddItem(item.Name, item.Kind+u.tr(" · Enter to view version and install"), 0, func() {
			u.startTask("Looking up %s…", func(ctx context.Context) (func(), error) {
				candidate, e := u.client.Describe(ctx, item.Kind, item.FullName)
				return func() {
					if !u.pages.HasPage("catalog-results") {
						return
					}
					if e != nil {
						u.message(u.tr("Lookup failed"), u.errorText(e))
						return
					}
					u.candidateDetail(candidate)
				}, nil
			}, item.Name)
		})
	}
	if len(items) == 0 {
		list.AddItem(u.tr("No results"), u.tr("Try a different part of the package name."), 0, nil)
	}
	close := u.showPage("catalog-results", list, list)
	list.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape {
			close()
			return nil
		}
		return e
	})
}
func (u *UI) candidateDetail(c brew.Candidate) {
	kind := u.tr("Formula · CLI tool or library")
	if c.Kind == "cask" {
		kind = u.tr("Application")
	}
	text := fmt.Sprintf(u.tr("%s\n\n%s\n\nAvailable version   %s\nType                %s\n\nProject website\n%s\n\nAn installation preview will be shown before confirmation."), c.FullName, c.Description, c.Version, kind, c.Homepage)
	actions := []menuItem{{label: u.tr("Back")}}
	if c.Installed {
		actions = append(actions, menuItem{label: u.tr("Upgrade"), run: func() { u.mutate("upgrade", c.Kind, c.FullName) }})
	} else {
		actions = append(actions, menuItem{label: u.tr("Install"), run: func() { u.mutate("install", c.Kind, c.FullName) }})
	}
	u.actionDialog("catalog-package", u.tr("Catalog package"), text, actions)
}
