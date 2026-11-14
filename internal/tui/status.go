package tui

import (
	"fmt"

	"brew-sentinel/internal/disk"
	"brew-sentinel/internal/inventory"
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
	if reason := u.securityGap(p); reason != "" {
		return u.tr("Not scanned") + ": " + reason
	}
	if u.limitedSecurityCoverage(p) {
		return u.tr("Limited coverage") + ": " + u.tr("The installed version or retained versions could not be fully verified.")
	}
	return u.tr("OK — no reported findings")
}

// Only assign a coverage gap to the formula named by the scanner. A report-wide
// incomplete flag does not identify which of the remaining packages were checked.
func (u *UI) securityGap(p inventory.Package) string {
	if u.report == nil {
		return ""
	}
	name := p.FullName
	if name == "" {
		name = p.Name
	}
	if reason, ok := u.report.Unscanned[name]; ok {
		if reason == "" {
			return u.tr("Not reported")
		}
		return reason
	}
	for _, item := range u.report.Skipped {
		if item == name {
			return u.tr("Skipped by the scanner")
		}
	}
	return ""
}

func (u *UI) limitedSecurityCoverage(p inventory.Package) bool {
	if len(retainedVersions(p)) > 1 {
		return true
	}
	if u.report == nil {
		return false
	}
	name := p.FullName
	if name == "" {
		name = p.Name
	}
	for _, item := range u.report.Outdated {
		if item == name {
			return true
		}
	}
	return false
}

// Multiple distinct formula kegs are evidence of retained versions. A cask's
// version string may contain a build number; it is not a list of old kegs.
func retainedVersions(p inventory.Package) []string {
	if p.Kind != "formula" {
		return nil
	}
	var versions []string
	seen := map[string]bool{}
	for _, v := range p.Versions {
		if v.Version != "" && !seen[v.Version] {
			versions = append(versions, v.Version)
			seen[v.Version] = true
		}
	}
	return versions
}
