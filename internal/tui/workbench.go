package tui

import (
	"brew-sentinel/internal/disk"
	"brew-sentinel/internal/inventory"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"sort"
	"strings"
)

func (u *UI) buildLayout() {
	u.header = textView()
	u.header.SetBackgroundColor(bg)
	summary := tview.NewFlex()
	summary.SetBackgroundColor(bg)
	for i := range u.metrics {
		v := textView()
		v.SetBorderPadding(0, 0, 2, 1)
		u.metrics[i] = v
		summary.AddItem(v, 0, 1, false)
	}
	u.nav = tview.NewTable().SetSelectable(true, false).SetSelectedStyle(tcell.StyleDefault.Foreground(ink).Background(line))
	panel(u.nav.Box, u.tr("Navigation"))
	u.nav.SetBorder(false).SetBorderPadding(1, 0, 1, 1)
	u.nav.Select(1, 0)
	u.nav.SetSelectedFunc(func(row, col int) { u.activateNav(row) })
	u.search = tview.NewInputField().SetLabel(u.tr("Search  ")).SetLabelColor(muted).SetFieldBackgroundColor(surface).SetFieldTextColor(ink).SetPlaceholder(u.tr("Name or description…")).SetPlaceholderTextColor(muted)
	u.search.SetBackgroundColor(surface)
	searchBox := tview.NewFlex().AddItem(u.search, 0, 1, true)
	searchBox.SetBackgroundColor(surface)
	searchBox.SetBorderPadding(0, 0, 1, 1)
	u.table = tview.NewTable().SetSelectable(true, false).SetFixed(1, 0).SetSelectedStyle(tcell.StyleDefault.Foreground(ink).Background(tcell.GetColor("#24453f")))
	panel(u.table.Box, u.tr("Packages"))
	u.inspector = textView().SetScrollable(true)
	panel(u.inspector.Box, u.tr("Package overview"))
	u.packageArea = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(searchBox, 2, 0, true).AddItem(u.table, 0, 1, true)
	u.body = tview.NewFlex()
	u.footer = textView()
	u.footer.SetBackgroundColor(bg)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(u.header, 2, 0, false).AddItem(summary, 3, 0, false).AddItem(u.body, 0, 1, true).AddItem(u.footer, 2, 0, false)
	layout.Box = tview.NewBox().SetBackgroundColor(bg)
	u.pages = tview.NewPages().AddPage("main", layout, true, true)
	u.app.SetRoot(u.pages, true).SetFocus(u.table)
}
func (u *UI) render() {
	u.updateSecurityView()
	if u.graph == nil {
		u.updateChrome()
		return
	}
	selected := ""
	if p, ok := u.selected(); ok {
		selected = p.ID
	}
	u.visible = nil
	for _, p := range Filter(u.snapshot.Packages, u.search.GetText()) {
		match := true
		switch u.section {
		case 1:
			match = p.Kind == "formula"
		case 2:
			match = p.Kind == "cask"
		case 3:
			match = p.Outdated
		case 4:
			match = u.findingsCount(p) > 0
		}
		if match {
			u.visible = append(u.visible, p)
		}
	}
	counts := map[string]int{}
	for _, p := range u.snapshot.Packages {
		counts[p.ID] = len(u.graph.Dependents(p.ID))
	}
	sort.SliceStable(u.visible, func(i, j int) bool {
		a, b := u.visible[i], u.visible[j]
		if u.sortSize {
			x, ox := u.sizes[a.ID]
			y, oy := u.sizes[b.ID]
			if ox != oy {
				return ox
			}
			if x != y {
				return x > y
			}
		}
		if u.sortDependents && counts[a.ID] != counts[b.ID] {
			return counts[a.ID] > counts[b.ID]
		}
		if a.Name == b.Name {
			return a.ID < b.ID
		}
		return a.Name < b.Name
	})
	u.table.Clear()
	listWidth := u.width - 24
	if u.width >= 140 {
		listWidth -= 36
	}
	wide := listWidth >= 88
	compact := u.width > 0 && u.width < 100
	headers := []string{u.tr("Package"), u.tr("Version"), u.tr("Size"), u.tr("Status")}
	if wide {
		headers = []string{u.tr("Package"), u.tr("Version"), u.tr("Size"), u.tr("Update"), u.tr("Security")}
	}
	if compact {
		headers = []string{u.tr("Package"), u.tr("Version"), u.tr("Status")}
	}
	for i, h := range headers {
		u.table.SetCell(0, i, tview.NewTableCell(h).SetTextColor(muted).SetAttributes(tcell.AttrBold).SetSelectable(false).SetBackgroundColor(surface))
	}
	selectedRow := 1
	for i, p := range u.visible {
		name := p.Name
		if p.Kind == "cask" {
			name += " · app"
		}
		state := "—"
		color := muted
		if p.Outdated {
			state = u.tr("Upgrade")
			color = amber
		}
		if p.Pinned {
			state = u.tr("pinned")
		}
		if p.Deprecated || p.Disabled {
			state = u.tr("deprecated")
			color = amber
		}
		if n := u.findingsCount(p); n > 0 {
			state = fmt.Sprintf(u.tr("%d findings"), n)
			color = danger
		}
		values := []string{name, Versions(p), u.sizeText(p.ID), state}
		if wide {
			update := u.tr("Up to date")
			if p.Outdated {
				update = u.tr("Available")
			}
			if p.Pinned {
				update = u.tr("Pinned")
			}
			if p.Deprecated || p.Disabled {
				update = u.tr("Deprecated")
			}
			values = []string{name, Versions(p), u.sizeText(p.ID), update, u.securityShort(p)}
		}
		if compact {
			values = []string{name, Versions(p), state}
		}
		for j, v := range values {
			cell := tview.NewTableCell(tview.Escape(v)).SetTextColor(ink).SetBackgroundColor(surface).SetMaxWidth(18)
			if j == 0 {
				cell.SetExpansion(1).SetMaxWidth(28)
			}
			if j == len(values)-1 {
				cell.SetTextColor(color).SetMaxWidth(16)
			}
			u.table.SetCell(i+1, j, cell)
		}
		if p.ID == selected {
			selectedRow = i + 1
		}
	}
	if len(u.visible) > 0 {
		offset, col := u.table.GetOffset()
		u.table.SetOffset(offset, col)
		u.table.Select(selectedRow, 0)
	} else {
		empty := u.tr("No packages match this selection. Esc clears search.")
		if u.section == 3 && u.search.GetText() == "" {
			empty = u.tr("No updates in local Homebrew metadata.")
		}
		if u.section == 4 {
			if u.report == nil {
				empty = u.tr("Press v to scan for vulnerabilities.")
			} else {
				empty = u.tr("No findings in the scanned scope. See the coverage summary.")
			}
		}
		u.table.SetCell(1, 0, tview.NewTableCell(empty).SetTextColor(muted).SetSelectable(false))
	}
	sections := []string{u.tr("All packages"), u.tr("Formulae"), u.tr("Applications"), u.tr("Updates"), u.tr("Vulnerabilities")}
	u.table.SetTitle(fmt.Sprintf(u.tr(" %s · %d of %d · %s "), sections[u.section], len(u.visible), len(u.snapshot.Packages), u.sortLabel()))
	u.updateInspector()
	u.updateChrome()
}
func (u *UI) sortLabel() string {
	if u.sortSize {
		return u.tr("largest size")
	}
	if u.sortDependents {
		return u.tr("most dependents")
	}
	return "A → Z"
}
func (u *UI) reasonLabel(r string) string {
	switch r {
	case "explicit":
		return u.tr("requested by you")
	case "dependency":
		return u.tr("dependency")
	default:
		return u.tr("unknown")
	}
}
func (u *UI) findingsCount(p inventory.Package) int {
	if u.report == nil || p.Kind != "formula" {
		return 0
	}
	for _, f := range u.report.Findings {
		if f.Formula == p.FullName {
			return len(f.Vulnerabilities)
		}
	}
	return 0
}
func (u *UI) updateChrome() {
	total, formulae, apps, explicit, updates, affected := len(u.snapshot.Packages), 0, 0, 0, 0, 0
	for _, p := range u.snapshot.Packages {
		if p.Kind == "formula" {
			formulae++
		} else {
			apps++
		}
		if p.Explicit {
			explicit++
		}
		if p.Outdated {
			updates++
		}
		if u.findingsCount(p) > 0 {
			affected++
		}
	}
	u.header.SetText(u.tr(" [#67dfb5::b]BREW SENTINEL[-::-]  [#8396ad]Packages, dependencies and security[-]"))
	u.metrics[0].SetText(fmt.Sprintf(u.tr("[#dce6f3::b]%d packages[-::-]\n[#8396ad]%d formulae/%d apps[-]"), total, formulae, apps))
	u.metrics[1].SetText(fmt.Sprintf(u.tr("[#dce6f3::b]%d requested[-::-]\n[#8396ad]%s[-]"), explicit, countLabel(updates, u.tr("update"), u.tr("updates"))))
	size := u.tr("Pending")
	coverage := u.tr("Disk usage")
	if u.sizes != nil {
		var n int64
		for _, v := range u.sizes {
			n += v
		}
		size = disk.Format(n)
		coverage = fmt.Sprintf(u.tr("%d/%d measured"), len(u.sizes), formulae)
	}
	if u.measuring {
		size = u.tr("Calculating…")
	}
	u.metrics[2].SetText("[#dce6f3::b]" + size + "[-::-]\n[#8396ad]" + coverage + "[-]")
	security := u.tr("Pending")
	note := u.tr("Vulnerabilities")
	color := "#8396ad"
	if u.scanning {
		security = u.tr("Scanning…")
		note = u.tr("Vulnerabilities")
		color = "#f1c079"
	} else if u.scanError != "" {
		security = u.tr("Unavailable")
		note = u.tr("View report")
		color = "#f1c079"
	} else if u.report != nil {
		security = fmt.Sprintf(u.tr("%d with findings"), affected)
		color = "#dce6f3"
		if affected > 0 {
			color = "#ff919b"
		}
		note = u.tr("Partial coverage")
		if u.report.Complete {
			note = u.tr("Formulae scanned")
		}
	}
	u.metrics[3].SetText("[" + color + "::b]" + security + "[-::-]\n[#8396ad]" + note + "[-]")
	rows := map[int]string{0: u.tr("Library"), 1: fmt.Sprintf(u.tr("All  %d"), total), 2: fmt.Sprintf(u.tr("Formulae  %d"), formulae), 3: fmt.Sprintf(u.tr("Applications  %d"), apps), 4: fmt.Sprintf(u.tr("Updates %d"), updates), 5: fmt.Sprintf(u.tr("With findings  %d"), affected), 6: "", 7: u.tr("Tools"), 8: u.tr("Search packages"), 9: u.tr("Install package"), 10: u.tr("Filter list"), 11: u.tr("Sort list"), 12: u.tr("Maintenance"), 13: u.tr("Scan report"), 14: u.tr("Navigation help"), 15: u.tr("Language"), 16: u.tr("Quit")}
	for row, label := range rows {
		c := tview.NewTableCell(label).SetTextColor(ink).SetBackgroundColor(surface)
		if row == 0 || row == 6 || row == 7 {
			c.SetSelectable(false).SetTextColor(muted)
		}
		if row == u.section+1 {
			c.SetTextColor(accent)
		}
		u.nav.SetCell(row, 0, c)
	}
	status := u.status
	if status == "" {
		status = u.tr("Installed packages · local Homebrew data")
	}
	if u.scanning && u.measuring {
		status = u.tr("Scanning vulnerabilities and measuring disk usage…")
	} else if u.scanning {
		status = u.tr("Scanning vulnerabilities…")
	} else if u.measuring {
		status = u.tr("Measuring formula disk usage…")
	}
	if u.task != "" {
		status = "● " + fmt.Sprintf(u.tr(u.task), u.taskArgs...)
	}
	hint := u.tr("↑↓ select   ←→ switch pane   Enter open   Esc back")
	if u.search.HasFocus() {
		hint = u.tr("Type to filter · ↑ sidebar · ↓ list · ←→ cursor · Enter apply")
	} else if u.nav.HasFocus() {
		hint = u.tr("↑↓ select · Enter open · → back to list")
	} else if u.inspector.HasFocus() {
		hint = u.tr("↑↓ scroll preview · ← list · Enter details")
	}
	u.footer.SetText(" [#8396ad]" + tview.Escape(strings.Join(strings.Fields(status), " ")) + "[-]\n [#67dfb5]" + hint + "[-]")
}
func (u *UI) updateInspector() {
	p, ok := u.selected()
	if !ok {
		empty := u.tr("[#8396ad]Select a package to view details.[-]")
		if u.section == 4 && u.report == nil {
			empty = u.tr("[#dce6f3::b]Understand your exposure[-::-]\n\n[#8396ad]The scan checks installed formulae for known vulnerabilities. Apps/casks are not covered.\n\n[#67dfb5]v[-] start scan")
		}
		u.inspector.SetText(empty)
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "[#67dfb5::b]%s[-::-]\n[#8396ad]%s[-]\n\n%s\n\n", tview.Escape(p.Name), tview.Escape(p.Kind+" · "+p.FullName), tview.Escape(p.Description))
	fmt.Fprintf(&b, u.tr("[#8396ad]Installed[-]\n%s\n[#8396ad]Available in Homebrew[-]\n%s\n\n[#8396ad]Why is it here[-]\n%s\n\n"), tview.Escape(Versions(p)), tview.Escape(p.Available), u.reasonLabel(p.Reason))
	deps := u.graph.Dependents(p.ID)
	fmt.Fprintf(&b, u.tr("[#8396ad]Relationships[-]\n%d dependencies · %d dependents\n"), len(p.Dependencies), len(deps))
	if len(deps) > 0 {
		b.WriteString(u.tr("\nUsed by\n"))
		for i, id := range deps {
			if i == 4 {
				fmt.Fprintf(&b, u.tr("  + %d more\n"), len(deps)-i)
				break
			}
			b.WriteString("  " + tview.Escape(strings.TrimPrefix(strings.TrimPrefix(id, "formula:"), "cask:")) + "\n")
		}
	}
	fmt.Fprintf(&b, u.tr("\n[#8396ad]Size[-]  %s\n[#8396ad]Security[-]\n%s\n\n[#86baff]Enter[-] open details and actions\n[#8396ad]←[-] back to list"), u.sizeText(p.ID), tview.Escape(u.securityText(p)))
	u.inspector.SetText(b.String()).ScrollToBeginning()
}

func (u *UI) resize(w int) {
	u.width = w
	u.body.Clear()
	side := 24
	if w < 100 {
		side = 22
	}
	u.body.AddItem(u.nav, side, 0, false).AddItem(u.packageArea, 0, 1, true)
	if w >= 140 {
		u.body.AddItem(u.inspector, 36, 0, false)
	} else if u.inspector.HasFocus() {
		go u.app.QueueUpdateDraw(func() {
			if u.inspector.HasFocus() {
				u.app.SetFocus(u.table)
			}
		})
	}
	u.render()
}

func countLabel(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", one)
	}
	return fmt.Sprintf("%d %s", n, many)
}

func (u *UI) securityShort(p inventory.Package) string {
	if p.Kind == "cask" {
		return u.tr("Not covered")
	}
	if u.scanning {
		return u.tr("Scanning")
	}
	if u.scanError != "" {
		return u.tr("Unavailable")
	}
	if u.report == nil {
		return u.tr("Pending")
	}
	if n := u.findingsCount(p); n > 0 {
		return fmt.Sprintf(u.tr("%d findings"), n)
	}
	if !u.report.Complete || len(p.Versions) > 1 {
		return u.tr("Partial")
	}
	return u.tr("No findings")
}
