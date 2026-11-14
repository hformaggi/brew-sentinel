package tui

import (
	"github.com/gdamore/tcell/v2"
	"testing"
)

func TestCommandPaletteShortcut(t *testing.T) {
	u := workbenchFixture()
	u.handleKey(tcell.NewEventKey(tcell.KeyCtrlP, 0, 0))
	name, _ := u.pages.GetFrontPage()
	if name != "menu" {
		t.Fatal("palette not opened", name)
	}
}
func TestFilterMenuDoesNotChangeSelectionUntilChosen(t *testing.T) {
	u := workbenchFixture()
	u.handleKey(tcell.NewEventKey(tcell.KeyRune, 'f', 0))
	name, _ := u.pages.GetFrontPage()
	if name != "menu" || u.search.GetText() != "" {
		t.Fatal("filter menu missing or applied prematurely")
	}
}
