package tui

import (
	"brew-sentinel/internal/brew"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainArrowNavigation(t *testing.T) {
	u := workbenchFixture()
	stop := runningUI(t, u)
	defer stop()
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
	waitUI(t, u, func() bool { return u.app.GetFocus() == u.nav })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyRight, 0, 0))
	waitUI(t, u, func() bool { return u.section == 1 && u.app.GetFocus() == u.table })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyUp, 0, 0))
	waitUI(t, u, func() bool { return u.app.GetFocus() == u.search })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	waitUI(t, u, func() bool { return u.app.GetFocus() == u.table })
}
func TestConfirmationUsesArrows(t *testing.T) {
	u := workbenchFixture()
	stop := runningUI(t, u)
	defer stop()
	called := false
	u.app.QueueUpdateDraw(func() { u.confirm("Confirmar", "Texto de prévia", func() { called = true }) })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyRight, 0, 0))
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	waitUI(t, u, func() bool { return called })
}
func TestDetailsUseArrowTabs(t *testing.T) {
	u := workbenchFixture()
	stop := runningUI(t, u)
	defer stop()
	u.app.QueueUpdateDraw(func() { u.detail(u.snapshot.Packages[0]) })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyRight, 0, 0))
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "detail-2" })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "detail" })
}

func TestCatalogFlowWithOnlyArrows(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fake-brew")
	script := `#!/bin/sh
printf '%s\n' "$*" >> "$0.log"
case "$1" in
 search) printf 'sample-app\n' ;;
 --cellar) dirname "$0" ;;
 info) case "$*" in
 *--installed*) printf '%s' '{"formulae":[],"casks":[]}' ;;
 *) printf '%s' '{"formulae":[],"casks":[{"token":"sample-app","full_token":"sample-app","desc":"A sample application","version":"2.0","installed":null}]}' ;;
 esac ;;
 install) case "$*" in *--dry-run) printf 'Would install sample-app\n';; *) printf MUTATION >> "$0.log" ;; esac ;;
esac
`
	if e := os.WriteFile(path, []byte(script), 0700); e != nil {
		t.Fatal(e)
	}
	u := workbenchFixture()
	u.client = &brew.Client{Path: path}
	stop := runningUI(t, u)
	defer stop()
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
	for i := 0; i < 6; i++ {
		u.app.QueueEvent(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	}
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "catalog-search" })
	for _, r := range "sample" {
		u.app.QueueEvent(tcell.NewEventKey(tcell.KeyRune, r, 0))
	}
	for _, k := range []tcell.Key{tcell.KeyDown, tcell.KeyRight, tcell.KeyDown, tcell.KeyEnter} {
		u.app.QueueEvent(tcell.NewEventKey(k, 0, 0))
	}
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "catalog-results" })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "catalog-package" })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyRight, 0, 0))
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "confirm" })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "catalog-results" })
	data, e := os.ReadFile(path + ".log")
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(string(data), "search --cask sample") || !strings.Contains(string(data), "--dry-run") || strings.Contains(string(data), "MUTATION") {
		t.Fatalf("incorrect catalog flow: %s", data)
	}
}

func TestSidebarToolsReachableWithArrows(t *testing.T) {
	for _, tc := range []struct {
		name  string
		steps int
		page  string
	}{{"catálogo", 6, "catalog-search"}, {"filtros", 7, "menu"}, {"ordenação", 8, "menu"}, {"manutenção", 9, "menu"}, {"relatório", 10, "detail"}, {"ajuda", 11, "detail"}} {
		t.Run(tc.name, func(t *testing.T) {
			u := workbenchFixture()
			stop := runningUI(t, u)
			defer stop()
			u.app.QueueEvent(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
			for i := 0; i < tc.steps; i++ {
				u.app.QueueEvent(tcell.NewEventKey(tcell.KeyDown, 0, 0))
			}
			u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
			waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == tc.page })
		})
	}
}

func TestTreeExpandsAndCollapsesWithArrows(t *testing.T) {
	u := workbenchFixture()
	stop := runningUI(t, u)
	defer stop()
	u.app.QueueUpdateDraw(func() {
		u.dependencies(u.snapshot.Packages[0])
		tree := u.app.GetFocus().(*tview.TreeView)
		tree.GetRoot().SetExpanded(false)
	})
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyRight, 0, 0))
	waitUI(t, u, func() bool { return u.app.GetFocus().(*tview.TreeView).GetRoot().IsExpanded() })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
	waitUI(t, u, func() bool { return !u.app.GetFocus().(*tview.TreeView).GetRoot().IsExpanded() })
}

func TestReadingPreviewDoesNotSelectConfirmation(t *testing.T) {
	u := workbenchFixture()
	stop := runningUI(t, u)
	defer stop()
	called := false
	u.app.QueueUpdateDraw(func() { u.confirm("Revisar", "Prévia longa", func() { called = true }) })
	for _, key := range []tcell.Key{tcell.KeyUp, tcell.KeyRight, tcell.KeyEnter} {
		u.app.QueueEvent(tcell.NewEventKey(key, 0, 0))
	}
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "main" })
	u.app.QueueUpdateDraw(func() {
		if called {
			t.Error("returning from text selected confirmation")
		}
	})
}
func TestCloseAfterResizeRestoresVisibleFocus(t *testing.T) {
	u := workbenchFixture()
	stop := runningUI(t, u)
	defer stop()
	u.app.QueueUpdateDraw(func() { u.resize(160); u.app.SetFocus(u.inspector); u.detail(u.snapshot.Packages[0]); u.resize(120) })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	waitUI(t, u, func() bool { return u.app.GetFocus() == u.table })
}

func TestRightFromToolsReturnsToPackageList(t *testing.T) {
	for row := 8; row <= 16; row++ {
		t.Run(fmt.Sprint(row), func(t *testing.T) {
			u := workbenchFixture()
			u.table.Select(2, 0)
			u.app.SetFocus(u.nav)
			u.nav.Select(row, 0)
			u.handleKey(tcell.NewEventKey(tcell.KeyRight, 0, 0))
			if u.app.GetFocus() != u.table {
				t.Error("right arrow did not return to package list")
			}
			if name, _ := u.pages.GetFrontPage(); name != "main" {
				t.Errorf("right arrow activated tool: %s", name)
			}
			if p, ok := u.selected(); !ok || p.ID != "formula:beta" {
				t.Error("package selection changed")
			}
		})
	}
}
