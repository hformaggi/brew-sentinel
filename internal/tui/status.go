package tui

import (
	"brew-sentinel/internal/disk"
	"brew-sentinel/internal/inventory"
	"fmt"
)

func (u *UI) sizeText(id string) string {
	if v, ok := u.sizes[id]; ok {
		return disk.Format(v)
	}
	return "—"
}
func (u *UI) securityText(p inventory.Package) string {
	if p.Kind == "cask" {
		return u.tr("not supported")
	}
	if u.scanning {
		return u.tr("scanning…")
	}
	if u.scanError != "" {
		return u.tr("scan unavailable")
	}
	if u.report == nil {
		return u.tr("not scanned")
	}
	for _, f := range u.report.Findings {
		if f.Formula == p.FullName {
			if len(f.Vulnerabilities) > 0 {
				return fmt.Sprintf(u.tr("%d findings (%s)"), len(f.Vulnerabilities), f.Version)
			}
		}
	}
	if !u.report.Complete {
		return u.tr("partial coverage")
	}
	if len(p.Versions) > 1 {
		return u.tr("no findings; partial keg coverage")
	}
	return u.tr("no findings (formula)")
}
