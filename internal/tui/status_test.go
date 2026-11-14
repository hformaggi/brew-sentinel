package tui

import (
	"brew-sentinel/internal/inventory"
	"brew-sentinel/internal/security"
	"strings"
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

func TestIncompleteReportDoesNotMarkEveryPackagePartial(t *testing.T) {
	u := workbenchFixture()
	u.report = &security.Report{Complete: false, Skipped: []string{"skipped"}, Outdated: []string{"outdated"}, Unscanned: map[string]string{"failed": "no SBOM"}}
	for _, name := range []string{"ordinary", "skipped", "outdated", "failed"} {
		p := inventory.Package{Kind: "formula", Name: name, FullName: name}
		want := "Not scanned"
		if name == "ordinary" {
			want = "OK"
		}
		if name == "outdated" {
			want = "Limited coverage"
		}
		if got := u.securityShort(p); got != want {
			t.Fatalf("%s: %s", name, got)
		}
		if name == "ordinary" && !strings.Contains(u.securityText(p), "OK — no reported findings") {
			t.Fatal(u.securityText(p))
		}
	}
	p := inventory.Package{Kind: "formula", Name: "skipped", FullName: "other/tap/skipped"}
	if u.securityShort(p) != "OK" {
		t.Fatal("coverage gap leaked across taps")
	}
	p.Versions = []inventory.InstalledVersion{{Version: "1"}, {Version: "2"}}
	if u.securityShort(p) != "Limited coverage" {
		t.Fatal("lost retained-version coverage warning")
	}
	u.report.Findings = []security.Finding{{Formula: "failed", Vulnerabilities: []security.Vulnerability{{ID: "CVE-2026-12345"}}}}
	if got := u.securityShort(inventory.Package{Kind: "formula", FullName: "failed"}); got != "1 findings" {
		t.Fatal("coverage gap hid finding:", got)
	}
}

func TestRetainedVersionsUseDistinctFormulaKegs(t *testing.T) {
	p := inventory.Package{Kind: "formula", Versions: []inventory.InstalledVersion{{Version: "1.0"}, {Version: "1.0"}, {Version: "2.0"}, {}}}
	if got := retainedVersions(p); len(got) != 2 || got[0] != "1.0" || got[1] != "2.0" {
		t.Fatal(got)
	}
	p.Kind = "cask"
	if len(retainedVersions(p)) != 0 {
		t.Fatal("cask treated as retained formula kegs")
	}
	u := workbenchFixture()
	u.report = &security.Report{Outdated: []string{"foo"}}
	p = inventory.Package{Kind: "formula", FullName: "foo"}
	if u.securityGap(p) != "" || u.securityShort(p) != "Limited coverage" {
		t.Fatal("queried package marked unscanned")
	}
	if !strings.Contains(u.securityText(p), "Limited coverage") {
		t.Fatal("detail omitted coverage limitation")
	}
}
