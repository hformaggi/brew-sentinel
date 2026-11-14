package tui

import (
	"brew-sentinel/internal/inventory"
	"testing"
)

func TestFilter(t *testing.T) {
	p := []inventory.Package{{Name: "openssl", Description: "TLS library"}, {Name: "git"}}
	if len(Filter(p, "tls")) != 1 {
		t.Fatal("description search failed")
	}
	if len(Filter(p, "GIT")) != 1 {
		t.Fatal("case search failed")
	}
}
func TestTypedFilters(t *testing.T) {
	p := []inventory.Package{{Name: "a", Kind: "formula", Reason: "explicit"}, {Name: "b", Kind: "cask", Reason: "unknown"}}
	if len(Filter(p, "type:cask")) != 1 || len(Filter(p, "reason:explicit")) != 1 {
		t.Fatal("typed filter failed")
	}
}
func BenchmarkFilter(b *testing.B) {
	p := make([]inventory.Package, 1000)
	for i := range p {
		p[i] = inventory.Package{Name: "openssl", Description: "TLS library"}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Filter(p, "tls")
	}
}
func TestUpdateFilter(t *testing.T) {
	p := []inventory.Package{{Name: "new", Outdated: true}, {Name: "current"}}
	got := Filter(p, "outdated:true")
	if len(got) != 1 || got[0].Name != "new" {
		t.Fatal(got)
	}
}
