package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"brew-sentinel/internal/brew"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestFormulaOperationCanBeCancelled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "brew-fake")
	script := `#!/bin/sh
case "$1" in
 upgrade) printf 'operation started\n'; exec sleep 30 ;;
 --cellar) dirname "$0" ;;
 info) printf '%s' '{"formulae":[],"casks":[]}' ;;
 vulns) printf '%s' '{"findings":[],"checked":0,"complete":true}' ;;
esac
`
	if err := os.WriteFile(path, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	u := workbenchFixture()
	ctx, cancel := context.WithCancel(context.Background())
	u.ctx = ctx
	u.client = &brew.Client{Path: path}
	stop := runningUI(t, u)
	defer stop()
	defer func() {
		cancel()
		waitUI(t, u, func() bool { return !u.busy && !u.loading && !u.scanning && !u.measuring })
	}()
	u.app.QueueUpdateDraw(func() { u.perform([]string{"upgrade", "--formula", "foo"}, "formula") })
	waitUI(t, u, func() bool {
		view, ok := u.app.GetFocus().(*tview.TextView)
		return ok && strings.Contains(view.GetText(false), "operation started")
	})
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyCtrlC, 0, 0))
	waitUI(t, u, func() bool { return !u.busy && !u.loading && !u.scanning && !u.measuring })
	u.app.QueueUpdateDraw(func() {
		if u.cancel != nil || u.task != "" {
			t.Error("operation retained task state after cancellation")
		}
	})
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "main" })
}

func TestCancelledPreparationDoesNotShowConfirmation(t *testing.T) {
	u := workbenchFixture()
	stop := runningUI(t, u)
	defer stop()
	confirmed := false
	u.app.QueueUpdateDraw(func() {
		u.startTask("Preparing", func(ctx context.Context) (func(), error) {
			<-ctx.Done()
			return func() { confirmed = true; u.confirm("Unexpected", "Cancelled", func() {}) }, nil
		})
	})
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyCtrlC, 0, 0))
	waitUI(t, u, func() bool { return !u.busy })
	u.app.QueueUpdateDraw(func() {
		if confirmed {
			t.Error("cancelled preparation opened a confirmation")
		}
	})
}
