package brew

import "testing"

func TestParseSearch(t *testing.T) {
	got := ParseSearch("==> Formulae\nfoo ✔\nbar@2\nhomebrew/tools/baz\n", "formula")
	if len(got) != 3 || got[2].FullName != "homebrew/tools/baz" {
		t.Fatal(got)
	}
}
func TestDescribeCandidate(t *testing.T) {
	c, e := ParseCandidate([]byte(`{"formulae":[{"name":"foo","full_name":"foo","desc":"Useful tool","versions":{"stable":"2"},"installed":[]}],"casks":[]}`), "formula", "foo")
	if e != nil || c.Description != "Useful tool" || c.Version != "2" || c.Installed {
		t.Fatalf("%+v %v", c, e)
	}
}
