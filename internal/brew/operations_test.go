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
