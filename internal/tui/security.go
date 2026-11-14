package tui

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"brew-sentinel/internal/security"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var cvePattern = regexp.MustCompile(`\bCVE-[0-9]{4}-[0-9]{4,}\b`)

// Apply terminal hyperlinks after drawing so CVEs in any widget are covered,
// including plain command output. Preserve the cell's colors and attributes.
func linkCVEs(screen tcell.Screen) {
	w, h := screen.Size()
	row := make([]byte, w)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, _, _, _ := screen.GetContent(x, y)
			row[x] = ' '
			if r >= 32 && r < 127 {
				row[x] = byte(r)
			}
		}
		for _, match := range cvePattern.FindAllIndex(row, -1) {
			target := "https://nvd.nist.gov/vuln/detail/" + string(row[match[0]:match[1]])
			for x := match[0]; x < match[1]; x++ {
				r, combining, style, _ := screen.GetContent(x, y)
				screen.SetContent(x, y, r, combining, style.Url(target).Underline(true))
			}
		}
	}
}

func (u *UI) advisoryText(v security.Vulnerability, version string, patched bool) string {
	color := "#8396ad"
	severity := strings.ToUpper(strings.TrimSpace(v.Severity))
	switch severity {
	case "CRITICAL", "HIGH":
		color = "#ff919b"
	case "MEDIUM", "MODERATE":
		color = "#f1c079"
	case "LOW":
		color = "#86baff"
	}
	switch severity {
	case "LOW", "MEDIUM", "MODERATE", "HIGH", "CRITICAL":
	default:
		severity = ""
	}
	if patched {
		color = "#67dfb5"
		severity = u.tr("PATCHED")
	}
	id := tview.Escape(v.ID)
	target := "https://osv.dev/vulnerability/" + url.PathEscape(v.ID)
	var b strings.Builder
	fmt.Fprintf(&b, "\n  [#86baff::b:%s]%s[-::BU:-]", target, id)
	if severity != "" {
		fmt.Fprintf(&b, "  [%s]%s[-]", color, u.tr(severity))
	}
	b.WriteString("\n")
	if !patched {
		if v.Summary != "" {
			fmt.Fprintf(&b, "  %s\n", tview.Escape(v.Summary))
		}
		if len(v.Aliases) > 0 {
			fmt.Fprintf(&b, "  [#8396ad]%s[-]\n", tview.Escape(strings.Join(v.Aliases, " · ")))
		}
		if len(v.FixedVersions) > 0 {
			fmt.Fprintf(&b, "  [#8396ad]%s: %s[-]\n", u.tr("Upstream fixes"), tview.Escape(strings.Join(v.FixedVersions, ", ")))
		}
	}

	return b.String()
}

func (u *UI) securityReportText() string {
	r := u.report
	active, patched := 0, 0
	for _, f := range r.Findings {
		active += len(f.Vulnerabilities)
		patched += len(f.Patched)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "[#67dfb5::b]%s[-::-]\n", u.tr("Security overview"))
	fmt.Fprintf(&b, "%s · %s\n\n", u.tr("Last scan"), r.CollectedAt.Format("2006-01-02 15:04"))
	fmt.Fprintf(&b, "[#ff919b::b]%d[-::-] %s   [#67dfb5::b]%d[-::-] %s   [#86baff::b]%d[-::-] %s\n", active, u.tr("findings"), patched, u.tr("patched"), r.Checked, u.tr("formulae checked"))
	b.WriteString("[#8396ad]" + u.tr("⌘+click to open advisory") + "[-]\n")
	if active == 0 {
		b.WriteString("\n" + u.tr("No known vulnerabilities found in the scanned scope.") + "\n")
	}
	for _, f := range r.Findings {
		if len(f.Vulnerabilities)+len(f.Patched) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n[#67dfb5::b]%s[-::-]  [#8396ad]%s[-]\n", tview.Escape(f.Formula), tview.Escape(f.Version))
		for _, v := range f.Vulnerabilities {
			b.WriteString(u.advisoryText(v, f.Version, false))
		}
		for _, v := range f.Patched {
			b.WriteString(u.advisoryText(v, f.Version, true))
		}
	}
	b.WriteString("\n[#86baff::b]" + u.tr("Scan coverage") + "[-::-]\n")
	coverage := u.tr("Incomplete")
	if r.Complete {
		coverage = u.tr("Complete within the scanner's scope")
	}
	b.WriteString(coverage + "\n" + u.tr("Casks and all retained kegs are not confirmed as covered.") + "\n")
	for _, group := range []struct {
		label string
		items []string
	}{{"Skipped", r.Skipped}, {"Missing installation evidence", r.Outdated}} {
		if len(group.items) > 0 {
			fmt.Fprintf(&b, "\n[#f1c079]%s[-]\n  %s\n", u.tr(group.label), tview.Escape(strings.Join(group.items, ", ")))
		}
	}
	keys := make([]string, 0, len(r.Unscanned))
	for key := range r.Unscanned {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) > 0 {
		b.WriteString("\n[#f1c079]" + u.tr("Not scanned") + "[-]\n")
	}
	for _, key := range keys {
		fmt.Fprintf(&b, "  %s · %s\n", tview.Escape(key), tview.Escape(r.Unscanned[key]))
	}
	return b.String()
}
