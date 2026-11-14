package inventory

import (
	"fmt"
	"testing"
	"time"
)

func TestSharedDependency(t *testing.T) {
	g := NewGraph(Snapshot{Packages: []Package{{ID: "a", Explicit: true, Dependencies: []string{"b"}}, {ID: "b", Dependencies: []string{"d"}}, {ID: "c", Explicit: true, Dependencies: []string{"d"}}, {ID: "d"}}})
	if got := g.Dependents("d"); len(got) != 2 {
		t.Fatalf("dependents: %v", got)
	}
	if got := g.RootPaths("d"); len(got) != 2 {
		t.Fatalf("paths: %v", got)
	}
}
func TestCycleTerminates(t *testing.T) {
	g := NewGraph(Snapshot{Packages: []Package{{ID: "a", Explicit: true, Dependencies: []string{"b"}}, {ID: "b", Dependencies: []string{"a"}}}})
	if len(g.RootPaths("b")) != 1 {
		t.Fatal("missing root path")
	}
}

func TestUnreachablePaths(t *testing.T) {
	s := Snapshot{}
	for i := 0; i < 45; i++ {
		s.Packages = append(s.Packages, Package{ID: fmt.Sprint(i), Explicit: i == 0, Dependencies: []string{fmt.Sprint(i + 1), fmt.Sprint(i + 2)}})
	}
	done := make(chan bool)
	go func() { NewGraph(s).RootPaths("unreachable"); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("unbounded path enumeration")
	}
}

func TestRuntimeDependentsDoNotBecomeDirectDependencies(t *testing.T) {
	g := NewGraph(Snapshot{Packages: []Package{{ID: "formula:ada-url", Observed: []string{"formula:icu4c@78", "formula:icu4c@78"}}, {ID: "formula:même", Dependencies: []string{"formula:icu4c@78"}, Observed: []string{"formula:icu4c@78"}}, {ID: "formula:icu4c@78"}}})
	if got := g.Dependents("formula:icu4c@78"); len(got) != 2 {
		t.Fatal(got)
	}
	if len(g.Dependencies("formula:ada-url")) != 0 {
		t.Fatal("runtime relationship presented as direct")
	}
}
