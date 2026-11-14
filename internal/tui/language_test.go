package tui

import (
	"brew-sentinel/internal/brew"
	"brew-sentinel/internal/i18n"
	"context"
	"github.com/gdamore/tcell/v2"
	"path/filepath"
	"strings"
	"testing"
)

func TestLanguageMenuPersistsAndUpdatesInterface(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	u := workbenchFixture()
	u.configPath = path
	stop := runningUI(t, u)
	defer stop()
	u.app.QueueUpdateDraw(func() { u.languageMenu() })
	for _, key := range []tcell.Key{tcell.KeyDown, tcell.KeyDown, tcell.KeyEnter} {
		u.app.QueueEvent(tcell.NewEventKey(key, 0, 0))
	}
	waitUI(t, u, func() bool { return u.language.Language == "pt" })
	u.app.QueueUpdateDraw(func() {
		if !strings.Contains(u.header.GetText(false), "Pacotes") {
			t.Error("header not translated")
		}
		if u.search.GetLabel() != "Buscar  " {
			t.Error("search not translated")
		}
	})
	pref, e := i18n.Load(path)
	if e != nil || pref != "pt" {
		t.Fatalf("%s %v", pref, e)
	}
	next := NewWithOptions(context.Background(), &brew.Client{Path: "/nonexistent"}, Options{ConfigPath: path})
	if next.language.Language != "pt" {
		t.Fatal("preference not restored")
	}
	override := NewWithOptions(context.Background(), &brew.Client{Path: "/nonexistent"}, Options{ConfigPath: path, Language: "en"})
	if override.language.Language != "en" {
		t.Fatal("explicit override ignored")
	}
}
func TestSystemLanguageAndEnglishFallback(t *testing.T) {
	for _, locale := range []string{"pt_BR.UTF-8", "fr_FR.UTF-8", "C"} {
		t.Run(locale, func(t *testing.T) {
			t.Setenv("LC_ALL", locale)
			u := NewWithOptions(context.Background(), &brew.Client{Path: "/nonexistent"}, Options{ConfigPath: filepath.Join(t.TempDir(), "settings.json")})
			want := "en"
			if strings.HasPrefix(locale, "pt") {
				want = "pt"
			}
			if u.language.Language != want {
				t.Fatalf("got %s want %s", u.language.Language, want)
			}
		})
	}
}
func TestEnglishInspectorAndPortugueseValidation(t *testing.T) {
	u := workbenchFixture()
	u.table.Select(2, 0)
	u.updateInspector()
	if !strings.Contains(u.inspector.GetText(false), "Used by") {
		t.Error("English inspector contains untranslated label")
	}
	u.useLanguage("pt")
	_, e := u.client.Search(context.Background(), "formula", "git foo")
	if got := u.errorText(e); got != "use parte do nome do pacote, sem espaços ou opções" {
		t.Fatal(got)
	}
}
func TestTaskUpdatesLanguage(t *testing.T) {
	u := workbenchFixture()
	u.task = "Preparing preview for %s…"
	u.taskArgs = []any{"git"}
	u.useLanguage("pt")
	if !strings.Contains(u.footer.GetText(false), "Preparando prévia de git") {
		t.Fatal(u.footer.GetText(false))
	}
}
