package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"brew-sentinel/internal/brew"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func updatesFixture(t *testing.T) (*UI, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "brew-fake")
	script := `#!/bin/sh
printf '%s\n' "$*" >> "$0.log"
case "$1" in
 upgrade) case "$*" in
  *--dry-run*) printf 'Would upgrade sample 1.0 -> 2.0\n' ;;
  *) printf 'PACKAGES_CHANGED\n' >> "$0.log" ;;
 esac ;;
 update) printf 'CATALOG_CHANGED\n' >> "$0.log" ;;
 --cellar) dirname "$0" ;;
 info) printf '%s' '{"formulae":[],"casks":[]}' ;;
 vulns) printf '%s' '{"findings":[],"checked":0,"complete":true}' ;;
 *) exit 99 ;;
esac
`
	if err := os.WriteFile(path, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	u := workbenchFixture()
	u.client = &brew.Client{Path: path}
	return u, path + ".log"
}

func chooseUpdateItem(t *testing.T, u *UI, label string) {
	t.Helper()
	found := false
	u.app.QueueUpdateDraw(func() {
		list, ok := u.app.GetFocus().(*tview.List)
		if !ok {
			return
		}
		for i := 0; i < list.GetItemCount(); i++ {
			main, _ := list.GetItemText(i)
			if main == label {
				list.SetCurrentItem(i)
				found = true
				return
			}
		}
	})
	if !found {
		t.Fatalf("menu item %q not found", label)
	}
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
}

func openUpdates(t *testing.T, u *UI) {
	t.Helper()
	u.app.QueueUpdateDraw(func() { u.maintenance() })
	chooseUpdateItem(t, u, "Updates")
	waitUI(t, u, func() bool {
		list, ok := u.app.GetFocus().(*tview.List)
		if !ok {
			return false
		}
		text, _ := list.GetItemText(0)
		return text == "Upgrade all packages"
	})
}

func TestMaintenanceUpgradePreviewsCanBeCancelled(t *testing.T) {
	for _, tc := range []struct{ label, command string }{
		{"Upgrade all packages", "upgrade --dry-run"},
		{"Upgrade formulae", "upgrade --formula --dry-run"},
		{"Upgrade applications", "upgrade --cask --dry-run"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			u, logPath := updatesFixture(t)
			stop := runningUI(t, u)
			defer stop()
			openUpdates(t, u)
			chooseUpdateItem(t, u, tc.label)
			waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "confirm" })
			u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
			waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "main" })
			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != tc.command+"\n" {
				t.Fatalf("unexpected commands before confirmation: %s", data)
			}
		})
	}
}

func TestMaintenanceFormulaUpgradeRequiresConfirmation(t *testing.T) {
	u, logPath := updatesFixture(t)
	stop := runningUI(t, u)
	defer stop()
	openUpdates(t, u)
	chooseUpdateItem(t, u, "Upgrade formulae")
	waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "confirm" })
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyRight, 0, 0))
	u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	waitUI(t, u, func() bool {
		name, _ := u.pages.GetFrontPage()
		return name == "operation" && !u.busy && !u.loading && !u.scanning && !u.measuring
	})
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "upgrade --formula --dry-run\nupgrade --formula\nPACKAGES_CHANGED\n") {
		t.Fatalf("wrong upgrade scope or order: %s", data)
	}
}

func TestMaintenanceHomebrewUpdateRequiresConfirmation(t *testing.T) {
	for _, confirm := range []bool{false, true} {
		t.Run(map[bool]string{false: "cancel", true: "confirm"}[confirm], func(t *testing.T) {
			u, logPath := updatesFixture(t)
			stop := runningUI(t, u)
			defer stop()
			openUpdates(t, u)
			chooseUpdateItem(t, u, "Update Homebrew")
			waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "confirm" })
			if _, err := os.Stat(logPath); !os.IsNotExist(err) {
				t.Fatal("Homebrew ran before confirmation")
			}
			if !confirm {
				u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
				waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return name == "main" })
				if _, err := os.Stat(logPath); !os.IsNotExist(err) {
					t.Fatal("cancel ran Homebrew")
				}
				return
			}
			u.app.QueueEvent(tcell.NewEventKey(tcell.KeyRight, 0, 0))
			u.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
			waitUI(t, u, func() bool {
				name, _ := u.pages.GetFrontPage()
				return name == "operation" && !u.busy && !u.loading && !u.scanning && !u.measuring
			})
			data, err := os.ReadFile(logPath)
			if err != nil || !strings.HasPrefix(string(data), "update\nCATALOG_CHANGED\n--cellar\ninfo --json=v2 --installed\n") {
				t.Fatalf("catalog update did not refresh inventory: %s %v", data, err)
			}
		})
	}
}

func TestMaintenanceUpgradeDoesNotConfirmFailedOrEmptyPreview(t *testing.T) {
	for _, tc := range []struct{ name, script, message string }{
		{"failed", "printf 'catalog unavailable' >&2; exit 1", "catalog unavailable"},
		{"empty", "exit 0", "Homebrew did not list any packages to upgrade."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			u, logPath := updatesFixture(t)
			path := strings.TrimSuffix(logPath, ".log")
			if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$0.log\"\n"+tc.script+"\n"), 0700); err != nil {
				t.Fatal(err)
			}
			stop := runningUI(t, u)
			defer stop()
			openUpdates(t, u)
			chooseUpdateItem(t, u, "Upgrade all packages")
			waitUI(t, u, func() bool { name, _ := u.pages.GetFrontPage(); return !u.busy && name == "detail" })
			u.app.QueueUpdateDraw(func() {
				view, ok := u.app.GetFocus().(*tview.TextView)
				if !ok || !strings.Contains(view.GetText(false), tc.message) {
					t.Error("preview result was not explained")
				}
			})
			data, err := os.ReadFile(logPath)
			if err != nil || string(data) != "upgrade --dry-run\n" {
				t.Fatalf("unexpected command after preview: %s %v", data, err)
			}
		})
	}
}

func TestMaintenanceUpdatesAreTranslated(t *testing.T) {
	u := workbenchFixture()
	u.useLanguage("pt")
	stop := runningUI(t, u)
	defer stop()
	u.app.QueueUpdateDraw(func() { u.maintenance() })
	chooseUpdateItem(t, u, "Atualizações")
	waitUI(t, u, func() bool {
		list, ok := u.app.GetFocus().(*tview.List)
		if !ok {
			return false
		}
		main, _ := list.GetItemText(0)
		return main == "Atualizar tudo"
	})
	u.app.QueueUpdateDraw(func() {
		list := u.app.GetFocus().(*tview.List)
		for i, want := range []string{"Atualizar tudo", "Atualizar fórmulas", "Atualizar aplicativos", "Atualizar Homebrew"} {
			if got, _ := list.GetItemText(i); got != want {
				t.Errorf("label = %q, want %q", got, want)
			}
		}
	})
}
