package tui

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"brew-sentinel/internal/inventory"

	"github.com/gdamore/tcell/v2"
)

func TestWorkbenchLayouts(t *testing.T) {
	for _, locale := range []string{"en", "pt"} {
		for _, size := range [][2]int{{80, 24}, {120, 32}, {160, 40}} {
			t.Run(locale+fmt.Sprint(size), func(t *testing.T) {
				u := workbenchFixture()
				u.useLanguage(locale)
				screen := tcell.NewSimulationScreen("UTF-8")
				if e := screen.Init(); e != nil {
					t.Fatal(e)
				}
				defer screen.Fini()
				screen.SetSize(size[0], size[1])
				u.resize(size[0])
				u.pages.SetRect(0, 0, size[0], size[1])
				u.pages.Draw(screen)
				screen.Show()
				cells, w, h := screen.GetContents()
				var b strings.Builder
				for y := 0; y < h; y++ {
					for x := 0; x < w; x++ {
						r := cells[y*w+x].Runes
						if len(r) > 0 {
							b.WriteRune(r[0])
						} else {
							b.WriteByte(' ')
						}
					}
					b.WriteByte('\n')
				}
				text := b.String()
				if !strings.Contains(text, "SENTINEL") || !strings.Contains(text, "alpha") {
					t.Fatalf("missing overview\n%s", text)
				}
				why := "Why is it here"
				if locale == "pt" {
					why = "Por que está aqui"
				}
				if size[0] >= 140 && !strings.Contains(text, why) {
					t.Fatalf("missing inspector\n%s", text)
				}
				os.WriteFile(fmt.Sprintf("/private/tmp/brew-sentinel-%s-%d.txt", locale, size[0]), []byte(text), 0600)
			})
		}
	}
}
func TestInitialEmptyDrawDoesNotScrollToEnd(t *testing.T) {
	u := workbenchFixture()
	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	defer screen.Fini()
	screen.SetSize(80, 24)
	u.table.Clear()
	u.table.SetRect(0, 0, 80, 12)
	u.table.Draw(screen)
	u.snapshot.Packages = nil
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("pkg-%03d", i)
		u.snapshot.Packages = append(u.snapshot.Packages, inventory.Package{ID: "formula:" + name, Name: name, Kind: "formula"})
	}
	u.graph = inventory.NewGraph(u.snapshot)
	u.render()
	u.table.Draw(screen)
	row, _ := u.table.GetOffset()
	if row != 0 {
		t.Fatalf("first load starts scrolled to row %d", row)
	}
}
