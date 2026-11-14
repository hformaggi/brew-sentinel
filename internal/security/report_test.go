package security

import "testing"

func TestCoverage(t *testing.T) {
	r, e := Parse([]byte(`{"findings":[{"formula":"foo","version":"1","vulnerabilities":[{"id":"CVE-1","severity":"HIGH"}]}],"checked":1,"complete":false,"skipped_formulae":["bar"]}`))
	if e != nil || r.Complete || len(r.Findings) != 1 {
		t.Fatalf("%+v %v", r, e)
	}
}
func TestInvalid(t *testing.T) {
	for _, s := range []string{`{}`, `{`} {
		if _, e := Parse([]byte(s)); e == nil {
			t.Fatal("accepted incomplete report")
		}
	}
}
func TestNullCoverageRejected(t *testing.T) {
	for _, s := range []string{`{"findings":null,"checked":null,"complete":null}`, `{"findings":[],"checked":-1,"complete":true}`} {
		if _, e := Parse([]byte(s)); e == nil {
			t.Fatal("accepted invalid coverage", s)
		}
	}
}
