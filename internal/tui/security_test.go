package tui

import (
	"brew-sentinel/internal/security"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestCVELinksAcrossWidgets(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	defer screen.Fini()
	screen.SetSize(100, 12)
	text := "ação 漢 CVE-2026-12345, CVE-2025-9999."
	view := textView().SetText(text)
	view.SetRect(0, 0, 100, 2)
	view.Draw(screen)
	table := tview.NewTable()
	table.SetCell(0, 0, tview.NewTableCell("CVE-2024-123456"))
	table.SetRect(0, 3, 100, 2)
	table.Draw(screen)
	linkCVEs(screen)
	found := 0
	for y := 0; y < 12; y++ {
		for x := 0; x < 100; x++ {
			r, _, style, _ := screen.GetContent(x, y)
			if r != 'C' {
				continue
			}
			for _, id := range []string{"CVE-2026-12345", "CVE-2025-9999", "CVE-2024-123456"} {
				var visible strings.Builder
				for i := 0; i < len(id); i++ {
					c, _, _, _ := screen.GetContent(x+i, y)
					visible.WriteRune(c)
				}
				if visible.String() != id {
					continue
				}
				found++
				for i := 0; i < len(id); i++ {
					_, _, cellStyle, _ := screen.GetContent(x+i, y)
					if cellStyle != cellStyle.Url("https://nvd.nist.gov/vuln/detail/"+id) {
						t.Fatalf("missing URL on %s at %d", id, i)
					}
				}
				_, _, next, _ := screen.GetContent(x+len(id), y)
				if next == style {
					t.Fatal("link leaked past identifier")
				}
			}
		}
	}
	if found != 3 {
		t.Fatalf("linked %d identifiers", found)
	}
}

func securityFixture() *UI {
	u := workbenchFixture()
	u.report = &security.Report{Checked: 42, Complete: false, CollectedAt: time.Date(2026, 11, 14, 15, 0, 0, 0, time.UTC), Skipped: []string{"skipped"}, Unscanned: map[string]string{"zeta": "missing receipt", "beta": "no SBOM"}, Findings: []security.Finding{{Formula: u.snapshot.Packages[0].FullName, Version: "1.0", Vulnerabilities: []security.Vulnerability{{ID: "CVE-2026-12345", Severity: "HIGH", Summary: "A vulnerability in request parsing [red]must stay literal", Aliases: []string{"GHSA-2345-6789-abcd"}, FixedVersions: []string{"1.1", "2.0"}}}, Patched: []security.Vulnerability{{ID: "CVE-2025-98765", Severity: "LOW", Summary: "Already fixed by the formula"}}}}}
	return u
}

func TestSecurityReportLayoutsAndDetails(t *testing.T) {
	for _, lang := range []string{"en", "pt"} {
		for _, width := range []int{80, 120, 160} {
			for _, detail := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s-%d-detail-%t", lang, width, detail), func(t *testing.T) {
					u := securityFixture()
					u.useLanguage(lang)
					u.resize(width)
					if detail {
						u.detail(u.snapshot.Packages[0])
						for i := 0; i < 3; i++ {
							u.pages.InputHandler()(tcell.NewEventKey(tcell.KeyRight, 0, 0), func(p tview.Primitive) { u.app.SetFocus(p) })
						}
					} else {
						u.securityReport()
					}
					screen := tcell.NewSimulationScreen("UTF-8")
					screen.Init()
					defer screen.Fini()
					screen.SetSize(width, 40)
					u.pages.SetRect(0, 0, width, 40)
					u.pages.Draw(screen)
					linkCVEs(screen)
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
					for _, want := range []string{"CVE-2026-12345", "CVE-2025-98765", "1.1, 2.0", "[red]must stay literal"} {
						if !strings.Contains(text, want) {
							t.Fatalf("missing %q\n%s", want, text)
						}
					}
					os.WriteFile(fmt.Sprintf("/tmp/brew-sentinel-security-%s-%d-%t.txt", lang, width, detail), []byte(text), 0600)
				})
			}
		}
	}
}

func TestSecurityCoverageIsPreserved(t *testing.T) {
	u := securityFixture()
	u.report.Findings = nil
	text := u.securityReportText()
	for _, want := range []string{"No known vulnerabilities", "Incomplete", "Casks", "Skipped", "skipped", "beta · no SBOM", "zeta · missing receipt"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s", want)
		}
	}
	if strings.Index(text, "beta ·") > strings.Index(text, "zeta ·") {
		t.Fatal("coverage errors not sorted")
	}
}

func TestAdvisoryDoesNotLeakLinkOrUnderline(t *testing.T) {
	u := securityFixture()
	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	defer screen.Fini()
	screen.SetSize(100, 20)
	view := textView().SetText(u.advisoryText(u.report.Findings[0].Vulnerabilities[0], "1.0", false))
	view.SetRect(0, 0, 100, 20)
	view.Draw(screen)
	linkCVEs(screen)
	for y := 3; y < 20; y++ {
		for x := 0; x < 100; x++ {
			_, _, style, _ := screen.GetContent(x, y)
			if style != style.Underline(false) {
				t.Fatalf("underline leaked onto metadata at %d,%d", x, y)
			}
			if style != style.Url("") {
				t.Fatalf("hyperlink leaked onto metadata at %d,%d", x, y)
			}
		}
	}
}

func TestOpenSecurityReportUpdatesAfterScan(t *testing.T) {
	for _, failed := range []bool{false, true} {
		t.Run(fmt.Sprint(failed), func(t *testing.T) {
			u := securityFixture()
			u.scanning = true
			u.securityReport()
			view := u.securityView
			if !strings.Contains(view.GetText(false), "update automatically") {
				t.Fatal("missing loading state")
			}
			u.scanning = false
			if failed {
				u.scanError = "offline"
			}
			u.render()
			if u.securityView != view || u.app.GetFocus() != view {
				t.Fatal("report view or focus was replaced")
			}
			want := "CVE-2026-12345"
			if failed {
				want = "offline"
			}
			if !strings.Contains(view.GetText(false), want) {
				t.Fatalf("missing completion state: %s", view.GetText(false))
			}
			view.InputHandler()(tcell.NewEventKey(tcell.KeyEscape, 0, 0), func(p tview.Primitive) { u.app.SetFocus(p) })
			if u.securityView != nil {
				t.Fatal("closed view retained")
			}
			u.render()
			if name, _ := u.pages.GetFrontPage(); name != "main" {
				t.Fatal("completion reopened report")
			}
		})
	}
}
