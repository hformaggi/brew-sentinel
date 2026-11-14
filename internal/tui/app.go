package tui

import (
	"brew-sentinel/internal/brew"
	"brew-sentinel/internal/i18n"
	"brew-sentinel/internal/inventory"
	"brew-sentinel/internal/security"
	"context"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"reflect"
	"time"
)

type UI struct {
	taskArgs                  []any
	language                  i18n.Translator
	preference, configPath    string
	cancel                    context.CancelFunc
	initialized               bool
	busy, loading             bool
	securityView              *tview.TextView
	report                    *security.Report
	scanError                 string
	scanning                  bool
	measuring                 bool
	insightCancels            map[string]context.CancelFunc
	sizes                     map[string]int64
	app                       *tview.Application
	client                    *brew.Client
	pages                     *tview.Pages
	table                     *tview.Table
	search                    *tview.InputField
	header, footer, inspector *tview.TextView
	metrics                   [4]*tview.TextView
	nav                       *tview.Table
	body                      *tview.Flex
	packageArea               *tview.Flex
	snapshot                  inventory.Snapshot
	graph                     *inventory.Graph
	visible                   []inventory.Package
	sortDependents, sortSize  bool
	section, width            int
	task, status              string
	ctx                       context.Context
}

func New(ctx context.Context, c *brew.Client) *UI { return NewWithOptions(ctx, c, Options{}) }
func NewWithOptions(ctx context.Context, c *brew.Client, options Options) *UI {
	u := &UI{app: tview.NewApplication().EnableMouse(false), client: c, ctx: ctx}
	u.configPath = options.ConfigPath
	if u.configPath == "" {
		u.configPath, _ = i18n.ConfigPath()
	}
	preference := options.Language
	if preference == "" {
		preference, _ = i18n.Load(u.configPath)
	}
	if normalized, e := i18n.Preference(preference); e == nil {
		preference = normalized
	} else {
		preference = "system"
	}
	u.preference = preference
	system := "en"
	if preference == "system" {
		system = i18n.System()
	}
	u.language = i18n.New(i18n.Resolve(preference, system))
	u.buildLayout()
	for _, box := range []*tview.Box{u.table.Box, u.nav.Box, u.inspector.Box} {
		b := box
		b.SetFocusFunc(func() { b.SetBorderColor(accent) })
		b.SetBlurFunc(func() { b.SetBorderColor(line) })
	}
	u.nav.SetFocusFunc(func() { u.nav.SetSelectedStyle(tcell.StyleDefault.Foreground(bg).Background(accent)) })
	u.nav.SetBlurFunc(func() { u.nav.SetSelectedStyle(tcell.StyleDefault.Foreground(ink).Background(surface)) })
	u.table.SetSelectionChangedFunc(func(row, col int) { u.updateInspector() })
	u.table.SetSelectedFunc(func(row, col int) {
		if p, ok := u.selected(); ok {
			u.detail(p)
		}
	})
	u.search.SetChangedFunc(func(string) { u.render() })
	u.search.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			u.search.SetText("")
		}
		u.app.SetFocus(u.table)
	})
	u.app.SetInputCapture(u.handleKey)
	u.app.SetBeforeDrawFunc(func(s tcell.Screen) bool {
		w, _ := s.Size()
		if w != u.width {
			u.resize(w)
		}
		u.updateChrome()
		if u.graph == nil && !u.loading {
			u.refresh()
		}
		return false
	})
	u.app.SetAfterDrawFunc(linkCVEs)
	u.updateChrome()
	return u
}
func (u *UI) handleKey(e *tcell.EventKey) *tcell.EventKey {
	if e.Key() == tcell.KeyTab || e.Key() == tcell.KeyBacktab {
		return nil
	}
	if e.Key() == tcell.KeyCtrlC && (u.busy || u.scanning || u.measuring) {
		for _, cancel := range u.insightCancels {
			cancel()
		}
		if u.cancel != nil {
			u.cancel()
			u.status = u.tr("Stopping query…")
			u.updateChrome()
		}
		return nil
	}

	if name, _ := u.pages.GetFrontPage(); name != "main" {
		return e
	}
	if e.Key() == tcell.KeyCtrlP {
		u.commands()
		return nil
	}
	switch u.app.GetFocus() {
	case u.search:
		switch e.Key() {
		case tcell.KeyUp:
			u.app.SetFocus(u.nav)
			return nil
		case tcell.KeyDown:
			u.app.SetFocus(u.table)
			return nil
		}
		return e
	case u.nav:
		switch e.Key() {
		case tcell.KeyRight:
			row, _ := u.nav.GetSelection()
			if row >= 1 && row <= 5 {
				u.setSection(row - 1)
			}
			u.app.SetFocus(u.table)
			return nil
		case tcell.KeyEnter:
			row, _ := u.nav.GetSelection()
			u.activateNav(row)
			return nil
		case tcell.KeyLeft:
			return nil
		}
	case u.table:
		switch e.Key() {
		case tcell.KeyLeft:
			u.app.SetFocus(u.nav)
			return nil
		case tcell.KeyRight:
			if u.width >= 140 {
				u.app.SetFocus(u.inspector)
			} else if p, ok := u.selected(); ok {
				u.detail(p)
			}
			return nil
		case tcell.KeyUp:
			row, _ := u.table.GetSelection()
			if row <= 1 {
				u.app.SetFocus(u.search)
				return nil
			}
		}
	case u.inspector:
		switch e.Key() {
		case tcell.KeyLeft:
			u.app.SetFocus(u.table)
			return nil
		case tcell.KeyRight:
			return nil
		case tcell.KeyEnter:
			if p, ok := u.selected(); ok {
				u.detail(p)
			}
			return nil
		}
	}
	if e.Key() == tcell.KeyEscape {
		u.search.SetText("")
		return nil
	}
	if e.Key() == tcell.KeyCtrlR {
		u.refresh()
		return nil
	}
	switch e.Rune() {
	case '1', '2', '3', '4', '5':
		u.setSection(int(e.Rune() - '1'))
	case 'q':
		u.app.Stop()
	case '/':
		u.app.SetFocus(u.search)
	case 'r':
		u.refresh()
	case 'o':
		if u.sortSize {
			u.sortSize = false
			u.sortDependents = false
		} else if u.sortDependents {
			u.sortDependents = false
			u.sortSize = true
		} else {
			u.sortDependents = true
		}
		u.render()
	case 'f':
		u.filterMenu()
	case 's':
		u.securityReport()
	case ':':
		u.commands()
	case 'v':
		u.scan()
	case 'd':
		u.measure()
	case 'i':
		u.install()
	case 'm', 'c', 'a':
		u.maintenance()
	case 'C':
		u.executeCleanup("cleanup")
	case 'A':
		u.executeCleanup("autoremove")
	case 'u', 'x':
		if p, ok := u.selected(); ok {
			a := "upgrade"
			if e.Rune() == 'x' {
				a = "uninstall"
			}
			u.mutate(a, p.Kind, p.FullName)
		}
	case 't':
		if p, ok := u.selected(); ok {
			u.dependencies(p)
		}
	case ' ':
		if p, ok := u.selected(); ok {
			u.packageActions(p)
		}
	case '?':
		u.help()
	default:
		return e
	}
	return nil
}
func (u *UI) selected() (inventory.Package, bool) {
	r, _ := u.table.GetSelection()
	if r < 1 || r > len(u.visible) {
		return inventory.Package{}, false
	}
	return u.visible[r-1], true
}
func (u *UI) setSection(section int) { u.section = section; u.render() }
func (u *UI) refresh() {
	if !u.ready() {
		return
	}
	u.loading = true
	u.task = "Loading packages and receipts…"
	u.taskArgs = nil
	u.updateChrome()
	go func() {
		ctx, cancel := context.WithTimeout(u.ctx, 2*time.Minute)
		defer cancel()
		s, e := u.client.Inventory(ctx)
		u.app.QueueUpdateDraw(func() {
			u.loading = false
			u.task = ""
			if e != nil {
				u.status = u.tr("Refresh failed. Previous data has been retained.")
				if u.graph == nil {
					u.graph = inventory.NewGraph(inventory.Snapshot{})
				}
				u.message(u.tr("Could not read inventory"), u.errorText(e))
				u.updateChrome()
				return
			}
			needsInsights := !u.initialized || !reflect.DeepEqual(u.snapshot.Packages, s.Packages) || u.report == nil || u.sizes == nil
			u.acceptSnapshot(s)
			u.status = u.tr("Inventory refreshed · local Homebrew metadata")
			u.render()
			if needsInsights {
				u.initialized = true
				u.scan()
				u.measure()
			}
		})
	}()
}
func (u *UI) acceptSnapshot(s inventory.Snapshot) {
	if !reflect.DeepEqual(u.snapshot.Packages, s.Packages) {
		u.report = nil
		u.scanError = ""
		u.sizes = nil
	}
	u.snapshot = s
	u.graph = inventory.NewGraph(s)
}
func (u *UI) ready() bool {
	if u.busy || u.loading || u.scanning || u.measuring {
		u.status = u.tr("Wait for the current task to finish. You can keep browsing.")
		u.updateChrome()
		return false
	}
	return true
}
func Run(ctx context.Context, c *brew.Client) error { return New(ctx, c).app.Run() }

func (u *UI) Run() error { return u.app.Run() }
