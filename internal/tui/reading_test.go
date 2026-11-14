package tui

import (
	"brew-sentinel/internal/brew"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
	"strings"
	"testing"
)

func TestReadingLayouts(t *testing.T) {
	cases := []struct {
		name, expected string
		open           func(*UI)
	}{
		{"detalhes", "Why is it installed?", func(u *UI) { u.detail(u.snapshot.Packages[0]) }},
		{"dependencias", "beta", func(u *UI) {
			u.detail(u.snapshot.Packages[0])
			u.pages.InputHandler()(tcell.NewEventKey(tcell.KeyRight, 0, 0), func(p tview.Primitive) { u.app.SetFocus(p) })
		}},
		{"confirmacao", "Confirm", func(u *UI) {
			u.confirm("Revisar atualização", strings.Repeat("Atualizar ferramenta 1.0 → 2.0\n", 40), func() { t.Error("rendering executed operation") })
		}},
		{"catalogo", "Applications", func(u *UI) { u.install() }},
		{"candidato", "Install", func(u *UI) {
			u.candidateDetail(brew.Candidate{Name: "sample", FullName: "sample", Kind: "formula", Version: "2.0", Description: "Uma ferramenta para trabalhar com arquivos e dependências no terminal."})
		}},
		{"manutencao", "Clean old versions", func(u *UI) { u.maintenance() }},
	}
	for _, locale := range []string{"en", "pt"} {
		for _, width := range []int{80, 120, 160} {
			for _, tc := range cases {
				t.Run(fmt.Sprintf("%s-%s-%d", locale, tc.name, width), func(t *testing.T) {
					u := workbenchFixture()
					u.useLanguage(locale)
					u.resize(width)
					tc.open(u)
					screen := tcell.NewSimulationScreen("UTF-8")
					screen.Init()
					defer screen.Fini()
					screen.SetSize(width, 24)
					u.pages.SetRect(0, 0, width, 24)
					u.pages.Draw(screen)
					screen.Show()
					cells, w, h := screen.GetContents()
					var out strings.Builder
					for y := 0; y < h; y++ {
						for x := 0; x < w; x++ {
							r := cells[y*w+x].Runes
							if len(r) > 0 {
								out.WriteRune(r[0])
							} else {
								out.WriteByte(' ')
							}
						}
						out.WriteByte('\n')
					}
					text := out.String()
					want := tc.expected
					if locale == "pt" {
						if translated, ok := map[string]string{"Why is it installed?": "Por que está instalado?", "Confirm": "Confirmar", "Applications": "Aplicativos", "Install": "Instalar", "Clean old versions": "Limpar versões"}[want]; ok {
							want = translated
						}
					}
					if !strings.Contains(text, want) {
						t.Fatalf("missing %s\n%s", tc.expected, text)
					}
					if (tc.name == "catalogo" || tc.name == "confirmacao" || tc.name == "candidato" || tc.name == "manutencao") && strings.Contains(text, "BREW SENTINEL") {
						t.Fatal("previous page leaks through dialog")
					}
					if strings.Contains(text, "Tab ") {
						t.Fatal("Tab instruction remains")
					}
					os.WriteFile(fmt.Sprintf("/private/tmp/brew-sentinel-%s-%s-%d.txt", locale, tc.name, width), []byte(text), 0600)
				})
			}
		}
	}
}
