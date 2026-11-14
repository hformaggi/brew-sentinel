package i18n

import (
	"fmt"
	"testing"
)

func TestLocalizedErrorArguments(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", Errorf("could not resolve %s as %s", "foo", "cask"))
	if got := ErrorText(err, New("pt").T); got != "não foi possível resolver foo como cask" {
		t.Fatal(got)
	}
	if got := ErrorText(err, New("en").T); got != "could not resolve foo as cask" {
		t.Fatal(got)
	}
}
