package brew

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseInventory(t *testing.T) {
	root := t.TempDir()
	keg := filepath.Join(root, "foo", "1.0")
	os.MkdirAll(keg, 0755)
	os.WriteFile(filepath.Join(keg, "INSTALL_RECEIPT.json"), []byte(`{"installed_on_request":true,"runtime_dependencies":[{"full_name":"bar","declared_directly":true},{"full_name":"baz","declared_directly":false}]}`), 0644)
	s, err := ParseInventory([]byte(`{"formulae":[{"name":"foo","full_name":"foo","versions":{"stable":"2.0"},"installed":[{"version":"1.0"}],"dependencies":["bar","baz"]}],"casks":[{"token":"foo","full_token":"foo","version":"3.0","installed":"2.0"}]}`), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Packages) != 2 {
		t.Fatal(s)
	}
	p := s.Packages[0]
	if !p.Explicit || len(p.Dependencies) != 1 || len(p.Observed) != 2 {
		t.Fatalf("%+v", p)
	}
	if p.ID == s.Packages[1].ID {
		t.Fatal("identity collision")
	}
}
func TestMissingReceiptUnknown(t *testing.T) {
	s, e := ParseInventory([]byte(`{"formulae":[{"name":"a","full_name":"a","installed":[{"version":"1"}],"dependencies":["b"]}],"casks":[]}`), t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	if s.Packages[0].Reason != "unknown" || s.Packages[0].DependencySource != "current recipe" {
		t.Fatal(s)
	}
}
func TestMalformedInventory(t *testing.T) {
	if _, e := ParseInventory([]byte(`{`), t.TempDir()); e == nil {
		t.Fatal("accepted malformed JSON")
	}
}

func TestCaskCommaVersion(t *testing.T) {
	s, e := ParseInventory([]byte(`{"formulae":[],"casks":[{"token":"app","installed":"1.2,345"}]}`), t.TempDir())
	if e != nil || len(s.Packages[0].Versions) != 1 || s.Packages[0].Versions[0].Version != "1.2,345" {
		t.Fatalf("%+v %v", s, e)
	}
}
func TestPreserveReceiptEdges(t *testing.T) {
	root := t.TempDir()
	keg := filepath.Join(root, "a", "1")
	os.MkdirAll(keg, 0755)
	os.WriteFile(filepath.Join(keg, "INSTALL_RECEIPT.json"), []byte(`{"runtime_dependencies":[{"full_name":"old","declared_directly":true},{"full_name":"unknown"}]}`), 0600)
	s, e := ParseInventory([]byte(`{"formulae":[{"name":"a","installed":[{"version":"1"}],"dependencies":["new"]}],"casks":[]}`), root)
	if e != nil {
		t.Fatal(e)
	}
	if len(s.Packages[0].Dependencies) != 1 || s.Packages[0].Dependencies[0] != "formula:old" {
		t.Fatal(s.Packages[0])
	}
}

func TestOutdatedFromBrew(t *testing.T) {
	s, e := ParseInventory([]byte(`{"formulae":[{"name":"foo","outdated":true}],"casks":[{"token":"bar","installed":"1","outdated":true}]}`), t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	if !s.Packages[0].Outdated || !s.Packages[1].Outdated {
		t.Fatal("missing update state")
	}
}
