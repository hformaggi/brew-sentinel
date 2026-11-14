package tui

import (
	"context"
	"testing"
	"time"

	"brew-sentinel/internal/brew"
	"brew-sentinel/internal/inventory"

	"github.com/gdamore/tcell/v2"
)

func TestSimulatedScreen(t *testing.T) {
	u := New(context.Background(), &brew.Client{Path: "/nonexistent"})
	u.snapshot = inventory.Snapshot{Packages: []inventory.Package{{ID: "formula:a", Name: "a", Explicit: true, Dependencies: []string{"formula:b"}}, {ID: "formula:b", Name: "b"}}}
	u.graph = inventory.NewGraph(u.snapshot)
	u.render()
	screen := tcell.NewSimulationScreen("UTF-8")
	screen.SetSize(100, 30)
	u.app.SetScreen(screen)
	done := make(chan error, 1)
	go func() { done <- u.app.Run() }()
	u.app.QueueUpdateDraw(func() {
		if u.table.GetRowCount() != 3 {
			t.Error("missing rows")
		}
		u.dependencies(u.snapshot.Packages[0])
		if name, _ := u.pages.GetFrontPage(); name != "tree" {
			t.Error("missing dependency tree")
		}
		u.pages.RemovePage("tree")
		u.detail(u.snapshot.Packages[0])
		if name, _ := u.pages.GetFrontPage(); name != "detail" {
			t.Error("missing detail")
		}
	})
	u.app.Stop()
	select {
	case e := <-done:
		if e != nil {
			t.Fatal(e)
		}
	case <-time.After(time.Second * 3):
		t.Fatal("screen did not stop")
	}
}
