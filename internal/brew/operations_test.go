package brew

import "testing"

func TestActionArguments(t *testing.T) {
	a, e := ActionArgs("uninstall", "cask", "foo")
	if e != nil || len(a) != 3 || a[1] != "--cask" {
		t.Fatalf("%v %v", a, e)
	}
	for _, name := range []string{"--force", "foo; echo x", "../x", ""} {
		if _, e := ActionArgs("install", "formula", name); e == nil {
			t.Fatalf("accepted %q", name)
		}
	}
	if _, e := ActionArgs("destroy", "formula", "foo"); e == nil {
		t.Fatal("unknown action")
	}
}
func TestPreviewArgs(t *testing.T) {
	a, e := PreviewArgs("upgrade", "formula", "git")
	if e != nil || a[len(a)-1] != "--dry-run" {
		t.Fatalf("%v %v", a, e)
	}
	if _, e := PreviewArgs("uninstall", "formula", "git"); e == nil {
		t.Fatal("invented uninstall dry-run")
	}
}
func TestStreamBound(t *testing.T) {
	var size int
	w := streamWriter{emit: func(s string) { size += len(s) }}
	p := make([]byte, 2*1024*1024)
	n, e := w.Write(p)
	if e != nil || n != len(p) || size != 1024*1024 || len(w.data) != size {
		t.Fatal("unbounded transcript")
	}
}
