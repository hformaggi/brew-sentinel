package brew

import (
	"context"
	"strings"
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

func TestCapturedCommandsDisableHomebrewColors(t *testing.T) {
	c := Client{Path: "/bin/sh"}
	r, err := c.Run(context.Background(), "-c", `test "$HOMEBREW_NO_COLOR" = 1 && test "$NO_COLOR" = 1`)
	if err != nil || r.ExitCode != 0 {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestCommandsDisableImplicitPackageChanges(t *testing.T) {
	for _, key := range []string{"HOMEBREW_NO_AUTO_UPDATE", "HOMEBREW_NO_INSTALL_CLEANUP", "HOMEBREW_NO_AUTOREMOVE"} {
		t.Setenv(key, "0")
	}
	c := Client{Path: "/bin/sh"}
	args := []string{"-c", `test "$HOMEBREW_NO_AUTO_UPDATE" = 1 && test "$HOMEBREW_NO_INSTALL_CLEANUP" = 1 && test "$HOMEBREW_NO_AUTOREMOVE" = 1`}
	t.Run("captured", func(t *testing.T) {
		r, err := c.Run(context.Background(), args...)
		if err != nil || r.ExitCode != 0 {
			t.Fatalf("implicit package changes permitted: %+v %v", r, err)
		}
	})
	t.Run("streamed", func(t *testing.T) {
		r, err := c.Stream(context.Background(), args, func(string) {})
		if err != nil || r.ExitCode != 0 {
			t.Fatalf("implicit package changes permitted: %+v %v", r, err)
		}
	})
	t.Run("interactive", func(t *testing.T) {
		if err := c.Interactive(args); err != nil {
			t.Fatalf("implicit package changes permitted: %v", err)
		}
	})
}

func TestStreamDisablesHomebrewColors(t *testing.T) {
	c := Client{Path: "/bin/sh"}
	var output strings.Builder
	r, err := c.Stream(context.Background(), []string{"-c", `printf '%s:%s' "$HOMEBREW_NO_COLOR" "$NO_COLOR"`}, func(s string) { output.WriteString(s) })
	if err != nil || r.ExitCode != 0 || output.String() != "1:1" {
		t.Fatalf("unexpected color settings: %q %+v %v", output.String(), r, err)
	}
}
