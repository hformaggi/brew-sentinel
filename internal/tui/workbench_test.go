package tui

import (
	"context"
	"testing"

	"brew-sentinel/internal/brew"
	"brew-sentinel/internal/inventory"
)

func workbenchFixture() *UI {
	u := NewWithOptions(context.Background(), &brew.Client{Path: "/nonexistent"}, Options{Language: "en"})
	u.snapshot = inventory.Snapshot{Packages: []inventory.Package{{ID: "formula:alpha", Name: "alpha", Kind: "formula", Reason: "explicit", Explicit: true, Outdated: true, Dependencies: []string{"formula:beta"}}, {ID: "formula:beta", Name: "beta", Kind: "formula", Reason: "dependency"}, {ID: "cask:browser", Name: "browser", Kind: "cask", Reason: "unknown"}}}
	for i := range u.snapshot.Packages {
		p := &u.snapshot.Packages[i]
		p.FullName = p.Name
		p.Available = "2.1.0"
		p.Versions = []inventory.InstalledVersion{{Version: "2.0.0"}}
		p.Description = "Ferramenta de exemplo para organizar o ambiente local."
		p.DependencySource = "installation receipt"
	}
	u.graph = inventory.NewGraph(u.snapshot)
	u.render()
	return u
}
func TestWorkbenchSections(t *testing.T) {
	u := workbenchFixture()
	u.setSection(2)
	if len(u.visible) != 1 || u.visible[0].Kind != "cask" {
		t.Fatal("apps filter")
	}
	u.setSection(3)
	if len(u.visible) != 1 || !u.visible[0].Outdated {
		t.Fatal("updates filter")
	}
	u.setSection(0)
	if len(u.visible) != 3 {
		t.Fatal("all filter")
	}
}
func TestSelectionRemainsStable(t *testing.T) {
	u := workbenchFixture()
	u.table.Select(2, 0)
	u.sortDependents = true
	u.render()
	p, ok := u.selected()
	if !ok || p.ID != "formula:beta" {
		t.Fatal("selection changed", p)
	}
}
func TestEmptySearchKeepsSummary(t *testing.T) {
	u := workbenchFixture()
	u.search.SetText("missing-package")
	if len(u.visible) != 0 {
		t.Fatal("not empty")
	}
	if u.metrics[0].GetText(false) == "" {
		t.Fatal("summary disappeared")
	}
}
