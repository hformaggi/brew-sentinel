package security

import (
	"encoding/json"
	"fmt"
	"time"

	"brew-sentinel/internal/i18n"
)

type Vulnerability struct {
	ID            string   `json:"id"`
	Severity      string   `json:"severity"`
	Summary       string   `json:"summary"`
	Aliases       []string `json:"aliases"`
	FixedVersions []string `json:"fixed_versions"`
}
type Finding struct {
	Formula         string          `json:"formula"`
	Version         string          `json:"version"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	Patched         []Vulnerability `json:"patched"`
}
type Report struct {
	Findings    []Finding         `json:"findings"`
	Checked     int               `json:"checked"`
	Complete    bool              `json:"complete"`
	Skipped     []string          `json:"skipped_formulae"`
	Outdated    []string          `json:"outdated_without_sbom"`
	Unscanned   map[string]string `json:"unscanned_formulae"`
	CollectedAt time.Time         `json:"-"`
}

func Parse(data []byte) (Report, error) {
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(data, &fields); e != nil {
		return Report{}, e
	}
	for _, k := range []string{"findings", "checked", "complete"} {
		if raw, ok := fields[k]; !ok || string(raw) == "null" {
			return Report{}, i18n.Errorf("scan is missing field %s", k)
		}
	}
	var r Report
	e := json.Unmarshal(data, &r)
	if e == nil && r.Checked < 0 {
		return Report{}, i18n.Errorf("invalid checked coverage count")
	}
	r.CollectedAt = time.Now()
	return r, e
}
func (r Report) Text() string { return r.TextWith(func(s string) string { return s }) }
func (r Report) TextWith(tr func(string) string) string {
	s := fmt.Sprintf(tr("Scan: %s\nFormulae checked: %d · Complete coverage: %t\n\nCoverage is NOT confirmed for casks or all retained kegs.\nNo findings means no known matches in the scanned scope.\n"), r.CollectedAt.Format(time.RFC3339), r.Checked, r.Complete)
	for _, f := range r.Findings {
		s += fmt.Sprintf("\n%s — upstream %s\n", f.Formula, f.Version)
		for _, v := range f.Vulnerabilities {
			s += fmt.Sprintf(tr("  %s [%s] %s\n  Upstream fixes: %v\n  https://osv.dev/vulnerability/%s\n"), v.ID, v.Severity, v.Summary, v.FixedVersions, v.ID)
		}
		for _, v := range f.Patched {
			s += tr("  Resolved by a formula patch: ") + v.ID + "\n"
		}
	}
	s += fmt.Sprintf(tr("\nSkipped: %v\nMissing installation evidence: %v\nNot scanned: %v"), r.Skipped, r.Outdated, r.Unscanned)
	return s
}
