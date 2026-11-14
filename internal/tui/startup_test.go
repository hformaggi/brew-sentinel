package tui

import (
	"brew-sentinel/internal/brew"
	"context"
	"github.com/gdamore/tcell/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartupScansAutomaticallyOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "brew-fake")
	script := `#!/bin/sh
printf '%s\n' "$*" >> "$0.log"
case "$1" in
 --cellar) dirname "$0" ;;
 info) printf '%s' '{"formulae":[{"name":"foo","full_name":"foo","installed":[]}],"casks":[]}' ;;
 vulns) printf '%s' '{"findings":[{"formula":"foo","version":"1","vulnerabilities":[{"id":"CVE-TEST"}]}],"checked":1,"complete":true}'; exit 1 ;;
esac
`
	if e := os.WriteFile(path, []byte(script), 0700); e != nil {
		t.Fatal(e)
	}
	u := New(context.Background(), &brew.Client{Path: path})
	stop := runningUI(t, u)
	defer stop()
	waitUI(t, u, func() bool { return u.report != nil && !u.busy && !u.loading && !u.measuring })
	u.app.QueueUpdateDraw(func() {
		if name, _ := u.pages.GetFrontPage(); name != "main" {
			t.Error("startup interrupted navigation", name)
		}
		u.render()
	})
	data, e := os.ReadFile(path + ".log")
	if e != nil {
		t.Fatal(e)
	}
	if strings.Count(string(data), "vulns --json") != 1 {
		t.Fatalf("expected one automatic scan: %s", data)
	}
}
func TestStartupScanFailureDoesNotBlockInventory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "brew-fake")
	os.WriteFile(path, []byte(`#!/bin/sh
case "$1" in
 --cellar) dirname "$0" ;;
 info) printf '%s' '{"formulae":[],"casks":[]}' ;;
 vulns) printf offline >&2; exit 2 ;;
esac
`), 0700)
	u := New(context.Background(), &brew.Client{Path: path})
	stop := runningUI(t, u)
	defer stop()
	waitUI(t, u, func() bool { return u.scanError != "" && !u.busy && !u.loading && !u.measuring })
	u.app.QueueUpdateDraw(func() {
		if u.report != nil {
			t.Error("failed scan became clean report")
		}
		if name, _ := u.pages.GetFrontPage(); name != "main" {
			t.Error("error blocked navigation", name)
		}
	})
}

func TestDiskDoesNotWaitForScanAndNavigationPreservesResult(t *testing.T) {
	path := filepath.Join(t.TempDir(), "brew-fake")
	os.WriteFile(path, []byte(`#!/bin/sh
case "$1" in
 --cellar) dirname "$0" ;;
 info) printf '%s' '{"formulae":[],"casks":[]}' ;;
 vulns) sleep 1; printf '%s' '{"findings":[],"checked":0,"complete":true}' ;;
esac
`), 0700)
	u := New(context.Background(), &brew.Client{Path: path})
	stop := runningUI(t, u)
	defer stop()
	waitUI(t, u, func() bool { return u.scanning && u.sizes != nil })
	u.app.QueueUpdateDraw(func() { u.message("Pacote", "Detalhes") })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	waitUI(t, u, func() bool { return u.report != nil && !u.scanning })
	u.app.QueueUpdateDraw(func() {
		name, _ := u.pages.GetFrontPage()
		if name != "main" {
			t.Error(name)
		}
		u.acceptSnapshot(u.snapshot)
		if u.report == nil {
			t.Error("unchanged inventory discarded scan")
		}
	})
}
