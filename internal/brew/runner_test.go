package brew

import (
	"context"
	"testing"
)

func TestRunnerPreservesFailure(t *testing.T) {
	c := Client{Path: "/bin/sh"}
	r, e := c.Run(context.Background(), "-c", "printf output; printf diagnostic >&2; exit 3")
	if e != nil || r.ExitCode != 3 || string(r.Stdout) != "output" || string(r.Stderr) != "diagnostic" {
		t.Fatalf("%+v %v", r, e)
	}
}
func TestRunnerMissingCommand(t *testing.T) {
	c := Client{Path: "/not/a/program"}
	if _, e := c.Run(context.Background()); e == nil {
		t.Fatal("missing launch error")
	}
}
