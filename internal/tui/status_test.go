package tui

import (
	"brew-sentinel/internal/inventory"
	"brew-sentinel/internal/security"
	"testing"
)

func TestSecurityDoesNotConfuseTaps(t *testing.T) {
	u := workbenchFixture()
	u.report = &security.Report{Complete: true, Findings: []security.Finding{{Formula: "foo", Version: "1", Vulnerabilities: []security.Vulnerability{{ID: "CVE-1"}}}}}
	p := inventory.Package{Kind: "formula", Name: "foo", FullName: "other/tap/foo"}
	if u.securityText(p) == "1 achados (1)" {
		t.Fatal("matched unrelated tap")
	}
}
func TestFailedScanIsNotClean(t *testing.T) {
	u := workbenchFixture()
	u.scanError = "offline"
	if got := u.securityText(inventory.Package{Kind: "formula"}); got != "scan unavailable" {
		t.Fatal(got)
	}
}
