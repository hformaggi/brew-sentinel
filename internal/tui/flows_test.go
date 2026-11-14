package tui

import (
	"brew-sentinel/internal/brew"
	"github.com/gdamore/tcell/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runningUI(t *testing.T, u *UI) func() {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	u.app.SetScreen(screen)
	done := make(chan error, 1)
	go func() { done <- u.app.Run() }()
	u.app.QueueUpdateDraw(func() {})
	return func() {
		u.app.Stop()
		select {
		case e := <-done:
			if e != nil {
				t.Error(e)
			}
		case <-time.After(3 * time.Second):
			t.Error("UI failed to stop")
		}
	}
}
func waitUI(t *testing.T, u *UI, predicate func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		ok := make(chan bool, 1)
		u.app.QueueUpdateDraw(func() { ok <- predicate() })
		if <-ok {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("UI state timed out")
}
func TestConfirmationCancelDoesNotExecute(t *testing.T) {
	u := workbenchFixture()
	stop := runningUI(t, u)
	defer stop()
	called := false
	u.app.QueueUpdateDraw(func() { u.confirm("Remove", "Review", func() { called = true }) })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "main" })
	if called {
		t.Fatal("cancel executed mutation")
	}
}
func TestCatalogSelectionResolvesCandidate(t *testing.T) {
	u := workbenchFixture()
	stop := runningUI(t, u)
	defer stop()
	u.app.QueueUpdateDraw(func() {
		u.candidateDetail(brew.Candidate{Kind: "formula", Name: "foo", FullName: "foo", Version: "2", Description: "Tool"})
		if name, _ := u.pages.GetFrontPage(); name != "catalog-package" {
			t.Error("missing contextual install")
		}
	})
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "main" })
}
func TestPreviewCancellationUsesOnlyFakeBrew(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "brew-fake")
	script := `#!/bin/sh
printf '%s\n' "$*" >> "$0.log"
case "$1" in
 --cellar) dirname "$0" ;;
 info) printf '%s' '{"formulae":[],"casks":[]}' ;;
 install) case "$*" in *--dry-run) printf 'Would install foo\n';; *) printf 'MUTATION\n' >> "$0.log";; esac ;;
esac
`
	if e := os.WriteFile(path, []byte(script), 0700); e != nil {
		t.Fatal(e)
	}
	u := workbenchFixture()
	u.client = &brew.Client{Path: path}
	stop := runningUI(t, u)
	defer stop()
	u.app.QueueUpdateDraw(func() { u.mutate("install", "formula", "foo") })
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "confirm" })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "main" })
	data, e := os.ReadFile(path + ".log")
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(string(data), "--dry-run") || strings.Contains(string(data), "MUTATION") {
		t.Fatalf("unsafe preview: %s", data)
	}
}
func TestSearchTypingDoesNotTriggerActions(t *testing.T) {
	u := workbenchFixture()
	u.app.SetFocus(u.search)
	e := tcell.NewEventKey(tcell.KeyRune, 'x', 0)
	if u.handleKey(e) != e {
		t.Fatal("search key treated as action")
	}
}

func TestConfirmationKeyboardCanConfirm(t *testing.T) {
	u := workbenchFixture()
	stop := runningUI(t, u)
	defer stop()
	called := false
	u.app.QueueUpdateDraw(func() { u.confirm("Confirm", "Review", func() { called = true }) })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyRight, 0, 0))
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	time.Sleep(30 * time.Millisecond)
	ok := make(chan bool, 1)
	u.app.QueueUpdateDraw(func() { ok <- called })
	if !<-ok {
		t.Fatal("Right then Enter did not reach Confirmar")
	}
}
func TestDismissedCatalogDoesNotReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fake-brew")
	os.WriteFile(path, []byte("#!/bin/sh\nsleep 0.2\nprintf 'foo\\n'\n"), 0700)
	u := workbenchFixture()
	u.client = &brew.Client{Path: path}
	stop := runningUI(t, u)
	defer stop()
	u.app.QueueUpdateDraw(func() { u.install() })
	for _, r := range "foo" {
		u.app.QueueEvent(tcell.NewEventKey(tcell.KeyRune, r, 0))
	}
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	waitUI(t, u, func() bool { return u.busy })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	waitUI(t, u, func() bool { return !u.busy })
	result := make(chan string, 1)
	u.app.QueueUpdateDraw(func() { name, _ := u.pages.GetFrontPage(); result <- name })
	if got := <-result; got != "main" {
		t.Fatal("dismissed catalog reopened", got)
	}
}
